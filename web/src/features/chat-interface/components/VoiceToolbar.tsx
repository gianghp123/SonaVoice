import { MicOff, PhoneOff } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"

export function VoiceToolbar() {
  return (
    <div className="flex">
      <Button variant="secondary" className="rounded-l-full gap-0">
        <MicOff />
        Mic Muted
      </Button>
      <Button variant="secondary" className="rounded-r-full">
        <PhoneOff className="text-destructive"/>
        End Session
      </Button>
    </div>
  )
}
