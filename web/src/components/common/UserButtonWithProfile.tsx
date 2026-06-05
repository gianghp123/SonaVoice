"use client"

import { UserButton } from "@clerk/nextjs"
import { useState } from "react"
import { ProfileSheet } from "@/features/profile/components/ProfileSheet"
import type { IUserProfile } from "@/lib/types/user-profile.interface"
import { UserIcon } from "lucide-react"
import { useT } from "next-i18next/client"

interface UserButtonWithProfileProps {
  profile: IUserProfile | null
}

export function UserButtonWithProfile({ profile }: UserButtonWithProfileProps) {
  const [open, setOpen] = useState(false)
  const { t } = useT("common")

  return (
    <>
      <UserButton>
        <UserButton.MenuItems>
          <UserButton.Action
            label={t("edit_profile")}
            labelIcon={<UserIcon className="size-4" />}
            onClick={() => setOpen(true)}
          />
        </UserButton.MenuItems>
      </UserButton>
      <ProfileSheet open={open} onOpenChange={setOpen} profile={profile} />
    </>
  )
}
