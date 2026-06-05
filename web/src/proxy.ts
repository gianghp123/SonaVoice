import { clerkMiddleware } from "@clerk/nextjs/server"
import type { NextRequest } from "next/server"
import { NextResponse } from "next/server"

import { FALLBACK_LANGUAGE } from "./lib/i18n/i18n"
import { PAGE_ROUTES } from "./lib/routes"
import { getLocaleFromPathname, stripLocalePrefix } from "./lib/utils/path"

const defaultLocale = FALLBACK_LANGUAGE

export function getLocaleFromReferer(req: NextRequest): string | undefined {
  const referer = req.headers.get("referer")

  if (!referer) return undefined

  try {
    const url = new URL(referer)
    return getLocaleFromPathname(url.pathname)
  } catch {
    return undefined
  }
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
  const { isAuthenticated, sessionClaims } = await auth()
  const pathname = req.nextUrl.pathname
  const pathnameWithoutLocale = stripLocalePrefix(pathname)

  const pathnameLocale = getLocaleFromPathname(pathname)
  const refererLocale = getLocaleFromReferer(req)

  const locale =
    pathnameLocale ??
    refererLocale ??
    defaultLocale

  const canonicalPathname = withPublicLocale(locale, pathnameWithoutLocale)

  // 1. Normalize locale URL first
  if (req.nextUrl.pathname !== canonicalPathname) {
    return redirectTo(req, canonicalPathname)
  }

  // 2. Auth/onboarding guard
  if (isAuthenticated) {
    const publicMetadata = sessionClaims?.metadata

    const onboardingCompleted =
      publicMetadata?.onboardingCompleted === true

    const isOnboardingPath =
      pathnameWithoutLocale === "/onboarding" ||
      pathnameWithoutLocale.startsWith("/onboarding/")

    if (!onboardingCompleted && !isOnboardingPath) {
      return redirectTo(req, withPublicLocale(locale, "/onboarding"))
    }
  }

  if (!isAuthenticated && isProtectedPath(pathnameWithoutLocale)) {
    return redirectTo(req, withPublicLocale(locale, PAGE_ROUTES.HOME))
  }

  // 3. Internal rewrite
  const internalPathname = `/${locale}${pathnameWithoutLocale === "/" ? "" : pathnameWithoutLocale
    }`

  if (req.nextUrl.pathname === internalPathname) {
    return NextResponse.next()
  }

  return rewriteTo(req, internalPathname)
})
export const config = {
  matcher: [
    "/((?!_next|[^?]*\\.(?:html?|css|js(?!on)|jpe?g|webp|png|gif|svg|ttf|woff2?|ico|csv|docx?|xlsx?|zip|webmanifest)).*)",
    "/(api|trpc)(.*)",
  ],
}