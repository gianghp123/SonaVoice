from pipecat.audio.vad.silero import SileroVADAnalyzer
from pipecat.pipeline.pipeline import Pipeline
from pipecat.pipeline.task import PipelineParams, PipelineTask
from pipecat.frames.frames import LLMRunFrame, LLMMessagesAppendFrame
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
from src.agents.tools import tools
from src.pipelines.processors import FrameProcessor

async def create_voice_bot_task(
    transport, stt, llm, tts, memory_processor: FrameProcessor, app_resources
) -> PipelineTask:
    
    context = LLMContext(tools=tools)
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
        )
    )

    # 2. Define Pipeline
    pipeline = Pipeline([
        transport.input(),
        stt,
        user_aggregator,
        memory_processor,  # Your CustomMem0Service
        llm,
        tts,
        transport.output(),
        assistant_aggregator,
    ])

    task = PipelineTask(pipeline, params=PipelineParams(enable_metrics=True), tool_resources=app_resources)

    # 3. Attach Pipeline/Aggregator Event Handlers
    @user_aggregator.event_handler("on_user_turn_stopped")
    async def on_user_turn_stopped(aggregator, strategy, message):
        print(f"User: {message.content}")

    @user_aggregator.event_handler("on_user_turn_idle")
    async def on_user_turn_idle(aggregator):
        msg = {"role": "developer", "content": "The user is quiet. Ask if they are there."}
        await aggregator.push_frame(LLMMessagesAppendFrame([msg], run_llm=True))

    @transport.event_handler("on_first_participant_joined")
    async def on_first_participant_joined(transport, participant):
        context.add_message({"role": "developer", "content": "Introduce yourself."})
        await task.queue_frames([LLMRunFrame()])

    @transport.event_handler("on_participant_left")
    async def on_participant_left(transport, participant, reason):
         await task.cancel()

    return task