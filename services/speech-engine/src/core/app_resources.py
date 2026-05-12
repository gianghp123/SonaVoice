from dataclasses import dataclass
from mem0 import AsyncMemory

@dataclass
class AppResources:
    user_id: str
    session_id: str
    memory_client: AsyncMemory