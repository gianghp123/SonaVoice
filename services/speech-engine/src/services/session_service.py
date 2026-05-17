import os
from typing import Any

import aiohttp


class SessionServiceError(Exception):
    pass


class SessionService:
    def __init__(self) -> None:
        self.base_url = os.getenv("SESSION_SERVICE_URL")
        if not self.base_url:
            raise SessionServiceError("SESSION_SERVICE_URL is not set")

        self.base_url = self.base_url.rstrip("/")

    async def close_session(self, session_id: str, actual_usage: int) -> dict[str, Any]:
        url = f"{self.base_url}/model-gateway/sessions/{session_id}/close"

        payload = {
            "session_id": session_id,
            "actual_usage": actual_usage,
        }

        try:
            async with aiohttp.ClientSession() as session:
                async with session.post(url, json=payload, timeout=10) as response:
                    response.raise_for_status()

                    if response.content_length == 0:
                        return {}

                    return await response.json()

        except aiohttp.ClientError as exc:
            raise SessionServiceError(
                f"Failed to close session {session_id}"
            ) from exc