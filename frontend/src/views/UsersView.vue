<script setup lang="ts">
import { ref } from 'vue'
import api from '@/api'
const users = ref([]); const form = ref({ name: '', email: '', password: '', role: 'petugas' })
async function fetch() { users.value = (await api.get('/users')).data }
async function create() { await api.post('/users', form.value); form.value = { name: '', email: '', password: '', role: 'petugas' }; fetch() }
fetch()
</script>
<template>
  <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
    <div class="card bg-base-100 border border-base-200 shadow-sm"><div class="card-body">
      <h3 class="font-bold">Tambah Pengguna</h3>
      <form @submit.prevent="create" class="space-y-2">
        <input v-model="form.name" placeholder="Nama" class="input input-bordered input-sm w-full" />
        <input v-model="form.email" placeholder="Email" class="input input-bordered input-sm w-full" />
        <input v-model="form.password" type="password" placeholder="Password (min 8)" class="input input-bordered input-sm w-full" />
        <select v-model="form.role" class="select select-bordered select-sm w-full"><option value="petugas">Petugas</option><option value="admin">Admin</option></select>
        <button class="btn btn-primary btn-sm w-full">Simpan</button>
      </form>
    </div></div>
    <div class="card bg-base-100 border border-base-200 shadow-sm"><div class="overflow-x-auto"><table class="table table-sm">
      <thead><tr><th>Nama</th><th>Email</th><th>Role</th></tr></thead>
      <tbody><tr v-for="u in users" :key="u.id"><td>{{ u.name }}</td><td class="text-base-content/50">{{ u.email }}</td><td><span class="badge badge-sm" :class="u.role === 'admin' ? 'badge-secondary' : 'badge-ghost'">{{ u.role }}</span></td></tr></tbody>
    </table></div></div>
  </div>
</template>
