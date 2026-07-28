import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import zhCN from './locales/zh-CN.json'
import en from './locales/en.json'

export const SUPPORTED_LANGUAGES = [
  { code: 'zh-CN', label: '中文' },
  { code: 'en', label: 'English' },
] as const

export type AppLanguage = (typeof SUPPORTED_LANGUAGES)[number]['code']

const STORAGE_KEY = 'i18n-lng'

function isAppLanguage(value: string | null): value is AppLanguage {
  return SUPPORTED_LANGUAGES.some((l) => l.code === value)
}

function detectLanguage(): AppLanguage {
  const stored = localStorage.getItem(STORAGE_KEY)
  if (isAppLanguage(stored)) return stored
  const nav = navigator.language.toLowerCase()
  if (nav.startsWith('zh')) return 'zh-CN'
  if (nav.startsWith('en')) return 'en'
  return 'zh-CN'
}

void i18n.use(initReactI18next).init({
  resources: {
    'zh-CN': { translation: zhCN },
    en: { translation: en },
  },
  lng: detectLanguage(),
  fallbackLng: 'en',
  interpolation: { escapeValue: false },
})

export function setLanguage(lng: AppLanguage) {
  localStorage.setItem(STORAGE_KEY, lng)
  void i18n.changeLanguage(lng)
}

export default i18n
