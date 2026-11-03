import { createI18n } from 'vue-i18n'
import zhCN from './locales/zh-CN'
import en from './locales/en'
import ja from './locales/ja'

export type Locale = 'zh-CN' | 'en' | 'ja'

const LOCALE_KEY = 'go-ffmpeg-tools-locale'

export const LOCALE_OPTIONS: Array<{ value: Locale; label: string }> = [
  { value: 'zh-CN', label: '中文' },
  { value: 'en', label: 'English' },
  { value: 'ja', label: '日本語' },
]

// 初始化：读取上次保存的语言，否则按系统语言推断
export function detectLocale(): Locale {
  const saved = localStorage.getItem(LOCALE_KEY)
  if (saved === 'zh-CN' || saved === 'en' || saved === 'ja') return saved
  const nav = (navigator.language || '').toLowerCase()
  if (nav.startsWith('zh')) return 'zh-CN'
  if (nav.startsWith('ja')) return 'ja'
  return 'en'
}

export function saveLocale(locale: Locale) {
  localStorage.setItem(LOCALE_KEY, locale)
}

const i18n = createI18n({
  legacy: false,
  locale: detectLocale(),
  fallbackLocale: 'zh-CN',
  messages: {
    'zh-CN': zhCN,
    en,
    ja,
  },
})

export default i18n
