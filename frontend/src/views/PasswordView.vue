<script setup lang="ts">
import { ref } from 'vue'
import api from '@/api'
const oldPw = ref(''); const newPw = ref(''); const confirm = ref(''); const msg = ref('')
async function submit() {
  if (newPw.value !== confirm.value) { msg.value = 'Konfirmasi tidak cocok'; return }
  try { await api.post('/auth/password', { old: oldPw.value, new: newPw.value }); msg.value = 'Password berhasil diganti' } catch (e: any) { msg.value = e.response?.data?.error || 'Gagal' }
}
</script>
<template>
  <div class="card bg-base-100 border border-base-200 shadow-sm max-w-md"><div class="card-body">
    <h3 class="font-bold">Ganti Password</h3>
    <div v-if="msg" class="alert alert-info text-sm">{{ msg }}</div>
    <form @submit.prevent="submit" class="space-y-2">
      <input v-model="oldPw" type="password" placeholder="Password lama" class="input input-bordered input-sm w-full" />
      <input v-model="newPw" type="password" placeholder="Password baru (min 8)" class="input input-bordered input-sm w-full" />
      <input v-model="confirm" type="password" placeholder="Ulangi" class="input input-bordered input-sm w-full" />
      <button class="btn btn-primary btn-sm w-full">Simpan</button>
    </form>
  </div></div>
</template>
