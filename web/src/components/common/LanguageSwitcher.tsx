"use client"

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { FALLBACK_LANGUAGE, isSupportedLanguage, LANGUAGE_LABELS, SUPPORTED_LANGUAGES } from "@/lib/i18n"
import { useT } from "next-i18next/client"
import { usePathname, useRouter } from "next/navigation"

export function LanguageSwitcher() {
  const pathname = usePathname()
  const router = useRouter()
  const { i18n } = useT()
  const currentLng = i18n.language

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