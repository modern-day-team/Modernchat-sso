<script setup lang="ts">
// SSO 门户外壳：居中卡片容器。语言由 i18n 启动时探测（zh 开头中文，其余英文）。
</script>

<template>
  <div class="app-shell">
    <router-view />
  </div>
</template>

<style scoped>
.app-shell {
  min-height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
}
</style>
