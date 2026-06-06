import { HomePageLayout } from "@/components/common/HomePageLayout"
import { SidebarFooterUI } from "@/components/common/SidebarFooter"
import { getProfile } from "@/features/profile/services/profile.get"
import { getSessions } from "@/features/session-history/services/session.get"
import { FALLBACK_LANGUAGE, isSupportedLanguage, SupportedLanguage } from "@/lib/i18n/i18n"
import { auth } from "@clerk/nextjs/server"

export default async function SidebarLayout({
  children,
  breadcrumb,
  params,
}: {
  children: React.ReactNode
  breadcrumb: React.ReactNode
  params: Promise<{ lng: string }>
}) {
  const { lng } = await params
  const { sessionClaims, isAuthenticated } = await auth()

  const onboardingCompleted = isAuthenticated && sessionClaims?.metadata?.onboardingCompleted

  const [sessions, profile] = onboardingCompleted
    ? await Promise.all([
      getSessions().then((res) => res.data ?? []),
      getProfile().then((res) => res.data ?? null),
    ])
    : [[], null]

  const currentLanguage: SupportedLanguage = isSupportedLanguage(lng)
    ? lng
    : FALLBACK_LANGUAGE

  return (
    <HomePageLayout
      sessions={sessions}
      sidebarFooter={<SidebarFooterUI profile={profile} />}
      breadcrumb={breadcrumb}
      currentLanguage={currentLanguage}
    >
      {children}
    </HomePageLayout>
  )
}
