import { MicOff, PhoneOff } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"

export function VoiceToolbar() {
  return (
    <div className="flex rounded-md border border-secondary bg-secondary shadow-sm overflow-hidden">
      <Button variant="ghost" className="text-primary rounded-none">
        <MicOff />
        Mic Muted
      </Button>
      <Separator orientation="vertical" className="bg-primary/20" />
      <Button variant="ghost" className="text-primary rounded-none">
        <PhoneOff />
        End Session
      </Button>
    </div>
  )
}
