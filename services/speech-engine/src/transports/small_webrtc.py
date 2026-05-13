from pipecat.transports.smallwebrtc.transport import SmallWebRTCTransport
from pipecat.transports.smallwebrtc.connection import SmallWebRTCConnection
from pipecat.transports.base_transport import TransportParams
from loguru import logger

def create_small_webrtc_transport(webrtc_connection: SmallWebRTCConnection, params: TransportParams = None) -> SmallWebRTCTransport:
    transport = SmallWebRTCTransport(
        webrtc_connection=webrtc_connection,
        params=params
    )

    # @transport.event_handler("on_joined")
    # async def on_joined(transport, data):
    #     logger.info(f"Bot joined the room: {data}")

    # @transport.event_handler("on_left")
    # async def on_left(transport):
    #     logger.info("Bot left the room")

    # @transport.event_handler("on_error")
    # async def on_error(transport, error):
    #     logger.error(f"Transport error: {error}")

    return transport