<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import api from '@/api'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Dialog from 'primevue/dialog'
import PageHeader from '@/components/PageHeader.vue'
import Panel from '@/components/Panel.vue'
import StatusChip from '@/components/StatusChip.vue'
import EmptyState from '@/components/EmptyState.vue'

const router = useRouter()

const txs = ref<any[]>([])
const total = ref(0)
const page = ref(0)
const perPage = ref(25)
const sku = ref('')
const type = ref('')
const loadingTx = ref(false)
const expandedTx = ref<string | null>(null)
const photoLightbox = ref(false)
const lightboxUrl = ref('')

const typeOptions = [
  { label: 'Semua jenis mutasi', value: '' },
  { label: 'Barang Masuk', value: 'IN' },
  { label: 'Barang Keluar', value: 'OUT' },
  { label: 'Opname (+)', value: 'ADJUST+' },
  { label: 'Opname (−)', value: 'ADJUST-' },
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
  if (!total.value) return '0'
  const from = page.value * perPage.value + 1
  const to = Math.min(total.value, (page.value + 1) * perPage.value)
  return `${from}–${to} dari ${total.value}`
})

const summary = computed(() => {
  const safeList = Array.isArray(txs.value) ? txs.value : []
  return {
    in: safeList.filter((t) => t.type === 'IN').reduce((a, t) => a + (t.quantity || 0), 0),
    out: safeList.filter((t) => t.type === 'OUT').reduce((a, t) => a + (t.quantity || 0), 0),
    adj: safeList.filter((t) => (t.type || '').startsWith('ADJUST')).length,
  }
})

function exportAs(kind: 'csv' | 'xlsx' | 'pdf') {
  const url = kind === 'csv' ? '/api/export/tx.csv' : kind === 'xlsx' ? '/api/export/items.xlsx' : '/api/report.pdf'
  window.open(url, '_blank')
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
    .map((r) => r.map((v) => `"${String(v ?? '').replace(/"/g, '""')}"`).join(','))
    .join('\n')
  const blob = new Blob(['﻿' + csv], { type: 'text/csv;charset=utf-8' })
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = `riwayat_mutasi_halaman_${page.value + 1}.csv`
  a.click()
  URL.revokeObjectURL(a.href)
}
</script>

