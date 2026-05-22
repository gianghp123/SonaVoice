from mem0 import AsyncMemory
from pipecat.frames.frames import Frame, LLMContextFrame
from pipecat.processors.frame_processor import FrameDirection, FrameProcessor


class CustomMem0Processor(FrameProcessor):
    def __init__(self, memory_client: AsyncMemory, user_id: str, **kwargs):
        super().__init__(**kwargs)
        self.memory_client = memory_client
        self.user_id = user_id
        self.last_query = None

    async def _retrieve_memories(self, query: str, top_k = 20):
        try:
            results = await self.memory_client.search(
                query=query,
                top_k=top_k,
                filters={
                  "user_id": self.user_id
                },
            )
            return results["results"]
        except Exception as e:
            return []

    async def process_frame(self, frame: Frame, direction: FrameDirection):
        await super().process_frame(frame, direction)

        if isinstance(frame, LLMContextFrame):
            context = frame.context
            context_messages = context.get_messages()
            latest_user_message = None

            for message in reversed(context_messages):
                if message.get("role") == "user" and isinstance(message.get("content"), str):
                    latest_user_message = message.get("content")
                    break

            if latest_user_message and self.last_query != latest_user_message:
                self.last_query = latest_user_message
                memories = await self._retrieve_memories(latest_user_message, top_k=10)
                if memories:
                    memory_text = "Based on previous conversations, I recall:\n\n"
                    for i, mem in enumerate(memories):
                        memory_text += f"{i}. {mem.get('memory', '')}\n\n"

                    memory_message = {"role": "system", "content": memory_text}

                    # Insert at the desired position
                    position = max(0, min(1, len(context_messages)))
                    context_messages.insert(position, memory_message)
                    context.set_messages(context_messages)

        await self.push_frame(frame, direction)