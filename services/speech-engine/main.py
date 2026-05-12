from pipecat.pipeline.runner import PipelineRunner
from pipecat.services.openai.llm import OpenAILLMService
from pipecat.services.deepgram.stt import DeepgramSTTService
from pipecat.services.cartesia.tts import CartesiaTTSService
from src.core.config import settings
from src.pipelines.processors import CustomMem0Processor
from src.transports.daily import create_daily_transport
from src.pipelines.tasks import create_voice_bot_task
from mem0 import AsyncMemory
from src.core.app_resources import AppResources
from src.agents.prompts import ENGLIST_TEACHER_SYSTEM_INSTRUCTION
from pipecat.runner.types import RunnerArguments
from pipecat.runner.types import DailyRunnerArguments, RunnerArguments, SmallWebRTCRunnerArguments
from loguru import logger
from pipecat.transports.daily.transport import DailyTransport
from dotenv import load_dotenv
load_dotenv()

async def bot(runner_args: RunnerArguments):
    transport = None

    if isinstance(runner_args, DailyRunnerArguments):
        user_id = runner_args.body.get("user_id")
        session_id = runner_args.body.get("session_id")
        transport = create_daily_transport(room_url=runner_args.room_url, token=runner_args.token, bot=settings.BOT_NAME)
        
    else:
        logger.error(f"Unsupported runner arguments type: {type(runner_args)}")
        return

    if transport is None:
        logger.error("Failed to create transport")
        return

    await run_bot(transport, user_id=user_id, session_id=session_id)

async def run_bot(transport: DailyTransport, user_id: str, session_id: str):
    stt = DeepgramSTTService(api_key=settings.DEEPGRAM_API_KEY)
    llm = OpenAILLMService(
        api_key=settings.OPENCODE_API_KEY,
        base_url=settings.OPENAI_BASE_URL,
        settings=OpenAILLMService.Settings(
            model=settings.LLM_NAME,
            system_instruction=ENGLIST_TEACHER_SYSTEM_INSTRUCTION,
        ),
    )

    tts = CartesiaTTSService(
        api_key=settings.CARTESIA_API_KEY,
        settings=CartesiaTTSService.Settings(
            voice=settings.CARTESIA_VOICE_ID, # Sophie - Female - British English
        ),
    )
    
    memory_client = AsyncMemory.from_config(settings.memory_config)
    
    app_resource = AppResources(memory_client=memory_client, user_id=user_id, session_id=session_id)
    
    # Initialize Memory Processor
    memory_processor = CustomMem0Processor(
        memory_client=memory_client,
        user_id=user_id,
        session_id=session_id
    )

    # Create the Task using our factory
    task = await create_voice_bot_task(transport, stt, llm, tts, memory_processor, app_resource)

    # Run
    runner = PipelineRunner(handle_sigint=True)
    await runner.run(task)
    
    
if __name__ == "__main__":
    from pipecat.runner.run import main
    main()