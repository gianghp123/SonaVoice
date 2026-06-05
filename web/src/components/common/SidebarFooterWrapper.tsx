import { getProfile } from "@/features/profile/services/profile.get"
import { SidebarFooterUI } from "./SidebarFooter"
import { auth } from "@clerk/nextjs/server"

export async function SidebarFooter() {
  const { sessionClaims } = await auth()
  const onboardingCompleted = sessionClaims?.metadata?.onboardingCompleted === true

  const profile = onboardingCompleted ? (await getProfile()).data ?? null : null

  return <SidebarFooterUI profile={profile} />
}
