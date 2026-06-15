"use client"

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Globe } from "lucide-react"
import { FALLBACK_LANGUAGE, LANGUAGE_LABELS, SUPPORTED_LANGUAGES, SupportedLanguage } from "@/lib/i18n/i18n"
import { useChangeLanguage, useT } from "next-i18next/client"

export function LanguageSwitcher() {
  const { i18n } = useT()
  const changeLanguage = useChangeLanguage()
  const currentLanguage = (i18n.language as SupportedLanguage) || FALLBACK_LANGUAGE

  return (
    <Select value={currentLanguage} onValueChange={(lng) => changeLanguage(lng as SupportedLanguage)}>
      <SelectTrigger className="w-9 md:w-[140px] px-0 md:px-3 [&>span]:hidden md:[&>span]:flex">
        <Globe className="size-4 shrink-0" />
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
