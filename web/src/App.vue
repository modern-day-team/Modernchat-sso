<script setup lang="ts">
// SSO 门户外壳：品牌头 + 居中内容 + 页脚。语言由 i18n 启动时探测（zh 开头中文，其余英文）。
</script>

<template>
  <div class="app-shell">
    <header class="app-shell__header mc-container mc-row mc-row--between">
      <router-link to="/" class="app-shell__brand" aria-label="Modernchat SSO">
        <span class="app-shell__mark" aria-hidden="true">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <path d="M4 6.5A2.5 2.5 0 0 1 6.5 4h11A2.5 2.5 0 0 1 20 6.5v7a2.5 2.5 0 0 1-2.5 2.5H10l-4.5 3.5V16A2.5 2.5 0 0 1 4 13.5v-7Z" />
            <path d="M9 9.5h6M9 12.5h4" />
          </svg>
        </span>
        <span class="app-shell__name">Modernchat <b>SSO</b></span>
      </router-link>
      <span class="app-shell__tagline">统一身份认证</span>
    </header>

    <main class="app-shell__main">
      <router-view />
    </main>

    <footer class="app-shell__footer mc-container">
      <span>Modernchat SSO</span>
      <span class="app-shell__dot">|</span>
      <span>项目组统一登录入口</span>
    </footer>
  </div>
</template>

<style scoped>
.app-shell {
  position: relative;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: var(--td-bg-color-page);
  /* 背景网格：方格边长 = 2 × 栅格槽宽（32 / 48px），与页面留白同一节奏 */
  background-image:
    radial-gradient(60% 45% at 50% 0%, var(--mc-brand-soft), transparent 70%),
    linear-gradient(var(--mc-app-pattern) 1px, transparent 1px),
    linear-gradient(90deg, var(--mc-app-pattern) 1px, transparent 1px);
  background-size: 100% 100%, calc(var(--mc-gutter) * 2) calc(var(--mc-gutter) * 2), calc(var(--mc-gutter) * 2) calc(var(--mc-gutter) * 2);
  background-position: center top, -1px -1px, -1px -1px;
}

/* 头 / 主体 / 页脚都是标准档容器（--mc-container），与栅格叠加层同一套列 */
.app-shell__header {
  position: relative;
  z-index: 1;
  padding-block: var(--mc-gutter);
}

.app-shell__brand {
  display: inline-flex;
  align-items: center;
  gap: var(--mc-sp-2);
  padding: var(--mc-sp-1);
  margin: calc(var(--mc-sp-1) * -1);
  border-radius: var(--mc-radius-md);
  color: var(--td-text-color-primary);
  text-decoration: none;
  transition: background-color var(--mc-dur-instant) var(--mc-ease-out);
}

.app-shell__brand:hover {
  background: var(--td-bg-color-container-hover);
}

.app-shell__mark {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: var(--mc-ctl-lg);
  height: var(--mc-ctl-lg);
  border-radius: var(--mc-radius-md);
  background: var(--td-brand-color);
  color: var(--td-text-color-anti);
  box-shadow: var(--mc-elev-1);
}

.app-shell__name {
  font-size: var(--mc-fs-emphasis);
  font-weight: 500;
  line-height: var(--mc-lh-snug);
  letter-spacing: -0.01em;
}

.app-shell__name b {
  color: var(--td-brand-color);
  font-weight: 600;
}

.app-shell__tagline {
  font-size: var(--mc-fs-sm);
  line-height: var(--mc-lh-snug);
  color: var(--td-text-color-secondary);
}

/* 左右留白交给各页面的容器/页面骨架（都是 --mc-gutter），这里只负责垂直居中 */
.app-shell__main {
  position: relative;
  z-index: 1;
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 0;
}

.app-shell__footer {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--mc-sp-2);
  padding-block: var(--mc-sp-4) var(--mc-sp-5);
  font-size: var(--mc-fs-sm);
  line-height: var(--mc-lh-normal);
  color: var(--td-text-color-placeholder);
}

.app-shell__dot {
  opacity: 0.5;
}

@media (max-width: 479px) {
  .app-shell__tagline {
    display: none;
  }
}
</style>
