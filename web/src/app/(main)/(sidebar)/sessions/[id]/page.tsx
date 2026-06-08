import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { SessionMessageList } from "@/features/session-history/components/SessionMessageList"
import { getMessages } from "@/features/session-history/services/messages.get"
import { getT } from "next-i18next/server"

export default async function SessionPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params
  const { t } = await getT("session")
  const messagesRes = await getMessages(id)

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

  return (
    <div className="flex flex-col flex-1 justify-center items-center overflow-y-auto">
      <SessionMessageList messages={messagesRes.data ?? []} />
    </div>
  )
}
