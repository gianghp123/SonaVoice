import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { ConnectNow } from "@/features/landing/components/ConnectNow"
import { Show, SignInButton } from "@clerk/nextjs"

export default function HomePage() {
  return (
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
  )
}
