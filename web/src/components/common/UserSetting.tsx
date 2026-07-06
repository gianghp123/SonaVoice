import { getProfile } from "@/features/profile/services/profile.get"
import { SidebarFooterUI } from "./SidebarFooter"

export async function UserSetting() {
  const res = await getProfile()
  const profile = res.data ?? null

  return <SidebarFooterUI profile={profile} />
}
