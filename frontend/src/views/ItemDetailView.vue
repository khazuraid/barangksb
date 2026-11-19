<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
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
const activeTab = ref<'stock_history' | 'units' | 'spec' | 'maintenance'>('spec')

// Physical Units State (Aset Tetap / BMD Multi-Unit)
const unitsList = ref<any[]>([])
const loadingUnits = ref(false)
const showUnitModal = ref(false)
const editingUnitId = ref<string | null>(null)
const submittingUnit = ref(false)
const unitPhotoRef = ref<any>(null)
const unitForm = ref({
  unit_number: 1,
  register_number: '',
  barcode: '',
  serial_number: '',
  condition_status: 'Baik (B)',
  location: '',
  location_code: '',
  notes: '',
  photo_url: '',
  geo_lat: null as number | null,
  geo_lng: null as number | null,
  geo_acc: null as number | null,
  geo_name: '',
})

async function fetchUnits() {
  const currentId = route.params.id as string
  if (!currentId || currentId === 'undefined') return
  loadingUnits.value = true
  try {
    const res = await api.get(`/items/${currentId}/units`)
    unitsList.value = res.data?.data || []
  } catch {
    unitsList.value = []
  } finally {
    loadingUnits.value = false
  }
}

function openAddUnitModal() {
  editingUnitId.value = null
  const nextNum = (unitsList.value?.length || 0) + 1
  unitForm.value = {
    unit_number: nextNum,
    register_number: String(nextNum).padStart(4, '0'),
    barcode: `${item.value?.sku || 'ITEM'}-U${nextNum}`,
    serial_number: '',
    condition_status: item.value?.condition_status || 'Baik (B)',
    location: item.value?.location || '',
    location_code: item.value?.location_code || '',
    notes: '',
    photo_url: '',
    geo_lat: null,
    geo_lng: null,
    geo_acc: null,
    geo_name: '',
  }
  showUnitModal.value = true
}

function openEditUnitModal(u: any) {
  editingUnitId.value = u.id
  unitForm.value = {
    unit_number: u.unit_number,
    register_number: u.register_number,
    barcode: u.barcode,
    serial_number: u.serial_number || '',
    condition_status: u.condition_status || 'Baik (B)',
    location: u.location || item.value?.location || '',
    location_code: u.location_code || '',
    notes: u.notes || '',
    photo_url: u.photo_url || '',
    geo_lat: u.geo_lat,
    geo_lng: u.geo_lng,
    geo_acc: u.geo_acc,
    geo_name: u.geo_name || '',
  }
  showUnitModal.value = true
}

async function submitUnit() {
  submittingUnit.value = true
  try {
    if (unitPhotoRef.value?.hasPendingPhoto) {
      const url = await unitPhotoRef.value.uploadPending()
      if (url) {
        unitForm.value.photo_url = url
      }
    }
    const currentId = route.params.id as string
    if (editingUnitId.value) {
      await api.put(`/items/${currentId}/units/${editingUnitId.value}`, unitForm.value)
      toast.add({ severity: 'success', summary: 'Unit fisik diperbarui', life: 2500 })
    } else {
      await api.post(`/items/${currentId}/units`, unitForm.value)
      toast.add({ severity: 'success', summary: 'Unit fisik berhasil ditambahkan', life: 2500 })
    }
    showUnitModal.value = false
    await fetchUnits()
    await fetchItem()
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Gagal menyimpan unit', detail: e.response?.data?.error || e.message, life: 3500 })
  } finally {
    submittingUnit.value = false
  }
}

function confirmDeleteUnit(u: any) {
  confirm.require({
    message: `Hapus unit fisik Unit #${u.unit_number} (Reg: ${u.register_number || u.barcode})? Tindakan ini tidak dapat dibatalkan.`,
    header: 'Konfirmasi Hapus Unit Fisik',
    icon: 'pi pi-exclamation-triangle',
    acceptClass: 'p-button-danger',
    acceptLabel: 'Hapus Unit',
    rejectLabel: 'Batal',
    accept: async () => {
      try {
        const currentId = route.params.id as string
        await api.delete(`/items/${currentId}/units/${u.id}`)
        toast.add({ severity: 'success', summary: 'Unit fisik dihapus', life: 2500 })
        await fetchUnits()
        await fetchItem()
      } catch (e: any) {
        toast.add({ severity: 'error', summary: 'Gagal menghapus unit', detail: e.response?.data?.error || e.message, life: 3500 })
      }
    },
  })
}

