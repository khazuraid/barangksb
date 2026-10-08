<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api'
import Card from 'primevue/card'
import Button from 'primevue/button'
import Checkbox from 'primevue/checkbox'
import { useToast } from 'primevue/usetoast'
const toast = useToast()
const items = ref([]); const selected = ref<string[]>([])
onMounted(async () => { const res = await api.get('/items', { params: { per_page: 100 } }); items.value = res.data.data })
function printSheet() {
  if (!selected.value.length) { toast.add({severity:'warn',summary:'Pilih barang',life:2000}); return }
  window.open(`/api/barcode/sheet?ids=${selected.value.join(',')}`, '_blank')
}
</script>
<template>
  <div class="flex flex-col gap-4">
    <div class="flex justify-between items-center"><h2 class="text-xl font-bold">QR Generator</h2><Button label="Cetak Terpilih" icon="pi pi-print" @click="printSheet" /></div>
    <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-3">
      <Card v-for="item in items" :key="item.id" class="shadow-sm cursor-pointer" @click="() => { const idx = selected.indexOf(item.id); if (idx >= 0) selected.splice(idx, 1); else selected.push(item.id) }">
        <template #content>
          <div class="text-center relative">
            <div class="absolute top-0 left-0"><Checkbox :modelValue="selected.includes(item.id)" binary /></div>
            <img :src="`/api/barcode/${item.id}.png?fmt=qr`" class="w-20 h-20 mx-auto rounded-lg border border-gray-200" />
            <div class="text-xs font-medium mt-1 truncate">{{ item.name }}</div>
            <code class="text-xs text-gray-400">{{ item.sku }}</code>
          </div>
        </template>
      </Card>
    </div>
  </div>
</template>
