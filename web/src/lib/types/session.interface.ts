export interface ISession {
  id: string
  userId: string
  status: "pending" | "active" | "inactive" | "failed"
  maxDuration: number
  actualUsage: number
  createdAt: string
  endedAt: string | null
}