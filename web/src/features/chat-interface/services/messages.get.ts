import "server-only"

import { apiFetch } from "@/lib/api-fetch"
import { API_ROUTES } from "@/lib/routes"
import type { IMessage } from "@/lib/types/message.interface"

export async function getMessages(sessionId: string) {
  return apiFetch<IMessage[]>(API_ROUTES.SESSIONS.MESSAGES(sessionId), {
    withCredentials: true,
    query: { limit: 1000 },
  })
}
