<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '@/api'
import { useConfirm } from 'primevue/useconfirm'
import { useToast } from 'primevue/usetoast'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import InputNumber from 'primevue/inputnumber'
import Textarea from 'primevue/textarea'
import Select from 'primevue/select'
import StatusChip from '@/components/StatusChip.vue'
import StatCard from '@/components/StatCard.vue'
import Panel from '@/components/Panel.vue'
import EmptyState from '@/components/EmptyState.vue'
import PhotoUploader from '@/components/PhotoUploader.vue'

const route = useRoute()
const router = useRouter()
const confirm = useConfirm()
const toast = useToast()

const item = ref<any>(null)
const loading = ref(true)
const photoLightbox = ref(false)
const lightboxUrl = ref('')
const deleting = ref(false)

// Stock Movement State (Stok Masuk & Stok Keluar dengan bukti foto geotag)
const stockModalType = ref<'IN' | 'OUT' | null>(null)
const submittingStock = ref(false)
const stockPhotoRef = ref<any>(null)
const maintPhotoRef = ref<any>(null)
const stockForm = ref({
  quantity: 1,
  received_by: '',
  notes: '',
  photo_url: '',
  geo_lat: null,
  geo_lng: null,
  geo_acc: null,
  geo_name: '',
})
const stockTransactions = ref<any[]>([])
const loadingStockHistory = ref(false)

// Maintenance state
const maintenanceList = ref<any[]>([])
const loadingMaint = ref(false)
const showMaintModal = ref(false)
const submittingMaint = ref(false)
const activeTab = ref<'stock_history' | 'spec' | 'maintenance'>('spec')

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

async function fetchItem() {
  const currentId = route.params.id as string
  if (!currentId || currentId === 'undefined') {
    router.push('/items')
    return
  }
  loading.value = true
  try {
    const res = await api.get(`/items/${currentId}`)
    item.value = res.data
    maintForm.value.update_condition = res.data.condition_status || 'Berfungsi'
    if (res.data.track_stock !== false) {
      activeTab.value = 'stock_history'
    } else {
      activeTab.value = 'spec'
    }
  } catch {
    toast.add({ severity: 'error', summary: 'Barang tidak ditemukan', life: 3000 })
    router.push('/items')
  } finally {
    loading.value = false
  }
}

async function fetchMaintenance() {
  const currentId = route.params.id as string
  if (!currentId || currentId === 'undefined') return
  loadingMaint.value = true
  try {
    const res = await api.get(`/items/${currentId}/maintenance`)
    maintenanceList.value = res.data
  } catch {
    // silent
  } finally {
    loadingMaint.value = false
  }
}

async function fetchStockHistory() {
  if (!item.value?.sku) return
  loadingStockHistory.value = true
  try {
    const res = await api.get('/transactions', { params: { sku: item.value.sku, per_page: 50 } })
    const list = res.data?.data || (Array.isArray(res.data) ? res.data : [])
    stockTransactions.value = Array.isArray(list) ? list : []
  } catch {
    stockTransactions.value = []
  } finally {
    loadingStockHistory.value = false
  }
}

onMounted(async () => {
  await fetchItem()
  if (item.value) {
    fetchMaintenance()
    if (item.value.track_stock !== false) {
      fetchStockHistory()
    }
  }
})

const hasCoords = computed(() => {
  return item.value && item.value.geo_lat != null && item.value.geo_lng != null
})

const googleMapsUrl = computed(() => {
  if (!hasCoords.value) return ''
  return `https://www.google.com/maps?q=${item.value.geo_lat},${item.value.geo_lng}`
})

const totalAssetValue = computed(() => {
  if (!item.value) return 0
  const qty = item.value.track_stock !== false ? (item.value.current_stock || 0) : 1
  const price = Number(item.value.price_per_unit) || 0
  return qty * price
})

function formatRupiah(val: number) {
  if (!val) return '—'
  return 'Rp ' + Number(val).toLocaleString('id-ID')
}

async function downloadImage(url: string, filename = 'foto_geotag.jpg') {
  if (!url) return
  try {
    const res = await fetch(url)
    const blob = await res.blob()
    const blobUrl = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = blobUrl
    a.download = filename
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(blobUrl)
  } catch {
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    a.target = '_blank'
    a.click()
  }
}

function openStockModal(type: 'IN' | 'OUT') {
  stockModalType.value = type
  stockForm.value = {
    quantity: 1,
    received_by: '',
    notes: '',
    photo_url: '',
    geo_lat: null,
    geo_lng: null,
    geo_acc: null,
    geo_name: item.value?.location || '',
  }
}

