from typing import TypedDict, Literal, Union


class UserMessage(TypedDict):
    role: Literal["user"]
    created_at: str
    transcript: str


class AssistantMessage(TypedDict):
    role: Literal["assistant"]
    transcript: str
    created_at: str
    was_interrupted: bool = False


SessionMessage = Union[UserMessage, AssistantMessage]