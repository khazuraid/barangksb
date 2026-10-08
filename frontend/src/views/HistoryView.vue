<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue'
import api from '@/api'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import PageHeader from '@/components/PageHeader.vue'
import Panel from '@/components/Panel.vue'
import StatusChip from '@/components/StatusChip.vue'
import EmptyState from '@/components/EmptyState.vue'

const txs = ref<any[]>([])
const total = ref(0)
const page = ref(0)
const perPage = ref(25)
const sku = ref('')
const type = ref('')
const loading = ref(false)
const expanded = ref<string | null>(null)

const typeOptions = [
  { label: 'Semua jenis', value: '' },
  { label: 'Barang Masuk', value: 'IN' },
  { label: 'Barang Keluar', value: 'OUT' },
  { label: 'Opname (+)', value: 'ADJUST+' },
  { label: 'Opname (−)', value: 'ADJUST-' },
]
const perPageOptions = [25, 50, 100]

async function fetch_(reset = false) {
  if (reset) page.value = 0
  loading.value = true
  try {
    const res = await api.get('/transactions', {
      params: { sku: sku.value, type: type.value, page: page.value + 1, per_page: perPage.value },
    })
    txs.value = res.data.data
    total.value = res.data.total
  } finally { loading.value = false }
}
onMounted(() => fetch_())
watch([page, perPage], () => fetch_())

const lastPage = computed(() => Math.max(0, Math.ceil(total.value / perPage.value) - 1))
const range = computed(() => {
  if (!total.value) return '0'
  const from = page.value * perPage.value + 1
  const to = Math.min(total.value, (page.value + 1) * perPage.value)
  return `${from}–${to} dari ${total.value}`
})

const summary = computed(() => ({
  in: txs.value.filter(t => t.type === 'IN').reduce((a, t) => a + t.quantity, 0),
  out: txs.value.filter(t => t.type === 'OUT').reduce((a, t) => a + t.quantity, 0),
  adj: txs.value.filter(t => t.type.startsWith('ADJUST')).length,
}))

function exportAs(kind: 'csv' | 'xlsx' | 'pdf') {
  const url = kind === 'csv' ? '/api/export/tx.csv' : kind === 'xlsx' ? '/api/export/items.xlsx' : '/api/report.pdf'
  window.open(url, '_blank')
}
function exportFiltered() {
  const header = ['Waktu', 'Jenis', 'SKU', 'Nama', 'Qty', 'Satuan', 'Sebelum', 'Sesudah', 'Petugas', 'Catatan']
  const rows = txs.value.map(t => [t.timestamp, t.type, t.item_sku, t.item_name, t.quantity, t.unit,
                                   t.previous_stock, t.new_stock, t.received_by || '', (t.notes || '').replace(/\n/g, ' ')])
  const csv = [header, ...rows]
    .map(r => r.map(v => `"${String(v ?? '').replace(/"/g, '""')}"`).join(','))
    .join('\n')
  const blob = new Blob(['\uFEFF' + csv], { type: 'text/csv;charset=utf-8' })
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = `riwayat_mutasi_halaman_${page.value + 1}.csv`
  a.click()
  URL.revokeObjectURL(a.href)
}
</script>

