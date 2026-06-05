import { getProfile } from "@/features/onboarding/services/profile.get"
import { SidebarFooterUI } from "./SidebarFooter"

export async function SidebarFooter() {
  const res = await getProfile()
  return <SidebarFooterUI profile={res.data ?? null} />
}
