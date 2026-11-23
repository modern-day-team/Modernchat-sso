<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { MessagePlugin } from 'tdesign-vue-next'

import { settingsApi, type SiteSettings } from '@/api'
import { ApiError } from '@/api/client'
import UiButton from '@/components/ui/UiButton.vue'
import UiInput from '@/components/ui/UiInput.vue'
import UiInputNumber from '@/components/ui/UiInputNumber.vue'
import UiSwitch from '@/components/ui/UiSwitch.vue'

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
const loading = ref(true)
const saving = ref(false)
const dirty = ref(false)

async function load() {
  loading.value = true
  try {
    Object.assign(form, await settingsApi.getAdmin())
    dirty.value = false
  } catch (e: unknown) {
    MessagePlugin.error(e instanceof ApiError ? e.message : t('auth.failed'))
  } finally {
    loading.value = false
  }
}

function set<K extends keyof SiteSettings>(key: K, value: SiteSettings[K]) {
  form[key] = value
  dirty.value = true
}

/** 与后端 handler/settings.go 的校验口径一致：失败阈值 0 或 2-20，锁定时长 1-1440。 */
function validate(): boolean {
  if (form.loginMaxAttempts !== 0 && (form.loginMaxAttempts < 2 || form.loginMaxAttempts > 20)) {
    MessagePlugin.warning(t('settings.invalidAttempts'))
    return false
  }
  if (form.loginLockMinutes < 1 || form.loginLockMinutes > 1440) {
    MessagePlugin.warning(t('settings.invalidMinutes'))
    return false
  }
  return true
}

async function save() {
  if (saving.value || !dirty.value || !validate()) return
  saving.value = true
  try {
    Object.assign(form, await settingsApi.update({ ...form }))
    dirty.value = false
    MessagePlugin.success(t('settings.saved'))
  } catch (e: unknown) {
    MessagePlugin.error(e instanceof ApiError ? e.message : t('auth.failed'))
  } finally {
    saving.value = false
  }
}

const saveLabel = computed(() => (dirty.value ? t('settings.saveChanges') : t('settings.savedState')))

onMounted(() => {
  void load()
})
</script>

