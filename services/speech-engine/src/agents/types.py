from typing import TypedDict, Literal, Union


class UserMessage(TypedDict):
    role: Literal["user"]
    timestamp: str
    content: str


class AssistantMessage(TypedDict):
    role: Literal["assistant"]
    content: str
    timestamp: str
    is_interrupted: bool = False


SessionMessage = Union[UserMessage, AssistantMessage]