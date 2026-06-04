"use client"

import { usePathname, useRouter } from "next/navigation"
import { useT } from "next-i18next/client"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

const LANGUAGES = [
  { code: 'en', label: 'EN' },
  { code: 'vi', label: 'VI' },
]

export function LanguageSwitcher() {
  const pathname = usePathname()
  const router = useRouter()
  const { i18n } = useT()

  const currentLng = i18n?.language || 'en'

  const switchLocale = (locale: string) => {
    const segments = pathname.split('/')
    // Check if the first segment is a locale
    if (segments[1] && LANGUAGES.some(l => l.code === segments[1])) {
      if (locale === 'en') {
        // Remove locale prefix for default language (hideDefaultLocale)
        segments.splice(1, 1)
      } else {
        segments[1] = locale
      }
    } else {
      // No locale in path, add one (unless it's the default)
      if (locale !== 'en') {
        segments.splice(1, 0, locale)
      }
    }
    router.push(segments.join('/') || '/')
  }

  return (
    <div className="flex gap-1">
      {LANGUAGES.map((lng) => (
        <Button
          key={lng.code}
          variant={currentLng === lng.code ? "default" : "ghost"}
          size="sm"
          className={cn(
            "h-7 px-2 text-xs",
            currentLng === lng.code && "pointer-events-none"
          )}
          onClick={() => switchLocale(lng.code)}
        >
          {lng.label}
        </Button>
      ))}
    </div>
  )
}
