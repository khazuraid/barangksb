<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  icon?: string
  title: string
  sub?: string
  tone?: 'neutral' | 'ok' | 'warn' | 'bad' | 'info'
  loading?: boolean
}>()

const isLoading = computed(() => props.loading || !!props.icon?.includes('pi-spinner'))

const iconToneClass = computed(() => {
  switch (props.tone) {
    case 'ok':   return 'text-sig-ok bg-[var(--sig-ok-bg)] border-[var(--sig-ok-line)]'
    case 'warn': return 'text-sig-warn bg-[var(--sig-warn-bg)] border-[var(--sig-warn-line)]'
    case 'bad':  return 'text-sig-bad bg-[var(--sig-bad-bg)] border-[var(--sig-bad-line)]'
    case 'info': return 'text-sig-info bg-[var(--sig-info-bg)] border-[var(--sig-info-line)]'
    default:     return 'text-txt-dim bg-panel-2 border-line'
  }
})
</script>

<template>
  <!-- Skeleton state replaces bare spinners across all data tables & panels -->
  <div v-if="isLoading" class="p-4 space-y-2.5" role="status" aria-live="polite">
    <div v-for="n in 5" :key="n" class="flex items-center gap-3 py-1.5">
      <div class="skeleton h-4 w-20 shrink-0" />
      <div class="skeleton h-4 flex-1" :style="{ maxWidth: `${68 - (n % 3) * 14}%` }" />
      <div class="skeleton h-4 w-16 shrink-0 hidden sm:block" />
      <div class="skeleton h-5 w-14 shrink-0" />
    </div>
    <span class="sr-only">{{ title }}</span>
  </div>

  <!-- Composed empty state -->
  <div v-else class="grid-bg grid place-items-center px-6 py-12 text-center">
    <div
      class="w-10 h-10 grid place-items-center rounded-[9px] border mb-3"
      :class="iconToneClass"
    >
      <i :class="icon || 'pi pi-inbox'" class="text-[15px]" aria-hidden="true" />
    </div>
    <div class="text-[13.5px] font-semibold text-txt">{{ title }}</div>
    <div v-if="sub" class="text-[12px] mt-1 max-w-sm text-txt-dim leading-relaxed">{{ sub }}</div>
    <div v-if="$slots.default" class="mt-3.5"><slot /></div>
  </div>
</template>
