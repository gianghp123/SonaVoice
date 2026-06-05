import type { I18nConfig } from 'next-i18next/proxy'
import { FALLBACK_LANGUAGE, SUPPORTED_LANGUAGES } from './src/lib/i18n/i18n'

const i18nConfig: I18nConfig = {
  supportedLngs: [...SUPPORTED_LANGUAGES],
  fallbackLng: FALLBACK_LANGUAGE,
  defaultNS: 'common',
  ns: ['common', 'home', 'chat', 'session'],
  hideDefaultLocale: true,
  reloadOnPrerender: process.env.NODE_ENV === 'development',
  resourceLoader:
    process.env.NODE_ENV === 'development'
      ? async (lng: string, ns: string) => {
          const fs = await import('fs/promises')
          const path = await import('path')
          const content = await fs.readFile(
            path.resolve(process.cwd(), `src/app/i18n/locales/${lng}/${ns}.json`),
            'utf-8'
          )
          return JSON.parse(content)
        }
      : (lng: string, ns: string) =>
          import(`./src/app/i18n/locales/${lng}/${ns}.json`),
}

export default i18nConfig
