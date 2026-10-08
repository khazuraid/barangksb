<script setup lang="ts">
defineProps<{ title: string; sub?: string; crumb?: string }>()
</script>

<template>
  <header class="flex flex-wrap items-end justify-between gap-3 pb-3.5 mb-4 border-b" style="border-color: var(--line)">
    <div class="min-w-0">
      <div v-if="crumb" class="t-label mb-1">{{ crumb }}</div>
      <h1 class="t-page truncate">{{ title }}</h1>
      <p v-if="sub" class="text-[12.5px] mt-1" style="color: var(--txt-dim)">{{ sub }}</p>
    </div>
    <div class="flex items-center gap-2 shrink-0">
      <slot name="actions" />
    </div>
  </header>
</template>
