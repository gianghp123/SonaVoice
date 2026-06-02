# Refactor Speech Engine Structure

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move misplaced functions to correct modules, split god functions, deduplicate code, and remove unnecessary files.

**Architecture:** Extract service initialization from `main.py` into `src/pipelines/services.py`, split event handlers from `src/pipelines/tasks.py` into `src/pipelines/handlers.py`, move `get_piper_models_dir()` to `src/core/config.py`, add pydantic input validation, deduplicate Sentry setup, remove `src/transports/` thin wrapper.

**Tech Stack:** Python 3.12, Pipecat 1.1.0, pydantic-settings

---

### Task 1: Move `get_piper_models_dir()` to `src/core/config.py`

**Files:**
- Modify: `src/core/config.py`
- Modify: `main.py`

- [ ] **Step 1: Add property to Settings class**

Insert after the `EMBEDDING_DIMS` field (after line 39) in `src/core/config.py`:

```python
    @property
    def piper_models_dir(self) -> Path:
        if os.environ.get("MODEL_ENVIRONMENT") == "local":
            return Path("./models")
        return Path("/root/models")
```

Also add `Path` to the imports at the top of `src/core/config.py`. Change line 1 from `import os` to:

```python
import os
from pathlib import Path
```

- [ ] **Step 2: Update `main.py` to use the new property**

Remove the `get_piper_models_dir()` function (lines 33-37) in `main.py`:

```python
def get_piper_models_dir() -> Path:
    if os.environ.get("MODEL_ENVIRONMENT") == "local":
        return Path("./models")

    return Path("/root/models")
```

Replace the call on line 169 from `download_dir=get_piper_models_dir()` to `download_dir=settings.piper_models_dir`.

Remove the now-unused imports `from pathlib import Path` and `import os` from `main.py` (lines 28-29), since `Path` and `os` are no longer used directly.

- [ ] **Step 3: Verify the file runs without import errors**

```bash
python -c "from src.core.config import settings; print(settings.piper_models_dir)"
```

Expected: prints a Path object (no errors).

---

### Task 2: Add `RunnerBody` pydantic model for input validation

**Files:**
- Modify: `src/core/config.py`

- [ ] **Step 1: Add RunnerBody model to config.py**

Insert after the `Settings` class and before `settings = Settings()` (before line 73) in `src/core/config.py`:

```python
from pydantic import BaseModel


class RunnerBody(BaseModel):
    user_id: str
    session_id: str
    max_duration: int
```

- [ ] **Step 2: Verify the model works**

```bash
python -c "
from src.core.config import RunnerBody
body = RunnerBody(user_id='u1', session_id='s1', max_duration=300)
print(body.model_dump())
"
```

Expected: `{'user_id': 'u1', 'session_id': 's1', 'max_duration': 300}`

---

### Task 3: Replace manual field validation in `main.py` with `RunnerBody`

**Files:**
- Modify: `main.py`

- [ ] **Step 1: Rewrite field extraction and validation in `run_bot()`**

Replace lines 85-106 in `main.py`:

```python
async def run_bot(transport: BaseTransport, runner_args: RunnerArguments):
    if not runner_args.body:
        raise ValueError("runner body is required")

    user_id = runner_args.body.get("user_id")
    session_id = runner_args.body.get("session_id")
    max_duration = runner_args.body.get("max_duration")

    missing_fields = [
        field
        for field, value in {
            "user_id": user_id,
            "session_id": session_id,
            "max_duration": max_duration,
        }.items()
        if value is None
    ]

    if missing_fields:
        raise ValueError(
            f"Missing required fields: {', '.join(missing_fields)}"
        )
```

Replace with:

```python
from src.core.config import RunnerBody
```

(add to existing imports at top of file) and:

```python
async def run_bot(transport: BaseTransport, runner_args: RunnerArguments):
    if not runner_args.body:
        raise ValueError("runner body is required")

    body = RunnerBody(**runner_args.body)
    user_id = body.user_id
    session_id = body.session_id
    max_duration = body.max_duration
```

- [ ] **Step 2: Verify the refactored file has no syntax errors**

```bash
python -c "import main" 2>&1 | head -5
```

Expected: no syntax errors (import-time side effects like sentry init are fine).

---

### Task 4: Extract service initialization to `src/pipelines/services.py`

**Files:**
- Create: `src/pipelines/services.py`
- Modify: `main.py`

- [ ] **Step 1: Create `src/pipelines/services.py`**

