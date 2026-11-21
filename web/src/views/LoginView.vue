<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  Button as TButton,
  Form as TForm,
  FormItem as TFormItem,
  Input as TInput,
  MessagePlugin,
} from 'tdesign-vue-next'
import type { FormProps } from 'tdesign-vue-next'

import { ApiError } from '@/api/client'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const form = reactive({ username: '', password: '' })
const loading = ref(false)

const rules: FormProps['rules'] = {
  username: [{ required: true, message: t('auth.required') }],
  password: [{ required: true, message: t('auth.required') }],
}

/** 登录成功后回来源页（client 在 401 时会带 ?redirect=... 回来） */
function afterLogin() {
  const q = route.query.redirect
  const redirect = typeof q === 'string' && q.startsWith('/') ? q : '/'
  void router.push(redirect)
}

/** TForm 的 submit：Enter 与按钮统一走校验→提交（沿用主仓库 AuthView 模式） */
function onSubmit({ validateResult }: { validateResult: boolean | Record<string, unknown> }) {
  if (validateResult !== true || loading.value) return
  loading.value = true
  auth
    .login(form.username, form.password)
    .then(afterLogin)
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
    <h1 class="auth-card__title">{{ t('auth.welcomeBack') }}</h1>
    <TForm :data="form" :rules="rules" label-width="0" @submit="onSubmit">
      <TFormItem name="username">
        <TInput v-model="form.username" :placeholder="t('auth.username')" clearable />
      </TFormItem>
      <TFormItem name="password">
        <TInput v-model="form.password" type="password" :placeholder="t('auth.password')" clearable />
      </TFormItem>
      <TButton theme="primary" type="submit" block :loading="loading">
        {{ t('auth.login') }}
      </TButton>
    </TForm>
    <p class="auth-card__footer">
      <router-link to="/register">{{ t('auth.toRegister') }}</router-link>
    </p>
  </main>
</template>

<style scoped>
.auth-card {
  width: 100%;
  max-width: 360px;
  background: var(--td-bg-color-container);
  border: 1px solid var(--mc-hairline);
  border-radius: var(--td-radius-large);
  padding: 32px 28px;
}

.auth-card__title {
  font-size: 20px;
  font-weight: 600;
  text-align: center;
  margin: 0 0 24px;
}

.auth-card__footer {
  text-align: center;
  font-size: 13px;
  color: var(--td-text-color-secondary);
  margin-top: 16px;
}

.auth-card__footer a {
  color: var(--td-brand-color);
  text-decoration: none;
}
</style>
