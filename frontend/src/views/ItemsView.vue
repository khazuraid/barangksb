<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import api from '@/api'
import { useRouter } from 'vue-router'

const router = useRouter()
const items = ref([])
const total = ref(0)
const page = ref(1)
const perPage = ref(25)
const query = ref('')
const category = ref('')
const location = ref('')
const categories = ref([])
const locations = ref([])
const loading = ref(false)

async function fetchItems() {
  loading.value = true
  try {
    const res = await api.get('/items', { params: { q: query.value, cat: category.value, loc: location.value, page: page.value, per_page: perPage.value } })
    items.value = res.data.data
    total.value = res.data.total
  } finally { loading.value = false }
}

async function fetchFilters() {
  const [cats, locs] = await Promise.all([api.get('/categories'), api.get('/locations')])
  categories.value = cats.data
  locations.value = locs.data
}

function deleteItem(id: string) {
  if (!confirm('Hapus barang ini?')) return
  api.delete(`/items/${id}`).then(() => fetchItems())
}

const pages = () => Math.ceil(total.value / perPage.value)

onMounted(() => { fetchItems(); fetchFilters() })
watch([page, perPage], () => fetchItems())
</script>

<template>
  <div class="space-y-4">
    <div class="flex justify-between items-center">
      <h2 class="text-xl font-bold">Daftar Barang ({{ total }})</h2>
      <button class="btn btn-primary btn-sm" @click="router.push('/items/new')">+ Tambah</button>
    </div>

    <div class="flex flex-wrap gap-2">
      <input v-model="query" @keyup.enter="fetchItems" type="search" placeholder="Cari..." class="input input-bordered input-sm" />
      <select v-model="category" @change="fetchItems" class="select select-bordered select-sm">
        <option value="">Semua Kategori</option>
        <option v-for="c in categories" :key="c.id" :value="c.name">{{ c.name }}</option>
      </select>
      <select v-model="location" @change="fetchItems" class="select select-bordered select-sm">
        <option value="">Semua Lokasi</option>
        <option v-for="l in locations" :key="l.id" :value="l.name">{{ l.name }}</option>
      </select>
      <button class="btn btn-sm btn-primary" @click="fetchItems">Filter</button>
    </div>

    <div class="card bg-base-100 shadow-sm border border-base-200">
      <div class="overflow-x-auto">
        <table class="table table-sm">
          <thead><tr><th>SKU</th><th>Nama</th><th>Kategori</th><th>Stok</th><th></th></tr></thead>
          <tbody>
            <tr v-for="item in items" :key="item.id">
              <td><code class="text-xs">{{ item.sku }}</code></td>
              <td class="font-medium">{{ item.name }}</td>
              <td class="text-base-content/50">{{ item.category }}</td>
              <td><span class="badge badge-sm" :class="item.current_stock <= item.min_stock ? 'badge-error' : 'badge-success'">{{ item.current_stock }} {{ item.unit }}</span></td>
              <td>
                <div class="flex gap-1">
                  <button class="btn btn-xs btn-ghost" @click="router.push(`/items/${item.id}/edit`)">Edit</button>
                  <button class="btn btn-xs btn-ghost text-error" @click="deleteItem(item.id)">Hapus</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div class="card-actions justify-between items-center p-4">
        <div class="join">
          <button v-for="p in pages()" :key="p" class="join-item btn btn-sm" :class="{ 'btn-active': page === p }" @click="page = p">{{ p }}</button>
        </div>
        <select v-model="perPage" class="select select-bordered select-sm">
          <option :value="20">20/hal</option><option :value="25">25/hal</option><option :value="50">50/hal</option><option :value="100">100/hal</option>
        </select>
      </div>
    </div>
  </div>
</template>
