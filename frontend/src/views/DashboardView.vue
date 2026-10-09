<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import api from '@/api'
import { useRouter } from 'vue-router'
import { useThemeStore } from '@/stores/theme'
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
const theme = useThemeStore()
const stats = ref<any>({ total_items: 0, total_stock: 0, low_stock: [], recent_tx: [], stock_by_cat: {} })
const loading = ref(true)

onMounted(async () => {
  try { stats.value = (await api.get('/dashboard')).data } finally { loading.value = false }
})

const catEntries = computed(() => Object.entries(stats.value.stock_by_cat || {}) as [string, number][])

const chartData = computed(() => ({
  labels: catEntries.value.map(e => e[0]),
  datasets: [{
    data: catEntries.value.map(e => e[1]),
    backgroundColor: theme.isDark ? '#6366f1' : '#4f46e5',
    hoverBackgroundColor: theme.isDark ? '#818cf8' : '#4338ca',
    borderRadius: 6,
    barThickness: 24
  }],
}))

const chartOptions = computed(() => ({
  indexAxis: 'y' as const,
  responsive: true,
  maintainAspectRatio: false,
  plugins: { legend: { display: false } },
  scales: {
    x: {
      grid: { color: theme.isDark ? '#1e293b' : '#f1f5f9' },
      ticks: { font: { size: 11 }, color: '#94a3b8' }
    },
    y: {
      grid: { display: false },
      ticks: { font: { size: 11.5 }, color: theme.isDark ? '#cbd5e1' : '#475569' }
    },
  },
}))
</script>

<template>
  <div>
    <PageHeader crumb="Ringkasan Operasional" title="Dashboard Inventaris"
      sub="Pantau stok real-time, sebaran kategori, dan pergerakan mutasi">
      <template #actions>
        <Button label="Mutasi Masuk" icon="pi pi-arrow-down-left" size="small" severity="success"
                @click="router.push('/movement')" />
        <Button label="Mutasi Keluar" icon="pi pi-arrow-up-right" size="small" severity="warn"
                @click="router.push('/movement')" />
        <Button label="Tambah Barang" icon="pi pi-plus" size="small"
                @click="router.push('/items/new')" />
      </template>
    </PageHeader>

    <div v-if="loading" class="grid place-items-center py-24">
      <i class="pi pi-spin pi-spinner text-2xl" style="color: var(--txt-dim)" />
    </div>

    <template v-else>
      <!-- KPI cards row -->
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <StatCard label="Total Barang" :value="stats.total_items" icon="pi pi-box" tone="info" hint="SKU terdaftar aktif" />
        <StatCard label="Total Unit Fisik" :value="stats.total_stock" icon="pi pi-database" tone="accent" hint="Akumulasi seluruh stok" />
        <StatCard label="Stok Kritis" :value="stats.low_stock?.length || 0" icon="pi pi-exclamation-triangle"
                  :tone="(stats.low_stock?.length || 0) > 0 ? 'bad' : 'ok'" hint="Stok ≤ batas minimum" />
        <StatCard label="Mutasi Terbaru" :value="stats.recent_tx?.length || 0" icon="pi pi-history" tone="neutral"
                  hint="Transaksi tercatat" />
      </div>

      <div class="grid lg:grid-cols-[1.4fr_1fr] gap-5 mb-5 items-start">
        <!-- chart -->
        <Panel title="Sebaran Stok per Kategori" icon="pi pi-chart-bar">
          <EmptyState v-if="!catEntries.length" icon="pi pi-chart-bar" title="Belum ada data stok"
                      sub="Tambahkan barang untuk melihat sebaran per kategori." />
          <div v-else style="height: 320px"><Bar :data="chartData" :options="chartOptions" /></div>
        </Panel>

        <!-- low stock -->
        <Panel title="Perlu Perhatian (Stok Menipis)" icon="pi pi-exclamation-circle" dense>
          <EmptyState v-if="!stats.low_stock?.length" icon="pi pi-check-circle" tone="ok"
                      title="Semua stok aman" sub="Tidak ada barang di bawah batas minimum." />
          <div v-else class="divide-y max-h-[320px] overflow-y-auto" style="border-color: var(--line)">
            <div
              v-for="l in stats.low_stock"
              :key="l.sku"
              class="flex items-center gap-3 px-4 py-3 hover:bg-slate-50 dark:hover:bg-slate-800/40 transition-colors cursor-pointer"
              @click="router.push('/items')"
            >
              <div class="w-8 h-8 rounded-lg grid place-items-center bg-rose-50 dark:bg-rose-950/40 border border-rose-200 dark:border-rose-800 shrink-0">
                <i class="pi pi-exclamation-triangle text-rose-500 text-[12px]" />
              </div>
              <div class="flex-1 min-w-0">
                <div class="text-[13px] font-semibold truncate">{{ l.name }}</div>
                <div class="t-mono text-[11px]" style="color: var(--txt-dim)">{{ l.sku }} · {{ l.location }}</div>
              </div>
              <div class="text-right shrink-0">
                <div class="t-num text-[13px] font-bold text-rose-600 dark:text-rose-400">
                  {{ l.current }} unit
                </div>
                <div class="text-[11px]" style="color: var(--txt-dim)">min {{ l.min }}</div>
              </div>
            </div>
          </div>
        </Panel>
      </div>

      <!-- recent tx -->
      <Panel title="Pergerakan Mutasi Terkini" icon="pi pi-history" dense>
        <template #actions>
          <Button label="Lihat Semua Mutasi" icon="pi pi-arrow-right" iconPos="right" text size="small"
                  @click="router.push('/history')" />
        </template>
        <EmptyState v-if="!stats.recent_tx?.length" icon="pi pi-history" title="Belum ada pergerakan"
                    sub="Catat barang masuk atau keluar untuk memulai jejak transaksi." />
        <div v-else class="overflow-x-auto">
          <table class="w-full text-[13px]">
            <thead>
              <tr class="text-left border-b" style="border-color: var(--line); background: var(--panel-2)">
                <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider" style="color: var(--txt-dim)">Waktu</th>
                <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider" style="color: var(--txt-dim)">Barang</th>
                <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider" style="color: var(--txt-dim)">Jenis</th>
                <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider text-right" style="color: var(--txt-dim)">Jumlah</th>
              </tr>
            </thead>
            <tbody class="divide-y" style="border-color: var(--line)">
              <tr v-for="(t, i) in stats.recent_tx" :key="i" class="hover:bg-slate-50 dark:hover:bg-slate-800/30 transition-colors">
                <td class="px-4 py-3 whitespace-nowrap text-[12px]" style="color: var(--txt-dim)">{{ t.time }}</td>
                <td class="px-4 py-3">
                  <span class="font-semibold" style="color: var(--txt)">{{ t.item_name }}</span>
                  <span class="t-mono ml-2 text-[11px] px-1.5 py-0.5 rounded border"
                        style="border-color: var(--line); color: var(--txt-dim); background: var(--panel-2)">{{ t.item_sku }}</span>
                </td>
                <td class="px-4 py-3"><StatusChip :kind="t.type" /></td>
                <td class="px-4 py-3 text-right t-num font-bold">{{ t.quantity }} {{ t.unit }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </Panel>
    </template>
  </div>
</template>
