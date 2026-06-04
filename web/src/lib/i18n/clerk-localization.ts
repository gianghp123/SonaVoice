import { SupportedLanguage } from "./i18n"
import { viVN, enUS } from "@clerk/localizations"

export const getClerkLanguageKey = (locale: SupportedLanguage) => {
  switch (locale) {
    case "en":
      return enUS
    case "vi":
      return viVN
    default:
      return enUS
  }
}