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
import Tag from 'primevue/tag'
import PhotoUploader from '@/components/PhotoUploader.vue'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const isEdit = !!route.params.id
const loading = ref(false)
const loadingData = ref(true)

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
  loading.value = true
  try {
    if (isEdit) {
      await api.put(`/items/${route.params.id}`, form.value)
      toast.add({ severity: 'success', summary: 'Perubahan disimpan', detail: form.value.name, life: 2500 })
    } else {
      const res = await api.post('/items', form.value)
      toast.add({ severity: 'success', summary: 'Barang ditambahkan', detail: `SKU ${res.data.sku}`, life: 3000 })
    }
    router.push('/items')
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Gagal menyimpan', detail: e.response?.data?.error, life: 4000 })
  } finally { loading.value = false }
}
</script>

<template>
  <div>
    <PageHeader
      :crumb="isEdit ? 'Barang · Edit' : 'Barang · Baru'"
      :title="isEdit ? 'Edit Barang' : 'Tambah Barang'"
      :sub="isEdit ? 'Perbarui data master barang ini' : 'Isi data barang baru — SKU dibuat otomatis bila dikosongkan'">
      <template #actions>
        <Button label="Batal" icon="pi pi-times" size="small" text severity="secondary" @click="router.push('/items')" />
        <Button :label="isEdit ? 'Simpan Perubahan' : 'Simpan Barang'" icon="pi pi-save" size="small"
                :loading="loading" :disabled="!dirty || loadingData" @click="submit" />
      </template>
    </PageHeader>

    <div v-if="loadingData" class="grid place-items-center py-20">
      <i class="pi pi-spin pi-spinner text-xl" style="color: var(--txt-dim)" />
    </div>

    <form v-else @submit.prevent="submit" class="grid xl:grid-cols-[1fr_300px] gap-4 items-start">
      <div class="flex flex-col gap-4">
        <Panel title="Identitas Barang" icon="pi pi-tag">
          <div class="grid sm:grid-cols-2 gap-4">
            <Field label="Nama Barang" required span>
              <InputText v-model="form.name" required placeholder="mis. Tensimeter Digital" class="w-full" />
            </Field>
            <Field label="Kategori" required>
              <Select v-model="form.category" :options="categories" optionLabel="name" optionValue="name"
                      required placeholder="Pilih kategori" class="w-full" filter />
            </Field>
            <Field label="Lokasi Penyimpanan" required>
              <Select v-model="form.location" :options="locations" optionLabel="name" optionValue="name"
                      required placeholder="Pilih lokasi" class="w-full" filter />
            </Field>
            <Field label="SKU" :hint="isEdit ? 'Ubah hanya bila perlu — SKU dipakai pada label QR.' : 'Kosongkan untuk membuat otomatis.'">
              <InputText v-model="form.sku" :placeholder="skuPreview" class="w-full t-mono" />
            </Field>
            <Field label="Satuan">
              <Select v-model="form.unit" :options="unitOptions" class="w-full" />
            </Field>
            <Field label="Merk">
              <InputText v-model="form.merk" placeholder="mis. Omron" class="w-full" />
            </Field>
            <Field label="Model / Tipe">
              <InputText v-model="form.type_model" class="w-full" />
            </Field>
          </div>
        </Panel>

        <Panel title="Stok &amp; Nilai" icon="pi pi-database">
          <div class="grid sm:grid-cols-3 gap-4">
            <Field label="Stok Saat Ini">
              <InputNumber v-model="form.current_stock" :min="0" showButtons class="w-full" />
            </Field>
            <Field label="Stok Minimum" hint="Baris ditandai merah bila stok ≤ nilai ini.">
              <InputNumber v-model="form.min_stock" :min="0" showButtons class="w-full" />
            </Field>
            <Field label="Harga Satuan (Rp)">
              <InputNumber v-model="form.price_per_unit" :min="0" mode="currency" currency="IDR"
                           locale="id-ID" class="w-full" />
            </Field>
          </div>
        </Panel>

        <Panel title="Kondisi &amp; Pengadaan" icon="pi pi-clipboard">
          <div class="grid sm:grid-cols-3 gap-4">
            <Field label="Kondisi">
              <Select v-model="form.condition_status" :options="conditionOptions" class="w-full" />
            </Field>
            <Field label="Tahun Pengadaan">
              <Select v-model="form.procurement_year" :options="yearOptions" showClear placeholder="Tahun" class="w-full" />
            </Field>
            <Field label="Sumber Dana">
              <InputText v-model="form.funding_source" placeholder="mis. APBD / BLU" class="w-full" />
            </Field>
            <Field label="Distributor / Penyedia">
              <InputText v-model="form.distributor" class="w-full" />
            </Field>
            <Field label="No. Serial">
              <InputText v-model="form.serial_number" class="w-full t-mono" />
            </Field>
            <Field label="AKL / AKD">
              <InputText v-model="form.akl_akd" placeholder="mis. AKL 1234567890" class="w-full" />
            </Field>
            <Field label="Catatan / Deskripsi" span>
              <Textarea v-model="form.description" rows="3" autoResize class="w-full"
                        placeholder="Spesifikasi, kelengkapan, atau keterangan lain" />
            </Field>
          </div>
        </Panel>
      </div>

      <div class="flex flex-col gap-4 xl:sticky xl:top-4">
        <Panel title="Foto &amp; Geotag" icon="pi pi-camera">
          <PhotoUploader
            v-model="form.photo_url"
            v-model:geo-lat="form.geo_lat"
            v-model:geo-lng="form.geo_lng"
            v-model:geo-acc="form.geo_acc"
            v-model:geo-name="form.geo_name"
            :location-name="form.location"
            label="Foto Fisik Barang"
            hint="Cap GPS &amp; waktu otomatis"
          />
        </Panel>

        <Panel title="Ringkasan" icon="pi pi-eye">
          <div class="flex flex-col gap-3">
            <div>
              <div class="t-label mb-1">SKU</div>
              <div class="t-mono font-semibold" :class="!form.sku ? 'text-acc-600' : ''">{{ skuPreview }}</div>
            </div>
            <div class="pt-3 border-t" style="border-color: var(--line)">
              <div class="t-label mb-1">Nama</div>
              <div class="text-[12.5px] font-semibold">{{ form.name || '—' }}</div>
            </div>
            <div>
              <div class="t-label mb-1">Penempatan</div>
              <div class="text-[12.5px]">{{ form.category || '—' }}</div>
              <div class="text-[12px]" style="color: var(--txt-dim)">{{ form.location || '—' }}</div>
            </div>
            <div class="pt-3 border-t flex items-end justify-between" style="border-color: var(--line)">
              <div>
                <div class="t-label mb-1">Stok / Min</div>
                <div class="t-num text-[20px] font-bold"
                     :class="form.current_stock <= form.min_stock ? 'text-sig-bad' : 'text-sig-ok'">
                  {{ form.current_stock }}<span class="text-[13px] font-normal" style="color: var(--txt-dim)"> / {{ form.min_stock }}</span>
                </div>
              </div>
              <span class="text-[11px]" style="color: var(--txt-dim)">{{ form.unit }}</span>
            </div>
            <Tag v-if="form.current_stock <= form.min_stock" severity="danger"
                 class="w-full justify-center mt-3.5"
                 icon="pi pi-exclamation-triangle" value="Stok di bawah minimum" />
          </div>
        </Panel>

        <Panel title="Checklist" icon="pi pi-check-square">
          <ul class="flex flex-col gap-2 text-[12px]">
            <li v-for="c in [
              { ok: !!form.name, t: 'Nama barang terisi' },
              { ok: !!form.category, t: 'Kategori dipilih' },
              { ok: !!form.location, t: 'Lokasi dipilih' },
              { ok: !!form.unit, t: 'Satuan ditentukan' },
              { ok: form.min_stock >= 0, t: 'Batas minimum wajar' },
            ]" :key="c.t" class="flex items-center gap-2.5">
              <i class="pi text-[11px]" :class="c.ok ? 'pi-check-circle text-sig-ok' : 'pi-circle text-ink-300'" />
              <span :style="c.ok ? '' : 'color: var(--txt-dim)'">{{ c.t }}</span>
            </li>
          </ul>
        </Panel>

        <Button :label="isEdit ? 'Simpan Perubahan' : 'Simpan Barang'" icon="pi pi-save"
                :loading="loading" :disabled="!dirty" @click="submit" class="w-full" />
      </div>
    </form>
  </div>
</template>