async function submitStockMovement() {
  if (stockForm.value.quantity <= 0) {
    toast.add({ severity: 'warn', summary: 'Jumlah harus lebih dari 0', life: 2500 })
    return
  }
  submittingStock.value = true
  const endpoint = stockModalType.value === 'IN' ? '/movement/in' : '/movement/out'
  try {
    if (stockPhotoRef.value?.hasPendingPhoto) {
      const u = await stockPhotoRef.value.uploadPending()
      if (u) stockForm.value.photo_url = u
    }
    const res = await api.post(endpoint, {
      item_id: item.value.id,
      quantity: stockForm.value.quantity,
      received_by: stockForm.value.received_by,
      notes: stockForm.value.notes,
      photo_url: stockForm.value.photo_url,
    })
    const d = res.data
    toast.add({
      severity: 'success',
      summary: stockModalType.value === 'IN' ? 'Stok Masuk Berhasil' : 'Stok Keluar Berhasil',
      detail: `Stok berubah: ${d.previous} ➔ ${d.new} ${d.unit}`,
      life: 3500,
    })
    stockModalType.value = null
    await fetchItem()
    await fetchStockHistory()
  } catch (err: any) {
    toast.add({
      severity: 'error',
      summary: 'Gagal mencatat mutasi stok',
      detail: err.response?.data?.error || err.message,
      life: 4000,
    })
  } finally {
    submittingStock.value = false
  }
}

function remove() {
  const targetId = item.value?.id || (route.params.id as string)
  if (!targetId || targetId === 'undefined') {
    toast.add({ severity: 'error', summary: 'ID barang tidak valid', life: 3000 })
    return
  }
  confirm.require({
    message: `Hapus barang "${item.value?.name || 'ini'}" (${item.value?.sku || ''}) beserta seluruh riwayatnya? Tindakan ini tidak dapat dibatalkan.`,
    header: 'Konfirmasi Hapus Barang',
    icon: 'pi pi-exclamation-triangle',
    acceptClass: 'p-button-danger',
    accept: async () => {
      deleting.value = true
      try {
        await api.delete(`/items/${targetId}`)
        toast.add({ severity: 'success', summary: 'Barang berhasil dihapus', detail: item.value?.name, life: 2500 })
        router.push('/items')
      } catch (e: any) {
        toast.add({ severity: 'error', summary: 'Gagal menghapus barang', detail: e.response?.data?.error || e.message, life: 4000 })
      } finally {
        deleting.value = false
      }
    },
  })
}

async function submitMaintenance() {
  submittingMaint.value = true
  try {
    if (maintPhotoRef.value?.hasPendingPhoto) {
      const u = await maintPhotoRef.value.uploadPending()
      if (u) maintForm.value.photo_url = u
    }
    await api.post(`/items/${route.params.id}/maintenance`, maintForm.value)
    toast.add({ severity: 'success', summary: 'Catatan servis tersimpan', life: 3000 })
    showMaintModal.value = false
    await fetchMaintenance()
    await fetchItem()
  } catch (err: any) {
    toast.add({ severity: 'error', summary: 'Gagal menyimpan', detail: err.response?.data?.error || err.message, life: 3500 })
  } finally {
    submittingMaint.value = false
  }
}

async function deleteMaintRecord(mId: string) {
  try {
    await api.delete(`/items/${route.params.id}/maintenance/${mId}`)
    toast.add({ severity: 'info', summary: 'Catatan servis dihapus', life: 2500 })
    fetchMaintenance()
  } catch (err: any) {
    toast.add({ severity: 'error', summary: 'Gagal menghapus', detail: err.response?.data?.error || err.message, life: 3000 })
  }
}
</script>

