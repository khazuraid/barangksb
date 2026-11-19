<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue'
import api, { downloadAuthenticatedFile, safeHref } from '@/api'
import { useRouter } from 'vue-router'
import { useConfirm } from 'primevue/useconfirm'
import { useToast } from 'primevue/usetoast'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import InputNumber from 'primevue/inputnumber'
import Textarea from 'primevue/textarea'
import Select from 'primevue/select'
import Dialog from 'primevue/dialog'
import PageHeader from '@/components/PageHeader.vue'
import Panel from '@/components/Panel.vue'
import StatusChip from '@/components/StatusChip.vue'
import EmptyState from '@/components/EmptyState.vue'
import PhotoUploader from '@/components/PhotoUploader.vue'

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
const sortBy = ref('newest')
const loading = ref(true)
const viewMode = ref<'table' | 'grid'>('table')

const categories = ref<any[]>([])
const locations = ref<any[]>([])

const perPageOptions = [
  { label: '20 per hal', value: 20 },
  { label: '25 per hal', value: 25 },
  { label: '50 per hal', value: 50 },
  { label: '100 per hal', value: 100 },
]

const sortOptions = [
  { label: 'Terkini / Terbaru', value: 'newest' },
  { label: 'Nama (A–Z)', value: 'name_asc' },
  { label: 'Nama (Z–A)', value: 'name_desc' },
  { label: 'Stok Terbanyak', value: 'stock_desc' },
  { label: 'Stok Terdikit', value: 'stock_asc' },
  { label: 'Paling Lama', value: 'oldest' },
]

// --- Import Excel / CSV state ---
const importVisible = ref(false)
const importing = ref(false)
const importFileInput = ref<HTMLInputElement | null>(null)
const selectedImportFile = ref<File | null>(null)
const importResult = ref<{ total: number; count: number; skipped?: number; failed: number; errors: string[] } | null>(null)

// --- Maintenance & Calibration state ---
const maintenanceVisible = ref(false)
const selectedItemForMaint = ref<any>(null)
const maintenanceList = ref<any[]>([])
const loadingMaint = ref(false)
const showAddMaintForm = ref(false)
const submittingMaint = ref(false)
const maintPhotoRef = ref<any>(null)
const serviceTypeOptions = ['Servis Rutin', 'Kalibrasi', 'Perbaikan', 'Pemeriksaan Berkala', 'Penggantian Suku Cadang']
const conditionOptions = ['Berfungsi', 'Rusak Ringan', 'Rusak Berat', 'Perlu Kalibrasi']
const maintForm = ref({
  service_type: 'Servis Rutin',
  service_date: new Date().toISOString().split('T')[0],
  vendor_or_technician: '',
  cost: 0,
  next_service_date: '',
  description: '',
  photo_url: '',
  update_condition: '',
})

async function fetchItems() {
  loading.value = true
  try {
    const res = await api.get('/items', {
      params: {
        q: q.value,
        cat: category.value,
        loc: location.value,
        sort: sortBy.value,
        low_stock: onlyLow.value ? 'true' : undefined,
        page: page.value + 1,
        per_page: perPage.value,
      },
    })
    items.value = res.data?.data || []
    total.value = res.data?.total || 0
  } finally {
    loading.value = false
  }
}

async function fetchFilters() {
  const [c, l] = await Promise.all([api.get('/categories'), api.get('/locations')])
  categories.value = c.data || []
  locations.value = l.data || []
}

onMounted(() => {
  fetchItems()
  fetchFilters()
})

let searchDebounce: any = null
watch(q, () => {
  clearTimeout(searchDebounce)
  searchDebounce = setTimeout(() => {
    page.value = 0
    fetchItems()
  }, 350)
})

watch([category, location, sortBy, onlyLow], () => {
  page.value = 0
  fetchItems()
})

watch(perPage, () => {
  page.value = 0
  fetchItems()
})

watch(page, () => {
  fetchItems()
})

const shown = computed(() => items.value)
const lastPage = computed(() => Math.max(0, Math.ceil(total.value / perPage.value) - 1))
const range = computed(() => {
  if (!total.value) return '0 barang'
  const from = page.value * perPage.value + 1
  const to = Math.min(total.value, (page.value + 1) * perPage.value)
  return `${from}–${to} dari ${total.value} barang`
})
const lowCount = computed(() => items.value.filter(i => i.track_stock !== false && i.current_stock <= i.min_stock).length)

function resetFilters() {
  q.value = ''
  category.value = ''
  location.value = ''
  onlyLow.value = false
  sortBy.value = 'newest'
  page.value = 0
  fetchItems()
}

