"use client"

import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet"
import { EditProfileForm } from "@/features/profile/components/EditProfileForm"
import type { IUserProfile } from "@/lib/types/user-profile.interface"

interface ProfileSheetProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  profile: IUserProfile | null
}

export function ProfileSheet({ open, onOpenChange, profile }: ProfileSheetProps) {
  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="overflow-y-auto">
        <SheetHeader>
          <SheetTitle>Edit Profile</SheetTitle>
          <SheetDescription>
            Update your learning preferences.
          </SheetDescription>
        </SheetHeader>
        <div className="px-4 pb-4">
          <EditProfileForm
            profile={profile}
            onSuccess={() => onOpenChange(false)}
          />
        </div>
      </SheetContent>
    </Sheet>
  )
}
