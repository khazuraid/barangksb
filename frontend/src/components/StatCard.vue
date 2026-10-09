<script setup lang="ts">
defineProps<{
  label?: string
  value?: string | number
  hint?: string
  icon?: string
  tone?: 'neutral' | 'ok' | 'warn' | 'bad' | 'info' | 'accent'
  mono?: boolean
}>()

const toneRing: Record<string, string> = {
  neutral: 'text-slate-500 dark:text-slate-400 bg-slate-100 dark:bg-slate-800/80 border-slate-200 dark:border-slate-700',
  ok: 'text-emerald-600 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-950/40 border-emerald-200 dark:border-emerald-800',
  warn: 'text-amber-600 dark:text-amber-400 bg-amber-50 dark:bg-amber-950/40 border-amber-200 dark:border-amber-800',
  bad: 'text-rose-600 dark:text-rose-400 bg-rose-50 dark:bg-rose-950/40 border-rose-200 dark:border-rose-800',
  info: 'text-sky-600 dark:text-sky-400 bg-sky-50 dark:bg-sky-950/40 border-sky-200 dark:border-sky-800',
  accent: 'text-indigo-600 dark:text-indigo-400 bg-indigo-50 dark:bg-indigo-950/40 border-indigo-200 dark:border-indigo-800',
}
</script>

<template>
  <div class="panel p-4 relative overflow-hidden transition-all duration-200 hover:shadow-md">
    <div class="flex items-start justify-between gap-3">
      <div class="text-[12px] font-medium tracking-wide uppercase" style="color: var(--txt-dim)">{{ label }}</div>
      <span
        v-if="icon"
        class="w-8 h-8 shrink-0 grid place-items-center rounded-lg border transition-transform duration-200"
        :class="toneRing[tone || 'neutral']"
      >
        <i :class="icon" class="text-[14px]" />
      </span>
    </div>
    <div
      class="mt-2 text-[26px] leading-tight font-bold tracking-tight tabular"
      :class="mono ? 't-mono !text-[22px]' : ''"
      style="color: var(--txt)"
    >{{ value ?? '—' }}</div>
    <div v-if="hint" class="mt-1 text-[12px]" style="color: var(--txt-dim)">{{ hint }}</div>
  </div>
</template>
