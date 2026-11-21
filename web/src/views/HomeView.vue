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
    <UiButton variant="outline" block @click="showPwdForm = !showPwdForm">
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
        <UiInput v-model="pwdForm.oldPassword" type="password" :placeholder="t('home.oldPassword')" clearable />
      </TFormItem>
      <TFormItem name="newPassword">
        <UiInput v-model="pwdForm.newPassword" type="password" :placeholder="t('home.newPassword')" clearable />
      </TFormItem>
      <TFormItem name="confirm">
        <UiInput v-model="pwdForm.confirm" type="password" :placeholder="t('home.confirmNew')" clearable />
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
</template>

<style scoped>
.auth-card {
  width: 100%;
  max-width: 400px;
  background: var(--td-bg-color-container);
  border: 1px solid var(--mc-hairline);
  border-radius: var(--mc-radius-lg);
  padding: 32px 28px;
}

.auth-card__title {
  font-size: var(--mc-fs-heading);
  font-weight: 600;
  text-align: center;
  margin: 0 0 24px;
}

.home-user {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
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
  color: #fff;
  font-weight: 600;
  flex-shrink: 0;
}

.home-user__meta {
  flex: 1;
  min-width: 0;
}

.home-user__name {
  font-size: var(--mc-fs-emphasis);
  font-weight: 600;
}

.home-user__sub {
  font-size: var(--mc-fs-sm);
  color: var(--td-text-color-secondary);
}

.home-hint {
  font-size: var(--mc-fs-sm);
  color: var(--td-text-color-secondary);
  line-height: 1.6;
  margin: 16px 0 20px;
}

.home-hint--gap {
  margin-top: 20px;
}

.pwd-form {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--mc-hairline);
}
</style>
