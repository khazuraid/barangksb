<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import api from '@/api'
import Card from 'primevue/card'
import Tag from 'primevue/tag'
const route = useRoute()
const item = ref(null)
onMounted(async () => { try { item.value = (await api.get(`/items/${route.params.id}`)).data } catch {} })
</script>
<template>
  <div class="min-h-screen flex items-center justify-center p-4 bg-gradient-to-br from-emerald-50 to-teal-50">
    <Card class="shadow-xl max-w-md w-full"><template #content>
      <div class="text-center">
        <Tag severity="success" value="QR Terdeteksi" class="mb-4" />
        <img :src="`/api/barcode/${route.params.id}.png?fmt=qr`" class="w-32 h-32 mx-auto rounded-xl border-2 border-gray-200" />
        <h2 class="text-xl font-bold mt-4">{{ item?.name }}</h2>
        <code class="text-sm text-gray-400">{{ item?.sku }}</code>
        <div class="text-left mt-4 space-y-2 text-sm">
          <div class="flex justify-between"><span class="text-gray-500">Kategori:</span><span>{{ item?.category }}</span></div>
          <div class="flex justify-between"><span class="text-gray-500">Lokasi:</span><span>{{ item?.location }}</span></div>
          <div class="flex justify-between"><span class="text-gray-500">Stok:</span><span>{{ item?.current_stock }} {{ item?.unit }}</span></div>
        </div>
        <a href="/login" class="btn btn-ghost btn-sm mt-4 inline-block text-emerald-500">Petugas? Masuk</a>
      </div>
    </template></Card>
  </div>
</template>
