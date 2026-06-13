import { clerkMiddleware } from "@clerk/nextjs/server"
import { NextResponse } from "next/server"

import { PAGE_ROUTES } from "./lib/routes"

function isProtectedPath(pathname: string): boolean {
  return (
    pathname === "/sessions" ||
    pathname.startsWith("/sessions/") ||
    pathname === "/chat" ||
    pathname.startsWith("/chat/")
  )
}

function redirectTo(req: Request, pathname: string) {
  const url = new URL(req.url)
  url.pathname = pathname

  return NextResponse.redirect(url)
}

export default clerkMiddleware(async (auth, req) => {
  const { isAuthenticated, sessionClaims } = await auth()
  const pathname = req.nextUrl.pathname

  // Auth/onboarding guard
  if (isAuthenticated) {
    const publicMetadata = sessionClaims?.metadata

    const onboardingCompleted =
      publicMetadata?.onboardingCompleted === true

    const isOnboardingPath =
      pathname === "/onboarding" ||
      pathname.startsWith("/onboarding/")

    if (!onboardingCompleted && !isOnboardingPath) {
      return redirectTo(req, "/onboarding")
    }
  }

  if (!isAuthenticated && isProtectedPath(pathname)) {
    return redirectTo(req, PAGE_ROUTES.HOME)
  }

  return NextResponse.next()
})

export const config = {
  matcher: [
    "/((?!_next|[^?]*\\.(?:html?|css|js(?!on)|jpe?g|webp|png|gif|svg|ttf|woff2?|ico|csv|docx?|xlsx?|zip|webmanifest)).*)",
    "/(api|trpc)(.*)",
  ],
}