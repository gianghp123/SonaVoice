import { ChatPageClient } from "@/features/chat-interface/components/ChatPageClient"
import { SessionErrorModal } from "@/features/chat-interface/components/SessionErrorModal"
import { getSession } from "@/features/session-history/services/session.get"
import { SessionStatus } from "@/lib/enums/session-status.enum"
import { getT } from "next-i18next/server"

export default async function ChatPage({ params }: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params

  const sessionRes = await getSession(id)
  const { t } = await getT("session")

  if (sessionRes.error) {
    return <SessionErrorModal message={t("failed_to_load_session")} />
  }

  const session = sessionRes.data

  if (!session) {
    return <SessionErrorModal message={t("session_not_found")} />
  }

  if (session.status === SessionStatus.Inactive || session.status === SessionStatus.Failed) {
    return <SessionErrorModal message={t("session_not_accessible")} />
  }

  if (session.status === SessionStatus.Active) {
    return <SessionErrorModal message={t("session_ended_start_new")} />
  }


  return <ChatPageClient sessionId={id} />
}
