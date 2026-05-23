import { HomePageLayout } from "@/components/common/HomePageLayout"
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { getMessages } from "@/features/chat-interface/services/messages.get"
import { getSessions } from "@/features/chat-interface/services/session.get"
import { SessionMessageList } from "@/features/session-history/components/SessionMessageList"
import { PAGE_ROUTES } from "@/lib/routes"
import Link from "next/link"

export default async function SessionPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params
  const [sessionRes, messagesRes] = await Promise.all([
    getSessions(),
    getMessages(id),
  ])
  const sessions = sessionRes.data ?? []

  if (messagesRes.error) {
    return (
      <HomePageLayout sessions={sessions}>
        <Breadcrumb className="px-8 pt-4">
          <BreadcrumbList>
            <BreadcrumbItem>
              <BreadcrumbLink asChild>
                <Link href={PAGE_ROUTES.HOME}>Home</Link>
              </BreadcrumbLink>
            </BreadcrumbItem>
            <BreadcrumbSeparator />
            <BreadcrumbItem>
              <BreadcrumbPage>Session</BreadcrumbPage>
            </BreadcrumbItem>
          </BreadcrumbList>
        </Breadcrumb>
        <div className="flex flex-1 items-center justify-center p-8">
          <Card className="max-w-md w-full">
            <CardHeader>
              <CardTitle>Error</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-muted-foreground">
                {messagesRes.error.message || "Failed to load messages"}
              </p>
            </CardContent>
          </Card>
        </div>
      </HomePageLayout>
    )
  }

  return (
    <HomePageLayout sessions={sessions} breadcrumb={
      <Breadcrumb>
        <BreadcrumbList>
          <BreadcrumbItem>
            <BreadcrumbLink asChild>
              <Link href={PAGE_ROUTES.HOME}>Home</Link>
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbPage>Session History</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
    }>

      <div className="flex flex-col flex-1 justify-center items-center overflow-y-auto">
        <SessionMessageList messages={messagesRes.data ?? []} />
      </div>
    </HomePageLayout>
  )
}
