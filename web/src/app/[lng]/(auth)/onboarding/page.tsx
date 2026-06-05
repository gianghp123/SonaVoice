import { OnboardingForm } from "@/features/onboarding/components/OnboardingForm"
import { getProfile } from "@/features/onboarding/services/profile.get"
import { PAGE_ROUTES } from "@/lib/routes"
import { auth } from "@clerk/nextjs/server"
import { redirect } from "next/navigation"

export default async function OnboardingPage() {
  const { userId } = await auth()

  if (!userId) {
    redirect("/")
  }

  const profile = await getProfile()

  if (profile.data?.displayName) {
    redirect(PAGE_ROUTES.HOME)
  }

  return (
      <OnboardingForm />
  )
}
