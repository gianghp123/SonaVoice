#
# Copyright (c) 2024-2026, Daily
#
# SPDX-License-Identifier: BSD 2-Clause License
#

"""Pipecat development runner.

This development runner executes Pipecat bots and provides the supporting
infrastructure they need for WebRTC connections.

Install with::

    pip install pipecat-ai[runner]

All bots must implement a ``bot(runner_args)`` async function as the entry point.
The server automatically discovers and executes this function when connections
are established.

Example::

    async def bot(runner_args: SmallWebRTCRunnerArguments):
        transport = SmallWebRTCTransport(
            webrtc_connection=runner_args.webrtc_connection,
            params=SmallWebRTCParams(
                audio_enabled=True,
                video_enabled=True,
            ),
        )
        await run_pipeline(transport)

    if __name__ == "__main__":
        from pipecat.runner.run import main
        main()

To run locally::

    python bot.py
"""

import argparse
import json
import mimetypes
import os
import sys
import uuid
from contextlib import asynccontextmanager
from http import HTTPMethod
from pathlib import Path
from typing import Any, TypedDict

from fastapi.responses import FileResponse, Response
from loguru import logger

from pipecat.runner.types import SmallWebRTCRunnerArguments

from src.core.config import settings
from pipecat.transports.smallwebrtc.connection import IceServer as PipecatIceServer

DEFAULT_ICE_SERVERS: list[PipecatIceServer] = [
    PipecatIceServer(urls=["stun:stun.l.google.com:19302"]),
]

def build_ice_servers() -> list[PipecatIceServer]:
    raw = settings.ICE_SERVERS.strip()
    if not raw:
        return DEFAULT_ICE_SERVERS
    try:
        configs = json.loads(raw)
        return [
            PipecatIceServer(
                urls=c["urls"],
                username=c.get("username"),
                credential=c.get("credential"),
            )
            for c in configs
        ]
    except (json.JSONDecodeError, KeyError) as e:
        logger.warning(f"Invalid ICE_SERVERS config, using defaults: {e}")
        return DEFAULT_ICE_SERVERS

try:
    import uvicorn
    from dotenv import load_dotenv
    from fastapi import BackgroundTasks, FastAPI, HTTPException, Request
    from fastapi.middleware.cors import CORSMiddleware
    from fastapi.responses import RedirectResponse
except ImportError as e:
    logger.error(f"Runner dependencies not available: {e}")
    logger.error("To use Pipecat runners, install with: pip install pipecat-ai[runner]")
    raise ImportError(
        "Runner dependencies required. Install with: pip install pipecat-ai[runner]"
    ) from e


load_dotenv(override=True)
os.environ["ENV"] = "local"

RUNNER_DOWNLOADS_FOLDER: str | None = None
RUNNER_HOST: str = "localhost"
RUNNER_PORT: int = 7860

app: FastAPI = FastAPI()
"""The FastAPI application instance.

Import this to add custom routes from other packages before calling
:func:`main`::

    from pipecat.runner.run import app, main

    @app.get("/my-route")
    async def my_route():
        return {"hello": "world"}

    if __name__ == "__main__":
        main()
"""


def _get_bot_module():
    """Get the bot module from the calling script."""
    import importlib.util

    # Get the main module (the file that was executed)
    main_module = sys.modules["__main__"]

    # Check if it has a bot function
    if hasattr(main_module, "bot"):
        return main_module

    # Try to import 'bot' module from current directory
    try:
        import bot  # type: ignore[import-untyped]

        return bot
    except ImportError:
        pass

    # Look for any .py file in current directory that has a bot function
    # (excluding server.py).
    cwd = os.getcwd()
    for filename in os.listdir(cwd):
        if filename.endswith(".py") and filename != "server.py":
            try:
                module_name = filename[:-3]  # Remove .py extension
                spec = importlib.util.spec_from_file_location(
                    module_name, os.path.join(cwd, filename)
                )
                if spec is None or spec.loader is None:
                    continue
                module = importlib.util.module_from_spec(spec)
                spec.loader.exec_module(module)

                if hasattr(module, "bot"):
                    return module
            except Exception:
                continue

    raise ImportError(
        "Could not find 'bot' function. Make sure your bot file has a 'bot' function."
    )


