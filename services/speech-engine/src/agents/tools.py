from pipecat.frames.frames import LLMSummarizeContextFrame
from pipecat.services.llm_service import FunctionCallParams
from pipecat.adapters.schemas.function_schema import FunctionSchema
from pipecat.adapters.schemas.tools_schema import ToolsSchema

async def summarize_conversation(params: FunctionCallParams):
    """Trigger manual context summarization via a pipeline frame."""
    await params.result_callback({"status": "summarization_requested"})
    await params.llm.queue_frame(LLMSummarizeContextFrame())
    
async def save_user_preferences(params: FunctionCallParams):
    user_id = params.app_resources.user_id
    session_id = params.app_resources.session_id
    memory_client = params.app_resources.memory_client
    messages = params.context.messages
    await memory_client.add(messages, user_id=user_id, run_id=session_id)
    await params.result_callback({"success": True})
    
summarize_function = FunctionSchema(
    name="summarize_conversation",
    description=(
        "Summarize and compress the conversation history. "
        "Call this only when the user asks you to summarize the conversation."
    ),
    properties={},
    required=[],
)

save_memory_function = FunctionSchema(
    name="save_user_preferences",
    description=(
        "Saves meaningful, long-term personal insights about the user to persistent memory. "
        "This helps make future English conversations more natural, relevant, and personalized. "
        "\n\n"
        "Use this for:"
        "- Strong user personal interests, hobbies, or favorites (e.g., 'I love Interstellar and Cyberpunk 2077')"
        "- Important personal background (job, culture, life goals)"
        "- Speaking breakthroughs, strengths, weaknesses, or learning preferences"
        "- Any specific detail that can enrich future speaking practice"
        "\n\n"
        "You should proactively save information when the user naturally shares something "
        "that reveals their personality, tastes, or background — no need for the user to explicitly say 'save this'. "
        "NOTE: no need any input for this tool since it operates automatically"
    ),
    properties={},
    required=[],
)

tools = ToolsSchema([summarize_function, save_memory_function])
