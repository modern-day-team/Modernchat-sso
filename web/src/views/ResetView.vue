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
  <div class="mc-container mc-container--narrow auth-wrap">
  <main class="auth-card">
    <h1 class="auth-card__title">{{ t('reset.title') }}</h1>
    <template v-if="!done">
      <TForm :data="form" :rules="rules" label-width="0" @submit="onSubmit">
        <TFormItem name="token">
          <UiInput v-model="form.token" :placeholder="t('reset.token')" :aria-label="t('reset.token')" clearable />
        </TFormItem>
        <TFormItem name="password">
          <UiInput v-model="form.password" type="password" :placeholder="t('reset.newPassword')" :aria-label="t('reset.newPassword')" clearable />
        </TFormItem>
        <TFormItem name="confirm">
          <UiInput v-model="form.confirm" type="password" :placeholder="t('reset.confirmNew')" :aria-label="t('reset.confirmNew')" clearable />
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
  text-align: center;
  margin: 0 0 var(--mc-sp-6);
}

.auth-card__hint {
  font-size: var(--mc-fs-sm);
  color: var(--td-text-color-secondary);
  line-height: var(--mc-lh-normal);
  margin: var(--mc-sp-3) 0 0;
}

.auth-card__footer {
  text-align: center;
  font-size: var(--mc-fs-sm);
  color: var(--td-text-color-secondary);
  margin-top: var(--mc-sp-4);
}
</style>
