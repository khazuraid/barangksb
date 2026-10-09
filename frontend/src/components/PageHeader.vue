<script setup lang="ts">
defineProps<{ title: string; sub?: string; crumb?: string }>()
</script>

<template>
  <header class="flex flex-wrap items-center justify-between gap-4 pb-4 mb-5 border-b" style="border-color: var(--line)">
    <div class="min-w-0">
      <div v-if="crumb" class="text-[11.5px] font-medium tracking-wide text-acc-500 mb-0.5 flex items-center gap-1.5">
        <i class="pi pi-compass text-[10px]" />
        <span>{{ crumb }}</span>
      </div>
      <h1 class="text-[22px] font-bold tracking-tight" style="color: var(--txt)">{{ title }}</h1>
      <p v-if="sub" class="text-[13px] mt-0.5 font-normal" style="color: var(--txt-dim)">{{ sub }}</p>
    </div>
    <div class="flex items-center gap-2.5 shrink-0 flex-wrap">
      <slot name="actions" />
    </div>
  </header>
</template>
