from pipecat.transports.smallwebrtc.transport import SmallWebRTCTransport
from pipecat.transports.smallwebrtc.connection import SmallWebRTCConnection
from pipecat.transports.base_transport import TransportParams
from loguru import logger

def create_small_webrtc_transport(webrtc_connection: SmallWebRTCConnection, params: TransportParams = None) -> SmallWebRTCTransport:
    transport = SmallWebRTCTransport(
        webrtc_connection=webrtc_connection,
        params=params
    )

    return transport