<template>
  <div>
    <PageHeader crumb="Operasional" title="Riwayat Mutasi"
      sub="Jejak lengkap setiap perubahan stok, terbaru di atas">
      <template #actions>
        <Button label="Halaman ini (CSV)" icon="pi pi-file" size="small" severity="secondary" outlined
                :disabled="!txs.length" @click="exportFiltered" />
        <Button label="Ekspor" icon="pi pi-download" size="small"
                :disabled="!txs.length" @click="exportAs('csv')" />
      </template>
    </PageHeader>

    <!-- quick strip -->
    <div class="grid grid-cols-3 gap-px mb-4 rounded-lg overflow-hidden border max-w-xl"
         style="border-color: var(--line); background: var(--line)">
      <div class="px-4 py-3" style="background: var(--panel)">
        <div class="t-label">Unit Masuk (hal. ini)</div>
        <div class="t-num text-[20px] font-bold mt-1 text-sig-ok">+{{ summary.in }}</div>
      </div>
      <div class="px-4 py-3" style="background: var(--panel)">
        <div class="t-label">Unit Keluar (hal. ini)</div>
        <div class="t-num text-[20px] font-bold mt-1 text-sig-bad">−{{ summary.out }}</div>
      </div>
      <div class="px-4 py-3" style="background: var(--panel)">
        <div class="t-label">Baris Opname</div>
        <div class="t-num text-[20px] font-bold mt-1 text-acc-600">{{ summary.adj }}</div>
      </div>
    </div>

    <Panel title="Transaksi" icon="pi pi-history" dense>
      <template #actions>
        <InputText v-model="sku" placeholder="Cari SKU…" class="!text-[12px] !py-1.5 w-[150px]"
                   @keyup.enter="fetch_(true)" />
        <Select v-model="type" :options="typeOptions" optionLabel="label" optionValue="value"
                class="!text-[12px] w-[155px]" @change="fetch_(true)" />
        <Button icon="pi pi-search" size="small" severity="secondary" outlined @click="fetch_(true)" />
      </template>

      <EmptyState v-if="loading" icon="pi pi-spin pi-spinner" title="Memuat riwayat…" />
      <EmptyState v-else-if="!txs.length" icon="pi pi-history" title="Belum ada transaksi"
                  sub="Riwayat transaksi akan tampil di sini saat terjadi perubahan stok.">
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
              <tr class="hover:bg-paper-2 transition-colors cursor-pointer"
                  @click="expanded = expanded === t.id ? null : t.id">
                <td class="px-4 py-2.5 whitespace-nowrap" style="color: var(--txt-dim)">{{ t.timestamp }}</td>
                <td class="px-4 py-2.5"><StatusChip :kind="t.type" /></td>
                <td class="px-4 py-2.5">
                  <div class="flex items-center gap-1.5 font-semibold">
                    <span>{{ t.item_name }}</span>
                    <a v-if="t.photo_url" :href="t.photo_url" target="_blank" @click.stop
                       title="Foto bukti berstempel geotag" class="inline-flex text-acc-500 hover:text-acc-400">
                      <i class="pi pi-camera text-[11px]" />
                    </a>
                  </div>
                  <div class="t-mono" style="color: var(--txt-dim)">{{ t.item_sku }}</div>
                </td>
                <td class="px-4 py-2.5 text-right">
                  <span class="t-num font-bold"
                        :class="t.type === 'IN' || t.type === 'ADJUST+' ? 'text-sig-ok' : 'text-sig-bad'">
                    {{ t.type === 'IN' || t.type === 'ADJUST+' ? '+' : '−' }}{{ t.quantity }} {{ t.unit }}
                  </span>
                </td>
                <td class="px-4 py-2.5 text-right t-num whitespace-nowrap" style="color: var(--txt-dim)">
                  {{ t.previous_stock }} <i class="pi pi-arrow-right text-[9px] mx-0.5" /> <span class="font-semibold" style="color: var(--txt)">{{ t.new_stock }}</span>
                </td>
                <td class="px-4 py-2.5 text-right">
                  <i class="pi text-[10px]" :class="expanded === t.id ? 'pi-chevron-up' : 'pi-chevron-down'"
                     style="color: var(--txt-dim)" />
                </td>
              </tr>
              <tr v-if="expanded === t.id">
                <td colspan="6" class="px-4 pb-3.5 pt-0" style="background: var(--paper-1)">
                  <div class="flex flex-wrap sm:flex-nowrap gap-4 items-center pt-2">
                    <div v-if="t.photo_url" class="relative group w-24 h-20 shrink-0 rounded overflow-hidden border bg-black/10"
                         style="border-color: var(--line)">
                      <img :src="t.photo_url" alt="Bukti Transaksi" class="w-full h-full object-cover" />
                      <a :href="t.photo_url" target="_blank"
                         class="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 flex items-center justify-center text-white transition-opacity"
                         title="Lihat foto bukti geotag asli">
                        <i class="pi pi-external-link text-xs" />
                      </a>
                    </div>
                    <div class="grid sm:grid-cols-3 gap-3 text-[12.5px] flex-1">
                      <div>
                        <div class="t-label mb-1">Petugas / Penerima</div>
                        <div>{{ t.received_by || '—' }}</div>
                      </div>
                      <div>
                        <div class="t-label mb-1">Catatan</div>
                        <div>{{ t.notes || '—' }}</div>
                      </div>
                      <div>
                        <div class="t-label mb-1">ID Transaksi</div>
                        <div class="t-mono">{{ t.id }}</div>
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
            <Button icon="pi pi-angle-left" size="small" text severity="secondary"
                    :disabled="page === 0" @click="page--" />
            <span class="t-num text-[12px] px-1">Hal. {{ page + 1 }} / {{ lastPage + 1 }}</span>
            <Button icon="pi pi-angle-right" size="small" text severity="secondary"
                    :disabled="page >= lastPage" @click="page++" />
          </div>
        </div>
      </template>
    </Panel>
  </div>
</template>
