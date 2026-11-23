<script setup lang="ts">
// SSO 门户外壳：品牌头 + 居中内容 + 页脚。语言由 i18n 启动时探测（zh 开头中文，其余英文）。
</script>

<template>
  <div class="app-shell">
    <header class="app-shell__header">
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

    <footer class="app-shell__footer">
      <span>Modernchat SSO</span>
      <span class="app-shell__dot">·</span>
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
  background-image:
    radial-gradient(60% 45% at 50% 0%, var(--mc-brand-soft), transparent 70%),
    linear-gradient(var(--mc-app-pattern) 1px, transparent 1px),
    linear-gradient(90deg, var(--mc-app-pattern) 1px, transparent 1px);
  background-size: 100% 100%, 40px 40px, 40px 40px;
  background-position: center top, -1px -1px, -1px -1px;
}

.app-shell__header {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 28px;
  max-width: 1160px;
  width: 100%;
  margin: 0 auto;
}

.app-shell__brand {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  color: var(--td-text-color-primary);
  text-decoration: none;
}

.app-shell__mark {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  border-radius: var(--mc-radius-md);
  background: var(--td-brand-color);
  color: #fff;
  box-shadow: var(--mc-elev-1);
}

.app-shell__name {
  font-size: var(--mc-fs-emphasis);
  font-weight: 500;
  letter-spacing: -0.01em;
}

.app-shell__name b {
  color: var(--td-brand-color);
  font-weight: 600;
}

.app-shell__tagline {
  font-size: var(--mc-fs-sm);
  color: var(--td-text-color-secondary);
}

.app-shell__main {
  position: relative;
  z-index: 1;
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px 24px 40px;
}

.app-shell__footer {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 20px;
  font-size: var(--mc-fs-sm);
  color: var(--td-text-color-placeholder);
}

.app-shell__dot {
  opacity: 0.5;
}

@media (max-width: 640px) {
  .app-shell__header { padding: 16px 20px; }
  .app-shell__tagline { display: none; }
  .app-shell__main { align-items: flex-start; padding: 8px 16px 32px; }
}
</style>
