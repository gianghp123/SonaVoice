"use client"

import { MicOff, Mic, PhoneOff } from "lucide-react"
import { PipecatClientMicToggle } from "@pipecat-ai/client-react"
import { ConnectButton } from "@pipecat-ai/voice-ui-kit"
import { Button } from "@/components/ui/button"

export function VoiceToolbar({ handleDisconnect }: { handleDisconnect: () => void | Promise<void> }) {
  return (
    <div className="flex lg:mb-10">
      <PipecatClientMicToggle>
        {({ disabled, isMicEnabled, onClick }) => (
          <Button variant="ghost" disabled={disabled} onClick={onClick}>
            {isMicEnabled ? <MicOff /> : <Mic />}
            {isMicEnabled ? "Mic Muted" : "Unmute"}
          </Button>
        )}
      </PipecatClientMicToggle>
      <ConnectButton
        className="hover:bg-muted px-2.5 text-destructive"
        onDisconnect={handleDisconnect}
      />
    </div>
  )
}
