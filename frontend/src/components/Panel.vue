<script setup lang="ts">
defineProps<{ title?: string; dense?: boolean; icon?: string }>()
</script>

<template>
  <section class="panel overflow-hidden">
    <div
      v-if="title || $slots.actions"
      class="flex items-center justify-between gap-3 px-4 border-b"
      style="border-color: var(--line); height: 46px; background: var(--panel-2)"
    >
      <div class="flex items-center gap-2.5 min-w-0">
        <i v-if="icon" :class="icon" class="text-[13px] text-acc-500" />
        <h2 class="text-[13px] font-semibold tracking-normal" style="color: var(--txt)">{{ title }}</h2>
      </div>
      <div class="flex items-center gap-2 shrink-0"><slot name="actions" /></div>
    </div>
    <div :class="dense ? 'p-0' : 'p-4'"><slot /></div>
    <div v-if="$slots.footer" class="px-4 py-3 border-t" style="border-color: var(--line); background: var(--panel-2)">
      <slot name="footer" />
    </div>
  </section>
</template>
