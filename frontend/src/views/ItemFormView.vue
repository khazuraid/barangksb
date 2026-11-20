<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '@/api'
import { useToast } from 'primevue/usetoast'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import InputNumber from 'primevue/inputnumber'
import Select from 'primevue/select'
import Textarea from 'primevue/textarea'
import Tag from 'primevue/tag'
import ToggleSwitch from 'primevue/toggleswitch'
import PageHeader from '@/components/PageHeader.vue'
import Panel from '@/components/Panel.vue'
import Field from '@/components/Field.vue'
import PhotoUploader from '@/components/PhotoUploader.vue'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const isEdit = !!route.params.id
const loading = ref(false)
const loadingData = ref(true)
const photoUploaderRef = ref<any>(null)

const enableKIR = ref(false)
const enableProcurement = ref(false)
const enablePrice = ref(false)
const copiedFromSource = ref<any>(null)

function resetToEmptyForm() {
  copiedFromSource.value = null
  form.value = {
    sku: '',
    name: '',
    category: '',
    location: '',
    location_code: '',
    track_stock: true,
    unit: 'buah',
    current_stock: 0,
    min_stock: 0,
    price_per_unit: 0,
    description: '',
    merk: '',
    type_model: '',
    serial_number: '',
    procurement_year: '',
    condition_status: 'Baik (B)',
    size: '',
    material: '',
    item_code: '',
    register_number: '',
    funding_source: '',
    distributor: '',
    akl_akd: '',
    photo_url: '',
    geo_lat: null,
    geo_lng: null,
    geo_acc: null,
    geo_name: '',
  }
  enableKIR.value = false
  enableProcurement.value = false
  enablePrice.value = false
  router.replace({ path: '/items/new' })
  toast.add({
    severity: 'info',
    summary: 'Mode Salin Dibatalkan',
    detail: 'Formulir dikembalikan ke form baru kosong (tidak tambah unit).',
    life: 2500,
  })
}

const form = ref<any>({
  sku: '',
  name: '',
  category: '',
  location: '',
  location_code: '',
  track_stock: true, // true = barang konsumabel/pensil, false = aset tetap/laptop
  unit: 'buah',
  current_stock: 0,
  min_stock: 0,
  price_per_unit: 0,
  description: '',
  merk: '',
  type_model: '',
  serial_number: '',
  procurement_year: '',
  condition_status: 'Baik (B)',
  size: '',
  material: '',
  item_code: '',
  register_number: '',
  funding_source: '',
  distributor: '',
  akl_akd: '',
  photo_url: '',
  geo_lat: null,
  geo_lng: null,
  geo_acc: null,
  geo_name: '',
})

const categories = ref<any[]>([])
const locations = ref<any[]>([])
const unitOptions = ['buah', 'unit', 'set', 'box', 'rim', 'pack', 'dus', 'botol', 'roll', 'lembar']
const conditionOptions = [
  'Baik (B)',
  'Kurang Baik (KB)',
  'Rusak Berat (RB)',
  'Berfungsi',
  'Rusak Ringan',
  'Perlu Kalibrasi',
]

const selectedLocationCode = computed(() => {
  const found = locations.value.find((l: any) => l.name === form.value.location)
  return found?.code || ''
})

function syncLocationCodeFromRoom(locName: string) {
  const found = locations.value.find((l: any) => l.name === locName)
  if (found) {
    form.value.location_code = found.code || ''
  }
}

function onLocationSelect(locName: string) {
  form.value.location = locName
  syncLocationCodeFromRoom(locName)
}

