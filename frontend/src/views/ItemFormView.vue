<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '@/api'
import { useToast } from 'primevue/usetoast'
import Card from 'primevue/card'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import InputNumber from 'primevue/inputnumber'
import Select from 'primevue/select'
import Textarea from 'primevue/textarea'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const isEdit = !!route.params.id
const loading = ref(false)
const form = ref({ sku: '', name: '', category: '', location: '', unit: 'Unit', current_stock: 0, min_stock: 0, price_per_unit: 0, merk: '', condition_status: 'Berfungsi' })
const categories = ref([])
const locations = ref([])

const unitOptions = ['Unit', 'Pcs', 'Set', 'Box', 'Rim', 'Pack', 'Dus', 'Botol', 'Roll']
const conditionOptions = ['Berfungsi', 'Rusak Ringan', 'Rusak Berat', 'Perlu Kalibrasi']

onMounted(async () => {
  const [cats, locs] = await Promise.all([api.get('/categories'), api.get('/locations')])
  categories.value = cats.data.map((c: any) => ({ label: c.name, value: c.name }))
  locations.value = locs.data.map((l: any) => ({ label: l.name, value: l.name }))
  if (isEdit) {
    const res = await api.get(`/items/${route.params.id}`)
    Object.assign(form.value, res.data)
  }
})

async function submit() {
  loading.value = true
  try {
    if (isEdit) {
      await api.put(`/items/${route.params.id}`, form.value)
    } else {
      await api.post('/items', form.value)
    }
    toast.add({ severity: 'success', summary: 'Berhasil', detail: 'Barang disimpan', life: 2000 })
    router.push('/items')
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Gagal', detail: e.response?.data?.error, life: 3000 })
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="max-w-2xl">
    <div class="flex items-center gap-3 mb-4">
      <Button icon="pi pi-arrow-left" text rounded @click="router.push('/items')" />
      <h2 class="text-xl font-bold">{{ isEdit ? 'Edit' : 'Tambah' }} Barang</h2>
    </div>
    <Card class="shadow-sm">
      <template #content>
        <form @submit.prevent="submit" class="flex flex-col gap-4">
          <div class="grid grid-cols-2 gap-4">
            <div><label class="block text-sm font-medium mb-1">SKU</label><InputText v-model="form.sku" class="w-full" placeholder="otomatis" /></div>
            <div><label class="block text-sm font-medium mb-1">Nama *</label><InputText v-model="form.name" required class="w-full" /></div>
            <div><label class="block text-sm font-medium mb-1">Kategori *</label><Select v-model="form.category" :options="categories" optionLabel="label" optionValue="value" required class="w-full" /></div>
            <div><label class="block text-sm font-medium mb-1">Lokasi *</label><Select v-model="form.location" :options="locations" optionLabel="label" optionValue="value" required class="w-full" /></div>
            <div><label class="block text-sm font-medium mb-1">Stok</label><InputNumber v-model="form.current_stock" :min="0" class="w-full" /></div>
            <div><label class="block text-sm font-medium mb-1">Min Stok</label><InputNumber v-model="form.min_stock" :min="0" class="w-full" /></div>
            <div><label class="block text-sm font-medium mb-1">Satuan</label><Select v-model="form.unit" :options="unitOptions" class="w-full" /></div>
            <div><label class="block text-sm font-medium mb-1">Harga (Rp)</label><InputNumber v-model="form.price_per_unit" :min="0" class="w-full" /></div>
            <div><label class="block text-sm font-medium mb-1">Merk</label><InputText v-model="form.merk" class="w-full" /></div>
            <div><label class="block text-sm font-medium mb-1">Kondisi</label><Select v-model="form.condition_status" :options="conditionOptions" class="w-full" /></div>
          </div>
          <div class="flex gap-3 mt-2">
            <Button type="submit" label="Simpan" icon="pi pi-save" :loading="loading" />
            <Button label="Batal" icon="pi pi-times" text @click="router.push('/items')" />
          </div>
        </form>
      </template>
    </Card>
  </div>
</template>
