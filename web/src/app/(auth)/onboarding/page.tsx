import { OnboardingForm } from "@/features/onboarding/components/OnboardingForm"
import { PAGE_ROUTES } from "@/lib/routes"
import { auth } from "@clerk/nextjs/server"
import { redirect } from "next/navigation"

export default async function OnboardingPage() {
  const { userId, sessionClaims, } = await auth()

  if (!userId) {
    redirect("/")
  }

  const metadata = sessionClaims?.metadata

  const onboardingCompleted =
    metadata?.onboardingCompleted === true

  if (onboardingCompleted) {
    redirect(PAGE_ROUTES.HOME)
  }

  return <OnboardingForm/>
}