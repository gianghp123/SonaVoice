import { isSupportedLanguage } from '@/lib/i18n/i18n'

export function stripLocalePrefix(pathname: string): string {
  const segments = pathname.split('/')
  if (segments[1] && isSupportedLanguage(segments[1])) {
    return '/' + segments.slice(2).join('/')
  }
  return pathname
}

export function getLocaleFromPathname(pathname: string): string | undefined {
  const segments = pathname.split('/')
  if (segments[1] && isSupportedLanguage(segments[1])) {
    return segments[1]
  }
  return undefined
}
