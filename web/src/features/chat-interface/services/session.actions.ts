'use server'

import { apiFetch } from "@/lib/api-fetch"
import { ISession } from "@/lib/types/session.interface"

export async function createSession() {
  return apiFetch<ISession>("/model-gateway/sessions", {
    method: "POST",
    withCredentials: true
  })
}