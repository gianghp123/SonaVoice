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
from pipecat.processors.aggregators.llm_context import LLMContext
from pipecat.processors.aggregators.llm_response_universal import (
    UserTurnStoppedMessage,
    AssistantTurnStoppedMessage,
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
    context: LLMContext,
    user_aggregator,
    assistant_aggregator,
    session_id: str,
    user_id: str,
    start_time: float,
    max_duration,
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
    async def on_user_turn_stopped(aggregator, strategy, message: UserTurnStoppedMessage):
        session_messages.append(
            UserMessage(
                role="user",
                transcript=message.content,
                created_at=message.timestamp,
            )
        )

    @assistant_aggregator.event_handler("on_assistant_turn_stopped")
    async def on_assistant_turn_stopped(aggregator, message: AssistantTurnStoppedMessage):
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
