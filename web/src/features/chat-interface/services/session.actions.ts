'use server'

import { revalidateTag } from "next/cache"
import { apiFetch } from "@/lib/api-fetch"
import { tags } from "@/lib/tags"
import { ICreateSessionRes } from "../dtos/create-session.dto.res"

export async function createSession() {
  const result = await apiFetch<ICreateSessionRes>("/sessions", {
    method: "POST",
    withCredentials: true
  })

  if (!result.error) {
    revalidateTag(tags.sessions, "max")
  }

  return result
}

export async function cancelSession(sessionId: string) {
  const result = await apiFetch<void>(`/sessions/${sessionId}/cancel`, {
    method: "POST",
    withCredentials: true
  })

  if (!result.error) {
    revalidateTag(tags.sessions, "max")
  }

  return result
}