const serviceTypeOptions = ['Servis Rutin', 'Kalibrasi', 'Perbaikan', 'Pemeriksaan Berkala', 'Penggantian Suku Cadang']
const conditionOptions = ['Baik (B)', 'Kurang Baik (KB)', 'Rusak Berat (RB)', 'Berfungsi', 'Rusak Ringan', 'Perlu Kalibrasi']
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
    maintForm.value.update_condition = res.data.condition_status || 'Baik (B)'
    if (res.data.track_stock !== false) {
      activeTab.value = 'stock_history'
    } else {
      activeTab.value = 'units'
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

const siblingUnits = ref<any[]>([])
const loadingSiblings = ref(false)

async function fetchSiblingUnits() {
  if (!item.value?.item_code) {
    siblingUnits.value = []
    return
  }
  loadingSiblings.value = true
  try {
    const res = await api.get('/items', { params: { q: item.value.item_code, per_page: 100 } })
    const list = res.data?.data || []
    siblingUnits.value = list.filter((i: any) => i.item_code && i.item_code === item.value.item_code)
  } catch {
    siblingUnits.value = []
  } finally {
    loadingSiblings.value = false
  }
}

onMounted(async () => {
  await fetchItem()
  if (item.value) {
    fetchMaintenance()
    fetchSiblingUnits()
    if (item.value.track_stock !== false) {
      fetchStockHistory()
    } else {
      fetchUnits()
    }
  }
})

watch(
  () => route.params.id,
  async (newId) => {
    if (newId && newId !== 'undefined') {
      await fetchItem()
      if (item.value) {
        fetchMaintenance()
        fetchSiblingUnits()
        if (item.value.track_stock !== false) {
          fetchStockHistory()
        } else {
          fetchUnits()
        }
      }
    }
  }
)

const hasCoords = computed(() => {
  return item.value && item.value.geo_lat != null && item.value.geo_lng != null
})

const googleMapsUrl = computed(() => {
  if (!hasCoords.value) return ''
  return `https://www.google.com/maps?q=${item.value.geo_lat},${item.value.geo_lng}`
})

const totalAssetValue = computed(() => {
  if (!item.value) return 0
  const qty = item.value.track_stock !== false ? (item.value.current_stock || 0) : (unitsList.value?.length || item.value.current_stock || 1)
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

        <!-- Action buttons for fixed asset -->
        <template v-else>
          <Button
            label="+ Tambah Unit Fisik"
            icon="pi pi-plus"
            size="small"
            severity="primary"
            @click="openAddUnitModal"
          />
          <Button
            label="Catat Servis"
            icon="pi pi-wrench"
            size="small"
            severity="warn"
            @click="showMaintModal = true"
          />
        </template>

        <Button
          label="Edit Data"
          icon="pi pi-pencil"
          size="small"
          severity="secondary"
          outlined
          @click="router.push('/items/' + route.params.id + '/edit')"
        />
        <Button
          label="Salin Spesifikasi"
          icon="pi pi-copy"
          size="small"
          severity="secondary"
          outlined
          @click="router.push({ path: '/items/new', query: { copy_from: route.params.id } })"
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
    <div v-if="loading" class="flex flex-col gap-4">
      <div class="panel p-5 flex flex-col gap-3">
        <div class="flex gap-2">
          <div class="skeleton h-5 w-28" />
          <div class="skeleton h-5 w-24" />
          <div class="skeleton h-5 w-20" />
        </div>
        <div class="skeleton h-8 w-2/3" />
        <div class="skeleton h-4 w-1/2" />
      </div>
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-3">
        <div v-for="i in 4" :key="i" class="panel p-4 flex flex-col gap-2">
          <div class="skeleton h-3 w-24" />
          <div class="skeleton h-7 w-28" />
          <div class="skeleton h-3 w-32" />
        </div>
      </div>
      <div class="panel p-5 flex flex-col gap-3">
        <div class="skeleton h-10 w-full" />
        <div class="grid sm:grid-cols-3 gap-3">
          <div v-for="i in 6" :key="i" class="skeleton h-16 rounded-[8px]" />
        </div>
      </div>
    </div>

    <!-- Main Detail Content -->
    <div v-else-if="item" class="flex flex-col gap-4">
      <!-- 1. Header Hero Card -->
      <div class="panel p-5 flex flex-col gap-3">
        <div class="flex items-center gap-2 flex-wrap">
          <span class="t-mono text-[12px] font-bold text-acc-500 bg-[var(--acc-wash)] px-2.5 py-0.5 rounded-[6px] border border-[var(--acc-border)]">
            {{ item.sku }}
          </span>
          <Tag v-if="item.item_code" severity="info" class="t-mono !text-[10px]" :value="'KODE: ' + item.item_code" />
          <Tag v-if="item.register_number" severity="secondary" class="t-mono !text-[10px]" :value="'REG: ' + item.register_number" />
          <Tag
            :severity="item.track_stock !== false ? 'warn' : 'info'"
            :value="item.track_stock !== false ? 'BARANG KONSUMABEL' : 'ASET TETAP'"
            class="!text-[10px]"
          />
          <StatusChip :kind="item.condition_status" />
          <Tag v-if="item.is_available" severity="success" value="TERSEDIA" class="!text-[10px]" />
          <Tag v-else severity="danger" value="TIDAK TERSEDIA" class="!text-[10px]" />
        </div>

        <h1 class="text-2xl sm:text-3xl font-extrabold leading-tight tracking-tight break-words text-txt">
          {{ item.name }}
        </h1>

        <div class="text-[12.5px] flex items-center gap-3 flex-wrap pt-2 border-t border-line text-txt-dim">
          <span>Kategori: <strong class="text-txt">{{ item.category }}</strong></span>
          <span>·</span>
          <span class="flex items-center gap-1.5">
            <i class="pi pi-map-marker text-acc-500 text-xs" aria-hidden="true" />
            Ruangan: <strong class="text-txt">{{ item.location }}</strong>
            <span v-if="item.location_code" class="t-mono text-[11px] text-acc-500 font-semibold">[{{ item.location_code }}]</span>
          </span>
          <span v-if="item.merk">·</span>
          <span v-if="item.merk">Merk: <strong class="text-txt">{{ item.merk }}</strong></span>
        </div>
      </div>

      <!-- Sibling Units Strip (Jika ada lebih dari 1 unit dengan No. Kode Barang sama) -->
      <div v-if="siblingUnits.length > 1" class="panel p-3.5 flex flex-col gap-2.5">
        <div class="flex items-center justify-between flex-wrap gap-2 text-[12px]">
          <span class="font-bold flex items-center gap-1.5 text-txt">
            <i class="pi pi-clone text-acc-500" aria-hidden="true" />
            Unit Terdaftar (No. Kode Barang: <span class="t-mono text-acc-500">{{ item.item_code }}</span>)
            <Tag severity="info" class="!text-[10px]" :value="`${siblingUnits.length} Unit Fisik`" />
          </span>
          <Button
            label="+ Tambah Unit Lagi"
            icon="pi pi-plus"
            size="small"
            text
            severity="primary"
            class="!text-xs !p-1"
            @click="router.push({ path: '/items/new', query: { copy_from: route.params.id } })"
          />
        </div>
        <div class="flex items-center gap-2 flex-wrap">
          <button
            v-for="u in siblingUnits"
            :key="u.id"
            type="button"
            class="px-2.5 py-1.5 rounded-[8px] border text-xs flex items-center gap-2 cursor-pointer transition-colors"
            :class="u.id === item.id
              ? 'border-acc-500 bg-[var(--acc-wash)] font-bold text-acc-500 ring-1 ring-acc-500'
              : 'border-line hover:border-[var(--acc-border)] bg-panel-2 text-txt-dim hover:text-txt'"
            @click="u.id !== item.id && router.push('/items/' + u.id)"
          >
            <i :class="u.id === item.id ? 'pi pi-check-circle text-acc-500' : 'pi pi-box text-txt-muted'" class="text-[11px]" aria-hidden="true" />
            <span>{{ u.register_number ? 'Reg: ' + u.register_number : (u.sku || u.name) }}</span>
            <span class="text-[10px] opacity-75">· {{ u.location }}</span>
          </button>
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
          featured
          :hint="item.current_stock <= item.min_stock ? 'Di bawah batas minimum' : 'Stok dalam batas aman'"
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
          :tone="item.condition_status === 'Berfungsi' || item.condition_status === 'Baik (B)' ? 'ok' : 'bad'"
          featured
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
              class="text-xs font-semibold text-acc-500 hover:underline flex items-center gap-1 px-2 py-1 rounded-[6px] bg-panel border border-line"
            >
              <i class="pi pi-external-link text-[10px]" aria-hidden="true" /> Google Maps
            </a>
          </div>
        </template>

        <div
          class="relative w-full max-h-[420px] sm:max-h-[500px] flex items-center justify-center cursor-pointer group p-3 bg-ink-950"
          @click="lightboxUrl = item.photo_url; photoLightbox = true"
        >
          <img
            :src="item.photo_url"
            :alt="item.name"
            class="max-h-[400px] sm:max-h-[480px] w-full object-contain rounded-[8px] group-hover:scale-[1.008] transition-transform"
          />
          <div class="absolute bottom-5 left-5 bg-black/80 backdrop-blur-xs px-3 py-1.5 rounded-[7px] text-[11px] text-white flex items-center gap-2 border border-white/10">
            <i class="pi pi-map-marker text-acc-300 text-xs" aria-hidden="true" />
            <span>Foto Berstempel GPS Map Camera</span>
          </div>
          <div class="absolute top-5 right-5 bg-black/80 backdrop-blur-xs px-3 py-1.5 rounded-[7px] text-[11px] text-white flex items-center gap-1.5 opacity-0 group-hover:opacity-100 transition-opacity border border-white/10">
            <i class="pi pi-search-plus text-xs" aria-hidden="true" />
            <span>Perbesar Foto</span>
          </div>
        </div>

        <template #footer>
          <div class="flex items-center justify-between text-[11.5px] flex-wrap gap-2 text-txt-dim">
            <div v-if="hasCoords" class="t-mono font-semibold flex items-center gap-2 text-sig-ok">
              <i class="pi pi-compass text-xs" aria-hidden="true" />
              <span>GPS: {{ item.geo_lat.toFixed(6) }}, {{ item.geo_lng.toFixed(6) }}</span>
              <span v-if="item.geo_acc" class="font-normal opacity-80">(akurasi ±{{ Math.round(item.geo_acc) }}m)</span>
            </div>
            <div v-else class="text-[11.5px]">Tidak ada data koordinat GPS</div>

            <div v-if="item.geo_name" class="truncate max-w-md font-medium text-txt">
              <i class="pi pi-building text-[10px] mr-1 text-acc-500" aria-hidden="true" />
              {{ item.geo_name }}
            </div>
          </div>
        </template>
      </Panel>

      <!-- 4. Segmented Tabs: Riwayat Mutasi / Spesifikasi / Pemeliharaan -->
      <div class="panel overflow-hidden">
        <!-- Tab Navigation Bar -->
        <div class="flex border-b border-line bg-panel-2 text-[12.5px] font-bold overflow-x-auto" role="tablist">
          <!-- Tab 1: Riwayat Mutasi Stok (Hanya barang konsumabel) -->
          <button
            v-if="item.track_stock !== false"
            type="button"
            role="tab"
            :aria-selected="activeTab === 'stock_history'"
            class="py-3 px-4 flex items-center justify-center gap-2 transition-colors cursor-pointer border-b-2 whitespace-nowrap"
            :class="activeTab === 'stock_history'
              ? 'border-acc-500 text-acc-500 bg-panel'
              : 'border-transparent text-txt-dim hover:text-txt'"
            @click="activeTab = 'stock_history'"
          >
            <i class="pi pi-history text-xs" aria-hidden="true" />
            <span>Riwayat Mutasi Stok</span>
            <span v-if="stockTransactions.length"
                  class="t-num text-[10px] px-2 py-0.5 rounded-full font-bold bg-[var(--acc-50)] text-acc-500 border border-[var(--acc-border)]">
              {{ stockTransactions.length }}
            </span>
          </button>

          <!-- Tab 1B: Daftar Unit Fisik (Hanya Aset Tetap / Non-Konsumabel) -->
          <button
            v-if="item.track_stock === false"
            type="button"
            role="tab"
            :aria-selected="activeTab === 'units'"
            class="py-3 px-4 flex items-center justify-center gap-2 transition-colors cursor-pointer border-b-2 whitespace-nowrap"
            :class="activeTab === 'units'
              ? 'border-acc-500 text-acc-500 bg-panel'
              : 'border-transparent text-txt-dim hover:text-txt'"
            @click="activeTab = 'units'"
          >
            <i class="pi pi-th-large text-xs" aria-hidden="true" />
            <span>Daftar Unit Fisik</span>
            <span v-if="unitsList.length"
                  class="t-num text-[10px] px-2 py-0.5 rounded-full font-bold bg-[var(--acc-50)] text-acc-500 border border-[var(--acc-border)]">
              {{ unitsList.length }}
            </span>
          </button>

          <!-- Tab 2: Spesifikasi Lengkap -->
          <button
            type="button"
            role="tab"
            :aria-selected="activeTab === 'spec'"
            class="py-3 px-4 flex items-center justify-center gap-2 transition-colors cursor-pointer border-b-2 whitespace-nowrap"
            :class="activeTab === 'spec'
              ? 'border-acc-500 text-acc-500 bg-panel'
              : 'border-transparent text-txt-dim hover:text-txt'"
            @click="activeTab = 'spec'"
          >
            <i class="pi pi-list text-xs" aria-hidden="true" />
            <span>Spesifikasi &amp; Pengadaan</span>
          </button>

          <!-- Tab 3: Riwayat Servis & Kalibrasi -->
          <button
            type="button"
            role="tab"
            :aria-selected="activeTab === 'maintenance'"
            class="py-3 px-4 flex items-center justify-center gap-2 transition-colors cursor-pointer border-b-2 whitespace-nowrap"
            :class="activeTab === 'maintenance'
              ? 'border-acc-500 text-acc-500 bg-panel'
              : 'border-transparent text-txt-dim hover:text-txt'"
            @click="activeTab = 'maintenance'"
          >
            <i class="pi pi-wrench text-xs" aria-hidden="true" />
            <span>Riwayat Pemeliharaan &amp; Servis</span>
            <span v-if="maintenanceList.length"
                  class="t-num text-[10px] px-2 py-0.5 rounded-full font-bold bg-[var(--acc-50)] text-acc-500 border border-[var(--acc-border)]">
              {{ maintenanceList.length }}
            </span>
          </button>
        </div>

        <!-- CONTENT TAB 1: Riwayat Mutasi Stok dengan Bukti Foto Geotag -->
        <div v-if="activeTab === 'stock_history'" class="p-5 flex flex-col gap-4">
          <div class="flex items-center justify-between flex-wrap gap-2">
            <div>
              <div class="font-bold text-[13px] text-txt">Log Transaksi Masuk &amp; Keluar</div>
              <div class="text-[11.5px] text-txt-dim">Setiap pergerakan fisik tercatat lengkap dengan bukti foto geotag</div>
            </div>
            <div class="flex gap-2">
              <Button label="+ Stok Masuk" size="small" severity="success" outlined @click="openStockModal('IN')" />
              <Button label="− Stok Keluar" size="small" severity="danger" outlined :disabled="item.current_stock <= 0" @click="openStockModal('OUT')" />
            </div>
          </div>

          <EmptyState v-if="loadingStockHistory" loading title="Memuat riwayat mutasi stok…" />

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
              class="p-4 rounded-[10px] border border-line bg-panel-2 flex flex-col gap-2.5 transition-colors"
            >
              <div class="flex items-start justify-between gap-3 flex-wrap">
                <div class="flex items-center gap-2.5 flex-wrap">
                  <span
                    class="px-2.5 py-0.5 rounded-[6px] text-[11px] font-bold border"
                    :class="tx.type === 'IN' || tx.type === 'ADJUST+'
                      ? 'bg-[var(--sig-ok-bg)] text-sig-ok border-[var(--sig-ok-line)]'
                      : 'bg-[var(--sig-bad-bg)] text-sig-bad border-[var(--sig-bad-line)]'"
                  >
                    {{ tx.type === 'IN' ? 'Stok Masuk' : tx.type === 'OUT' ? 'Stok Keluar' : 'Penyesuaian' }}
                  </span>

                  <span
                    class="t-num font-bold text-[14px]"
                    :class="tx.type === 'IN' || tx.type === 'ADJUST+' ? 'text-sig-ok' : 'text-sig-bad'"
                  >
                    {{ tx.type === 'IN' || tx.type === 'ADJUST+' ? '+' : '−' }}{{ tx.quantity }} {{ tx.unit }}
                  </span>

                  <span class="text-[11.5px] t-mono text-txt-dim">
                    ({{ tx.previous_stock }} → {{ tx.new_stock }})
                  </span>
                </div>

                <span class="t-mono text-[11.5px] font-medium text-txt-dim">{{ tx.timestamp }}</span>
              </div>

              <div class="flex items-center gap-3 text-[12px] flex-wrap text-txt-dim">
                <span v-if="tx.received_by">
                  Petugas / Penerima: <strong class="text-txt">{{ tx.received_by }}</strong>
                </span>
                <span v-if="tx.notes" class="italic text-txt">
                  "{{ tx.notes }}"
                </span>
              </div>

              <!-- Foto Bukti Geotag Transaksi -->
              <div v-if="tx.photo_url" class="mt-1 pt-2 border-t border-line flex items-center justify-between">
                <button
                  type="button"
                  class="text-[11.5px] font-semibold text-acc-500 hover:underline flex items-center gap-1.5 cursor-pointer"
                  @click="lightboxUrl = tx.photo_url; photoLightbox = true"
                >
                  <i class="pi pi-camera text-xs" aria-hidden="true" />
                  <span>Lihat Foto Bukti Geotag</span>
                </button>
                <span class="text-[10.5px] text-sig-ok font-semibold flex items-center gap-1">
                  <i class="pi pi-check-circle text-[10px]" aria-hidden="true" /> Bukti GPS Tervalidasi
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- CONTENT TAB 1B: Daftar Unit Fisik (Aset Tetap / BMD Multi-Unit) -->
        <div v-if="activeTab === 'units'" class="p-5 flex flex-col gap-4">
          <div class="flex items-center justify-between flex-wrap gap-2">
            <div>
              <div class="font-bold text-[13px] text-txt">Daftar Unit Fisik Terdaftar</div>
              <div class="text-[11.5px] text-txt-dim">
                Setiap unit fisik memiliki nomor register, barcode, dan foto stempel GPS tersendiri
              </div>
            </div>
            <div class="flex gap-2">
              <Button label="+ Tambah Unit Fisik" icon="pi pi-plus" size="small" severity="primary" @click="openAddUnitModal" />
            </div>
          </div>

          <EmptyState v-if="loadingUnits" loading title="Memuat daftar unit fisik…" />

          <EmptyState
            v-else-if="!unitsList.length"
            icon="pi pi-th-large"
            title="Belum ada unit fisik terpisah"
            sub="Tambahkan unit fisik untuk mengelola nomor register, serial number, barcode, dan foto masing-masing unit."
          >
            <Button label="+ Tambah Unit Fisik Pertama" icon="pi pi-plus" size="small" severity="primary" class="mt-2" @click="openAddUnitModal" />
          </EmptyState>

          <div v-else class="flex flex-col gap-3">
            <div
              v-for="u in unitsList"
              :key="u.id"
              class="p-4 rounded-[10px] border border-line bg-panel-2 flex flex-col gap-2.5 transition-colors"
            >
              <div class="flex items-start justify-between gap-3 flex-wrap">
                <div class="flex items-center gap-2.5 flex-wrap">
                  <span class="px-2.5 py-0.5 rounded-[6px] text-[11px] font-bold border bg-[var(--acc-50)] text-acc-500 border-[var(--acc-border)] t-num">
                    Unit #{{ u.unit_number }}
                  </span>

                  <span v-if="u.register_number" class="px-2 py-0.5 rounded-[6px] text-[11px] font-bold border t-mono bg-panel text-txt border-line">
                    Reg: {{ u.register_number }}
                  </span>

                  <span class="font-bold text-[13px] t-mono text-acc-500">
                    {{ u.barcode }}
                  </span>

                  <StatusChip :kind="u.condition_status" />
                </div>

                <div class="flex items-center gap-1.5 ml-auto">
                  <a :href="`/api/barcode/${u.barcode}.png?fmt=qr&size=300`" target="_blank" rel="noopener noreferrer" aria-label="Lihat atau cetak QR unit">
                    <Button icon="pi pi-qrcode" size="small" text severity="secondary" aria-label="Lihat atau cetak QR unit" v-tooltip.top="'Lihat / Cetak QR'" class="!p-1.5" />
                  </a>
                  <Button icon="pi pi-pencil" size="small" text severity="info" aria-label="Edit unit fisik" v-tooltip.top="'Edit Unit'" class="!p-1.5" @click="openEditUnitModal(u)" />
                  <Button icon="pi pi-trash" size="small" text severity="danger" aria-label="Hapus unit fisik" v-tooltip.top="'Hapus Unit'" class="!p-1.5" @click="confirmDeleteUnit(u)" />
                </div>
              </div>

              <!-- Location, Serial Number, Notes -->
              <div class="grid sm:grid-cols-3 gap-2 text-[12px] pt-1 text-txt-dim">
                <div v-if="u.location" class="flex items-center gap-1.5">
                  <i class="pi pi-building text-[11px] text-acc-500" aria-hidden="true" />
                  <span>Ruangan: <strong class="text-txt">{{ u.location }}</strong></span>
                </div>
                <div v-if="u.serial_number" class="flex items-center gap-1.5">
                  <i class="pi pi-hashtag text-[11px] text-acc-500" aria-hidden="true" />
                  <span>SN: <strong class="t-mono text-txt">{{ u.serial_number }}</strong></span>
                </div>
                <div v-if="u.notes" class="sm:col-span-3 italic text-[11.5px] text-txt">
                  "{{ u.notes }}"
                </div>
              </div>

              <!-- Photo with Geotag if exists -->
              <div v-if="u.photo_url" class="mt-1 pt-2 border-t border-line flex items-center justify-between flex-wrap gap-2">
                <button
                  type="button"
                  class="text-[11.5px] font-semibold text-acc-500 hover:underline flex items-center gap-1.5 cursor-pointer"
                  @click="lightboxUrl = u.photo_url; photoLightbox = true"
                >
                  <i class="pi pi-camera text-xs" aria-hidden="true" />
                  <span>Lihat Foto Geotag Unit #{{ u.unit_number }}</span>
                </button>
                <div v-if="u.geo_lat != null && u.geo_lng != null" class="text-[10.5px] text-sig-ok font-semibold flex items-center gap-1 t-mono">
                  <i class="pi pi-map-marker text-[10px]" aria-hidden="true" />
                  <span>GPS: {{ u.geo_lat.toFixed(5) }}, {{ u.geo_lng.toFixed(5) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- CONTENT TAB 2: Spesifikasi & Detail Pengadaan -->
        <div v-if="activeTab === 'spec'" class="p-5">
          <div class="grid sm:grid-cols-2 lg:grid-cols-3 gap-3.5 text-[12.5px]">
            <div class="p-3.5 rounded-[8px] border border-line bg-panel-2 flex flex-col gap-1">
              <span class="t-label">No. Kode Barang</span>
              <span class="t-mono font-bold text-[13px] text-acc-500">{{ item.item_code || '—' }}</span>
            </div>

            <div class="p-3.5 rounded-[8px] border border-line bg-panel-2 flex flex-col gap-1">
              <span class="t-label">Jumlah / No. Register</span>
              <span class="t-mono font-bold text-[13px] text-txt">{{ item.register_number || '—' }}</span>
            </div>

            <div class="p-3.5 rounded-[8px] border border-line bg-panel-2 flex flex-col gap-1">
              <span class="t-label">No. Kode Lokasi</span>
              <span class="t-mono font-semibold text-[13px] text-txt">{{ item.location_code || '—' }}</span>
            </div>

            <div class="p-3.5 rounded-[8px] border border-line bg-panel-2 flex flex-col gap-1">
              <span class="t-label">Merk / Pabrikan</span>
              <span class="font-bold text-[13px] text-txt">{{ item.merk || '—' }}</span>
            </div>

            <div class="p-3.5 rounded-[8px] border border-line bg-panel-2 flex flex-col gap-1">
              <span class="t-label">Model / Tipe</span>
              <span class="font-bold text-[13px] text-txt">{{ item.type_model || '—' }}</span>
            </div>

            <div class="p-3.5 rounded-[8px] border border-line bg-panel-2 flex flex-col gap-1">
              <span class="t-label">Nomor Seri Pabrik (SN)</span>
              <span class="t-mono font-bold text-[13px] text-acc-500">{{ item.serial_number || '—' }}</span>
            </div>

            <div class="p-3.5 rounded-[8px] border border-line bg-panel-2 flex flex-col gap-1">
              <span class="t-label">Ukuran / Dimensi</span>
              <span class="font-bold text-[13px] text-txt">{{ item.size || '—' }}</span>
            </div>

            <div class="p-3.5 rounded-[8px] border border-line bg-panel-2 flex flex-col gap-1">
              <span class="t-label">Bahan / Material</span>
              <span class="font-bold text-[13px] text-txt">{{ item.material || '—' }}</span>
            </div>

            <div class="p-3.5 rounded-[8px] border border-line bg-panel-2 flex flex-col gap-1">
              <span class="t-label">Tahun Pembuatan / Pembelian</span>
              <span class="t-num font-bold text-[13px] text-txt">{{ item.procurement_year || '—' }}</span>
            </div>

            <div class="p-3.5 rounded-[8px] border border-line bg-panel-2 flex flex-col gap-1">
              <span class="t-label">Keadaan Barang</span>
              <span class="font-bold text-[13px] text-txt">{{ item.condition_status || '—' }}</span>
            </div>

            <div class="p-3.5 rounded-[8px] border border-line bg-panel-2 flex flex-col gap-1">
              <span class="t-label">Sumber Dana</span>
              <span class="font-semibold text-txt">{{ item.funding_source || '—' }}</span>
            </div>

            <div class="p-3.5 rounded-[8px] border border-line bg-panel-2 flex flex-col gap-1">
              <span class="t-label">Distributor / Rekanan</span>
              <span class="font-semibold text-txt">{{ item.distributor || '—' }}</span>
            </div>

            <div class="p-3.5 rounded-[8px] border border-line bg-panel-2 flex flex-col gap-1 sm:col-span-2 lg:col-span-3">
              <span class="t-label">Izin Edar AKL / AKD</span>
              <span class="t-mono font-semibold text-[13px] text-txt">{{ item.akl_akd || '—' }}</span>
            </div>

            <div v-if="item.description" class="p-3.5 rounded-[8px] border border-line bg-panel-2 flex flex-col gap-1 sm:col-span-2 lg:col-span-3">
              <span class="t-label">Keterangan / Catatan Tambahan</span>
              <p class="leading-relaxed text-[12px] text-txt">{{ item.description }}</p>
            </div>
          </div>
        </div>

        <!-- CONTENT TAB 3: Riwayat Pemeliharaan & Servis -->
        <div v-if="activeTab === 'maintenance'" class="p-5 flex flex-col gap-4">
          <div class="flex items-center justify-between flex-wrap gap-2">
            <div>
              <div class="font-bold text-[13px] text-txt">Jadwal &amp; Log Servis Aset</div>
              <div class="text-[11.5px] text-txt-dim">Catatan perbaikan, servis berkala, dan masa kalibrasi alat</div>
            </div>
            <Button
              label="+ Catat Servis / Kalibrasi"
              icon="pi pi-plus"
              size="small"
              severity="primary"
              @click="showMaintModal = true"
            />
          </div>

          <EmptyState v-if="loadingMaint" loading title="Memuat riwayat pemeliharaan…" />

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
              class="p-4 rounded-[10px] border border-line bg-panel-2 flex flex-col gap-2.5"
            >
              <div class="flex items-start justify-between gap-2 flex-wrap">
                <div class="flex items-center gap-2.5 flex-wrap">
                  <span class="px-2.5 py-0.5 rounded-[6px] text-[11px] font-bold border"
                        :class="m.service_type === 'Kalibrasi'
                          ? 'bg-[var(--sig-info-bg)] text-sig-info border-[var(--sig-info-line)]'
                          : 'bg-[var(--acc-50)] text-acc-500 border-[var(--acc-border)]'">
                    {{ m.service_type }}
                  </span>
                  <span class="t-mono font-bold text-[13px] text-txt">{{ m.service_date }}</span>
                  <span v-if="m.vendor_or_technician" class="text-[12px] text-txt-dim">
                    · Oleh: <strong class="text-txt">{{ m.vendor_or_technician }}</strong>
                  </span>
                </div>
                <div class="flex items-center gap-2">
                  <span v-if="m.cost" class="t-num font-bold text-sig-ok text-[12px]">
                    Rp {{ Number(m.cost).toLocaleString('id-ID') }}
                  </span>
                  <Button icon="pi pi-trash" text rounded size="small" severity="danger" aria-label="Hapus catatan servis" v-tooltip.top="'Hapus'"
                          @click="deleteMaintRecord(m.id)" />
                </div>
              </div>

              <p v-if="m.description" class="text-[12px] leading-relaxed text-txt">
                {{ m.description }}
              </p>

              <div class="flex items-center justify-between text-[11.5px] pt-2 border-t border-line">
                <div v-if="m.next_service_date" class="flex items-center gap-1.5 font-bold text-sig-ok t-mono">
                  <i class="pi pi-calendar text-[11px]" aria-hidden="true" />
                  <span>Jatuh Tempo Berikutnya: {{ m.next_service_date }}</span>
                </div>
                <div v-else></div>

                <button
                  v-if="m.photo_url"
                  type="button"
                  class="text-acc-500 hover:underline flex items-center gap-1 font-semibold cursor-pointer"
                  @click="lightboxUrl = m.photo_url; photoLightbox = true"
                >
                  <i class="pi pi-image text-[11px]" aria-hidden="true" /> Bukti Nota / Sertifikat
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
      :visible="stockModalType !== null"
      @update:visible="(val) => { if (!val) stockModalType = null }"
      modal
      :header="stockModalType === 'IN' ? 'Catat Stok Masuk (Penerimaan Barang)' : 'Catat Stok Keluar (Pemakaian Barang)'"
      :style="{ width: '560px' }"
      class="p-fluid"
    >
      <div class="flex flex-col gap-3.5 text-[12.5px] pt-1">
        <div
          class="p-3 rounded-[8px] border text-[12px] flex items-center justify-between"
          :class="stockModalType === 'IN'
            ? 'bg-[var(--sig-ok-bg)] border-[var(--sig-ok-line)]'
            : 'bg-[var(--sig-bad-bg)] border-[var(--sig-bad-line)]'"
        >
          <div>
            <div class="t-label">Nama Barang</div>
            <div class="font-bold text-[13px] text-txt">{{ item?.name }}</div>
          </div>
          <div class="text-right">
            <div class="t-label">Stok Saat Ini</div>
            <div class="t-num font-bold text-[14px] text-acc-500">{{ item?.current_stock }} {{ item?.unit }}</div>
          </div>
        </div>

        <div class="grid sm:grid-cols-2 gap-3">
          <label class="flex flex-col gap-1">
            <span class="text-[11.5px] font-semibold text-txt-dim">
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
            <span class="text-[11.5px] font-semibold text-txt-dim">
              {{ stockModalType === 'IN' ? 'Petugas / Penerima' : 'Diberikan Kepada / Pemohon' }}
            </span>
            <InputText
              v-model="stockForm.received_by"
              :placeholder="stockModalType === 'IN' ? 'Nama penerima barang' : 'Nama orang / ruangan pemohon'"
              class="w-full !text-[12px]"
            />
          </label>

          <label class="flex flex-col gap-1 sm:col-span-2">
            <span class="text-[11.5px] font-semibold text-txt-dim">
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

        <div class="flex justify-end gap-2 pt-2 border-t border-line">
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

    <!-- Modal Form: Tambah / Edit Unit Fisik (Aset Tetap / BMD Multi-Unit) -->
    <Dialog
      v-model:visible="showUnitModal"
      modal
      :header="editingUnitId ? 'Edit Unit Fisik' : 'Tambah Unit Fisik Baru'"
      :style="{ width: '560px' }"
      class="p-fluid"
    >
      <div class="flex flex-col gap-3.5 text-[12.5px] pt-1">
        <div class="p-3 rounded-[8px] border border-[var(--acc-border)] bg-[var(--acc-wash)] text-[12px] flex items-center justify-between">
          <div>
            <div class="t-label">Nama Barang (Spesifikasi Sama)</div>
            <div class="font-bold text-[13px] text-txt">{{ item?.name }}</div>
          </div>
          <div class="text-right">
            <div class="t-label">No. Kode Barang</div>
            <div class="font-bold text-[13px] t-mono text-acc-500">{{ item?.item_code || '—' }}</div>
          </div>
        </div>

        <div class="grid sm:grid-cols-2 gap-3">
          <label class="flex flex-col gap-1">
            <span class="text-[11.5px] font-semibold text-txt-dim">Nomor Urut Unit *</span>
            <InputNumber v-model="unitForm.unit_number" :min="1" :max="9999" showButtons class="w-full !text-[12px]" />
          </label>

          <label class="flex flex-col gap-1">
            <span class="text-[11.5px] font-semibold text-txt-dim">No. Register *</span>
            <InputText v-model="unitForm.register_number" placeholder="mis. 0001" class="w-full t-mono !text-[12px]" />
          </label>

          <label class="flex flex-col gap-1">
            <span class="text-[11.5px] font-semibold text-txt-dim">Kode Barcode / QR Unit *</span>
            <InputText v-model="unitForm.barcode" placeholder="mis. BARANG-U1" class="w-full t-mono !text-[12px]" />
          </label>

          <label class="flex flex-col gap-1">
            <span class="text-[11.5px] font-semibold text-txt-dim">Kondisi Fisik Unit</span>
            <Select v-model="unitForm.condition_status" :options="conditionOptions" class="w-full !text-[12px]" />
          </label>

          <label class="flex flex-col gap-1">
            <span class="text-[11.5px] font-semibold text-txt-dim">Nomor Seri Pabrik (SN)</span>
            <InputText v-model="unitForm.serial_number" placeholder="Nomor seri pada fisik barang" class="w-full t-mono !text-[12px]" />
          </label>

          <label class="flex flex-col gap-1">
            <span class="text-[11.5px] font-semibold text-txt-dim">Ruangan / Lokasi Unit</span>
            <InputText v-model="unitForm.location" placeholder="mis. Ruang Sanitasi" class="w-full !text-[12px]" />
          </label>

          <label class="flex flex-col gap-1 sm:col-span-2">
            <span class="text-[11.5px] font-semibold text-txt-dim">Catatan Khusus Unit Ini</span>
            <Textarea v-model="unitForm.notes" rows="2" placeholder="Catatan kondisi, penempatan khusus, atau keterangan lain..." class="w-full !text-[12px]" />
          </label>

          <!-- Upload Foto Geotag Unit Fisik -->
          <div class="sm:col-span-2">
            <PhotoUploader
              ref="unitPhotoRef"
              :defer-upload="true"
              v-model="unitForm.photo_url"
              v-model:geo-lat="unitForm.geo_lat"
              v-model:geo-lng="unitForm.geo_lng"
              v-model:geo-acc="unitForm.geo_acc"
              v-model:geo-name="unitForm.geo_name"
              :item-name="item?.name"
              :location-name="unitForm.location || item?.location"
              label="Foto Fisik Unit Berstempel Geotag (GPS Map Camera)"
              hint="Foto fisik unit ini dengan cap GPS Map Camera sebagai identitas otentik"
            />
          </div>
        </div>

        <div class="flex justify-end gap-2 pt-2 border-t border-line">
          <Button label="Batal" size="small" severity="secondary" text @click="showUnitModal = false" />
          <Button
            :label="editingUnitId ? 'Simpan Perubahan' : 'Tambah Unit'"
            icon="pi pi-check"
            size="small"
            severity="primary"
            :loading="submittingUnit"
            @click="submitUnit"
          />
        </div>
      </div>
    </Dialog>

    <!-- Modal Form: Catat Servis & Kalibrasi -->
    <Dialog v-model:visible="showMaintModal" modal header="Catat Pemeliharaan / Kalibrasi Aset" :style="{ width: '560px' }" class="p-fluid">
      <div class="flex flex-col gap-3 text-[12.5px] pt-1">
        <div class="grid sm:grid-cols-2 gap-3">
          <label class="flex flex-col gap-1">
            <span class="text-[11px] font-semibold text-txt-dim">Jenis Tindakan</span>
            <Select v-model="maintForm.service_type" :options="serviceTypeOptions" class="w-full !text-[12px]" />
          </label>
          <label class="flex flex-col gap-1">
            <span class="text-[11px] font-semibold text-txt-dim">Tanggal Servis</span>
            <InputText v-model="maintForm.service_date" type="date" class="w-full !text-[12px]" />
          </label>
          <label class="flex flex-col gap-1">
            <span class="text-[11px] font-semibold text-txt-dim">Teknisi / Vendor</span>
            <InputText v-model="maintForm.vendor_or_technician" placeholder="mis. PT Medika Solusi" class="w-full !text-[12px]" />
          </label>
          <label class="flex flex-col gap-1">
            <span class="text-[11px] font-semibold text-txt-dim">Biaya (Rp)</span>
            <InputNumber v-model="maintForm.cost" :min="0" mode="currency" currency="IDR" locale="id-ID" class="w-full !text-[12px]" />
          </label>
          <label class="flex flex-col gap-1">
            <span class="text-[11px] font-semibold text-txt-dim">Jadwal Servis / Kalibrasi Berikutnya</span>
            <InputText v-model="maintForm.next_service_date" type="date" class="w-full !text-[12px]" />
          </label>
          <label class="flex flex-col gap-1">
            <span class="text-[11px] font-semibold text-txt-dim">Update Kondisi Fisik</span>
            <Select v-model="maintForm.update_condition" :options="conditionOptions" class="w-full !text-[12px]" />
          </label>
          <label class="flex flex-col gap-1 sm:col-span-2">
            <span class="text-[11px] font-semibold text-txt-dim">Keterangan / Hasil Pemeriksaan</span>
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

        <div class="flex justify-end gap-2 pt-2 border-t border-line">
          <Button label="Batal" size="small" severity="secondary" text @click="showMaintModal = false" />
          <Button label="Simpan Catatan Servis" icon="pi pi-check" size="small" :loading="submittingMaint"
                  @click="submitMaintenance" />
        </div>
      </div>
    </Dialog>

    <!-- Modal Lightbox Foto -->
    <Dialog v-model:visible="photoLightbox" modal header="Pratinjau Foto Berstempel Geotag" :style="{ width: '700px' }" class="p-fluid">
      <div v-if="lightboxUrl" class="flex flex-col gap-3">
        <div class="p-2 bg-ink-950 rounded-[8px] border border-line flex items-center justify-center">
          <img :src="lightboxUrl" alt="Foto Geotag" class="max-h-[75vh] object-contain rounded" />
        </div>
        <div class="flex items-center justify-between pt-2 border-t border-line">
          <span class="text-xs text-txt-dim">Berkas gambar berstempel GPS Map Camera</span>
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
