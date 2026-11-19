<script setup lang="ts">
defineProps<{ label: string; hint?: string; required?: boolean; span?: boolean }>()
</script>

<template>
  <label class="flex flex-col gap-1.5 min-w-0" :class="span ? 'sm:col-span-2' : ''">
    <span class="text-[12px] font-medium text-txt-dim">
      {{ label }}<span v-if="required" class="text-sig-bad ml-0.5">*</span>
    </span>
    <slot />
    <span v-if="hint" class="text-[11px] text-txt-dim leading-snug">{{ hint }}</span>
  </label>
</template>
