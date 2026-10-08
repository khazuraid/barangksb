<script setup lang="ts">
import { onMounted } from 'vue'
import { RouterView } from 'vue-router'
import Toast from 'primevue/toast'
import ConfirmDialog from 'primevue/confirmdialog'

// The app is dark-only: PrimeVue's darkModeSelector is body.dark, so the class
// must exist before PrimeVue styles its overlays (dialogs, toasts, menus).
onMounted(() => document.body.classList.add('dark'))
</script>

<template>
  <RouterView />
  <Toast />
  <ConfirmDialog />
</template>