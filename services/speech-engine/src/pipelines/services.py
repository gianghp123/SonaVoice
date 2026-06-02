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


def _get_metrics():
    return SentryMetrics() if settings.SENTRY_DSN else None


def build_stt() -> DeepgramSTTService:
    return DeepgramSTTService(
        api_key=settings.DEEPGRAM_API_KEY,
        metrics=_get_metrics(),
        ttfs_p99_latency=0.4,
        settings=DeepgramSTTService.Settings(
            interim_results=True,
            punctuate=False,
            endpointing=300,
        ),
    )


def build_llm() -> OpenAILLMService:
    llm = OpenAILLMService(
        api_key=settings.OPENCODE_API_KEY,
        base_url=settings.OPENAI_BASE_URL,
        metrics=_get_metrics(),
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


def build_tts() -> PiperTTSService:
    return PiperTTSService(
        metrics=_get_metrics(),
        settings=PiperTTSService.Settings(
            voice="en_US-lessac-high",
        ),
        download_dir=settings.piper_models_dir,
    )


def build_context() -> LLMContext:
    tools = ToolsSchema([summarize_function])
    return LLMContext(tools=tools)


def build_memory(user_id: str, session_id: str) -> Mem0MemoryService:
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