```python
from pipecat.services.openai.llm import OpenAILLMService
from pipecat.services.deepgram.stt import DeepgramSTTService
from pipecat.services.piper.tts import PiperTTSService
from pipecat.processors.aggregators.llm_context import LLMContext
from pipecat.adapters.schemas.tools_schema import ToolsSchema
from pipecat.processors.metrics.sentry import SentryMetrics
from pipecat.services.mem0.memory import Mem0MemoryService

from src.core.config import settings
from src.agents.prompts import ENGLIST_TEACHER_SYSTEM_INSTRUCTION
from src.agents.tools import summarize_conversation, summarize_function


def build_stt():
    return DeepgramSTTService(
        api_key=settings.DEEPGRAM_API_KEY,
        metrics=SentryMetrics(),
        ttfs_p99_latency=0.4,
        settings=DeepgramSTTService.Settings(
            interim_results=True,
            punctuate=False,
            endpointing=300,
        ),
    )


def build_llm():
    llm = OpenAILLMService(
        api_key=settings.OPENCODE_API_KEY,
        base_url=settings.OPENAI_BASE_URL,
        metrics=SentryMetrics(),
        settings=OpenAILLMService.Settings(
            model=settings.LLM_NAME,
            system_instruction=ENGLIST_TEACHER_SYSTEM_INSTRUCTION.format(
                name=settings.BOT_NAME
            ),
            extra={"extra_body": {"thinking": {"type": "disabled"}}},
        ),
    )
    llm.register_function("summarize_conversation", summarize_conversation)
    return llm


def build_tts():
    return PiperTTSService(
        metrics=SentryMetrics(),
        settings=PiperTTSService.Settings(
            voice="en_US-lessac-high",
        ),
        download_dir=settings.piper_models_dir,
    )


def build_context():
    tools = ToolsSchema([summarize_function])
    return LLMContext(tools=tools)


def build_memory(user_id: str, session_id: str):
    return Mem0MemoryService(
        local_config=settings.memory_config,
        user_id=user_id,
        run_id=session_id,
        params=Mem0MemoryService.InputParams(
            search_limit=10,
            search_threshold=0.5,
            add_as_system_message=True,
            position=1,
        ),
    )
```

- [ ] **Step 2: Rewrite `main.py` imports and service initialization**

Remove these imports from `main.py` (lines 1-3, 7, 19-23, 26-27):

```python
from pipecat.services.openai.llm import OpenAILLMService
from pipecat.services.deepgram.stt import DeepgramSTTService
from src.agents.prompts import ENGLIST_TEACHER_SYSTEM_INSTRUCTION
from src.agents.tools import summarize_conversation
from pipecat.processors.aggregators.llm_context import LLMContext
from pipecat.adapters.schemas.tools_schema import ToolsSchema
from src.agents.tools import summarize_function
from pipecat.services.piper.tts import PiperTTSService
from pipecat.processors.metrics.sentry import SentryMetrics
from pipecat.services.mem0.memory import Mem0MemoryService
```

Add this import:

```python
from src.pipelines.services import build_stt, build_llm, build_tts, build_context, build_memory
```

Replace the service initialization block in `run_bot()` (lines 136-187):

```python
        log.info("Initializing voice services")

        stt = DeepgramSTTService(
            api_key=settings.DEEPGRAM_API_KEY,
            metrics=SentryMetrics(),
            ttfs_p99_latency=0.4,
            settings=DeepgramSTTService.Settings(
                interim_results=True,
                punctuate=False,
                endpointing=300,
            )
        )

        llm = OpenAILLMService(
            api_key=settings.OPENCODE_API_KEY,
            base_url=settings.OPENAI_BASE_URL,
            metrics=SentryMetrics(),
            settings=OpenAILLMService.Settings(
                model=settings.LLM_NAME,
                system_instruction=ENGLIST_TEACHER_SYSTEM_INSTRUCTION.format(
                    name=settings.BOT_NAME
                ),
                extra={"extra_body": {"thinking": {"type": "disabled"}}},
            ),
        )

        llm.register_function("summarize_conversation", summarize_conversation)

        tts = PiperTTSService(
            metrics=SentryMetrics(),
            settings=PiperTTSService.Settings(
                voice="en_US-lessac-high",
            ),
            download_dir=get_piper_models_dir(),
        )

        tools = ToolsSchema([summarize_function])
        log.info("Memory enabled for session")

        context = LLMContext(tools=tools)

        memory_processor = Mem0MemoryService(
            local_config=settings.memory_config,
            user_id=user_id,
            run_id=session_id,
            params=Mem0MemoryService.InputParams(
                search_limit=10,
                search_threshold=0.5,
                add_as_system_message=True,
                position=1,
            ),
        )
```

Replace with:

