<script setup lang="ts">
import { onMounted } from 'vue'
import { RouterView } from 'vue-router'
import Toast from 'primevue/toast'
import ConfirmDialog from 'primevue/confirmdialog'
import { useThemeStore } from '@/stores/theme'

const theme = useThemeStore()
onMounted(() => {
  theme.apply()
  if ((window as any).Telegram?.WebApp) {
    const tg = (window as any).Telegram.WebApp
    try {
      tg.ready()
      tg.expand()
    } catch {
      // ignore
    }
  }
})
</script>

<template>
  <RouterView />
  <Toast />
  <ConfirmDialog />
</template>