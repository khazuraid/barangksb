<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import Card from 'primevue/card'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import Button from 'primevue/button'
import { useToast } from 'primevue/usetoast'

const auth = useAuthStore()
const router = useRouter()
const toast = useToast()
const email = ref('')
const password = ref('')
const loading = ref(false)

async function login() {
  loading.value = true
  try {
    await auth.login(email.value, password.value)
    toast.add({ severity: 'success', summary: 'Berhasil', detail: 'Login berhasil', life: 2000 })
    router.push('/')
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Gagal', detail: e.response?.data?.error || 'Login gagal', life: 3000 })
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-emerald-50 via-teal-50 to-cyan-50 dark:from-gray-950 dark:via-gray-900 dark:to-emerald-950 p-4">
    <div class="w-full max-w-sm">
      <!-- Logo -->
      <div class="text-center mb-8">
        <div class="w-16 h-16 rounded-2xl bg-gradient-to-br from-emerald-500 to-teal-600 mx-auto flex items-center justify-center text-white font-bold text-2xl shadow-xl shadow-emerald-500/40 mb-4">
          IK
        </div>
        <h1 class="text-2xl font-bold text-gray-800 dark:text-gray-100">Inventaris Kantor</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">Sistem Barcode & Stok Real-Time</p>
      </div>

      <Card class="shadow-xl border-0">
        <template #content>
          <form @submit.prevent="login" class="flex flex-col gap-4">
            <div class="flex flex-col gap-2">
              <label class="text-sm font-medium text-gray-700 dark:text-gray-300">Email</label>
              <InputText v-model="email" type="email" required placeholder="email@kantor.id" class="w-full" />
            </div>
            <div class="flex flex-col gap-2">
              <label class="text-sm font-medium text-gray-700 dark:text-gray-300">Password</label>
              <Password v-model="password" required :feedback="false" toggleMask placeholder="••••••••" class="w-full" inputClass="w-full" />
            </div>
            <Button type="submit" label="Masuk" icon="pi pi-sign-in" :loading="loading" class="w-full" />
          </form>
        </template>
      </Card>

      <p class="text-center text-xs text-gray-400 mt-6">
        <i class="pi pi-shield-lock"></i> Akses terbatas untuk petugas & admin kantor
      </p>
    </div>
  </div>
</template>
