<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue'
import api from '@/api'
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
import Tag from 'primevue/tag'
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
const loading = ref(true)
const viewMode = ref<'table' | 'grid'>('table')

const categories = ref<any[]>([])
const locations = ref<any[]>([])
const perPageOptions = [25, 50, 100]

// --- Import Excel / CSV state ---
const importVisible = ref(false)
const importing = ref(false)
const importFileInput = ref<HTMLInputElement | null>(null)
const selectedImportFile = ref<File | null>(null)
const importResult = ref<{ total: number; count: number; failed: number; errors: string[] } | null>(null)

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
const lowCount = computed(() => items.value.filter(i => i.current_stock <= i.min_stock).length)

function resetFilters() {
  q.value = ''; category.value = ''; location.value = ''; onlyLow.value = false
  page.value = 0; fetchItems()
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

function exportCsv() { window.open('/api/export/items.csv', '_blank') }
function exportXlsx() { window.open('/api/export/items.xlsx', '_blank') }
function printReport() { window.open('/api/report.pdf', '_blank') }

// --- Import functions ---
function downloadTemplate(fmt: 'xlsx' | 'csv') {
  window.open(`/api/items/template?fmt=${fmt}`, '_blank')
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
      toast.add({ severity: 'success', summary: 'Import Berhasil', detail: `${res.data.count} barang ditambahkan`, life: 3500 })
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

// --- Maintenance functions ---
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

async function deleteMaintRecord(mId: string) {
  if (!selectedItemForMaint.value) return
  try {
    await api.delete(`/items/${selectedItemForMaint.value.id}/maintenance/${mId}`)
    toast.add({ severity: 'info', summary: 'Catatan servis dihapus', life: 2500 })
    fetchMaintenance(selectedItemForMaint.value.id)
  } catch (err: any) {
    toast.add({ severity: 'error', summary: 'Gagal menghapus', detail: err.response?.data?.error || err.message, life: 3000 })
  }
}
</script>

<template>
  <div>
    <PageHeader crumb="Operasional" title="Daftar Barang"
      sub="Master inventaris kantor — setiap baris merepresentasikan satu SKU">
      <template #actions>
        <Button label="Import Excel / CSV" icon="pi pi-file-import" size="small" severity="secondary" outlined
                @click="importVisible = true; importResult = null; selectedImportFile = null" />
        <Button label="Tambah Barang" icon="pi pi-plus" size="small" @click="router.push('/items/new')" />
      </template>
    </PageHeader>

    <!-- Clean Unified 1-Row Toolbar -->
    <div class="panel p-3 mb-4 flex flex-wrap items-center justify-between gap-2.5">
      <div class="flex flex-wrap items-center gap-2 flex-1 min-w-[280px]">
        <!-- Search bar with icon -->
        <div class="relative flex-1 min-w-[190px] max-w-sm">
          <i class="pi pi-search absolute left-3 top-1/2 -translate-y-1/2 text-ink-400 text-xs" />
          <InputText
            v-model="q"
            placeholder="Cari nama, SKU, atau ruangan…"
            class="w-full !pl-8 !text-[12px] !py-1.5"
            @keyup.enter="page = 0; fetchItems()"
          />
        </div>

        <!-- Filter Kategori -->
        <Select
          v-model="category"
          :options="categories"
          optionLabel="name"
          optionValue="name"
          showClear
          placeholder="Kategori"
          class="!text-[12px] !py-0.5 w-[150px]"
          @change="page = 0; fetchItems()"
          filter
        />

        <!-- Filter Lokasi -->
        <Select
          v-model="location"
          :options="locations"
          optionLabel="name"
          optionValue="name"
          showClear
          placeholder="Ruangan"
          class="!text-[12px] !py-0.5 w-[140px]"
          @change="page = 0; fetchItems()"
          filter
        />

        <!-- Chip filter stok menipis -->
        <button
          class="px-2.5 py-1.5 rounded-full text-[11px] font-semibold flex items-center gap-1.5 transition-colors cursor-pointer border"
          :class="onlyLow
            ? 'bg-rose-500/20 text-rose-300 border-rose-500/50'
            : 'bg-paper-2 text-ink-300 border-line hover:border-acc-500/40'"
          @click="onlyLow = !onlyLow; page = 0; fetchItems()"
        >
          <i class="pi pi-exclamation-triangle text-[10px]" :class="onlyLow ? 'text-rose-400' : 'text-amber-400'" />
          <span>Stok Menipis</span>
          <span v-if="lowCount" class="px-1.5 py-0.2 rounded-full text-[10px] bg-rose-500/30 text-rose-300 font-bold">
            {{ lowCount }}
          </span>
        </button>

        <Button
          v-if="q || category || location || onlyLow"
          icon="pi pi-filter-slash"
          text
          rounded
          size="small"
          severity="secondary"
          v-tooltip.top="'Reset Semua Filter'"
          @click="resetFilters"
        />
      </div>

      <!-- Export & View Switcher Buttons -->
      <div class="flex items-center gap-2 shrink-0">
        <div class="flex items-center rounded-lg border p-0.5" style="border-color: var(--line); background: var(--panel-2)">
          <Button icon="pi pi-list" size="small" :text="viewMode !== 'table'" class="!p-1.5 !w-7 !h-7"
                  v-tooltip.top="'Tampilan Tabel'" @click="viewMode = 'table'" />
          <Button icon="pi pi-th-large" size="small" :text="viewMode !== 'grid'" class="!p-1.5 !w-7 !h-7"
                  v-tooltip.top="'Tampilan Kartu Grid'" @click="viewMode = 'grid'" />
        </div>
        <Button label="CSV" icon="pi pi-file" size="small" text severity="secondary" @click="exportCsv" />
        <Button label="Excel" icon="pi pi-file-excel" size="small" text severity="secondary" @click="exportXlsx" />
        <Button label="PDF" icon="pi pi-file-pdf" size="small" text severity="secondary" @click="printReport" />
      </div>
    </div>

    <!-- Calm & Clean Table -->
    <Panel title="Inventaris Aset" icon="pi pi-box" dense>
      <template #actions>
        <Tag severity="secondary" :value="range" />
      </template>

      <EmptyState v-if="loading" icon="pi pi-spin pi-spinner" title="Memuat data barang…" />
      <EmptyState v-else-if="!shown.length" icon="pi pi-box" title="Tidak ada barang"
                  sub="Ubah kata kunci pencarian atau tambahkan barang baru.">
        <Button label="Tambah barang" icon="pi pi-plus" size="small" @click="router.push('/items/new')" />
      </EmptyState>

      <!-- View: Table or Grid -->
      <div v-else-if="viewMode === 'grid'" class="p-4 grid sm:grid-cols-2 md:grid-cols-3 xl:grid-cols-4 gap-3.5" style="background: var(--ink-950)">
        <div
          v-for="it in shown"
          :key="it.id"
          class="panel p-3.5 flex flex-col justify-between hover:border-indigo-400 dark:hover:border-indigo-500 hover:shadow-md transition-all cursor-pointer group"
          @click="router.push('/items/' + it.id)"
        >
          <div>
            <div class="flex items-start justify-between gap-2 mb-2.5">
              <span class="t-mono text-[11px] font-semibold px-2 py-0.5 rounded border"
                    style="border-color: var(--line); color: var(--txt-dim); background: var(--panel-2)">
                {{ it.sku }}
              </span>
              <StatusChip :kind="it.condition_status" />
            </div>

            <div class="flex items-center gap-3 mb-3">
              <div class="w-12 h-12 rounded-lg border overflow-hidden shrink-0 grid place-items-center"
                   style="border-color: var(--line); background: var(--panel-2)">
                <img v-if="it.photo_url" :src="it.photo_url" class="w-full h-full object-cover" />
                <i v-else class="pi pi-box text-acc-500 text-[18px]" />
              </div>
              <div class="min-w-0">
                <div class="font-bold text-[13.5px] truncate group-hover:text-indigo-600 dark:group-hover:text-indigo-400 transition-colors">
                  {{ it.name }}
                </div>
                <div class="text-[12px] truncate" style="color: var(--txt-dim)">{{ it.category }}</div>
              </div>
            </div>
          </div>

          <div class="pt-2.5 border-t flex items-center justify-between text-[12px]" style="border-color: var(--line)">
            <span class="truncate flex items-center gap-1" style="color: var(--txt-dim)">
              <i class="pi pi-map-marker text-[10px]" /> {{ it.location }}
            </span>
            <div class="flex items-center gap-2">
              <div v-if="it.track_stock !== false" class="text-right shrink-0">
                <span class="font-bold text-[14px]" :class="it.current_stock <= it.min_stock ? 'text-rose-500' : 'text-emerald-500'">
                  {{ it.current_stock }}
                </span>
                <span class="text-[11px] ml-1" style="color: var(--txt-dim)">{{ it.unit }}</span>
              </div>
              <div v-else class="text-right shrink-0">
                <span class="text-[10px] font-semibold text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-950/40 px-2 py-0.5 rounded border border-blue-200 dark:border-blue-800">
                  Aset Tetap
                </span>
              </div>
              <Button
                icon="pi pi-trash"
                text
                rounded
                size="small"
                severity="danger"
                v-tooltip.top="'Hapus Barang'"
                @click.stop="remove(it)"
              />
            </div>
          </div>
        </div>
      </div>

      <div v-else class="overflow-x-auto">
        <table class="w-full text-[13px]">
          <thead>
            <tr class="text-left border-b" style="border-color: var(--line); background: var(--panel-2)">
              <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider" style="color: var(--txt-dim)">Barang &amp; SKU</th>
              <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider w-[220px]" style="color: var(--txt-dim)">Kategori &amp; Ruangan</th>
              <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider w-[140px] text-right" style="color: var(--txt-dim)">Stok</th>
              <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider w-[130px]" style="color: var(--txt-dim)">Kondisi</th>
              <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider w-[130px] text-center" style="color: var(--txt-dim)">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y" style="border-color: var(--line)">
            <tr
              v-for="it in shown"
              :key="it.id"
              class="hover:bg-slate-50 dark:hover:bg-slate-800/40 transition-colors cursor-pointer group"
              @click="router.push('/items/' + it.id)"
            >
              <!-- 1. Barang & Foto -->
              <td class="px-4 py-3">
                <div class="flex items-center gap-3">
                  <div class="w-10 h-10 shrink-0 rounded-lg overflow-hidden border flex items-center justify-center"
                       style="border-color: var(--line); background: var(--panel-2)">
                    <img v-if="it.photo_url" :src="it.photo_url" :alt="it.name" class="w-full h-full object-cover" />
                    <i v-else class="pi pi-box text-[14px] text-acc-500" />
                  </div>
                  <div class="min-w-0">
                    <div class="font-bold text-[13.5px] leading-snug group-hover:text-indigo-600 dark:group-hover:text-indigo-400 transition-colors">
                      {{ it.name }}
                    </div>
                    <div class="t-mono text-[11px] mt-0.5 flex items-center gap-2" style="color: var(--txt-dim)">
                      <span>{{ it.sku }}</span>
                      <span v-if="it.merk" class="text-[11px]">· {{ it.merk }}</span>
                    </div>
                  </div>
                </div>
              </td>

              <!-- 2. Kategori & Ruangan -->
              <td class="px-4 py-3">
                <div class="font-semibold text-[12.5px] truncate">{{ it.category }}</div>
                <div class="text-[11.5px] flex items-center gap-1 mt-0.5" style="color: var(--txt-dim)">
                  <i class="pi pi-map-marker text-[10px]" />
                  <span class="truncate">{{ it.location }}</span>
                </div>
              </td>

              <!-- 3. Stok -->
              <td class="px-4 py-3 text-right whitespace-nowrap">
                <template v-if="it.track_stock !== false">
                  <span class="t-num font-bold text-[14px]" :class="it.current_stock <= it.min_stock ? 'text-rose-500' : 'text-emerald-500'">
                    {{ it.current_stock }}
                  </span>
                  <span class="text-[11.5px] ml-1" style="color: var(--txt-dim)">{{ it.unit }}</span>
                  <div v-if="it.current_stock <= it.min_stock" class="text-[10.5px] text-rose-500 font-semibold">
                    min {{ it.min_stock }}
                  </div>
                </template>
                <template v-else>
                  <span class="text-[11px] font-semibold text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-950/40 px-2 py-0.5 rounded border border-blue-200 dark:border-blue-800">
                    Aset Tetap
                  </span>
                </template>
              </td>

              <!-- 4. Kondisi -->
              <td class="px-4 py-3 whitespace-nowrap">
                <StatusChip :kind="it.condition_status" />
              </td>

              <!-- 5. Aksi / Hapus / Detail -->
              <td class="px-4 py-3 text-center whitespace-nowrap">
                <div class="flex items-center justify-center gap-1" @click.stop>
                  <Button
                    icon="pi pi-pencil"
                    text
                    rounded
                    size="small"
                    severity="secondary"
                    v-tooltip.top="'Ubah Barang'"
                    @click="router.push('/items/' + it.id + '/edit')"
                  />
                  <Button
                    icon="pi pi-trash"
                    text
                    rounded
                    size="small"
                    severity="danger"
                    v-tooltip.top="'Hapus Barang'"
                    @click="remove(it)"
                  />
                  <Button
                    icon="pi pi-chevron-right"
                    text
                    rounded
                    size="small"
                    severity="secondary"
                    class="group-hover:text-indigo-600 dark:group-hover:text-indigo-400 group-hover:translate-x-0.5 transition-all"
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

    <!-- Modal Dialog: Import Excel / CSV -->
    <Dialog v-model:visible="importVisible" modal header="Import Master Barang (Excel / CSV)" :style="{ width: '560px' }" class="p-fluid">
      <div class="flex flex-col gap-4 text-[12.5px] pt-1">
        <div class="p-3 rounded-lg border flex flex-col gap-2" style="background: var(--paper-2); border-color: var(--line)">
          <div class="font-bold flex items-center gap-1.5 text-acc-500">
            <i class="pi pi-info-circle text-[13px]" /> Langkah 1: Unduh Format Template
          </div>
          <p class="text-[11.5px] text-ink-300 leading-relaxed">
            Gunakan template spreadsheet resmi agar kolom terpetakan otomatis ke sistem (SKU akan dibuat otomatis jika dikosongkan).
          </p>
          <div class="flex gap-2 pt-1">
            <Button label="Unduh Template Excel (.xlsx)" icon="pi pi-file-excel" size="small" severity="success"
                    @click="downloadTemplate('xlsx')" />
            <Button label="Unduh CSV (.csv)" icon="pi pi-file" size="small" severity="secondary" outlined
                    @click="downloadTemplate('csv')" />
          </div>
        </div>

        <div class="flex flex-col gap-2">
          <span class="font-bold">Langkah 2: Pilih Berkas yang Telah Diisi</span>
          <input
            ref="importFileInput"
            type="file"
            accept=".xlsx, .xls, .csv"
            class="block w-full text-[12px] text-ink-300 file:mr-3 file:py-2 file:px-3 file:rounded-md file:border-0 file:text-[12px] file:font-semibold file:bg-acc-500 file:text-ink-950 hover:file:bg-acc-400 cursor-pointer"
            @change="onImportFileChanged"
          />
        </div>

        <!-- Result Card -->
        <div v-if="importResult" class="p-3 rounded-lg border text-[12px] flex flex-col gap-2"
             :style="{ background: importResult.count > 0 ? 'rgba(16, 185, 129, 0.1)' : 'rgba(239, 68, 68, 0.1)', borderColor: 'var(--line)' }">
          <div class="flex items-center justify-between font-bold">
            <span :class="importResult.count > 0 ? 'text-emerald-400' : 'text-rose-400'">
              {{ importResult.count > 0 ? 'Import Selesai' : 'Import Gagal' }}
            </span>
            <span>Total: {{ importResult.total }} baris</span>
          </div>
          <div class="flex gap-4 text-[11.5px]">
            <span class="text-emerald-400 font-semibold">✓ Berhasil: {{ importResult.count }}</span>
            <span class="text-rose-400 font-semibold">✗ Gagal: {{ importResult.failed }}</span>
          </div>
          <div v-if="importResult.errors && importResult.errors.length" class="max-h-32 overflow-y-auto bg-black/60 p-2 rounded text-[11px] font-mono text-rose-300 flex flex-col gap-1">
            <div v-for="(err, i) in importResult.errors" :key="i">• {{ err }}</div>
          </div>
        </div>

        <div class="flex justify-end gap-2 pt-2 border-t" style="border-color: var(--line)">
          <Button label="Batal" size="small" severity="secondary" text @click="importVisible = false" />
          <Button label="Mulai Import Data" icon="pi pi-upload" size="small" :loading="importing"
                  :disabled="!selectedImportFile || importing" @click="doImport" />
        </div>
      </div>
    </Dialog>

    <!-- Modal Dialog: Servis & Kalibrasi -->
    <Dialog v-model:visible="maintenanceVisible" modal :style="{ width: '680px' }" class="p-fluid">
      <template #header>
        <div class="flex items-center gap-2">
          <i class="pi pi-wrench text-acc-500 text-lg" />
          <div class="leading-tight">
            <div class="font-bold text-[14px]">Riwayat Servis &amp; Kalibrasi</div>
            <div class="t-mono text-[11px] text-ink-400">{{ selectedItemForMaint?.name }} ({{ selectedItemForMaint?.sku }})</div>
          </div>
        </div>
      </template>

      <div class="flex flex-col gap-4 pt-1 text-[12.5px]">
        <div class="flex items-center justify-between">
          <span class="font-semibold text-ink-300">Catatan Servis &amp; Pemeriksaan</span>
          <Button :label="showAddMaintForm ? 'Tutup Formulir' : '+ Catat Servis Baru'"
                  :icon="showAddMaintForm ? 'pi pi-times' : 'pi pi-plus'"
                  size="small" :severity="showAddMaintForm ? 'secondary' : 'primary'"
                  @click="showAddMaintForm = !showAddMaintForm" />
        </div>

        <div v-if="showAddMaintForm" class="p-4 rounded-lg border flex flex-col gap-3"
             style="background: var(--paper-2); border-color: var(--line)">
          <div class="font-bold text-acc-500 text-[12px] flex items-center gap-1.5">
            <i class="pi pi-plus-circle" /> Formulir Pemeliharaan &amp; Kalibrasi
          </div>

          <div class="grid sm:grid-cols-2 gap-3">
            <label class="flex flex-col gap-1">
              <span class="text-[11px] font-semibold text-ink-400">Jenis Tindakan</span>
              <Select v-model="maintForm.service_type" :options="serviceTypeOptions" class="w-full !text-[12px]" />
            </label>
            <label class="flex flex-col gap-1">
              <span class="text-[11px] font-semibold text-ink-400">Tanggal Servis / Kalibrasi</span>
              <InputText v-model="maintForm.service_date" type="date" class="w-full !text-[12px]" />
            </label>
            <label class="flex flex-col gap-1">
              <span class="text-[11px] font-semibold text-ink-400">Teknisi / Vendor Pelaksana</span>
              <InputText v-model="maintForm.vendor_or_technician" placeholder="mis. PT Medika Solusi / BPFK" class="w-full !text-[12px]" />
            </label>
            <label class="flex flex-col gap-1">
              <span class="text-[11px] font-semibold text-ink-400">Biaya Servis (Rp)</span>
              <InputNumber v-model="maintForm.cost" :min="0" mode="currency" currency="IDR" locale="id-ID" class="w-full !text-[12px]" />
            </label>
            <label class="flex flex-col gap-1">
              <span class="text-[11px] font-semibold text-ink-400">Jadwal Kalibrasi Berikutnya (Opsional)</span>
              <InputText v-model="maintForm.next_service_date" type="date" class="w-full !text-[12px]" />
            </label>
            <label class="flex flex-col gap-1">
              <span class="text-[11px] font-semibold text-ink-400">Update Kondisi Barang</span>
              <Select v-model="maintForm.update_condition" :options="conditionOptions" class="w-full !text-[12px]" />
            </label>
            <label class="flex flex-col gap-1 sm:col-span-2">
              <span class="text-[11px] font-semibold text-ink-400">Keterangan / Hasil Pengujian</span>
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

          <div class="flex justify-end gap-2 pt-2 border-t" style="border-color: var(--line)">
            <Button label="Batal" size="small" severity="secondary" text @click="showAddMaintForm = false" />
            <Button label="Simpan Catatan Servis" icon="pi pi-check" size="small" :loading="submittingMaint"
                    @click="submitMaintenance" />
          </div>
        </div>

        <div v-if="loadingMaint" class="text-center py-8 text-ink-400">
          <i class="pi pi-spin pi-spinner text-xl text-acc-500" />
        </div>

        <div v-else-if="!maintenanceList.length" class="text-center py-10 border rounded-lg text-ink-400 flex flex-col items-center gap-2"
             style="border-color: var(--line); background: var(--paper-1)">
          <i class="pi pi-wrench text-3xl text-ink-500" />
          <span>Belum ada catatan servis atau kalibrasi untuk barang ini.</span>
          <Button label="Tambah Catatan Pertama" icon="pi pi-plus" size="small" severity="secondary" outlined
                  class="mt-1" @click="showAddMaintForm = true" />
        </div>

        <div v-else class="flex flex-col gap-2.5 max-h-[50vh] overflow-y-auto pr-1">
          <div
            v-for="m in maintenanceList"
            :key="m.id"
            class="p-3.5 rounded-lg border flex flex-col gap-2 transition-colors hover:border-acc-500/50"
            style="background: var(--paper-1); border-color: var(--line)"
          >
            <div class="flex items-start justify-between gap-2">
              <div class="flex items-center gap-2 flex-wrap">
                <span class="px-2 py-0.5 rounded text-[11px] font-bold"
                      :class="m.service_type === 'Kalibrasi' ? 'bg-purple-900/60 text-purple-300 border border-purple-700/50' : 'bg-acc-500/20 text-acc-400 border border-acc-500/40'">
                  {{ m.service_type }}
                </span>
                <span class="font-bold text-[12.5px]">{{ m.service_date }}</span>
                <span v-if="m.vendor_or_technician" class="text-ink-400 text-[11.5px]">· Oleh: {{ m.vendor_or_technician }}</span>
              </div>
              <div class="flex items-center gap-1.5 shrink-0">
                <span v-if="m.cost" class="t-num font-semibold text-emerald-400 text-[11.5px]">
                  Rp {{ Number(m.cost).toLocaleString('id-ID') }}
                </span>
                <Button icon="pi pi-trash" text rounded size="small" severity="danger" v-tooltip.top="'Hapus Catatan'"
                        @click="deleteMaintRecord(m.id)" />
              </div>
            </div>

            <p v-if="m.description" class="text-ink-300 text-[12px] leading-relaxed">
              {{ m.description }}
            </p>

            <div class="flex items-center justify-between text-[11px] pt-1 border-t" style="border-color: var(--line-soft)">
              <div v-if="m.next_service_date" class="flex items-center gap-1.5 font-semibold text-sig-ok">
                <i class="pi pi-calendar text-[10px]" />
                <span>Jadwal Berikutnya: {{ m.next_service_date }}</span>
              </div>
              <div v-else></div>

              <a v-if="m.photo_url" :href="m.photo_url" target="_blank"
                 class="text-acc-500 hover:underline flex items-center gap-1 font-medium">
                <i class="pi pi-image text-[10px]" /> Lihat Bukti Nota
              </a>
            </div>
          </div>
        </div>

        <div class="flex justify-end pt-2 border-t" style="border-color: var(--line)">
          <Button label="Tutup" size="small" severity="secondary" @click="maintenanceVisible = false" />
        </div>
      </div>
    </Dialog>
  </div>
</template>
