import { clerkMiddleware, createRouteMatcher } from "@clerk/nextjs/server"
import { NextResponse } from "next/server"
import { PAGE_ROUTES } from "./lib/routes"

const isProtectedRoute = createRouteMatcher(["/sessions(.*)", "/chat(.*)"])

export default clerkMiddleware(async (auth, req) => {
  const { isAuthenticated } = await auth()

  if (!isAuthenticated && isProtectedRoute(req)) {
    return NextResponse.redirect(new URL(PAGE_ROUTES.HOME, req.url))
  }
})

export const config = {
  matcher: [
    "/((?!_next|[^?]*\\.(?:html?|css|js(?!on)|jpe?g|webp|png|gif|svg|ttf|woff2?|ico|csv|docx?|xlsx?|zip|webmanifest)).*)",
    "/(api|trpc)(.*)",
    "/__clerk/(.*)",
  ],
}