from pipecat.transports.daily.transport import DailyParams, DailyTransport
from loguru import logger
from src.core.config import settings

def create_daily_transport(room_url: str, token: str, bot: str, params: DailyParams = None) -> DailyTransport:
    transport = DailyTransport(
        room_url,
        token,
        bot,
        params=params
    )

    return transport