import en from '../locales/en/translation.json'
import zh from '../locales/zh/translation.json'

export const languageResources = {
  en: { translation: en },
  zh: { translation: zh },
} as const

export const supportedLanguageOptions = [
  { code: 'en', label: 'EN' },
  { code: 'zh', label: '中文' },
] as const satisfies ReadonlyArray<{ code: keyof typeof languageResources; label: string }>

export type SupportedLanguage = (typeof supportedLanguageOptions)[number]['code']

export const supportedLanguages = supportedLanguageOptions.map(({ code }) => code) as SupportedLanguage[]
export const defaultLanguage: SupportedLanguage = 'en'