function remove(item: any) {
  if (!item?.id || item.id === 'undefined') {
    toast.add({ severity: 'error', summary: 'ID barang tidak valid', life: 3000 })
    return
  }
  confirm.require({
    message: `Hapus barang "${item.name}" (${item.sku})? Seluruh riwayat transaksi barang ini ikut terhapus.`,
    header: 'Konfirmasi hapus barang',
    icon: 'pi pi-exclamation-triangle',
    acceptClass: 'p-button-danger',
    accept: async () => {
      try {
        await api.delete(`/items/${item.id}`)
        toast.add({ severity: 'success', summary: 'Barang dihapus', detail: item.name, life: 2500 })
        fetchItems()
      } catch (err: any) {
        toast.add({ severity: 'error', summary: 'Gagal menghapus barang', detail: err.response?.data?.error || err.message, life: 3500 })
      }
    },
  })
}

async function exportCsv() {
  try {
    await downloadAuthenticatedFile('/export/items.csv', 'master_barang.csv')
  } catch (err: any) {
    toast.add({ severity: 'error', summary: 'Gagal ekspor CSV', detail: err.response?.data?.error || err.message, life: 3500 })
  }
}
async function exportXlsx() {
  try {
    await downloadAuthenticatedFile('/export/items.xlsx', 'master_barang.xlsx')
  } catch (err: any) {
    toast.add({ severity: 'error', summary: 'Gagal ekspor Excel', detail: err.response?.data?.error || err.message, life: 3500 })
  }
}
async function printReport() {
  try {
    await downloadAuthenticatedFile('/report.pdf', 'laporan_inventaris.pdf', true)
  } catch (err: any) {
    toast.add({ severity: 'error', summary: 'Gagal membuka laporan PDF', detail: err.response?.data?.error || err.message, life: 3500 })
  }
}

async function downloadTemplate(fmt: 'xlsx' | 'csv') {
  try {
    await downloadAuthenticatedFile(`/items/template?fmt=${fmt}`, `template_import_barang.${fmt}`)
  } catch (err: any) {
    toast.add({ severity: 'error', summary: 'Gagal mengunduh template', detail: err.response?.data?.error || err.message, life: 3500 })
  }
}

function onImportFileChanged(e: Event) {
  const target = e.target as HTMLInputElement
  if (target.files?.[0]) {
    selectedImportFile.value = target.files[0]
  }
}

async function doImport() {
  if (!selectedImportFile.value) {
    toast.add({ severity: 'warn', summary: 'Pilih berkas Excel/CSV terlebih dahulu', life: 2500 })
    return
  }
  importing.value = true
  importResult.value = null
  try {
    const fd = new FormData()
    fd.append('file', selectedImportFile.value)
    const res = await api.post('/items/import', fd)
    importResult.value = res.data
    if (res.data.count > 0) {
      toast.add({ severity: 'success', summary: 'Import berhasil', detail: `${res.data.count} barang ditambahkan`, life: 3500 })
      fetchItems()
      fetchFilters()
    } else {
      toast.add({ severity: 'warn', summary: 'Tidak ada data terimport', detail: 'Periksa baris error di bawah', life: 3500 })
    }
  } catch (err: any) {
    toast.add({ severity: 'error', summary: 'Gagal mengimport file', detail: err.response?.data?.error || err.message, life: 4000 })
  } finally {
    importing.value = false
  }
}

async function openMaintenance(it: any) {
  selectedItemForMaint.value = it
  maintenanceVisible.value = true
  showAddMaintForm.value = false
  maintForm.value = {
    service_type: 'Servis Rutin',
    service_date: new Date().toISOString().split('T')[0],
    vendor_or_technician: '',
    cost: 0,
    next_service_date: '',
    description: '',
    photo_url: '',
    update_condition: it.condition_status || 'Berfungsi',
  }
  await fetchMaintenance(it.id)
}

async function fetchMaintenance(id: string) {
  loadingMaint.value = true
  try {
    const res = await api.get(`/items/${id}/maintenance`)
    maintenanceList.value = res.data
  } catch {
    toast.add({ severity: 'error', summary: 'Gagal memuat riwayat servis', life: 3000 })
  } finally {
    loadingMaint.value = false
  }
}

async function submitMaintenance() {
  if (!selectedItemForMaint.value) return
  submittingMaint.value = true
  try {
    if (maintPhotoRef.value?.hasPendingPhoto) {
      const u = await maintPhotoRef.value.uploadPending()
      if (u) maintForm.value.photo_url = u
    }
    await api.post(`/items/${selectedItemForMaint.value.id}/maintenance`, maintForm.value)
    toast.add({ severity: 'success', summary: 'Catatan servis tersimpan', life: 3000 })
    showAddMaintForm.value = false
    await fetchMaintenance(selectedItemForMaint.value.id)
    fetchItems()
  } catch (err: any) {
    toast.add({ severity: 'error', summary: 'Gagal menyimpan', detail: err.response?.data?.error || err.message, life: 3500 })
  } finally {
    submittingMaint.value = false
  }
}

