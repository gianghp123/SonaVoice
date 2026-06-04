"use client"

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { FALLBACK_LANGUAGE, isSupportedLanguage, LANGUAGE_LABELS, SUPPORTED_LANGUAGES } from "@/lib/i18n"
import { getLocaleFromPathname } from "@/lib/utils/path"
import { usePathname, useRouter } from "next/navigation"

export function LanguageSwitcher() {
  const pathname = usePathname()
  const router = useRouter()

  const currentLng = getLocaleFromPathname(pathname) ?? FALLBACK_LANGUAGE

  const switchLocale = (locale: string) => {
    const segments = pathname.split('/').filter(Boolean)
    const pathWithoutLocale = isSupportedLanguage(segments[0])
      ? segments.slice(1)
      : segments
    const nextPath =
      locale === FALLBACK_LANGUAGE
        ? `/${pathWithoutLocale.join('/')}`
        : `/${locale}/${pathWithoutLocale.join('/')}`

    router.push(nextPath === '/' ? '/' : nextPath.replace(/\/$/, ''))
  }

  console.log(currentLng)

  return (
    <Select value={currentLng} onValueChange={switchLocale}>
      <SelectTrigger className="w-[140px]">
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        {SUPPORTED_LANGUAGES.map((lng) => (
          <SelectItem key={lng} value={lng}>
            {LANGUAGE_LABELS[lng]}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}