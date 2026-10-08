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
const showAdvanced = ref(false)

const form = ref<any>({
  sku: '', name: '', category: '', location: '',
  unit: 'buah', current_stock: 0, min_stock: 0, price_per_unit: 0,
  description: '', merk: '', type_model: '', serial_number: '',
  procurement_year: '', condition_status: 'Berfungsi', funding_source: '',
  distributor: '', akl_akd: '', photo_url: '',
  geo_lat: null, geo_lng: null, geo_acc: null, geo_name: '',
})

const categories = ref<any[]>([])
const locations = ref<any[]>([])
const unitOptions = ['buah', 'unit', 'set', 'box', 'rim', 'pack', 'dus', 'botol', 'roll', 'lembar']
const conditionOptions = ['Berfungsi', 'Rusak Ringan', 'Rusak Berat', 'Perlu Kalibrasi']
const yearOptions = Array.from({ length: 16 }, (_, i) => String(new Date().getFullYear() - i))

onMounted(async () => {
  try {
    const [c, l] = await Promise.all([api.get('/categories'), api.get('/locations')])
    categories.value = c.data
    locations.value = l.data
    if (isEdit) {
      const res = await api.get(`/items/${route.params.id}`)
      Object.assign(form.value, res.data)
      form.value.sku = res.data.sku ?? ''
      // If editing and has advanced fields filled, expand advanced section
      if (res.data.merk || res.data.serial_number || res.data.funding_source || res.data.akl_akd) {
        showAdvanced.value = true
      }
    }
  } catch {
    toast.add({ severity: 'error', summary: 'Gagal memuat data barang', life: 3500 })
  } finally { loadingData.value = false }
})

const skuPreview = computed(() => {
  if (form.value.sku) return form.value.sku
  const cat = categories.value.find(c => c.name === form.value.category)
  if (!cat) return '— pilih kategori dulu —'
  const prefix = (cat.id || '').split('-')[0].slice(0, 4).toUpperCase()
  return `${prefix}-${new Date().getFullYear()}-###`
})

const dirty = computed(() => !!form.value.name && !!form.value.category && !!form.value.location)

async function submit() {
  if (!dirty.value) {
    toast.add({ severity: 'warn', summary: 'Lengkapi field wajib', detail: 'Nama, kategori, dan lokasi wajib diisi', life: 3000 })
    return
  }
  loading.value = true
  try {
    if (isEdit) {
      await api.put(`/items/${route.params.id}`, form.value)
      toast.add({ severity: 'success', summary: 'Perubahan disimpan', detail: form.value.name, life: 2500 })
      router.push(`/items/${route.params.id}`)
    } else {
      const res = await api.post('/items', form.value)
      toast.add({ severity: 'success', summary: 'Barang ditambahkan', detail: `SKU ${res.data.sku}`, life: 3000 })
      router.push(`/items/${res.data.id}`)
    }
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Gagal menyimpan', detail: e.response?.data?.error || e.message, life: 4000 })
  } finally { loading.value = false }
}
</script>

