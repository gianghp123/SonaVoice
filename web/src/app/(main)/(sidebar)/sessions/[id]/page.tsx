import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { SessionMessageList } from "@/features/session-history/components/SessionMessageList"
import { getMessages } from "@/features/session-history/services/messages.get"
import { getGrammarAnalyses } from "@/features/shared/grammar/services/grammar.get"
import type { SessionItem } from "@/features/session-history/types"
import { getT } from "next-i18next/server"

export default async function SessionPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params
  const { t } = await getT("session")

  const [messagesRes, analysesRes] = await Promise.all([
    getMessages(id),
    getGrammarAnalyses(id),
  ])

  if (messagesRes.error) {
    return (
      <div className="flex flex-1 items-center justify-center p-8">
        <Card className="max-w-md w-full">
          <CardHeader>
            <CardTitle>{t("error")}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-muted-foreground">
              {messagesRes.error.message || t("failed_to_load_messages")}
            </p>
          </CardContent>
        </Card>
      </div>
    )
  }

  const analysisMap = new Map(
    (analysesRes.data ?? []).map((a) => [a.messageId, a])
  )

  const items: SessionItem[] = []
  for (const msg of messagesRes.data ?? []) {
    items.push({ type: "message", data: msg })
    const analysis = analysisMap.get(msg.id)
    if (analysis) {
      items.push({ type: "analysis", data: analysis })
    }
  }

  return (
    <div className="flex flex-col flex-1 justify-center items-center py-10">
      <SessionMessageList items={items} />
    </div>
  )
}
