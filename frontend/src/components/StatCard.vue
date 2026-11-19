<script setup lang="ts">
defineProps<{
  label?: string
  value?: string | number
  hint?: string
  icon?: string
  tone?: 'neutral' | 'ok' | 'warn' | 'bad' | 'info' | 'accent'
  mono?: boolean
  featured?: boolean
}>()

const toneChip: Record<string, string> = {
  neutral: 'text-txt-dim bg-paper-2 border-line',
  ok: 'text-sig-ok bg-[var(--sig-ok-bg)] border-[var(--sig-ok-line)]',
  warn: 'text-sig-warn bg-[var(--sig-warn-bg)] border-[var(--sig-warn-line)]',
  bad: 'text-sig-bad bg-[var(--sig-bad-bg)] border-[var(--sig-bad-line)]',
  info: 'text-sig-info bg-[var(--sig-info-bg)] border-[var(--sig-info-line)]',
  accent: 'text-acc-500 bg-[var(--acc-50)] border-[var(--acc-border)]',
}
</script>

<template>
  <div
    class="rounded-[10px] border p-4 flex flex-col justify-between transition-colors duration-150"
    :class="featured
      ? 'bg-[var(--acc-wash)] border-[var(--acc-border)]'
      : 'bg-panel border-line'"
  >
    <div class="flex items-center justify-between gap-2.5">
      <div class="text-[12px] font-medium text-txt-dim truncate">{{ label }}</div>
      <span
        v-if="icon"
        class="w-7 h-7 shrink-0 grid place-items-center rounded-[7px] border"
        :class="toneChip[tone || (featured ? 'accent' : 'neutral')]"
      >
        <i :class="icon" class="text-[12.5px]" aria-hidden="true" />
      </span>
    </div>

    <div
      class="mt-2.5 font-bold tracking-[-0.025em] leading-none tabular"
      :class="[
        featured ? 'text-[30px] text-acc-500' : 'text-[26px] text-txt',
        mono ? 't-mono !text-[20px] !leading-tight' : ''
      ]"
    >
      {{ value ?? '—' }}
    </div>

    <div v-if="hint" class="mt-2 text-[11.5px] text-txt-dim truncate">{{ hint }}</div>
  </div>
</template>
