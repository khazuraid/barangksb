<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ kind: string; label?: string }>()

type Tone = 'ok' | 'warn' | 'bad' | 'info' | 'neutral'

const MAP: Record<string, { tone: Tone; icon?: string; text: string }> = {
  IN:                { tone: 'ok',      icon: 'pi pi-arrow-down', text: 'Masuk' },
  OUT:               { tone: 'bad',     icon: 'pi pi-arrow-up',   text: 'Keluar' },
  'ADJUST+':         { tone: 'warn',    icon: 'pi pi-sliders-h',  text: 'Opname +' },
  'ADJUST-':         { tone: 'warn',    icon: 'pi pi-sliders-h',  text: 'Opname −' },
  INSERT:            { tone: 'ok',      text: 'Insert' },
  UPDATE:            { tone: 'info',    text: 'Update' },
  DELETE:            { tone: 'bad',     text: 'Delete' },
  admin:             { tone: 'warn',    icon: 'pi pi-shield',     text: 'Admin' },
  petugas:           { tone: 'neutral', icon: 'pi pi-user',       text: 'Petugas' },
  'Baik (B)':        { tone: 'ok',      text: 'Baik (B)' },
  'Kurang Baik (KB)':{ tone: 'warn',    text: 'Kurang Baik (KB)' },
  'Rusak Berat (RB)':{ tone: 'bad',     text: 'Rusak Berat (RB)' },
  Berfungsi:         { tone: 'ok',      text: 'Berfungsi' },
  'Rusak Ringan':    { tone: 'warn',    text: 'Rusak Ringan' },
  'Rusak Berat':     { tone: 'bad',     text: 'Rusak Berat' },
  'Perlu Kalibrasi': { tone: 'info',    text: 'Kalibrasi' },
  PENDING:           { tone: 'warn',    icon: 'pi pi-clock',        text: 'Menunggu' },
  IN_PROGRESS:       { tone: 'info',    icon: 'pi pi-wrench',       text: 'Diproses' },
  RESOLVED:          { tone: 'ok',      icon: 'pi pi-check',        text: 'Selesai' },
  REJECTED:          { tone: 'bad',     icon: 'pi pi-times',        text: 'Ditolak' },
  BURST_ATTACK:      { tone: 'warn',    icon: 'pi pi-bolt',         text: 'Burst Attack' },
  BRUTE_FORCE:       { tone: 'bad',     icon: 'pi pi-key',          text: 'Brute Force' },
  BLOCKED:           { tone: 'bad',     icon: 'pi pi-ban',          text: 'Diblokir' },
  ACTIVE:            { tone: 'ok',      text: 'Aktif' },
  INACTIVE:          { tone: 'neutral', text: 'Nonaktif' },
}

const TONE_CLASS: Record<Tone, string> = {
  ok:      'text-sig-ok bg-[var(--sig-ok-bg)] border-[var(--sig-ok-line)]',
  warn:    'text-sig-warn bg-[var(--sig-warn-bg)] border-[var(--sig-warn-line)]',
  bad:     'text-sig-bad bg-[var(--sig-bad-bg)] border-[var(--sig-bad-line)]',
  info:    'text-sig-info bg-[var(--sig-info-bg)] border-[var(--sig-info-line)]',
  neutral: 'text-txt-dim bg-panel-2 border-line',
}

const DOT_CLASS: Record<Tone, string> = {
  ok: 'bg-sig-ok',
  warn: 'bg-sig-warn',
  bad: 'bg-sig-bad',
  info: 'bg-sig-info',
  neutral: 'bg-txt-muted',
}

const meta = computed(() => MAP[props.kind] ?? { tone: 'neutral' as Tone, text: props.kind || '—' })
</script>

<template>
  <span
    class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-[6px] border text-[11px] font-semibold leading-tight whitespace-nowrap"
    :class="TONE_CLASS[meta.tone]"
  >
    <i v-if="meta.icon" :class="meta.icon" class="text-[10px]" aria-hidden="true" />
    <span v-else class="w-1.5 h-1.5 rounded-full shrink-0" :class="DOT_CLASS[meta.tone]" aria-hidden="true" />
    <span>{{ label || meta.text }}</span>
  </span>
</template>
