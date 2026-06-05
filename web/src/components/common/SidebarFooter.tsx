"use client"

import { Button } from "@/components/ui/button"
import { useT } from "next-i18next/client"
import { Show, SignInButton } from "@clerk/nextjs"
import { UserButtonWithProfile } from "./UserButtonWithProfile"
import type { IUserProfile } from "@/lib/types/user-profile.interface"

interface SidebarFooterUIProps {
  profile?: IUserProfile | null
}

export function SidebarFooterUI({ profile }: SidebarFooterUIProps) {
  const { t } = useT('common')
  return (
    <div className="px-4 py-5">
      <Show when="signed-in">
        <UserButtonWithProfile profile={profile ?? null} />
      </Show>

      <Show when="signed-out">
        <div className="space-y-5">
          <div className="space-y-3">
            <h3 className="text-[15px] font-semibold leading-none text-foreground">
              {t('get_responses_tailored')}
            </h3>
            <p className="text-[14px] leading-5 text-muted-foreground">
              {t('login_description')}
            </p>
          </div>
          <SignInButton>
            <Button className="w-full rounded-full border-border">
              {t('login')}
            </Button>
          </SignInButton>
        </div>
      </Show>
    </div>
  )
}