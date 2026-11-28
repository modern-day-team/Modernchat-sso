<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Form as TForm, FormItem as TFormItem, MessagePlugin } from 'tdesign-vue-next'
import type { FormProps } from 'tdesign-vue-next'

import { authApi } from '@/api'
import { ApiError } from '@/api/client'
import UiButton from '@/components/ui/UiButton.vue'
import UiInput from '@/components/ui/UiInput.vue'
import UiTag from '@/components/ui/UiTag.vue'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const router = useRouter()
const auth = useAuthStore()

// 进入门户即拉取最新资料；401 由 client 统一清理并带回登录页
onMounted(() => {
  void auth.fetchMe()
})

async function onLogout() {
  await auth.logout()
  MessagePlugin.success(t('home.loggedOut'))
  void router.push('/login')
}

// ===== 自身改密（成功后全端失效，需重新登录） =====
const pwdForm = reactive({ oldPassword: '', newPassword: '', confirm: '' })
const pwdLoading = ref(false)
const showPwdForm = ref(false)

const pwdRules: FormProps['rules'] = {
  oldPassword: [{ required: true, message: t('auth.required') }],
  newPassword: [
    { required: true, message: t('auth.required') },
    { min: 6, max: 72, message: t('auth.passwordLen') },
  ],
  confirm: [
    {
      required: true,
      validator: (v: string) => v === pwdForm.newPassword,
      message: t('auth.pwdMismatch'),
    },
  ],
}

function onPwdSubmit({ validateResult }: { validateResult: boolean | Record<string, unknown> }) {
  if (validateResult !== true || pwdLoading.value) return
  pwdLoading.value = true
  authApi
    .changePassword(pwdForm.oldPassword, pwdForm.newPassword)
    .then(() => {
      MessagePlugin.success(t('home.pwdChanged'))
      // tv 已 bump：本地会话立即失效，清态回登录页
      return auth.logout()
    })
    .then(() => {
      void router.push('/login')
    })
    .catch((e: unknown) => {
      MessagePlugin.error(e instanceof ApiError ? e.message : t('auth.failed'))
    })
    .finally(() => {
      pwdLoading.value = false
    })
}
</script>

<template>
  <div class="mc-container mc-container--narrow auth-wrap">
  <main class="auth-card">
    <h1 class="auth-card__title">{{ t('home.title') }}</h1>
    <div class="home-user">
      <span class="home-user__avatar" :style="{ background: auth.user?.avatarColor }">
        {{ auth.user?.nickname?.slice(0, 1) || '?' }}
      </span>
      <div class="home-user__meta">
        <div class="home-user__name">{{ auth.user?.nickname }}（@{{ auth.user?.username }}）</div>
        <div class="home-user__sub">{{ auth.user?.email || t('home.noEmail') }}</div>
      </div>
      <UiTag :theme="auth.user?.role === 'admin' ? 'primary' : 'default'">
        {{ auth.user?.role }}
      </UiTag>
    </div>
    <p class="home-hint">{{ t('home.hint') }}</p>
    <UiButton v-if="auth.user?.role === 'admin'" variant="outline" block @click="router.push('/admin')">
      {{ t('home.adminConsole') }}
    </UiButton>
    <UiButton class="home-gap" variant="outline" block @click="showPwdForm = !showPwdForm">
      {{ t('home.changePassword') }}
    </UiButton>
    <TForm
      v-if="showPwdForm"
      class="pwd-form"
      :data="pwdForm"
      :rules="pwdRules"
      label-width="0"
      @submit="onPwdSubmit"
    >
      <TFormItem name="oldPassword">
        <UiInput v-model="pwdForm.oldPassword" type="password" :placeholder="t('home.oldPassword')" :aria-label="t('home.oldPassword')" clearable />
      </TFormItem>
      <TFormItem name="newPassword">
        <UiInput v-model="pwdForm.newPassword" type="password" :placeholder="t('home.newPassword')" :aria-label="t('home.newPassword')" clearable />
      </TFormItem>
      <TFormItem name="confirm">
        <UiInput v-model="pwdForm.confirm" type="password" :placeholder="t('home.confirmNew')" :aria-label="t('home.confirmNew')" clearable />
      </TFormItem>
      <UiButton theme="primary" type="submit" block :loading="pwdLoading">
        {{ t('home.pwdSubmit') }}
      </UiButton>
    </TForm>
    <p class="home-hint home-hint--gap">{{ t('home.logoutHint') }}</p>
    <UiButton theme="danger" variant="outline" block @click="onLogout">
      {{ t('home.logout') }}
    </UiButton>
  </main>
  </div>
</template>

<style scoped>
/* 外框留白 = --mc-gutter（由 mc-container 提供），上下留白走刻度 */
.auth-wrap {
  padding-block: var(--mc-sp-4) var(--mc-sp-8);
}

.auth-card {
/* 卡片宽度 = --mc-container-narrow（外层 mc-container） */
  width: 100%;
  background: var(--td-bg-color-container);
  border: 1px solid var(--mc-hairline);
  border-radius: var(--mc-radius-lg);
  padding: var(--mc-sp-8) var(--mc-sp-6) var(--mc-sp-6);
  box-shadow: var(--mc-elev-1);
}

.auth-card__title {
  font-size: var(--mc-fs-heading);
  font-weight: 600;
  line-height: var(--mc-lh-tight);
  letter-spacing: -0.02em;
  text-align: center;
  margin: 0 0 var(--mc-sp-6);
}

.home-user {
  display: flex;
  align-items: center;
  gap: var(--mc-sp-3);
  padding: var(--mc-sp-3);
  border-radius: var(--mc-radius-md);
  background: var(--td-bg-color-secondarycontainer);
}

.home-user__avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--td-text-color-anti);
  font-weight: 600;
  flex-shrink: 0;
  font-size: var(--mc-fs-title);
}

.home-user__meta {
  flex: 1;
  min-width: 0;
}

.home-user__name {
  font-size: var(--mc-fs-emphasis);
  font-weight: 600;
  line-height: var(--mc-lh-snug);
  overflow-wrap: anywhere;
}

.home-user__sub {
  font-size: var(--mc-fs-sm);
  line-height: var(--mc-lh-snug);
  color: var(--td-text-color-secondary);
  overflow-wrap: anywhere;
}

.home-hint {
  font-size: var(--mc-fs-sm);
  color: var(--td-text-color-secondary);
  line-height: var(--mc-lh-normal);
  margin: var(--mc-sp-4) 0 var(--mc-sp-5);
}

.home-hint--gap {
  margin-top: var(--mc-sp-5);
}

.home-gap {
  margin-top: var(--mc-sp-2);
}

.pwd-form {
  margin-top: var(--mc-sp-4);
  padding-top: var(--mc-sp-4);
  border-top: 1px solid var(--mc-hairline);
}
</style>
