from pipecat.frames.frames import LLMSummarizeContextFrame
from pipecat.services.llm_service import FunctionCallParams
from pipecat.adapters.schemas.function_schema import FunctionSchema


async def summarize_conversation(params: FunctionCallParams):
    """Trigger manual context summarization via a pipeline frame."""
    await params.result_callback({"status": "summarization_requested"})
    await params.llm.queue_frame(LLMSummarizeContextFrame())

summarize_function = FunctionSchema(
    name="summarize_conversation",
    description=(
        "Summarize and compress the conversation history. "
        "Call this only when the user asks you to summarize the conversation."
    ),
    properties={},
    required=[],
)
