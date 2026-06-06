import os
from typing import Any

import aiohttp

from src.core.config import settings


class SessionServiceError(Exception):
    pass


class SessionService:
    def __init__(self) -> None:
        self.base_url = settings.API_URL.rstrip("/")

    async def finalize_session(self, session_id: str, actual_usage: int) -> dict[str, Any]:
        url = f"{self.base_url}/sessions/{session_id}/finalize"

        payload = {
            "session_id": session_id,
            "actual_usage": actual_usage,
        }

        headers = {"X-Internal-Secret": os.getenv("INTERNAL_SECRET", "")}

        try:
            async with aiohttp.ClientSession() as session:
                async with session.post(url, json=payload, headers=headers, timeout=10) as response:
                    response.raise_for_status()

                    if response.content_length == 0:
                        return {}

                    return await response.json()

        except aiohttp.ClientError as exc:
            raise SessionServiceError(
                f"Failed to close session {session_id}"
            ) from exc