<script setup lang="ts">
withDefaults(
  defineProps<{
    variant?: 'full' | 'icon' | 'stacked'
    size?: 'sm' | 'md' | 'lg' | 'xl'
    hideText?: boolean
  }>(),
  {
    variant: 'full',
    size: 'md',
    hideText: false,
  }
)
</script>

<template>
  <div
    class="inline-flex items-center select-none"
    :class="variant === 'stacked' ? 'flex-col text-center gap-2.5' : 'gap-2.5'"
  >
    <div
      class="shrink-0 grid place-items-center rounded-[9px] bg-acc-500 text-white dark:text-[#042f2e] border border-white/10"
      :class="[
        size === 'sm' ? 'w-7 h-7' : size === 'lg' ? 'w-10 h-10 rounded-[11px]' : size === 'xl' ? 'w-12 h-12 rounded-[13px]' : 'w-8 h-8'
      ]"
    >
      <svg
        viewBox="0 0 32 32"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        class="w-[68%] h-[68%]"
        aria-hidden="true"
      >
        <path
          d="M16 4.5L26 10V22L16 27.5L6 22V10L16 4.5Z"
          stroke="currentColor"
          stroke-width="2.3"
          stroke-linejoin="round"
        />
        <path
          d="M6.5 10.3L16 15.8L25.5 10.3M16 15.8V27"
          stroke="currentColor"
          stroke-width="2.2"
          stroke-linejoin="round"
        />
        <path
          d="M16 7.8V12.2M13.8 10H18.2"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
        />
      </svg>
    </div>

    <div
      v-if="!hideText && variant !== 'icon'"
      class="flex flex-col leading-tight min-w-0"
      :class="variant === 'stacked' ? 'items-center text-center' : 'text-left'"
    >
      <span
        class="font-bold tracking-[-0.025em] text-txt truncate"
        :class="[
          size === 'sm' ? 'text-[12.5px]' : size === 'lg' ? 'text-[16px]' : size === 'xl' ? 'text-[20px]' : 'text-[13.5px]'
        ]"
      >
        Inventaris
      </span>
      <span
        class="t-mono font-semibold uppercase tracking-[0.08em] text-acc-500 leading-none mt-0.5"
        :class="[
          size === 'sm' ? '!text-[9px]' : size === 'lg' ? '!text-[10.5px]' : size === 'xl' ? '!text-[12px]' : '!text-[9.5px]'
        ]"
      >
        Puskesmas
      </span>
    </div>
  </div>
</template>
