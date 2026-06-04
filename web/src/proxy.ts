import { clerkMiddleware, createRouteMatcher } from "@clerk/nextjs/server"
import { match as matchLocale } from "@formatjs/intl-localematcher"
import Negotiator from "negotiator"
import type { NextRequest } from "next/server"
import { NextResponse } from "next/server"
import { FALLBACK_LANGUAGE, SUPPORTED_LANGUAGES } from "./lib/i18n/i18n"
import { PAGE_ROUTES } from "./lib/routes"

const locales: string[] = [...SUPPORTED_LANGUAGES]
const defaultLocale = FALLBACK_LANGUAGE
const LOCALE_COOKIE = "NEXT_LOCALE"

function getLocaleFromHeaders(request: NextRequest): string {
  const headers: Record<string, string> = {}
  request.headers.forEach((value, key) => {
    headers[key] = value
  })
  const languages = new Negotiator({ headers }).languages(locales)
  return matchLocale(languages, locales, defaultLocale)
}

function getLocale(request: NextRequest): string {
  const cookieLocale = request.cookies.get(LOCALE_COOKIE)?.value
  if (cookieLocale && locales.includes(cookieLocale)) {
    return cookieLocale
  }
  return getLocaleFromHeaders(request)
}

function hasLocalePrefix(pathname: string): boolean {
  return locales.some(
    (locale) => pathname.startsWith(`/${locale}/`) || pathname === `/${locale}`
  )
}

const isProtectedRoute = createRouteMatcher(["/(.*)/sessions(.*)", "/(.*)/chat(.*)"])

export default clerkMiddleware(async (auth, req) => {
  const { isAuthenticated } = await auth()
  const pathname = req.nextUrl.pathname

  if (hasLocalePrefix(pathname)) {
    if (!isAuthenticated && isProtectedRoute(req)) {
      const locale = getLocale(req)
      return NextResponse.redirect(
        new URL(`/${locale}${PAGE_ROUTES.HOME}`, req.url)
      )
    }
    return NextResponse.next()
  }

  const locale = getLocale(req)
  const url = req.nextUrl.clone()
  url.pathname = `/${locale}${pathname}`
  return NextResponse.rewrite(url)
})

export const config = {
  matcher: [
    "/((?!_next|[^?]*\\.(?:html?|css|js(?!on)|jpe?g|webp|png|gif|svg|ttf|woff2?|ico|csv|docx?|xlsx?|zip|webmanifest)).*)",
    "/(api|trpc)(.*)",
    "/__clerk/(.*)",
  ],
}
