export const SUPPORTED_LANGUAGES = ['en', 'vi'] as const

export type SupportedLanguage = (typeof SUPPORTED_LANGUAGES)[number]

export const FALLBACK_LANGUAGE: SupportedLanguage = 'en'

export const LANGUAGE_LABELS: Record<SupportedLanguage, string> = {
  en: '🇬🇧 English',
  vi: '🇻🇳 Tiếng Việt',
}

export function isSupportedLanguage(
  locale: string | undefined
): locale is SupportedLanguage {
  return !!locale && SUPPORTED_LANGUAGES.includes(locale as SupportedLanguage)
}
