import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import LanguageDetector from 'i18next-browser-languagedetector'
import { en, zh } from './resources'

export const supportedLanguages = ['en', 'zh'] as const
export type SupportedLanguage = (typeof supportedLanguages)[number]

export const defaultLanguage: SupportedLanguage = 'en'

export function isSupportedLanguage(language: string): language is SupportedLanguage {
  return supportedLanguages.some((supportedLanguage) => supportedLanguage === language)
}

export function resolveSupportedLanguage(language?: string | null): SupportedLanguage {
  const normalized = language?.trim().toLowerCase()
  if (!normalized) {
    return defaultLanguage
  }

  if (isSupportedLanguage(normalized)) {
    return normalized
  }

  const baseLanguage = normalized.split(/[-_]/)[0]
  if (isSupportedLanguage(baseLanguage)) {
    return baseLanguage
  }

  return defaultLanguage
}

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources: {
      en: { translation: en },
      zh: { translation: zh },
    },
    fallbackLng: defaultLanguage,
    supportedLngs: supportedLanguages,
    nonExplicitSupportedLngs: true,
    detection: {
      order: ['localStorage'],
      caches: ['localStorage'],
    },
    interpolation: {
      escapeValue: false,
    },
  })

export default i18n

export function changeLanguage(language: string) {
  return i18n.changeLanguage(resolveSupportedLanguage(language))
}

export function getCurrentLanguage(): SupportedLanguage {
  return resolveSupportedLanguage(i18n.resolvedLanguage ?? i18n.language)
}
