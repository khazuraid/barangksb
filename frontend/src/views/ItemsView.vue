<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import api from '@/api'
import { useRouter } from 'vue-router'
import { useConfirm } from 'primevue/useconfirm'
import { useToast } from 'primevue/usetoast'
import Card from 'primevue/card'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Tag from 'primevue/tag'
import Paginator from 'primevue/paginator'

const router = useRouter()
const confirm = useConfirm()
const toast = useToast()
const items = ref([])
const total = ref(0)
const page = ref(0)
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
    const res = await api.get('/items', { params: { q: query.value, cat: category.value, loc: location.value, page: page.value + 1, per_page: perPage.value } })
    items.value = res.data.data
    total.value = res.data.total
  } finally { loading.value = false }
}

async function fetchFilters() {
  const [cats, locs] = await Promise.all([api.get('/categories'), api.get('/locations')])
  categories.value = cats.data.map((c: any) => ({ label: c.name, value: c.name }))
  locations.value = locs.data.map((l: any) => ({ label: l.name, value: l.name }))
}

function deleteItem(id: string, name: string) {
  confirm.require({
    message: `Hapus barang "${name}"?`,
    header: 'Konfirmasi',
    icon: 'pi pi-exclamation-triangle',
    acceptClass: 'p-button-danger',
    accept: () => {
      api.delete(`/items/${id}`).then(() => {
        toast.add({ severity: 'success', summary: 'Dihapus', detail: `${name} berhasil dihapus`, life: 2000 })
        fetchItems()
      })
    },
  })
}

onMounted(() => { fetchItems(); fetchFilters() })
watch([page, perPage], () => fetchItems())
</script>

<template>
  <div class="flex flex-col gap-4">
    <div class="flex justify-between items-center">
      <h2 class="text-xl font-bold">Daftar Barang ({{ total }})</h2>
      <Button label="Tambah Barang" icon="pi pi-plus" @click="router.push('/items/new')" />
    </div>

    <!-- Filter bar -->
    <Card class="shadow-sm">
      <template #content>
        <div class="flex flex-wrap gap-3 items-end">
          <div class="flex-1 min-w-[200px]">
            <label class="block text-xs font-medium text-gray-500 mb-1">Cari</label>
            <InputText v-model="query" @keyup.enter="fetchItems" placeholder="Nama / SKU / lokasi" class="w-full" />
          </div>
          <div class="min-w-[180px]">
            <label class="block text-xs font-medium text-gray-500 mb-1">Kategori</label>
            <Select v-model="category" :options="categories" optionLabel="label" optionValue="value" showClear placeholder="Semua" class="w-full" @change="fetchItems" />
          </div>
          <div class="min-w-[180px]">
            <label class="block text-xs font-medium text-gray-500 mb-1">Lokasi</label>
            <Select v-model="location" :options="locations" optionLabel="label" optionValue="value" showClear placeholder="Semua" class="w-full" @change="fetchItems" />
          </div>
          <Button label="Cari" icon="pi pi-search" @click="fetchItems" />
        </div>
      </template>
    </Card>

    <!-- Table -->
    <Card class="shadow-sm">
      <template #content>
        <DataTable :value="items" :loading="loading" responsiveLayout="scroll" class="p-datatable-sm" stripedRows>
          <Column field="sku" header="SKU" style="width: 120px">
            <template #body="{ data }"><code class="text-xs font-mono bg-gray-100 dark:bg-gray-800 px-2 py-0.5 rounded">{{ data.sku }}</code></template>
          </Column>
          <Column field="name" header="Nama">
            <template #body="{ data }"><strong>{{ data.name }}</strong></template>
          </Column>
          <Column field="category" header="Kategori" style="width: 150px">
            <template #body="{ data }"><span class="text-gray-500 text-sm">{{ data.category }}</span></template>
          </Column>
          <Column field="current_stock" header="Stok" style="width: 100px">
            <template #body="{ data }">
              <Tag :severity="data.current_stock <= data.min_stock ? 'danger' : 'success'" :value="`${data.current_stock} ${data.unit}`" />
            </template>
          </Column>
          <Column header="Aksi" style="width: 120px">
            <template #body="{ data }">
              <div class="flex gap-1">
                <Button icon="pi pi-pencil" text rounded size="small" @click="router.push(`/items/${data.id}/edit`)" />
                <Button icon="pi pi-trash" text rounded size="small" severity="danger" @click="deleteItem(data.id, data.name)" />
              </div>
            </template>
          </Column>
        </DataTable>
        <Paginator :rows="perPage" :totalRecords="total" v-model:first="page" :rowsPerPageOptions="[20, 25, 50, 100]" @page="(e: any) => { perPage = e.rows; page = e.page }" />
      </template>
    </Card>
  </div>
</template>
