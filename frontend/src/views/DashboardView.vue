<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import api from '@/api'
import { useRouter } from 'vue-router'
import Button from 'primevue/button'
import PageHeader from '@/components/PageHeader.vue'
import StatCard from '@/components/StatCard.vue'
import Panel from '@/components/Panel.vue'
import StatusChip from '@/components/StatusChip.vue'
import EmptyState from '@/components/EmptyState.vue'
import { Bar } from 'vue-chartjs'
import { Chart as ChartJS, Tooltip, BarElement, CategoryScale, LinearScale } from 'chart.js'

ChartJS.register(Tooltip, BarElement, CategoryScale, LinearScale)

const router = useRouter()
const stats = ref<any>({ total_items: 0, total_stock: 0, low_stock: [], recent_tx: [], stock_by_cat: {} })
const loading = ref(true)

onMounted(async () => {
  try { stats.value = (await api.get('/dashboard')).data } finally { loading.value = false }
})

const catEntries = computed(() => Object.entries(stats.value.stock_by_cat || {}) as [string, number][])
const catMax = computed(() => Math.max(1, ...catEntries.value.map(e => e[1])))

const chartData = computed(() => ({
  labels: catEntries.value.map(e => e[0]),
  datasets: [{ data: catEntries.value.map(e => e[1]), backgroundColor: '#f5a524', hoverBackgroundColor: '#d98806', borderRadius: 3, barThickness: 26 }],
}))
const chartOptions = {
  indexAxis: 'y' as const,
  responsive: true,
  maintainAspectRatio: false,
  plugins: { legend: { display: false } },
  scales: {
    x: { grid: { color: '#ecf0f5' }, ticks: { font: { size: 10 }, color: '#8791a3' } },
    y: { grid: { display: false }, ticks: { font: { size: 10.5 }, color: '#3d4759' } },
  },
}
</script>

<template>
  <div>
    <PageHeader crumb="Ringkasan" title="Dashboard"
      sub="Kondisi stok dan pergerakan barang terkini">
      <template #actions>
        <Button label="Mutasi" icon="pi pi-plus" size="small" @click="router.push('/movement')" />
        <Button label="Barang" icon="pi pi-box" size="small" severity="secondary" outlined
                @click="router.push('/items')" />
      </template>
    </PageHeader>

    <div v-if="loading" class="grid place-items-center py-20">
      <i class="pi pi-spin pi-spinner text-xl" style="color: var(--txt-dim)" />
    </div>

    <template v-else>
      <!-- KPI row -->
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-3 mb-5">
        <StatCard label="Jenis Barang" :value="stats.total_items" icon="pi pi-box" hint="SKU terdaftar" />
        <StatCard label="Total Unit" :value="stats.total_stock" icon="pi pi-database" tone="accent" hint="Akumulasi seluruh stok" />
        <StatCard label="Stok Menipis" :value="stats.low_stock?.length || 0" icon="pi pi-exclamation-triangle"
                  :tone="(stats.low_stock?.length || 0) > 0 ? 'bad' : 'ok'" hint="Stok ≤ batas minimum" />
        <StatCard label="Transaksi Terakhir" :value="stats.recent_tx?.length || 0" icon="pi pi-history"
                  hint="8 pergerakan terbaru" />
      </div>

      <div class="grid lg:grid-cols-[1.35fr_1fr] gap-4 mb-4">
        <!-- chart -->
        <Panel title="Sebaran Stok per Kategori" icon="pi pi-chart-bar">
          <EmptyState v-if="!catEntries.length" icon="pi pi-chart-bar" title="Belum ada data stok"
                      sub="Tambahkan barang untuk melihat sebaran per kategori." />
          <div v-else style="height: 300px"><Bar :data="chartData" :options="chartOptions" /></div>
        </Panel>

        <!-- low stock -->
        <Panel title="Perlu Perhatian" icon="pi pi-exclamation-triangle" dense>
          <EmptyState v-if="!stats.low_stock?.length" icon="pi pi-check-circle" tone="ok"
                      title="Semua stok aman" sub="Tidak ada barang di bawah batas minimum." />
          <ul v-else class="divide-y" style="border-color: var(--line)">
            <li v-for="l in stats.low_stock" :key="l.sku"
                class="flex items-center gap-3 px-4 py-2.5 hover:bg-paper-2 transition-colors cursor-pointer"
                @click="router.push('/items')">
              <div class="flex-1 min-w-0">
                <div class="text-[12.5px] font-semibold truncate">{{ l.name }}</div>
                <div class="t-mono" style="color: var(--txt-dim)">{{ l.sku }} · {{ l.location }}</div>
              </div>
              <div class="text-right shrink-0">
                <div class="t-num text-[13px] font-bold text-sig-bad">{{ l.current }}</div>
                <div class="t-label !text-[9.5px]">min {{ l.min }}</div>
              </div>
            </li>
          </ul>
        </Panel>
      </div>

      <!-- recent tx -->
      <Panel title="Pergerakan Terbaru" icon="pi pi-history" dense>
        <template #actions>
          <Button label="Lihat semua" icon="pi pi-arrow-right" iconPos="right" text size="small"
                  @click="router.push('/history')" />
        </template>
        <EmptyState v-if="!stats.recent_tx?.length" icon="pi pi-history" title="Belum ada pergerakan"
                    sub="Catat barang masuk atau keluar untuk memulai jejak transaksi." />
        <div v-else class="overflow-x-auto">
          <table class="w-full text-[12.5px]">
            <thead>
              <tr class="text-left" style="background: var(--paper-2)">
                <th class="t-label px-4 py-2.5">Waktu</th>
                <th class="t-label px-4 py-2.5">Barang</th>
                <th class="t-label px-4 py-2.5">Jenis</th>
                <th class="t-label px-4 py-2.5 text-right">Jumlah</th>
              </tr>
            </thead>
            <tbody class="divide-y" style="border-color: var(--line-soft)">
              <tr v-for="(t, i) in stats.recent_tx" :key="i" class="hover:bg-paper-2 transition-colors">
                <td class="px-4 py-2.5 whitespace-nowrap" style="color: var(--txt-dim)">{{ t.time }}</td>
                <td class="px-4 py-2.5">
                  <span class="font-semibold">{{ t.item_name }}</span>
                  <span class="t-mono ml-1.5" style="color: var(--txt-dim)">{{ t.item_sku }}</span>
                </td>
                <td class="px-4 py-2.5"><StatusChip :kind="t.type" /></td>
                <td class="px-4 py-2.5 text-right t-num font-semibold">{{ t.quantity }} {{ t.unit }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </Panel>
    </template>
  </div>
</template>
