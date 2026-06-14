'use server'

import { apiFetch } from "@/lib/api-fetch"
import { API_ROUTES } from "@/lib/routes"
import { tags } from "@/lib/tags"
import { refresh, updateTag } from "next/cache"
import { ICreateSessionRes } from "../dtos/create-session.dto.res"

export async function createSession() {
  const result = await apiFetch<ICreateSessionRes>(API_ROUTES.SESSIONS.LIST, {
    method: "POST",
    withCredentials: true
  })

  if (!result.error) {
    updateTag(tags.sessions)
  }

  return result
}

export async function cancelSession(sessionId: string) {
  const result = await apiFetch<void>(API_ROUTES.SESSIONS.CANCEL(sessionId), {
    method: "POST",
    withCredentials: true
  })

  if (!result.error) {
    updateTag(tags.sessions)
    refresh()
  }

  return result
}