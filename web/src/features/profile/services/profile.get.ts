import "server-only"

import { apiFetch } from "@/lib/api-fetch"
import { API_ROUTES } from "@/lib/routes"
import { tags } from "@/lib/tags"
import type { IUserProfile } from "@/lib/types/user-profile.interface"

export async function getProfile() {
  return apiFetch<IUserProfile>(API_ROUTES.PROFILE.GET, {
    withCredentials: true,
    next: { tags: [tags.profile] },
  })
}
