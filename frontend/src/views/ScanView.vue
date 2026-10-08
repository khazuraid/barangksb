<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import api from '@/api'
const route = useRoute()
const item = ref(null)
onMounted(async () => { try { item.value = (await api.get(`/items/${route.params.id}`)).data } catch {} })
</script>
<template>
  <div class="min-h-screen flex items-center justify-center p-4">
    <div class="card bg-base-100 shadow-xl max-w-md w-full">
      <div class="card-body text-center">
        <div class="badge badge-success badge-lg mx-auto mb-3">QR Terdeteksi</div>
        <img :src="`/api/barcode/${route.params.id}.png?fmt=qr`" class="w-32 h-32 mx-auto rounded-xl" />
        <h2 class="text-xl font-bold">{{ item?.name }}</h2>
        <code class="text-sm">{{ item?.sku }}</code>
        <div class="text-left mt-4 space-y-1 text-sm">
          <div><span class="text-base-content/50">Kategori:</span> {{ item?.category }}</div>
          <div><span class="text-base-content/50">Lokasi:</span> {{ item?.location }}</div>
          <div><span class="text-base-content/50">Stok:</span> {{ item?.current_stock }} {{ item?.unit }}</div>
        </div>
        <a href="/login" class="btn btn-ghost btn-sm mt-4">Petugas? Masuk</a>
      </div>
    </div>
  </div>
</template>
