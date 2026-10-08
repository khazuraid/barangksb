<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api'
const logs = ref([])
onMounted(async () => { logs.value = (await api.get('/audit')).data })
</script>
<template>
  <div class="card bg-base-100 border border-base-200 shadow-sm"><div class="overflow-x-auto"><table class="table table-sm">
    <thead><tr><th>Waktu</th><th>Tabel</th><th>Aksi</th><th>Row ID</th></tr></thead>
    <tbody><tr v-for="l in logs" :key="l.id"><td class="text-base-content/50 text-xs">{{ l.at }}</td><td>{{ l.table_name }}</td><td><span class="badge badge-sm" :class="l.op === 'INSERT' ? 'badge-success' : l.op === 'DELETE' ? 'badge-error' : 'badge-warning'">{{ l.op }}</span></td><td><code class="text-xs">{{ l.row_id }}</code></td></tr></tbody>
  </table></div></div>
</template>
