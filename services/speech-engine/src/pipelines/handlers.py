import asyncio
import time

from loguru import logger
from pipecat.frames.frames import (
    LLMRunFrame,
    EndTaskFrame,
    ErrorFrame,
)
from pipecat.processors.aggregators.llm_context import LLMContext
from pipecat.processors.aggregators.llm_response_universal import (
    UserTurnStoppedMessage,
    AssistantTurnStoppedMessage,
)
from pipecat.processors.frame_processor import FrameDirection

from src.types.messages import SessionMessage, UserMessage, AssistantMessage
from src.utils.error_msg import get_custom_error_message
from src.services.session_service import SessionService
from src.services.messages_service import MessageService

def merge_consecutive_user_messages(
    messages: list[SessionMessage],
) -> list[SessionMessage]:
    merged: list[SessionMessage] = []

    for message in messages:
        if (
            merged
            and merged[-1]["role"] == "user"
            and message["role"] == "user"
        ):
            prev = merged[-1]

            merged[-1] = {
                "role": "user",
                "transcript": f'{prev["transcript"]} {message["transcript"]}'.strip(),
                "created_at": prev["created_at"],
            }
        else:
            merged.append(message)

    return merged

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
    timer_task = None

    log = logger.bind(
        area="voice-task",
        user_id=user_id,
        session_id=session_id,
        max_duration=max_duration,
        has_memory=has_memory,
    )
    
    @task.event_handler("on_idle_timeout")
    async def on_idle_timeout(task):
        await task.queue_frame(ErrorFrame(error="408: Request Timeout", fatal=True))

    @user_aggregator.event_handler("on_user_turn_stopped")
    async def on_user_turn_stopped(aggregator, strategy, message: UserTurnStoppedMessage):
        if message.content:
            session_messages.append(
                UserMessage(
                    role="user",
                    transcript=message.content,
                    created_at=message.timestamp,
                )
            )

    @assistant_aggregator.event_handler("on_assistant_turn_stopped")
    async def on_assistant_turn_stopped(aggregator, message: AssistantTurnStoppedMessage):
        if message.content:
            session_messages.append(
                AssistantMessage(
                    role="assistant",
                    transcript=message.content,
                    was_interrupted=message.interrupted,
                    created_at=message.timestamp,
                )
            )

    # @user_aggregator.event_handler("on_user_turn_idle")
    # async def on_user_turn_idle(aggregator):
    #     if task.has_finished():
    #         return
    #     log.info("User idle")
    #     msg = {
    #         "role": "system",
    #         "content": "The user is quiet. Ask if they are there.",
    #     }
    #     await aggregator.push_frame(LLMMessagesAppendFrame([msg], run_llm=True))
        

    @transport.event_handler("on_client_connected")
    async def on_client_connected(transport, client):
        nonlocal timer_task
        log.info("Client connected")
        context.add_message(
            {
                "role": "system",
                "content": "Say hello and briefly introduce yourself.",
            }
        )
        await task.queue_frames([LLMRunFrame()])

        if max_duration is not None:
            timer_task = asyncio.create_task(
                session_timer(
                    task,
                    user_aggregator,
                    timeout_secs=max_duration,
                )
            )

    @transport.event_handler("on_client_disconnected")
    async def on_client_disconnected(transport, client):
        log.info("Client disconnected")
        nonlocal timer_task
        if timer_task:
            timer_task.cancel()
            
        await task.cancel()
        

    async def session_timer(task, aggregator, timeout_secs=300):
        await asyncio.sleep(timeout_secs)
        
        if task.has_finished():
            return
                
        actual_duration = time.time() - start_time
        log.info(
            "Session timer finished",
            actual_duration=actual_duration,
        )
    
        log.info(
            "Session max duration reached",
            timeout_secs=timeout_secs,
        )
        await aggregator.push_frame(EndTaskFrame(), FrameDirection.UPSTREAM)

    @task.event_handler("on_pipeline_finished")
    async def on_pipeline_finished(task, frame):
        actual_usage = int(time.time() - start_time)
        
        merged_messages = merge_consecutive_user_messages(session_messages)

        log.info(
            "Pipeline finished",
            actual_usage=actual_usage,
            message_count=len(session_messages),
        )

        try:
            if merged_messages:
                await message_service.save_messages(session_id, merged_messages)

                log.info(
                    "Session messages saved",
                    actual_usage=actual_usage,
                    message_count=len(merged_messages),
                )
            else:
                log.info(
                    "Session closed with no messages to save",
                    actual_usage=actual_usage,
                )

        except Exception:
            log.exception(
                "Failed to save session messages",
                actual_usage=actual_usage,
                message_count=len(merged_messages),
            )

        finally:
            try:
                await session_service.finalize_session(
                    session_id=session_id,
                    actual_usage=actual_usage,
                )

                log.info(
                    "Session closed",
                    actual_usage=actual_usage,
                )

            except Exception:
                log.exception(
                    "Failed to close session",
                    actual_usage=actual_usage,
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
        if not frame.fatal:
            await task.queue_frames([ErrorFrame(error=custom_msg, fatal=True)])
        else:
            log.error("Fatal error already in progress", error=error_msg)
