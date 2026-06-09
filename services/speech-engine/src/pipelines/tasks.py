import time

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
        start_time = time.time()

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
            audio_in_sample_rate=16000,
            audio_out_sample_rate=22050,
        ),
        cancel_on_idle_timeout=False,
        idle_timeout_secs=30, 
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
