<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '@/api'

const route = useRoute()
const router = useRouter()
const isEdit = !!route.params.id
const form = ref({ sku: '', name: '', category: '', location: '', unit: 'Unit', current_stock: 0, min_stock: 0, price_per_unit: 0, merk: '', condition_status: 'Berfungsi' })
const categories = ref([])
const locations = ref([])

onMounted(async () => {
  const [cats, locs] = await Promise.all([api.get('/categories'), api.get('/locations')])
  categories.value = cats.data; locations.value = locs.data
  if (isEdit) {
    const res = await api.get(`/items/${route.params.id}`)
    Object.assign(form.value, res.data)
  }
})

async function submit() {
  if (isEdit) {
    await api.put(`/items/${route.params.id}`, form.value)
  } else {
    await api.post('/items', form.value)
  }
  router.push('/items')
}
</script>

<template>
  <div class="card bg-base-100 shadow-sm border border-base-200 max-w-2xl">
    <div class="card-body">
      <h2 class="card-title">{{ isEdit ? 'Edit' : 'Tambah' }} Barang</h2>
      <form @submit.prevent="submit" class="grid grid-cols-2 gap-3">
        <label class="form-control"><span class="label-text">SKU</span><input v-model="form.sku" class="input input-bordered input-sm" placeholder="otomatis" /></label>
        <label class="form-control"><span class="label-text">Nama *</span><input v-model="form.name" required class="input input-bordered input-sm" /></label>
        <label class="form-control"><span class="label-text">Kategori *</span><select v-model="form.category" required class="select select-bordered select-sm"><option value="">—</option><option v-for="c in categories" :key="c.id" :value="c.name">{{ c.name }}</option></select></label>
        <label class="form-control"><span class="label-text">Lokasi *</span><select v-model="form.location" required class="select select-bordered select-sm"><option value="">—</option><option v-for="l in locations" :key="l.id" :value="l.name">{{ l.name }}</option></select></label>
        <label class="form-control"><span class="label-text">Stok</span><input v-model.number="form.current_stock" type="number" class="input input-bordered input-sm" /></label>
        <label class="form-control"><span class="label-text">Min Stok</span><input v-model.number="form.min_stock" type="number" class="input input-bordered input-sm" /></label>
        <label class="form-control"><span class="label-text">Satuan</span><select v-model="form.unit" class="select select-bordered select-sm"><option>Unit</option><option>Pcs</option><option>Set</option><option>Box</option></select></label>
        <label class="form-control"><span class="label-text">Harga (Rp)</span><input v-model.number="form.price_per_unit" type="number" class="input input-bordered input-sm" /></label>
        <label class="form-control"><span class="label-text">Merk</span><input v-model="form.merk" class="input input-bordered input-sm" /></label>
        <label class="form-control"><span class="label-text">Kondisi</span><select v-model="form.condition_status" class="select select-bordered select-sm"><option>Berfungsi</option><option>Rusak Ringan</option><option>Rusak Berat</option></select></label>
        <div class="col-span-2 flex gap-2 mt-2"><button type="submit" class="btn btn-primary btn-sm">Simpan</button><button class="btn btn-ghost btn-sm" @click="router.push('/items')">Batal</button></div>
      </form>
    </div>
  </div>
</template>
