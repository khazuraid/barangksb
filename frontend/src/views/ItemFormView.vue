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
const yearOptions = Array.from({ length: 25 }, (_, i) => String(new Date().getFullYear() - i))

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
      const res = await api.post('/items', form.value)
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
          @click="router.push(isEdit ? '/items/' + route.params.id : '/items')"
        />
        <Button
          :label="isEdit ? 'Simpan Perubahan' : 'Simpan Barang'"
          icon="pi pi-check"
          size="small"
          :loading="loading"
          :disabled="!dirty || loadingData"
          @click="submit"
        />
      </template>
    </PageHeader>

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

            <Field label="Estimasi Harga Satuan (Rp)" hint="Harga beli atau nilai perolehan per unit">
              <InputNumber
                v-model="form.price_per_unit"
                :min="0"
                mode="currency"
                currency="IDR"
                locale="id-ID"
                class="w-full !text-[12.5px]"
                fluid
              />
            </Field>

            <Field label="SKU Kustom (Opsional)" hint="Kosongkan untuk penomoran otomatis sistem">
              <InputText v-model="form.sku" :placeholder="skuPreview" class="w-full t-mono !text-[12.5px]" fluid />
            </Field>
          </div>
        </Panel>

        <!-- 3. Buku Inventaris (KIR / KIB) - Opsi On/Off -->
        <Panel title="Buku Inventaris (KIR / KIB)" icon="pi pi-book">
          <template #actions>
            <div class="flex items-center gap-2">
              <span class="text-[11.5px] font-semibold" :style="{ color: enableKIR ? 'var(--primary-color, #6366f1)' : 'var(--txt-dim)' }">
                {{ enableKIR ? 'Aktif' : 'Nonaktif' }}
              </span>
              <ToggleSwitch v-model="enableKIR" />
            </div>
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

            <Field label="Tahun Pembuatan / Pembelian" span>
              <Select v-model="form.procurement_year" :options="yearOptions" showClear placeholder="Pilih tahun pembuatan / pembelian" class="w-full !text-[12.5px]" fluid />
            </Field>
          </div>

          <div v-else class="p-3.5 rounded-xl border flex items-center justify-between gap-3"
               style="background: var(--panel-2); border-color: var(--line)">
            <div class="flex items-center gap-2.5 text-[12px]" style="color: var(--txt-dim)">
              <i class="pi pi-info-circle text-indigo-500 text-sm" />
              <span>Opsi format inventaris KIR/KIB (Kode Barang, Register, Kode Lokasi, Merk, Seri, Ukuran, Bahan, Tahun, Kondisi) sedang nonaktif.</span>
            </div>
            <Button label="Aktifkan Opsi KIR/KIB" icon="pi pi-plus" size="small" text severity="primary" @click="enableKIR = true" />
          </div>
        </Panel>

        <!-- 4. Detail Pengadaan & Keterangan - Opsi On/Off -->
        <Panel title="Detail Pengadaan &amp; Keterangan" icon="pi pi-clipboard">
          <template #actions>
            <div class="flex items-center gap-2">
              <span class="text-[11.5px] font-semibold" :style="{ color: enableProcurement ? 'var(--primary-color, #6366f1)' : 'var(--txt-dim)' }">
                {{ enableProcurement ? 'Aktif' : 'Nonaktif' }}
              </span>
              <ToggleSwitch v-model="enableProcurement" />
            </div>
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

          <div v-else class="p-3.5 rounded-xl border flex items-center justify-between gap-3"
               style="background: var(--panel-2); border-color: var(--line)">
            <div class="flex items-center gap-2.5 text-[12px]" style="color: var(--txt-dim)">
              <i class="pi pi-info-circle text-indigo-500 text-sm" />
              <span>Opsi detail pengadaan (Sumber Dana, Rekanan Vendor, AKL/AKD) &amp; catatan keterangan tambahan sedang nonaktif.</span>
            </div>
            <Button label="Aktifkan Pengadaan" icon="pi pi-plus" size="small" text severity="primary" @click="enableProcurement = true" />
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
            @click="router.push(isEdit ? '/items/' + route.params.id : '/items')"
          />
          <div class="flex items-center gap-2">
            <Button
              :label="isEdit ? 'Simpan Perubahan' : 'Daftarkan Barang'"
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
