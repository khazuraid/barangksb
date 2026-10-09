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
    class="app-logo flex items-center select-none"
    :class="[
      variant === 'stacked' ? 'flex-col text-center gap-2' : 'gap-2.5',
      size === 'sm' ? 'scale-90' : size === 'lg' ? 'scale-110' : size === 'xl' ? 'scale-125' : ''
    ]"
  >
    <!-- SVG Icon Mark -->
    <div
      class="logo-mark relative shrink-0"
      :class="[
        size === 'sm' ? 'w-8 h-8' : size === 'lg' ? 'w-11 h-11' : size === 'xl' ? 'w-14 h-14' : 'w-9 h-9'
      ]"
    >
      <svg
        viewBox="0 0 90 90"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        class="w-full h-full drop-shadow-xs"
      >
        <!-- CLIPBOARD (Behind) -->
        <g class="clipboard-group">
          <!-- Clip top hanger ring -->
          <circle cx="63" cy="8" r="4.5" stroke="#0284c7" stroke-width="3" fill="none" />
          <!-- Clipboard Body -->
          <rect
            x="44"
            y="13"
            width="38"
            height="52"
            rx="4"
            fill="#ffffff"
            stroke="#0d9488"
            stroke-width="3.5"
            stroke-linejoin="round"
          />
          <!-- Blue Top Clip Clamp -->
          <rect x="53" y="10" width="20" height="9" rx="2.5" fill="#0284c7" />

          <!-- Blue Checkmark -->
          <path
            d="M52 23.5L57 28.5L67 18.5"
            stroke="#0284c7"
            stroke-width="3.2"
            stroke-linecap="round"
            stroke-linejoin="round"
          />

          <!-- Blue Checklist Lines -->
          <line x1="72" y1="23.5" x2="75" y2="23.5" stroke="#0284c7" stroke-width="3" stroke-linecap="round" />
          <line x1="52" y1="34" x2="75" y2="34" stroke="#0284c7" stroke-width="3" stroke-linecap="round" />
          <line x1="52" y1="42.5" x2="75" y2="42.5" stroke="#0284c7" stroke-width="3" stroke-linecap="round" />
          <line x1="52" y1="51" x2="75" y2="51" stroke="#0284c7" stroke-width="3" stroke-linecap="round" />
        </g>

        <!-- 3D MEDICAL INVENTORY BOX (Front-Left) -->
        <g class="box-group">
          <!-- Top Face -->
          <polygon
            points="12,38 36,27 64,38 40,49"
            fill="#ffffff"
            stroke="#0d9488"
            stroke-width="3.5"
            stroke-linejoin="round"
          />

          <!-- Tape on Top Face -->
          <polygon
            points="22,33 34,28 44,32.5 32,37.5"
            fill="#0d9488"
          />

          <!-- Right Face (Darker Teal Green) -->
          <polygon
            points="40,49 64,38 64,68 40,79"
            fill="#0d9488"
            stroke="#0d9488"
            stroke-width="3.5"
            stroke-linejoin="round"
          />

          <!-- Front Face (White with Teal Border) -->
          <polygon
            points="12,38 40,49 40,79 12,68"
            fill="#ffffff"
            stroke="#0d9488"
            stroke-width="3.5"
            stroke-linejoin="round"
          />

          <!-- Blue Medical Cross on Front Face -->
          <g transform="translate(19, 51)">
            <!-- Cross Vertical -->
            <path
              d="M5.5 0.5H9.5V14.5H5.5Z"
              fill="#0284c7"
            />
            <!-- Cross Horizontal -->
            <path
              d="M0.5 5H14.5V9.5H0.5Z"
              fill="#0284c7"
            />
          </g>
        </g>
      </svg>
    </div>

    <!-- Typography Text -->
    <div
      v-if="!hideText && variant !== 'icon'"
      class="logo-text flex flex-col leading-none text-left"
      :class="variant === 'stacked' ? 'items-center text-center' : ''"
    >
      <span
        class="font-black tracking-tight"
        :class="[
          size === 'sm' ? 'text-[12px]' : size === 'lg' ? 'text-[16px]' : size === 'xl' ? 'text-[22px]' : 'text-[13.5px]'
        ]"
        style="color: var(--txt); letter-spacing: -0.02em;"
      >
        INVENTARIS
      </span>
      <span
        class="font-black text-emerald-600 dark:text-emerald-400"
        :class="[
          size === 'sm' ? 'text-[11.5px]' : size === 'lg' ? 'text-[15.5px]' : size === 'xl' ? 'text-[21px]' : 'text-[13px]'
        ]"
        style="color: #0d9488; letter-spacing: 0.02em;"
      >
        PUSKESMAS
      </span>
    </div>
  </div>
</template>

<style scoped>
.app-logo {
  transition: transform 0.15s ease;
}
</style>
