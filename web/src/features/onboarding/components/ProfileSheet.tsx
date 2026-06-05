"use client"

import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet"
import { OnboardingForm } from "./OnboardingForm"
import type { IUserProfile } from "@/lib/types/user-profile.interface"
import type { OnboardingFormValues } from "../schemas/onboarding.schema"

interface ProfileSheetProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  profile: IUserProfile | null
}

function mapProfileToFormValues(profile: IUserProfile): Partial<OnboardingFormValues> {
  const prefs = profile.preferences || {}
  return {
    displayName: profile.displayName,
    englishLevel: profile.englishLevel as OnboardingFormValues["englishLevel"],
    nativeLanguage: prefs.nativeLanguage,
    improvementGoals: prefs.improvementGoals,
    topics: prefs.topics,
    customTopics: prefs.customTopics,
    learningReason: prefs.learningReason,
    customLearningReason: prefs.customLearningReason,
  }
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
          <OnboardingForm
            defaultValues={profile ? mapProfileToFormValues(profile) : undefined}
            onSuccess={() => onOpenChange(false)}
          />
        </div>
      </SheetContent>
    </Sheet>
  )
}
