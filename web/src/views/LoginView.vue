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
/* 提交失败的内联错误：比 toast 更靠近表单且不自动消失（toast 仍保留给全局观察） */
const error = ref('')

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
  error.value = ''
  auth
    .login(form.username, form.password)
    .then(afterLogin)
    .catch((e: unknown) => {
      error.value = e instanceof ApiError ? e.message : t('auth.failed')
      MessagePlugin.error(error.value)
    })
    .finally(() => {
      loading.value = false
    })
}
</script>

<template>
  <div class="mc-container mc-container--narrow auth-wrap">
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
      <!-- 提交失败的持久内联提示（错误说明靠近表单，不随 toast 消失） -->
      <p v-if="error" class="auth-card__error" role="alert">{{ error }}</p>
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

.auth-card__mark {
  width: var(--mc-ctl-lg);
  height: var(--mc-ctl-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto var(--mc-sp-4);
  border-radius: var(--mc-radius-md);
  background: var(--mc-brand-soft);
  color: var(--td-brand-color);
}

.auth-card__title {
  font-size: var(--mc-fs-heading);
  font-weight: 600;
  line-height: var(--mc-lh-tight);
  text-align: center;
  letter-spacing: -0.02em;
  margin: 0;
}

.auth-card__subtitle {
  text-align: center;
  font-size: var(--mc-fs-sm);
  color: var(--td-text-color-secondary);
  margin: var(--mc-sp-1) 0 var(--mc-sp-6);
}

.auth-card__submit {
  margin-top: var(--mc-sp-2);
}

.auth-card__error {
  margin: var(--mc-sp-2) 0 0;
  padding: var(--mc-sp-2) var(--mc-sp-3);
  border-radius: var(--mc-radius-sm);
  background: var(--td-error-color-1, rgba(227, 77, 89, 0.08));
  color: var(--td-error-color);
  font-size: var(--mc-fs-sm);
  line-height: var(--mc-lh-normal);
}

.auth-card__divider {
  height: 1px;
  background: var(--mc-hairline);
  margin: var(--mc-sp-6) 0 var(--mc-sp-4);
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

.auth-card__footer a:hover,
.auth-card__forgot:hover {
  color: var(--td-brand-color-hover);
}

.auth-card__forgot {
  display: block;
  text-align: center;
  font-size: var(--mc-fs-sm);
  color: var(--td-text-color-secondary);
  margin-top: var(--mc-sp-3);
  transition: color var(--mc-dur-instant) var(--mc-ease-out);
}
</style>
