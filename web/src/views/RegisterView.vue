<script setup lang="ts">
import { onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Form as TForm, FormItem as TFormItem, MessagePlugin } from 'tdesign-vue-next'
import type { FormProps } from 'tdesign-vue-next'

import { authApi, settingsApi } from '@/api'
import { ApiError } from '@/api/client'
import UiButton from '@/components/ui/UiButton.vue'
import UiInput from '@/components/ui/UiInput.vue'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const router = useRouter()
const auth = useAuthStore()

const form = reactive({ username: '', password: '', confirm: '', nickname: '', email: '', emailCode: '' })
const loading = ref(false)

// 站点开关（公开端点）：决定邮箱/验证码字段是否渲染与必填
const needEmail = ref(true)
const registerClosed = ref(false)

// 发送验证码冷却倒计时（秒）
const cooldown = ref(0)
let cooldownTimer: ReturnType<typeof setInterval> | undefined

onMounted(async () => {
  try {
    const s = await settingsApi.get()
    needEmail.value = s.requireEmailVerify
    registerClosed.value = !s.allowRegister
  } catch {
    /* 读不到开关按默认（需要邮箱验证）渲染 */
  }
})

onBeforeUnmount(() => {
  if (cooldownTimer) clearInterval(cooldownTimer)
})

function startCooldown(seconds: number) {
  cooldown.value = seconds
  cooldownTimer = setInterval(() => {
    cooldown.value -= 1
    if (cooldown.value <= 0 && cooldownTimer) clearInterval(cooldownTimer)
  }, 1000)
}

/** 发送验证码：冷却期内按钮禁用（服务端还有 60s 冷却双保险） */
async function onSendCode() {
  if (!form.email || cooldown.value > 0 || loading.value) return
  try {
    const r = await authApi.sendEmailCode(form.email)
    startCooldown(r.cooldownSeconds)
    MessagePlugin.success(t('auth.codeSent'))
  } catch (e: unknown) {
    MessagePlugin.error(e instanceof ApiError ? e.message : t('auth.failed'))
  }
}

const rules: FormProps['rules'] = {
  username: [
    { required: true, message: t('auth.required') },
    { max: 20, message: t('auth.usernameMax') },
  ],
  email: needEmail.value
    ? [
        { required: true, message: t('auth.required') },
        { email: true, message: t('auth.emailInvalid') },
      ]
    : [],
  emailCode: needEmail.value ? [{ required: true, message: t('auth.required') }] : [],
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

/** TForm 的 submit：Enter 与按钮统一走校验→提交（沿用主仓库 AuthView 模式） */
function onSubmit({ validateResult }: { validateResult: boolean | Record<string, unknown> }) {
  if (validateResult !== true || loading.value) return
  loading.value = true
  auth
    // 注册成功后 store 内自动登录
    .register(form.username, form.password, form.nickname, form.email, form.emailCode)
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
  <div class="mc-container mc-container--narrow auth-wrap">
  <main class="auth-card">
    <h1 class="auth-card__title">{{ t('auth.register') }}</h1>
    <div v-if="registerClosed" class="auth-card__notice">{{ t('auth.registerClosed') }}</div>
    <TForm v-else :data="form" :rules="rules" label-width="0" @submit="onSubmit">
      <TFormItem name="username">
        <UiInput v-model="form.username" :placeholder="t('auth.username')" :aria-label="t('auth.username')" clearable />
      </TFormItem>
      <TFormItem name="nickname">
        <UiInput v-model="form.nickname" :placeholder="t('auth.nickname')" :aria-label="t('auth.nickname')" clearable />
      </TFormItem>
      <TFormItem name="email">
        <div class="code-row">
          <UiInput v-model="form.email" type="email" :placeholder="t('auth.email')" :aria-label="t('auth.email')" clearable />
          <UiButton
            variant="outline"
            size="medium"
            :disabled="cooldown > 0 || !form.email"
            class="code-row__btn"
            @click="onSendCode"
          >
            {{ cooldown > 0 ? t('auth.resendIn', { n: cooldown }) : t('auth.sendCode') }}
          </UiButton>
        </div>
      </TFormItem>
      <TFormItem v-if="needEmail" name="emailCode">
        <UiInput v-model="form.emailCode" :placeholder="t('auth.emailCode')" :aria-label="t('auth.emailCode')" :maxlength="6" />
      </TFormItem>
      <TFormItem name="password">
        <UiInput v-model="form.password" type="password" :placeholder="t('auth.password')" :aria-label="t('auth.password')" clearable />
      </TFormItem>
      <TFormItem name="confirm">
        <UiInput v-model="form.confirm" type="password" :placeholder="t('auth.confirmPassword')" :aria-label="t('auth.confirmPassword')" clearable />
      </TFormItem>
      <UiButton theme="primary" type="submit" block :loading="loading">
        {{ t('auth.register') }}
      </UiButton>
    </TForm>
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

.auth-card__notice {
  text-align: center;
  font-size: var(--mc-fs-body);
  color: var(--td-warning-color);
  margin-bottom: var(--mc-sp-4);
}

/* 验证码行：输入框 + 发送按钮同行 */
.code-row {
  display: flex;
  gap: var(--mc-sp-2);
  width: 100%;
}

.code-row__btn {
  flex-shrink: 0;
  white-space: nowrap;
}

.auth-card__footer {
  text-align: center;
  font-size: var(--mc-fs-sm);
  line-height: var(--mc-lh-normal);
  color: var(--td-text-color-secondary);
  margin-top: var(--mc-sp-4);
}
</style>