<template>
  <div class="pb-24 max-w-5xl mx-auto">
    <!-- Page Header -->
    <PageHeader
      :crumb="isEdit ? 'Barang · Edit' : 'Barang · Baru'"
      :title="isEdit ? 'Edit Data Barang' : 'Tambah Barang Baru'"
      :sub="isEdit ? 'Perbarui identitas, stok, atau spesifikasi aset' : 'Lengkapi data pokok barang — foto otomatis dicap geotag GPS'">
      <template #actions>
        <Button label="Batal" icon="pi pi-times" size="small" text severity="secondary" @click="router.push(isEdit ? '/items/' + route.params.id : '/items')" />
        <Button :label="isEdit ? 'Simpan Perubahan' : 'Simpan Barang'" icon="pi pi-save" size="small"
                :loading="loading" :disabled="!dirty || loadingData" @click="submit" />
      </template>
    </PageHeader>

    <div v-if="loadingData" class="panel p-16 grid place-items-center">
      <i class="pi pi-spin pi-spinner text-2xl text-acc-500" />
    </div>

    <!-- Responsive Mobile-First Layout -->
    <form v-else @submit.prevent="submit" class="grid lg:grid-cols-[340px_1fr] gap-4 items-start">
      <!-- Left Column on Desktop / Top Column on Mobile: Foto & Kamera -->
      <div class="flex flex-col gap-4 order-first lg:order-last lg:sticky lg:top-4">
        <Panel title="Foto Fisik &amp; Geotag" icon="pi pi-camera">
          <PhotoUploader
            v-model="form.photo_url"
            v-model:geo-lat="form.geo_lat"
            v-model:geo-lng="form.geo_lng"
            v-model:geo-acc="form.geo_acc"
            v-model:geo-name="form.geo_name"
            :location-name="form.location"
            label="Foto Barang Langsung"
            hint="Cap GPS Map Camera otomatis"
          />
        </Panel>

        <!-- SKU Live Preview Card -->
        <div class="panel p-3.5 rounded-lg border text-[12px] flex flex-col gap-1.5"
             style="background: var(--paper-1); border-color: var(--line)">
          <div class="text-[10.5px] font-bold text-ink-400 uppercase tracking-wider">Preview Label SKU</div>
          <div class="t-mono font-bold text-[14px] text-acc-400">
            {{ skuPreview }}
          </div>
          <p class="text-[11px] text-ink-400 leading-tight">
            {{ form.sku ? 'SKU manual kustom' : 'SKU otomatis di-generate berdasarkan kategori dan tahun berjalan' }}
          </p>
        </div>
      </div>

      <!-- Main Column: Clean Form Fields -->
      <div class="flex flex-col gap-4">
        <!-- 1. Data Pokok Barang (Wajib & Pokok) -->
        <Panel title="Data Pokok Barang" icon="pi pi-box">
          <div class="grid sm:grid-cols-2 gap-3.5 text-[12.5px]">
            <Field label="Nama Barang" required span>
              <InputText v-model="form.name" required placeholder="mis. Tensimeter Digital / Laptop Kantor" class="w-full !text-[13px]" />
            </Field>

            <Field label="Kategori Barang" required>
              <Select v-model="form.category" :options="categories" optionLabel="name" optionValue="name"
                      required placeholder="Pilih kategori" class="w-full !text-[12.5px]" filter />
            </Field>

            <Field label="Ruangan / Lokasi Penyimpanan" required>
              <Select v-model="form.location" :options="locations" optionLabel="name" optionValue="name"
                      required placeholder="Pilih ruangan" class="w-full !text-[12.5px]" filter />
            </Field>

            <Field label="Satuan Unit" required>
              <Select v-model="form.unit" :options="unitOptions" class="w-full !text-[12.5px]" />
            </Field>

            <Field label="Kondisi Fisik">
              <Select v-model="form.condition_status" :options="conditionOptions" class="w-full !text-[12.5px]" />
            </Field>

            <Field label="SKU (Opsional)" hint="Kosongkan untuk auto-generate">
              <InputText v-model="form.sku" :placeholder="skuPreview" class="w-full t-mono !text-[12px]" />
            </Field>

            <Field label="Harga Satuan (Rp)">
              <InputNumber v-model="form.price_per_unit" :min="0" mode="currency" currency="IDR"
                           locale="id-ID" class="w-full !text-[12.5px]" />
            </Field>
          </div>
        </Panel>

        <!-- 2. Stok Barang -->
        <Panel title="Stok Inventaris" icon="pi pi-database">
          <div class="grid sm:grid-cols-2 gap-3.5 text-[12.5px]">
            <Field label="Jumlah Stok Saat Ini">
              <InputNumber v-model="form.current_stock" :min="0" showButtons class="w-full" />
            </Field>

            <Field label="Batas Minimum Stok" hint="Peringatan merah bila stok ≤ angka ini">
              <InputNumber v-model="form.min_stock" :min="0" showButtons class="w-full" />
            </Field>
          </div>
        </Panel>

        <!-- 3. Spesifikasi Lanjutan & Pengadaan (Collapsible / Rapi) -->
        <Panel title="Spesifikasi &amp; Dokumen Pengadaan" icon="pi pi-clipboard">
          <template #actions>
            <Button
              :label="showAdvanced ? 'Tutup Spesifikasi' : '+ Lengkapi Detail Pengadaan'"
              :icon="showAdvanced ? 'pi pi-chevron-up' : 'pi pi-chevron-down'"
              size="small"
              text
              severity="secondary"
              @click="showAdvanced = !showAdvanced"
            />
          </template>

          <div v-if="!showAdvanced" class="py-2 text-[12px] text-ink-400 flex items-center justify-between">
            <span>Merk, No. Seri, Tahun, Sumber Dana, dan AKL/AKD bersifat opsional.</span>
            <Button label="Buka Kolom Spesifikasi" size="small" text @click="showAdvanced = true" />
          </div>

          <div v-else class="grid sm:grid-cols-2 gap-3.5 text-[12.5px] pt-1">
            <Field label="Merk / Pabrikan">
              <InputText v-model="form.merk" placeholder="mis. Omron / Lenovo" class="w-full" />
            </Field>

            <Field label="Model / Tipe">
              <InputText v-model="form.type_model" placeholder="mis. HEM-7120 / ThinkPad E14" class="w-full" />
            </Field>

            <Field label="Nomor Seri (Serial Number)">
              <InputText v-model="form.serial_number" placeholder="Nomor seri resmi dari pabrik" class="w-full t-mono" />
            </Field>

            <Field label="Tahun Pengadaan">
              <Select v-model="form.procurement_year" :options="yearOptions" showClear placeholder="Pilih tahun" class="w-full" />
            </Field>

            <Field label="Sumber Dana">
              <InputText v-model="form.funding_source" placeholder="mis. APBD / BLU / DAK" class="w-full" />
            </Field>

            <Field label="Distributor / Vendor Penyedia">
              <InputText v-model="form.distributor" placeholder="Nama PT atau rekanan pengadaan" class="w-full" />
            </Field>

            <Field label="Izin Edar AKL / AKD" span>
              <InputText v-model="form.akl_akd" placeholder="mis. KEMENKES RI AKL 20501812345" class="w-full" />
            </Field>

            <Field label="Keterangan / Deskripsi Tambahan" span>
              <Textarea v-model="form.description" rows="2" autoResize class="w-full"
                        placeholder="Kelengkapan aksesoris, catatan khusus, garansi, dll." />
            </Field>
          </div>
        </Panel>
      </div>

      <!-- Sticky Mobile Bottom Bar -->
      <div class="fixed bottom-0 inset-x-0 z-30 p-3 border-t bg-[var(--paper-1)]/95 backdrop-blur-md shadow-2xl flex items-center justify-between gap-3 max-w-5xl mx-auto"
           style="border-color: var(--line)">
        <Button
          label="Batal"
          size="small"
          severity="secondary"
          text
          @click="router.push(isEdit ? '/items/' + route.params.id : '/items')"
        />

        <div class="flex items-center gap-2">
          <Button
            :label="isEdit ? 'Simpan Perubahan' : 'Simpan Barang'"
            icon="pi pi-save"
            size="small"
            severity="primary"
            :loading="loading"
            :disabled="!dirty || loadingData"
            @click="submit"
          />
        </div>
      </div>
    </form>
  </div>
</template>