```python
        log.info("Initializing voice services")

        stt = build_stt()
        llm = build_llm()
        tts = build_tts()
        context = build_context()
        memory_processor = build_memory(user_id, session_id)
```

- [ ] **Step 3: Verify the refactored file has no import errors**

```bash
python -c "
from src.pipelines.services import build_stt, build_llm, build_tts, build_context, build_memory
print('services module OK')
"
```

Expected: prints "services module OK".

---

### Task 5: Split event handlers into `src/pipelines/handlers.py`

**Files:**
- Create: `src/pipelines/handlers.py`
- Modify: `src/pipelines/tasks.py`

- [ ] **Step 1: Create `src/pipelines/handlers.py`**

```python
import asyncio
import time

from loguru import logger
from pipecat.frames.frames import (
    LLMRunFrame,
    LLMMessagesAppendFrame,
    EndTaskFrame,
    ErrorFrame,
    LLMSummarizeContextFrame,
)
from pipecat.processors.frame_processor import FrameDirection
from pipecat.utils.context.llm_context_summarization import LLMContextSummaryConfig

from src.types.messages import SessionMessage, UserMessage, AssistantMessage
from src.utils.error_msg import get_custom_error_message
from src.services.session_service import SessionService
from src.services.messages_service import MessageService


def register_event_handlers(
    task,
    transport,
    stt,
    llm,
    tts,
    context,
    user_aggregator,
    assistant_aggregator,
    session_id: str,
    user_id: str,
    start_time: float,
    max_duration: int | None,
    has_memory: bool,
):
    session_messages: list[SessionMessage] = []
    session_service = SessionService()
    message_service = MessageService()

    log = logger.bind(
        area="voice-task",
        user_id=user_id,
        session_id=session_id,
        max_duration=max_duration,
        has_memory=has_memory,
    )

    @user_aggregator.event_handler("on_user_turn_stopped")
    async def on_user_turn_stopped(aggregator, strategy, message):
        session_messages.append(
            UserMessage(
                role="user",
                transcript=message.content,
                created_at=message.timestamp,
            )
        )

    @assistant_aggregator.event_handler("on_assistant_turn_stopped")
    async def on_assistant_turn_stopped(aggregator, message):
        if message.interrupted:
            session_messages.append(
                AssistantMessage(
                    role="assistant",
                    transcript=message.content,
                    was_interrupted=True,
                    created_at=message.timestamp,
                )
            )
            log.info(
                "Assistant interrupted",
                content_length=len(message.content or ""),
            )
        elif message.content:
            session_messages.append(
                AssistantMessage(
                    role="assistant",
                    transcript=message.content,
                    was_interrupted=False,
                    created_at=message.timestamp,
                )
            )

    @user_aggregator.event_handler("on_user_turn_idle")
    async def on_user_turn_idle(aggregator):
        log.info("User idle")
        msg = {
            "role": "developer",
            "content": "The user is quiet. Ask if they are there.",
        }
        await aggregator.push_frame(LLMMessagesAppendFrame([msg], run_llm=True))

    @transport.event_handler("on_client_disconnected")
    async def on_client_disconnected(transport, client):
        actual_usage = int(time.time() - start_time)
        log.info(
            "Client disconnected",
            actual_usage=actual_usage,
            message_count=len(session_messages),
        )

        try:
            await session_service.close_session(
                session_id=session_id,
                actual_usage=actual_usage,
            )

            if session_messages and len(session_messages) > 0:
                await message_service.save_messages(session_id, session_messages)
                log.info(
                    "Session closed and messages saved",
                    actual_usage=actual_usage,
                    message_count=len(session_messages),
                )
            else:
                log.info(
                    "Session closed with no messages to save",
                    actual_usage=actual_usage,
                )
        except Exception:
            log.exception(
                "Failed to close session or save messages",
                actual_usage=actual_usage,
                message_count=len(session_messages),
            )
        finally:
            await task.cancel()

    async def session_timer(task, aggregator, timeout_secs=300):
        await asyncio.sleep(timeout_secs)
        log.info(
            "Session max duration reached",
            timeout_secs=timeout_secs,
        )
        await aggregator.push_frame(
            LLMMessagesAppendFrame(
                messages=[
                    {
                        "role": "system",
                        "content": "Say goodbye to the user since the session is over.",
                    }
                ],
                run_llm=True,
            )
        )
        await aggregator.push_frame(EndTaskFrame(), FrameDirection.UPSTREAM)
        actual_duration = time.time() - start_time
        log.info(
            "Session timer finished",
            actual_duration=actual_duration,
        )

    @transport.event_handler("on_client_connected")
    async def on_client_connected(transport, client):
        log.info("Client connected")
        context.add_message(
            {
                "role": "system",
                "content": "Say hello and briefly introduce yourself.",
            }
        )
        await task.queue_frames([LLMRunFrame()])

        if max_duration is not None:
            asyncio.create_task(
                session_timer(
                    task,
                    user_aggregator,
                    timeout_secs=max_duration,
                )
            )

    @task.event_handler("on_pipeline_finished")
    async def on_pipeline_finished(task, frame):
        log.info("Pipeline finished")
        await task.queue_frames(
            [
                LLMSummarizeContextFrame(
                    config=LLMContextSummaryConfig(
                        target_context_tokens=4000,
                        min_messages_after_summary=0,
                    )
                )
            ]
        )

    @tts.event_handler("on_connection_error")
    async def on_tts_connection_error(service, error):
        log.error("TTS connection error", stage="tts", error=str(error))

    @stt.event_handler("on_connection_error")
    async def on_stt_connection_error(service, error):
        log.error("STT connection error", stage="stt", error=str(error))

    @llm.event_handler("on_connection_error")
    async def on_llm_connection_error(service, error):
        log.error("LLM connection error", stage="llm", error=str(error))

    @task.event_handler("on_pipeline_error")
    async def on_pipeline_error(task, frame):
        error_msg = str(frame.error)
        log.error(
            "Pipeline error",
            stage="pipeline",
            fatal=frame.fatal,
            error=error_msg,
        )

        if "Cartesia" in error_msg or "TTS" in error_msg or "WebSocket" in error_msg:
            service_name = "TTS service"
        elif "STT" in error_msg:
            service_name = "STT service"
        elif "LLM" in error_msg:
            service_name = "LLM service"
        else:
            service_name = "Service"

        custom_msg = get_custom_error_message(error_msg, service_name)
        await task.queue_frames([ErrorFrame(error=custom_msg, fatal=True)])
```

