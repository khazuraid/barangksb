<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue'
import api from '@/api'
import { useRouter } from 'vue-router'
import { useConfirm } from 'primevue/useconfirm'
import { useToast } from 'primevue/usetoast'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import PageHeader from '@/components/PageHeader.vue'
import Panel from '@/components/Panel.vue'
import StatusChip from '@/components/StatusChip.vue'
import EmptyState from '@/components/EmptyState.vue'
import Tag from 'primevue/tag'

const router = useRouter()
const confirm = useConfirm()
const toast = useToast()

const items = ref<any[]>([])
const total = ref(0)
const page = ref(0)
const perPage = ref(25)
const q = ref('')
const category = ref('')
const location = ref('')
const onlyLow = ref(false)
const loading = ref(true)
const expanded = ref<string | null>(null)

const categories = ref<any[]>([])
const locations = ref<any[]>([])
const perPageOptions = [25, 50, 100]

async function fetchItems() {
  loading.value = true
  try {
    const res = await api.get('/items', {
      params: { q: q.value, cat: category.value, loc: location.value, page: page.value + 1, per_page: perPage.value },
    })
    items.value = res.data.data
    total.value = res.data.total
  } finally { loading.value = false }
}
async function fetchFilters() {
  const [c, l] = await Promise.all([api.get('/categories'), api.get('/locations')])
  categories.value = c.data
  locations.value = l.data
}
onMounted(() => { fetchItems(); fetchFilters() })
watch([page, perPage], fetchItems)

const shown = computed(() => onlyLow.value ? items.value.filter(i => i.current_stock <= i.min_stock) : items.value)
const lastPage = computed(() => Math.max(0, Math.ceil(total.value / perPage.value) - 1))
const range = computed(() => {
  if (!total.value) return '0 barang'
  return `${page.value * perPage.value + 1}–${Math.min(total.value, (page.value + 1) * perPage.value)} dari ${total.value}`
})
const lowCount = computed(() => shown.value.filter(i => i.current_stock <= i.min_stock).length)

function resetFilters() {
  q.value = ''; category.value = ''; location.value = ''; onlyLow.value = false
  page.value = 0; fetchItems()
}

function remove(item: any) {
  confirm.require({
    message: `Hapus barang "${item.name}" (${item.sku})? Seluruh riwayat transaksi barang ini ikut terhapus.`,
    header: 'Konfirmasi hapus barang',
    icon: 'pi pi-exclamation-triangle',
    acceptClass: 'p-button-danger',
    accept: async () => {
      await api.delete(`/items/${item.id}`)
      toast.add({ severity: 'success', summary: 'Barang dihapus', detail: item.name, life: 2500 })
      fetchItems()
    },
  })
}

function exportCsv() { window.open('/api/export/items.csv', '_blank') }
function exportXlsx() { window.open('/api/export/items.xlsx', '_blank') }
function printReport() { window.open('/api/report.pdf', '_blank') }
</script>