function deleteMaintRecord(mId: string) {
  if (!selectedItemForMaint.value) return
  confirm.require({
    message: 'Hapus catatan servis ini dari riwayat pemeliharaan?',
    header: 'Konfirmasi hapus catatan',
    icon: 'pi pi-exclamation-triangle',
    acceptClass: 'p-button-danger',
    accept: async () => {
      try {
        await api.delete(`/items/${selectedItemForMaint.value.id}/maintenance/${mId}`)
        toast.add({ severity: 'info', summary: 'Catatan servis dihapus', life: 2500 })
        fetchMaintenance(selectedItemForMaint.value.id)
      } catch (err: any) {
        toast.add({ severity: 'error', summary: 'Gagal menghapus', detail: err.response?.data?.error || err.message, life: 3000 })
      }
    },
  })
}
</script>

<template>
  <div>
    <PageHeader
      crumb="Operasional"
      title="Daftar Barang"
      sub="Master inventaris puskesmas — setiap baris merepresentasikan satu SKU atau aset tetap."
    >
      <template #actions>
        <Button
          label="Import Excel / CSV"
          icon="pi pi-file-import"
          size="small"
          severity="secondary"
          outlined
          @click="importVisible = true; importResult = null; selectedImportFile = null"
        />
        <Button label="Tambah Barang" icon="pi pi-plus" size="small" @click="router.push('/items/new')" />
      </template>
    </PageHeader>

    <!-- Dense Filter Toolbar -->
    <div class="panel p-3 mb-4 flex flex-wrap items-center justify-between gap-2.5">
      <div class="flex flex-wrap items-center gap-2 flex-1 min-w-[280px]">
        <div class="relative flex-1 min-w-[200px] max-w-sm">
          <i class="pi pi-search absolute left-3 top-1/2 -translate-y-1/2 text-txt-dim text-[12px]" aria-hidden="true" />
          <InputText
            v-model="q"
            placeholder="Cari nama, SKU, kode barang, merk, lokasi…"
            class="w-full !pl-8 !pr-7 !text-[12.5px]"
          />
          <button
            v-if="q"
            type="button"
            aria-label="Bersihkan pencarian"
            class="absolute right-2 top-1/2 -translate-y-1/2 text-txt-dim hover:text-txt text-xs p-1 cursor-pointer"
            @click="q = ''"
          >
            <i class="pi pi-times" aria-hidden="true" />
          </button>
        </div>

        <Select
          v-model="category"
          :options="categories"
          optionLabel="name"
          optionValue="name"
          showClear
          placeholder="Kategori"
          class="!text-[12.5px] w-[150px]"
          filter
        />

        <Select
          v-model="location"
          :options="locations"
          optionLabel="name"
          optionValue="name"
          showClear
          placeholder="Ruangan"
          class="!text-[12.5px] w-[145px]"
          filter
        />

        <Select
          v-model="sortBy"
          :options="sortOptions"
          optionLabel="label"
          optionValue="value"
          placeholder="Urutan"
          class="!text-[12.5px] w-[160px]"
        />

        <!-- Outline Filter Chip -->
        <button
          type="button"
          class="h-[34px] px-2.5 rounded-[7px] text-[12px] font-semibold flex items-center gap-1.5 transition-colors cursor-pointer border"
          :class="onlyLow
            ? 'bg-[var(--sig-bad-bg)] text-sig-bad border-[var(--sig-bad-line)]'
            : 'bg-panel-2 text-txt-dim border-line hover:border-[var(--acc-border)] hover:text-txt'"
          @click="onlyLow = !onlyLow"
        >
          <i class="pi pi-exclamation-triangle text-[11px]" :class="onlyLow ? 'text-sig-bad' : 'text-sig-warn'" aria-hidden="true" />
          <span>Stok Menipis</span>
          <span v-if="lowCount" class="t-mono !text-[10.5px] font-bold px-1 rounded bg-[var(--sig-bad-bg)] text-sig-bad">
            {{ lowCount }}
          </span>
        </button>

        <Button
          v-if="q || category || location || onlyLow || sortBy !== 'newest'"
          icon="pi pi-filter-slash"
          text
          rounded
          size="small"
          severity="secondary"
          aria-label="Reset semua filter"
          v-tooltip.top="'Reset Semua Filter'"
          @click="resetFilters"
        />
      </div>

      <div class="flex items-center gap-1.5 shrink-0">
        <div class="flex items-center rounded-[7px] border border-line bg-panel-2 p-0.5">
          <Button
            icon="pi pi-list"
            size="small"
            :text="viewMode !== 'table'"
            :severity="viewMode === 'table' ? 'primary' : 'secondary'"
            class="!p-1.5 !w-7 !h-7"
            aria-label="Tampilan tabel"
            v-tooltip.top="'Tampilan Tabel'"
            @click="viewMode = 'table'"
          />
          <Button
            icon="pi pi-th-large"
            size="small"
            :text="viewMode !== 'grid'"
            :severity="viewMode === 'grid' ? 'primary' : 'secondary'"
            class="!p-1.5 !w-7 !h-7"
            aria-label="Tampilan kartu grid"
            v-tooltip.top="'Tampilan Kartu Grid'"
            @click="viewMode = 'grid'"
          />
        </div>
        <Button label="CSV" icon="pi pi-file" size="small" text severity="secondary" @click="exportCsv" />
        <Button label="Excel" icon="pi pi-file-excel" size="small" text severity="secondary" @click="exportXlsx" />
        <Button label="PDF" icon="pi pi-file-pdf" size="small" text severity="secondary" @click="printReport" />
      </div>
    </div>

    <!-- Inventory Panel -->
    <Panel title="Inventaris Aset" icon="pi pi-box" dense>
      <template #actions>
        <span class="t-mono !text-[11px] text-txt-dim">{{ range }}</span>
      </template>

      <EmptyState v-if="loading" loading title="Memuat data barang…" />
      <EmptyState
        v-else-if="!shown.length"
        icon="pi pi-box"
        title="Tidak ada barang"
        sub="Ubah kata kunci pencarian atau tambahkan barang baru."
      >
        <Button label="Tambah barang" icon="pi pi-plus" size="small" @click="router.push('/items/new')" />
      </EmptyState>

      <!-- Grid View -->
      <div v-else-if="viewMode === 'grid'" class="p-4 grid sm:grid-cols-2 md:grid-cols-3 xl:grid-cols-4 gap-3.5 bg-ink-950">
        <div
          v-for="it in shown"
          :key="it.id"
          class="panel p-3.5 flex flex-col justify-between hover:border-[var(--acc-border)] transition-colors cursor-pointer group"
          @click="router.push('/items/' + it.id)"
        >
          <div>
            <div class="flex items-start justify-between gap-1.5 mb-2.5 flex-wrap">
              <span class="t-mono !text-[10.5px] font-semibold px-2 py-0.5 rounded-[5px] border border-line bg-panel-2 text-txt-dim">
                {{ it.sku }}
              </span>
              <span
                v-if="it.item_code"
                class="t-mono !text-[10px] font-semibold px-1.5 py-0.5 rounded-[5px] border border-[var(--acc-border)] bg-[var(--acc-wash)] text-acc-500"
              >
                {{ it.item_code }}
              </span>
              <StatusChip :kind="it.condition_status" />
            </div>

            <div class="flex items-center gap-3 mb-3">
              <div class="w-11 h-11 rounded-[8px] border border-line bg-panel-2 overflow-hidden shrink-0 grid place-items-center">
                <img v-if="it.photo_url" :src="it.photo_url" :alt="it.name" class="w-full h-full object-cover" />
                <i v-else class="pi pi-box text-acc-500 text-[16px]" aria-hidden="true" />
              </div>
              <div class="min-w-0">
                <div class="font-semibold text-[13.5px] text-txt truncate group-hover:text-acc-500 transition-colors">
                  {{ it.name }}
                </div>
                <div class="text-[12px] text-txt-dim truncate">{{ it.category }}</div>
              </div>
            </div>
          </div>

          <div class="pt-2.5 border-t border-line flex items-center justify-between text-[12px]">
            <span class="truncate flex items-center gap-1 text-txt-dim">
              <i class="pi pi-map-marker text-[10px]" aria-hidden="true" />
              <span v-if="it.location_code" class="t-mono !text-[10.5px] text-acc-500 font-semibold">[{{ it.location_code }}]</span>
              <span class="truncate">{{ it.location }}</span>
            </span>
            <div class="flex items-center gap-1.5 shrink-0">
              <div v-if="it.track_stock !== false" class="text-right">
                <span class="t-num font-bold text-[13.5px]" :class="it.current_stock <= it.min_stock ? 'text-sig-bad' : 'text-sig-ok'">
                  {{ it.current_stock }}
                </span>
                <span class="text-[11px] ml-1 text-txt-dim">{{ it.unit }}</span>
              </div>
              <span
                v-else
                class="text-[10.5px] font-semibold text-sig-info bg-[var(--sig-info-bg)] border border-[var(--sig-info-line)] px-2 py-0.5 rounded-[5px]"
              >
                Aset Tetap
              </span>
              <Button
                icon="pi pi-trash"
                text
                rounded
                size="small"
                severity="danger"
                aria-label="Hapus barang"
                v-tooltip.top="'Hapus Barang'"
                @click.stop="remove(it)"
              />
            </div>
          </div>
        </div>
      </div>

      <!-- Dense Table View -->
      <div v-else class="overflow-x-auto">
        <table class="w-full text-[13px] border-collapse">
          <thead>
            <tr class="text-left border-b border-line bg-panel-2">
              <th class="px-3.5 py-2.5 font-semibold text-[11px] uppercase tracking-[0.05em] text-txt-dim">Barang &amp; SKU</th>
              <th class="px-3.5 py-2.5 font-semibold text-[11px] uppercase tracking-[0.05em] text-txt-dim w-[220px]">Kategori &amp; Ruangan</th>
              <th class="px-3.5 py-2.5 font-semibold text-[11px] uppercase tracking-[0.05em] text-txt-dim w-[140px] text-right">Stok</th>
              <th class="px-3.5 py-2.5 font-semibold text-[11px] uppercase tracking-[0.05em] text-txt-dim w-[140px]">Kondisi</th>
              <th class="px-3.5 py-2.5 font-semibold text-[11px] uppercase tracking-[0.05em] text-txt-dim w-[140px] text-center">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-line">
            <tr
              v-for="it in shown"
              :key="it.id"
              class="data-row cursor-pointer group"
              @click="router.push('/items/' + it.id)"
            >
              <td class="px-3.5 py-2.5">
                <div class="flex items-center gap-3">
                  <div class="w-9 h-9 shrink-0 rounded-[7px] overflow-hidden border border-line bg-panel-2 flex items-center justify-center">
                    <img v-if="it.photo_url" :src="it.photo_url" :alt="it.name" class="w-full h-full object-cover" />
                    <i v-else class="pi pi-box text-[13px] text-acc-500" aria-hidden="true" />
                  </div>
                  <div class="min-w-0">
                    <div class="font-semibold text-[13px] leading-snug text-txt group-hover:text-acc-500 transition-colors">
                      {{ it.name }}
                    </div>
                    <div class="t-mono !text-[11px] mt-0.5 flex items-center gap-2 flex-wrap text-txt-dim">
                      <span>{{ it.sku }}</span>
                      <span v-if="it.item_code" class="text-acc-500 font-semibold">Kode: {{ it.item_code }}</span>
                      <span v-if="it.merk">· {{ it.merk }}</span>
                    </div>
                  </div>
                </div>
              </td>

              <td class="px-3.5 py-2.5">
                <div class="font-medium text-[12.5px] text-txt truncate">{{ it.category }}</div>
                <div class="text-[11.5px] flex items-center gap-1 mt-0.5 text-txt-dim">
                  <i class="pi pi-map-marker text-[10px]" aria-hidden="true" />
                  <span v-if="it.location_code" class="t-mono !text-[10.5px] text-acc-500 font-semibold">[{{ it.location_code }}]</span>
                  <span class="truncate">{{ it.location }}</span>
                </div>
              </td>

              <td class="px-3.5 py-2.5 text-right whitespace-nowrap">
                <template v-if="it.track_stock !== false">
                  <span class="t-num font-bold text-[13.5px]" :class="it.current_stock <= it.min_stock ? 'text-sig-bad' : 'text-sig-ok'">
                    {{ it.current_stock }}
                  </span>
                  <span class="text-[11.5px] ml-1 text-txt-dim">{{ it.unit }}</span>
                  <div v-if="it.current_stock <= it.min_stock" class="t-mono !text-[10px] text-sig-bad font-semibold">
                    min {{ it.min_stock }}
                  </div>
                </template>
                <template v-else>
                  <span class="text-[11px] font-semibold text-sig-info bg-[var(--sig-info-bg)] border border-[var(--sig-info-line)] px-2 py-0.5 rounded-[5px]">
                    Aset Tetap
                  </span>
                </template>
              </td>

              <td class="px-3.5 py-2.5 whitespace-nowrap">
                <StatusChip :kind="it.condition_status" />
              </td>

              <td class="px-3.5 py-2.5 text-center whitespace-nowrap">
                <div class="flex items-center justify-center gap-0.5" @click.stop>
                  <Button
                    icon="pi pi-wrench"
                    text
                    rounded
                    size="small"
                    severity="secondary"
                    aria-label="Riwayat servis"
                    v-tooltip.top="'Servis & Kalibrasi'"
                    @click="openMaintenance(it)"
                  />
                  <Button
                    icon="pi pi-pencil"
                    text
                    rounded
                    size="small"
                    severity="secondary"
                    aria-label="Ubah barang"
                    v-tooltip.top="'Ubah Barang'"
                    @click="router.push('/items/' + it.id + '/edit')"
                  />
                  <Button
                    icon="pi pi-trash"
                    text
                    rounded
                    size="small"
                    severity="danger"
                    aria-label="Hapus barang"
                    v-tooltip.top="'Hapus Barang'"
                    @click="remove(it)"
                  />
                  <Button
                    icon="pi pi-chevron-right"
                    text
                    rounded
                    size="small"
                    severity="secondary"
                    class="group-hover:text-acc-500 group-hover:translate-x-0.5 transition-all"
                    aria-label="Lihat detail"
                    v-tooltip.top="'Lihat Detail Lengkap'"
                    @click="router.push('/items/' + it.id)"
                  />
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <template #footer>
        <div class="flex flex-wrap items-center justify-between gap-3">
          <span class="t-mono !text-[11.5px] text-txt-dim">{{ range }}</span>
          <div class="flex items-center gap-2">
            <Select
              v-model="perPage"
              :options="perPageOptions"
              optionLabel="label"
              optionValue="value"
              class="!text-[12px] w-[135px]"
            />
            <Button
              icon="pi pi-angle-left"
              size="small"
              text
              severity="secondary"
              aria-label="Halaman sebelumnya"
              :disabled="page === 0"
              @click="page--"
            />
            <span class="t-num text-[12px] px-1 text-txt">Hal. {{ page + 1 }} / {{ lastPage + 1 }}</span>
            <Button
              icon="pi pi-angle-right"
              size="small"
              text
              severity="secondary"
              aria-label="Halaman berikutnya"
              :disabled="page >= lastPage"
              @click="page++"
            />
          </div>
        </div>
      </template>
    </Panel>

    <!-- Modal Dialog: Import Excel / CSV -->
    <Dialog v-model:visible="importVisible" modal header="Import Master Barang (Excel / CSV)" :style="{ width: '620px', maxWidth: '94vw' }" class="p-fluid">
      <div class="flex flex-col gap-4 text-[12.5px] pt-1">
        <div class="p-3.5 rounded-[8px] border border-line bg-panel-2 flex flex-col gap-2.5">
          <div class="font-semibold flex items-center gap-1.5 text-acc-500">
            <i class="pi pi-info-circle text-[13px]" aria-hidden="true" /> Langkah 1: Unduh Format Template
          </div>
          <p class="text-[12px] text-txt-dim leading-relaxed">
            Template spreadsheet resmi mendukung seluruh kolom formulir barang, termasuk format Buku Inventaris KIR/KIB (Kode Lokasi, Kode Barang, Register, Merk, Seri, Ukuran, Bahan, Tahun, Kondisi) serta detail pengadaan.
          </p>
          <div class="flex flex-wrap gap-2 pt-1">
            <Button label="Unduh Template Excel (.xlsx)" icon="pi pi-file-excel" size="small" severity="success" @click="downloadTemplate('xlsx')" />
            <Button label="Unduh Template CSV (.csv)" icon="pi pi-file" size="small" severity="secondary" outlined @click="downloadTemplate('csv')" />
          </div>
          <div class="text-[11px] text-txt-dim border-t border-line pt-2 mt-1 flex flex-col gap-1">
            <div><strong class="text-txt">Kolom Didukung:</strong> Nama Barang, Kategori, Lokasi, No. Kode Lokasi, No. Kode Barang, Register, Merk, Model, No. Seri, Ukuran, Bahan, Tahun, Kondisi, Stok, Satuan, Min Stok, Harga, Sumber Dana, Distributor, AKL/AKD, Keterangan, SKU.</div>
            <div class="text-acc-500"><em>*SKU kustom opsional (otomatis dibuatkan jika kosong). Kolom kosong pada format KIR/KIB tetap valid untuk barang habis pakai.</em></div>
          </div>
        </div>

        <div class="flex flex-col gap-2">
          <span class="font-semibold text-txt">Langkah 2: Pilih Berkas yang Telah Diisi</span>
          <input
            ref="importFileInput"
            type="file"
            accept=".xlsx, .xls, .csv"
            class="block w-full text-[12px] text-txt-dim file:mr-3 file:py-2 file:px-3 file:rounded-[7px] file:border file:border-line file:text-[12px] file:font-semibold file:bg-acc-500 file:text-white hover:file:bg-acc-600 cursor-pointer"
            @change="onImportFileChanged"
          />
        </div>

        <div
          v-if="importResult"
          class="p-3 rounded-[8px] border text-[12px] flex flex-col gap-2"
          :class="(importResult.count > 0 || (importResult.skipped ?? 0) > 0)
            ? 'bg-[var(--sig-ok-bg)] border-[var(--sig-ok-line)]'
            : 'bg-[var(--sig-bad-bg)] border-[var(--sig-bad-line)]'"
        >
          <div class="flex items-center justify-between font-semibold">
            <span :class="(importResult.count > 0 || (importResult.skipped ?? 0) > 0) ? 'text-sig-ok' : 'text-sig-bad'">
              {{ (importResult.count > 0 || (importResult.skipped ?? 0) > 0) ? 'Import Selesai' : 'Import Gagal' }}
            </span>
            <span class="t-num text-txt">Total: {{ importResult.total }} baris</span>
          </div>
          <div class="flex items-center gap-3 text-[11.5px] flex-wrap">
            <span class="text-sig-ok font-semibold">Barang Baru: {{ importResult.count }}</span>
            <span v-if="importResult.skipped" class="text-sig-warn font-semibold">Dilewati: {{ importResult.skipped }}</span>
            <span class="text-sig-bad font-semibold">Gagal: {{ importResult.failed }}</span>
          </div>
          <div v-if="importResult.errors && importResult.errors.length" class="max-h-32 overflow-y-auto bg-panel border border-line p-2 rounded-[6px] t-mono !text-[11px] text-sig-bad flex flex-col gap-1">
            <div v-for="(err, i) in importResult.errors" :key="i">• {{ err }}</div>
          </div>
        </div>

        <div class="flex justify-end gap-2 pt-2 border-t border-line">
          <Button label="Batal" size="small" severity="secondary" text @click="importVisible = false" />
          <Button
            label="Mulai Import Data"
            icon="pi pi-upload"
            size="small"
            :loading="importing"
            :disabled="!selectedImportFile || importing"
            @click="doImport"
          />
        </div>
      </div>
    </Dialog>

    <!-- Modal Dialog: Servis & Kalibrasi -->
    <Dialog v-model:visible="maintenanceVisible" modal :style="{ width: '680px', maxWidth: '94vw' }" class="p-fluid">
      <template #header>
        <div class="flex items-center gap-2.5">
          <i class="pi pi-wrench text-acc-500 text-base" aria-hidden="true" />
          <div class="leading-tight">
            <div class="font-bold text-[14px] text-txt">Riwayat Servis &amp; Kalibrasi</div>
            <div class="t-mono !text-[11px] text-txt-dim">{{ selectedItemForMaint?.name }} ({{ selectedItemForMaint?.sku }})</div>
          </div>
        </div>
      </template>

      <div class="flex flex-col gap-4 pt-1 text-[12.5px]">
        <div class="flex items-center justify-between">
          <span class="font-semibold text-txt-dim">Catatan Servis &amp; Pemeriksaan</span>
          <Button
            :label="showAddMaintForm ? 'Tutup Formulir' : 'Catat Servis Baru'"
            :icon="showAddMaintForm ? 'pi pi-times' : 'pi pi-plus'"
            size="small"
            :severity="showAddMaintForm ? 'secondary' : 'primary'"
            @click="showAddMaintForm = !showAddMaintForm"
          />
        </div>

        <div v-if="showAddMaintForm" class="p-4 rounded-[8px] border border-line bg-panel-2 flex flex-col gap-3">
          <div class="font-semibold text-acc-500 text-[12px] flex items-center gap-1.5">
            <i class="pi pi-plus-circle" aria-hidden="true" /> Formulir Pemeliharaan &amp; Kalibrasi
          </div>

          <div class="grid sm:grid-cols-2 gap-3">
            <label class="flex flex-col gap-1">
              <span class="text-[11.5px] font-medium text-txt-dim">Jenis Tindakan</span>
              <Select v-model="maintForm.service_type" :options="serviceTypeOptions" class="w-full !text-[12px]" />
            </label>
            <label class="flex flex-col gap-1">
              <span class="text-[11.5px] font-medium text-txt-dim">Tanggal Servis / Kalibrasi</span>
              <InputText v-model="maintForm.service_date" type="date" class="w-full !text-[12px]" />
            </label>
            <label class="flex flex-col gap-1">
              <span class="text-[11.5px] font-medium text-txt-dim">Teknisi / Vendor Pelaksana</span>
              <InputText v-model="maintForm.vendor_or_technician" placeholder="mis. PT Medika Solusi / BPFK" class="w-full !text-[12px]" />
            </label>
            <label class="flex flex-col gap-1">
              <span class="text-[11.5px] font-medium text-txt-dim">Biaya Servis (Rp)</span>
              <InputNumber v-model="maintForm.cost" :min="0" mode="currency" currency="IDR" locale="id-ID" class="w-full !text-[12px]" />
            </label>
            <label class="flex flex-col gap-1">
              <span class="text-[11.5px] font-medium text-txt-dim">Jadwal Kalibrasi Berikutnya (Opsional)</span>
              <InputText v-model="maintForm.next_service_date" type="date" class="w-full !text-[12px]" />
            </label>
            <label class="flex flex-col gap-1">
              <span class="text-[11.5px] font-medium text-txt-dim">Update Kondisi Barang</span>
              <Select v-model="maintForm.update_condition" :options="conditionOptions" class="w-full !text-[12px]" />
            </label>
            <label class="flex flex-col gap-1 sm:col-span-2">
              <span class="text-[11.5px] font-medium text-txt-dim">Keterangan / Hasil Pengujian</span>
              <Textarea v-model="maintForm.description" rows="2" placeholder="Catatan hasil servis, nomor sertifikat kalibrasi, dll." class="w-full !text-[12px]" />
            </label>
            <div class="sm:col-span-2">
              <PhotoUploader
                ref="maintPhotoRef"
                :defer-upload="true"
                v-model="maintForm.photo_url"
                :location-name="selectedItemForMaint?.location"
                label="Foto Nota / Sertifikat Kalibrasi (Opsional)"
                hint="Dapat dicap lokasi GPS"
              />
            </div>
          </div>

          <div class="flex justify-end gap-2 pt-2 border-t border-line">
            <Button label="Batal" size="small" severity="secondary" text @click="showAddMaintForm = false" />
            <Button label="Simpan Catatan Servis" icon="pi pi-check" size="small" :loading="submittingMaint" @click="submitMaintenance" />
          </div>
        </div>

        <div v-if="loadingMaint" class="space-y-2.5 py-2">
          <div v-for="n in 3" :key="n" class="skeleton h-16 w-full" />
        </div>

        <EmptyState
          v-else-if="!maintenanceList.length"
          icon="pi pi-wrench"
          title="Belum ada catatan servis"
          sub="Belum ada catatan servis atau kalibrasi untuk barang ini."
        >
          <Button label="Tambah Catatan Pertama" icon="pi pi-plus" size="small" severity="secondary" outlined @click="showAddMaintForm = true" />
        </EmptyState>

        <div v-else class="flex flex-col gap-2.5 max-h-[50vh] overflow-y-auto pr-1">
          <div
            v-for="m in maintenanceList"
            :key="m.id"
            class="p-3.5 rounded-[8px] border border-line bg-panel-2 flex flex-col gap-2"
          >
            <div class="flex items-start justify-between gap-2">
              <div class="flex items-center gap-2 flex-wrap">
                <span
                  class="px-2 py-0.5 rounded-[5px] text-[11px] font-semibold border"
                  :class="m.service_type === 'Kalibrasi'
                    ? 'bg-[var(--sig-info-bg)] text-sig-info border-[var(--sig-info-line)]'
                    : 'bg-[var(--acc-wash)] text-acc-500 border-[var(--acc-border)]'"
                >
                  {{ m.service_type }}
                </span>
                <span class="t-mono !text-[12px] font-semibold text-txt">{{ m.service_date }}</span>
                <span v-if="m.vendor_or_technician" class="text-txt-dim text-[11.5px]">· Oleh: {{ m.vendor_or_technician }}</span>
              </div>
              <div class="flex items-center gap-1.5 shrink-0">
                <span v-if="m.cost" class="t-num font-semibold text-sig-ok text-[12px]">
                  Rp {{ Number(m.cost).toLocaleString('id-ID') }}
                </span>
                <Button
                  icon="pi pi-trash"
                  text
                  rounded
                  size="small"
                  severity="danger"
                  aria-label="Hapus catatan servis"
                  v-tooltip.top="'Hapus Catatan'"
                  @click="deleteMaintRecord(m.id)"
                />
              </div>
            </div>

            <p v-if="m.description" class="text-txt-dim text-[12px] leading-relaxed">
              {{ m.description }}
            </p>

            <div class="flex items-center justify-between text-[11px] pt-1.5 border-t border-line-soft">
              <div v-if="m.next_service_date" class="flex items-center gap-1.5 font-semibold text-sig-ok">
                <i class="pi pi-calendar text-[10px]" aria-hidden="true" />
                <span>Jadwal Berikutnya: {{ m.next_service_date }}</span>
              </div>
              <div v-else />

              <a
                v-if="safeHref(m.photo_url)"
                :href="safeHref(m.photo_url)"
                target="_blank"
                rel="noopener noreferrer"
                class="text-acc-500 hover:underline flex items-center gap-1 font-medium"
              >
                <i class="pi pi-image text-[10px]" aria-hidden="true" /> Lihat Bukti Nota
              </a>
            </div>
          </div>
        </div>

        <div class="flex justify-end pt-2 border-t border-line">
          <Button label="Tutup" size="small" severity="secondary" @click="maintenanceVisible = false" />
        </div>
      </div>
    </Dialog>
  </div>
</template>
