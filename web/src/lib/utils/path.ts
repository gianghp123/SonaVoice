import { FALLBACK_LANGUAGE, isSupportedLanguage } from "@/lib/i18n/i18n"

export function stripLocalePrefix(pathname: string): string {
  const segments = pathname.split("/").filter(Boolean)

  if (segments[0] && isSupportedLanguage(segments[0])) {
    const pathWithoutLocale = segments.slice(1).join("/")
    return pathWithoutLocale ? `/${pathWithoutLocale}` : "/"
  }

  return pathname || "/"
}

export function getLocaleFromPathname(pathname: string): string | undefined {
  const segments = pathname.split("/").filter(Boolean)
  const maybeLocale = segments[0] ?? FALLBACK_LANGUAGE

  return maybeLocale && isSupportedLanguage(maybeLocale)
    ? maybeLocale
    : undefined
}