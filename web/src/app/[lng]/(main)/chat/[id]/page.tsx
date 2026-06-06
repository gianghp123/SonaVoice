import { getSession } from "@/features/session-history/services/session.get"
import { ChatPageClient } from "@/features/chat-interface/components/ChatPageClient"
import { SessionErrorModal } from "@/features/chat-interface/components/SessionErrorModal"

export default async function ChatPage({ params }: {
  params: Promise<{ lng: string; id: string }>
}) {
  const { id } = await params

  const sessionRes = await getSession(id)

  if (sessionRes.error) {
    return <SessionErrorModal message="Failed to load session" />
  }

  const session = sessionRes.data

  if (!session) {
    return <SessionErrorModal message="Session not found" />
  }

  if (session.status === "inactive" || session.status === "failed") {
    return <SessionErrorModal message="This session is not accessible or does not exist" />
  }

  if (session.status === "active") {
    return <SessionErrorModal message="Session ended, please start a new one" />
  }

  return <ChatPageClient sessionId={id} />
}
