"use client"

import { Button } from "@/components/ui/button"
import { useT } from "next-i18next/client"
import { PipecatClientMicToggle } from "@pipecat-ai/client-react"
import { ConnectButton } from "@pipecat-ai/voice-ui-kit"
import { Mic, MicOff } from "lucide-react"

export function VoiceToolbar({ handleDisconnect }: { handleDisconnect: () => void | Promise<void> }) {
  const { t } = useT('chat')
  return (
    <div className="flex lg:mb-10">
      <PipecatClientMicToggle>
        {({ disabled, isMicEnabled, onClick }) => (
          <Button variant="ghost" disabled={disabled} onClick={onClick}>
            {isMicEnabled ? <MicOff /> : <Mic />}
            {isMicEnabled ? t('mic_muted') : t('unmute')}
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
