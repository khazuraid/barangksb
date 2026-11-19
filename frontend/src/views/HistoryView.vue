<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import api, { downloadAuthenticatedFile } from '@/api'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Dialog from 'primevue/dialog'
import PageHeader from '@/components/PageHeader.vue'
import Panel from '@/components/Panel.vue'
import StatCard from '@/components/StatCard.vue'
import StatusChip from '@/components/StatusChip.vue'
import EmptyState from '@/components/EmptyState.vue'

const router = useRouter()

const txs = ref<any[]>([])
const total = ref(0)
const page = ref(0)
const perPage = ref(25)
const sku = ref('')
const type = ref('')
const loadingTx = ref(true)
const expandedTx = ref<string | null>(null)
const photoLightbox = ref(false)
const lightboxUrl = ref('')

const typeOptions = [
  { label: 'Semua jenis mutasi', value: '' },
  { label: 'Barang Masuk (IN)', value: 'IN' },
  { label: 'Barang Keluar (OUT)', value: 'OUT' },
  { label: 'Opname Bertambah (ADJUST+)', value: 'ADJUST+' },
  { label: 'Opname Berkurang (ADJUST-)', value: 'ADJUST-' },
]
const perPageOptions = [25, 50, 100]

async function fetchTx(reset = false) {
  if (reset) page.value = 0
  loadingTx.value = true
  try {
    const res = await api.get('/transactions', {
      params: { sku: sku.value, type: type.value, page: page.value + 1, per_page: perPage.value },
    })
    const list = res.data?.data || (Array.isArray(res.data) ? res.data : [])
    txs.value = Array.isArray(list) ? list : []
    total.value = res.data?.total || txs.value.length
  } catch {
    txs.value = []
    total.value = 0
  } finally {
    loadingTx.value = false
  }
}

onMounted(() => fetchTx())
watch([page, perPage], () => fetchTx())

const lastPage = computed(() => Math.max(0, Math.ceil(total.value / perPage.value) - 1))
const range = computed(() => {
  if (!total.value) return '0 transaksi'
  const from = page.value * perPage.value + 1
  const to = Math.min(total.value, (page.value + 1) * perPage.value)
  return `${from}–${to} dari ${total.value}`
})

const summary = computed(() => {
  const safeList = Array.isArray(txs.value) ? txs.value : []
  return {
    total: total.value,
    in: safeList.filter((t) => t.type === 'IN').reduce((a, t) => a + (t.quantity || 0), 0),
    out: safeList.filter((t) => t.type === 'OUT').reduce((a, t) => a + (t.quantity || 0), 0),
    adj: safeList.filter((t) => (t.type || '').startsWith('ADJUST')).length,
  }
})

async function exportAs(kind: 'csv' | 'xlsx' | 'pdf') {
  const path = kind === 'csv' ? '/export/tx.csv' : kind === 'xlsx' ? '/export/items.xlsx' : '/report.pdf'
  const filename = kind === 'csv' ? 'riwayat_mutasi.csv' : kind === 'xlsx' ? 'master_barang.xlsx' : 'laporan_inventaris.pdf'
  await downloadAuthenticatedFile(path, filename, kind === 'pdf')
}

function sanitizeCsvCell(v: unknown): string {
  const s = String(v ?? '')
  if (/^[\s]*[=+\-@]/.test(s) || /^[\t\r\n]/.test(s)) {
    return `'${s}`
  }
  return s
}

function exportFiltered() {
  const header = ['Waktu', 'Jenis', 'SKU', 'Nama', 'Qty', 'Satuan', 'Sebelum', 'Sesudah', 'Petugas', 'Catatan']
  const rows = (txs.value || []).map((t) => [
    t.timestamp,
    t.type,
    t.item_sku,
    t.item_name,
    t.quantity,
    t.unit,
    t.previous_stock,
    t.new_stock,
    t.received_by || '',
    (t.notes || '').replace(/\n/g, ' '),
  ])
  const csv = [header, ...rows]
    .map((r) => r.map((v) => `"${sanitizeCsvCell(v).replace(/"/g, '""')}"`).join(','))
    .join('\n')
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'riwayat_mutasi_terfilter.csv'
  a.click()
  URL.revokeObjectURL(url)
}

function openPhoto(url: string) {
  lightboxUrl.value = url
  photoLightbox.value = true
}
</script>