onMounted(async () => {
  try {
    const [c, l] = await Promise.all([api.get('/categories'), api.get('/locations')])
    categories.value = c.data
    locations.value = l.data
    if (isEdit) {
      const res = await api.get(`/items/${route.params.id}`)
      Object.assign(form.value, res.data)
      form.value.sku = res.data.sku ?? ''
      form.value.track_stock = res.data.track_stock !== false
      if (!form.value.track_stock && (Number(form.value.current_stock) || 0) < 1) {
        form.value.current_stock = 1
      }
      if (selectedLocationCode.value) {
        form.value.location_code = selectedLocationCode.value
      }

      // Auto-aktifkan opsi KIR / KIB jika data aset memiliki atribut KIR
      if (
        res.data.item_code ||
        res.data.register_number ||
        res.data.location_code ||
        res.data.merk ||
        res.data.type_model ||
        res.data.serial_number ||
        res.data.size ||
        res.data.material ||
        res.data.procurement_year ||
        (res.data.condition_status && !['Berfungsi', 'Baik (B)'].includes(res.data.condition_status))
      ) {
        enableKIR.value = true
      }

      // Auto-aktifkan opsi pengadaan jika ada data pengadaan atau keterangan
      if (
        res.data.funding_source ||
        res.data.distributor ||
        res.data.akl_akd ||
        res.data.description
      ) {
        enableProcurement.value = true
      }

      // Auto-aktifkan opsi harga jika barang memiliki nilai harga satuan
      if (res.data.price_per_unit && res.data.price_per_unit > 0) {
        enablePrice.value = true
      }
    } else if (route.query.copy_from) {
      try {
        const res = await api.get(`/items/${route.query.copy_from}`)
        const src = res.data
        copiedFromSource.value = src

        // Salin spesifikasi pokok & No. Kode Barang
        form.value.name = src.name ?? ''
        form.value.category = src.category ?? ''
        form.value.location = src.location ?? ''
        form.value.location_code = src.location_code ?? ''
        if (selectedLocationCode.value) {
          form.value.location_code = selectedLocationCode.value
        }
        form.value.track_stock = false // Selalu unit mandiri / aset tetap
        form.value.unit = src.unit || 'unit'
        form.value.current_stock = 1
        form.value.min_stock = 0
        form.value.price_per_unit = src.price_per_unit ?? 0
        form.value.description = src.description ?? ''
        form.value.merk = src.merk ?? ''
        form.value.type_model = src.type_model ?? ''
        form.value.procurement_year = src.procurement_year ?? ''
        form.value.condition_status = src.condition_status || 'Baik (B)'
        form.value.size = src.size ?? ''
        form.value.material = src.material ?? ''
        form.value.item_code = src.item_code ?? '' // No. Kode Barang SAMA PERSIS
        form.value.funding_source = src.funding_source ?? ''
        form.value.distributor = src.distributor ?? ''
        form.value.akl_akd = src.akl_akd ?? ''

        // Auto-increment No. Register jika berupa angka
        if (src.register_number && /^\d+$/.test(src.register_number)) {
          const nextNum = parseInt(src.register_number, 10) + 1
          form.value.register_number = String(nextNum).padStart(src.register_number.length, '0')
        } else {
          form.value.register_number = ''
        }

        // Field fisik unik dikosongkan agar pengguna input data unit ke-2
        form.value.sku = ''
        form.value.serial_number = ''
        form.value.photo_url = ''
        form.value.geo_lat = null
        form.value.geo_lng = null
        form.value.geo_acc = null
        form.value.geo_name = ''

        if (
          src.item_code ||
          src.register_number ||
          src.location_code ||
          src.merk ||
          src.type_model ||
          src.size ||
          src.material ||
          src.procurement_year
        ) {
          enableKIR.value = true
        }
        if (src.funding_source || src.distributor || src.akl_akd || src.description) {
          enableProcurement.value = true
        }
        if (src.price_per_unit && src.price_per_unit > 0) {
          enablePrice.value = true
        }
      } catch {
        toast.add({ severity: 'warn', summary: 'Gagal memuat barang sumber salin', life: 3000 })
      }
    }
  } catch {
    toast.add({ severity: 'error', summary: 'Gagal memuat data barang', life: 3500 })
  } finally {
    loadingData.value = false
  }
})

const skuPreview = computed(() => {
  if (form.value.sku) return form.value.sku
  const cat = categories.value.find((c: any) => c.name === form.value.category)
  if (!cat) return '— pilih kategori dulu —'
  const prefix = (cat.id || '').split('-')[0].slice(0, 4).toUpperCase()
  return `${prefix}-${new Date().getFullYear()}-###`
})

const dirty = computed(() => !!form.value.name && !!form.value.category && !!form.value.location)