<template>
  <main class="settings-page">
    <header class="settings-head">
      <div>
        <h1 class="settings-head__title">{{ t('settings.title') }}</h1>
        <p class="settings-head__subtitle">{{ t('settings.subtitle') }}</p>
      </div>
      <UiButton variant="outline" size="small" @click="router.push('/admin')">
        {{ t('settings.back') }}
      </UiButton>
    </header>

    <div v-if="loading" class="settings-loading">
      <div v-for="n in 3" :key="n" class="settings-skeleton" />
    </div>

    <template v-else>
      <section class="settings-group">
        <div class="group-head">
          <span class="group-head__icon group-head__icon--blue" aria-hidden="true">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">
              <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" /><circle cx="9" cy="7" r="4" /><path d="M19 8v6M22 11h-6" />
            </svg>
          </span>
          <div>
            <h2 class="group-head__title">{{ t('settings.groupRegister') }}</h2>
            <p class="group-head__desc">{{ t('settings.groupRegisterDesc') }}</p>
          </div>
        </div>

        <div class="setting-row">
          <div class="setting-row__info">
            <h3>{{ t('settings.allowRegister') }}</h3>
            <p>{{ t('settings.allowRegisterHint') }}</p>
          </div>
          <UiSwitch :value="form.allowRegister" @change="(v: boolean) => set('allowRegister', v)" />
        </div>
        <div class="setting-row">
          <div class="setting-row__info">
            <h3>{{ t('settings.requireInviteCode') }}</h3>
            <p>{{ t('settings.requireInviteCodeHint') }}</p>
          </div>
          <UiSwitch :value="form.requireInviteCode" @change="(v: boolean) => set('requireInviteCode', v)" />
        </div>
        <div class="setting-row">
          <div class="setting-row__info">
            <h3>{{ t('settings.requireEmailVerify') }}</h3>
            <p>{{ t('settings.requireEmailVerifyHint') }}</p>
          </div>
          <UiSwitch :value="form.requireEmailVerify" @change="(v: boolean) => set('requireEmailVerify', v)" />
        </div>
      </section>

      <section class="settings-group">
        <div class="group-head">
          <span class="group-head__icon group-head__icon--amber" aria-hidden="true">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">
              <path d="M12 9v4M12 17h.01" /><path d="M10.3 3.9 1.8 18a2 2 0 0 0 1.7 3h17a2 2 0 0 0 1.7-3L13.7 3.9a2 2 0 0 0-3.4 0Z" />
            </svg>
          </span>
          <div>
            <h2 class="group-head__title">{{ t('settings.groupMaintenance') }}</h2>
            <p class="group-head__desc">{{ t('settings.groupMaintenanceDesc') }}</p>
          </div>
        </div>

        <div class="setting-row">
          <div class="setting-row__info">
            <h3>{{ t('settings.maintenanceMode') }}</h3>
            <p>{{ t('settings.maintenanceModeHint') }}</p>
          </div>
          <UiSwitch :value="form.maintenanceMode" @change="(v: boolean) => set('maintenanceMode', v)" />
        </div>
        <div v-if="form.maintenanceMode" class="settings-warning">
          <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" aria-hidden="true">
            <circle cx="12" cy="12" r="9" /><path d="M12 8v5M12 16h.01" />
          </svg>
          <span>{{ t('settings.maintenanceWarning') }}</span>
        </div>
        <div class="setting-field">
          <label class="setting-field__label" for="maintenance-message">{{ t('settings.maintenanceMessage') }}</label>
          <UiInput
            id="maintenance-message"
            :model-value="form.maintenanceMessage"
            :placeholder="t('settings.maintenanceMessagePlaceholder')"
            :maxlength="200"
            @update:model-value="(v: string) => set('maintenanceMessage', v)"
          />
          <p class="setting-field__hint">{{ t('settings.maintenanceMessageHint') }}</p>
        </div>
      </section>

      <section class="settings-group">
        <div class="group-head">
          <span class="group-head__icon group-head__icon--violet" aria-hidden="true">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">
              <rect x="4" y="10" width="16" height="10" rx="2" /><path d="M8 10V7a4 4 0 0 1 8 0v3" />
            </svg>
          </span>
          <div>
            <h2 class="group-head__title">{{ t('settings.groupSecurity') }}</h2>
            <p class="group-head__desc">{{ t('settings.groupSecurityDesc') }}</p>
          </div>
        </div>

        <div class="setting-field setting-field--inline">
          <div class="setting-field__text">
            <label class="setting-field__label" for="login-attempts">{{ t('settings.loginMaxAttempts') }}</label>
            <p class="setting-field__hint">{{ t('settings.loginMaxAttemptsHint') }}</p>
          </div>
          <UiInputNumber
            :model-value="form.loginMaxAttempts"
            :min="0"
            :max="20"
            suffix=""
            class="setting-field__number"
            @update:model-value="(v: number) => set('loginMaxAttempts', v)"
          />
        </div>
        <div class="setting-field setting-field--inline">
          <div class="setting-field__text">
            <label class="setting-field__label" for="login-minutes">{{ t('settings.loginLockMinutes') }}</label>
            <p class="setting-field__hint">{{ t('settings.loginLockMinutesHint') }}</p>
          </div>
          <UiInputNumber
            :model-value="form.loginLockMinutes"
            :min="1"
            :max="1440"
            suffix=""
            class="setting-field__number"
            @update:model-value="(v: number) => set('loginLockMinutes', v)"
          />
        </div>
      </section>

      <div class="settings-actions">
        <span class="settings-actions__status" :class="{ 'is-dirty': dirty }">
          <span class="settings-actions__dot" />
          {{ saveLabel }}
        </span>
        <UiButton theme="primary" :loading="saving" :disabled="!dirty" @click="save">
          {{ t('settings.save') }}
        </UiButton>
      </div>
    </template>
  </main>
