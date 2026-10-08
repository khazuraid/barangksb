<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()
const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function login() {
  loading.value = true
  error.value = ''
  try {
    await auth.login(email.value, password.value)
    router.push('/')
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Login gagal'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-primary/10 to-secondary/10">
    <div class="card w-full max-w-sm bg-base-100 shadow-xl">
      <div class="card-body">
        <div class="text-center mb-6">
          <div class="w-16 h-16 rounded-2xl bg-primary text-primary-content flex items-center justify-center font-bold text-2xl mx-auto mb-3">IK</div>
          <h2 class="text-xl font-bold">Inventaris Kantor</h2>
          <p class="text-sm text-base-content/50">Sistem Barcode & Stok Real-Time</p>
        </div>
        <div v-if="error" class="alert alert-error text-sm">{{ error }}</div>
        <form @submit.prevent="login" class="space-y-3">
          <label class="form-control w-full">
            <div class="label"><span class="label-text">Email</span></div>
            <input v-model="email" type="email" required placeholder="email@kantor.id" class="input input-bordered w-full" />
          </label>
          <label class="form-control w-full">
            <div class="label"><span class="label-text">Password</span></div>
            <input v-model="password" type="password" required placeholder="••••••••" class="input input-bordered w-full" />
          </label>
          <button type="submit" class="btn btn-primary w-full" :disabled="loading">
            {{ loading ? 'Loading...' : 'Masuk' }}
          </button>
        </form>
      </div>
    </div>
  </div>
</template>
