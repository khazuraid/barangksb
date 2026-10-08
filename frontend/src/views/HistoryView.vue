<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '@/api'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Dialog from 'primevue/dialog'
import PageHeader from '@/components/PageHeader.vue'
import Panel from '@/components/Panel.vue'
import StatusChip from '@/components/StatusChip.vue'
import EmptyState from '@/components/EmptyState.vue'

const route = useRoute()
const router = useRouter()

// Tab state: 'changes' (Histori Perubahan) vs 'audit' (Audit Log / Keamanan)
const activeTab = ref<'changes' | 'audit'>('changes')

onMounted(() => {
  if (route.query.tab === 'audit' || route.path === '/audit') {
    activeTab.value = 'audit'
  }
})

watch(
  () => [route.query.tab, route.path],
  ([tab, path]) => {
    if (tab === 'audit' || path === '/audit') {
      activeTab.value = 'audit'
    } else if (tab === 'changes' || path === '/history') {
      activeTab.value = 'changes'
    }
  }
)

function switchTab(t: 'changes' | 'audit') {
  activeTab.value = t
  router.replace({ path: '/history', query: t === 'audit' ? { tab: 'audit' } : {} })
}

// ==========================================
// TAB 1: HISTORI PERUBAHAN & MUTASI STOK
// ==========================================
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
  { label: 'Semua jenis', value: '' },
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
    txs.value = res.data.data
    total.value = res.data.total
  } finally {
    loadingTx.value = false
  }
}

const lastPage = computed(() => Math.max(0, Math.ceil(total.value / perPage.value) - 1))
const range = computed(() => {
  if (!total.value) return '0'
  const from = page.value * perPage.value + 1
  const to = Math.min(total.value, (page.value + 1) * perPage.value)
  return `${from}–${to} dari ${total.value}`
})

const summary = computed(() => ({
  in: txs.value.filter((t) => t.type === 'IN').reduce((a, t) => a + t.quantity, 0),
  out: txs.value.filter((t) => t.type === 'OUT').reduce((a, t) => a + t.quantity, 0),
  adj: txs.value.filter((t) => t.type.startsWith('ADJUST')).length,
}))

function exportAs(kind: 'csv' | 'xlsx' | 'pdf') {
  const url = kind === 'csv' ? '/api/export/tx.csv' : kind === 'xlsx' ? '/api/export/items.xlsx' : '/api/report.pdf'
  window.open(url, '_blank')
}

