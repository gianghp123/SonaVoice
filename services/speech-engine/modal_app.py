import modal
import os

APP_NAME = "sona-speech-engine"
MODAL_SECRET_NAME = "sona-voice-runtime"

image = (
    modal.Image.debian_slim(python_version="3.12")
    .apt_install("ffmpeg")
    .pip_install_from_requirements("requirements.txt")
    .add_local_dir("src", remote_path="/root/src")
    .add_local_file("main.py", remote_path="/root/main.py")
)

app = modal.App(APP_NAME, secrets=[modal.Secret.from_name(MODAL_SECRET_NAME)])

@app.function(image=image, min_containers=0, scaledown_window=60)
@modal.concurrent(max_inputs=3)
@modal.asgi_app()
def fastapi_app():
    """Create and configure the FastAPI application for Modal deployment.

    This mirrors the local development server (src.runner) but runs inside
    Modal's serverless containers. WebRTC connections are handled concurrently
    via SmallWebRTC transport.
    """
    import sys

    # Ensure /root is on the path so ``src`` and ``main`` are importable.
    if "/root" not in sys.path:
        sys.path.insert(0, "/root")

    import main as bot_module
    from src.runner import create_app

    web_app = create_app(bot_module=bot_module)
    return web_app
