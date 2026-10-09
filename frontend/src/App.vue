<script setup lang="ts">
import { onMounted } from 'vue'
import { RouterView } from 'vue-router'
import Toast from 'primevue/toast'
import ConfirmDialog from 'primevue/confirmdialog'
import { useThemeStore } from '@/stores/theme'

const theme = useThemeStore()
onMounted(() => theme.apply())
</script>

<template>
  <RouterView />
  <Toast />
  <ConfirmDialog />
</template>