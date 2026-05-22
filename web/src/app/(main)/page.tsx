import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { ConnectNow } from "@/features/landing/components/ConnectNow"
import { getSessions } from "@/features/chat-interface/services/session.get"
import { HomePageLayout } from "@/features/landing/components/HomePageLayout"
import { Show, SignInButton } from "@clerk/nextjs"

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
    <HomePageLayout sessions={sessions}>
      <div className="flex flex-1 items-center justify-center">
        <Card className="max-w-md w-full">
          <CardHeader>
            <CardTitle>Sona Voice</CardTitle>
            <CardDescription>
              Your AI voice companion. Connect to start a conversation.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Show when="signed-in">
              <ConnectNow />
            </Show>
            <Show when="signed-out">
              <p>Please sign in to start a conversation.</p>
              <SignInButton />
            </Show>
          </CardContent>
        </Card>
      </div>
    </HomePageLayout>
  )
}
