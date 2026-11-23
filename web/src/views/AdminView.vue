<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { MessagePlugin } from 'tdesign-vue-next'

import { adminApi } from '@/api'
import { ApiError } from '@/api/client'
import UiButton from '@/components/ui/UiButton.vue'
import UiDialog from '@/components/ui/UiDialog.vue'
import UiInput from '@/components/ui/UiInput.vue'
import UiTag from '@/components/ui/UiTag.vue'
import type { User } from '@/types'

const { t } = useI18n()
const router = useRouter()

const users = ref<User[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const kw = ref('')
const loading = ref(false)

const pageAdmins = computed(() => users.value.filter((u) => u.role === 'admin').length)
const pageBanned = computed(() => users.value.filter(isBanned).length)
const totalPages = computed(() => Math.ceil(total.value / pageSize) || 1)

/** 拉取用户列表 */
async function load() {
  loading.value = true
  try {
    const r = await adminApi.listUsers({ page: page.value, pageSize, kw: kw.value })
    users.value = r.users
    total.value = r.total
  } catch (e: unknown) {
    MessagePlugin.error(e instanceof ApiError ? e.message : t('auth.failed'))
  } finally {
    loading.value = false
  }
}

function search() {
  page.value = 1
  void load()
}

function isBanned(u: User): boolean {
  return !!u.bannedUntil && new Date(u.bannedUntil).getTime() > Date.now()
}

// 封禁对话框：天数（0=永久）与原因一次填完，替换浏览器原生 prompt
const banVisible = ref(false)
const banTarget = ref<User | null>(null)
const banDaysText = ref('0')
const banReason = ref('')
const banSaving = ref(false)

function onBan(u: User) {
  banTarget.value = u
  banDaysText.value = '0'
  banReason.value = ''
  banVisible.value = true
}

function onBanConfirm() {
  const u = banTarget.value
  if (!u || banSaving.value) return
  const days = Number(banDaysText.value.trim())
  if (!Number.isInteger(days) || days < 0 || days > 3650) {
    MessagePlugin.warning(t('admin.banInvalid'))
    return
  }
  banSaving.value = true
  adminApi
    .ban(u.id, days, banReason.value.trim())
    .then(() => {
      banVisible.value = false
      MessagePlugin.success(t('admin.opDone'))
      void load()
    })
    .catch((e: unknown) => {
      MessagePlugin.error(e instanceof ApiError ? e.message : t('auth.failed'))
    })
    .finally(() => {
      banSaving.value = false
    })
}

function onUnban(u: User) {
  adminApi
    .unban(u.id)
    .then(() => {
      MessagePlugin.success(t('admin.opDone'))
      void load()
    })
    .catch((e: unknown) => {
      MessagePlugin.error(e instanceof ApiError ? e.message : t('auth.failed'))
    })
}

function onResetPassword(u: User) {
  adminApi
    .resetPassword(u.id)
    .then((r) => {
      MessagePlugin.success(t('admin.newPwd', { pwd: r.newPassword }))
    })
    .catch((e: unknown) => {
      MessagePlugin.error(e instanceof ApiError ? e.message : t('auth.failed'))
    })
}

function onToggleRole(u: User) {
  const role = u.role === 'admin' ? 'user' : 'admin'
  adminApi
    .setRole(u.id, role)
    .then(() => {
      MessagePlugin.success(t('admin.opDone'))
      void load()
    })
    .catch((e: unknown) => {
      MessagePlugin.error(e instanceof ApiError ? e.message : t('auth.failed'))
    })
}

onMounted(() => {
  void load()
})
</script>

<template>
  <main class="admin-page">
    <header class="admin-head">
      <div>
        <h1 class="admin-head__title">{{ t('admin.title') }}</h1>
        <p class="admin-head__subtitle">{{ t('admin.subtitle') }}</p>
      </div>
      <UiButton variant="outline" size="small" @click="router.push('/admin/settings')">
        <template #icon>
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <circle cx="12" cy="12" r="3" /><path d="M19.4 15a1.7 1.7 0 0 0 .3 1.9l.1.1-2.8 2.8-.1-.1a1.7 1.7 0 0 0-1.9-.3 1.7 1.7 0 0 0-1 1.5V21h-4v-.1a1.7 1.7 0 0 0-1-1.5 1.7 1.7 0 0 0-1.9.3l-.1.1-2.8-2.8.1-.1a1.7 1.7 0 0 0 .3-1.9 1.7 1.7 0 0 0-1.5-1H3v-4h.1a1.7 1.7 0 0 0 1.5-1 1.7 1.7 0 0 0-.3-1.9l-.1-.1 2.8-2.8.1.1a1.7 1.7 0 0 0 1.9.3H10a1.7 1.7 0 0 0 1-1.5V3h4v.1a1.7 1.7 0 0 0 1 1.5 1.7 1.7 0 0 0 1.9-.3l.1-.1 2.8 2.8-.1.1a1.7 1.7 0 0 0-.3 1.9V10a1.7 1.7 0 0 0 1.5 1H21v4h-.1a1.7 1.7 0 0 0-1.5 1Z" />
          </svg>
        </template>
        {{ t('admin.siteSettings') }}
      </UiButton>
    </header>

    <div class="admin-stats">
      <div class="stat-card">
        <span class="stat-card__value">{{ total }}</span>
        <span class="stat-card__label">{{ t('admin.totalUsers') }}</span>
      </div>
      <div class="stat-card">
        <span class="stat-card__value">{{ pageAdmins }}</span>
        <span class="stat-card__label">{{ t('admin.pageAdmins') }}</span>
      </div>
      <div class="stat-card">
        <span class="stat-card__value" :class="{ 'is-alert': pageBanned > 0 }">{{ pageBanned }}</span>
        <span class="stat-card__label">{{ t('admin.pageBanned') }}</span>
      </div>
    </div>

    <div class="admin-toolbar">
      <UiInput v-model="kw" :placeholder="t('admin.searchPlaceholder')" clearable @enter="search">
        <template #prefix-icon>
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" aria-hidden="true">
            <circle cx="11" cy="11" r="7" /><path d="m16 16 4 4" />
          </svg>
        </template>
      </UiInput>
      <UiButton theme="primary" @click="search">{{ t('admin.search') }}</UiButton>
      <UiButton variant="outline" :loading="loading" @click="load">{{ t('admin.refresh') }}</UiButton>
    </div>

    <div class="admin-table-wrap">
      <table class="admin-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>{{ t('admin.colUser') }}</th>
            <th>{{ t('admin.colEmail') }}</th>
            <th>{{ t('admin.colRole') }}</th>
            <th>{{ t('admin.colStatus') }}</th>
            <th>{{ t('admin.colJoined') }}</th>
            <th class="col-actions">{{ t('admin.colActions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="u in users" :key="u.id">
            <td class="num">{{ u.id }}</td>
            <td>
              <div class="user-cell">
                <span class="user-cell__avatar">{{ u.nickname?.slice(0, 1) || '?' }}</span>
                <span class="user-cell__meta">
                  <b>{{ u.nickname }}</b>
                  <small>@{{ u.username }}</small>
                </span>
              </div>
            </td>
            <td class="email-cell">{{ u.email || '—' }}</td>
            <td>
              <UiTag :theme="u.role === 'admin' ? 'primary' : 'default'">{{ u.role }}</UiTag>
            </td>
            <td>
              <UiTag :theme="isBanned(u) ? 'danger' : 'success'">
                {{ isBanned(u) ? t('admin.banned') : t('admin.active') }}
              </UiTag>
            </td>
            <td class="num">{{ new Date(u.createdAt).toLocaleDateString() }}</td>
            <td>
              <div class="row-actions">
                <UiButton v-if="u.role !== 'admin'" variant="text" size="small" theme="danger" @click="onBan(u)">
                  {{ t('admin.ban') }}
                </UiButton>
                <UiButton v-if="isBanned(u)" variant="text" size="small" @click="onUnban(u)">
                  {{ t('admin.unban') }}
                </UiButton>
                <UiButton variant="text" size="small" @click="onResetPassword(u)">
                  {{ t('admin.resetPwd') }}
                </UiButton>
                <UiButton variant="text" size="small" @click="onToggleRole(u)">
                  {{ u.role === 'admin' ? t('admin.setUser') : t('admin.setAdmin') }}
                </UiButton>
              </div>
            </td>
          </tr>
          <tr v-if="!loading && users.length === 0">
            <td colspan="7" class="admin-empty">{{ t('admin.empty') }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="admin-pager">
      <span class="admin-pager__info">{{ t('admin.pageInfo', { page, total: totalPages }) }}</span>
      <div class="admin-pager__buttons">
        <UiButton size="small" :disabled="page <= 1" @click="page -= 1; load()">{{ t('admin.prev') }}</UiButton>
        <UiButton size="small" :disabled="page * pageSize >= total" @click="page += 1; load()">{{ t('admin.next') }}</UiButton>
      </div>
    </div>

    <UiDialog
      v-model:visible="banVisible"
      :header="t('admin.banTitle', { name: banTarget?.nickname ?? '' })"
      width="420"
      :confirm-btn="{ content: t('admin.ban'), theme: 'danger' }"
      :confirm-btn-loading="banSaving"
      @confirm="onBanConfirm"
    >
      <div class="ban-form">
        <div class="ban-form__warning">
          <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" aria-hidden="true">
            <circle cx="12" cy="12" r="9" /><path d="M12 8v5M12 16h.01" />
          </svg>
          <span>{{ t('admin.banWarning') }}</span>
        </div>
        <label class="ban-form__label" for="ban-days">{{ t('admin.banPrompt') }}</label>
        <UiInput id="ban-days" v-model="banDaysText" />
        <label class="ban-form__label" for="ban-reason">{{ t('admin.reasonPrompt') }}</label>
        <UiInput id="ban-reason" v-model="banReason" :maxlength="200" />
        <p class="ban-form__hint">{{ t('admin.banInvalid') }}</p>
      </div>
    </UiDialog>
  </main>
</template>

<style scoped>
.admin-page {
  width: 100%;
  max-width: 1000px;
  background: var(--td-bg-color-container);
  border: 1px solid var(--mc-hairline);
  border-radius: var(--mc-radius-lg);
  padding: 28px 32px 24px;
  box-shadow: var(--mc-elev-1);
}

.admin-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding-bottom: 20px;
  border-bottom: 1px solid var(--mc-hairline);
}

.admin-head__title {
  font-size: var(--mc-fs-heading);
  font-weight: 600;
  letter-spacing: -0.02em;
  margin: 0;
}

.admin-head__subtitle {
  font-size: var(--mc-fs-sm);
  color: var(--td-text-color-secondary);
  margin: 6px 0 0;
}

.admin-stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin: 20px 0;
}

.stat-card {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 14px 16px;
  border: 1px solid var(--mc-hairline);
  border-radius: var(--mc-radius-md);
  background: var(--td-bg-color-secondarycontainer);
}

.stat-card__value {
  font-size: var(--mc-fs-display);
  font-weight: 600;
  letter-spacing: -0.02em;
  color: var(--td-text-color-primary);
  line-height: 1.1;
}

.stat-card__value.is-alert { color: var(--td-error-color); }

.stat-card__label {
  font-size: var(--mc-fs-sm);
  color: var(--td-text-color-secondary);
}

.admin-toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
}

.admin-toolbar > :first-child {
  flex: 1;
  max-width: 360px;
}

.admin-table-wrap {
  border: 1px solid var(--mc-hairline);
  border-radius: var(--mc-radius-md);
  overflow: hidden;
}

.admin-table {
  width: 100%;
  border-collapse: collapse;
  font-size: var(--mc-fs-sm);
}

.admin-table th {
  text-align: left;
  padding: 10px 12px;
  background: var(--td-bg-color-secondarycontainer);
  color: var(--td-text-color-secondary);
  font-weight: 500;
  white-space: nowrap;
}

.admin-table td {
  padding: 12px;
  border-top: 1px solid var(--mc-hairline);
  vertical-align: middle;
}

.admin-table tbody tr:hover td {
  background: var(--td-bg-color-container-hover);
}

.col-actions { text-align: right; }
.num { font-variant-numeric: tabular-nums; }
.email-cell { color: var(--td-text-color-secondary); }

.user-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}

.user-cell__avatar {
  width: 30px;
  height: 30px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--mc-radius-md);
  background: var(--mc-brand-soft);
  color: var(--td-brand-color);
  font-size: var(--mc-fs-sm);
  font-weight: 600;
}

