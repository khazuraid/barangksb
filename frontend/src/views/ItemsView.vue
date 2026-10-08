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
const expanded = ref<string | null>(null)

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
              <th class="t-label px-4 py-2.5 w-[120px] text-right">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y" style="border-color: var(--line-soft)">
            <template v-for="it in shown" :key="it.id">
              <tr class="hover:bg-paper-2 transition-colors cursor-pointer"
                  @click="expanded = expanded === it.id ? null : it.id">
                <td class="px-4 py-2.5"><span class="t-mono font-semibold">{{ it.sku }}</span></td>
                <td class="px-4 py-2.5">
                  <div class="flex items-center gap-2.5">
                    <span v-if="it.photo_url" class="w-7 h-7 shrink-0 rounded overflow-hidden border"
                          style="border-color: var(--line)">
                      <img :src="it.photo_url" :alt="it.name" class="w-full h-full object-cover" />
                    </span>
                    <span v-else class="w-7 h-7 shrink-0 grid place-items-center rounded border"
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
                  <Button icon="pi pi-wrench" text rounded size="small" severity="warn"
                          v-tooltip.top="'Servis & Kalibrasi'" @click="openMaintenance(it)" />
                  <Button icon="pi pi-pencil" text rounded size="small" severity="secondary"
                          v-tooltip.top="'Edit'" @click="router.push(`/items/${it.id}/edit`)" />
                  <Button icon="pi pi-trash" text rounded size="small" severity="danger"
                          v-tooltip.top="'Hapus'" @click="remove(it)" />
                </td>
              </tr>
              <tr v-if="expanded === it.id">
                <td colspan="7" class="px-4 pb-3.5 pt-0" style="background: var(--paper-1)">
                  <div class="flex flex-wrap sm:flex-nowrap gap-4 text-[12px] items-center pt-2">
                    <div v-if="it.photo_url" class="relative group w-24 h-20 shrink-0 rounded overflow-hidden border bg-black/10"
                         style="border-color: var(--line)">
                      <img :src="it.photo_url" :alt="it.name" class="w-full h-full object-cover" />
                      <a :href="it.photo_url" target="_blank"
                         class="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 flex items-center justify-center text-white transition-opacity"
                         title="Lihat foto geotag asli">
                        <i class="pi pi-external-link text-xs" />
                      </a>
                    </div>
                    <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 flex-1">
                      <div>
                        <div class="t-label mb-1">Harga Satuan</div>
                        <div class="t-num">{{ it.price_per_unit ? 'Rp ' + Number(it.price_per_unit).toLocaleString('id-ID') : '—' }}</div>
                      </div>
                      <div>
                        <div class="t-label mb-1">Ketersediaan</div>
                        <div>{{ it.is_available ? 'Tersedia' : 'Tidak tersedia' }}</div>
                      </div>
                      <div>
                        <div class="t-label mb-1">Foto &amp; Geotag</div>
                        <div class="flex items-center gap-1.5">
                          <span>{{ it.photo_url ? 'Tersedia' : 'Belum ada' }}</span>
                          <a v-if="it.geo_lat && it.geo_lng"
                             :href="`https://www.google.com/maps?q=${it.geo_lat},${it.geo_lng}`"
                             target="_blank"
                             class="text-acc-500 hover:underline flex items-center gap-0.5 text-[11px]"
                             title="Buka peta lokasi di Google Maps">
                            <i class="pi pi-map-marker text-[10px]" /> Peta
                          </a>
                        </div>
                      </div>
                      <div class="flex items-end gap-1.5 flex-wrap">
                        <Button label="Servis &amp; Kalibrasi" icon="pi pi-wrench" size="small" text severity="warn"
                                @click="openMaintenance(it)" />
                        <Button label="Edit" icon="pi pi-pencil" size="small" text
                                @click="router.push(`/items/${it.id}/edit`)" />
                        <Button label="QR" icon="pi pi-qrcode" size="small" text
                                @click="router.push('/barcode')" />
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
        <!-- Action bar -->
        <div class="flex items-center justify-between">
          <span class="font-semibold text-ink-300">Catatan Servis &amp; Pemeriksaan</span>
          <Button :label="showAddMaintForm ? 'Tutup Formulir' : '+ Catat Servis Baru'"
                  :icon="showAddMaintForm ? 'pi pi-times' : 'pi pi-plus'"
                  size="small" :severity="showAddMaintForm ? 'secondary' : 'primary'"
                  @click="showAddMaintForm = !showAddMaintForm" />
        </div>

        <!-- Add Form Collapsible -->
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

        <!-- Maintenance List Table / Timeline -->
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
