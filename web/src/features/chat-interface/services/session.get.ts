import "server-only"

import { apiFetch } from "@/lib/api-fetch"
import { tags } from "@/lib/tags"
import type { GetSessionsDto } from "../dtos/get-sessions.dto"
import type { ISession } from "@/lib/types/session.interface"

export async function getSessions(dto: GetSessionsDto = {}) {
  return apiFetch<ISession[]>("/sessions", {
    withCredentials: true,
    query: dto,
    next: { tags: [tags.sessions] },
  })
}