function exportFiltered() {
  const header = ['Waktu', 'Jenis', 'SKU', 'Nama', 'Qty', 'Satuan', 'Sebelum', 'Sesudah', 'Petugas', 'Catatan']
  const rows = txs.value.map((t) => [
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

// ==========================================
// TAB 2: AUDIT LOG & KEAMANAN SISTEM
// ==========================================
const auditLogs = ref<any[]>([])
const loadingAudit = ref(false)
const auditQuery = ref('')
const opFilter = ref('')
const expandedAudit = ref<string | null>(null)

const opOptions = [
  { label: 'Semua aksi', value: '' },
  { label: 'INSERT (Tambah)', value: 'INSERT' },
  { label: 'UPDATE (Ubah)', value: 'UPDATE' },
  { label: 'DELETE (Hapus)', value: 'DELETE' },
]

async function fetchAudit() {
  loadingAudit.value = true
  try {
    const res = await api.get('/audit')
    auditLogs.value = res.data
  } finally {
    loadingAudit.value = false
  }
}

const filteredAudit = computed(() =>
  auditLogs.value.filter(
    (l) =>
      (!opFilter.value || l.op === opFilter.value) &&
      (!auditQuery.value ||
        [l.table_name, l.row_id, l.op].join(' ').toLowerCase().includes(auditQuery.value.toLowerCase()))
  )
)

const auditCounts = computed(() => ({
  total: auditLogs.value.length,
  ins: auditLogs.value.filter((l) => l.op === 'INSERT').length,
  upd: auditLogs.value.filter((l) => l.op === 'UPDATE').length,
  del: auditLogs.value.filter((l) => l.op === 'DELETE').length,
}))

function prettyJSON(raw: any) {
  if (!raw) return '—'
  try {
    return JSON.stringify(typeof raw === 'string' ? JSON.parse(raw) : raw, null, 2)
  } catch {
    return String(raw)
  }
}

// Initial fetch
onMounted(() => {
  fetchTx()
  fetchAudit()
})
watch([page, perPage], () => fetchTx())
</script>

<template>
  <div class="pb-16 max-w-6xl mx-auto">
    <PageHeader
      crumb="Pusat Log"
      title="Log &amp; Riwayat Sistem"
      sub="Pemantauan lengkap pergerakan stok, perubahan data, dan jejak audit keamanan"
    >
      <template #actions>
        <template v-if="activeTab === 'changes'">
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
        <template v-else>
          <Button
            label="Segarkan Log"
            icon="pi pi-refresh"
            size="small"
            severity="secondary"
            outlined
            :loading="loadingAudit"
            @click="fetchAudit"
          />
        </template>
      </template>
    </PageHeader>

    <!-- 2 Unified Tabs Switcher -->
    <div
      class="panel p-1.5 mb-5 rounded-xl border flex items-center gap-1.5"
      style="background: var(--paper-1); border-color: var(--line)"
    >
      <button
        class="flex-1 py-2.5 px-4 rounded-lg text-[13px] font-bold flex items-center justify-center gap-2 transition-all cursor-pointer border"
        :class="activeTab === 'changes'
          ? 'bg-acc-500 text-ink-950 border-acc-500 shadow-md font-extrabold'
          : 'bg-transparent text-ink-300 border-transparent hover:bg-paper-2 hover:text-ink-100'"
        @click="switchTab('changes')"
      >
        <i class="pi pi-history text-xs" />
        <span>Histori Perubahan &amp; Mutasi Stok</span>
        <span
          class="text-[10.5px] px-2 py-0.2 rounded-full font-bold"
          :class="activeTab === 'changes' ? 'bg-ink-950/20 text-ink-950' : 'bg-paper-2 text-ink-400'"
        >
          {{ total }}
        </span>
      </button>

      <button
        class="flex-1 py-2.5 px-4 rounded-lg text-[13px] font-bold flex items-center justify-center gap-2 transition-all cursor-pointer border"
        :class="activeTab === 'audit'
          ? 'bg-acc-500 text-ink-950 border-acc-500 shadow-md font-extrabold'
          : 'bg-transparent text-ink-300 border-transparent hover:bg-paper-2 hover:text-ink-100'"
        @click="switchTab('audit')"
      >
        <i class="pi pi-shield text-xs" />
        <span>Audit Log &amp; Keamanan</span>
        <span
          class="text-[10.5px] px-2 py-0.2 rounded-full font-bold"
          :class="activeTab === 'audit' ? 'bg-ink-950/20 text-ink-950' : 'bg-paper-2 text-ink-400'"
        >
          {{ auditLogs.length }}
        </span>
      </button>
    </div>

    <!-- ======================================================== -->
    <!-- TAB 1: HISTORI PERUBAHAN & MUTASI STOK -->
    <!-- ======================================================== -->
    <div v-if="activeTab === 'changes'" class="flex flex-col gap-4">
      <!-- Quick Metric KPI Strip -->
      <div
        class="grid grid-cols-3 gap-px rounded-xl overflow-hidden border max-w-xl shadow-xs"
        style="border-color: var(--line); background: var(--line)"
      >
        <div class="px-4 py-3" style="background: var(--panel)">
          <div class="t-label">Unit Masuk (hal. ini)</div>
          <div class="t-num text-[22px] font-bold mt-1 text-sig-ok">+{{ summary.in }}</div>
        </div>
        <div class="px-4 py-3" style="background: var(--panel)">
          <div class="t-label">Unit Keluar (hal. ini)</div>
          <div class="t-num text-[22px] font-bold mt-1 text-sig-bad">−{{ summary.out }}</div>
        </div>
        <div class="px-4 py-3" style="background: var(--panel)">
          <div class="t-label">Baris Opname</div>
          <div class="t-num text-[22px] font-bold mt-1 text-acc-500">{{ summary.adj }}</div>
        </div>
      </div>

      <!-- Transaction List Panel -->
      <Panel title="Jejak Mutasi Barang" icon="pi pi-history" dense>
        <template #actions>
          <InputText
            v-model="sku"
            placeholder="Cari SKU / nama…"
            class="!text-[12px] !py-1.5 w-[160px]"
            @keyup.enter="fetchTx(true)"
          />
          <Select
            v-model="type"
            :options="typeOptions"
            optionLabel="label"
            optionValue="value"
            class="!text-[12px] w-[155px]"
            @change="fetchTx(true)"
          />
          <Button icon="pi pi-search" size="small" severity="secondary" outlined @click="fetchTx(true)" />
        </template>

        <EmptyState v-if="loadingTx" icon="pi pi-spin pi-spinner" title="Memuat riwayat perubahan…" />
        <EmptyState
          v-else-if="!txs.length"
          icon="pi pi-history"
          title="Belum ada transaksi"
          sub="Riwayat transaksi pergerakan stok barang akan terekam otomatis di sini."
        >
          <Button label="Daftar Barang" icon="pi pi-box" size="small" @click="$router.push('/items')" />
        </EmptyState>

        <div v-else class="overflow-x-auto">
          <table class="w-full text-[12.5px]">
            <thead>
              <tr class="text-left" style="background: var(--paper-2)">
                <th class="t-label px-4 py-2.5 w-[170px]">Waktu</th>
                <th class="t-label px-4 py-2.5 w-[120px]">Jenis</th>
                <th class="t-label px-4 py-2.5">Barang</th>
                <th class="t-label px-4 py-2.5 w-[120px] text-right">Perubahan</th>
                <th class="t-label px-4 py-2.5 w-[130px] text-right">Stok</th>
                <th class="t-label px-4 py-2.5 w-[60px]"></th>
              </tr>
            </thead>
            <tbody class="divide-y" style="border-color: var(--line-soft)">
              <template v-for="t in txs" :key="t.id">
                <tr
                  class="hover:bg-paper-2 transition-colors cursor-pointer"
                  @click="expandedTx = expandedTx === t.id ? null : t.id"
                >
                  <td class="px-4 py-2.5 whitespace-nowrap" style="color: var(--txt-dim)">{{ t.timestamp }}</td>
                  <td class="px-4 py-2.5"><StatusChip :kind="t.type" /></td>
                  <td class="px-4 py-2.5">
                    <div class="flex items-center gap-1.5 font-semibold">
                      <span>{{ t.item_name }}</span>
                      <button
                        v-if="t.photo_url"
                        class="inline-flex text-acc-500 hover:text-acc-400 cursor-pointer"
                        title="Lihat foto bukti geotag"
                        @click.stop="lightboxUrl = t.photo_url; photoLightbox = true"
                      >
                        <i class="pi pi-camera text-[11px]" />
                      </button>
                    </div>
                    <div class="t-mono text-[11px]" style="color: var(--txt-dim)">{{ t.item_sku }}</div>
                  </td>
                  <td class="px-4 py-2.5 text-right whitespace-nowrap">
                    <span
                      class="t-num font-bold"
                      :class="t.type === 'IN' || t.type === 'ADJUST+' ? 'text-sig-ok' : 'text-sig-bad'"
                    >
                      {{ t.type === 'IN' || t.type === 'ADJUST+' ? '+' : '−' }}{{ t.quantity }} {{ t.unit }}
                    </span>
                  </td>
                  <td class="px-4 py-2.5 text-right t-num whitespace-nowrap" style="color: var(--txt-dim)">
                    {{ t.previous_stock }}
                    <i class="pi pi-arrow-right text-[9px] mx-0.5" />
                    <span class="font-semibold" style="color: var(--txt)">{{ t.new_stock }}</span>
                  </td>
                  <td class="px-4 py-2.5 text-right">
                    <i
                      class="pi text-[10px]"
                      :class="expandedTx === t.id ? 'pi-chevron-up' : 'pi-chevron-down'"
                      style="color: var(--txt-dim)"
                    />
                  </td>
                </tr>

                <!-- Expanded Details -->
                <tr v-if="expandedTx === t.id">
                  <td colspan="6" class="px-4 pb-3.5 pt-0" style="background: var(--paper-1)">
                    <div class="flex flex-wrap sm:flex-nowrap gap-4 items-center pt-2">
                      <div
                        v-if="t.photo_url"
                        class="relative group w-28 h-20 shrink-0 rounded overflow-hidden border bg-black/40 cursor-pointer"
                        style="border-color: var(--line)"
                        @click="lightboxUrl = t.photo_url; photoLightbox = true"
                      >
                        <img :src="t.photo_url" alt="Bukti Transaksi" class="w-full h-full object-cover" />
                        <div class="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 flex items-center justify-center text-white transition-opacity">
                          <i class="pi pi-search-plus text-xs" />
                        </div>
                      </div>

                      <div class="grid sm:grid-cols-3 gap-3 text-[12.5px] flex-1">
                        <div>
                          <div class="t-label mb-1">Petugas / Penerima</div>
                          <div class="font-medium">{{ t.received_by || '—' }}</div>
                        </div>
                        <div>
                          <div class="t-label mb-1">Catatan Transaksi</div>
                          <div class="text-ink-200">{{ t.notes || '—' }}</div>
                        </div>
                        <div>
                          <div class="t-label mb-1">ID Transaksi</div>
                          <div class="t-mono text-[11px] text-ink-400">{{ t.id }}</div>
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
    </div>

    <!-- ======================================================== -->
    <!-- TAB 2: AUDIT LOG & KEAMANAN SISTEM -->
    <!-- ======================================================== -->
    <div v-else class="flex flex-col gap-4">
      <!-- Audit KPI Summary Strip -->
      <div
        class="grid grid-cols-2 sm:grid-cols-4 gap-px rounded-xl overflow-hidden border shadow-xs"
        style="border-color: var(--line); background: var(--line)"
      >
        <div
          v-for="c in [
            { k: 'TOTAL LOG', v: auditCounts.total, cls: 'text-ink-100' },
            { k: 'INSERT (TAMBAH)', v: auditCounts.ins, cls: 'text-sig-ok' },
            { k: 'UPDATE (UBAH)', v: auditCounts.upd, cls: 'text-sig-info' },
            { k: 'DELETE (HAPUS)', v: auditCounts.del, cls: 'text-sig-bad' },
          ]"
          :key="c.k"
          class="px-4 py-3"
          style="background: var(--panel)"
        >
          <div class="t-label">{{ c.k }}</div>
          <div class="t-num text-[20px] font-bold mt-1" :class="c.cls">{{ c.v }}</div>
        </div>
      </div>

      <!-- Audit Table Panel -->
      <Panel title="Jejak Mutasi Database &amp; Keamanan" icon="pi pi-shield" dense>
        <template #actions>
          <InputText
            v-model="auditQuery"
            placeholder="Cari tabel / row id…"
            class="!text-[12px] !py-1.5 w-[180px]"
          />
          <Select
            v-model="opFilter"
            :options="opOptions"
            optionLabel="label"
            optionValue="value"
            class="!text-[12px] w-[150px]"
          />
        </template>

        <EmptyState v-if="loadingAudit" icon="pi pi-spin pi-spinner" title="Memuat audit log…" />
        <EmptyState
          v-else-if="!filteredAudit.length"
          icon="pi pi-shield"
          title="Tidak ada catatan audit"
          sub="Jejak audit terekam otomatis saat ada modifikasi data barang, pengguna, atau transaksi."
        />

        <div v-else class="overflow-x-auto">
          <table class="w-full text-[12.5px]">
            <thead>
              <tr class="text-left" style="background: var(--paper-2)">
                <th class="t-label px-4 py-2.5 w-[180px]">Waktu</th>
                <th class="t-label px-4 py-2.5 w-[170px]">Tabel Target</th>
                <th class="t-label px-4 py-2.5 w-[110px]">Aksi</th>
                <th class="t-label px-4 py-2.5">Row ID / Kunci</th>
                <th class="t-label px-4 py-2.5 w-[80px] text-right">Detail</th>
              </tr>
            </thead>
            <tbody class="divide-y" style="border-color: var(--line-soft)">
              <template v-for="(l, i) in filteredAudit" :key="i">
                <tr
                  class="hover:bg-paper-2 transition-colors cursor-pointer"
                  @click="expandedAudit = expandedAudit === l.row_id + i ? null : l.row_id + i"
                >
                  <td class="px-4 py-2.5 whitespace-nowrap" style="color: var(--txt-dim)">{{ l.at }}</td>
                  <td class="px-4 py-2.5">
                    <span class="t-mono font-semibold text-acc-400">{{ l.table_name }}</span>
                  </td>
                  <td class="px-4 py-2.5"><StatusChip :kind="l.op" /></td>
                  <td class="px-4 py-2.5">
                    <span class="t-mono text-[11px]" style="color: var(--txt-dim)">{{ l.row_id || '—' }}</span>
                  </td>
                  <td class="px-4 py-2.5 text-right">
                    <i
                      class="pi text-[10px]"
                      :class="expandedAudit === l.row_id + i ? 'pi-chevron-up' : 'pi-chevron-down'"
                      style="color: var(--txt-dim)"
                    />
                  </td>
                </tr>

                <!-- JSON Diff Expanded Row -->
                <tr v-if="expandedAudit === l.row_id + i">
                  <td colspan="5" class="px-4 pb-3.5 pt-0" style="background: var(--paper-1)">
                    <div class="grid md:grid-cols-2 gap-3 pt-2">
                      <div class="flex flex-col gap-1">
                        <div class="t-label text-rose-400">Data Sebelumnya (Old Data)</div>
                        <pre class="panel !rounded-md p-3 t-mono text-[11px] overflow-x-auto max-h-[220px] bg-black/60 border border-line">{{ prettyJSON(l.old_data) }}</pre>
                      </div>
                      <div class="flex flex-col gap-1">
                        <div class="t-label text-emerald-400">Data Baru (New Data)</div>
                        <pre class="panel !rounded-md p-3 t-mono text-[11px] overflow-x-auto max-h-[220px] bg-black/60 border border-line">{{ prettyJSON(l.new_data) }}</pre>
                      </div>
                    </div>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>

        <template #footer>
          <div class="flex items-center justify-between text-[11.5px]" style="color: var(--txt-dim)">
            <span>Menampilkan {{ filteredAudit.length }} dari {{ auditLogs.length }} catatan audit</span>
            <span>Klik baris untuk melihat perbedaan payload JSON</span>
          </div>
        </template>
      </Panel>
    </div>

    <!-- Modal Lightbox Foto Geotag -->
    <Dialog v-model:visible="photoLightbox" modal header="Foto Bukti Geotag Transaksi" :style="{ width: '640px' }" class="p-fluid">
      <div v-if="lightboxUrl" class="p-2 bg-black rounded-lg flex items-center justify-center">
        <img :src="lightboxUrl" alt="Foto Geotag Transaksi" class="max-h-[75vh] object-contain rounded" />
      </div>
    </Dialog>
  </div>
</template>
