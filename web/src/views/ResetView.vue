<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Form as TForm, FormItem as TFormItem, MessagePlugin } from 'tdesign-vue-next'
import type { FormProps } from 'tdesign-vue-next'

import { authApi } from '@/api'
import { ApiError } from '@/api/client'
import UiButton from '@/components/ui/UiButton.vue'
import UiInput from '@/components/ui/UiInput.vue'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()

// 重置链接回跳时带 ?token=...；也可手动粘贴
const routeToken = typeof route.query.token === 'string' ? route.query.token : ''
const form = reactive({ token: routeToken, password: '', confirm: '' })
const loading = ref(false)
const done = ref(false)

const hasToken = computed(() => form.token.trim().length > 0)

const rules: FormProps['rules'] = {
  token: [{ required: true, message: t('auth.required') }],
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
}

function onSubmit({ validateResult }: { validateResult: boolean | Record<string, unknown> }) {
  if (validateResult !== true || loading.value) return
  loading.value = true
  authApi
    .reset(form.token.trim(), form.password)
    .then(() => {
      done.value = true
      MessagePlugin.success(t('reset.success'))
      setTimeout(() => void router.push('/login'), 1500)
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
    <h1 class="auth-card__title">{{ t('reset.title') }}</h1>
    <template v-if="!done">
      <TForm :data="form" :rules="rules" label-width="0" @submit="onSubmit">
        <TFormItem name="token">
          <UiInput v-model="form.token" :placeholder="t('reset.token')" clearable />
        </TFormItem>
        <TFormItem name="password">
          <UiInput v-model="form.password" type="password" :placeholder="t('reset.newPassword')" clearable />
        </TFormItem>
        <TFormItem name="confirm">
          <UiInput v-model="form.confirm" type="password" :placeholder="t('reset.confirmNew')" clearable />
        </TFormItem>
        <UiButton theme="primary" type="submit" block :loading="loading">
          {{ t('reset.submit') }}
        </UiButton>
      </TForm>
      <p v-if="!hasToken" class="auth-card__hint">{{ t('reset.pasteHint') }}</p>
    </template>
    <template v-else>
      <p class="auth-card__hint">{{ t('reset.done') }}</p>
    </template>
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

.auth-card__hint {
  font-size: var(--mc-fs-sm);
  color: var(--td-text-color-secondary);
  line-height: 1.6;
  margin: 12px 0 0;
}

.auth-card__footer {
  text-align: center;
  font-size: var(--mc-fs-sm);
  color: var(--td-text-color-secondary);
  margin-top: 16px;
}
</style>
