<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '@/api'
import Button from 'primevue/button'
import StatusChip from '@/components/StatusChip.vue'
import Tag from 'primevue/tag'

const route = useRoute()
const router = useRouter()
const item = ref<any>(null)
const error = ref('')
const loading = ref(true)

onMounted(async () => {
  try {
    item.value = (await api.get(`/scan/${route.params.id}`)).data
  } catch {
    error.value = 'Barang tidak ditemukan atau kode QR tidak valid.'
  } finally { loading.value = false }
})

function rows() {
  return [
    ['Kategori', item.value?.category],
    ['Lokasi', item.value?.location],
    ['Stok Tersedia', `${item.value?.current_stock} ${item.value?.unit}`],
    ['Batas Minimum', `${item.value?.min_stock} ${item.value?.unit}`],
    ['Merk', item.value?.merk],
    ['Model / Tipe', item.value?.type_model],
    ['Serial Number', item.value?.serial_number],
    ['Tahun Pengadaan', item.value?.procurement_year],
    ['Sumber Dana', item.value?.funding_source],
    ['Akl / Akd', item.value?.akl_akd],
  ].filter(([, v]) => v !== undefined && v !== null && String(v).trim() !== '')
}
</script>

<template>
  <div class="min-h-screen shell grid place-items-center p-5">
    <div class="w-full max-w-[420px]">
      <!-- header -->
      <div class="flex items-center gap-3 mb-5">
        <div class="w-9 h-9 grid place-items-center rounded-md bg-acc-500 text-ink-950 font-extrabold text-[13px]">IK</div>
        <div class="leading-tight">
          <div class="text-[12.5px] font-bold tracking-wide">INVENTARIS KANTOR</div>
          <div class="t-label !text-[9.5px]">Hasil Pemindaian QR</div>
        </div>
      </div>

      <div v-if="loading" class="panel shell-panel-2 p-10 grid place-items-center">
        <i class="pi pi-spin pi-spinner text-xl text-acc-500" />
      </div>

      <div v-else-if="error" class="panel shell-panel-2 p-6 text-center">
        <i class="pi pi-times-circle text-3xl text-sig-bad" />
        <div class="text-[14px] font-bold mt-3">Tidak ditemukan</div>
        <p class="text-[12.5px] mt-1.5 text-ink-400">{{ error }}</p>
        <Button label="Ke panel utama" icon="pi pi-arrow-right" iconPos="right" size="small"
                class="mt-5" @click="router.push('/')" />
      </div>

      <template v-else>
        <div class="panel shell-panel-2 overflow-hidden">
          <!-- identity strip -->
          <div class="relative p-5 border-b" style="border-color: var(--line)">
            <span class="absolute left-0 top-0 bottom-0 w-[3px] bg-acc-500" />
            <div class="flex items-start justify-between gap-3">
              <div class="min-w-0">
                <div class="t-label mb-1">Barang Teridentifikasi</div>
                <h1 class="text-[20px] font-bold leading-tight tracking-tight">{{ item.name }}</h1>
                <div class="t-mono mt-1.5 text-ink-400">{{ item.sku }}</div>
              </div>
              <img :src="`/api/barcode/${route.params.id}.png?fmt=qr&size=120`"
                   class="w-[84px] h-[84px] shrink-0 rounded-md border bg-white p-1"
                   style="border-color: var(--line)" />
            </div>
            <div class="flex flex-wrap gap-1.5 mt-3.5">
              <StatusChip :kind="item.condition_status" />
              <Tag v-if="item.is_available" severity="success" value="TERSEDIA" icon="pi pi-check" />
              <Tag v-else severity="danger" value="TIDAK TERSEDIA" />
            </div>
          </div>

          <!-- stock hero -->
          <div class="grid grid-cols-2 gap-px" style="background: var(--line-soft)">
            <div class="p-4" style="background: var(--panel-2)">
              <div class="t-label">Stok Saat Ini</div>
              <div class="t-num text-[28px] font-extrabold mt-1 leading-none"
                   :class="item.current_stock <= item.min_stock ? 'text-sig-bad' : 'text-acc-500'">
                {{ item.current_stock }}
              </div>
              <div class="text-[11px] text-ink-400 mt-1">{{ item.unit }}</div>
            </div>
            <div class="p-4" style="background: var(--panel-2)">
              <div class="t-label">Batas Minimum</div>
              <div class="t-num text-[28px] font-extrabold mt-1 leading-none text-ink-300">{{ item.min_stock }}</div>
              <div class="text-[11px] text-ink-400 mt-1">{{ item.min_stock > item.current_stock ? 'Stok menipis' : 'Stok aman' }}</div>
            </div>
          </div>

          <!-- detail rows -->
          <dl class="divide-y" style="border-color: var(--line-soft)">
            <div v-for="([k, v]) in rows()" :key="String(k)"
                 class="flex items-start justify-between gap-4 px-5 py-2.5">
              <dt class="t-label shrink-0 pt-0.5">{{ k }}</dt>
              <dd class="text-[12.5px] font-medium text-right break-words">{{ v }}</dd>
            </div>
          </dl>
        </div>

        <div class="flex gap-2 mt-4">
          <Button label="Buka Panel" icon="pi pi-sign-in" size="small" class="flex-1"
                  @click="router.push('/')" />
          <Button label="Riwayat" icon="pi pi-history" size="small" severity="secondary" outlined
                  class="flex-1" @click="router.push('/history')" />
        </div>

        <p class="text-[11px] mt-4 text-ink-600 leading-relaxed text-center">
          Halaman ini hanya menampilkan data. Perubahan stok memerlukan login petugas.
        </p>
      </template>
    </div>
  </div>
</template>