def _setup_webrtc_routes(app: FastAPI, args: argparse.Namespace):
    """Set up WebRTC-specific routes."""
    try:
        from pipecat_ai_small_webrtc_prebuilt.frontend import SmallWebRTCPrebuiltUI

        from pipecat.transports.smallwebrtc.connection import SmallWebRTCConnection
        from pipecat.transports.smallwebrtc.request_handler import (
            IceCandidate,
            SmallWebRTCPatchRequest,
            SmallWebRTCRequest,
            SmallWebRTCRequestHandler,
        )
    except ImportError as e:
        logger.error(f"WebRTC transport dependencies not installed: {e}")
        return

    class IceServer(TypedDict, total=False):
        urls: str | list[str]

    class IceConfig(TypedDict):
        iceServers: list[IceServer]

    class StartBotResult(TypedDict, total=False):
        sessionId: str
        iceConfig: IceConfig | None

    # In-memory store of active sessions: session_id -> session info
    active_sessions: dict[str, dict[str, Any]] = {}

    # Mount the frontend
    app.mount("/client", SmallWebRTCPrebuiltUI)

    @app.get("/", include_in_schema=False)
    async def root_redirect():
        """Redirect root requests to client interface."""
        return RedirectResponse(url="/client/")

    @app.get("/files/{filename:path}")
    async def download_file(filename: str):
        """Handle file downloads."""
        if not args.folder:
            logger.warning(f"Attempting to dowload {filename}, but downloads folder not setup.")
            return

        file_path = Path(args.folder) / filename
        if not os.path.exists(file_path):
            raise HTTPException(404)

        media_type, _ = mimetypes.guess_type(file_path)

        return FileResponse(path=file_path, media_type=media_type, filename=filename)

    # Initialize the SmallWebRTC request handler
    small_webrtc_handler: SmallWebRTCRequestHandler = SmallWebRTCRequestHandler(
        ice_servers=build_ice_servers(), esp32_mode=args.esp32, host=args.host
    )

    @app.post("/api/offer")
    async def offer(request: SmallWebRTCRequest, background_tasks: BackgroundTasks):
        """Handle WebRTC offer requests via SmallWebRTCRequestHandler."""

        # Prepare runner arguments with the callback to run your bot
        async def webrtc_connection_callback(connection: SmallWebRTCConnection):
            bot_module = _get_bot_module()

            runner_args = SmallWebRTCRunnerArguments(
                webrtc_connection=connection, body=request.request_data
            )
            runner_args.cli_args = args
            background_tasks.add_task(bot_module.bot, runner_args)

        # Delegate handling to SmallWebRTCRequestHandler
        answer = await small_webrtc_handler.handle_web_request(
            request=request,
            webrtc_connection_callback=webrtc_connection_callback,
        )
        return answer

    @app.patch("/api/offer")
    async def ice_candidate(request: SmallWebRTCPatchRequest):
        """Handle WebRTC new ice candidate requests."""
        logger.debug(f"Received patch request: {request}")
        await small_webrtc_handler.handle_patch_request(request)
        return {"status": "success"}

    @app.post("/start")
    async def rtvi_start(request: Request):
        """Mimic Pipecat Cloud's /start endpoint."""
        # Parse the request body
        try:
            request_data = await request.json()
            logger.debug(f"Received request: {request_data}")
        except Exception as e:
            logger.error(f"Failed to parse request body: {e}")
            request_data = {}

        # Store session info immediately in memory, replicate the behavior expected on Pipecat Cloud
        session_id = str(uuid.uuid4())
        active_sessions[session_id] = request_data.get("body", {})

        ice_srvs = build_ice_servers()
        result: StartBotResult = {"sessionId": session_id}
        if ice_srvs:
            ice_config_servers = []
            for s in ice_srvs:
                server: dict[str, Any] = {"urls": s.urls}
                if getattr(s, "username", None):
                    server["username"] = s.username
                if getattr(s, "credential", None):
                    server["credential"] = s.credential
                ice_config_servers.append(server)
            result["iceConfig"] = IceConfig(iceServers=ice_config_servers)

        return result

    @app.api_route(
        "/sessions/{session_id}/{path:path}",
        methods=["GET", "POST", "PUT", "PATCH", "DELETE"],
    )
    async def proxy_request(
        session_id: str, path: str, request: Request, background_tasks: BackgroundTasks
    ):
        """Mimic Pipecat Cloud's proxy."""
        active_session = active_sessions.get(session_id)
        if active_session is None:
            return Response(content="Invalid or not-yet-ready session_id", status_code=404)

        if path.endswith("api/offer"):
            # Parse the request body and convert to SmallWebRTCRequest
            try:
                request_data = await request.json()
                if request.method == HTTPMethod.POST.value:
                    webrtc_request = SmallWebRTCRequest(
                        sdp=request_data["sdp"],
                        type=request_data["type"],
                        pc_id=request_data.get("pc_id"),
                        restart_pc=request_data.get("restart_pc"),
                        request_data=request_data.get("request_data")
                        or request_data.get("requestData")
                        or active_session,
                    )
                    return await offer(webrtc_request, background_tasks)
                elif request.method == HTTPMethod.PATCH.value:
                    patch_request = SmallWebRTCPatchRequest(
                        pc_id=request_data["pc_id"],
                        candidates=[IceCandidate(**c) for c in request_data.get("candidates", [])],
                    )
                    return await ice_candidate(patch_request)
            except Exception as e:
                logger.error(f"Failed to parse WebRTC request: {e}")
                return Response(content="Invalid WebRTC request", status_code=400)

        logger.info(f"Received request for path: {path}")
        return Response(status_code=200)

    @asynccontextmanager
    async def smallwebrtc_lifespan(app: FastAPI):
        """Manage FastAPI application lifecycle and cleanup connections."""
        yield
        await small_webrtc_handler.close()

    # Add the SmallWebRTC lifespan to the app
    _add_lifespan_to_app(app, smallwebrtc_lifespan)


