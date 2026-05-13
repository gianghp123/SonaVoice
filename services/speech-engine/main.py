from pipecat.pipeline.runner import PipelineRunner
from pipecat.services.openai.llm import OpenAILLMService
from pipecat.services.deepgram.stt import DeepgramSTTService
from pipecat.services.cartesia.tts import CartesiaTTSService
from src.core.config import settings
from src.pipelines.processors import CustomMem0Processor
from src.transports.daily import create_daily_transport
from src.transports.small_webrtc import create_small_webrtc_transport
from src.pipelines.tasks import create_voice_bot_task
from mem0 import AsyncMemory
from src.core.app_resources import AppResources
from src.agents.prompts import ENGLIST_TEACHER_SYSTEM_INSTRUCTION
from pipecat.runner.types import RunnerArguments
from pipecat.runner.types import DailyRunnerArguments, RunnerArguments, SmallWebRTCRunnerArguments
from pipecat.transports.smallwebrtc.transport import TransportParams, SmallWebRTCConnection
from pipecat.transports.base_transport import BaseTransport
from loguru import logger
from pipecat.transports.daily.transport import DailyParams
from dotenv import load_dotenv
from src.agents.tools import summarize_conversation, save_user_preferences, tools
from pipecat.processors.aggregators.llm_context import LLMContext

load_dotenv()

async def bot(runner_args: RunnerArguments):
    transport = None

    if isinstance(runner_args, DailyRunnerArguments):
        transport = create_daily_transport(
            room_url=runner_args.room_url,
            token=runner_args.token,
            bot=settings.BOT_NAME,
            params=DailyParams(
                audio_in_enabled=True,
                audio_out_enabled=True,
                transcription_enabled=True,
            ),
        )
    elif isinstance(runner_args, SmallWebRTCRunnerArguments):
        logger.info(f"Starting the bot, received body: {runner_args.body}")
        webrtc_connection: SmallWebRTCConnection = runner_args.webrtc_connection
        transport = create_small_webrtc_transport(
            webrtc_connection,
            params=TransportParams(
                audio_in_enabled=True,
                audio_out_enabled=True,
            )
        )
    
    else:
        logger.error(f"Unsupported runner arguments type: {type(runner_args)}")
        return

    if transport is None:
        logger.error("Failed to create transport")
        return

    await run_bot(transport, runner_args)

async def run_bot(transport: BaseTransport, runner_args: RunnerArguments):
    logger.info(f"Starting the bot, received body: {runner_args.body}") 
    user_id = "anonymous"
    session_id = "anonymous"
    if runner_args.body:
        logger.debug("No body received")
        user_id = runner_args.body.get("user_id")
        session_id = runner_args.body.get("session_id")
    else:
        logger.debug("No body received")
        
    stt = DeepgramSTTService(api_key=settings.DEEPGRAM_API_KEY)
    
    llm = OpenAILLMService(
        api_key=settings.OPENCODE_API_KEY,
        base_url=settings.OPENAI_BASE_URL,
        settings=OpenAILLMService.Settings(
            model=settings.LLM_NAME,
            system_instruction=ENGLIST_TEACHER_SYSTEM_INSTRUCTION,
            extra={
                "extra_body": {
                    "thinking": {"type": "disabled"}
                }
            }
        ),
    )
    
    llm.register_function("summarize_conversation", summarize_conversation)
    llm.register_function("save_user_preferences", save_user_preferences)

    tts = CartesiaTTSService(
        api_key=settings.CARTESIA_API_KEY,
        settings=CartesiaTTSService.Settings(
            voice=settings.CARTESIA_VOICE_ID, # Sophie - Female - British English
        ),
    )
    
    context = LLMContext(tools=tools)
    
    memory_client = AsyncMemory.from_config(settings.memory_config)
    
    app_resource = AppResources(memory_client=memory_client, user_id=user_id, session_id=session_id)
    
    # Initialize Memory Processor
    memory_processor = CustomMem0Processor(
        memory_client=memory_client,
        user_id=user_id,
        session_id=session_id
    )

    # Create the Task using our factory
    task = await create_voice_bot_task(transport, stt, llm, tts, memory_processor, context, app_resources=app_resource)

    # Run
    runner = PipelineRunner(handle_sigint=False)
    await runner.run(task)
    
    
if __name__ == "__main__":
    from pipecat.runner.run import main
    main()