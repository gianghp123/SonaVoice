from pipecat.audio.vad.silero import SileroVADAnalyzer
from pipecat.pipeline.pipeline import Pipeline
from pipecat.pipeline.task import PipelineParams, PipelineTask
from pipecat.frames.frames import LLMRunFrame, LLMMessagesAppendFrame, EndTaskFrame
from pipecat.processors.aggregators.llm_context import LLMContext
from pipecat.processors.aggregators.llm_response_universal import (
    LLMContextAggregatorPair,
    LLMUserAggregatorParams,
    LLMAssistantAggregatorParams,
    UserTurnStoppedMessage,
    AssistantTurnStoppedMessage,
)
from pipecat.utils.context.llm_context_summarization import (
    LLMAutoContextSummarizationConfig,
    LLMContextSummaryConfig,
)
from loguru import logger
from pipecat.frames.frames import ErrorFrame
from src.utils.error_msg import get_custom_error_message
import asyncio
from pipecat.processors.frame_processor import FrameDirection
from src.types.messages import SessionMessage, UserMessage, AssistantMessage
from pipecat.frames.frames import LLMSummarizeContextFrame
import time
from src.services.session_service import SessionService
from src.pipelines.processors import CustomMem0Processor

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
    memory_processor: CustomMem0Processor = None,
) -> PipelineTask:
    
    if start_time is None:
        start_time = time.time()
    
    session_messages: list[SessionMessage] = []
    service = SessionService()


    user_aggregator, assistant_aggregator = LLMContextAggregatorPair(
        context,
        user_params=LLMUserAggregatorParams(vad_analyzer=SileroVADAnalyzer()),
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

    pipeline = None

    # 2. Define Pipeline
    if memory_processor is None:
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
        pipeline = Pipeline(
            [
                transport.input(),
                stt,
                user_aggregator,
                memory_processor,  # Your CustomMem0Service
                llm,
                tts,
                transport.output(),
                assistant_aggregator,
            ]
        )

    task = PipelineTask(
        pipeline,
        params=PipelineParams(enable_metrics=True),
        tool_resources={
            "user_id": user_id,
            "session_id": session_id,
            "memory_client": memory_processor.memory_client if memory_processor else None,
        },
    )

    # Messages handler
    @user_aggregator.event_handler("on_user_turn_stopped")
    async def on_user_turn_stopped(aggregator, strategy, message: UserTurnStoppedMessage):
        session_messages.append(UserMessage(role="user", content=message.content))

    @assistant_aggregator.event_handler("on_assistant_turn_stopped")
    async def on_assistant_turn_stopped(aggregator, message: AssistantTurnStoppedMessage):
        if message.content:
            session_messages.append(AssistantMessage(role="assistant", content=message.content, is_interrupted=False, timestamp=message.timestamp))
        if message.interrupted:
            session_messages.append(AssistantMessage(role="assistant", content=message.content, is_interrupted=True, timestamp=message.timestamp))

    @user_aggregator.event_handler("on_user_turn_idle")
    async def on_user_turn_idle(aggregator):
        msg = {
            "role": "developer",
            "content": "The user is quiet. Ask if they are there.",
        }
        await aggregator.push_frame(LLMMessagesAppendFrame([msg], run_llm=True))



    # Disconnect handler
    @transport.event_handler("on_client_disconnected")
    async def on_client_disconnected(transport, client):
        logger.info("Client disconnected - cancelling pipeline")
        actual_usage = int(time.time() - start_time)

        await service.close_session(
            session_id=session_id,
            actual_usage=actual_usage,
        )
        
        await task.cancel()
        


    # Session timer
    async def session_timer(task, aggregator, timeout_secs=300):
        await asyncio.sleep(timeout_secs)
        # Trigger LLM to say goodbye
        await aggregator.push_frame(
            LLMMessagesAppendFrame(
                messages=[{"role": "system", "content": "Say goodbye to the user since the session is over."}],
                run_llm=True
            )
        )
        # EndFrame is queued, so the goodbye speech will complete before shutdown
        await aggregator.push_frame(EndTaskFrame(), FrameDirection.UPSTREAM)
        actual_duration = time.time() - start_time
        logger.info(f"Session duration: {actual_duration} seconds")
        

   

    @transport.event_handler("on_client_connected")
    async def on_client_connected(transport, client):
        logger.info(f"Client connected")
        # Add a greeting message to the context
        context.add_message(
            {"role": "system", "content": "Say hello and briefly introduce yourself."}
        )
        # Prompt the bot to start talking when the client connects
        await task.queue_frames([LLMRunFrame()])
        if max_duration is not None:
            asyncio.create_task(session_timer(task, user_aggregator, timeout_secs=max_duration))
            
        
    
    
    # Finish handler

    @task.event_handler("on_pipeline_finished")
    async def on_pipeline_finished(task, frame):
        await task.queue_frames([
            LLMSummarizeContextFrame(
                config=LLMContextSummaryConfig(
                    target_context_tokens=4000,
                    min_messages_after_summary=0,
                )
            )
        ])

        

    # Error handlers
    @tts.event_handler("on_connection_error")
    async def on_tts_connection_error(service, error):
        logger.error(f"TTS connection error: {error}")


    @stt.event_handler("on_connection_error")
    async def on_stt_connection_error(service, error):
        logger.error(f"STT connection error: {error}")


    @llm.event_handler("on_connection_error")
    async def on_llm_connection_error(service, error):
        logger.error(f"LLM connection error: {error}")


    @task.event_handler("on_pipeline_error")
    async def on_pipeline_error(task, frame):
        logger.error(f"Pipeline error: {frame.error}, fatal: {frame.fatal}")

        error_msg = str(frame.error)

        if "Cartesia" in error_msg or "TTS" in error_msg or "WebSocket" in error_msg:
            service_name = "TTS service"
        elif "STT" in error_msg:
            service_name = "STT service"
        elif "LLM" in error_msg:
            service_name = "LLM service"
        else:
            service_name = "Service"

        custom_msg = get_custom_error_message(error_msg, service_name)

        await task.queue_frames([
            ErrorFrame(error=custom_msg, fatal=True)
        ])
    return task
