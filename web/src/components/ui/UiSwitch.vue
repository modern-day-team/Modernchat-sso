<script setup lang="ts">
import { computed } from 'vue'

/** 开关（UnoCSS 实现，对齐主仓库 UiSwitch：value + change） */
const props = withDefaults(
  defineProps<{
    value?: boolean
    disabled?: boolean
    size?: 'small' | 'medium'
  }>(),
  { value: false, disabled: false, size: 'medium' },
)

const emit = defineEmits<{ change: [value: boolean] }>()

const trackCls = computed(() => [
  'relative inline-flex items-center flex-shrink-0 rounded-full border-none overflow-hidden',
  props.size === 'small' ? 'w-8 h-[18px]' : 'w-10 h-[22px]',
  props.value ? 'bg-brand' : 'bg-ink-4',
  props.disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer',
])

const knobCls = computed(() => [
  'ui-switch__knob absolute top-[2px] left-[2px] rounded-full bg-white shadow-sm',
  props.size === 'small' ? 'w-3.5 h-3.5' : 'w-[18px] h-[18px]',
  props.size === 'small'
    ? props.value ? 'translate-x-[14px]' : 'translate-x-0'
    : props.value ? 'translate-x-[18px]' : 'translate-x-0',
])

function toggle() {
  if (props.disabled) return
  emit('change', !props.value)
}
</script>

<template>
  <button
    type="button"
    role="switch"
    :aria-checked="value"
    :disabled="disabled"
    :class="trackCls"
    class="ui-switch__track"
    @click="toggle"
  >
    <span :class="knobCls" />
  </button>
</template>

<style scoped>
.ui-switch__track {
  transition:
    background-color 180ms var(--mc-ease-in-out),
    transform 100ms var(--mc-ease-out);
}

.ui-switch__track:active:not(:disabled) {
  transform: scale(0.96);
}

.ui-switch__knob {
  transition: transform var(--mc-dur-base) var(--mc-ease-spring);
}

@media (prefers-reduced-motion: reduce) {
  .ui-switch__knob {
    transition-duration: 1ms;
  }
}
</style>
