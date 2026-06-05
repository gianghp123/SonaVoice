"use client"

import { UserButton } from "@clerk/nextjs"
import { useState } from "react"
import { ProfileSheet } from "@/features/profile/components/ProfileSheet"
import type { IUserProfile } from "@/lib/types/user-profile.interface"
import { UserIcon } from "lucide-react"

interface UserButtonWithProfileProps {
  profile: IUserProfile | null
}

export function UserButtonWithProfile({ profile }: UserButtonWithProfileProps) {
  const [open, setOpen] = useState(false)

  return (
    <>
      <UserButton>
        <UserButton.MenuItems>
          <UserButton.Action
            label="Edit Profile"
            labelIcon={<UserIcon className="size-4" />}
            onClick={() => setOpen(true)}
          />
        </UserButton.MenuItems>
      </UserButton>
      <ProfileSheet open={open} onOpenChange={setOpen} profile={profile} />
    </>
  )
}
