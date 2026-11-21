import { createI18n } from 'vue-i18n'

import enUS from './en-US'
import zhCN from './zh-CN'

export type Lang = 'zh-CN' | 'en-US'

const LANG_KEY = 'sso_lang'

/** 默认语言：已保存 → 浏览器语言（zh 开头为中文，其余英文）。 */
export function detectLang(): Lang {
  const saved = localStorage.getItem(LANG_KEY)
  if (saved === 'zh-CN' || saved === 'en-US') return saved
  return navigator.language.toLowerCase().startsWith('zh') ? 'zh-CN' : 'en-US'
}

const i18n = createI18n({
  legacy: false,
  locale: detectLang(),
  fallbackLocale: 'zh-CN',
  messages: {
    'zh-CN': zhCN,
    'en-US': enUS,
  },
})

export { i18n }
export default i18n
