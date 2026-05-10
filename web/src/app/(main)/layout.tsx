import { PageLayout } from "@/components/layouts/PageLayout"

export default function MainLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return <PageLayout>{children}</PageLayout>
}
