<script setup lang="ts">
// Shared presentational primitives. Keep dumb — no data fetching here.
defineProps<{
  label?: string
  value?: string | number
  hint?: string
  icon?: string
  tone?: 'neutral' | 'ok' | 'warn' | 'bad' | 'info' | 'accent'
  mono?: boolean
}>()

const toneRing: Record<string, string> = {
  neutral: 'text-ink-400 bg-ink-800/60 border-ink-700',
  ok: 'text-sig-ok bg-[#e9f8ef] border-[#b9e8cd]',
  warn: 'text-acc-600 bg-acc-50 border-acc-300',
  bad: 'text-sig-bad bg-[#fdeded] border-[#f8c9c9]',
  info: 'text-sig-info bg-[#eaf1fe] border-[#c2d6fb]',
  accent: 'text-acc-700 bg-acc-50 border-acc-300',
}
</script>

<template>
  <div class="panel px-4 py-3.5 relative overflow-hidden">
    <div class="flex items-start justify-between gap-3">
      <div class="t-label">{{ label }}</div>
      <span
        v-if="icon"
        class="w-7 h-7 shrink-0 grid place-items-center rounded-md border"
        :class="toneRing[tone || 'neutral']"
      >
        <i :class="icon" class="text-[13px]" />
      </span>
    </div>
    <div
      class="mt-2 text-[26px] leading-none font-bold tracking-tight tabular"
      :class="mono ? 't-mono !text-[22px]' : ''"
    >{{ value ?? '—' }}</div>
    <div v-if="hint" class="mt-1.5 text-[11.5px]" style="color: var(--txt-dim)">{{ hint }}</div>
  </div>
</template>
