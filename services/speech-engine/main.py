from pipecat.pipeline.runner import PipelineRunner
from pipecat.services.openai.llm import OpenAILLMService
from pipecat.services.deepgram.stt import DeepgramSTTService
from src.core.config import settings
from src.pipelines.processors import CustomMem0Processor
from src.transports.daily import create_daily_transport
from src.transports.small_webrtc import create_small_webrtc_transport
from src.pipelines.tasks import create_voice_bot_task
from mem0 import AsyncMemory
from src.agents.prompts import ENGLIST_TEACHER_SYSTEM_INSTRUCTION
from pipecat.runner.types import (
    DailyRunnerArguments,
    RunnerArguments,
    SmallWebRTCRunnerArguments,
)
from pipecat.transports.smallwebrtc.transport import (
    TransportParams,
    SmallWebRTCConnection,
)
from pipecat.transports.base_transport import BaseTransport
from loguru import logger
from pipecat.transports.daily.transport import DailyParams
from dotenv import load_dotenv
from src.agents.tools import summarize_conversation, save_user_preferences
from pipecat.processors.aggregators.llm_context import LLMContext
from pipecat.adapters.schemas.tools_schema import ToolsSchema
from src.agents.tools import summarize_function, save_memory_function
from pipecat.services.piper.tts import PiperTTSService
import time
import sentry_sdk
from pipecat.processors.metrics.sentry import SentryMetrics

load_dotenv()

if settings.SENTRY_DSN:
    sentry_sdk.init(
        dsn=settings.SENTRY_DSN,
        traces_sample_rate=1.0,
        enable_logs=True,
        send_default_pii=False,
    )


async def bot(runner_args: RunnerArguments):
    log = logger.bind(area="bot")

    try:
        transport = None

        if isinstance(runner_args, DailyRunnerArguments):
            log.info("Creating Daily transport")

            transport = create_daily_transport(
                room_url=runner_args.room_url,
                token=runner_args.token,
                bot=settings.BOT_NAME,
                params=DailyParams(
                    audio_in_enabled=True,
                    audio_out_enabled=True,
                    transcription_enabled=True,
                ),
            )

        elif isinstance(runner_args, SmallWebRTCRunnerArguments):
            log.info("Creating SmallWebRTC transport")

            webrtc_connection: SmallWebRTCConnection = runner_args.webrtc_connection

            transport = create_small_webrtc_transport(
                webrtc_connection,
                params=TransportParams(
                    audio_in_enabled=True,
                    audio_out_enabled=True,
                ),
            )

        else:
            log.error(
                "Unsupported runner arguments type",
                runner_args_type=str(type(runner_args)),
            )
            return

        if transport is None:
            log.error("Failed to create transport")
            return

        await run_bot(transport, runner_args)

    except Exception:
        log.exception("Bot crashed")
        raise


async def run_bot(transport: BaseTransport, runner_args: RunnerArguments):
    user_id = "anonymous"
    session_id = None
    max_duration = None
    memory_client = None
    tools = None

    if runner_args.body:
        user_id = runner_args.body.get("user_id") or "anonymous"
        session_id = runner_args.body.get("session_id")
        max_duration = runner_args.body.get("max_duration")

        logger.bind(
            area="bot",
            user_id=user_id,
            session_id=session_id,
            max_duration=max_duration,
        ).info("Runner body received")
    else:
        max_duration = 60 * 5

        logger.bind(
            area="bot",
            user_id=user_id,
            session_id=session_id,
            max_duration=max_duration,
        ).warning("No runner body received, using default max duration")

    sentry_sdk.set_user({"id": user_id})
    sentry_sdk.set_context(
        "voice_session",
        {
            "session_id": session_id,
            "max_duration": max_duration,
            "transport_type": type(transport).__name__,
        },
    )

    log = logger.bind(
        area="bot",
        user_id=user_id,
        session_id=session_id,
        max_duration=max_duration,
        transport_type=type(transport).__name__,
    )

    start_time = time.time()

    try:
        log.info("Initializing voice services")

        stt = DeepgramSTTService(
            api_key=settings.DEEPGRAM_API_KEY,
            metrics=SentryMetrics(),
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
        llm.register_function("save_user_preferences", save_user_preferences)

        tts = PiperTTSService(
            metrics=SentryMetrics(),
            settings=PiperTTSService.Settings(
                voice="en_US-lessac-high",
            ),
        )

        if session_id is None or session_id == "":
            tools = ToolsSchema([summarize_function])
            log.info("Memory disabled for anonymous session")
        else:
            tools = ToolsSchema([summarize_function, save_memory_function])
            memory_client = AsyncMemory.from_config(settings.memory_config)
            log.info("Memory enabled for session")

        context = LLMContext(tools=tools)

        memory_processor = CustomMem0Processor(
            memory_client=memory_client,
            user_id=user_id,
            session_id=session_id,
        )

        task = await create_voice_bot_task(
            transport,
            stt,
            llm,
            tts,
            context,
            session_id,
            user_id,
            start_time=start_time,
            max_duration=max_duration,
            memory_processor=memory_processor,
        )

        log.info("Starting pipeline runner")

        runner = PipelineRunner(handle_sigint=False)
        await runner.run(task)

        duration_seconds = time.time() - start_time

        log.info(
            "Bot run finished",
            duration_seconds=duration_seconds,
        )

    except Exception:
        duration_seconds = time.time() - start_time

        log.exception(
            "Bot runtime crashed",
            duration_seconds=duration_seconds,
        )
        raise


if __name__ == "__main__":
    from src.runner import main

    main()