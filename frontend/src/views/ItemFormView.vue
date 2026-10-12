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
const initialUnits = ref(1)
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
  initialUnits.value = 1
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

function onLocationSelect(locName: string) {
  form.value.location = locName
  const found = locations.value.find((l: any) => l.name === locName)
  if (found && found.code) {
    form.value.location_code = found.code
  }
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
    form.value.location_code = ''
    form.value.merk = ''
    form.value.type_model = ''
    form.value.serial_number = ''
    form.value.size = ''
    form.value.material = ''
    form.value.procurement_year = ''
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
      await api.put(`/items/${route.params.id}`, form.value)
      toast.add({ severity: 'success', summary: 'Perubahan disimpan', detail: form.value.name, life: 2500 })
      router.push('/items/' + route.params.id)
    } else {
      const payload: any = { ...form.value }
      if (!form.value.track_stock) {
        payload.initial_units = initialUnits.value || 1
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
      class="panel p-4 mb-4 flex flex-wrap items-center justify-between gap-3 border-l-4 border-indigo-500 bg-indigo-50/70 dark:bg-indigo-950/40 shadow-xs"
    >
      <div class="flex items-center gap-3">
        <div class="w-9 h-9 rounded-lg bg-indigo-500 text-white grid place-items-center shrink-0">
          <i class="pi pi-clone text-base" />
        </div>
        <div class="text-[12.5px]">
          <div class="font-bold text-slate-900 dark:text-white flex items-center gap-2">
            <span>Mode: Tambah Unit Sejenis</span>
            <Tag severity="info" class="!text-[10px]" value="SALIN SPESIFIKASI" />
          </div>
          <div class="text-slate-600 dark:text-slate-400 mt-0.5">
            Spesifikasi disalin dari: <strong class="text-indigo-600 dark:text-indigo-400">{{ copiedFromSource.name }}</strong>
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
    <div v-if="loadingData" class="panel p-16 grid place-items-center">
      <div class="flex flex-col items-center gap-3">
        <i class="pi pi-spin pi-spinner text-3xl text-indigo-500" />
        <span class="text-[12.5px]" style="color: var(--txt-dim)">Memuat data inventaris...</span>
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
            <span class="text-[11px] font-semibold uppercase tracking-wider" style="color: var(--txt-dim)">
              Preview Identitas Barang
            </span>
            <Tag :severity="form.track_stock ? 'warn' : 'info'"
                 :value="form.track_stock ? 'KONSUMABEL' : 'ASET TETAP'" class="!text-[10px]" />
          </div>

          <div class="t-mono font-bold text-[18px] text-indigo-500 dark:text-indigo-400 p-2.5 rounded-lg border flex items-center justify-between"
               style="background: var(--panel-2); border-color: var(--line)">
            <span>{{ skuPreview }}</span>
            <i class="pi pi-barcode text-lg opacity-70" />
          </div>

          <!-- Live Nama Barang -->
          <div class="p-2.5 rounded-lg border flex flex-col gap-0.5"
               style="background: var(--panel-2); border-color: var(--line)">
            <span class="text-[10px] uppercase font-semibold" style="color: var(--txt-dim)">Nama Barang:</span>
            <span class="font-bold text-[13px] break-words" :class="form.name ? 'text-slate-900 dark:text-white' : 'italic text-slate-400'">
              {{ form.name || 'Belum diisi (Ketik nama di formulir)' }}
            </span>
          </div>

          <p class="text-[11.5px] leading-relaxed" style="color: var(--txt-dim)">
            {{ form.sku ? 'SKU manual kustom yang telah ditentukan pengguna.' : 'SKU otomatis di-generate berdasarkan kode kategori, tahun berjalan, dan urutan database.' }}
          </p>

          <div class="pt-2 border-t flex flex-col gap-1.5 text-[11.5px]" style="border-color: var(--line)">
            <div v-if="enableKIR && form.item_code" class="flex items-center justify-between">
              <span style="color: var(--txt-dim)">No. Kode Barang:</span>
              <span class="t-mono font-bold text-indigo-500">{{ form.item_code }}</span>
            </div>
            <div v-if="enableKIR && form.register_number" class="flex items-center justify-between">
              <span style="color: var(--txt-dim)">No. Register:</span>
              <span class="t-mono font-semibold" style="color: var(--txt)">{{ form.register_number }}</span>
            </div>
            <div v-if="enableKIR" class="flex items-center justify-between">
              <span style="color: var(--txt-dim)">Kondisi Fisik:</span>
              <span class="font-semibold" style="color: var(--txt)">{{ form.condition_status }}</span>
            </div>
            <div v-if="form.location" class="flex items-center justify-between">
              <span style="color: var(--txt-dim)">Ruangan:</span>
              <span class="font-semibold truncate max-w-[170px]" style="color: var(--txt)">
                <span v-if="enableKIR && form.location_code" class="text-indigo-500 font-mono mr-1">[{{ form.location_code }}]</span>
                {{ form.location }}
              </span>
            </div>
            <div v-if="form.track_stock" class="flex items-center justify-between">
              <span style="color: var(--txt-dim)">Satuan:</span>
              <span class="font-semibold" style="color: var(--txt)">{{ form.unit }}</span>
            </div>
            <div v-if="form.track_stock" class="flex items-center justify-between">
              <span style="color: var(--txt-dim)">Batas Min. Stok:</span>
              <span class="font-semibold text-amber-500">{{ form.min_stock }} {{ form.unit }}</span>
            </div>
            <div v-if="enablePrice && form.price_per_unit > 0" class="flex items-center justify-between">
              <span style="color: var(--txt-dim)">Estimasi Harga:</span>
              <span class="font-semibold text-emerald-500">Rp {{ Number(form.price_per_unit).toLocaleString('id-ID') }}</span>
            </div>
          </div>
        </div>

        <!-- Panduan Geotag Card -->
        <div class="panel p-3.5 flex items-start gap-3 text-[11.5px] leading-relaxed" style="background: var(--panel-2)">
          <i class="pi pi-info-circle text-indigo-500 text-sm mt-0.5 shrink-0" />
          <div style="color: var(--txt-dim)">
            Pastikan izin lokasi GPS browser aktif saat mengambil foto kamera agar banner geotag memuat alamat akurat, koordinat presisi, dan peta lokasi.
          </div>
        </div>
      </div>

      <!-- RIGHT COLUMN: Jenis Pengelolaan, Data Pokok, & Spesifikasi -->
      <div class="flex flex-col gap-4">
        <!-- 1. Pilihan Jenis Pengelolaan Inventaris -->
        <Panel title="Jenis Pengelolaan Barang" icon="pi pi-sliders-h">
          <div class="flex flex-col gap-3">
            <span class="text-[12px]" style="color: var(--txt-dim)">
              Pilih model pengelolaan aset inventaris di sistem:
            </span>

            <div class="grid sm:grid-cols-2 gap-3">
              <!-- Opsi A: Barang Konsumabel / Stok Habis Pakai -->
              <div
                class="p-4 rounded-xl border cursor-pointer transition-all flex items-start gap-3"
                :class="form.track_stock
                  ? 'border-indigo-500 bg-indigo-50/70 dark:bg-indigo-950/30 shadow-xs ring-1 ring-indigo-500'
                  : 'hover:border-slate-300 dark:hover:border-slate-700'"
                :style="!form.track_stock ? { background: 'var(--panel-2)', borderColor: 'var(--line)' } : {}"
                @click="form.track_stock = true"
              >
                <div class="w-8 h-8 rounded-lg grid place-items-center shrink-0 border mt-0.5"
                     :class="form.track_stock
                       ? 'bg-indigo-500 text-white border-indigo-600'
                       : 'bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400 border-slate-200 dark:border-slate-700'">
                  <i class="pi pi-box text-sm" />
                </div>
                <div class="flex flex-col">
                  <div class="flex items-center gap-2">
                    <span class="font-bold text-[13px]" style="color: var(--txt)">Barang Stok / Konsumabel</span>
                    <i v-if="form.track_stock" class="pi pi-check-circle text-xs text-indigo-500 font-bold" />
                  </div>
                  <span class="text-[11px] mt-1 leading-snug" style="color: var(--txt-dim)">
                    Contoh: Pensil, Kertas, Spidol, APD, Tinta. Jumlah bertambah/berkurang via menu <strong>Stok Masuk / Keluar</strong> dengan foto geotag.
                  </span>
                </div>
              </div>

              <!-- Opsi B: Aset Tetap / Unit Mandiri -->
              <div
                class="p-4 rounded-xl border cursor-pointer transition-all flex items-start gap-3"
                :class="!form.track_stock
                  ? 'border-indigo-500 bg-indigo-50/70 dark:bg-indigo-950/30 shadow-xs ring-1 ring-indigo-500'
                  : 'hover:border-slate-300 dark:hover:border-slate-700'"
                :style="form.track_stock ? { background: 'var(--panel-2)', borderColor: 'var(--line)' } : {}"
                @click="form.track_stock = false; form.unit = 'unit'"
              >
                <div class="w-8 h-8 rounded-lg grid place-items-center shrink-0 border mt-0.5"
                     :class="!form.track_stock
                       ? 'bg-indigo-500 text-white border-indigo-600'
                       : 'bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400 border-slate-200 dark:border-slate-700'">
                  <i class="pi pi-desktop text-sm" />
                </div>
                <div class="flex flex-col">
                  <div class="flex items-center gap-2">
                    <span class="font-bold text-[13px]" style="color: var(--txt)">Aset Tetap / Unit Mandiri</span>
                    <i v-if="!form.track_stock" class="pi pi-check-circle text-xs text-indigo-500 font-bold" />
                  </div>
                  <span class="text-[11px] mt-1 leading-snug" style="color: var(--txt-dim)">
                    Contoh: Laptop, Komputer, Kendaraan, Alat Medis. Setiap unit adalah satu aset bernomor seri tersendiri yang dipantau jadwal servis/kalibrasinya.
                  </span>
                </div>
              </div>
            </div>

            <!-- Notice Callout for New Stock Item -->
            <div
              v-if="form.track_stock && !isEdit"
              class="p-3 rounded-lg border text-[11.5px] flex items-center gap-2.5"
              style="background: rgba(99, 102, 241, 0.08); border-color: rgba(99, 102, 241, 0.25); color: var(--txt)"
            >
              <i class="pi pi-info-circle text-sm text-indigo-500 shrink-0" />
              <span style="color: var(--txt-dim)">
                Barang stok akan didaftarkan dengan stok awal 0. Penambahan stok dapat dicatat langsung setelah disimpan pada halaman detail melalui tombol <strong style="color: var(--txt)">+ Tambah Stok Masuk</strong> berbukti foto geotag.
              </span>
            </div>

            <!-- Opsi Unit Fisik Aset Tetap / BMD (Selaras dengan Tab Daftar Unit Fisik) -->
            <div
              v-if="!form.track_stock && !isEdit"
              class="p-3.5 rounded-xl border flex flex-col gap-3"
              style="background: rgba(99, 102, 241, 0.06); border-color: rgba(99, 102, 241, 0.25)"
            >
              <div class="flex items-start gap-2.5 text-[12px]">
                <i class="pi pi-clone text-indigo-500 text-sm mt-0.5 shrink-0" />
                <div class="flex flex-col gap-0.5">
                  <span class="font-bold" style="color: var(--txt)">Pencatatan Multi-Unit Fisik (Aset Tetap / BMD)</span>
                  <span style="color: var(--txt-dim)" class="text-[11.5px] leading-relaxed">
                    Untuk barang sejenis dalam satu ruangan (contoh: <strong>5 unit Meja / Kursi</strong> dengan No. Kode Barang sama), tentukan jumlah unit di bawah. Sistem akan mendaftarkan unit-unit tersebut, dan setiap unit dapat dilengkapi nomor seri, barcode, serta foto fisik masing-masing di tab <strong>Daftar Unit Fisik</strong> setelah disimpan.
                  </span>
                </div>
              </div>

              <div class="grid sm:grid-cols-2 gap-3 pt-2.5 border-t" style="border-color: rgba(99, 102, 241, 0.15)">
                <div class="flex flex-col gap-1">
                  <label class="text-[11.5px] font-semibold" style="color: var(--txt-dim)">
                    Jumlah Unit Fisik Awal
                  </label>
                  <InputNumber
                    v-model="initialUnits"
                    :min="1"
                    :max="500"
                    showButtons
                    buttonLayout="horizontal"
                    decrementButtonClass="p-button-secondary"
                    incrementButtonClass="p-button-secondary"
                    incrementButtonIcon="pi pi-plus"
                    decrementButtonIcon="pi pi-minus"
                    class="w-full !text-[12.5px]"
                    fluid
                  />
                  <span class="text-[10.5px]" style="color: var(--txt-dim)">
                    Unit 1 s/d {{ initialUnits }} otomatis dibuat berurutan.
                  </span>
                </div>

                <div class="flex flex-col gap-1">
                  <label class="text-[11.5px] font-semibold" style="color: var(--txt-dim)">
                    No. Register Awal (Opsional)
                  </label>
                  <InputText
                    v-model="form.register_number"
                    placeholder="mis. 0001"
                    class="w-full t-mono !text-[12.5px]"
                    fluid
                    @input="enableKIR = true"
                  />
                  <span class="text-[10.5px]" style="color: var(--txt-dim)">
                    Urutan register berikutnya otomatis diteruskan (0001, 0002, dst.).
                  </span>
                </div>
              </div>
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

            <Field label="Ruangan / Lokasi Penempatan" required>
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
                <span class="text-[11.5px] font-semibold" style="color: var(--txt-dim)">
                  Estimasi Harga Satuan (Rp)
                </span>
                <label class="flex items-center gap-1.5 cursor-pointer select-none" @click.stop>
                  <span class="text-[11px] font-semibold" :style="{ color: enablePrice ? 'var(--primary-color, #6366f1)' : 'var(--txt-dim)' }">
                    {{ enablePrice ? 'Aktif' : 'Nonaktif' }}
                  </span>
                  <ToggleSwitch v-model="enablePrice" />
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
                <span class="text-[11px]" style="color: var(--txt-dim)">Harga beli atau nilai perolehan per unit</span>
              </div>
              <div
                v-else
                class="p-2.5 rounded-lg border text-[11.5px] flex items-center justify-between gap-2 cursor-pointer hover:border-indigo-400 transition-colors"
                style="background: var(--panel-2); border-color: var(--line)"
                @click="enablePrice = true"
              >
                <span style="color: var(--txt-dim)">Pencatatan estimasi harga nonaktif (Rp 0)</span>
                <button type="button" class="text-indigo-500 font-semibold text-[11px] hover:underline" @click.stop="enablePrice = true">
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
              <span class="text-[11.5px] font-semibold" :style="{ color: enableKIR ? 'var(--primary-color, #6366f1)' : 'var(--txt-dim)' }">
                {{ enableKIR ? 'Aktif' : 'Nonaktif' }}
              </span>
              <ToggleSwitch v-model="enableKIR" />
            </label>
          </template>

          <div v-if="enableKIR" class="grid sm:grid-cols-2 gap-3.5 text-[12.5px]">
            <Field label="No. Kode Barang" hint="Kode barang standar inventaris daerah/KIB (cth: 02.03.01.02.05)">
              <InputText v-model="form.item_code" placeholder="mis. 02.03.01.02.05" class="w-full t-mono !text-[12.5px]" fluid />
            </Field>

            <Field label="Jumlah Barang / Register" hint="Nomor urut register pencatatan aset (cth: 0001, 0015)">
              <InputText v-model="form.register_number" placeholder="mis. 0001 / REG-2026-001" class="w-full t-mono !text-[12.5px]" fluid />
            </Field>

            <Field label="NO. KODE LOKASI" hint="Nomor kode lokasi / ruangan penempatan aset">
              <InputText v-model="form.location_code" placeholder="mis. 11.02.01" class="w-full t-mono !text-[12.5px]" fluid />
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

          <div v-else class="p-3.5 rounded-xl border flex items-center justify-between gap-3 cursor-pointer hover:border-indigo-400 transition-colors"
               style="background: var(--panel-2); border-color: var(--line)"
               @click="enableKIR = true">
            <div class="flex items-center gap-2.5 text-[12px]" style="color: var(--txt-dim)">
              <i class="pi pi-info-circle text-indigo-500 text-sm shrink-0" />
              <span>Opsi format inventaris KIR/KIB (Kode Barang, Register, Kode Lokasi, Merk, Seri, Ukuran, Bahan, Tahun, Kondisi) sedang nonaktif.</span>
            </div>
            <Button label="Nyalakan (ON)" icon="pi pi-power-off" size="small" severity="primary" @click.stop="enableKIR = true" />
          </div>
        </Panel>

        <!-- 4. Detail Pengadaan & Keterangan - Opsi On/Off -->
        <Panel title="Detail Pengadaan &amp; Keterangan" icon="pi pi-clipboard">
          <template #actions>
            <label class="flex items-center gap-2 cursor-pointer select-none" @click.stop>
              <span class="text-[11.5px] font-semibold" :style="{ color: enableProcurement ? 'var(--primary-color, #6366f1)' : 'var(--txt-dim)' }">
                {{ enableProcurement ? 'Aktif' : 'Nonaktif' }}
              </span>
              <ToggleSwitch v-model="enableProcurement" />
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

          <div v-else class="p-3.5 rounded-xl border flex items-center justify-between gap-3 cursor-pointer hover:border-indigo-400 transition-colors"
               style="background: var(--panel-2); border-color: var(--line)"
               @click="enableProcurement = true">
            <div class="flex items-center gap-2.5 text-[12px]" style="color: var(--txt-dim)">
              <i class="pi pi-info-circle text-indigo-500 text-sm shrink-0" />
              <span>Opsi detail pengadaan (Sumber Dana, Rekanan Vendor, AKL/AKD) &amp; catatan keterangan tambahan sedang nonaktif.</span>
            </div>
            <Button label="Nyalakan (ON)" icon="pi pi-power-off" size="small" severity="primary" @click.stop="enableProcurement = true" />
          </div>
        </Panel>

        <!-- Bottom Unified Action Footer -->
        <div class="panel p-4 flex items-center justify-between gap-3 shadow-xs">
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
