import { clerkMiddleware } from "@clerk/nextjs/server"
import { match as matchLocale } from "@formatjs/intl-localematcher"
import Negotiator from "negotiator"
import type { NextRequest } from "next/server"
import { NextResponse } from "next/server"

import { FALLBACK_LANGUAGE, SUPPORTED_LANGUAGES } from "./lib/i18n/i18n"
import { PAGE_ROUTES } from "./lib/routes"
import { LOCALE_COOKIE } from "./lib/cookies/cookie.contants"
import { stripLocalePrefix } from "./lib/utils/path"

const locales: string[] = [...SUPPORTED_LANGUAGES]
const defaultLocale = FALLBACK_LANGUAGE

function getLocaleFromHeaders(request: NextRequest): string {
  const headers = Object.fromEntries(request.headers.entries())
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


function withPublicLocale(locale: string, pathname: string): string {
  const normalizedPath = pathname.startsWith("/") ? pathname : `/${pathname}`

  if (locale === defaultLocale) {
    return normalizedPath
  }

  return `/${locale}${normalizedPath === "/" ? "" : normalizedPath}`
}

function isProtectedPath(pathname: string): boolean {
  return (
    pathname === "/sessions" ||
    pathname.startsWith("/sessions/") ||
    pathname === "/chat" ||
    pathname.startsWith("/chat/")
  )
}

function redirectTo(req: NextRequest, pathname: string) {
  const url = req.nextUrl.clone()
  url.pathname = pathname

  return NextResponse.redirect(url)
}

function rewriteTo(req: NextRequest, pathname: string) {
  const url = req.nextUrl.clone()
  url.pathname = pathname

  return NextResponse.rewrite(url)
}

export default clerkMiddleware(async (auth, req) => {
  const { isAuthenticated } = await auth()

  const locale = getLocale(req)

  const pathnameWithoutLocale = stripLocalePrefix(req.nextUrl.pathname)
  const canonicalPathname = withPublicLocale(locale, pathnameWithoutLocale)

  if (!isAuthenticated && isProtectedPath(pathnameWithoutLocale)) {
    return redirectTo(req, withPublicLocale(locale, PAGE_ROUTES.HOME))
  }

  if (req.nextUrl.pathname !== canonicalPathname) {
    return redirectTo(req, canonicalPathname)
  }

  return rewriteTo(req, `/${locale}${pathnameWithoutLocale === "/" ? "" : pathnameWithoutLocale}`)
})

export const config = {
  matcher: [
    "/((?!_next|[^?]*\\.(?:html?|css|js(?!on)|jpe?g|webp|png|gif|svg|ttf|woff2?|ico|csv|docx?|xlsx?|zip|webmanifest)).*)",
    "/(api|trpc)(.*)",
  ],
}