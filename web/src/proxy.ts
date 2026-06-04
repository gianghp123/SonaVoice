import { clerkMiddleware, createRouteMatcher } from "@clerk/nextjs/server"
import { createProxy } from "next-i18next/proxy"
import { NextResponse } from "next/server"
import i18nConfig from "../i18n.config"
import { PAGE_ROUTES } from "./lib/routes"

const i18nProxy = createProxy(i18nConfig)
const isProtectedRoute = createRouteMatcher(["/(.*)/sessions(.*)", "/(.*)/chat(.*)"])

export default clerkMiddleware(async (auth, req) => {
  const { isAuthenticated } = await auth()

  if (!isAuthenticated && isProtectedRoute(req)) {
    return NextResponse.redirect(new URL(PAGE_ROUTES.HOME, req.url))
  }
  return i18nProxy(req)
})

export const config = {
  matcher: [
    "/((?!_next|[^?]*\\.(?:html?|css|js(?!on)|jpe?g|webp|png|gif|svg|ttf|woff2?|ico|csv|docx?|xlsx?|zip|webmanifest)).*)",
    "/(api|trpc)(.*)",
    "/__clerk/(.*)",
  ],
}