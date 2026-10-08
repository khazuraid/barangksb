<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import api from '@/api'
import { useRouter } from 'vue-router'
import { useToast } from 'primevue/usetoast'
import Card from 'primevue/card'
import Button from 'primevue/button'
import Select from 'primevue/select'
import InputNumber from 'primevue/inputnumber'
import InputText from 'primevue/inputtext'
import Textarea from 'primevue/textarea'
import Tabs from 'primevue/tabs'
import TabList from 'primevue/tablist'
import Tab from 'primevue/tab'
import TabPanels from 'primevue/tabpanels'
import TabPanel from 'primevue/tabpanel'

const router = useRouter()
const toast = useToast()
const tab = ref('0')
const items = ref([])
const form = ref({ item_id: '', quantity: 0, received_by: '', notes: '' })

watch(tab, async () => {
  if (items.value.length === 0) {
    const res = await api.get('/items', { params: { per_page: 100 } })
    items.value = res.data.data.map((i: any) => ({ label: `${i.name} (${i.sku}) — stok ${i.current_stock}`, value: i.id }))
  }
}, { immediate: true })

async function submit() {
  const endpoint = tab.value === '1' ? '/movement/out' : tab.value === '2' ? '/movement/adjust' : '/movement/in'
  try {
    await api.post(endpoint, form.value)
    toast.add({ severity: 'success', summary: 'Berhasil', detail: 'Mutasi dicatat', life: 2000 })
    router.push('/history')
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Gagal', detail: e.response?.data?.error, life: 3000 })
  }
}
</script>

<template>
  <div class="flex flex-col gap-4">
    <h2 class="text-xl font-bold">Mutasi Barang</h2>
    <Card class="shadow-sm max-w-xl">
      <template #content>
        <Tabs v-model:value="tab">
          <TabList>
            <Tab :value="'0'"><i class="pi pi-arrow-down mr-2"></i>Barang Masuk</Tab>
            <Tab :value="'1'"><i class="pi pi-arrow-up mr-2"></i>Barang Keluar</Tab>
            <Tab :value="'2'"><i class="pi pi-sliders-h mr-2"></i>Opname</Tab>
          </TabList>
          <TabPanels>
            <TabPanel :value="'0'">
              <form @submit.prevent="submit" class="flex flex-col gap-3">
                <div><label class="block text-sm font-medium mb-1">Barang</label><Select v-model="form.item_id" :options="items" optionLabel="label" optionValue="value" required class="w-full" filter /></div>
                <div><label class="block text-sm font-medium mb-1">Jumlah Masuk</label><InputNumber v-model="form.quantity" :min="1" required class="w-full" /></div>
                <div><label class="block text-sm font-medium mb-1">Petugas</label><InputText v-model="form.received_by" class="w-full" /></div>
                <div><label class="block text-sm font-medium mb-1">Keterangan</label><Textarea v-model="form.notes" rows="2" class="w-full" /></div>
                <Button type="submit" label="Catat Masuk" icon="pi pi-arrow-down" severity="success" />
              </form>
            </TabPanel>
            <TabPanel :value="'1'">
              <form @submit.prevent="submit" class="flex flex-col gap-3">
                <div><label class="block text-sm font-medium mb-1">Barang</label><Select v-model="form.item_id" :options="items" optionLabel="label" optionValue="value" required class="w-full" filter /></div>
                <div><label class="block text-sm font-medium mb-1">Jumlah Keluar</label><InputNumber v-model="form.quantity" :min="1" required class="w-full" /></div>
                <div><label class="block text-sm font-medium mb-1">Diberikan Kepada</label><InputText v-model="form.received_by" class="w-full" /></div>
                <div><label class="block text-sm font-medium mb-1">Keterangan</label><Textarea v-model="form.notes" rows="2" class="w-full" /></div>
                <Button type="submit" label="Catat Keluar" icon="pi pi-arrow-up" severity="danger" />
              </form>
            </TabPanel>
            <TabPanel :value="'2'">
              <form @submit.prevent="submit" class="flex flex-col gap-3">
                <div><label class="block text-sm font-medium mb-1">Barang</label><Select v-model="form.item_id" :options="items" optionLabel="label" optionValue="value" required class="w-full" filter /></div>
                <div><label class="block text-sm font-medium mb-1">Stok Fisik</label><InputNumber v-model="form.quantity" :min="0" required class="w-full" /></div>
                <div><label class="block text-sm font-medium mb-1">Catatan</label><Textarea v-model="form.notes" rows="2" class="w-full" /></div>
                <Button type="submit" label="Simpan Opname" icon="pi pi-check" />
              </form>
            </TabPanel>
          </TabPanels>
        </Tabs>
      </template>
    </Card>
  </div>
</template>
