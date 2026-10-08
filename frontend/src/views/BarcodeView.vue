<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api'
const items = ref([]); const selected = ref<string[]>([])
onMounted(async () => { const res = await api.get('/items', { params: { per_page: 100 } }); items.value = res.data.data })
function printSheet() { if (selected.value.length) window.open(`/api/barcode/sheet?ids=${selected.value.join(',')}`, '_blank') }
</script>
<template>
  <div class="space-y-4">
    <div class="flex justify-between"><h2 class="text-xl font-bold">QR Generator</h2><button class="btn btn-primary btn-sm" @click="printSheet">Cetak Terpilih</button></div>
    <div class="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-3">
      <label v-for="item in items" :key="item.id" class="card bg-base-100 border border-base-200 shadow-sm p-3 text-center cursor-pointer">
        <input type="checkbox" :value="item.id" v-model="selected" class="checkbox checkbox-sm checkbox-primary mb-2" />
        <img :src="`/api/barcode/${item.id}.png?fmt=qr`" class="w-20 h-20 mx-auto rounded-lg" />
        <div class="text-xs font-medium mt-1 truncate">{{ item.name }}</div>
        <code class="text-xs">{{ item.sku }}</code>
      </label>
    </div>
  </div>
</template>