<template>
  <div class="pb-16 max-w-6xl mx-auto">
    <PageHeader
      crumb="Operasional"
      title="Log &amp; Riwayat Mutasi"
      sub="Rekaman transaksi pergerakan stok barang masuk, keluar, dan opname fisik"
    >
      <template #actions>
        <Button
          label="Halaman ini (CSV)"
          icon="pi pi-file"
          size="small"
          severity="secondary"
          outlined
          :disabled="!txs.length"
          @click="exportFiltered"
        />
        <Button
          label="Ekspor Semua"
          icon="pi pi-download"
          size="small"
          :disabled="!txs.length"
          @click="exportAs('csv')"
        />
      </template>
    </PageHeader>

    <!-- Quick Metric KPI Strip -->
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-3 mb-5 max-w-2xl">
      <div class="panel p-3.5 flex items-center justify-between">
        <div>
          <div class="text-[11px] font-semibold uppercase tracking-wider" style="color: var(--txt-dim)">Unit Masuk</div>
          <div class="t-num text-[22px] font-bold text-emerald-600 dark:text-emerald-400 mt-0.5">+{{ summary.in }}</div>
        </div>
        <div class="w-8 h-8 rounded-lg grid place-items-center bg-emerald-50 dark:bg-emerald-950/40 border border-emerald-200 dark:border-emerald-800 text-emerald-600 dark:text-emerald-400 text-xs">
          <i class="pi pi-arrow-down-left" />
        </div>
      </div>

      <div class="panel p-3.5 flex items-center justify-between">
        <div>
          <div class="text-[11px] font-semibold uppercase tracking-wider" style="color: var(--txt-dim)">Unit Keluar</div>
          <div class="t-num text-[22px] font-bold text-rose-600 dark:text-rose-400 mt-0.5">−{{ summary.out }}</div>
        </div>
        <div class="w-8 h-8 rounded-lg grid place-items-center bg-rose-50 dark:bg-rose-950/40 border border-rose-200 dark:border-rose-800 text-rose-600 dark:text-rose-400 text-xs">
          <i class="pi pi-arrow-up-right" />
        </div>
      </div>

      <div class="panel p-3.5 flex items-center justify-between">
        <div>
          <div class="text-[11px] font-semibold uppercase tracking-wider" style="color: var(--txt-dim)">Baris Opname</div>
          <div class="t-num text-[22px] font-bold text-indigo-600 dark:text-indigo-400 mt-0.5">{{ summary.adj }}</div>
        </div>
        <div class="w-8 h-8 rounded-lg grid place-items-center bg-indigo-50 dark:bg-indigo-950/40 border border-indigo-200 dark:border-indigo-800 text-indigo-600 dark:text-indigo-400 text-xs">
          <i class="pi pi-sliders-h" />
        </div>
      </div>
    </div>

    <!-- Transaction List Panel -->
    <Panel title="Jejak Mutasi Barang" icon="pi pi-history" dense>
      <template #actions>
        <InputText
          v-model="sku"
          placeholder="Cari SKU / nama…"
          class="!text-[12.5px] !py-1.5 w-[180px]"
          @keyup.enter="fetchTx(true)"
        />
        <Select
          v-model="type"
          :options="typeOptions"
          optionLabel="label"
          optionValue="value"
          class="!text-[12.5px] w-[180px]"
          @change="fetchTx(true)"
        />
        <Button icon="pi pi-search" size="small" severity="secondary" outlined @click="fetchTx(true)" />
      </template>

      <EmptyState v-if="loadingTx" icon="pi pi-spin pi-spinner" title="Memuat riwayat perubahan…" />
      <EmptyState
        v-else-if="!txs.length"
        icon="pi pi-history"
        title="Belum ada transaksi mutasi"
        sub="Pencatatan barang masuk atau keluar akan terekam otomatis di sini lengkap dengan foto geotag."
      >
        <Button label="Daftar Barang" icon="pi pi-box" size="small" @click="router.push('/items')" />
      </EmptyState>

      <div v-else class="overflow-x-auto">
        <table class="w-full text-[12.5px]">
          <thead>
            <tr class="text-left border-b" style="border-color: var(--line); background: var(--panel-2)">
              <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider w-[170px]" style="color: var(--txt-dim)">Waktu</th>
              <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider w-[120px]" style="color: var(--txt-dim)">Jenis</th>
              <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider" style="color: var(--txt-dim)">Barang</th>
              <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider w-[120px] text-right" style="color: var(--txt-dim)">Perubahan</th>
              <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider w-[130px] text-right" style="color: var(--txt-dim)">Stok</th>
              <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider w-[60px]" style="color: var(--txt-dim)"></th>
            </tr>
          </thead>
          <tbody class="divide-y" style="border-color: var(--line)">
            <template v-for="t in txs" :key="t.id">
              <tr
                class="hover:bg-slate-50 dark:hover:bg-slate-800/40 transition-colors cursor-pointer"
                @click="expandedTx = expandedTx === t.id ? null : t.id"
              >
                <td class="px-4 py-3 whitespace-nowrap" style="color: var(--txt-dim)">{{ t.timestamp }}</td>
                <td class="px-4 py-3"><StatusChip :kind="t.type" /></td>
                <td class="px-4 py-3">
                  <div class="flex items-center gap-2 font-semibold">
                    <span style="color: var(--txt)">{{ t.item_name }}</span>
                    <button
                      v-if="t.photo_url"
                      class="inline-flex text-indigo-600 dark:text-indigo-400 hover:opacity-80 cursor-pointer"
                      title="Lihat foto bukti geotag"
                      @click.stop="lightboxUrl = t.photo_url; photoLightbox = true"
                    >
                      <i class="pi pi-camera text-[12px]" />
                    </button>
                  </div>
                  <div class="t-mono text-[11px] mt-0.5" style="color: var(--txt-dim)">{{ t.item_sku }}</div>
                </td>
                <td class="px-4 py-3 text-right whitespace-nowrap">
                  <span
                    class="t-num font-bold"
                    :class="t.type === 'IN' || t.type === 'ADJUST+' ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400'"
                  >
                    {{ t.type === 'IN' || t.type === 'ADJUST+' ? '+' : '−' }}{{ t.quantity }} {{ t.unit }}
                  </span>
                </td>
                <td class="px-4 py-3 text-right t-num whitespace-nowrap" style="color: var(--txt-dim)">
                  {{ t.previous_stock }}
                  <i class="pi pi-arrow-right text-[9px] mx-1" />
                  <span class="font-bold" style="color: var(--txt)">{{ t.new_stock }}</span>
                </td>
                <td class="px-4 py-3 text-right">
                  <i
                    class="pi text-[11px]"
                    :class="expandedTx === t.id ? 'pi-chevron-up' : 'pi-chevron-down'"
                    style="color: var(--txt-dim)"
                  />
                </td>
              </tr>

              <!-- Expanded Details -->
              <tr v-if="expandedTx === t.id">
                <td colspan="6" class="px-4 py-3.5" style="background: var(--panel-2)">
                  <div class="flex flex-wrap sm:flex-nowrap gap-4 items-center">
                    <div
                      v-if="t.photo_url"
                      class="relative group w-28 h-20 shrink-0 rounded-lg overflow-hidden border bg-black/40 cursor-pointer"
                      style="border-color: var(--line)"
                      @click="lightboxUrl = t.photo_url; photoLightbox = true"
                    >
                      <img :src="t.photo_url" alt="Bukti Transaksi" class="w-full h-full object-cover" />
                      <div class="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 flex items-center justify-center text-white transition-opacity">
                        <i class="pi pi-search-plus text-sm" />
                      </div>
                    </div>

                    <div class="grid sm:grid-cols-3 gap-3 text-[12.5px] flex-1">
                      <div>
                        <div class="text-[10.5px] font-semibold uppercase tracking-wider mb-1" style="color: var(--txt-dim)">Petugas / Penerima</div>
                        <div class="font-semibold" style="color: var(--txt)">{{ t.received_by || '—' }}</div>
                      </div>
                      <div>
                        <div class="text-[10.5px] font-semibold uppercase tracking-wider mb-1" style="color: var(--txt-dim)">Catatan Transaksi</div>
                        <div style="color: var(--txt)">{{ t.notes || '—' }}</div>
                      </div>
                      <div>
                        <div class="text-[10.5px] font-semibold uppercase tracking-wider mb-1" style="color: var(--txt-dim)">ID Transaksi</div>
                        <div class="t-mono text-[11px]" style="color: var(--txt-dim)">{{ t.id }}</div>
                      </div>
                    </div>
                  </div>
                </td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>

      <template #footer>
        <div class="flex flex-wrap items-center justify-between gap-3">
          <span class="text-[11.5px]" style="color: var(--txt-dim)">{{ range }}</span>
          <div class="flex items-center gap-2">
            <Select v-model="perPage" :options="perPageOptions" class="!text-[12px] !py-1 w-[95px]" />
            <Button
              icon="pi pi-angle-left"
              size="small"
              text
              severity="secondary"
              :disabled="page === 0"
              @click="page--"
            />
            <span class="t-num text-[12px] px-1">Hal. {{ page + 1 }} / {{ lastPage + 1 }}</span>
            <Button
              icon="pi pi-angle-right"
              size="small"
              text
              severity="secondary"
              :disabled="page >= lastPage"
              @click="page++"
            />
          </div>
        </div>
      </template>
    </Panel>

    <!-- Modal Lightbox Foto Geotag -->
    <Dialog v-model:visible="photoLightbox" modal header="Foto Bukti Geotag Transaksi" :style="{ width: '640px' }" class="p-fluid">
      <div v-if="lightboxUrl" class="p-2 bg-black rounded-lg flex items-center justify-center">
        <img :src="lightboxUrl" alt="Foto Geotag Transaksi" class="max-h-[75vh] object-contain rounded" />
      </div>
    </Dialog>
  </div>
</template>
