<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Form as TForm, FormItem as TFormItem, MessagePlugin } from 'tdesign-vue-next'
import type { FormProps } from 'tdesign-vue-next'

import { ApiError } from '@/api/client'
import UiButton from '@/components/ui/UiButton.vue'
import UiInput from '@/components/ui/UiInput.vue'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const router = useRouter()
const auth = useAuthStore()

const form = reactive({ username: '', password: '', confirm: '', nickname: '', email: '' })
const loading = ref(false)

const rules: FormProps['rules'] = {
  username: [
    { required: true, message: t('auth.required') },
    { max: 20, message: t('auth.usernameMax') },
  ],
  password: [
    { required: true, message: t('auth.required') },
    { min: 6, max: 72, message: t('auth.passwordLen') },
  ],
  confirm: [
    {
      required: true,
      validator: (v: string) => v === form.password,
      message: t('auth.pwdMismatch'),
    },
  ],
  email: [{ validator: (v: string) => !v || /.+@.+\..+/.test(v), message: t('auth.emailInvalid') }],
}

/** TForm 的 submit：Enter 与按钮统一走校验→提交（沿用主仓库 AuthView 模式） */
function onSubmit({ validateResult }: { validateResult: boolean | Record<string, unknown> }) {
  if (validateResult !== true || loading.value) return
  loading.value = true
  auth
    // 注册成功后 store 内自动登录
    .register(form.username, form.password, form.nickname, form.email || undefined)
    .then(() => {
      MessagePlugin.success(t('auth.registered'))
      void router.push('/')
    })
    .catch((e: unknown) => {
      MessagePlugin.error(e instanceof ApiError ? e.message : t('auth.failed'))
    })
    .finally(() => {
      loading.value = false
    })
}
</script>

<template>
  <main class="auth-card">
    <h1 class="auth-card__title">{{ t('auth.register') }}</h1>
    <TForm :data="form" :rules="rules" label-width="0" @submit="onSubmit">
      <TFormItem name="username">
        <UiInput v-model="form.username" :placeholder="t('auth.username')" clearable />
      </TFormItem>
      <TFormItem name="nickname">
        <UiInput v-model="form.nickname" :placeholder="t('auth.nickname')" clearable />
      </TFormItem>
      <TFormItem name="email">
        <UiInput v-model="form.email" type="email" :placeholder="t('auth.email')" clearable />
      </TFormItem>
      <TFormItem name="password">
        <UiInput v-model="form.password" type="password" :placeholder="t('auth.password')" clearable />
      </TFormItem>
      <TFormItem name="confirm">
        <UiInput v-model="form.confirm" type="password" :placeholder="t('auth.confirmPassword')" clearable />
      </TFormItem>
      <UiButton theme="primary" type="submit" block :loading="loading">
        {{ t('auth.register') }}
      </UiButton>
    </TForm>
    <p class="auth-card__footer">
      <router-link to="/login">{{ t('auth.toLogin') }}</router-link>
    </p>
  </main>
</template>

<style scoped>
.auth-card {
  width: 100%;
  max-width: 360px;
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

.auth-card__footer {
  text-align: center;
  font-size: var(--mc-fs-sm);
  color: var(--td-text-color-secondary);
  margin-top: 16px;
}
</style>
