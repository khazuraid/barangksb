<script setup lang="ts">
defineProps<{ title?: string; dense?: boolean; icon?: string }>()
</script>

<template>
  <section class="panel overflow-hidden">
    <div
      v-if="title || $slots.actions"
      class="flex items-center justify-between gap-3 px-4 border-b"
      style="border-color: var(--line); height: 42px"
    >
      <div class="flex items-center gap-2 min-w-0">
        <i v-if="icon" :class="icon" class="text-[12px] text-acc-500" />
        <h2 class="text-[12.5px] font-bold tracking-wide uppercase" style="color: var(--txt-dim)">{{ title }}</h2>
      </div>
      <div class="flex items-center gap-2 shrink-0"><slot name="actions" /></div>
    </div>
    <div :class="dense ? 'p-0' : 'p-4'"><slot /></div>
    <div v-if="$slots.footer" class="px-4 py-3 border-t" style="border-color: var(--line)"><slot name="footer" /></div>
  </section>
</template>
