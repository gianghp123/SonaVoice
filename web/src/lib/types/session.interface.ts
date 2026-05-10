import { IMessage } from "./message.interface"

export interface ISession {
  id: string
  userId: string
  startedAt: Date
  endedAt: Date
  messages?: IMessage[]
  createdAt: Date
}