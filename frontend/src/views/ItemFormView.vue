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

const form = ref<any>({
  sku: '',
  name: '',
  category: '',
  location: '',
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
  condition_status: 'Berfungsi',
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
      form.value.track_stock = res.data.track_stock !== false
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
  loading.value = true
  try {
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
  <div class="pb-20 max-w-5xl mx-auto">
    <!-- Page Header -->
    <PageHeader
      :crumb="isEdit ? 'Barang · Edit' : 'Barang · Baru'"
      :title="isEdit ? 'Edit Data Barang' : 'Tambah Barang Baru'"
      :sub="isEdit ? 'Perbarui informasi identitas atau spesifikasi aset' : 'Daftarkan barang inventaris — foto otomatis dicap geotag GPS Map Camera'"
    >
      <template #actions>
        <Button
          label="Batal"
          icon="pi pi-times"
          size="small"
          text
          severity="secondary"
          @click="router.push(isEdit ? '/items/' + route.params.id : '/items')"
        />
        <Button
          :label="isEdit ? 'Simpan Perubahan' : 'Simpan Barang'"
          icon="pi pi-save"
          size="small"
          :loading="loading"
          :disabled="!dirty || loadingData"
          @click="submit"
        />
      </template>
    </PageHeader>

    <div v-if="loadingData" class="panel p-16 grid place-items-center">
      <i class="pi pi-spin pi-spinner text-2xl text-acc-500" />
    </div>

    <!-- Symmetrical & Clean Layout -->
    <form v-else @submit.prevent="submit" class="grid lg:grid-cols-[380px_1fr] gap-5 items-start">
      <!-- Left Column: Foto & Preview SKU -->
      <div class="flex flex-col gap-4">
        <Panel title="Foto Barang &amp; Geotag" icon="pi pi-camera">
          <PhotoUploader
            v-model="form.photo_url"
            v-model:geo-lat="form.geo_lat"
            v-model:geo-lng="form.geo_lng"
            v-model:geo-acc="form.geo_acc"
            v-model:geo-name="form.geo_name"
            :location-name="form.location"
            label="Foto Fisik Barang"
            hint="Cap GPS Map Camera otomatis"
          />
        </Panel>

        <!-- Preview Label SKU -->
        <div
          class="p-4 rounded-xl border flex flex-col gap-1.5 shadow-xs"
          style="background: var(--paper-1); border-color: var(--line)"
        >
          <div class="text-[10.5px] font-bold text-ink-400 uppercase tracking-wider">Preview Label SKU</div>
          <div class="t-mono font-bold text-[15px] text-acc-500">
            {{ skuPreview }}
          </div>
          <p class="text-[11px] text-ink-400 leading-tight">
            {{ form.sku ? 'SKU manual kustom' : 'SKU otomatis dibuat berdasarkan kategori dan tahun' }}
          </p>
        </div>
      </div>

      <!-- Right Column: Identitas, Pilihan Tipe, & Spesifikasi -->
      <div class="flex flex-col gap-4">
        <!-- 1. Tipe Pengelolaan Inventaris (Barang Stok vs Aset Tetap) -->
        <Panel title="Jenis Pengelolaan Inventaris" icon="pi pi-sliders-h">
          <div class="flex flex-col gap-3">
            <span class="text-[11.5px] text-ink-300">
              Pilih bagaimana barang ini dikelola dalam inventaris:
            </span>

            <div class="grid sm:grid-cols-2 gap-3">
              <!-- Opsi A: Barang Habis Pakai / Stok (Pensil, Kertas, Tinta) -->
              <div
                class="p-3.5 rounded-xl border cursor-pointer transition-all flex items-start gap-3"
                :class="form.track_stock ? 'bg-acc-500/10 border-acc-500/60 shadow-xs' : 'bg-paper-2 border-line hover:border-line-soft'"
                @click="form.track_stock = true"
              >
                <input
                  type="radio"
                  :checked="form.track_stock"
                  class="mt-1 accent-amber-500 cursor-pointer"
                  @change="form.track_stock = true"
                />
                <div class="flex flex-col">
                  <span class="font-bold text-[13px] text-ink-100 flex items-center gap-1.5">
                    <i class="pi pi-box text-acc-500 text-xs" /> Barang Stok / Konsumabel
                  </span>
                  <span class="text-[11px] text-ink-400 mt-0.5 leading-snug">
                    Contoh: Pensil, Kertas, Tinta, Sarung Tangan. Stok ditambah via menu <strong>Stok Masuk</strong> ber-geotag.
                  </span>
                </div>
              </div>

              <!-- Opsi B: Aset Tetap / Unit Mandiri (Laptop, Alkes, Kendaraan) -->
              <div
                class="p-3.5 rounded-xl border cursor-pointer transition-all flex items-start gap-3"
                :class="!form.track_stock ? 'bg-acc-500/10 border-acc-500/60 shadow-xs' : 'bg-paper-2 border-line hover:border-line-soft'"
                @click="form.track_stock = false"
              >
                <input
                  type="radio"
                  :checked="!form.track_stock"
                  class="mt-1 accent-amber-500 cursor-pointer"
                  @change="form.track_stock = false"
                />
                <div class="flex flex-col">
                  <span class="font-bold text-[13px] text-ink-100 flex items-center gap-1.5">
                    <i class="pi pi-desktop text-acc-500 text-xs" /> Aset Tetap / Unit Mandiri
                  </span>
                  <span class="text-[11px] text-ink-400 mt-0.5 leading-snug">
                    Contoh: Laptop, Komputer, Kendaraan, Mesin. Setiap unit adalah 1 barang tersendiri.
                  </span>
                </div>
              </div>
            </div>

            <!-- Notice box -->
            <div
              v-if="form.track_stock && !isEdit"
              class="p-2.5 rounded-lg border text-[11.5px] text-acc-400/90 flex items-center gap-2"
              style="background: rgba(245, 165, 36, 0.08); border-color: rgba(245, 165, 36, 0.25)"
            >
              <i class="pi pi-info-circle text-xs shrink-0" />
              <span>
                Barang akan didaftarkan dengan stok awal 0. Penambahan stok dilakukan di halaman detail barang melalui tombol <strong>+ Stok Masuk</strong> dengan bukti foto geotag.
              </span>
            </div>
          </div>
        </Panel>

        <!-- 2. Data Pokok Barang -->
        <Panel title="Data Pokok Barang" icon="pi pi-tag">
          <div class="grid sm:grid-cols-2 gap-3.5 text-[12.5px]">
            <Field label="Nama Barang" required span>
              <InputText v-model="form.name" required placeholder="mis. Pensil 2B Joyko / Laptop Lenovo ThinkPad" class="w-full" />
            </Field>

            <Field label="Kategori Barang" required>
              <Select
                v-model="form.category"
                :options="categories"
                optionLabel="name"
                optionValue="name"
                required
                placeholder="Pilih kategori"
                class="w-full"
                filter
              />
            </Field>

            <Field label="Ruangan / Lokasi Penyimpanan" required>
              <Select
                v-model="form.location"
                :options="locations"
                optionLabel="name"
                optionValue="name"
                required
                placeholder="Pilih ruangan"
                class="w-full"
                filter
              />
            </Field>

            <Field label="Satuan Unit" required>
              <Select v-model="form.unit" :options="unitOptions" class="w-full" />
            </Field>

            <Field label="Kondisi Fisik">
              <Select v-model="form.condition_status" :options="conditionOptions" class="w-full" />
            </Field>

            <!-- Batas Minimum Stok (Hanya untuk Barang Stok / Konsumabel) -->
            <Field v-if="form.track_stock" label="Batas Minimum Peringatan" hint="Peringatan bila stok tersisa ≤ angka ini">
              <InputNumber v-model="form.min_stock" :min="0" showButtons class="w-full" />
            </Field>

            <Field label="Estimasi Harga Satuan (Rp)">
              <InputNumber
                v-model="form.price_per_unit"
                :min="0"
                mode="currency"
                currency="IDR"
                locale="id-ID"
                class="w-full"
              />
            </Field>

            <Field label="SKU Kustom (Opsional)" hint="Kosongkan untuk auto-generate">
              <InputText v-model="form.sku" :placeholder="skuPreview" class="w-full t-mono" />
            </Field>
          </div>
        </Panel>

        <!-- 3. Spesifikasi & Pengadaan Aset -->
        <Panel title="Spesifikasi &amp; Detail Pengadaan" icon="pi pi-clipboard">
          <div class="grid sm:grid-cols-2 gap-3.5 text-[12.5px]">
            <Field label="Merk / Pabrikan">
              <InputText v-model="form.merk" placeholder="mis. Joyko / Omron / Lenovo" class="w-full" />
            </Field>

            <Field label="Model / Tipe">
              <InputText v-model="form.type_model" placeholder="mis. 2B / HEM-7120 / E14" class="w-full" />
            </Field>

            <Field label="Nomor Seri (Serial Number)">
              <InputText v-model="form.serial_number" placeholder="Nomor seri pabrik (jika ada)" class="w-full t-mono" />
            </Field>

            <Field label="Tahun Pengadaan">
              <Select v-model="form.procurement_year" :options="yearOptions" showClear placeholder="Pilih tahun" class="w-full" />
            </Field>

            <Field label="Sumber Dana">
              <InputText v-model="form.funding_source" placeholder="mis. APBD / BLU / DAK" class="w-full" />
            </Field>

            <Field label="Distributor / Rekanan Penyedia">
              <InputText v-model="form.distributor" placeholder="Nama rekanan penyedia" class="w-full" />
            </Field>

            <Field label="Izin Edar AKL / AKD" span>
              <InputText v-model="form.akl_akd" placeholder="mis. AKL 20501812345" class="w-full" />
            </Field>

            <Field label="Catatan / Deskripsi Tambahan" span>
              <Textarea
                v-model="form.description"
                rows="2"
                autoResize
                placeholder="Spesifikasi teknis, kelengkapan aksesoris, atau garansi"
                class="w-full"
              />
            </Field>
          </div>
        </Panel>

        <!-- Action Buttons -->
        <div class="p-4 rounded-xl border flex items-center justify-between gap-3 shadow-xs"
             style="background: var(--paper-1); border-color: var(--line)">
          <Button
            label="Batal"
            size="small"
            severity="secondary"
            text
            @click="router.push(isEdit ? '/items/' + route.params.id : '/items')"
          />
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