.user-cell__meta {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.user-cell__meta b {
  font-size: var(--mc-fs-body);
  font-weight: 500;
  color: var(--td-text-color-primary);
}

.user-cell__meta small {
  font-size: var(--mc-fs-sm);
  color: var(--td-text-color-placeholder);
}

.row-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 4px;
}

.admin-empty {
  text-align: center;
  padding: 40px 12px !important;
  color: var(--td-text-color-placeholder);
}

.admin-pager {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-top: 16px;
  font-size: var(--mc-fs-sm);
  color: var(--td-text-color-secondary);
}

.admin-pager__buttons {
  display: flex;
  gap: 8px;
}

.ban-form {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 4px 0 8px;
}

.ban-form__warning {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  margin-bottom: 6px;
  border-radius: var(--mc-radius-md);
  background: var(--mc-warning-soft);
  color: var(--td-warning-color);
  font-size: var(--mc-fs-sm);
}

.ban-form__label {
  font-size: var(--mc-fs-body);
  color: var(--td-text-color-primary);
}

.ban-form__hint {
  margin: 0;
  font-size: var(--mc-fs-sm);
  color: var(--td-text-color-secondary);
}

@media (max-width: 720px) {
  .admin-page { padding: 20px; }
  .admin-stats { grid-template-columns: 1fr; }
  .admin-table-wrap { overflow-x: auto; }
  .admin-table { min-width: 760px; }
}
</style>
