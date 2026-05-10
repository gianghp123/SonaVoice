import { ScrollArea } from "@/components/ui/scroll-area"
import { HistoryHeader } from "./HistoryHeader"
import { MessageBubble } from "./MessageBubble"
import { AnalysisCard } from "./AnalysisCard"
import { MessageRole } from "@/lib/enums/message-role.enum"

export function HistoryPanelContent() {
  return (
    <>
      <HistoryHeader />
      <ScrollArea className="flex-1">
        <div className="flex flex-col gap-6 p-6">
          <MessageBubble role={MessageRole.Assistant}>
            How are your language goals progressing today? You mentioned
            yesterday that you were visiting a specific city.
          </MessageBubble>

          <MessageBubble role={MessageRole.User}>
            Yesterday I go to Kyoto for the conference.
          </MessageBubble>

          <MessageBubble role={MessageRole.Analysis}>
            <AnalysisCard
              suggestions={{
                hint: "It sounds better as:",
                original: "Yesterday I",
                corrected: "went",
              }}
              pronunciation={{
                word: "Kyoto",
                phonetic: "/kiˈoʊtoʊ/",
              }}
            />
          </MessageBubble>

          <MessageBubble role={MessageRole.Assistant}>
            That makes sense. Kyoto is beautiful this time of year! Was the
            conference productive?
          </MessageBubble>

          <MessageBubble role={MessageRole.User}>
            I ate some really good ramen in a small shop.
          </MessageBubble>

          <MessageBubble role={MessageRole.Analysis}>
            <AnalysisCard
              suggestions={{
                hint: "Try adding more detail:",
                original: '"I ate some',
                corrected: "delicious",
              }}
              pronunciation={{
                word: "Ramen",
                phonetic: "/ˈrɑːmən/",
              }}
            />
          </MessageBubble>

          <MessageBubble role={MessageRole.Assistant}>
            That sounds delicious! Was it a tonkotsu or miso base?
          </MessageBubble>
        </div>
      </ScrollArea>
    </>
  )
}
