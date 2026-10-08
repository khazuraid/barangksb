<script setup lang="ts">
// Small labelled form field wrapper — keeps label/spacing consistent everywhere.
defineProps<{ label: string; hint?: string; required?: boolean; span?: boolean }>()
</script>

<template>
  <label class="flex flex-col gap-1.5 min-w-0" :class="span ? 'sm:col-span-2' : ''">
    <span class="text-[11.5px] font-semibold" style="color: var(--txt-dim)">
      {{ label }}<span v-if="required" class="text-sig-bad"> *</span>
    </span>
    <slot />
    <span v-if="hint" class="text-[11px]" style="color: var(--txt-dim)">{{ hint }}</span>
  </label>
</template>
