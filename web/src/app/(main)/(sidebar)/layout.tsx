import { HomePageLayout } from "@/components/common/HomePageLayout"
import { SidebarFooterUI } from "@/components/common/SidebarFooter"
import { getProfile } from "@/features/profile/services/profile.get"
import { getSessions } from "@/features/session-history/services/session.get"
import { auth } from "@clerk/nextjs/server"

export default async function SidebarLayout({
  children,
  breadcrumb,
}: {
  children: React.ReactNode
  breadcrumb: React.ReactNode
}) {
  const { sessionClaims, isAuthenticated } = await auth()

  const onboardingCompleted = isAuthenticated && sessionClaims?.metadata?.onboardingCompleted

  const [sessions, profile] = onboardingCompleted
    ? await Promise.all([
      getSessions().then((res) => res.data ?? []),
      getProfile().then((res) => res.data ?? null),
    ])
    : [[], null]

  return (
    <HomePageLayout
      sessions={sessions}
      sidebarFooter={<SidebarFooterUI profile={profile} />}
      breadcrumb={breadcrumb}
    >
      {children}
    </HomePageLayout>
  )
}