<template>
  <div>
    <PageHeader
      crumb="Operasional"
      title="Log &amp; Riwayat Mutasi"
      sub="Jejak lengkap seluruh pergerakan stok masuk, keluar, dan penyesuaian opname."
    >
      <template #actions>
        <Button label="Catat Mutasi" icon="pi pi-plus" size="small" @click="router.push('/movement')" />
      </template>
    </PageHeader>

    <div class="grid grid-cols-2 lg:grid-cols-4 gap-3.5 mb-5">
      <StatCard
        label="Total Transaksi"
        :value="summary.total"
        icon="pi pi-history"
        tone="accent"
        featured
        hint="Seluruh jejak mutasi"
      />
      <StatCard
        label="Unit Masuk (Hal. Ini)"
        :value="`+${summary.in}`"
        icon="pi pi-arrow-down"
        tone="ok"
        hint="Akumulasi barang masuk"
      />
      <StatCard
        label="Unit Keluar (Hal. Ini)"
        :value="`−${summary.out}`"
        icon="pi pi-arrow-up"
        tone="bad"
        hint="Akumulasi barang keluar"
      />
      <StatCard
        label="Opname / Adjust"
        :value="summary.adj"
        icon="pi pi-sliders-h"
        tone="warn"
        hint="Penyesuaian stok halaman ini"
      />
    </div>

    <!-- Filter & Export Toolbar -->
    <div class="panel p-3 mb-4 flex flex-wrap items-center justify-between gap-2.5">
      <div class="flex flex-wrap items-center gap-2 flex-1 min-w-[260px]">
        <div class="relative flex-1 min-w-[180px] max-w-xs">
          <i class="pi pi-search absolute left-3 top-1/2 -translate-y-1/2 text-txt-dim text-[12px]" aria-hidden="true" />
          <InputText
            v-model="sku"
            placeholder="Filter SKU barang…"
            class="w-full !pl-8 !text-[12.5px] t-mono"
            @keyup.enter="fetchTx(true)"
          />
        </div>
        <Select
          v-model="type"
          :options="typeOptions"
          optionLabel="label"
          optionValue="value"
          class="w-full sm:w-52 !text-[12.5px]"
          @change="fetchTx(true)"
        />
        <Button label="Terapkan" icon="pi pi-filter" size="small" @click="fetchTx(true)" />
        <Button
          v-if="sku || type"
          icon="pi pi-filter-slash"
          text
          rounded
          size="small"
          severity="secondary"
          aria-label="Reset filter"
          @click="sku = ''; type = ''; fetchTx(true)"
        />
      </div>

      <div class="flex items-center gap-1.5">
        <Button label="CSV Terfilter" icon="pi pi-download" size="small" text severity="secondary" @click="exportFiltered" />
        <Button label="CSV Semua" icon="pi pi-file" size="small" text severity="secondary" @click="exportAs('csv')" />
        <Button label="Excel" icon="pi pi-file-excel" size="small" text severity="secondary" @click="exportAs('xlsx')" />
        <Button label="PDF" icon="pi pi-file-pdf" size="small" text severity="secondary" @click="exportAs('pdf')" />
      </div>
    </div>

    <Panel title="Daftar Mutasi Stok" icon="pi pi-list" dense>
      <template #actions>
        <span class="t-mono !text-[11px] text-txt-dim">{{ range }}</span>
      </template>

      <EmptyState v-if="loadingTx" loading title="Memuat riwayat mutasi…" />

      <EmptyState
        v-else-if="!txs.length"
        icon="pi pi-history"
        title="Belum ada transaksi mutasi"
        sub="Riwayat barang masuk, keluar, dan opname akan muncul di sini."
      >
        <Button label="Catat Mutasi" icon="pi pi-plus" size="small" @click="router.push('/movement')" />
      </EmptyState>

      <div v-else class="overflow-x-auto">
        <table class="w-full text-[13px] border-collapse">
          <thead>
            <tr class="text-left border-b border-line bg-panel-2 text-[11px] uppercase tracking-[0.05em] text-txt-dim font-semibold">
              <th class="px-3.5 py-2.5 w-[160px]">Waktu</th>
              <th class="px-3.5 py-2.5 w-[120px]">Jenis</th>
              <th class="px-3.5 py-2.5">Barang &amp; SKU</th>
              <th class="px-3.5 py-2.5 text-right w-[110px]">Jumlah</th>
              <th class="px-3.5 py-2.5 text-right w-[130px]">Perubahan Stok</th>
              <th class="px-3.5 py-2.5">Petugas / Catatan</th>
              <th class="px-3.5 py-2.5 text-center w-[70px]">Bukti</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-line">
            <template v-for="(t, i) in txs" :key="t.id || i">
              <tr
                class="data-row cursor-pointer"
                :class="{ 'data-row-active': expandedTx === (t.id || String(i)) }"
                @click="expandedTx = expandedTx === (t.id || String(i)) ? null : (t.id || String(i))"
              >
                <td class="px-3.5 py-2.5 whitespace-nowrap t-mono !text-[11.5px] text-txt-dim">{{ t.timestamp }}</td>
                <td class="px-3.5 py-2.5 whitespace-nowrap"><StatusChip :kind="t.type" /></td>
                <td class="px-3.5 py-2.5">
                  <div class="font-semibold text-txt">{{ t.item_name }}</div>
                  <div class="t-mono !text-[11px] text-txt-dim">{{ t.item_sku }}</div>
                </td>
                <td class="px-3.5 py-2.5 text-right t-num font-bold whitespace-nowrap" :class="t.type === 'OUT' || t.type === 'ADJUST-' ? 'text-sig-bad' : 'text-sig-ok'">
                  {{ t.type === 'OUT' || t.type === 'ADJUST-' ? '−' : '+' }}{{ t.quantity }}
                  <span class="font-normal text-txt-dim text-[11.5px]">{{ t.unit }}</span>
                </td>
                <td class="px-3.5 py-2.5 text-right t-mono !text-[11.5px] text-txt-dim whitespace-nowrap">
                  {{ t.previous_stock }} → <strong class="text-txt">{{ t.new_stock }}</strong>
                </td>
                <td class="px-3.5 py-2.5 max-w-xs">
                  <div class="text-[12.5px] font-medium text-txt truncate">{{ t.received_by || '—' }}</div>
                  <div v-if="t.notes" class="text-[11.5px] text-txt-dim truncate">{{ t.notes }}</div>
                </td>
                <td class="px-3.5 py-2.5 text-center" @click.stop>
                  <Button
                    v-if="t.photo_url"
                    icon="pi pi-image"
                    text
                    rounded
                    size="small"
                    severity="secondary"
                    aria-label="Lihat foto bukti mutasi"
                    @click="openPhoto(t.photo_url)"
                  />
                  <span v-else class="text-txt-muted">—</span>
                </td>
              </tr>

              <tr v-if="expandedTx === (t.id || String(i))" class="bg-panel-2">
                <td colspan="7" class="px-4 py-3 text-[12px]">
                  <div class="flex flex-wrap items-center justify-between gap-3">
                    <div class="space-y-1">
                      <div><span class="text-txt-dim">Catatan lengkap:</span> <span class="text-txt">{{ t.notes || 'Tidak ada catatan' }}</span></div>
                      <div v-if="t.received_by"><span class="text-txt-dim">Penerima / Petugas:</span> <span class="text-txt font-medium">{{ t.received_by }}</span></div>
                    </div>
                    <Button
                      v-if="t.item_id"
                      label="Buka Detail Barang"
                      icon="pi pi-external-link"
                      size="small"
                      severity="secondary"
                      outlined
                      @click="router.push(`/items/${t.item_id}`)"
                    />
                  </div>
                </td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>

      <template #footer>
        <div class="flex flex-wrap items-center justify-between gap-3">
          <span class="t-mono !text-[11.5px] text-txt-dim">{{ range }}</span>
          <div class="flex items-center gap-2">
            <Select v-model="perPage" :options="perPageOptions" class="!text-[12px] w-[96px]" />
            <Button icon="pi pi-angle-left" size="small" text severity="secondary" aria-label="Halaman sebelumnya" :disabled="page === 0" @click="page--" />
            <span class="t-num text-[12px] px-1 text-txt">Hal. {{ page + 1 }} / {{ lastPage + 1 }}</span>
            <Button icon="pi pi-angle-right" size="small" text severity="secondary" aria-label="Halaman berikutnya" :disabled="page >= lastPage" @click="page++" />
          </div>
        </div>
      </template>
    </Panel>

    <Dialog v-model:visible="photoLightbox" modal header="Bukti Foto Mutasi" :style="{ width: '90vw', maxWidth: '720px' }">
      <div class="flex justify-center bg-ink-950 p-2 rounded-[8px] border border-line">
        <img :src="lightboxUrl" alt="Foto bukti transaksi" class="max-h-[70vh] object-contain rounded" />
      </div>
    </Dialog>
  </div>
</template>