def _add_lifespan_to_app(app: FastAPI, new_lifespan):
    """Add a new lifespan context manager to the app, combining with existing if present.

    Args:
        app: The FastAPI application instance
        new_lifespan: The new lifespan context manager to add
    """
    if hasattr(app.router, "lifespan_context") and app.router.lifespan_context is not None:
        # If there's already a lifespan context, combine them
        existing_lifespan = app.router.lifespan_context

        @asynccontextmanager
        async def combined_lifespan(app: FastAPI):
            async with existing_lifespan(app):
                async with new_lifespan(app):
                    yield

        app.router.lifespan_context = combined_lifespan
    else:
        # No existing lifespan, use the new one
        app.router.lifespan_context = new_lifespan


def runner_downloads_folder() -> str | None:
    """Returns the folder where files are stored for later download."""
    return RUNNER_DOWNLOADS_FOLDER


def runner_host() -> str:
    """Returns the host name of this runner."""
    return RUNNER_HOST


def runner_port() -> int:
    """Returns the port of this runner."""
    return RUNNER_PORT


def main(parser: argparse.ArgumentParser | None = None):
    """Start the Pipecat development runner.

    Parses command-line arguments and starts a FastAPI server configured
    for WebRTC transport.

    The runner discovers and runs any ``bot(runner_args)`` function found in the
    calling module.

    Command-line arguments:
       - --host: Server host address (default: localhost)
       - --port: Server port (default: 7860)
       - -v/--verbose: Increase logging verbosity

    Args:
        parser: Optional custom argument parser. If provided, default runner
            arguments are added to it so bots can define their own CLI
            arguments. Custom arguments should not conflict with the default
            ones. Custom args are accessible via ``runner_args.cli_args``.

    """
    global RUNNER_DOWNLOADS_FOLDER, RUNNER_HOST, RUNNER_PORT

    if not parser:
        parser = argparse.ArgumentParser(description="Pipecat Development Runner")
    parser.add_argument("--host", type=str, default=RUNNER_HOST, help="Host address")
    parser.add_argument("--port", type=int, default=RUNNER_PORT, help="Port number")
    parser.add_argument(
        "-v", "--verbose", action="count", default=0, help="Increase logging verbosity"
    )

    args = parser.parse_args()

    # Set defaults for args used by _setup_webrtc_routes
    args.folder = getattr(args, "folder", None)
    args.esp32 = getattr(args, "esp32", False)

    # Log level
    logger.remove()
    logger.add(sys.stderr, level="TRACE" if args.verbose else "DEBUG")

    RUNNER_DOWNLOADS_FOLDER = args.folder
    RUNNER_HOST = args.host
    RUNNER_PORT = args.port

    # Add CORS middleware and set up WebRTC routes
    app.add_middleware(
        CORSMiddleware,
        allow_origins=["*"],
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )
    _setup_webrtc_routes(app, args)

    # Run the server
    uvicorn.run(app, host=args.host, port=args.port)


if __name__ == "__main__":
    main()
