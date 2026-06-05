import { getSessions } from "@/features/session-history/services/session.get"
import { HomePageLayout } from "@/components/common/HomePageLayout"
import { SidebarFooter } from "@/components/common/SidebarFooterWrapper"
import { LandingHero } from "@/features/landing/components/LandingHero"

export default async function HomePage(props: {
  searchParams: Promise<{ page?: string; limit?: string }>
}) {
  const searchParams = await props.searchParams
  const res = await getSessions({
    page: Number(searchParams.page) || 1,
    limit: Number(searchParams.limit) || 10,
  })
  const sessions = res.data ?? []

  return (
    <HomePageLayout sessions={sessions} sidebarFooter={<SidebarFooter />}>
      <LandingHero />
    </HomePageLayout>
  )
}
