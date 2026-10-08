<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import api from '@/api'
import Card from 'primevue/card'
import Button from 'primevue/button'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Tag from 'primevue/tag'
import { Bar } from 'vue-chartjs'
import { Chart as ChartJS, Title, Tooltip, Legend, BarElement, CategoryScale, LinearScale } from 'chart.js'

ChartJS.register(Title, Tooltip, Legend, BarElement, CategoryScale, LinearScale)

const stats = ref({ total_items: 0, total_stock: 0, low_stock: [] as any[], recent_tx: [] as any[], stock_by_cat: {} as Record<string, number> })
const loading = ref(true)

onMounted(async () => {
  try {
    const res = await api.get('/dashboard')
    stats.value = res.data
  } finally {
    loading.value = false
  }
})

const chartData = computed(() => {
  const entries = Object.entries(stats.value.stock_by_cat || {})
  return {
    labels: entries.map(e => e[0]),
    datasets: [{ label: 'Stok', data: entries.map(e => e[1]), backgroundColor: '#10B981', borderRadius: 8 }],
  }
})
const chartOptions = { responsive: true, plugins: { legend: { display: false } } }
</script>

<template>
  <div v-if="loading" class="flex justify-center py-12">
    <i class="pi pi-spin pi-spinner text-4xl text-emerald-500"></i>
  </div>
  <div v-else class="flex flex-col gap-4">
    <!-- Stat cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <Card class="shadow-sm">
        <template #content>
          <div class="flex items-center gap-3">
            <div class="w-12 h-12 rounded-xl bg-emerald-100 dark:bg-emerald-900/30 flex items-center justify-center">
              <i class="pi pi-box text-2xl text-emerald-600 dark:text-emerald-400"></i>
            </div>
            <div>
              <div class="text-2xl font-bold text-gray-800 dark:text-gray-100">{{ stats.total_items }}</div>
              <div class="text-xs text-gray-500 uppercase font-semibold tracking-wide">Jenis Barang</div>
            </div>
          </div>
        </template>
      </Card>
      <Card class="shadow-sm">
        <template #content>
          <div class="flex items-center gap-3">
            <div class="w-12 h-12 rounded-xl bg-violet-100 dark:bg-violet-900/30 flex items-center justify-center">
              <i class="pi pi-folder-open text-2xl text-violet-600 dark:text-violet-400"></i>
            </div>
            <div>
              <div class="text-2xl font-bold text-gray-800 dark:text-gray-100">{{ stats.total_stock }}</div>
              <div class="text-xs text-gray-500 uppercase font-semibold tracking-wide">Total Unit</div>
            </div>
          </div>
        </template>
      </Card>
      <Card class="shadow-sm">
        <template #content>
          <div class="flex items-center gap-3">
            <div class="w-12 h-12 rounded-xl bg-red-100 dark:bg-red-900/30 flex items-center justify-center">
              <i class="pi pi-exclamation-triangle text-2xl text-red-600 dark:text-red-400"></i>
            </div>
            <div>
              <div class="text-2xl font-bold text-gray-800 dark:text-gray-100">{{ stats.low_stock?.length || 0 }}</div>
              <div class="text-xs text-gray-500 uppercase font-semibold tracking-wide">Stok Menipis</div>
            </div>
          </div>
        </template>
      </Card>
      <Card class="shadow-sm">
        <template #content>
          <div class="flex items-center gap-3">
            <div class="w-12 h-12 rounded-xl bg-amber-100 dark:bg-amber-900/30 flex items-center justify-center">
              <i class="pi pi-tags text-2xl text-amber-600 dark:text-amber-400"></i>
            </div>
            <div>
              <div class="text-2xl font-bold text-gray-800 dark:text-gray-100">{{ Object.keys(stats.stock_by_cat || {}).length }}</div>
              <div class="text-xs text-gray-500 uppercase font-semibold tracking-wide">Kategori</div>
            </div>
          </div>
        </template>
      </Card>
    </div>

    <!-- Chart + Low stock -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <Card class="shadow-sm">
        <template #title><i class="pi pi-chart-bar text-emerald-500 mr-2"></i>Stok per Kategori</template>
        <template #content>
          <Bar :data="chartData" :options="chartOptions" />
        </template>
      </Card>
      <Card class="shadow-sm">
        <template #title><i class="pi pi-exclamation-triangle text-red-500 mr-2"></i>Perlu Restock</template>
        <template #content>
          <div v-if="!stats.low_stock?.length" class="text-center py-8 text-gray-400">
            <i class="pi pi-check-circle text-4xl text-emerald-400 mb-2"></i>
            <div>Semua stok aman</div>
          </div>
          <div v-else class="flex flex-col gap-2">
            <div v-for="item in stats.low_stock" :key="item.sku" class="flex justify-between items-center py-2 border-b border-gray-100 dark:border-gray-800">
              <div>
                <div class="font-medium text-sm">{{ item.name }}</div>
                <code class="text-xs text-gray-400">{{ item.sku }}</code>
              </div>
              <Tag severity="danger" :value="`${item.current}/${item.min} ${item.unit}`" />
            </div>
          </div>
        </template>
      </Card>
    </div>

    <!-- Recent transactions -->
    <Card class="shadow-sm">
      <template #title><i class="pi pi-clock text-emerald-500 mr-2"></i>Mutasi Terbaru</template>
      <template #content>
        <DataTable :value="stats.recent_tx" :rows="8" responsiveLayout="scroll" class="p-datatable-sm">
          <Column field="item_name" header="Barang">
            <template #body="{ data }">
              <strong>{{ data.item_name }}</strong> <code class="text-xs text-gray-400">{{ data.item_sku }}</code>
            </template>
          </Column>
          <Column field="type" header="Jenis" style="width: 100px">
            <template #body="{ data }">
              <Tag :severity="data.type === 'IN' ? 'success' : data.type === 'OUT' ? 'danger' : 'warn'" :value="data.type" />
            </template>
          </Column>
          <Column field="quantity" header="Qty" style="width: 100px">
            <template #body="{ data }">{{ data.quantity }} {{ data.unit }}</template>
          </Column>
        </DataTable>
      </template>
    </Card>
  </div>
</template>
