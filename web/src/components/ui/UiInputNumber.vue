<script setup lang="ts">
import { computed, ref, watch } from 'vue'

/** 数字输入（右侧上下步进，对齐主仓库 UiInputNumber） */
const props = withDefaults(
  defineProps<{
    modelValue?: number
    min?: number
    max?: number
    step?: number
    suffix?: string
    disabled?: boolean
  }>(),
  { modelValue: 0, min: -Infinity, max: Infinity, step: 1, disabled: false },
)

const emit = defineEmits<{
  'update:modelValue': [value: number]
  change: [value: number]
}>()

const draft = ref(String(props.modelValue ?? 0))

watch(
  () => props.modelValue,
  (v) => { draft.value = String(v ?? 0) },
)

function commit(v: number) {
  const next = Math.min(props.max, Math.max(props.min, Number.isFinite(v) ? v : 0))
  draft.value = String(next)
  emit('update:modelValue', next)
  emit('change', next)
}

function onInput(e: Event) {
  draft.value = (e.target as HTMLInputElement).value
}

function onBlur() {
  const n = Number(draft.value)
  commit(Number.isFinite(n) ? n : props.min === -Infinity ? 0 : props.min)
}

function bump(dir: 1 | -1) {
  if (props.disabled) return
  commit((props.modelValue ?? 0) + dir * props.step)
}

const wrapCls = computed(() => [
  'flex items-center gap-1 w-full h-8 px-2.5 rounded-md border bg-page transition-colors',
  // 焦点环与全站一致：2px 品牌色描边 + 2px 偏移
  'focus-within:outline-2 focus-within:outline-solid focus-within:outline-brand focus-within:outline-offset-2',
  props.disabled
    ? 'border-line opacity-60'
    : 'border-line-strong hover:border-brand-line-strong focus-within:border-brand',
])
</script>

<template>
  <div :class="wrapCls">
    <input
      :value="draft"
      type="text"
      inputmode="numeric"
      :disabled="disabled"
      class="flex-1 min-w-0 bg-transparent outline-none border-none text-[13px] text-ink tabular-nums"
      @input="onInput"
      @blur="onBlur"
      @keydown.enter="onBlur"
    />
    <span v-if="suffix" class="flex-shrink-0 text-[12px] text-ink-3">{{ suffix }}</span>
    <span class="flex-shrink-0 flex flex-col -my-1">
      <!-- 按钮命中区 24×16：视觉箭头小，但触达面积不能跟着小 -->
      <button
        type="button"
        class="w-6 h-4 flex items-center justify-center border-none bg-transparent text-ink-3 hover:text-brand focus-visible:ring-1 focus-visible:ring-brand-ring"
        :disabled="disabled"
        aria-label="increase"
        @click="bump(1)"
      >
        <span class="text-[8px] leading-none">▲</span>
      </button>
      <button
        type="button"
        class="w-6 h-4 flex items-center justify-center border-none bg-transparent text-ink-3 hover:text-brand focus-visible:ring-1 focus-visible:ring-brand-ring"
        :disabled="disabled"
        aria-label="decrease"
        @click="bump(-1)"
      >
        <span class="text-[8px] leading-none">▼</span>
      </button>
    </span>
  </div>
</template>
