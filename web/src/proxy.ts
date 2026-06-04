import { clerkMiddleware } from "@clerk/nextjs/server"
import { match as matchLocale } from "@formatjs/intl-localematcher"
import Negotiator from "negotiator"
import type { NextRequest } from "next/server"
import { NextResponse } from "next/server"
import { FALLBACK_LANGUAGE, SUPPORTED_LANGUAGES } from "./lib/i18n/i18n"
import { PAGE_ROUTES } from "./lib/routes"
import { LOCALE_COOKIE } from "./lib/cookies/cookie.contants"

const locales: string[] = [...SUPPORTED_LANGUAGES]
const defaultLocale = FALLBACK_LANGUAGE

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

function splitLocale(pathname: string): {
  pathLocale: string | null
  pathnameWithoutLocale: string
} {
  const segments = pathname.split("/")
  const maybeLocale = segments[1]

  if (!locales.includes(maybeLocale)) {
    return {
      pathLocale: null,
      pathnameWithoutLocale: pathname,
    }
  }

  const rest = `/${segments.slice(2).join("/")}`

  return {
    pathLocale: maybeLocale,
    pathnameWithoutLocale: rest === "/" ? "/" : rest,
  }
}

function withPublicLocale(locale: string, path: string): string {
  const normalizedPath = path.startsWith("/") ? path : `/${path}`

  if (locale === FALLBACK_LANGUAGE) {
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

export default clerkMiddleware(async (auth, req) => {
  const { isAuthenticated } = await auth()

  const pathname = req.nextUrl.pathname
  const preferredLocale = getLocale(req)

  const { pathLocale, pathnameWithoutLocale } = splitLocale(pathname)

  const effectiveLocale = pathLocale ?? preferredLocale

  if (!isAuthenticated && isProtectedPath(pathnameWithoutLocale)) {
    return redirectTo(req, withPublicLocale(effectiveLocale, PAGE_ROUTES.HOME))
  }

  if (pathLocale) {
    const canonicalPath = withPublicLocale(pathLocale, pathnameWithoutLocale)

    if (pathname !== canonicalPath) {
      return redirectTo(req, canonicalPath)
    }

    return NextResponse.next()
  }

  const url = req.nextUrl.clone()
  url.pathname = `/${preferredLocale}${pathname === "/" ? "" : pathname}`

  return NextResponse.rewrite(url)
})

export const config = {
  matcher: [
    "/((?!_next|[^?]*\\.(?:html?|css|js(?!on)|jpe?g|webp|png|gif|svg|ttf|woff2?|ico|csv|docx?|xlsx?|zip|webmanifest)).*)",
    "/(api|trpc)(.*)",
  ],
}