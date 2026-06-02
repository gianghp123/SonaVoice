from pipecat.pipeline.runner import PipelineRunner
from src.core.config import RunnerBody, settings
from pipecat.transports.smallwebrtc.transport import SmallWebRTCTransport
from src.pipelines.tasks import create_voice_bot_task
from pipecat.runner.types import (
    RunnerArguments,
    SmallWebRTCRunnerArguments,
)
from pipecat.transports.smallwebrtc.transport import (
    TransportParams,
    SmallWebRTCConnection,
)
from pipecat.transports.base_transport import BaseTransport
from loguru import logger
from dotenv import load_dotenv
from src.pipelines.services import build_stt, build_llm, build_tts, build_context, build_memory
import time
import sentry_sdk
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

        if isinstance(runner_args, SmallWebRTCRunnerArguments):
            log.info("Creating SmallWebRTC transport")

            webrtc_connection: SmallWebRTCConnection = runner_args.webrtc_connection

            transport = SmallWebRTCTransport(
                webrtc_connection=webrtc_connection,
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
    if not runner_args.body:
        raise ValueError("runner body is required")

    body = RunnerBody(**runner_args.body)
    user_id = body.user_id
    session_id = body.session_id
    max_duration = body.max_duration

    logger.bind(
        area="bot",
        user_id=user_id,
        session_id=session_id,
        max_duration=max_duration,
    ).info("Runner body received")

    if settings.SENTRY_DSN:
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

        stt = build_stt()
        llm = build_llm()
        tts = build_tts()
        context = build_context()
        memory_processor = build_memory(user_id, session_id)
        log.info("Memory enabled for session")

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
