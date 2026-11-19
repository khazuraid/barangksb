<script setup lang="ts">
defineProps<{ title: string; sub?: string; crumb?: string }>()
</script>

<template>
  <header class="flex flex-wrap items-end justify-between gap-3 pb-4 mb-5 border-b border-line">
    <div class="min-w-0">
      <div
        v-if="crumb"
        class="t-mono !text-[10.5px] font-medium uppercase tracking-[0.08em] text-acc-500 mb-1 flex items-center gap-1.5"
      >
        <span class="w-1.5 h-1.5 rounded-full bg-acc-500 shrink-0" aria-hidden="true" />
        <span>{{ crumb }}</span>
      </div>
      <h1 class="text-[20px] font-bold tracking-[-0.025em] leading-tight text-txt">{{ title }}</h1>
      <p v-if="sub" class="text-[13px] mt-1 text-txt-dim max-w-[65ch]">{{ sub }}</p>
    </div>
    <div class="flex items-center gap-2 shrink-0 flex-wrap">
      <slot name="actions" />
    </div>
  </header>
</template>
