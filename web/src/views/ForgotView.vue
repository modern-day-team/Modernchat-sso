<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Form as TForm, FormItem as TFormItem, MessagePlugin } from 'tdesign-vue-next'
import type { FormProps } from 'tdesign-vue-next'

import { authApi } from '@/api'
import { ApiError } from '@/api/client'
import UiButton from '@/components/ui/UiButton.vue'
import UiInput from '@/components/ui/UiInput.vue'

const { t } = useI18n()

const form = reactive({ email: '' })
const loading = ref(false)
const sent = ref(false)

const rules: FormProps['rules'] = {
  email: [
    { required: true, message: t('auth.required') },
    { email: true, message: t('auth.emailInvalid') },
  ],
}

function onSubmit({ validateResult }: { validateResult: boolean | Record<string, unknown> }) {
  if (validateResult !== true || loading.value) return
  loading.value = true
  authApi
    .forgot(form.email)
    // 恒成功（后端不暴露邮箱存在性）：无论邮箱是否存在都提示已发送
    .then(() => {
      sent.value = true
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
    <h1 class="auth-card__title">{{ t('forgot.title') }}</h1>
    <template v-if="!sent">
      <p class="auth-card__hint">{{ t('forgot.hint') }}</p>
      <TForm :data="form" :rules="rules" label-width="0" @submit="onSubmit">
        <TFormItem name="email">
          <UiInput v-model="form.email" type="email" :placeholder="t('auth.email')" :aria-label="t('auth.email')" clearable />
        </TFormItem>
        <UiButton theme="primary" type="submit" block :loading="loading">
          {{ t('forgot.submit') }}
        </UiButton>
      </TForm>
    </template>
    <template v-else>
      <p class="auth-card__hint">{{ t('forgot.sent') }}</p>
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
  margin: 0 0 var(--mc-sp-4);
}

.auth-card__hint {
  font-size: var(--mc-fs-sm);
  color: var(--td-text-color-secondary);
  line-height: var(--mc-lh-normal);
  margin: 0 0 var(--mc-sp-4);
}

.auth-card__footer {
  text-align: center;
  font-size: var(--mc-fs-sm);
  color: var(--td-text-color-secondary);
  margin-top: var(--mc-sp-4);
}
</style>
