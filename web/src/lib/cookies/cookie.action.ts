'use server'

import { cookies } from 'next/headers'
import { FALLBACK_LANGUAGE, isSupportedLanguage } from '@/lib/i18n/i18n'
import { COOKIE_MAX_AGE, LOCALE_COOKIE } from './cookie.contants'


export async function setLocale(locale: string) {
  if (!isSupportedLanguage(locale)) return

  const cookieStore = await cookies()
  cookieStore.set(LOCALE_COOKIE, locale, {
    path: '/',
    maxAge: COOKIE_MAX_AGE,
    sameSite: 'lax',
  })
}

export async function getLocale(): Promise<string> {
  const cookieStore = await cookies()
  const cookie = cookieStore.get(LOCALE_COOKIE)?.value

  if (cookie && isSupportedLanguage(cookie)) {
    return cookie
  }

  return FALLBACK_LANGUAGE
}
