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

const LANGUAGES = [
  { code: 'en', label: 'English' },
  { code: 'vi', label: 'Tiếng Việt' },
]

export function LanguageSwitcher() {
  const pathname = usePathname()
  const router = useRouter()
  const { i18n } = useT()

  const currentLng = i18n?.language || 'en'

  const switchLocale = (locale: string) => {
    const segments = pathname.split('/')
    if (segments[1] && LANGUAGES.some(l => l.code === segments[1])) {
      if (locale === 'en') {
        segments.splice(1, 1)
      } else {
        segments[1] = locale
      }
    } else {
      if (locale !== 'en') {
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
        {LANGUAGES.map((lng) => (
          <SelectItem key={lng.code} value={lng.code}>
            {lng.label}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
