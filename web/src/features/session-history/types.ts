import type { IMessage } from "@/lib/types/message.interface"
import type { IGrammarAnalysis } from "@/lib/types/grammar-analysis.interface"

export type SessionItem =
  | { type: "message"; data: IMessage }
  | { type: "analysis"; data: IGrammarAnalysis }
