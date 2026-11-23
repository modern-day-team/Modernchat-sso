<script setup lang="ts">
import { onMounted, ref } from 'vue'
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
    <div class="settings-head">
      <h1 class="admin-page__title">{{ t('admin.title') }}</h1>
      <UiButton variant="outline" size="small" @click="router.push('/admin/settings')">
        {{ t('admin.siteSettings') }}
      </UiButton>
    </div>
    <div class="admin-toolbar">
      <UiInput v-model="kw" :placeholder="t('admin.searchPlaceholder')" clearable @enter="search" />
      <UiButton variant="outline" @click="search">{{ t('admin.search') }}</UiButton>
      <UiButton variant="outline" @click="load">{{ t('admin.refresh') }}</UiButton>
    </div>
    <table class="admin-table">
      <thead>
        <tr>
          <th>ID</th>
          <th>{{ t('admin.colUser') }}</th>
          <th>{{ t('admin.colEmail') }}</th>
          <th>{{ t('admin.colRole') }}</th>
          <th>{{ t('admin.colStatus') }}</th>
          <th>{{ t('admin.colJoined') }}</th>
          <th>{{ t('admin.colActions') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="u in users" :key="u.id">
          <td class="num">{{ u.id }}</td>
          <td>{{ u.nickname }}（@{{ u.username }}）</td>
          <td>{{ u.email || '—' }}</td>
          <td>
            <UiTag :theme="u.role === 'admin' ? 'primary' : 'default'">{{ u.role }}</UiTag>
          </td>
          <td>
            <UiTag :theme="isBanned(u) ? 'danger' : 'success'">
              {{ isBanned(u) ? t('admin.banned') : t('admin.active') }}
            </UiTag>
          </td>
          <td class="num">{{ new Date(u.createdAt).toLocaleDateString() }}</td>
          <td class="actions">
            <UiButton v-if="u.role !== 'admin'" variant="text" theme="danger" @click="onBan(u)">
              {{ t('admin.ban') }}
            </UiButton>
            <UiButton v-if="isBanned(u)" variant="text" @click="onUnban(u)">
              {{ t('admin.unban') }}
            </UiButton>
            <UiButton variant="text" @click="onResetPassword(u)">
              {{ t('admin.resetPwd') }}
            </UiButton>
            <UiButton variant="text" @click="onToggleRole(u)">
              {{ u.role === 'admin' ? t('admin.setUser') : t('admin.setAdmin') }}
            </UiButton>
          </td>
        </tr>
      </tbody>
    </table>
    <div class="admin-pager">
      <UiButton size="small" :disabled="page <= 1" @click="page -= 1; load()">
        {{ t('admin.prev') }}
      </UiButton>
      <span class="admin-pager__info">
        {{ t('admin.pageInfo', { page, total: Math.ceil(total / pageSize) || 1 }) }}
      </span>
      <UiButton size="small" :disabled="page * pageSize >= total" @click="page += 1; load()">
        {{ t('admin.next') }}
      </UiButton>
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
  max-width: 960px;
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
  margin-bottom: 16px;
}

.admin-page__title {
  font-size: var(--mc-fs-heading);
  font-weight: 600;
  margin: 0;
}

.admin-toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}

.admin-toolbar > :first-child {
  flex: 1;
}

.admin-table {
  width: 100%;
  border-collapse: collapse;
  font-size: var(--mc-fs-sm);
}

.admin-table th,
.admin-table td {
  text-align: left;
  padding: 8px 10px;
  border-bottom: 1px solid var(--mc-hairline);
}

.admin-table th {
  color: var(--td-text-color-secondary);
  font-weight: 500;
}

.num {
  font-variant-numeric: tabular-nums;
}

.actions {
  display: flex;
  gap: 4px;
}

.admin-pager {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 12px;
  font-size: var(--mc-fs-sm);
  color: var(--td-text-color-secondary);
}

.ban-form {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 4px 0 8px;
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
</style>
