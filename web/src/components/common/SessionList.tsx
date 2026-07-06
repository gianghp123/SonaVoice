import { getSessions } from "@/features/session-history/services/session.get"
import { SessionListClient } from "./SessionListClient"

export async function SessionList() {
  const res = await getSessions()
  const sessions = res.data ?? []

  return <SessionListClient sessions={sessions} />
}