async function submit() {
  if (!dirty.value) {
    toast.add({ severity: 'warn', summary: 'Field wajib belum diisi', detail: 'Nama, kategori, dan ruangan wajib diisi', life: 3000 })
    return
  }
  if (!form.value.track_stock) {
    form.value.unit = 'unit'
  }

  // Jika opsi dinonaktifkan, bersihkan data agar tidak tersimpan sembarangan
  if (!enableKIR.value) {
    form.value.item_code = ''
    form.value.register_number = ''
    if (!selectedLocationCode.value) {
      form.value.location_code = ''
    }
    form.value.merk = ''
    form.value.type_model = ''
    form.value.serial_number = ''
    form.value.size = ''
    form.value.material = ''
    form.value.procurement_year = ''
  }
  if (selectedLocationCode.value) {
    form.value.location_code = selectedLocationCode.value
  }
  if (!enableProcurement.value) {
    form.value.funding_source = ''
    form.value.distributor = ''
    form.value.akl_akd = ''
    form.value.description = ''
  }
  if (!enablePrice.value) {
    form.value.price_per_unit = 0
  }

  loading.value = true
  try {
    // Unggah foto fisik hanya saat pengguna benar-benar menyimpan formulir (hemat storage)
    if (photoUploaderRef.value?.hasPendingPhoto) {
      const uploadedUrl = await photoUploaderRef.value.uploadPending()
      if (uploadedUrl) {
        form.value.photo_url = uploadedUrl
      }
    }

    if (isEdit) {
      const payload: any = { ...form.value }
      if (!form.value.track_stock) {
        payload.current_stock = Math.max(1, Number(form.value.current_stock) || 1)
      }
      await api.put(`/items/${route.params.id}`, payload)
      toast.add({ severity: 'success', summary: 'Perubahan disimpan', detail: form.value.name, life: 2500 })
      router.push('/items/' + route.params.id)
    } else {
      const payload: any = { ...form.value }
      if (!form.value.track_stock) {
        const qty = Math.max(1, Number(form.value.current_stock) || 1)
        payload.initial_units = qty
        payload.current_stock = qty
      }
      const res = await api.post('/items', payload)
      toast.add({ severity: 'success', summary: 'Barang berhasil didaftarkan', detail: `SKU ${res.data.sku}`, life: 3000 })
      router.push('/items/' + res.data.id)
    }
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Gagal menyimpan', detail: e.response?.data?.error || e.message, life: 4000 })
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="pb-20 w-full">
    <!-- Clean Page Header -->
    <PageHeader
      :crumb="isEdit ? 'Operasional · Edit Data' : 'Operasional · Registrasi Baru'"
      :title="isEdit ? 'Edit Data Barang' : 'Pendaftaran Barang Baru'"
      :sub="isEdit ? 'Perbarui identitas, ruangan penempatan, dan spesifikasi aset inventaris' : 'Daftarkan aset atau barang konsumabel dengan bukti foto geotag GPS Map Camera'"
    >
      <template #actions>
        <Button
          label="Batal"
          icon="pi pi-times"
          size="small"
          severity="secondary"
          text
          @click="router.push(isEdit ? '/items/' + route.params.id : (copiedFromSource ? '/items/' + copiedFromSource.id : '/items'))"
        />
        <Button
          :label="isEdit ? 'Simpan Perubahan' : (copiedFromSource ? 'Simpan Unit Baru' : 'Simpan Barang')"
          icon="pi pi-check"
          size="small"
          :loading="loading"
          :disabled="!dirty || loadingData"
          @click="submit"
        />
      </template>
    </PageHeader>

    <!-- Banner Pilihan Salin Unit Sejenis & Opsi Batal/Tidak Tambah Unit -->
    <div
      v-if="!loadingData && copiedFromSource"
      class="panel p-4 mb-4 flex flex-wrap items-center justify-between gap-3 border-l-4 !border-l-acc-500 bg-[var(--acc-wash)]"
    >
      <div class="flex items-center gap-3">
        <div class="w-9 h-9 rounded-[8px] bg-acc-500 text-white grid place-items-center shrink-0">
          <i class="pi pi-clone text-base" aria-hidden="true" />
        </div>
        <div class="text-[12.5px]">
          <div class="font-bold text-txt flex items-center gap-2">
            <span>Mode: Tambah Unit Sejenis</span>
            <Tag severity="info" class="!text-[10px]" value="SALIN SPESIFIKASI" />
          </div>
          <div class="text-txt-dim mt-0.5">
            Spesifikasi disalin dari: <strong class="text-acc-500">{{ copiedFromSource.name }}</strong>
            <span v-if="copiedFromSource.item_code" class="t-mono font-semibold"> (No. Kode: {{ copiedFromSource.item_code }})</span>
          </div>
        </div>
      </div>
      <div class="flex items-center gap-2">
        <Button
          label="Tidak Tambah Unit (Form Kosong)"
          icon="pi pi-times"
          size="small"
          severity="secondary"
          outlined
          @click="resetToEmptyForm"
        />
        <Button
          label="Kembali ke Rincian"
          icon="pi pi-arrow-left"
          size="small"
          severity="secondary"
          text
          @click="router.push('/items/' + copiedFromSource.id)"
        />
      </div>
    </div>

    <!-- Loading Skeleton -->
    <div v-if="loadingData" class="grid lg:grid-cols-[380px_1fr] gap-5 items-start">
      <div class="flex flex-col gap-4">
        <div class="panel p-4 flex flex-col gap-3">
          <div class="skeleton h-4 w-40" />
          <div class="skeleton h-48 w-full rounded-[8px]" />
          <div class="skeleton h-9 w-full rounded-[8px]" />
        </div>
        <div class="panel p-4 flex flex-col gap-2.5">
          <div class="skeleton h-3.5 w-32" />
          <div class="skeleton h-11 w-full rounded-[8px]" />
          <div class="skeleton h-12 w-full rounded-[8px]" />
        </div>
      </div>
      <div class="flex flex-col gap-4">
        <div class="panel p-4 flex flex-col gap-3">
          <div class="skeleton h-4 w-44" />
          <div class="grid sm:grid-cols-2 gap-3">
            <div class="skeleton h-24 rounded-[10px]" />
            <div class="skeleton h-24 rounded-[10px]" />
          </div>
        </div>
        <div class="panel p-4 flex flex-col gap-3">
          <div class="skeleton h-4 w-36" />
          <div class="grid sm:grid-cols-2 gap-3">
            <div class="skeleton h-10 rounded-[8px] sm:col-span-2" />
            <div class="skeleton h-10 rounded-[8px]" />
            <div class="skeleton h-10 rounded-[8px]" />
            <div class="skeleton h-10 rounded-[8px]" />
            <div class="skeleton h-10 rounded-[8px]" />
          </div>
        </div>
      </div>
    </div>

    <!-- Main Symmetrical 2-Column Form Layout -->
    <form v-else @submit.prevent="submit" class="grid lg:grid-cols-[380px_1fr] gap-5 items-start">
      <!-- LEFT COLUMN: Foto, Geotag, & Preview Identitas -->
      <div class="flex flex-col gap-4 lg:sticky lg:top-4">
        <!-- Panel Foto & Geotag -->
        <Panel title="Foto Fisik &amp; Stempel Geotag" icon="pi pi-camera">
          <PhotoUploader
            ref="photoUploaderRef"
            :defer-upload="true"
            v-model="form.photo_url"
            v-model:geo-lat="form.geo_lat"
            v-model:geo-lng="form.geo_lng"
            v-model:geo-acc="form.geo_acc"
            v-model:geo-name="form.geo_name"
            :item-name="form.name"
            :location-name="form.location"
            label="Foto Fisik Barang"
            hint="Otomatis dicap banner GPS Map Camera proporsional"
          />
        </Panel>

        <!-- Preview Label SKU & Ringkasan Cepat -->
        <div class="panel p-4 flex flex-col gap-2.5">
          <div class="flex items-center justify-between">
            <span class="t-label">
              Preview Identitas Barang
            </span>
            <Tag :severity="form.track_stock ? 'warn' : 'info'"
                 :value="form.track_stock ? 'KONSUMABEL' : 'ASET TETAP'" class="!text-[10px]" />
          </div>

          <div class="t-mono font-bold text-[18px] text-acc-500 p-2.5 rounded-[8px] border border-line bg-panel-2 flex items-center justify-between">
            <span>{{ skuPreview }}</span>
            <i class="pi pi-barcode text-lg opacity-70" aria-hidden="true" />
          </div>

          <!-- Live Nama Barang -->
          <div class="p-2.5 rounded-[8px] border border-line bg-panel-2 flex flex-col gap-0.5">
            <span class="text-[10px] uppercase font-semibold text-txt-dim">Nama Barang:</span>
            <span class="font-bold text-[13px] break-words" :class="form.name ? 'text-txt' : 'italic text-txt-muted'">
              {{ form.name || 'Belum diisi (Ketik nama di formulir)' }}
            </span>
          </div>

          <p class="text-[11.5px] leading-relaxed text-txt-dim">
            {{ form.sku ? 'SKU manual kustom yang telah ditentukan pengguna.' : 'SKU otomatis di-generate berdasarkan kode kategori, tahun berjalan, dan urutan database.' }}
          </p>

          <div class="pt-2 border-t border-line flex flex-col gap-1.5 text-[11.5px]">
            <div v-if="enableKIR && form.item_code" class="flex items-center justify-between">
              <span class="text-txt-dim">No. Kode Barang:</span>
              <span class="t-mono font-bold text-acc-500">{{ form.item_code }}</span>
            </div>
            <div v-if="enableKIR && form.register_number" class="flex items-center justify-between">
              <span class="text-txt-dim">No. Register:</span>
              <span class="t-mono font-semibold text-txt">{{ form.register_number }}</span>
            </div>
            <div v-if="enableKIR" class="flex items-center justify-between">
              <span class="text-txt-dim">Kondisi Fisik:</span>
              <span class="font-semibold text-txt">{{ form.condition_status }}</span>
            </div>
            <div v-if="form.location" class="flex items-center justify-between">
              <span class="text-txt-dim">Ruangan:</span>
              <span class="font-semibold truncate max-w-[170px] text-txt">
                <span v-if="form.location_code" class="text-acc-500 t-mono mr-1">[{{ form.location_code }}]</span>
                {{ form.location }}
              </span>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-txt-dim">Jumlah Barang:</span>
              <span class="font-semibold text-txt t-num">{{ form.current_stock }} {{ form.track_stock ? form.unit : 'unit' }}</span>
            </div>
            <div v-if="form.track_stock" class="flex items-center justify-between">
              <span class="text-txt-dim">Satuan:</span>
              <span class="font-semibold text-txt">{{ form.unit }}</span>
            </div>
            <div v-if="form.track_stock" class="flex items-center justify-between">
              <span class="text-txt-dim">Batas Min. Stok:</span>
              <span class="font-semibold text-sig-warn t-num">{{ form.min_stock }} {{ form.unit }}</span>
            </div>
            <div v-if="enablePrice && form.price_per_unit > 0" class="flex items-center justify-between">
              <span class="text-txt-dim">Estimasi Harga:</span>
              <span class="font-semibold text-sig-ok t-num">Rp {{ Number(form.price_per_unit).toLocaleString('id-ID') }}</span>
            </div>
          </div>
        </div>

        <!-- Panduan Geotag Card -->
        <div class="panel p-3.5 flex items-start gap-3 text-[11.5px] leading-relaxed !bg-panel-2">
          <i class="pi pi-info-circle text-acc-500 text-sm mt-0.5 shrink-0" aria-hidden="true" />
          <div class="text-txt-dim">
            Pastikan izin lokasi GPS browser aktif saat mengambil foto kamera agar banner geotag memuat alamat akurat, koordinat presisi, dan peta lokasi.
          </div>
        </div>
      </div>

      <!-- RIGHT COLUMN: Jenis Pengelolaan, Data Pokok, & Spesifikasi -->
      <div class="flex flex-col gap-4">
        <!-- 1. Pilihan Jenis Pengelolaan Inventaris -->
        <Panel title="Jenis Pengelolaan Barang" icon="pi pi-sliders-h">
          <div class="flex flex-col gap-3">
            <span class="text-[12px] text-txt-dim">
              Pilih model pengelolaan aset inventaris di sistem:
            </span>

            <div class="grid sm:grid-cols-2 gap-3">
              <!-- Opsi A: Barang Konsumabel / Stok Habis Pakai -->
              <button
                type="button"
                class="p-4 rounded-[10px] border text-left cursor-pointer transition-colors flex items-start gap-3"
                :class="form.track_stock
                  ? 'border-acc-500 bg-[var(--acc-wash)] ring-1 ring-acc-500'
                  : 'border-line bg-panel-2 hover:border-[var(--acc-border)]'"
                @click="form.track_stock = true"
              >
                <div class="w-8 h-8 rounded-[8px] grid place-items-center shrink-0 border mt-0.5"
                     :class="form.track_stock
                       ? 'bg-acc-500 text-white border-acc-600'
                       : 'bg-panel text-txt-dim border-line'">
                  <i class="pi pi-box text-sm" aria-hidden="true" />
                </div>
                <div class="flex flex-col">
                  <div class="flex items-center gap-2">
                    <span class="font-bold text-[13px] text-txt">Barang Stok / Konsumabel</span>
                    <i v-if="form.track_stock" class="pi pi-check-circle text-xs text-acc-500 font-bold" aria-hidden="true" />
                  </div>
                  <span class="text-[11px] mt-1 leading-snug text-txt-dim">
                    Contoh: Pensil, Kertas, Spidol, APD, Tinta. Jumlah bertambah/berkurang via menu <strong class="text-txt">Stok Masuk / Keluar</strong> dengan foto geotag.
                  </span>
                </div>
              </button>

              <!-- Opsi B: Aset Tetap / Unit Mandiri -->
              <button
                type="button"
                class="p-4 rounded-[10px] border text-left cursor-pointer transition-colors flex items-start gap-3"
                :class="!form.track_stock
                  ? 'border-acc-500 bg-[var(--acc-wash)] ring-1 ring-acc-500'
                  : 'border-line bg-panel-2 hover:border-[var(--acc-border)]'"
                @click="form.track_stock = false; form.unit = 'unit'; if ((Number(form.current_stock) || 0) < 1) form.current_stock = 1"
              >
                <div class="w-8 h-8 rounded-[8px] grid place-items-center shrink-0 border mt-0.5"
                     :class="!form.track_stock
                       ? 'bg-acc-500 text-white border-acc-600'
                       : 'bg-panel text-txt-dim border-line'">
                  <i class="pi pi-desktop text-sm" aria-hidden="true" />
                </div>
                <div class="flex flex-col">
                  <div class="flex items-center gap-2">
                    <span class="font-bold text-[13px] text-txt">Aset Tetap / Unit Mandiri</span>
                    <i v-if="!form.track_stock" class="pi pi-check-circle text-xs text-acc-500 font-bold" aria-hidden="true" />
                  </div>
                  <span class="text-[11px] mt-1 leading-snug text-txt-dim">
                    Contoh: Laptop, Komputer, Kendaraan, Alat Medis. Setiap unit adalah satu aset bernomor seri tersendiri yang dipantau jadwal servis/kalibrasinya.
                  </span>
                </div>
              </button>
            </div>

            <!-- Notice Callout for New Stock Item -->
            <div
              v-if="form.track_stock && !isEdit"
              class="p-3 rounded-[8px] border border-[var(--acc-border)] bg-[var(--acc-wash)] text-[11.5px] flex items-center gap-2.5 text-txt"
            >
              <i class="pi pi-info-circle text-sm text-acc-500 shrink-0" aria-hidden="true" />
              <span class="text-txt-dim">
                Isi jumlah stok awal pada kolom <strong class="text-txt">Jumlah Barang</strong> di bawah (default 0). Penambahan atau pengurangan stok selanjutnya juga dapat dicatat pada halaman detail melalui tombol <strong class="text-txt">+ Tambah Stok Masuk</strong> berbukti foto geotag.
              </span>
            </div>

            <!-- Info Unit Fisik Aset Tetap / BMD -->
            <div
              v-if="!form.track_stock"
              class="p-3 rounded-[8px] border border-[var(--acc-border)] bg-[var(--acc-wash)] text-[11.5px] flex items-start gap-2.5 text-txt"
            >
              <i class="pi pi-clone text-sm text-acc-500 mt-0.5 shrink-0" aria-hidden="true" />
              <span class="text-txt-dim leading-relaxed">
                Jumlah unit fisik aset tetap diatur pada kolom <strong class="text-txt">Jumlah Barang</strong> di bawah (1–500), sedangkan <strong class="text-txt">No. Register</strong> diisi terpisah pada bagian <strong class="text-txt">Buku Inventaris (KIR / KIB)</strong>. Setiap unit dapat dilengkapi nomor seri, barcode, dan foto di tab <strong class="text-txt">Daftar Unit Fisik</strong>.
              </span>
            </div>
          </div>
        </Panel>

        <!-- 2. Data Pokok Barang -->
        <Panel title="Data Pokok Barang" icon="pi pi-box">
          <div class="grid sm:grid-cols-2 gap-3.5 text-[12.5px]">
            <Field label="Jenis Barang / Nama Barang" required span hint="Nama spesifik aset atau barang inventaris">
              <InputText v-model="form.name" required placeholder="mis. Kursi Kerja Putar / Laptop Lenovo ThinkPad / Tensimeter Digital" class="w-full !text-[12.5px]" fluid />
            </Field>

            <Field label="Kategori Barang" required>
              <Select
                v-model="form.category"
                :options="categories"
                optionLabel="name"
                optionValue="name"
                required
                placeholder="Pilih kategori"
                class="w-full !text-[12.5px]"
                fluid
                filter
              />
            </Field>

            <Field
              label="Ruangan / Lokasi Penempatan"
              required
              :hint="form.location_code ? `Kode Lokasi otomatis: ${form.location_code}` : 'Pilih ruangan penempatan barang'"
            >
              <Select
                v-model="form.location"
                :options="locations"
                optionLabel="name"
                optionValue="name"
                required
                placeholder="Pilih ruangan"
                class="w-full !text-[12.5px]"
                fluid
                filter
                @change="onLocationSelect($event.value)"
              >
                <template #option="{ option }">
                  <div class="flex items-center justify-between gap-2 w-full">
                    <span>{{ option.name }}</span>
                    <span v-if="option.code" class="t-mono text-[11px] text-acc-500">[{{ option.code }}]</span>
                  </div>
                </template>
              </Select>
            </Field>

            <Field
              :label="form.track_stock ? 'Jumlah Barang (Stok)' : 'Jumlah Barang (Unit)'"
              required
              :hint="form.track_stock
                ? (isEdit ? 'Jumlah stok barang saat ini' : 'Jumlah stok awal barang (default 0)')
                : (isEdit ? 'Jumlah unit fisik aset tetap (minimal 1)' : 'Jumlah unit fisik awal yang didaftarkan (1–500)')"
            >
              <InputNumber
                v-model="form.current_stock"
                :min="form.track_stock ? 0 : 1"
                :max="form.track_stock ? undefined : 500"
                showButtons
                buttonLayout="horizontal"
                decrementButtonClass="p-button-secondary"
                incrementButtonClass="p-button-secondary"
                incrementButtonIcon="pi pi-plus"
                decrementButtonIcon="pi pi-minus"
                class="w-full !text-[12.5px]"
                fluid
              />
            </Field>

            <!-- Satuan Unit (Hanya untuk Barang Stok / Konsumabel) -->
            <Field v-if="form.track_stock" label="Satuan Unit" required>
              <Select v-model="form.unit" :options="unitOptions" class="w-full !text-[12.5px]" fluid />
            </Field>

            <!-- Batas Minimum Stok (Hanya untuk Barang Konsumabel) -->
            <Field v-if="form.track_stock" label="Batas Minimum Peringatan (Stok Kritis)" hint="Peringatan otomatis dikirim ke Telegram bila stok ≤ angka ini">
              <InputNumber v-model="form.min_stock" :min="0" showButtons class="w-full !text-[12.5px]" fluid />
            </Field>

            <!-- Estimasi Harga Satuan (Rp) - Opsi On/Off -->
            <div class="flex flex-col gap-1.5 min-w-0">
              <div class="flex items-center justify-between">
                <span class="text-[11.5px] font-semibold text-txt-dim">
                  Estimasi Harga Satuan (Rp)
                </span>
                <label class="flex items-center gap-1.5 cursor-pointer select-none" @click.stop>
                  <span class="text-[11px] font-semibold" :class="enablePrice ? 'text-acc-500' : 'text-txt-dim'">
                    {{ enablePrice ? 'Aktif' : 'Nonaktif' }}
                  </span>
                  <ToggleSwitch v-model="enablePrice" aria-label="Aktifkan estimasi harga satuan" />
                </label>
              </div>

              <div v-if="enablePrice" class="flex flex-col gap-1">
                <InputNumber
                  v-model="form.price_per_unit"
                  :min="0"
                  mode="currency"
                  currency="IDR"
                  locale="id-ID"
                  class="w-full !text-[12.5px]"
                  fluid
                />
                <span class="text-[11px] text-txt-dim">Harga beli atau nilai perolehan per unit</span>
              </div>
              <div
                v-else
                class="p-2.5 rounded-[8px] border border-line bg-panel-2 text-[11.5px] flex items-center justify-between gap-2 cursor-pointer hover:border-[var(--acc-border)] transition-colors"
                @click="enablePrice = true"
              >
                <span class="text-txt-dim">Pencatatan estimasi harga nonaktif (Rp 0)</span>
                <button type="button" class="text-acc-500 font-semibold text-[11px] hover:underline cursor-pointer" @click.stop="enablePrice = true">
                  Nyalakan
                </button>
              </div>
            </div>

            <Field label="SKU Kustom (Opsional)" hint="Kosongkan untuk penomoran otomatis sistem">
              <InputText v-model="form.sku" :placeholder="skuPreview" class="w-full t-mono !text-[12.5px]" fluid />
            </Field>
          </div>
        </Panel>

        <!-- 3. Buku Inventaris (KIR / KIB) - Opsi On/Off -->
        <Panel title="Buku Inventaris (KIR / KIB)" icon="pi pi-book">
          <template #actions>
            <label class="flex items-center gap-2 cursor-pointer select-none" @click.stop>
              <span class="text-[11.5px] font-semibold" :class="enableKIR ? 'text-acc-500' : 'text-txt-dim'">
                {{ enableKIR ? 'Aktif' : 'Nonaktif' }}
              </span>
              <ToggleSwitch v-model="enableKIR" aria-label="Aktifkan buku inventaris KIR/KIB" />
            </label>
          </template>

          <div v-if="enableKIR" class="grid sm:grid-cols-2 gap-3.5 text-[12.5px]">
            <Field label="No. Kode Barang" hint="Kode barang standar inventaris daerah/KIB (cth: 02.03.01.02.05)">
              <InputText v-model="form.item_code" placeholder="mis. 02.03.01.02.05" class="w-full t-mono !text-[12.5px]" fluid />
            </Field>

            <Field label="No. Register" hint="Nomor urut register pencatatan aset (cth: 0001, REG-2026-001)">
              <InputText v-model="form.register_number" placeholder="mis. 0001 / REG-2026-001" class="w-full t-mono !text-[12.5px]" fluid />
            </Field>

            <Field
              label="NO. KODE LOKASI"
              :hint="selectedLocationCode ? 'Otomatis diambil dari kode ruangan yang dipilih' : 'Ruangan belum memiliki kode di menu Lokasi (bisa diisi manual)'"
            >
              <InputText
                v-model="form.location_code"
                :readonly="!!selectedLocationCode"
                :placeholder="selectedLocationCode || 'mis. 11.02.01'"
                :class="['w-full t-mono !text-[12.5px]', selectedLocationCode ? 'opacity-85 cursor-not-allowed' : '']"
                fluid
              />
            </Field>

            <Field label="Keadaan Barang" hint="Kondisi fisik aset inventaris saat didaftarkan">
              <Select v-model="form.condition_status" :options="conditionOptions" class="w-full !text-[12.5px]" fluid />
            </Field>

            <Field label="Merk / Model">
              <div class="grid grid-cols-2 gap-2">
                <InputText v-model="form.merk" placeholder="Merk (mis. Chitose / Lenovo)" class="w-full !text-[12.5px]" fluid />
                <InputText v-model="form.type_model" placeholder="Model (mis. L-14 / Ergo)" class="w-full !text-[12.5px]" fluid />
              </div>
            </Field>

            <Field label="No. Seri Pabrik" hint="Nomor seri unik pabrikan (Serial Number jika ada)">
              <InputText v-model="form.serial_number" placeholder="Nomor seri pabrik (barcode SN)" class="w-full t-mono !text-[12.5px]" fluid />
            </Field>

            <Field label="Ukuran" hint="Dimensi, volume, atau ukuran fisik aset">
              <InputText v-model="form.size" placeholder="mis. 120 x 80 x 75 cm / 14 inch / 2 Liter" class="w-full !text-[12.5px]" fluid />
            </Field>

            <Field label="Bahan" hint="Material utama pembentuk barang">
              <InputText v-model="form.material" placeholder="mis. Kayu Jati / Besi / Plastik / Aluminium / Kain" class="w-full !text-[12.5px]" fluid />
            </Field>

            <Field label="Tahun Pembuatan / Pembelian" span hint="Tahun pembuatan pabrik atau tahun perolehan/pembelian barang">
              <InputText v-model="form.procurement_year" placeholder="mis. 2024 / 2021" class="w-full t-mono !text-[12.5px]" fluid />
            </Field>
          </div>

          <div v-else class="p-3.5 rounded-[10px] border border-line bg-panel-2 flex items-center justify-between gap-3 cursor-pointer hover:border-[var(--acc-border)] transition-colors"
               @click="enableKIR = true">
            <div class="flex items-center gap-2.5 text-[12px] text-txt-dim">
              <i class="pi pi-info-circle text-acc-500 text-sm shrink-0" aria-hidden="true" />
              <span>Opsi format inventaris KIR/KIB (Kode Barang, Register, Kode Lokasi, Merk, Seri, Ukuran, Bahan, Tahun, Kondisi) sedang nonaktif.</span>
            </div>
            <Button label="Nyalakan (ON)" icon="pi pi-power-off" size="small" severity="primary" @click.stop="enableKIR = true" />
          </div>
        </Panel>

        <!-- 4. Detail Pengadaan & Keterangan - Opsi On/Off -->
        <Panel title="Detail Pengadaan &amp; Keterangan" icon="pi pi-clipboard">
          <template #actions>
            <label class="flex items-center gap-2 cursor-pointer select-none" @click.stop>
              <span class="text-[11.5px] font-semibold" :class="enableProcurement ? 'text-acc-500' : 'text-txt-dim'">
                {{ enableProcurement ? 'Aktif' : 'Nonaktif' }}
              </span>
              <ToggleSwitch v-model="enableProcurement" aria-label="Aktifkan detail pengadaan dan keterangan" />
            </label>
          </template>

          <div v-if="enableProcurement" class="grid sm:grid-cols-2 gap-3.5 text-[12.5px]">
            <Field label="Sumber Dana">
              <InputText v-model="form.funding_source" placeholder="mis. APBD / DAK / BLU / Hibah" class="w-full !text-[12.5px]" fluid />
            </Field>

            <Field label="Distributor / Vendor Penyedia">
              <InputText v-model="form.distributor" placeholder="Nama rekanan penyedia pengadaan" class="w-full !text-[12.5px]" fluid />
            </Field>

            <Field label="Izin Edar AKL / AKD" span hint="Khusus alat kesehatan atau reagen laboratorium">
              <InputText v-model="form.akl_akd" placeholder="mis. KEMENKES RI AKL 20501812345" class="w-full !text-[12.5px]" fluid />
            </Field>

            <Field label="Keterangan / Catatan Tambahan" span hint="Rincian kelengkapan aksesoris, spesifikasi, atau info garansi">
              <Textarea
                v-model="form.description"
                rows="3"
                autoResize
                placeholder="Tuliskan keterangan tambahan, kelengkapan aksesoris, catatan garansi, atau instruksi pemakaian..."
                class="w-full !text-[12.5px]"
                fluid
              />
            </Field>
          </div>

          <div v-else class="p-3.5 rounded-[10px] border border-line bg-panel-2 flex items-center justify-between gap-3 cursor-pointer hover:border-[var(--acc-border)] transition-colors"
               @click="enableProcurement = true">
            <div class="flex items-center gap-2.5 text-[12px] text-txt-dim">
              <i class="pi pi-info-circle text-acc-500 text-sm shrink-0" aria-hidden="true" />
              <span>Opsi detail pengadaan (Sumber Dana, Rekanan Vendor, AKL/AKD) &amp; catatan keterangan tambahan sedang nonaktif.</span>
            </div>
            <Button label="Nyalakan (ON)" icon="pi pi-power-off" size="small" severity="primary" @click.stop="enableProcurement = true" />
          </div>
        </Panel>

        <!-- Bottom Unified Action Footer -->
        <div class="panel p-4 flex items-center justify-between gap-3">
          <Button
            label="Batal &amp; Kembali"
            icon="pi pi-arrow-left"
            size="small"
            severity="secondary"
            text
            @click="router.push(isEdit ? '/items/' + route.params.id : (copiedFromSource ? '/items/' + copiedFromSource.id : '/items'))"
          />
          <div class="flex items-center gap-2">
            <Button
              v-if="copiedFromSource"
              label="Batal Salin (Form Kosong)"
              icon="pi pi-times"
              size="small"
              severity="secondary"
              outlined
              @click="resetToEmptyForm"
            />
            <Button
              :label="isEdit ? 'Simpan Perubahan' : (copiedFromSource ? 'Simpan Unit Baru' : 'Daftarkan Barang')"
              icon="pi pi-check"
              size="small"
              :loading="loading"
              :disabled="!dirty || loadingData"
              @click="submit"
            />
          </div>
        </div>
      </div>
    </form>
  </div>
</template>