<template>
  <div class="py-2 pb-16 w-full flex flex-col gap-4">
    <!-- Clean Unified Action Toolbar -->
    <div class="panel p-3 flex flex-wrap items-center justify-between gap-2.5">
      <Button
        label="Daftar Barang"
        icon="pi pi-arrow-left"
        size="small"
        text
        severity="secondary"
        @click="router.push('/items')"
      />

      <!-- Action buttons -->
      <div v-if="item" class="flex items-center gap-2 flex-wrap">
        <!-- Stock buttons for consumable -->
        <template v-if="item.track_stock !== false">
          <Button
            label="+ Stok Masuk"
            icon="pi pi-arrow-down-left"
            size="small"
            severity="success"
            @click="openStockModal('IN')"
          />
          <Button
            label="− Stok Keluar"
            icon="pi pi-arrow-up-right"
            size="small"
            severity="danger"
            :disabled="item.current_stock <= 0"
            @click="openStockModal('OUT')"
          />
        </template>

        <!-- Service button for fixed asset -->
        <Button
          v-else
          label="Catat Servis"
          icon="pi pi-wrench"
          size="small"
          severity="warn"
          @click="showMaintModal = true"
        />

        <Button
          label="Edit Data"
          icon="pi pi-pencil"
          size="small"
          severity="secondary"
          outlined
          @click="router.push('/items/' + route.params.id + '/edit')"
        />
        <Button
          label="Label QR"
          icon="pi pi-qrcode"
          size="small"
          severity="secondary"
          outlined
          @click="router.push('/barcode')"
        />
        <Button
          label="Hapus"
          icon="pi pi-trash"
          size="small"
          severity="danger"
          text
          :loading="deleting"
          @click="remove"
        />
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="panel p-16 grid place-items-center">
      <div class="flex flex-col items-center gap-3">
        <i class="pi pi-spin pi-spinner text-3xl text-indigo-500" />
        <span class="text-[12.5px]" style="color: var(--txt-dim)">Memuat rincian inventaris...</span>
      </div>
    </div>

    <!-- Main Detail Content -->
    <div v-else-if="item" class="flex flex-col gap-4">
      <!-- 1. Header Hero Card -->
      <div class="panel p-5 flex flex-col gap-3">
        <div class="flex items-center gap-2 flex-wrap">
          <span class="t-mono text-[12px] font-bold text-indigo-600 dark:text-indigo-400 bg-indigo-50 dark:bg-indigo-950/40 px-2.5 py-0.5 rounded border border-indigo-200 dark:border-indigo-800">
            {{ item.sku }}
          </span>
          <Tag
            :severity="item.track_stock !== false ? 'warn' : 'info'"
            :value="item.track_stock !== false ? 'BARANG KONSUMABEL' : 'ASET TETAP'"
            class="!text-[10px]"
          />
          <StatusChip :kind="item.condition_status" />
          <Tag v-if="item.is_available" severity="success" value="TERSEDIA" class="!text-[10px]" />
          <Tag v-else severity="danger" value="TIDAK TERSEDIA" class="!text-[10px]" />
        </div>

        <h1 class="text-2xl sm:text-3xl font-extrabold leading-tight tracking-tight break-words" style="color: var(--txt)">
          {{ item.name }}
        </h1>

        <div class="text-[12.5px] flex items-center gap-3 flex-wrap pt-1 border-t" style="border-color: var(--line); color: var(--txt-dim)">
          <span>Kategori: <strong style="color: var(--txt)">{{ item.category }}</strong></span>
          <span>·</span>
          <span class="flex items-center gap-1.5">
            <i class="pi pi-map-marker text-indigo-500 text-xs" />
            Ruangan: <strong style="color: var(--txt)">{{ item.location }}</strong>
          </span>
          <span v-if="item.merk">·</span>
          <span v-if="item.merk">Merk: <strong style="color: var(--txt)">{{ item.merk }}</strong></span>
        </div>
      </div>

      <!-- 2. KPI Metrics Strip with StatCard -->
      <!-- Case A: Consumable Stock Item -->
      <div v-if="item.track_stock !== false" class="grid grid-cols-2 lg:grid-cols-4 gap-3">
        <StatCard
          label="Stok Tersedia"
          :value="`${item.current_stock} ${item.unit}`"
          icon="pi pi-box"
          :tone="item.current_stock <= item.min_stock ? 'bad' : 'ok'"
          :hint="item.current_stock <= item.min_stock ? '⚠️ Di bawah batas minimum' : 'Stok dalam batas aman'"
        />
        <StatCard
          label="Batas Minimum"
          :value="`${item.min_stock} ${item.unit}`"
          icon="pi pi-bell"
          tone="warn"
          hint="Peringatan bot otomatis"
        />
        <StatCard
          label="Harga Satuan"
          :value="formatRupiah(item.price_per_unit)"
          icon="pi pi-tag"
          tone="neutral"
          hint="Estimasi harga pengadaan"
        />
        <StatCard
          label="Total Nilai Fisik"
          :value="formatRupiah(totalAssetValue)"
          icon="pi pi-wallet"
          tone="accent"
          hint="Akumulasi nilai stok fisik"
        />
      </div>

      <!-- Case B: Fixed Asset Item -->
      <div v-else class="grid grid-cols-2 lg:grid-cols-4 gap-3">
        <StatCard
          label="Kondisi Fisik"
          :value="item.condition_status"
          icon="pi pi-check-circle"
          :tone="item.condition_status === 'Berfungsi' ? 'ok' : 'bad'"
          hint="Kelaikan operasional unit"
        />
        <StatCard
          label="Tipe Pengelolaan"
          value="Unit Mandiri"
          icon="pi pi-desktop"
          tone="info"
          hint="Aset tidak habis pakai"
        />
        <StatCard
          label="Tahun Pengadaan"
          :value="item.procurement_year || '—'"
          icon="pi pi-calendar"
          tone="neutral"
          hint="Tahun perolehan barang"
        />
        <StatCard
          label="Nilai Perolehan"
          :value="formatRupiah(item.price_per_unit)"
          icon="pi pi-wallet"
          tone="accent"
          hint="Estimasi nilai aset"
        />
      </div>

      <!-- 3. Photo & Geotag Banner Card -->
      <Panel v-if="item.photo_url" title="Dokumentasi Fisik &amp; Stempel Geotag" icon="pi pi-camera" dense>
        <template #actions>
          <div class="flex items-center gap-2">
            <Button
              label="Unduh Foto (JPG)"
              icon="pi pi-download"
              size="small"
              text
              severity="secondary"
              @click.stop="downloadImage(item.photo_url, (item.sku || 'foto') + '_geotag.jpg')"
            />
            <a
              v-if="hasCoords"
              :href="googleMapsUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="text-xs font-semibold text-indigo-500 hover:underline flex items-center gap-1 px-2 py-1 rounded"
              style="background: var(--panel)"
            >
              <i class="pi pi-external-link text-[10px]" /> Google Maps
            </a>
          </div>
        </template>

        <div
          class="relative w-full max-h-[420px] sm:max-h-[500px] flex items-center justify-center cursor-pointer group p-3"
          style="background: #090d16"
          @click="lightboxUrl = item.photo_url; photoLightbox = true"
        >
          <img
            :src="item.photo_url"
            :alt="item.name"
            class="max-h-[400px] sm:max-h-[480px] w-full object-contain rounded-lg group-hover:scale-[1.008] transition-transform shadow-md"
          />
          <div class="absolute bottom-5 left-5 bg-black/85 backdrop-blur-xs px-3 py-1.5 rounded-lg text-[11px] text-white flex items-center gap-2 border border-white/10 shadow-lg">
            <i class="pi pi-map-marker text-emerald-400 text-xs" />
            <span>Foto Berstempel GPS Map Camera</span>
          </div>
          <div class="absolute top-5 right-5 bg-black/80 backdrop-blur-xs px-3 py-1.5 rounded-lg text-[11px] text-white flex items-center gap-1.5 opacity-0 group-hover:opacity-100 transition-opacity border border-white/10 shadow-lg">
            <i class="pi pi-search-plus text-xs" />
            <span>Perbesar Foto</span>
          </div>
        </div>

        <template #footer>
          <div class="flex items-center justify-between text-[11.5px] flex-wrap gap-2" style="color: var(--txt-dim)">
            <div v-if="hasCoords" class="t-mono font-semibold flex items-center gap-2 text-emerald-600 dark:text-emerald-400">
              <i class="pi pi-compass text-xs" />
              <span>GPS: {{ item.geo_lat.toFixed(6) }}, {{ item.geo_lng.toFixed(6) }}</span>
              <span v-if="item.geo_acc" class="font-normal opacity-80">(akurasi ±{{ Math.round(item.geo_acc) }}m)</span>
            </div>
            <div v-else class="text-[11.5px]">Tidak ada data koordinat GPS</div>

            <div v-if="item.geo_name" class="truncate max-w-md font-medium" style="color: var(--txt)">
              <i class="pi pi-building text-[10px] mr-1 text-indigo-500" />
              {{ item.geo_name }}
            </div>
          </div>
        </template>
      </Panel>

      <!-- 4. Segmented Tabs: Riwayat Mutasi / Spesifikasi / Pemeliharaan -->
      <div class="panel overflow-hidden shadow-xs">
        <!-- Tab Navigation Bar -->
        <div class="flex border-b text-[12.5px] font-bold overflow-x-auto" style="border-color: var(--line); background: var(--panel-2)">
          <!-- Tab 1: Riwayat Mutasi Stok (Hanya barang konsumabel) -->
          <button
            v-if="item.track_stock !== false"
            class="py-3 px-4 flex items-center justify-center gap-2 transition-colors cursor-pointer border-b-2 whitespace-nowrap"
            :class="activeTab === 'stock_history'
              ? 'border-indigo-500 text-indigo-600 dark:text-indigo-400'
              : 'border-transparent hover:text-slate-900 dark:hover:text-slate-100'"
            :style="activeTab === 'stock_history' ? { background: 'var(--panel)' } : { color: 'var(--txt-dim)' }"
            @click="activeTab = 'stock_history'"
          >
            <i class="pi pi-history text-xs" />
            <span>Riwayat Mutasi Stok</span>
            <span v-if="stockTransactions.length"
                  class="text-[10px] px-2 py-0.5 rounded-full font-bold bg-indigo-100 text-indigo-700 dark:bg-indigo-950 dark:text-indigo-300">
              {{ stockTransactions.length }}
            </span>
          </button>

          <!-- Tab 2: Spesifikasi Lengkap -->
          <button
            class="py-3 px-4 flex items-center justify-center gap-2 transition-colors cursor-pointer border-b-2 whitespace-nowrap"
            :class="activeTab === 'spec'
              ? 'border-indigo-500 text-indigo-600 dark:text-indigo-400'
              : 'border-transparent hover:text-slate-900 dark:hover:text-slate-100'"
            :style="activeTab === 'spec' ? { background: 'var(--panel)' } : { color: 'var(--txt-dim)' }"
            @click="activeTab = 'spec'"
          >
            <i class="pi pi-list text-xs" />
            <span>Spesifikasi &amp; Pengadaan</span>
          </button>

          <!-- Tab 3: Riwayat Servis & Kalibrasi -->
          <button
            class="py-3 px-4 flex items-center justify-center gap-2 transition-colors cursor-pointer border-b-2 whitespace-nowrap"
            :class="activeTab === 'maintenance'
              ? 'border-indigo-500 text-indigo-600 dark:text-indigo-400'
              : 'border-transparent hover:text-slate-900 dark:hover:text-slate-100'"
            :style="activeTab === 'maintenance' ? { background: 'var(--panel)' } : { color: 'var(--txt-dim)' }"
            @click="activeTab = 'maintenance'"
          >
            <i class="pi pi-wrench text-xs" />
            <span>Riwayat Pemeliharaan &amp; Servis</span>
            <span v-if="maintenanceList.length"
                  class="text-[10px] px-2 py-0.5 rounded-full font-bold bg-indigo-100 text-indigo-700 dark:bg-indigo-950 dark:text-indigo-300">
              {{ maintenanceList.length }}
            </span>
          </button>
        </div>

        <!-- CONTENT TAB 1: Riwayat Mutasi Stok dengan Bukti Foto Geotag -->
        <div v-if="activeTab === 'stock_history'" class="p-5 flex flex-col gap-4">
          <div class="flex items-center justify-between flex-wrap gap-2">
            <div>
              <div class="font-bold text-[13px]" style="color: var(--txt)">Log Transaksi Masuk &amp; Keluar</div>
              <div class="text-[11.5px]" style="color: var(--txt-dim)">Setiap pergerakan fisik tercatat lengkap dengan bukti foto geotag</div>
            </div>
            <div class="flex gap-2">
              <Button label="+ Stok Masuk" size="small" severity="success" outlined @click="openStockModal('IN')" />
              <Button label="− Stok Keluar" size="small" severity="danger" outlined :disabled="item.current_stock <= 0" @click="openStockModal('OUT')" />
            </div>
          </div>

          <div v-if="loadingStockHistory" class="py-12 text-center">
            <i class="pi pi-spin pi-spinner text-xl text-indigo-500" />
          </div>

          <EmptyState
            v-else-if="!stockTransactions.length"
            icon="pi pi-history"
            title="Belum ada riwayat mutasi stok"
            sub="Catat penerimaan barang pertama kali untuk menambah saldo stok fisik."
          >
            <Button label="Catat Stok Masuk Pertama" icon="pi pi-arrow-down-left" size="small" severity="success" class="mt-2" @click="openStockModal('IN')" />
          </EmptyState>

          <div v-else class="flex flex-col gap-3">
            <div
              v-for="tx in stockTransactions"
              :key="tx.id"
              class="p-4 rounded-xl border flex flex-col gap-2.5 transition-colors"
              style="background: var(--panel-2); border-color: var(--line)"
            >
              <div class="flex items-start justify-between gap-3 flex-wrap">
                <div class="flex items-center gap-2.5 flex-wrap">
                  <span
                    class="px-2.5 py-0.5 rounded text-[11px] font-bold border"
                    :class="tx.type === 'IN' || tx.type === 'ADJUST+'
                      ? 'bg-emerald-50 dark:bg-emerald-950/40 text-emerald-600 dark:text-emerald-400 border-emerald-200 dark:border-emerald-800'
                      : 'bg-rose-50 dark:bg-rose-950/40 text-rose-600 dark:text-rose-400 border-rose-200 dark:border-rose-800'"
                  >
                    {{ tx.type === 'IN' ? 'Stok Masuk' : tx.type === 'OUT' ? 'Stok Keluar' : 'Penyesuaian' }}
                  </span>

                  <span
                    class="font-bold text-[14px]"
                    :class="tx.type === 'IN' || tx.type === 'ADJUST+' ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400'"
                  >
                    {{ tx.type === 'IN' || tx.type === 'ADJUST+' ? '+' : '−' }}{{ tx.quantity }} {{ tx.unit }}
                  </span>

                  <span class="text-[11.5px] t-mono" style="color: var(--txt-dim)">
                    ({{ tx.previous_stock }} ➔ {{ tx.new_stock }})
                  </span>
                </div>

                <span class="text-[11.5px] font-medium" style="color: var(--txt-dim)">{{ tx.timestamp }}</span>
              </div>

              <div class="flex items-center gap-3 text-[12px] flex-wrap" style="color: var(--txt-dim)">
                <span v-if="tx.received_by">
                  Petugas / Penerima: <strong style="color: var(--txt)">{{ tx.received_by }}</strong>
                </span>
                <span v-if="tx.notes" class="italic" style="color: var(--txt)">
                  "{{ tx.notes }}"
                </span>
              </div>

              <!-- Foto Bukti Geotag Transaksi -->
              <div v-if="tx.photo_url" class="mt-1 pt-2 border-t flex items-center justify-between" style="border-color: var(--line)">
                <button
                  class="text-[11.5px] font-semibold text-indigo-500 hover:underline flex items-center gap-1.5 cursor-pointer"
                  @click="lightboxUrl = tx.photo_url; photoLightbox = true"
                >
                  <i class="pi pi-camera text-xs" />
                  <span>Lihat Foto Bukti Geotag</span>
                </button>
                <span class="text-[10.5px] text-emerald-600 dark:text-emerald-400 font-semibold flex items-center gap-1">
                  <i class="pi pi-check-circle text-[10px]" /> Bukti GPS Tervalidasi
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- CONTENT TAB 2: Spesifikasi & Detail Pengadaan -->
        <div v-if="activeTab === 'spec'" class="p-5">
          <div class="grid sm:grid-cols-2 lg:grid-cols-3 gap-3.5 text-[12.5px]">
            <div class="p-3.5 rounded-lg border flex flex-col gap-1" style="background: var(--panel-2); border-color: var(--line)">
              <span class="text-[11px] font-semibold uppercase tracking-wider" style="color: var(--txt-dim)">Merk / Pabrikan</span>
              <span class="font-bold text-[13px]" style="color: var(--txt)">{{ item.merk || '—' }}</span>
            </div>

            <div class="p-3.5 rounded-lg border flex flex-col gap-1" style="background: var(--panel-2); border-color: var(--line)">
              <span class="text-[11px] font-semibold uppercase tracking-wider" style="color: var(--txt-dim)">Model / Tipe</span>
              <span class="font-bold text-[13px]" style="color: var(--txt)">{{ item.type_model || '—' }}</span>
            </div>

            <div class="p-3.5 rounded-lg border flex flex-col gap-1" style="background: var(--panel-2); border-color: var(--line)">
              <span class="text-[11px] font-semibold uppercase tracking-wider" style="color: var(--txt-dim)">Nomor Seri (Serial Number)</span>
              <span class="t-mono font-bold text-[13px] text-indigo-500 dark:text-indigo-400">{{ item.serial_number || '—' }}</span>
            </div>

            <div class="p-3.5 rounded-lg border flex flex-col gap-1" style="background: var(--panel-2); border-color: var(--line)">
              <span class="text-[11px] font-semibold uppercase tracking-wider" style="color: var(--txt-dim)">Tahun Pengadaan</span>
              <span class="font-bold text-[13px]" style="color: var(--txt)">{{ item.procurement_year || '—' }}</span>
            </div>

            <div class="p-3.5 rounded-lg border flex flex-col gap-1" style="background: var(--panel-2); border-color: var(--line)">
              <span class="text-[11px] font-semibold uppercase tracking-wider" style="color: var(--txt-dim)">Sumber Dana</span>
              <span class="font-semibold" style="color: var(--txt)">{{ item.funding_source || '—' }}</span>
            </div>

            <div class="p-3.5 rounded-lg border flex flex-col gap-1" style="background: var(--panel-2); border-color: var(--line)">
              <span class="text-[11px] font-semibold uppercase tracking-wider" style="color: var(--txt-dim)">Distributor / Rekanan</span>
              <span class="font-semibold" style="color: var(--txt)">{{ item.distributor || '—' }}</span>
            </div>

            <div class="p-3.5 rounded-lg border flex flex-col gap-1 sm:col-span-2 lg:col-span-3" style="background: var(--panel-2); border-color: var(--line)">
              <span class="text-[11px] font-semibold uppercase tracking-wider" style="color: var(--txt-dim)">Izin Edar AKL / AKD</span>
              <span class="t-mono font-semibold text-[13px]" style="color: var(--txt)">{{ item.akl_akd || '—' }}</span>
            </div>

            <div v-if="item.description" class="p-3.5 rounded-lg border flex flex-col gap-1 sm:col-span-2 lg:col-span-3"
                 style="background: var(--panel-2); border-color: var(--line)">
              <span class="text-[11px] font-semibold uppercase tracking-wider" style="color: var(--txt-dim)">Catatan &amp; Keterangan Tambahan</span>
              <p class="leading-relaxed text-[12px]" style="color: var(--txt)">{{ item.description }}</p>
            </div>
          </div>
        </div>

        <!-- CONTENT TAB 3: Riwayat Pemeliharaan & Servis -->
        <div v-if="activeTab === 'maintenance'" class="p-5 flex flex-col gap-4">
          <div class="flex items-center justify-between flex-wrap gap-2">
            <div>
              <div class="font-bold text-[13px]" style="color: var(--txt)">Jadwal &amp; Log Servis Aset</div>
              <div class="text-[11.5px]" style="color: var(--txt-dim)">Catatan perbaikan, servis berkala, dan masa kalibrasi alat</div>
            </div>
            <Button
              label="+ Catat Servis / Kalibrasi"
              icon="pi pi-plus"
              size="small"
              severity="primary"
              @click="showMaintModal = true"
            />
          </div>

          <div v-if="loadingMaint" class="py-12 text-center">
            <i class="pi pi-spin pi-spinner text-xl text-indigo-500" />
          </div>

          <EmptyState
            v-else-if="!maintenanceList.length"
            icon="pi pi-wrench"
            title="Belum ada riwayat servis"
            sub="Catat kegiatan pemeliharaan atau kalibrasi untuk memantau kondisi fisik aset."
          >
            <Button label="Catat Servis Pertama" icon="pi pi-plus" size="small" outlined class="mt-2" @click="showMaintModal = true" />
          </EmptyState>

          <div v-else class="flex flex-col gap-3">
            <div
              v-for="m in maintenanceList"
              :key="m.id"
              class="p-4 rounded-xl border flex flex-col gap-2.5 shadow-xs"
              style="background: var(--panel-2); border-color: var(--line)"
            >
              <div class="flex items-start justify-between gap-2 flex-wrap">
                <div class="flex items-center gap-2.5 flex-wrap">
                  <span class="px-2.5 py-0.5 rounded text-[11px] font-bold border"
                        :class="m.service_type === 'Kalibrasi'
                          ? 'bg-purple-50 dark:bg-purple-950/40 text-purple-600 dark:text-purple-400 border-purple-200 dark:border-purple-800'
                          : 'bg-indigo-50 dark:bg-indigo-950/40 text-indigo-600 dark:text-indigo-400 border-indigo-200 dark:border-indigo-800'">
                    {{ m.service_type }}
                  </span>
                  <span class="font-bold text-[13px]" style="color: var(--txt)">{{ m.service_date }}</span>
                  <span v-if="m.vendor_or_technician" class="text-[12px]" style="color: var(--txt-dim)">
                    · Oleh: <strong style="color: var(--txt)">{{ m.vendor_or_technician }}</strong>
                  </span>
                </div>
                <div class="flex items-center gap-2">
                  <span v-if="m.cost" class="t-num font-bold text-emerald-600 dark:text-emerald-400 text-[12px]">
                    Rp {{ Number(m.cost).toLocaleString('id-ID') }}
                  </span>
                  <Button icon="pi pi-trash" text rounded size="small" severity="danger" v-tooltip.top="'Hapus'"
                          @click="deleteMaintRecord(m.id)" />
                </div>
              </div>

              <p v-if="m.description" class="text-[12px] leading-relaxed" style="color: var(--txt)">
                {{ m.description }}
              </p>

              <div class="flex items-center justify-between text-[11.5px] pt-2 border-t" style="border-color: var(--line)">
                <div v-if="m.next_service_date" class="flex items-center gap-1.5 font-bold text-emerald-600 dark:text-emerald-400">
                  <i class="pi pi-calendar text-[11px]" />
                  <span>Jatuh Tempo Berikutnya: {{ m.next_service_date }}</span>
                </div>
                <div v-else></div>

                <button
                  v-if="m.photo_url"
                  class="text-indigo-500 hover:underline flex items-center gap-1 font-semibold cursor-pointer"
                  @click="lightboxUrl = m.photo_url; photoLightbox = true"
                >
                  <i class="pi pi-image text-[11px]" /> Bukti Nota / Sertifikat
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- ======================================================== -->
    <!-- Modal: Pencatatan Stok Masuk & Keluar dengan Foto Geotag -->
    <!-- ======================================================== -->
    <Dialog
      v-model:visible="stockModalType"
      modal
      :header="stockModalType === 'IN' ? 'Catat Stok Masuk (Penerimaan Barang)' : 'Catat Stok Keluar (Pemakaian Barang)'"
      :style="{ width: '560px' }"
      class="p-fluid"
    >
      <div class="flex flex-col gap-3.5 text-[12.5px] pt-1">
        <div class="p-3 rounded-lg border text-[12px] flex items-center justify-between"
             :style="{ background: stockModalType === 'IN' ? 'rgba(16, 185, 129, 0.08)' : 'rgba(239, 68, 68, 0.08)', borderColor: 'var(--line)' }">
          <div>
            <div class="text-[11px]" style="color: var(--txt-dim)">Nama Barang:</div>
            <div class="font-bold text-[13px]" style="color: var(--txt)">{{ item?.name }}</div>
          </div>
          <div class="text-right">
            <div class="text-[11px]" style="color: var(--txt-dim)">Stok Saat Ini:</div>
            <div class="font-bold text-[14px] text-indigo-500">{{ item?.current_stock }} {{ item?.unit }}</div>
          </div>
        </div>

        <div class="grid sm:grid-cols-2 gap-3">
          <label class="flex flex-col gap-1">
            <span class="text-[11.5px] font-semibold" style="color: var(--txt-dim)">
              Jumlah {{ stockModalType === 'IN' ? 'Masuk' : 'Keluar' }} ({{ item?.unit }}) *
            </span>
            <InputNumber
              v-model="stockForm.quantity"
              :min="1"
              :max="stockModalType === 'OUT' ? item?.current_stock : 999999"
              showButtons
              class="w-full !text-[12px]"
            />
          </label>

          <label class="flex flex-col gap-1">
            <span class="text-[11.5px] font-semibold" style="color: var(--txt-dim)">
              {{ stockModalType === 'IN' ? 'Petugas / Penerima' : 'Diberikan Kepada / Pemohon' }}
            </span>
            <InputText
              v-model="stockForm.received_by"
              :placeholder="stockModalType === 'IN' ? 'Nama penerima barang' : 'Nama orang / ruangan pemohon'"
              class="w-full !text-[12px]"
            />
          </label>

          <label class="flex flex-col gap-1 sm:col-span-2">
            <span class="text-[11.5px] font-semibold" style="color: var(--txt-dim)">
              {{ stockModalType === 'IN' ? 'Sumber / Catatan Pengadaan' : 'Keperluan / Catatan Pemakaian' }}
            </span>
            <Textarea
              v-model="stockForm.notes"
              rows="2"
              autoResize
              :placeholder="stockModalType === 'IN' ? 'Nomor invoice, sumber pengadaan, atau keterangan penerimaan...' : 'Keperluan kegiatan, peruntukan ruangan, dll...'"
              class="w-full !text-[12px]"
            />
          </label>

          <!-- Upload Foto Geotag Bukti Serah Terima -->
          <div class="sm:col-span-2">
            <PhotoUploader
              ref="stockPhotoRef"
              :defer-upload="true"
              v-model="stockForm.photo_url"
              v-model:geo-lat="stockForm.geo_lat"
              v-model:geo-lng="stockForm.geo_lng"
              v-model:geo-acc="stockForm.geo_acc"
              v-model:geo-name="stockForm.geo_name"
              :item-name="item?.name"
              :location-name="item?.location"
              :label="stockModalType === 'IN' ? 'Foto Bukti Penerimaan Barang (Wajib Cap Geotag)' : 'Foto Bukti Penyerahan Barang (Wajib Cap Geotag)'"
              hint="Foto fisik barang dengan cap GPS Map Camera sebagai bukti otentik"
            />
          </div>
        </div>

        <div class="flex justify-end gap-2 pt-2 border-t" style="border-color: var(--line)">
          <Button label="Batal" size="small" severity="secondary" text @click="stockModalType = null" />
          <Button
            :label="stockModalType === 'IN' ? 'Simpan Stok Masuk' : 'Simpan Stok Keluar'"
            icon="pi pi-check"
            size="small"
            :severity="stockModalType === 'IN' ? 'success' : 'danger'"
            :loading="submittingStock"
            @click="submitStockMovement"
          />
        </div>
      </div>
    </Dialog>

    <!-- Modal Form: Catat Servis & Kalibrasi -->
    <Dialog v-model:visible="showMaintModal" modal header="Catat Pemeliharaan / Kalibrasi Aset" :style="{ width: '560px' }" class="p-fluid">
      <div class="flex flex-col gap-3 text-[12.5px] pt-1">
        <div class="grid sm:grid-cols-2 gap-3">
          <label class="flex flex-col gap-1">
            <span class="text-[11px] font-semibold" style="color: var(--txt-dim)">Jenis Tindakan</span>
            <Select v-model="maintForm.service_type" :options="serviceTypeOptions" class="w-full !text-[12px]" />
          </label>
          <label class="flex flex-col gap-1">
            <span class="text-[11px] font-semibold" style="color: var(--txt-dim)">Tanggal Servis</span>
            <InputText v-model="maintForm.service_date" type="date" class="w-full !text-[12px]" />
          </label>
          <label class="flex flex-col gap-1">
            <span class="text-[11px] font-semibold" style="color: var(--txt-dim)">Teknisi / Vendor</span>
            <InputText v-model="maintForm.vendor_or_technician" placeholder="mis. PT Medika Solusi" class="w-full !text-[12px]" />
          </label>
          <label class="flex flex-col gap-1">
            <span class="text-[11px] font-semibold" style="color: var(--txt-dim)">Biaya (Rp)</span>
            <InputNumber v-model="maintForm.cost" :min="0" mode="currency" currency="IDR" locale="id-ID" class="w-full !text-[12px]" />
          </label>
          <label class="flex flex-col gap-1">
            <span class="text-[11px] font-semibold" style="color: var(--txt-dim)">Jadwal Servis / Kalibrasi Berikutnya</span>
            <InputText v-model="maintForm.next_service_date" type="date" class="w-full !text-[12px]" />
          </label>
          <label class="flex flex-col gap-1">
            <span class="text-[11px] font-semibold" style="color: var(--txt-dim)">Update Kondisi Fisik</span>
            <Select v-model="maintForm.update_condition" :options="conditionOptions" class="w-full !text-[12px]" />
          </label>
          <label class="flex flex-col gap-1 sm:col-span-2">
            <span class="text-[11px] font-semibold" style="color: var(--txt-dim)">Keterangan / Hasil Pemeriksaan</span>
            <Textarea v-model="maintForm.description" rows="2" placeholder="Nomor sertifikat kalibrasi atau rincian perbaikan..." class="w-full !text-[12px]" />
          </label>
          <div class="sm:col-span-2">
            <PhotoUploader
              ref="maintPhotoRef"
              :defer-upload="true"
              v-model="maintForm.photo_url"
              :item-name="item?.name"
              :location-name="item?.location"
              label="Foto Nota / Sertifikat Kalibrasi (Opsional)"
              hint="Dapat dicap lokasi GPS"
            />
          </div>
        </div>

        <div class="flex justify-end gap-2 pt-2 border-t" style="border-color: var(--line)">
          <Button label="Batal" size="small" severity="secondary" text @click="showMaintModal = false" />
          <Button label="Simpan Catatan Servis" icon="pi pi-check" size="small" :loading="submittingMaint"
                  @click="submitMaintenance" />
        </div>
      </div>
    </Dialog>

    <!-- Modal Lightbox Foto -->
    <Dialog v-model:visible="photoLightbox" modal header="Pratinjau Foto Berstempel Geotag" :style="{ width: '700px' }" class="p-fluid">
      <div v-if="lightboxUrl" class="flex flex-col gap-3">
        <div class="p-2 bg-black rounded-lg flex items-center justify-center">
          <img :src="lightboxUrl" alt="Foto Geotag" class="max-h-[75vh] object-contain rounded" />
        </div>
        <div class="flex items-center justify-between pt-2 border-t" style="border-color: var(--line)">
          <span class="text-xs" style="color: var(--txt-dim)">Berkas gambar berstempel GPS Map Camera</span>
          <div class="flex gap-2">
            <Button
              label="Unduh Foto (JPG)"
              icon="pi pi-download"
              size="small"
              severity="success"
              @click="downloadImage(lightboxUrl, (item?.sku || 'foto') + '_geotag.jpg')"
            />
            <Button label="Tutup" size="small" severity="secondary" text @click="photoLightbox = false" />
          </div>
        </div>
      </div>
    </Dialog>
  </div>
</template>
