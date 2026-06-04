"use client"

import { usePathname, useRouter } from "next/navigation"
import { useT } from "next-i18next/client"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { LANGUAGE_LABELS, SUPPORTED_LANGUAGES, FALLBACK_LANGUAGE } from "@/lib/i18n"
import { getLocaleFromPathname } from "@/lib/utils/path"

export function LanguageSwitcher() {
  const pathname = usePathname()
  const router = useRouter()
  const { i18n } = useT()

  const currentLng = i18n?.language || FALLBACK_LANGUAGE

  const switchLocale = (locale: string) => {
    const segments = pathname.split('/')
    const existingLocale = getLocaleFromPathname(pathname)

    if (existingLocale) {
      if (locale === FALLBACK_LANGUAGE) {
        segments.splice(1, 1)
      } else {
        segments[1] = locale
      }
    } else {
      if (locale !== FALLBACK_LANGUAGE) {
        segments.splice(1, 0, locale)
      }
    }
    router.push(segments.join('/') || '/')
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
