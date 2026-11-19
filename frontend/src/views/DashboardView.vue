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
  try {
    stats.value = (await api.get('/dashboard')).data
  } finally {
    loading.value = false
  }
})

const catEntries = computed(() =>
  (Object.entries(stats.value.stock_by_cat || {}) as [string, number][]).sort((a, b) => b[1] - a[1])
)

const chartData = computed(() => ({
  labels: catEntries.value.map(e => e[0]),
  datasets: [{
    data: catEntries.value.map(e => e[1]),
    backgroundColor: theme.isDark ? 'rgba(45, 212, 191, 0.78)' : 'rgba(13, 148, 136, 0.85)',
    hoverBackgroundColor: theme.isDark ? '#2dd4bf' : '#0f766e',
    borderRadius: 5,
    barThickness: 20,
  }],
}))

const chartOptions = computed(() => ({
  indexAxis: 'y' as const,
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: {
      backgroundColor: theme.isDark ? '#182231' : '#0f172a',
      titleColor: '#f1f5f9',
      bodyColor: '#cbd5e1',
      borderColor: theme.isDark ? '#27364c' : '#334155',
      borderWidth: 1,
      padding: 10,
      cornerRadius: 7,
      displayColors: false,
      callbacks: {
        label: (ctx: any) => `${Number(ctx.raw).toLocaleString('id-ID')} unit`,
      },
    },
  },
  scales: {
    x: {
      border: { display: false },
      grid: { color: theme.isDark ? 'rgba(148, 163, 184, 0.08)' : 'rgba(15, 23, 42, 0.06)' },
      ticks: {
        font: { family: '"IBM Plex Mono", monospace', size: 11 },
        color: theme.isDark ? '#94a3b8' : '#64748b',
      },
    },
    y: {
      border: { display: false },
      grid: { display: false },
      ticks: {
        font: { family: '"Plus Jakarta Sans", sans-serif', size: 12, weight: 500 },
        color: theme.isDark ? '#cbd5e1' : '#334155',
      },
    },
  },
}))
</script>

