import { SessionStatus } from "@/lib/enums/session-status.enum"

export interface ISession {
  id: string
  userId: string
  status: SessionStatus
  maxDuration: number
  actualUsage: number
  createdAt: string
  endedAt: string | null
}