</template>

<style scoped>
.settings-page {
  width: 100%;
  max-width: 720px;
  background: var(--td-bg-color-container);
  border: 1px solid var(--mc-hairline);
  border-radius: var(--mc-radius-lg);
  padding: 28px 32px 24px;
  box-shadow: var(--mc-elev-1);
}

.settings-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding-bottom: 20px;
  border-bottom: 1px solid var(--mc-hairline);
}

.settings-head__title {
  font-size: var(--mc-fs-heading);
  font-weight: 600;
  letter-spacing: -0.02em;
  margin: 0;
}

.settings-head__subtitle {
  font-size: var(--mc-fs-sm);
  color: var(--td-text-color-secondary);
  margin: 6px 0 0;
}

.settings-group {
  padding: 22px 0 8px;
  border-bottom: 1px solid var(--mc-hairline);
}

.group-head {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.group-head__icon {
  width: 34px;
  height: 34px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--mc-radius-md);
}

.group-head__icon--blue { background: var(--mc-brand-soft); color: var(--td-brand-color); }
.group-head__icon--amber { background: var(--mc-warning-soft); color: var(--td-warning-color); }
.group-head__icon--violet { background: var(--mc-brand-soft); color: var(--td-brand-color); }

.group-head__title {
  font-size: var(--mc-fs-title);
  font-weight: 600;
  margin: 0;
}

.group-head__desc {
  font-size: var(--mc-fs-sm);
  color: var(--td-text-color-secondary);
  margin: 2px 0 0;
}

.setting-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  padding: 12px 0;
}

.setting-row + .setting-row {
  border-top: 1px solid var(--mc-hairline);
}

.setting-row__info { min-width: 0; }

.setting-row__info h3 {
  font-size: var(--mc-fs-body);
  font-weight: 500;
  margin: 0 0 4px;
}

.setting-row__info p,
.setting-field__hint {
  font-size: var(--mc-fs-sm);
  color: var(--td-text-color-secondary);
  line-height: 1.5;
  margin: 0;
}

.settings-warning {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  margin-bottom: 12px;
  border-radius: var(--mc-radius-md);
  background: var(--mc-warning-soft);
  color: var(--td-warning-color);
  font-size: var(--mc-fs-sm);
}

.setting-field {
  padding: 12px 0;
}

.setting-field__label {
  display: block;
  font-size: var(--mc-fs-body);
  font-weight: 500;
  margin-bottom: 8px;
}

.setting-field__hint { margin-top: 6px; }

.setting-field--inline {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
}

.setting-field__text { min-width: 0; }
.setting-field__number { width: 120px; flex-shrink: 0; }

.settings-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding-top: 20px;
}

.settings-actions__status {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: var(--mc-fs-sm);
  color: var(--td-text-color-placeholder);
}

.settings-actions__dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--td-success-color);
}

.settings-actions__status.is-dirty { color: var(--td-warning-color); }
.settings-actions__status.is-dirty .settings-actions__dot { background: var(--td-warning-color); }

.settings-loading { display: grid; gap: 16px; padding: 28px 0; }
.settings-skeleton {
  height: 88px;
  border-radius: var(--mc-radius-md);
  background: linear-gradient(90deg, var(--td-bg-color-secondarycontainer) 25%, var(--td-bg-color-container-hover) 50%, var(--td-bg-color-secondarycontainer) 75%);
  background-size: 200% 100%;
  animation: skeleton-shine 1.4s ease-in-out infinite;
}

@keyframes skeleton-shine {
  from { background-position: 200% 0; }
  to { background-position: -200% 0; }
}

@media (prefers-reduced-motion: reduce) {
  .settings-skeleton { animation: none; }
}

@media (max-width: 640px) {
  .settings-page { padding: 20px; }
  .setting-field--inline { align-items: flex-start; }
  .setting-field__number { width: 96px; }
}
</style>