- [ ] **Step 2: Update `src/pipelines/tasks.py` to call the handler registration**

Replace the entire file content of `src/pipelines/tasks.py`:

```python
from pipecat.audio.vad.silero import SileroVADAnalyzer
from pipecat.pipeline.pipeline import Pipeline
from pipecat.pipeline.task import PipelineParams, PipelineTask
from pipecat.processors.aggregators.llm_context import LLMContext
from pipecat.processors.aggregators.llm_response_universal import (
    LLMContextAggregatorPair,
    LLMUserAggregatorParams,
    LLMAssistantAggregatorParams,
)
from pipecat.utils.context.llm_context_summarization import (
    LLMAutoContextSummarizationConfig,
    LLMContextSummaryConfig,
)
from loguru import logger

from src.pipelines.handlers import register_event_handlers


async def create_voice_bot_task(
    transport,
    stt,
    llm,
    tts,
    context: LLMContext,
    session_id,
    user_id,
    start_time=None,
    max_duration=None,
    memory_processor=None,
) -> PipelineTask:
    if start_time is None:
        start_time = __import__("time").time()

    log = logger.bind(
        area="voice-task",
        user_id=user_id,
        session_id=session_id,
        max_duration=max_duration,
        has_memory=memory_processor is not None,
    )

    user_aggregator, assistant_aggregator = LLMContextAggregatorPair(
        context,
        user_params=LLMUserAggregatorParams(
            vad_analyzer=SileroVADAnalyzer(),
        ),
        assistant_params=LLMAssistantAggregatorParams(
            enable_auto_context_summarization=True,
            auto_context_summarization_config=LLMAutoContextSummarizationConfig(
                max_context_tokens=8000,
                max_unsummarized_messages=20,
                summary_config=LLMContextSummaryConfig(
                    target_context_tokens=6000,
                    min_messages_after_summary=4,
                ),
            ),
        ),
    )

    if memory_processor is None:
        log.info("Creating pipeline without memory processor")
        pipeline = Pipeline(
            [
                transport.input(),
                stt,
                user_aggregator,
                llm,
                tts,
                transport.output(),
                assistant_aggregator,
            ]
        )
    else:
        log.info("Creating pipeline with memory processor")
        pipeline = Pipeline(
            [
                transport.input(),
                stt,
                user_aggregator,
                memory_processor,
                llm,
                tts,
                transport.output(),
                assistant_aggregator,
            ]
        )

    task = PipelineTask(
        pipeline,
        params=PipelineParams(
            enable_metrics=True,
            enable_usage_metrics=True,
        ),
    )

    register_event_handlers(
        task=task,
        transport=transport,
        stt=stt,
        llm=llm,
        tts=tts,
        context=context,
        user_aggregator=user_aggregator,
        assistant_aggregator=assistant_aggregator,
        session_id=session_id,
        user_id=user_id,
        start_time=start_time,
        max_duration=max_duration,
        has_memory=memory_processor is not None,
    )

    return task
```

