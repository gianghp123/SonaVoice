"use client"

import { MicOff, Mic, PhoneOff } from "lucide-react"
import { PipecatClientMicToggle } from "@pipecat-ai/client-react"
import { ConnectButton } from "@pipecat-ai/voice-ui-kit"
import { Button } from "@/components/ui/button"

export function VoiceToolbar({ handleDisconnect }: { handleDisconnect: () => void | Promise<void> }) {
  return (
    <div className="flex">
      <PipecatClientMicToggle>
        {({ disabled, isMicEnabled, onClick }) => (
          <Button variant="secondary" className="rounded-l-full gap-0" disabled={disabled} onClick={onClick}>
            {isMicEnabled ? <MicOff /> : <Mic />}
            {isMicEnabled ? "Mic Muted" : "Unmute"}
          </Button>
        )}
      </PipecatClientMicToggle>
      <ConnectButton
        className="rounded-r-full"
        defaultVariant="secondary"
        onDisconnect={handleDisconnect}
        stateContent={{
          ready: {
            variant: "secondary",
            children: <><PhoneOff className="text-destructive" /> End Session</>
          },
        }}
      />
    </div>
  )
}
