'use server'

import { apiFetch } from "@/lib/api-fetch"
import { ICreateSessionRes } from "../dtos/create-session.dto.res"

export async function createSession() {
  return apiFetch<ICreateSessionRes>("/model-gateway/start", {
    method: "POST",
    withCredentials: true
  })
}