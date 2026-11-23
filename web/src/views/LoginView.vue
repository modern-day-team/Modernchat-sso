<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Form as TForm, FormItem as TFormItem, MessagePlugin } from 'tdesign-vue-next'
import type { FormProps } from 'tdesign-vue-next'

import { ApiError } from '@/api/client'
import UiButton from '@/components/ui/UiButton.vue'
import UiInput from '@/components/ui/UiInput.vue'
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
    <div class="auth-card__mark" aria-hidden="true">
      <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
        <rect x="4" y="10" width="16" height="10" rx="2" />
        <path d="M8 10V7a4 4 0 0 1 8 0v3" />
        <circle cx="12" cy="15" r="1" fill="currentColor" stroke="none" />
      </svg>
    </div>
    <h1 class="auth-card__title">{{ t('auth.welcomeBack') }}</h1>
    <p class="auth-card__subtitle">{{ t('auth.loginSubtitle') }}</p>

    <TForm :data="form" :rules="rules" label-width="0" @submit="onSubmit">
      <TFormItem name="username">
        <UiInput v-model="form.username" :placeholder="t('auth.username')" clearable ariaLabel="username" />
      </TFormItem>
      <TFormItem name="password">
        <UiInput v-model="form.password" type="password" :placeholder="t('auth.password')" clearable ariaLabel="password" />
      </TFormItem>
      <UiButton theme="primary" type="submit" block :loading="loading" class="auth-card__submit">
        {{ t('auth.login') }}
      </UiButton>
    </TForm>

    <div class="auth-card__divider" />
    <p class="auth-card__footer">
      {{ t('auth.noAccount') }} <router-link to="/register">{{ t('auth.toRegister') }}</router-link>
    </p>
    <router-link class="auth-card__forgot" to="/forgot">{{ t('auth.forgotLink') }}</router-link>
  </main>
</template>

<style scoped>
.auth-card {
  width: 100%;
  max-width: 400px;
  background: var(--td-bg-color-container);
  border: 1px solid var(--mc-hairline);
  border-radius: var(--mc-radius-lg);
  padding: 36px 32px 28px;
  box-shadow: var(--mc-elev-1);
}

.auth-card__mark {
  width: 44px;
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 16px;
  border-radius: var(--mc-radius-md);
  background: var(--mc-brand-soft);
  color: var(--td-brand-color);
}

.auth-card__title {
  font-size: var(--mc-fs-heading);
  font-weight: 600;
  text-align: center;
  letter-spacing: -0.02em;
  margin: 0;
}

.auth-card__subtitle {
  text-align: center;
  font-size: var(--mc-fs-sm);
  color: var(--td-text-color-secondary);
  margin: 6px 0 28px;
}

.auth-card__submit {
  margin-top: 8px;
}

.auth-card__divider {
  height: 1px;
  background: var(--mc-hairline);
  margin: 24px 0 16px;
}

.auth-card__footer {
  text-align: center;
  font-size: var(--mc-fs-body);
  color: var(--td-text-color-secondary);
  margin: 0;
}

.auth-card__footer a {
  color: var(--td-brand-color);
  font-weight: 500;
}

.auth-card__forgot {
  display: block;
  text-align: center;
  font-size: var(--mc-fs-sm);
  color: var(--td-text-color-secondary);
  margin-top: 12px;
}

.auth-card__forgot:hover {
  color: var(--td-brand-color);
}
</style>
