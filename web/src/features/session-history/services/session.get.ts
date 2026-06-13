import "server-only"

import { apiFetch } from "@/lib/api-fetch"
import { API_ROUTES } from "@/lib/routes"
import { tags } from "@/lib/tags"
import type { GetSessionsDto } from "../dtos/get-sessions.dto"
import type { ISession } from "@/lib/types/session.interface"

export async function getSessions(dto: GetSessionsDto = {}) {
  return apiFetch<ISession[]>(API_ROUTES.SESSIONS.LIST, {
    withCredentials: true,
    query: { ...dto },
    next: { tags: [tags.sessions] },
  })
}

export async function getSession(sessionId: string) {
  return apiFetch<ISession>(API_ROUTES.SESSIONS.BY_ID(sessionId), {
    withCredentials: true,
    next: { tags: [tags.sessions] },
  })
}
