from typing import Any

import aiohttp

from src.core.config import settings
from src.types.messages import SessionMessage


class MessageServiceError(Exception):
    pass


class MessageService:
    def __init__(self) -> None:
        self.base_url = settings.API_URL.rstrip("/")

    async def save_messages(
        self,
        session_id: str,
        messages: list[SessionMessage],
    ) -> dict[str, Any]:
        url = f"{self.base_url}/sessions/{session_id}/messages"

        payload = {
            "messages": messages,
        }

        try:
            async with aiohttp.ClientSession() as session:
                async with session.post(url, json=payload, timeout=10) as response:
                    response.raise_for_status()

                    if response.content_length == 0:
                        return {}

                    return await response.json()

        except aiohttp.ClientError as exc:
            raise MessageServiceError(
                f"Failed to save messages for session {session_id}"
            ) from exc