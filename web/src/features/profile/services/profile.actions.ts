"use server"

import { apiFetch } from "@/lib/api-fetch"
import { API_ROUTES } from "@/lib/routes"
import { tags } from "@/lib/tags"
import { updateTag } from "next/cache"
import type { IUpsertProfileDto } from "@/lib/types/user-profile.interface"

export async function updateProfile(data: IUpsertProfileDto) {
  const result = await apiFetch<boolean>(API_ROUTES.PROFILE.UPDATE, {
    method: "PATCH",
    withCredentials: true,
    body: data,
  })

  if (!result.error) {
    updateTag(tags.profile)
  }

  return result
}
