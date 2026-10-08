<script setup lang="ts">
defineProps<{ icon?: string; title: string; sub?: string }>()
</script>

<template>
  <div class="grid-bg panel grid place-items-center px-6 py-14 text-center">
    <div class="w-11 h-11 grid place-items-center rounded-lg border mb-3"
         style="border-color: var(--line); background: var(--panel)">
      <i :class="icon || 'pi pi-inbox'" class="text-base" style="color: var(--txt-dim)" />
    </div>
    <div class="text-[13.5px] font-semibold">{{ title }}</div>
    <div v-if="sub" class="text-[12px] mt-1 max-w-sm" style="color: var(--txt-dim)">{{ sub }}</div>
    <div class="mt-3.5"><slot /></div>
  </div>
</template>
