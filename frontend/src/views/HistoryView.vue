<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import api from '@/api'
const txs = ref([]); const total = ref(0); const page = ref(1); const perPage = ref(25)
const sku = ref(''); const type = ref('')
async function fetch() {
  const res = await api.get('/transactions', { params: { sku: sku.value, type: type.value, page: page.value, per_page: perPage.value } })
  txs.value = res.data.data; total.value = res.data.total
}
const pages = () => Math.ceil(total.value / perPage.value)
onMounted(fetch); watch([page, perPage], fetch)
</script>
<template>
  <div class="space-y-4">
    <h2 class="text-xl font-bold">Riwayat Mutasi ({{ total }})</h2>
    <div class="flex gap-2 flex-wrap">
      <input v-model="sku" @keyup.enter="fetch" placeholder="SKU..." class="input input-bordered input-sm" />
      <select v-model="type" @change="fetch" class="select select-bordered select-sm"><option value="">Semua</option><option>IN</option><option>OUT</option><option>ADJUST+</option><option>ADJUST-</option></select>
      <button class="btn btn-sm btn-primary" @click="fetch">Filter</button>
    </div>
    <div class="card bg-base-100 shadow-sm border border-base-200">
      <div class="overflow-x-auto">
        <table class="table table-sm">
          <thead><tr><th>Waktu</th><th>Jenis</th><th>Barang</th><th>Qty</th><th>Person</th></tr></thead>
          <tbody>
            <tr v-for="t in txs" :key="t.id">
              <td class="text-base-content/50 text-xs">{{ t.timestamp }}</td>
              <td><span class="badge badge-sm" :class="t.type === 'IN' ? 'badge-success' : t.type === 'OUT' ? 'badge-error' : 'badge-warning'">{{ t.type }}</span></td>
              <td><code class="text-xs">{{ t.item_sku }}</code> {{ t.item_name }}</td>
              <td>{{ t.quantity }} {{ t.unit }}</td>
              <td class="text-base-content/50">{{ t.received_by }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <div class="p-4 flex justify-between items-center">
        <div class="join"><button v-for="p in pages()" :key="p" class="join-item btn btn-sm" :class="{ 'btn-active': page === p }" @click="page = p">{{ p }}</button></div>
        <select v-model="perPage" class="select select-bordered select-sm"><option :value="20">20</option><option :value="25">25</option><option :value="50">50</option><option :value="100">100</option></select>
      </div>
    </div>
  </div>
</template>
