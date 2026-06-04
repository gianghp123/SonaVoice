import { SUPPORTED_LANGUAGES } from '@/lib/i18n'

export function stripLocalePrefix(pathname: string): string {
  const segments = pathname.split('/')
  if (segments[1] && SUPPORTED_LANGUAGES.includes(segments[1] as any)) {
    return '/' + segments.slice(2).join('/')
  }
  return pathname
}

export function getLocaleFromPathname(pathname: string): string | undefined {
  const segments = pathname.split('/')
  if (segments[1] && SUPPORTED_LANGUAGES.includes(segments[1] as any)) {
    return segments[1]
  }
  return undefined
}
