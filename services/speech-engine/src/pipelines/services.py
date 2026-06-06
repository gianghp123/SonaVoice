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

import json
import re


async def clean_for_tts(text: str, type: str) -> str:
    if not text:
        return ""

    # Remove code blocks
    text = re.sub(r"```[\s\S]*?```", " code omitted ", text)

    # Remove inline code
    text = re.sub(r"`([^`]*)`", r"\1", text)

    # Markdown links: [text](url) -> text
    text = re.sub(r"\[([^\]]+)\]\((.*?)\)", r"\1", text)

    # Remove raw URLs
    text = re.sub(r"https?://\S+", "", text)

    # Remove markdown formatting
    text = re.sub(r"[*_~>#|]", "", text)

    # Speak-friendly replacements
    replacements = {
        "&": " and ",
        "@": " at ",
        "%": " percent ",
        "$": " dollars ",
    }
    for old, new in replacements.items():
        text = text.replace(old, new)

    # Remove excessive punctuation
    text = re.sub(r"[!?]{2,}", ".", text)
    text = re.sub(r"\.{3,}", ".", text)

    # Normalize whitespace
    text = re.sub(r"\s+", " ", text)

    return text.strip()


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


def build_llm(user_profile: dict | None = None) -> OpenAILLMService:
    system_instruction = ENGLIST_TEACHER_SYSTEM_INSTRUCTION.format(
        name=settings.BOT_NAME,
        user_profile="",
    )
    if user_profile:
        profile_block = json.dumps(user_profile, indent=2, ensure_ascii=False)
        system_instruction = ENGLIST_TEACHER_SYSTEM_INSTRUCTION.format(
            name=settings.BOT_NAME,
            user_profile=f"USER PROFILE:\n{profile_block}",
        )

    llm = OpenAILLMService(
        api_key=settings.OPENCODE_API_KEY,
        base_url=settings.OPENAI_BASE_URL,
        metrics=_get_metrics(),
        settings=OpenAILLMService.Settings(
            model=settings.LLM_NAME,
            system_instruction=system_instruction,
            extra={"extra_body": {"thinking": {"type": "disabled"}}},
        ),
    )
    llm.register_function("summarize_conversation", summarize_conversation)
    return llm


def build_tts() -> PiperTTSService:
    tts = PiperTTSService(
        metrics=_get_metrics(),
        settings=PiperTTSService.Settings(
            voice="en_US-lessac-high",
        ),
        download_dir=settings.piper_models_dir,
    )
    
    tts.add_text_transformer(clean_for_tts, "*")
    return tts

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
