<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api'
const locs = ref([]); const newName = ref('')
onMounted(async () => { const res = await api.get('/locations'); locs.value = res.data })
async function add() { if (!newName.value) return; await api.post('/locations', { name: newName.value }); newName.value = ''; const res = await api.get('/locations'); locs.value = res.data }
async function del(id: string) { if (!confirm('Hapus?')) return; await api.delete(`/locations/${id}`); const res = await api.get('/locations'); locs.value = res.data }
</script>
<template>
  <div class="space-y-4">
    <h2 class="text-xl font-bold">Lokasi ({{ locs.length }})</h2>
    <div class="card bg-base-100 border border-base-200 shadow-sm"><div class="card-body">
      <div class="flex gap-2"><input v-model="newName" placeholder="Nama ruangan" class="input input-bordered input-sm flex-1" @keyup.enter="add" /><button class="btn btn-primary btn-sm" @click="add">Tambah</button></div>
    </div></div>
    <div class="card bg-base-100 border border-base-200 shadow-sm"><div class="table"><table class="table table-sm"><thead><tr><th>Nama</th><th></th></tr></thead><tbody>
      <tr v-for="l in locs" :key="l.id"><td class="font-medium">{{ l.name }}</td><td><button class="btn btn-xs btn-ghost text-error" @click="del(l.id)">Hapus</button></td></tr>
    </tbody></table></div></div>
  </div>
</template>
