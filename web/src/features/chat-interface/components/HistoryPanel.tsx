import { ScrollArea } from "@/components/ui/scroll-area"
import { HistoryHeader } from "./HistoryHeader"
import { MessageBubble } from "./MessageBubble"
import { AnalysisCard } from "./AnalysisCard"

interface HistoryPanelProps {
  onClose: () => void
}

export function HistoryPanel({ onClose }: HistoryPanelProps) {
  return (
    <section className="flex w-[40%] flex-col border-l border-border bg-accent">
      <HistoryHeader onClose={onClose} />
      <ScrollArea className="flex-1 h-full md:pb-20">
        <div className="flex flex-col gap-6 p-6">
          <MessageBubble role="sona">
            How are your language goals progressing today? You mentioned
            yesterday that you were visiting a specific city.
          </MessageBubble>

          <MessageBubble role="user">
            Yesterday I go to Kyoto for the conference.
          </MessageBubble>

          <MessageBubble role="analysis">
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

          <MessageBubble role="sona">
            That makes sense. Kyoto is beautiful this time of year! Was the
            conference productive?
          </MessageBubble>

          <MessageBubble role="user">
            I ate some really good ramen in a small shop.
          </MessageBubble>

          <MessageBubble role="analysis">
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

          <MessageBubble role="sona">
            That sounds delicious! Was it a tonkotsu or miso base?
          </MessageBubble>
        </div>
      </ScrollArea>
    </section>
  )
}
