<script setup lang="ts">
import Tag from 'primevue/tag'

const props = defineProps<{ kind: string; label?: string }>()

const MAP: Record<string, { sev: 'success' | 'info' | 'warn' | 'danger' | 'secondary'; icon?: string; text?: string }> = {
  IN:       { sev: 'success', icon: 'pi pi-arrow-down',  text: 'MASUK' },
  OUT:      { sev: 'danger',  icon: 'pi pi-arrow-up',    text: 'KELUAR' },
  'ADJUST+':{ sev: 'warn',    icon: 'pi pi-sliders-h',   text: 'OPNAME +' },
  'ADJUST-':{ sev: 'warn',    icon: 'pi pi-sliders-h',   text: 'OPNAME −' },
  INSERT:   { sev: 'success', text: 'INSERT' },
  UPDATE:   { sev: 'info',    text: 'UPDATE' },
  DELETE:   { sev: 'danger',  text: 'DELETE' },
  admin:    { sev: 'warn',    icon: 'pi pi-shield',      text: 'ADMIN' },
  petugas:  { sev: 'secondary', icon: 'pi pi-user',      text: 'PETUGAS' },
  Berfungsi:       { sev: 'success', text: 'BERFUNGSI' },
  'Rusak Ringan':  { sev: 'warn',    text: 'RUSAK RINGAN' },
  'Rusak Berat':   { sev: 'danger',  text: 'RUSAK BERAT' },
  'Perlu Kalibrasi': { sev: 'info',  text: 'KALIBRASI' },
}

const meta = () => MAP[props.kind] || { sev: 'secondary' as const }
</script>

<template>
  <Tag :severity="meta().sev" :value="label ?? meta().text ?? kind" :icon="meta().icon" />
</template>