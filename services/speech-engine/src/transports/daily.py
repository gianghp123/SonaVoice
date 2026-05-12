from pipecat.transports.daily.transport import DailyParams, DailyTransport
from loguru import logger
from src.core.config import settings

def create_daily_transport(room_url: str, token: str, bot: str) -> DailyTransport:
    transport = DailyTransport(
        room_url,
        token,
        bot,
        DailyParams(
            audio_in_enabled=True,
            audio_out_enabled=True,
            transcription_enabled=True,
        ),
    )

    @transport.event_handler("on_joined")
    async def on_joined(transport, data):
        logger.info(f"Bot joined the room: {data}")

    @transport.event_handler("on_left")
    async def on_left(transport):
        logger.info("Bot left the room")

    @transport.event_handler("on_error")
    async def on_error(transport, error):
        logger.error(f"Transport error: {error}")

    return transport