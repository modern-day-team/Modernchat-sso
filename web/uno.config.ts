import { defineConfig, presetWind3 } from 'unocss'

/**
 * UnoCSS 配置（与主仓库 Modernchat-new 同体系）
 *
 * 颜色值全部指向产品的 CSS 变量（--td-* / --mc-*），因此：
 * - 亮/暗主题切换自动生效（变量由 TDesign 主题体系维护）
 * - 原子类与手写样式取同一套色值
 *
 * 用法示例：`bg-surface text-ink-2 border-line rounded-lg px-3 py-2`
 */
export default defineConfig({
  presets: [presetWind3()],
  theme: {
    colors: {
      // 品牌
      brand: 'var(--td-brand-color)',
      'brand-hover': 'var(--td-brand-color-hover)',
      'brand-active': 'var(--td-brand-color-active)',
      'brand-light': 'var(--td-brand-color-light)',
      // 文本
      ink: 'var(--td-text-color-primary)',
      'ink-2': 'var(--td-text-color-secondary)',
      'ink-3': 'var(--td-text-color-placeholder)',
      'ink-4': 'var(--td-text-color-disabled)',
      // 表面
      surface: 'var(--td-bg-color-container)',
      page: 'var(--td-bg-color-page)',
      hover: 'var(--td-bg-color-container-hover)',
      active: 'var(--td-bg-color-container-active)',
      line: 'var(--mc-hairline)',
      // 语义
      danger: 'var(--td-error-color)',
      success: 'var(--td-success-color)',
      warning: 'var(--td-warning-color)',
      info: 'var(--td-brand-color)',
    },
  },
})