- [ ] **Step 3: Verify no import errors**

```bash
python -c "from src.pipelines.tasks import create_voice_bot_task; print('tasks module OK')"
python -c "from src.pipelines.handlers import register_event_handlers; print('handlers module OK')"
```

Expected: both print OK without errors.

---

### Task 6: Remove `src/transports/` directory, inline into `main.py`

**Files:**
- Remove: `src/transports/__init__.py`
- Remove: `src/transports/small_webrtc.py`
- Delete: `src/transports/` directory
- Modify: `main.py`

- [ ] **Step 1: Inline transport creation into `main.py`**

Replace the import on line 5 of `main.py`:

```python
from src.transports.small_webrtc import create_small_webrtc_transport
```

With:

```python
from pipecat.transports.smallwebrtc.transport import SmallWebRTCTransport
```

Replace the call on lines 59-65:

```python
            transport = create_small_webrtc_transport(
                webrtc_connection,
                params=TransportParams(
                    audio_in_enabled=True,
                    audio_out_enabled=True,
                ),
            )
```

With:

```python
            transport = SmallWebRTCTransport(
                webrtc_connection=webrtc_connection,
                params=TransportParams(
                    audio_in_enabled=True,
                    audio_out_enabled=True,
                ),
            )
```

- [ ] **Step 2: Delete the transports directory**

```bash
rm -f src/transports/__init__.py src/transports/small_webrtc.py && rmdir src/transports/ 2>/dev/null; echo "done"
```

- [ ] **Step 3: Verify import of `main` module still works**

```bash
python -c "import main" 2>&1 | head -5
```

Expected: no errors (sentry init log lines are fine).

---

### Task 7: Deduplicate Sentry context setup

**Files:**
- Modify: `src/pipelines/tasks.py`

- [ ] **Step 1: Remove Sentry setup from `tasks.py`**

The `create_voice_bot_task()` function in `src/pipelines/tasks.py` no longer has the Sentry setup lines (they were part of the old content removed in Task 5). Verify that the new `tasks.py` from Task 5 does NOT include:

```python
import sentry_sdk
sentry_sdk.set_user({"id": user_id})
sentry_sdk.set_context("voice_session", {...})
```

The Sentry setup already happens in `main.py:run_bot()` at lines 115-123, which is the correct single place for it.

- [ ] **Step 2: Confirm Sentry is only set in `main.py`**

```bash
grep -rn "sentry_sdk.set_user\|sentry_sdk.set_context" src/ main.py
```

Expected: only matches in `main.py`, not in `src/`.

---

### Task 8: Final verification

**Files:** All modified files

- [ ] **Step 1: Run a full import check**

```bash
python -c "
import main
from src.core.config import settings, RunnerBody
from src.pipelines.services import build_stt, build_llm, build_tts, build_context, build_memory
from src.pipelines.tasks import create_voice_bot_task
from src.pipelines.handlers import register_event_handlers
print('All imports OK')
"
```

Expected: prints "All imports OK".

- [ ] **Step 2: Verify module structure**

```bash
find src -name "*.py" | sort
```

Expected output:
```
src/__init__.py
src/agents/__init__.py
src/agents/prompts.py
src/agents/tools.py
src/core/__init__.py
src/core/config.py
src/pipelines/__init__.py
src/pipelines/handlers.py
src/pipelines/services.py
src/pipelines/tasks.py
src/runner.py
src/services/__init__.py
src/services/messages_service.py
src/services/session_service.py
src/types/__init__.py
src/types/messages.py
src/utils/__init__.py
src/utils/error_msg.py
```

No `transports/` directory should exist.

- [ ] **Step 3: Final `main.py` review**

Read the final `main.py`. It should be ~120 lines (down from 227), containing only:
- Imports (standard + pipecat + src)
- Sentry init
- `bot()` — transport dispatch
- `run_bot()` — validation + service init + pipeline creation + runner execution
- `if __name__ == "__main__"` block

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "refactor: split god functions, deduplicate, remove dead transports module"

Changes:
- Move get_piper_models_dir() to Settings.piper_models_dir property
- Add RunnerBody pydantic model for input validation
- Extract service init to src/pipelines/services.py
- Split event handlers to src/pipelines/handlers.py
- Remove src/transports/ (inline into main.py)
- Deduplicate Sentry context setup (keep in main.py only)
- main.py reduced from 227 to ~120 lines
```
