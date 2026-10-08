<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api'
import { Bar } from 'vue-chartjs'
import { Chart as ChartJS, Title, Tooltip, Legend, BarElement, CategoryScale, LinearScale } from 'chart.js'

ChartJS.register(Title, Tooltip, Legend, BarElement, CategoryScale, LinearScale)

const stats = ref({ total_items: 0, total_stock: 0, low_stock: [], recent_tx: [], stock_by_cat: {} })
const loading = ref(true)

onMounted(async () => {
  try {
    const res = await api.get('/dashboard')
    stats.value = res.data
  } finally {
    loading.value = false
  }
})

const chartData = ref({ labels: [] as string[], datasets: [{ data: [] as number[], backgroundColor: '#10B981' }] })
const chartOptions = { responsive: true, plugins: { legend: { display: false } } }

function updateChart() {
  const entries = Object.entries(stats.value.stock_by_cat || {})
  chartData.value = {
    labels: entries.map(e => e[0]),
    datasets: [{ label: 'Stok', data: entries.map(e => e[1]), backgroundColor: '#10B981' }],
  }
}
</script>

<template>
  <div v-if="loading" class="flex justify-center py-12"><span class="loading loading-spinner loading-lg"></span></div>
  <div v-else>
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
      <div class="stat bg-base-100 rounded-xl border border-base-200 shadow-sm">
        <div class="stat-title">Jenis Barang</div>
        <div class="stat-value text-primary">{{ stats.total_items }}</div>
      </div>
      <div class="stat bg-base-100 rounded-xl border border-base-200 shadow-sm">
        <div class="stat-title">Total Unit</div>
        <div class="stat-value text-secondary">{{ stats.total_stock }}</div>
      </div>
      <div class="stat bg-base-100 rounded-xl border border-base-200 shadow-sm">
        <div class="stat-title">Stok Menipis</div>
        <div class="stat-value text-error">{{ stats.low_stock?.length || 0 }}</div>
      </div>
      <div class="stat bg-base-100 rounded-xl border border-base-200 shadow-sm">
        <div class="stat-title">Kategori</div>
        <div class="stat-value text-accent">{{ Object.keys(stats.stock_by_cat || {}).length }}</div>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <div class="card bg-base-100 shadow-sm border border-base-200">
        <div class="card-body">
          <h2 class="card-title">Stok per Kategori</h2>
          <Bar :data="chartData" :options="chartOptions" @click="updateChart" />
        </div>
      </div>
      <div class="card bg-base-100 shadow-sm border border-base-200">
        <div class="card-body">
          <h2 class="card-title">Perlu Restock</h2>
          <div v-if="!stats.low_stock?.length" class="text-base-content/50 py-8 text-center">Semua stok aman ✓</div>
          <ul v-else class="space-y-2">
            <li v-for="item in stats.low_stock" :key="item.sku" class="flex justify-between items-center pb-2 border-b border-base-100">
              <div><div class="font-medium text-sm">{{ item.name }}</div><code class="text-xs">{{ item.sku }}</code></div>
              <span class="badge badge-error badge-sm">{{ item.current }}/{{ item.min }} {{ item.unit }}</span>
            </li>
          </ul>
        </div>
      </div>
    </div>

    <div class="card bg-base-100 shadow-sm border border-base-200 mt-4">
      <div class="card-body">
        <h2 class="card-title">Mutasi Terbaru</h2>
        <div class="overflow-x-auto">
          <table class="table table-sm">
            <thead><tr><th>Barang</th><th>Jenis</th><th>Qty</th></tr></thead>
            <tbody>
              <tr v-for="tx in stats.recent_tx" :key="tx.item_sku">
                <td>{{ tx.item_name }} <code class="text-xs">{{ tx.item_sku }}</code></td>
                <td><span class="badge badge-sm" :class="tx.type === 'IN' ? 'badge-success' : 'badge-error'">{{ tx.type }}</span></td>
                <td>{{ tx.quantity }} {{ tx.unit }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>