<template>
  <div>
    <PageHeader
      crumb="Ringkasan Operasional"
      title="Dashboard Inventaris"
      sub="Pantau ketersediaan stok, sebaran kategori, dan pergerakan mutasi secara langsung."
    >
      <template #actions>
        <Button
          label="Mutasi Stok"
          icon="pi pi-arrow-right-arrow-left"
          size="small"
          severity="secondary"
          outlined
          @click="router.push('/movement')"
        />
        <Button
          label="Tambah Barang"
          icon="pi pi-plus"
          size="small"
          @click="router.push('/items/new')"
        />
      </template>
    </PageHeader>

    <!-- Layout-matching skeleton while loading -->
    <div v-if="loading" class="space-y-5" role="status" aria-live="polite">
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-3.5">
        <div v-for="n in 4" :key="n" class="panel p-4 space-y-3">
          <div class="flex justify-between">
            <div class="skeleton h-3.5 w-24" />
            <div class="skeleton h-7 w-7 rounded-[7px]" />
          </div>
          <div class="skeleton h-7 w-20" />
          <div class="skeleton h-3 w-32" />
        </div>
      </div>
      <div class="grid lg:grid-cols-[1.45fr_1fr] gap-4">
        <div class="panel p-4 space-y-3 h-[340px]">
          <div class="skeleton h-4 w-44" />
          <div v-for="n in 6" :key="n" class="skeleton h-5" :style="{ width: `${92 - n * 10}%` }" />
        </div>
        <div class="panel p-4 space-y-3 h-[340px]">
          <div class="skeleton h-4 w-40" />
          <div v-for="n in 5" :key="n" class="skeleton h-11 w-full" />
        </div>
      </div>
      <span class="sr-only">Memuat data dashboard…</span>
    </div>

    <template v-else>
      <!-- Differentiated KPI Row: Primary Featured Metric + 3 Secondary Metrics -->
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-3.5 mb-5">
        <StatCard
          label="Total Unit Fisik"
          :value="Number(stats.total_stock || 0).toLocaleString('id-ID')"
          icon="pi pi-database"
          tone="accent"
          featured
          hint="Akumulasi seluruh stok aktif"
        />
        <StatCard
          label="Jenis Barang"
          :value="Number(stats.total_items || 0).toLocaleString('id-ID')"
          icon="pi pi-box"
          tone="info"
          hint="SKU terdaftar di katalog"
        />
        <StatCard
          label="Stok Kritis"
          :value="stats.low_stock?.length || 0"
          icon="pi pi-exclamation-triangle"
          :tone="(stats.low_stock?.length || 0) > 0 ? 'bad' : 'ok'"
          hint="Berada di bawah batas minimum"
        />
        <StatCard
          label="Mutasi Tercatat"
          :value="stats.recent_tx?.length || 0"
          icon="pi pi-history"
          tone="neutral"
          hint="Aktivitas keluar-masuk terbaru"
        />
      </div>

      <!-- Analytics + Low Stock Attention -->
      <div class="grid lg:grid-cols-[1.45fr_1fr] gap-4 mb-5 items-start">
        <Panel title="Sebaran Stok per Kategori" icon="pi pi-chart-bar">
          <template #actions>
            <span class="t-mono !text-[11px] text-txt-dim">{{ catEntries.length }} kategori</span>
          </template>
          <EmptyState
            v-if="!catEntries.length"
            icon="pi pi-chart-bar"
            title="Belum ada data stok"
            sub="Tambahkan barang untuk melihat sebaran stok per kategori."
          />
          <div v-else class="h-[310px]">
            <Bar :data="chartData" :options="chartOptions" />
          </div>
        </Panel>

        <Panel title="Perlu Perhatian — Stok Menipis" icon="pi pi-exclamation-circle" dense>
          <template #actions>
            <span
              v-if="stats.low_stock?.length"
              class="t-mono !text-[11px] font-semibold text-sig-bad"
            >
              {{ stats.low_stock.length }} item
            </span>
          </template>
          <EmptyState
            v-if="!stats.low_stock?.length"
            icon="pi pi-check-circle"
            tone="ok"
            title="Semua stok aman"
            sub="Tidak ada barang di bawah batas minimum."
          />
          <div v-else class="divide-y divide-line max-h-[342px] overflow-y-auto">
            <button
              v-for="l in stats.low_stock"
              :key="l.sku"
              type="button"
              class="data-row w-full flex items-center gap-3 px-4 py-2.5 text-left cursor-pointer"
              @click="router.push(l.id ? `/items/${l.id}` : '/items')"
            >
              <div class="w-7 h-7 rounded-[7px] grid place-items-center bg-[var(--sig-bad-bg)] border border-[var(--sig-bad-line)] shrink-0">
                <i class="pi pi-exclamation-triangle text-sig-bad text-[11px]" aria-hidden="true" />
              </div>
              <div class="flex-1 min-w-0">
                <div class="text-[13px] font-semibold text-txt truncate">{{ l.name }}</div>
                <div class="t-mono !text-[11px] text-txt-dim truncate">{{ l.sku }} · {{ l.location || 'Tanpa lokasi' }}</div>
              </div>
              <div class="text-right shrink-0">
                <div class="t-num text-[13px] font-bold text-sig-bad">
                  {{ l.current }} unit
                </div>
                <div class="t-mono !text-[10.5px] text-txt-dim">min {{ l.min }}</div>
              </div>
            </button>
          </div>
        </Panel>
      </div>

      <!-- Recent Transactions Dense Table -->
      <Panel title="Pergerakan Mutasi Terkini" icon="pi pi-history" dense>
        <template #actions>
          <Button
            label="Semua Mutasi"
            icon="pi pi-arrow-right"
            iconPos="right"
            text
            size="small"
            severity="secondary"
            @click="router.push('/history')"
          />
        </template>

        <EmptyState
          v-if="!stats.recent_tx?.length"
          icon="pi pi-history"
          title="Belum ada pergerakan"
          sub="Catat barang masuk atau keluar untuk memulai jejak transaksi."
        >
          <Button label="Catat Mutasi" icon="pi pi-plus" size="small" @click="router.push('/movement')" />
        </EmptyState>

        <div v-else class="overflow-x-auto">
          <table class="w-full text-[13px] border-collapse">
            <thead>
              <tr class="text-left border-b border-line bg-panel-2">
                <th class="px-3.5 py-2.5 text-[11px] font-semibold uppercase tracking-[0.05em] text-txt-dim">Waktu</th>
                <th class="px-3.5 py-2.5 text-[11px] font-semibold uppercase tracking-[0.05em] text-txt-dim">Barang</th>
                <th class="px-3.5 py-2.5 text-[11px] font-semibold uppercase tracking-[0.05em] text-txt-dim">Jenis</th>
                <th class="px-3.5 py-2.5 text-[11px] font-semibold uppercase tracking-[0.05em] text-txt-dim text-right">Jumlah</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-line">
              <tr v-for="(t, i) in stats.recent_tx" :key="i" class="data-row">
                <td class="px-3.5 py-2.5 whitespace-nowrap t-mono !text-[11.5px] text-txt-dim">{{ t.time }}</td>
                <td class="px-3.5 py-2.5">
                  <div class="flex items-center gap-2 flex-wrap">
                    <span class="font-semibold text-txt">{{ t.item_name }}</span>
                    <span class="t-mono !text-[10.5px] px-1.5 py-0.5 rounded border border-line bg-panel-2 text-txt-dim">
                      {{ t.item_sku }}
                    </span>
                  </div>
                </td>
                <td class="px-3.5 py-2.5"><StatusChip :kind="t.type" /></td>
                <td class="px-3.5 py-2.5 text-right t-num font-bold text-txt whitespace-nowrap">
                  {{ t.quantity }} <span class="font-normal text-txt-dim text-[12px]">{{ t.unit }}</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </Panel>
    </template>
  </div>
</template>
