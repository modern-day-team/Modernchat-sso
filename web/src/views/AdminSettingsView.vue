<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { MessagePlugin } from 'tdesign-vue-next'

import { settingsApi, type SiteSettings } from '@/api'
import { ApiError } from '@/api/client'
import UiButton from '@/components/ui/UiButton.vue'
import UiCheckbox from '@/components/ui/UiCheckbox.vue'
import UiInput from '@/components/ui/UiInput.vue'

const { t } = useI18n()
const router = useRouter()

const form = reactive<SiteSettings>({
  allowRegister: true,
  requireEmailVerify: true,
  requireInviteCode: false,
  maintenanceMode: false,
  maintenanceMessage: '',
  loginMaxAttempts: 5,
  loginLockMinutes: 15,
})
// UiInput 的 modelValue 是 string：数字字段用文本态承接，保存时再解析（避免 .number 与组件类型不匹配）
const attemptsText = ref('5')
const minutesText = ref('15')
const loading = ref(false)
const saving = ref(false)

async function load() {
  loading.value = true
  try {
    Object.assign(form, await settingsApi.getAdmin())
    attemptsText.value = String(form.loginMaxAttempts)
    minutesText.value = String(form.loginLockMinutes)
  } catch (e: unknown) {
    MessagePlugin.error(e instanceof ApiError ? e.message : t('auth.failed'))
  } finally {
    loading.value = false
  }
}

/** 与后端 handler/settings.go 的校验口径一致：失败阈值 0 或 2-20，锁定时长 1-1440。 */
function parseNumbers(): { attempts: number; minutes: number } | null {
  const attempts = Number(attemptsText.value.trim())
  const minutes = Number(minutesText.value.trim())
  if (!Number.isInteger(attempts) || (attempts !== 0 && (attempts < 2 || attempts > 20))) {
    MessagePlugin.warning(t('settings.invalidAttempts'))
    return null
  }
  if (!Number.isInteger(minutes) || minutes < 1 || minutes > 1440) {
    MessagePlugin.warning(t('settings.invalidMinutes'))
    return null
  }
  return { attempts, minutes }
}

async function save() {
  if (saving.value) return
  const nums = parseNumbers()
  if (!nums) return
  saving.value = true
  try {
    const next = await settingsApi.update({
      ...form,
      loginMaxAttempts: nums.attempts,
      loginLockMinutes: nums.minutes,
    })
    Object.assign(form, next)
    attemptsText.value = String(next.loginMaxAttempts)
    minutesText.value = String(next.loginLockMinutes)
    MessagePlugin.success(t('settings.saved'))
  } catch (e: unknown) {
    MessagePlugin.error(e instanceof ApiError ? e.message : t('auth.failed'))
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  void load()
})
</script>

<template>
  <main class="settings-page">
    <div class="settings-head">
      <h1 class="settings-page__title">{{ t('settings.title') }}</h1>
      <UiButton variant="outline" size="small" @click="router.push('/admin')">
        {{ t('settings.back') }}
      </UiButton>
    </div>

    <section class="settings-group">
      <h2 class="settings-group__title">{{ t('settings.groupRegister') }}</h2>
      <div class="settings-field">
        <UiCheckbox v-model="form.allowRegister">
          <span class="settings-field__label">{{ t('settings.allowRegister') }}</span>
          <span class="settings-field__hint">{{ t('settings.allowRegisterHint') }}</span>
        </UiCheckbox>
      </div>
      <div class="settings-field">
        <UiCheckbox v-model="form.requireInviteCode">
          <span class="settings-field__label">{{ t('settings.requireInviteCode') }}</span>
          <span class="settings-field__hint">{{ t('settings.requireInviteCodeHint') }}</span>
        </UiCheckbox>
      </div>
      <div class="settings-field">
        <UiCheckbox v-model="form.requireEmailVerify">
          <span class="settings-field__label">{{ t('settings.requireEmailVerify') }}</span>
          <span class="settings-field__hint">{{ t('settings.requireEmailVerifyHint') }}</span>
        </UiCheckbox>
      </div>
    </section>

    <section class="settings-group">
      <h2 class="settings-group__title">{{ t('settings.groupMaintenance') }}</h2>
      <div class="settings-field">
        <UiCheckbox v-model="form.maintenanceMode">
          <span class="settings-field__label">{{ t('settings.maintenanceMode') }}</span>
          <span class="settings-field__hint">{{ t('settings.maintenanceModeHint') }}</span>
        </UiCheckbox>
      </div>
      <div class="settings-row">
        <label class="settings-row__label" for="maintenance-message">
          {{ t('settings.maintenanceMessage') }}
        </label>
        <UiInput
          id="maintenance-message"
          v-model="form.maintenanceMessage"
          :placeholder="t('settings.maintenanceMessagePlaceholder')"
          :maxlength="200"
        />
        <span class="settings-field__hint">{{ t('settings.maintenanceMessageHint') }}</span>
      </div>
    </section>

    <section class="settings-group">
      <h2 class="settings-group__title">{{ t('settings.groupSecurity') }}</h2>
      <div class="settings-row">
        <label class="settings-row__label" for="login-attempts">
          {{ t('settings.loginMaxAttempts') }}
        </label>
        <UiInput id="login-attempts" v-model="attemptsText" />
        <span class="settings-field__hint">{{ t('settings.loginMaxAttemptsHint') }}</span>
      </div>
      <div class="settings-row">
        <label class="settings-row__label" for="login-minutes">
          {{ t('settings.loginLockMinutes') }}
        </label>
        <UiInput id="login-minutes" v-model="minutesText" />
        <span class="settings-field__hint">{{ t('settings.loginLockMinutesHint') }}</span>
      </div>
    </section>

    <div class="settings-actions">
      <UiButton theme="primary" :loading="saving" :disabled="loading" @click="save">
        {{ t('settings.save') }}
      </UiButton>
    </div>
  </main>
</template>

<style scoped>
.settings-page {
  width: 100%;
  max-width: 640px;
  background: var(--td-bg-color-container);
  border: 1px solid var(--mc-hairline);
  border-radius: var(--mc-radius-lg);
  padding: 24px;
}

.settings-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 20px;
}

.settings-page__title {
  font-size: var(--mc-fs-heading);
  font-weight: 600;
  margin: 0;
}

.settings-group {
  padding: 16px 0;
  border-top: 1px solid var(--mc-hairline);
}

.settings-group__title {
  font-size: var(--mc-fs-emphasis);
  font-weight: 600;
  margin: 0 0 12px;
}

.settings-field {
  margin-bottom: 12px;
}

.settings-field__label {
  display: block;
  font-size: var(--mc-fs-body);
  color: var(--td-text-color-primary);
}

.settings-field__hint {
  display: block;
  font-size: var(--mc-fs-sm);
  color: var(--td-text-color-secondary);
  line-height: 1.5;
}

.settings-row {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 14px;
}

.settings-row__label {
  font-size: var(--mc-fs-body);
  color: var(--td-text-color-primary);
}

.settings-row .ui-input,
.settings-row > :nth-child(2) {
  max-width: 320px;
}

.settings-actions {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
  border-top: 1px solid var(--mc-hairline);
}
</style>
