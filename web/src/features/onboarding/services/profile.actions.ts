"use server"

import { apiFetch } from "@/lib/api-fetch"
import { API_ROUTES } from "@/lib/routes"
import { tags } from "@/lib/tags"
import { updateTag } from "next/cache"
import type { IUserProfile, IUpsertProfileDto } from "@/lib/types/user-profile.interface"

export async function upsertProfile(data: IUpsertProfileDto) {
  const result = await apiFetch<IUserProfile>(API_ROUTES.PROFILE.UPSERT, {
    method: "POST",
    withCredentials: true,
    body: JSON.stringify(data),
  })

  if (!result.error) {
    updateTag(tags.profile)
  }

  return result
}
