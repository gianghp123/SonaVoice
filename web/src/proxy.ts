import { clerkMiddleware, createRouteMatcher } from "@clerk/nextjs/server"
import { match as matchLocale } from "@formatjs/intl-localematcher"
import Negotiator from "negotiator"
import { NextResponse } from "next/server"
import type { NextRequest } from "next/server"
import { SUPPORTED_LANGUAGES, FALLBACK_LANGUAGE } from "./lib/i18n"
import { PAGE_ROUTES } from "./lib/routes"

const locales: string[] = [...SUPPORTED_LANGUAGES]
const defaultLocale = FALLBACK_LANGUAGE

function getLocale(request: NextRequest): string {
  const headers: Record<string, string> = {}
  request.headers.forEach((value, key) => {
    headers[key] = value
  })
  const languages = new Negotiator({ headers }).languages(locales)
  return matchLocale(languages, locales, defaultLocale)
}

function hasLocalePrefix(pathname: string): boolean {
  return locales.some(
    (locale) => pathname.startsWith(`/${locale}/`) || pathname === `/${locale}`
  )
}

const isProtectedRoute = createRouteMatcher(["/(.*)/sessions(.*)", "/(.*)/chat(.*)"])

export default clerkMiddleware(async (auth, req) => {
  const { isAuthenticated } = await auth()

  if (!isAuthenticated && isProtectedRoute(req)) {
    return NextResponse.redirect(new URL(PAGE_ROUTES.HOME, req.url))
  }

  const pathname = req.nextUrl.pathname

  if (hasLocalePrefix(pathname)) {
    return NextResponse.next()
  }

  const locale = getLocale(req)
  const newUrl = new URL(`/${locale}${pathname}`, req.url)
  return NextResponse.redirect(newUrl)
})

export const config = {
  matcher: [
    "/((?!_next|[^?]*\\.(?:html?|css|js(?!on)|jpe?g|webp|png|gif|svg|ttf|woff2?|ico|csv|docx?|xlsx?|zip|webmanifest)).*)",
    "/(api|trpc)(.*)",
    "/__clerk/(.*)",
  ],
}
