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
import PhotoUploader from '@/components/PhotoUploader.vue'

const route = useRoute()
const router = useRouter()
const confirm = useConfirm()
const toast = useToast()

const item = ref<any>(null)
const loading = ref(true)
const photoLightbox = ref(false)
const lightboxUrl = ref('')

// Stock Movement State (Stok Masuk & Stok Keluar dengan bukti foto geotag)
const stockModalType = ref<'IN' | 'OUT' | null>(null)
const submittingStock = ref(false)
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
  loading.value = true
  try {
    const res = await api.get(`/items/${route.params.id}`)
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
  loadingMaint.value = true
  try {
    const res = await api.get(`/items/${route.params.id}/maintenance`)
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
    stockTransactions.value = res.data.data
  } catch {
    // silent
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
  confirm.require({
    message: `Hapus barang "${item.value.name}" (${item.value.sku})? Seluruh data barang ini akan dihapus.`,
    header: 'Konfirmasi Hapus Barang',
    icon: 'pi pi-exclamation-triangle',
    acceptClass: 'p-button-danger',
    accept: async () => {
      try {
        await api.delete(`/items/${item.value.id}`)
        toast.add({ severity: 'success', summary: 'Barang dihapus', detail: item.value.name, life: 2500 })
        router.push('/items')
      } catch (e: any) {
        toast.add({ severity: 'error', summary: 'Gagal menghapus', detail: e.response?.data?.error, life: 3000 })
      }
    },
  })
}

async function submitMaintenance() {
  submittingMaint.value = true
  try {
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
  <div class="pb-24 max-w-4xl mx-auto">
    <!-- Top Nav Back Button -->
    <div class="flex items-center justify-between mb-4">
      <button
        class="text-[12px] font-semibold text-ink-300 hover:text-ink-100 flex items-center gap-1.5 transition-colors cursor-pointer py-1"
        @click="router.push('/items')"
      >
        <i class="pi pi-arrow-left text-[11px]" />
        <span>Kembali ke Daftar Barang</span>
      </button>

      <div class="flex items-center gap-1.5">
        <Button
          label="Edit"
          icon="pi pi-pencil"
          size="small"
          severity="secondary"
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
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="panel p-16 grid place-items-center">
      <i class="pi pi-spin pi-spinner text-2xl text-acc-500" />
    </div>

    <!-- Main Detail Card -->
    <div v-else-if="item" class="flex flex-col gap-4">
      <!-- Header Banner Card -->
      <div class="panel p-5 rounded-xl border flex flex-col gap-3 shadow-md"
           style="background: var(--paper-1); border-color: var(--line)">
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2 flex-wrap mb-1.5">
              <span class="t-mono text-[12px] font-bold text-acc-500 bg-acc-500/10 px-2 py-0.5 rounded border border-acc-500/20">
                {{ item.sku }}
              </span>
              <span
                class="px-2 py-0.5 rounded text-[11px] font-bold border"
                :class="item.track_stock !== false ? 'bg-amber-500/10 text-amber-300 border-amber-500/30' : 'bg-blue-500/10 text-blue-300 border-blue-500/30'"
              >
                {{ item.track_stock !== false ? 'Barang Stok / Konsumabel' : 'Aset Tetap / Unit Mandiri' }}
              </span>
              <StatusChip :kind="item.condition_status" />
              <Tag v-if="item.is_available" severity="success" value="TERSEDIA" class="!text-[10px]" />
              <Tag v-else severity="danger" value="TIDAK TERSEDIA" class="!text-[10px]" />
            </div>

            <h1 class="text-xl sm:text-2xl font-bold leading-tight tracking-tight break-words">
              {{ item.name }}
            </h1>

            <div class="text-[12px] text-ink-400 mt-1 flex items-center gap-3 flex-wrap">
              <span>Kategori: <strong class="text-ink-200">{{ item.category }}</strong></span>
              <span>·</span>
              <span class="flex items-center gap-1">
                <i class="pi pi-map-marker text-acc-500 text-[10px]" />
                Ruangan: <strong class="text-ink-200">{{ item.location }}</strong>
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- Hero Photo Card with Geotag -->
      <div
        v-if="item.photo_url"
        class="panel rounded-xl overflow-hidden border bg-black/40 flex flex-col shadow-md"
        style="border-color: var(--line)"
      >
        <div
          class="relative w-full max-h-[380px] sm:max-h-[460px] bg-black/70 flex items-center justify-center cursor-pointer group p-2"
          @click="lightboxUrl = item.photo_url; photoLightbox = true"
        >
          <img
            :src="item.photo_url"
            :alt="item.name"
            class="max-h-[360px] sm:max-h-[440px] w-full object-contain rounded group-hover:scale-[1.01] transition-transform"
          />
          <div class="absolute bottom-3 left-3 bg-black/85 px-2.5 py-1 rounded text-[11px] text-white flex items-center gap-1.5 border border-white/10">
            <i class="pi pi-map-marker text-sig-ok text-[10px]" /> Foto Berstempel GPS Map Camera
          </div>
          <div class="absolute top-3 right-3 bg-black/80 px-2.5 py-1 rounded text-[11px] text-white flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
            <i class="pi pi-search-plus text-[10px]" /> Perbesar
          </div>
        </div>

        <div v-if="hasCoords" class="px-4 py-2.5 border-t flex items-center justify-between text-[11.5px]"
             style="border-color: var(--line); background: var(--paper-2)">
          <div class="t-mono text-sig-ok font-semibold flex items-center gap-1.5">
            <i class="pi pi-compass text-[11px]" />
            <span>GPS: {{ item.geo_lat.toFixed(5) }}, {{ item.geo_lng.toFixed(5) }}</span>
            <span v-if="item.geo_acc" class="text-ink-400 font-normal">(±{{ Math.round(item.geo_acc) }}m)</span>
          </div>
          <a
            :href="googleMapsUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="text-acc-500 hover:underline flex items-center gap-1 font-semibold"
          >
            <i class="pi pi-external-link text-[10px]" /> Buka di Google Maps
          </a>
        </div>
      </div>

      <!-- Quick Metrics Hero Card & Dedicated Stock Movement Action -->
      <div v-if="item.track_stock !== false" class="panel p-4 rounded-xl border flex flex-col gap-4 shadow-md"
           style="background: var(--paper-1); border-color: var(--line)">
        <div class="flex flex-wrap items-center justify-between gap-3 pb-3 border-b" style="border-color: var(--line)">
          <div>
            <div class="text-[11px] font-bold text-ink-400 uppercase tracking-wider">Kondisi Stok Barang</div>
            <div class="flex items-baseline gap-2 mt-1">
              <span class="t-num text-3xl sm:text-4xl font-black"
                    :class="item.current_stock <= item.min_stock ? 'text-rose-400' : 'text-emerald-400'">
                {{ item.current_stock }}
              </span>
              <span class="text-sm font-semibold text-ink-300">{{ item.unit }}</span>
              <span v-if="item.current_stock <= item.min_stock" class="text-[11px] font-bold text-rose-400 bg-rose-500/10 px-2 py-0.5 rounded ml-2">
                Stok Menipis (min {{ item.min_stock }})
              </span>
            </div>
          </div>

          <!-- Prominent Stock Action Buttons -->
          <div class="flex items-center gap-2 flex-wrap">
            <Button
              label="+ Tambah Stok Masuk"
              icon="pi pi-arrow-down"
              size="small"
              severity="success"
              @click="openStockModal('IN')"
            />
            <Button
              label="− Catat Stok Keluar"
              icon="pi pi-arrow-up"
              size="small"
              severity="danger"
              :disabled="item.current_stock <= 0"
              @click="openStockModal('OUT')"
            />
          </div>
        </div>

        <div class="grid grid-cols-2 gap-3 text-center text-[12px]">
          <div class="p-2.5 rounded-lg border bg-paper-2 border-line">
            <div class="text-[11px] text-ink-400">Batas Minimum Peringatan</div>
            <div class="t-num font-bold text-[15px] mt-0.5">{{ item.min_stock }} {{ item.unit }}</div>
          </div>
          <div class="p-2.5 rounded-lg border bg-paper-2 border-line">
            <div class="text-[11px] text-ink-400">Estimasi Nilai per Unit</div>
            <div class="t-num font-bold text-[15px] text-acc-400 mt-0.5">
              {{ item.price_per_unit ? 'Rp ' + Number(item.price_per_unit).toLocaleString('id-ID') : '—' }}
            </div>
          </div>
        </div>
      </div>

      <!-- Fixed Asset Card (Non-Consumable / Laptop) -->
      <div v-else class="panel p-4 rounded-xl border flex items-center justify-between gap-3 shadow-md"
           style="background: var(--paper-1); border-color: var(--line)">
        <div>
          <div class="text-[11px] font-bold text-ink-400 uppercase tracking-wider">Status Aset Tetap</div>
          <div class="text-[14px] font-bold mt-0.5 text-ink-100 flex items-center gap-2">
            <i class="pi pi-desktop text-acc-500" />
            <span>Unit Mandiri (Stok Tetap: 1 {{ item.unit }})</span>
          </div>
          <div class="text-[11.5px] text-ink-400 mt-0.5">
            Unit aset ini teridentifikasi unik berdasarkan nomor seri dan lokasi ruangan.
          </div>
        </div>

        <Button
          label="Catat Servis"
          icon="pi pi-wrench"
          size="small"
          severity="warn"
          @click="showMaintModal = true"
        />
      </div>

      <!-- Segment Tabs: Riwayat Mutasi vs Spesifikasi vs Riwayat Servis -->
      <div class="panel rounded-xl overflow-hidden border shadow-md"
           style="background: var(--paper-1); border-color: var(--line)">
        <div class="flex border-b text-[12.5px] font-bold overflow-x-auto" style="border-color: var(--line); background: var(--paper-2)">
          <!-- Tab 1: Riwayat Mutasi Stok (Hanya barang stok) -->
          <button
            v-if="item.track_stock !== false"
            class="py-3 px-4 flex items-center justify-center gap-2 transition-colors cursor-pointer border-b-2 whitespace-nowrap"
            :class="activeTab === 'stock_history' ? 'border-acc-500 text-acc-500 bg-paper-1' : 'border-transparent text-ink-400 hover:text-ink-200'"
            @click="activeTab = 'stock_history'"
          >
            <i class="pi pi-history text-[12px]" />
            <span>Riwayat Mutasi Stok</span>
            <span v-if="stockTransactions.length" class="text-[10px] px-1.5 py-0.2 rounded-full bg-acc-500 text-ink-950 font-bold">
              {{ stockTransactions.length }}
            </span>
          </button>

          <!-- Tab 2: Spesifikasi Lengkap -->
          <button
            class="py-3 px-4 flex items-center justify-center gap-2 transition-colors cursor-pointer border-b-2 whitespace-nowrap"
            :class="activeTab === 'spec' ? 'border-acc-500 text-acc-500 bg-paper-1' : 'border-transparent text-ink-400 hover:text-ink-200'"
            @click="activeTab = 'spec'"
          >
            <i class="pi pi-list text-[12px]" />
            <span>Spesifikasi Lengkap</span>
          </button>

          <!-- Tab 3: Riwayat Servis & Kalibrasi -->
          <button
            class="py-3 px-4 flex items-center justify-center gap-2 transition-colors cursor-pointer border-b-2 whitespace-nowrap"
            :class="activeTab === 'maintenance' ? 'border-acc-500 text-acc-500 bg-paper-1' : 'border-transparent text-ink-400 hover:text-ink-200'"
            @click="activeTab = 'maintenance'"
          >
            <i class="pi pi-wrench text-[12px]" />
            <span>Riwayat Servis &amp; Kalibrasi</span>
            <span v-if="maintenanceList.length" class="text-[10px] px-1.5 py-0.2 rounded-full bg-acc-500 text-ink-950 font-bold">
              {{ maintenanceList.length }}
            </span>
          </button>
        </div>

        <!-- Content Tab 1: Riwayat Mutasi Stok Lengkap dengan Foto Geotag -->
        <div v-if="activeTab === 'stock_history'" class="p-5 flex flex-col gap-4">
          <div class="flex items-center justify-between">
            <span class="font-bold text-[13px]">Jejak Keluar / Masuk Stok</span>
            <div class="flex gap-2">
              <Button label="+ Masuk" size="small" severity="success" text @click="openStockModal('IN')" />
              <Button label="− Keluar" size="small" severity="danger" text :disabled="item.current_stock <= 0" @click="openStockModal('OUT')" />
            </div>
          </div>

          <div v-if="loadingStockHistory" class="py-12 text-center text-ink-400">
            <i class="pi pi-spin pi-spinner text-xl text-acc-500" />
          </div>

          <div v-else-if="!stockTransactions.length" class="text-center py-10 border rounded-lg text-ink-400 flex flex-col items-center gap-2"
               style="border-color: var(--line); background: var(--paper-2)">
            <i class="pi pi-history text-3xl text-ink-500" />
            <span class="text-[12.5px]">Belum ada catatan mutasi stok untuk barang ini.</span>
            <Button label="Catat Stok Masuk Pertama" icon="pi pi-plus" size="small" outlined class="mt-1" @click="openStockModal('IN')" />
          </div>

          <div v-else class="flex flex-col gap-3">
            <div
              v-for="tx in stockTransactions"
              :key="tx.id"
              class="p-4 rounded-xl border flex flex-col gap-2.5 shadow-xs transition-colors hover:border-acc-500/40"
              style="background: var(--paper-2); border-color: var(--line)"
            >
              <div class="flex items-start justify-between gap-3 flex-wrap">
                <div class="flex items-center gap-2 flex-wrap">
                  <span
                    class="px-2 py-0.5 rounded text-[11px] font-bold border"
                    :class="tx.type === 'IN' || tx.type === 'ADJUST+' ? 'bg-emerald-500/15 text-emerald-400 border-emerald-500/30' : 'bg-rose-500/15 text-rose-400 border-rose-500/30'"
                  >
                    {{ tx.type === 'IN' ? 'Stok Masuk' : tx.type === 'OUT' ? 'Stok Keluar' : 'Penyesuaian' }}
                  </span>
                  <span class="font-bold text-[14px]"
                        :class="tx.type === 'IN' || tx.type === 'ADJUST+' ? 'text-emerald-400' : 'text-rose-400'">
                    {{ tx.type === 'IN' || tx.type === 'ADJUST+' ? '+' : '−' }}{{ tx.quantity }} {{ tx.unit }}
                  </span>
                  <span class="text-ink-400 text-[11.5px] t-mono">({{ tx.previous_stock }} ➔ {{ tx.new_stock }})</span>
                </div>
                <span class="text-[11.5px] text-ink-400 font-medium">{{ tx.timestamp }}</span>
              </div>

              <div class="flex items-center gap-3 text-[12px] text-ink-300 flex-wrap">
                <span v-if="tx.received_by">
                  Petugas / Penerima: <strong class="text-ink-100">{{ tx.received_by }}</strong>
                </span>
                <span v-if="tx.notes" class="text-ink-400 italic">
                  "{{ tx.notes }}"
                </span>
              </div>

              <!-- Foto Bukti Geotag Transaksi -->
              <div v-if="tx.photo_url" class="mt-1 pt-2 border-t border-line-soft flex items-center justify-between">
                <button
                  class="text-[11px] text-acc-500 hover:underline flex items-center gap-1.5 cursor-pointer font-medium"
                  @click="lightboxUrl = tx.photo_url; photoLightbox = true"
                >
                  <i class="pi pi-camera text-xs" />
                  <span>Lihat Foto Bukti Berstempel Geotag</span>
                </button>
                <span class="text-[10px] text-sig-ok flex items-center gap-1">
                  <i class="pi pi-check text-[9px]" /> Bukti GPS Tervalidasi
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- Content Tab 2: Spesifikasi Lengkap -->
        <div v-if="activeTab === 'spec'" class="p-5">
          <div class="grid sm:grid-cols-2 gap-4 text-[12.5px]">
            <div class="p-3 rounded-lg border flex flex-col gap-1" style="background: var(--paper-2); border-color: var(--line)">
              <span class="text-[11px] font-semibold text-ink-400">Merk / Pabrikan</span>
              <span class="font-bold text-[13px]">{{ item.merk || '—' }}</span>
            </div>

            <div class="p-3 rounded-lg border flex flex-col gap-1" style="background: var(--paper-2); border-color: var(--line)">
              <span class="text-[11px] font-semibold text-ink-400">Model / Tipe</span>
              <span class="font-bold text-[13px]">{{ item.type_model || '—' }}</span>
            </div>

            <div class="p-3 rounded-lg border flex flex-col gap-1" style="background: var(--paper-2); border-color: var(--line)">
              <span class="text-[11px] font-semibold text-ink-400">Nomor Seri</span>
              <span class="t-mono font-bold text-[13px]">{{ item.serial_number || '—' }}</span>
            </div>

            <div class="p-3 rounded-lg border flex flex-col gap-1" style="background: var(--paper-2); border-color: var(--line)">
              <span class="text-[11px] font-semibold text-ink-400">Tahun Pengadaan</span>
              <span class="font-bold text-[13px]">{{ item.procurement_year || '—' }}</span>
            </div>

            <div class="p-3 rounded-lg border flex flex-col gap-1" style="background: var(--paper-2); border-color: var(--line)">
              <span class="text-[11px] font-semibold text-ink-400">Sumber Dana</span>
              <span class="font-semibold">{{ item.funding_source || '—' }}</span>
            </div>

            <div class="p-3 rounded-lg border flex flex-col gap-1" style="background: var(--paper-2); border-color: var(--line)">
              <span class="text-[11px] font-semibold text-ink-400">Distributor / Vendor</span>
              <span class="font-semibold">{{ item.distributor || '—' }}</span>
            </div>

            <div class="p-3 rounded-lg border flex flex-col gap-1 sm:col-span-2" style="background: var(--paper-2); border-color: var(--line)">
              <span class="text-[11px] font-semibold text-ink-400">Izin Edar AKL / AKD</span>
              <span class="t-mono font-semibold text-[13px]">{{ item.akl_akd || '—' }}</span>
            </div>

            <div v-if="item.description" class="p-3 rounded-lg border flex flex-col gap-1 sm:col-span-2"
                 style="background: var(--paper-2); border-color: var(--line)">
              <span class="text-[11px] font-semibold text-ink-400">Keterangan Tambahan</span>
              <p class="text-ink-200 leading-relaxed">{{ item.description }}</p>
            </div>
          </div>
        </div>

        <!-- Content Tab 3: Riwayat Servis & Kalibrasi -->
        <div v-if="activeTab === 'maintenance'" class="p-5 flex flex-col gap-4">
          <div class="flex items-center justify-between">
            <span class="font-bold text-[13px]">Catatan Pemeliharaan Aset</span>
            <Button
              label="+ Catat Servis / Kalibrasi"
              icon="pi pi-plus"
              size="small"
              severity="primary"
              @click="showMaintModal = true"
            />
          </div>

          <div v-if="loadingMaint" class="py-12 text-center text-ink-400">
            <i class="pi pi-spin pi-spinner text-xl text-acc-500" />
          </div>

          <div v-else-if="!maintenanceList.length" class="text-center py-10 border rounded-lg text-ink-400 flex flex-col items-center gap-2"
               style="border-color: var(--line); background: var(--paper-2)">
            <i class="pi pi-wrench text-3xl text-ink-500" />
            <span class="text-[12.5px]">Belum ada riwayat servis atau kalibrasi untuk barang ini.</span>
            <Button label="Catat Sekarang" size="small" outlined class="mt-1" @click="showMaintModal = true" />
          </div>

          <div v-else class="flex flex-col gap-3">
            <div
              v-for="m in maintenanceList"
              :key="m.id"
              class="p-4 rounded-xl border flex flex-col gap-2 shadow-xs"
              style="background: var(--paper-2); border-color: var(--line)"
            >
              <div class="flex items-start justify-between gap-2 flex-wrap">
                <div class="flex items-center gap-2 flex-wrap">
                  <span class="px-2 py-0.5 rounded text-[11px] font-bold"
                        :class="m.service_type === 'Kalibrasi' ? 'bg-purple-900/60 text-purple-300 border border-purple-700/50' : 'bg-acc-500/20 text-acc-400 border border-acc-500/40'">
                    {{ m.service_type }}
                  </span>
                  <span class="font-bold text-[13px]">{{ m.service_date }}</span>
                  <span v-if="m.vendor_or_technician" class="text-ink-400 text-[12px]">· Oleh: {{ m.vendor_or_technician }}</span>
                </div>
                <div class="flex items-center gap-2">
                  <span v-if="m.cost" class="t-num font-bold text-emerald-400 text-[12px]">
                    Rp {{ Number(m.cost).toLocaleString('id-ID') }}
                  </span>
                  <Button icon="pi pi-trash" text rounded size="small" severity="danger" v-tooltip.top="'Hapus'"
                          @click="deleteMaintRecord(m.id)" />
                </div>
              </div>

              <p v-if="m.description" class="text-ink-200 text-[12px] leading-relaxed">
                {{ m.description }}
              </p>

              <div class="flex items-center justify-between text-[11px] pt-1.5 border-t" style="border-color: var(--line-soft)">
                <div v-if="m.next_service_date" class="flex items-center gap-1.5 font-bold text-sig-ok">
                  <i class="pi pi-calendar text-[10px]" />
                  <span>Jatuh Tempo Berikutnya: {{ m.next_service_date }}</span>
                </div>
                <div v-else></div>

                <button
                  v-if="m.photo_url"
                  class="text-acc-500 hover:underline flex items-center gap-1 font-semibold cursor-pointer"
                  @click="lightboxUrl = m.photo_url; photoLightbox = true"
                >
                  <i class="pi pi-image text-[10px]" /> Bukti Nota
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Sticky Bottom Mobile Action Bar -->
    <div class="fixed bottom-0 inset-x-0 z-30 p-3 border-t bg-[var(--paper-1)]/95 backdrop-blur-md shadow-2xl flex items-center justify-between gap-2 max-w-4xl mx-auto"
         style="border-color: var(--line)">
      <div class="flex items-center gap-2 flex-wrap">
        <Button
          v-if="item?.track_stock !== false"
          label="+ Stok Masuk"
          icon="pi pi-arrow-down"
          size="small"
          severity="success"
          @click="openStockModal('IN')"
        />
        <Button
          v-if="item?.track_stock !== false"
          label="− Stok Keluar"
          icon="pi pi-arrow-up"
          size="small"
          severity="danger"
          :disabled="!item || item.current_stock <= 0"
          @click="openStockModal('OUT')"
        />
        <Button
          label="Edit"
          icon="pi pi-pencil"
          size="small"
          severity="secondary"
          @click="router.push('/items/' + route.params.id + '/edit')"
        />
        <Button
          label="Label QR"
          icon="pi pi-qrcode"
          size="small"
          text
          severity="secondary"
          @click="router.push('/barcode')"
        />
      </div>

      <Button
        icon="pi pi-trash"
        size="small"
        text
        severity="danger"
        v-tooltip.top="'Hapus Barang'"
        @click="remove"
      />
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
            <div class="text-[11px] text-ink-400">Barang:</div>
            <div class="font-bold text-[13px]">{{ item?.name }}</div>
          </div>
          <div class="text-right">
            <div class="text-[11px] text-ink-400">Stok Saat Ini:</div>
            <div class="font-bold text-[14px] text-acc-400">{{ item?.current_stock }} {{ item?.unit }}</div>
          </div>
        </div>

        <div class="grid sm:grid-cols-2 gap-3">
          <label class="flex flex-col gap-1">
            <span class="text-[11.5px] font-semibold text-ink-300">
              Jumlah {{ stockModalType === 'IN' ? 'Masuk' : 'Keluar' }} ({{ item?.unit }}) *
            </span>
            <InputNumber
              v-model="stockForm.quantity"
              :min="1"
              :max="stockModalType === 'OUT' ? item?.current_stock : 999999"
              showButtons
              class="w-full"
            />
          </label>

          <label class="flex flex-col gap-1">
            <span class="text-[11.5px] font-semibold text-ink-300">
              {{ stockModalType === 'IN' ? 'Petugas / Penerima' : 'Diberikan Kepada / Pemohon' }}
            </span>
            <InputText
              v-model="stockForm.received_by"
              :placeholder="stockModalType === 'IN' ? 'Nama penerima barang' : 'Nama orang / ruangan pemakai'"
              class="w-full !text-[12px]"
            />
          </label>

          <label class="flex flex-col gap-1 sm:col-span-2">
            <span class="text-[11.5px] font-semibold text-ink-300">
              {{ stockModalType === 'IN' ? 'Sumber / Catatan Pengadaan' : 'Keperluan / Catatan Pemakaian' }}
            </span>
            <Textarea
              v-model="stockForm.notes"
              rows="2"
              autoResize
              :placeholder="stockModalType === 'IN' ? 'Nomor invoice, sumber dana, atau keterangan tambahan' : 'Keperluan kegiatan, peruntukan ruangan, dll.'"
              class="w-full !text-[12px]"
            />
          </label>

          <!-- Upload Foto Geotag Bukti Serah Terima -->
          <div class="sm:col-span-2">
            <PhotoUploader
              v-model="stockForm.photo_url"
              v-model:geo-lat="stockForm.geo_lat"
              v-model:geo-lng="stockForm.geo_lng"
              v-model:geo-acc="stockForm.geo_acc"
              v-model:geo-name="stockForm.geo_name"
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
            :icon="stockModalType === 'IN' ? 'pi pi-check' : 'pi pi-check'"
            size="small"
            :severity="stockModalType === 'IN' ? 'success' : 'danger'"
            :loading="submittingStock"
            @click="submitStockMovement"
          />
        </div>
      </div>
    </Dialog>

    <!-- Modal Form: Catat Servis & Kalibrasi -->
    <Dialog v-model:visible="showMaintModal" modal header="Catat Servis / Kalibrasi Aset" :style="{ width: '560px' }" class="p-fluid">
      <div class="flex flex-col gap-3 text-[12.5px] pt-1">
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
            <span class="text-[11px] font-semibold text-ink-400">Teknisi / Vendor</span>
            <InputText v-model="maintForm.vendor_or_technician" placeholder="mis. PT Medika Solusi" class="w-full !text-[12px]" />
          </label>
          <label class="flex flex-col gap-1">
            <span class="text-[11px] font-semibold text-ink-400">Biaya (Rp)</span>
            <InputNumber v-model="maintForm.cost" :min="0" mode="currency" currency="IDR" locale="id-ID" class="w-full !text-[12px]" />
          </label>
          <label class="flex flex-col gap-1">
            <span class="text-[11px] font-semibold text-ink-400">Jadwal Kalibrasi Berikutnya</span>
            <InputText v-model="maintForm.next_service_date" type="date" class="w-full !text-[12px]" />
          </label>
          <label class="flex flex-col gap-1">
            <span class="text-[11px] font-semibold text-ink-400">Update Kondisi Barang</span>
            <Select v-model="maintForm.update_condition" :options="conditionOptions" class="w-full !text-[12px]" />
          </label>
          <label class="flex flex-col gap-1 sm:col-span-2">
            <span class="text-[11px] font-semibold text-ink-400">Keterangan / Hasil Pemeriksaan</span>
            <Textarea v-model="maintForm.description" rows="2" placeholder="Nomor sertifikat kalibrasi atau rincian perbaikan" class="w-full !text-[12px]" />
          </label>
          <div class="sm:col-span-2">
            <PhotoUploader
              v-model="maintForm.photo_url"
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
    <Dialog v-model:visible="photoLightbox" modal :style="{ width: '640px' }" class="p-fluid">
      <div v-if="lightboxUrl" class="p-2 bg-black rounded-lg flex items-center justify-center">
        <img :src="lightboxUrl" alt="Foto Geotag" class="max-h-[75vh] object-contain rounded" />
      </div>
    </Dialog>
  </div>
</template>
