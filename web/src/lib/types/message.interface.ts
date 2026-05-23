import { MessageRole } from "../enums/message-role.enum"

export interface IMessage {
	id: string
	sessionId: string
	role: MessageRole
	transcript: string
	wasInterrupted: boolean
  createdAt: string
}