<template>
  <div>
    <PageHeader crumb="Operasional" title="Daftar Barang"
      sub="Master inventaris kantor — setiap baris merepresentasikan satu SKU">
      <template #actions>
        <Button label="Tambah Barang" icon="pi pi-plus" size="small" @click="router.push('/items/new')" />
      </template>
    </PageHeader>

    <div class="panel p-3.5 mb-4">
      <div class="flex flex-wrap gap-3 items-end">
        <label class="flex flex-col gap-1.5 flex-1 min-w-[200px]">
          <span class="text-[11.5px] font-semibold" style="color: var(--txt-dim)">Cari</span>
          <InputText v-model="q" placeholder="Nama, SKU, atau lokasi…" class="w-full !text-[12.5px]"
                     @keyup.enter="page = 0; fetchItems()" />
        </label>
        <label class="flex flex-col gap-1.5 min-w-[180px]">
          <span class="text-[11.5px] font-semibold" style="color: var(--txt-dim)">Kategori</span>
          <Select v-model="category" :options="categories" optionLabel="name" optionValue="name" showClear
                  placeholder="Semua" class="!text-[12.5px]" @change="page = 0; fetchItems()" filter />
        </label>
        <label class="flex flex-col gap-1.5 min-w-[170px]">
          <span class="text-[11.5px] font-semibold" style="color: var(--txt-dim)">Lokasi</span>
          <Select v-model="location" :options="locations" optionLabel="name" optionValue="name" showClear
                  placeholder="Semua" class="!text-[12.5px]" @change="page = 0; fetchItems()" filter />
        </label>
        <div class="flex items-center gap-2 pb-1.5">
          <Button :label="onlyLow ? 'Hanya menipis (aktif)' : 'Hanya stok menipis'"
                  icon="pi pi-exclamation-triangle" text size="small"
                  :severity="onlyLow ? 'danger' : 'secondary'"
                  @click="onlyLow = !onlyLow" />
          <Button icon="pi pi-search" size="small" @click="page = 0; fetchItems()" />
          <Button icon="pi pi-filter-slash" size="small" text severity="secondary"
                  v-tooltip.top="'Reset filter'" @click="resetFilters" />
        </div>
      </div>
    </div>

    <Panel title="Data Barang" icon="pi pi-box" dense>
      <template #actions>
        <Tag severity="secondary" :value="range" />
        <Tag v-if="lowCount" severity="danger" :value="lowCount + ' menipis'" />
        <Button icon="pi pi-file" size="small" text severity="secondary" v-tooltip.top="'Ekspor CSV'"
                @click="exportCsv" />
        <Button icon="pi pi-file-excel" size="small" text severity="secondary" v-tooltip.top="'Ekspor Excel'"
                @click="exportXlsx" />
        <Button icon="pi pi-file-pdf" size="small" text severity="secondary" v-tooltip.top="'Laporan PDF'"
                @click="printReport" />
      </template>

      <EmptyState v-if="loading" icon="pi pi-spin pi-spinner" title="Memuat data barang…" />
      <EmptyState v-else-if="!shown.length" icon="pi pi-box" title="Tidak ada barang"
                  sub="Ubah kata kunci atau tambahkan barang baru ke inventaris.">
        <Button label="Tambah barang" icon="pi pi-plus" size="small" @click="router.push('/items/new')" />
      </EmptyState>

      <div v-else class="overflow-x-auto">
        <table class="w-full text-[12.5px]">
          <thead>
            <tr class="text-left" style="background: var(--paper-2)">
              <th class="t-label px-4 py-2.5 w-[140px]">SKU</th>
              <th class="t-label px-4 py-2.5">Nama Barang</th>
              <th class="t-label px-4 py-2.5 w-[160px]">Kategori</th>
              <th class="t-label px-4 py-2.5 w-[140px]">Lokasi</th>
              <th class="t-label px-4 py-2.5 w-[130px] text-right">Stok</th>
              <th class="t-label px-4 py-2.5 w-[130px]">Kondisi</th>
              <th class="t-label px-4 py-2.5 w-[100px] text-right">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y" style="border-color: var(--line-soft)">
            <template v-for="it in shown" :key="it.id">
              <tr class="hover:bg-paper-2 transition-colors cursor-pointer"
                  @click="expanded = expanded === it.id ? null : it.id">
                <td class="px-4 py-2.5"><span class="t-mono font-semibold">{{ it.sku }}</span></td>
                <td class="px-4 py-2.5">
                  <div class="flex items-center gap-2.5">
                    <span class="w-6 h-6 shrink-0 grid place-items-center rounded border"
                          style="border-color: var(--line); background: var(--paper-2)">
                      <i class="pi pi-box text-[10px]" style="color: var(--txt-dim)" />
                    </span>
                    <span class="font-semibold">{{ it.name }}</span>
                  </div>
                </td>
                <td class="px-4 py-2.5" style="color: var(--txt-dim)">{{ it.category }}</td>
                <td class="px-4 py-2.5" style="color: var(--txt-dim)">{{ it.location }}</td>
                <td class="px-4 py-2.5 text-right">
                  <span class="t-num font-bold" :class="it.current_stock <= it.min_stock ? 'text-sig-bad' : ''">
                    {{ it.current_stock }}
                  </span>
                  <span class="text-[11px] ml-1" style="color: var(--txt-dim)">{{ it.unit }}</span>
                  <div v-if="it.current_stock <= it.min_stock" class="t-label !text-[9px] text-sig-bad">min {{ it.min_stock }}</div>
                </td>
                <td class="px-4 py-2.5"><StatusChip :kind="it.condition_status" /></td>
                <td class="px-4 py-2.5 text-right" @click.stop>
                  <Button icon="pi pi-pencil" text rounded size="small" severity="secondary"
                          v-tooltip.top="'Edit'" @click="router.push(`/items/${it.id}/edit`)" />
                  <Button icon="pi pi-trash" text rounded size="small" severity="danger"
                          v-tooltip.top="'Hapus'" @click="remove(it)" />
                </td>
              </tr>
              <tr v-if="expanded === it.id">
                <td colspan="7" class="px-4 pb-3.5 pt-0" style="background: var(--paper-1)">
                  <div class="grid sm:grid-cols-4 gap-3 text-[12px]">
                    <div>
                      <div class="t-label mb-1">Harga Satuan</div>
                      <div class="t-num">{{ it.price_per_unit ? 'Rp ' + Number(it.price_per_unit).toLocaleString('id-ID') : '—' }}</div>
                    </div>
                    <div>
                      <div class="t-label mb-1">Ketersediaan</div>
                      <div>{{ it.is_available ? 'Tersedia' : 'Tidak tersedia' }}</div>
                    </div>
                    <div>
                      <div class="t-label mb-1">ID Barang</div>
                      <div class="t-mono">{{ it.id }}</div>
                    </div>
                    <div class="flex items-end gap-2">
                      <Button label="Edit" icon="pi pi-pencil" size="small" text
                              @click="router.push(`/items/${it.id}/edit`)" />
                      <Button label="QR" icon="pi pi-qrcode" size="small" text
                              @click="router.push('/barcode')" />
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
            <Button icon="pi pi-angle-double-left" size="small" text severity="secondary"
                    :disabled="page === 0" @click="page = 0" />
            <Button icon="pi pi-angle-left" size="small" text severity="secondary"
                    :disabled="page === 0" @click="page--" />
            <span class="t-num text-[12px] px-1">Hal. {{ page + 1 }} / {{ lastPage + 1 }}</span>
            <Button icon="pi pi-angle-right" size="small" text severity="secondary"
                    :disabled="page >= lastPage" @click="page++" />
            <Button icon="pi pi-angle-double-right" size="small" text severity="secondary"
                    :disabled="page >= lastPage" @click="page = lastPage" />
          </div>
        </div>
      </template>
    </Panel>
  </div>
</template>