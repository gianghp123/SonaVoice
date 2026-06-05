import { OnboardingForm } from "@/features/onboarding/components/OnboardingForm"
import { auth } from "@clerk/nextjs/server"
import { redirect } from "next/navigation"
import { getProfile } from "@/features/onboarding/services/profile.get"
import { PAGE_ROUTES } from "@/lib/routes"

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
    <div className="w-full max-w-lg mx-auto px-4 py-8">
      <div className="mb-8 text-center">
        <h1 className="text-2xl font-bold">Welcome to Sona Voice</h1>
        <p className="text-muted-foreground mt-2">
          Tell us about yourself so we can personalize your experience.
        </p>
      </div>
      <OnboardingForm/>
    </div>
  )
}
