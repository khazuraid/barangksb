<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
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
const showMap = ref(false)
const photoModal = ref(false)
const selectedPhoto = ref('')
const activeTab = ref<'detail' | 'maintenance'>('detail')
const isLoggedIn = ref(false)

onMounted(async () => {
  isLoggedIn.value = !!localStorage.getItem('token')
  try {
    item.value = (await api.get(`/scan/${route.params.id}`)).data
  } catch {
    error.value = 'Barang tidak ditemukan atau kode QR tidak valid.'
  } finally { loading.value = false }
})

const hasCoords = computed(() => {
  return item.value && item.value.geo_lat != null && item.value.geo_lng != null
})

const osmEmbedUrl = computed(() => {
  if (!hasCoords.value) return ''
  const lat = item.value.geo_lat
  const lng = item.value.geo_lng
  const delta = 0.0035
  const minLng = lng - delta
  const minLat = lat - delta
  const maxLng = lng + delta
  const maxLat = lat + delta
  const bbox = `${minLng}%2C${minLat}%2C${maxLng}%2C${maxLat}`
  return `https://www.openstreetmap.org/export/embed.html?bbox=${bbox}&layer=mapnik&marker=${lat}%2C${lng}`
})

const googleMapsUrl = computed(() => {
  if (!hasCoords.value) return ''
  return `https://www.google.com/maps?q=${item.value.geo_lat},${item.value.geo_lng}`
})

const maintenances = computed(() => item.value?.maintenances || [])

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
    <div class="w-full max-w-[460px]">
      <!-- header -->
      <div class="flex items-center gap-3 mb-5">
        <div class="w-9 h-9 grid place-items-center rounded-md bg-acc-500 text-ink-950 font-extrabold text-[13px]">IK</div>
        <div class="leading-tight">
          <div class="text-[12.5px] font-bold tracking-wide">INVENTARIS KANTOR</div>
          <div class="t-label !text-[9.5px]">Hasil Pemindaian QR &amp; Riwayat Aset</div>
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
        <div class="panel shell-panel-2 overflow-hidden shadow-xl">
          <!-- Photo with Geotag if available -->
          <div v-if="item.photo_url" class="relative bg-black/40 border-b overflow-hidden cursor-pointer group"
               style="border-color: var(--line)"
               @click="selectedPhoto = item.photo_url; photoModal = true">
            <img :src="item.photo_url" alt="Foto Barang Berstempel Geotag" class="w-full max-h-64 object-contain group-hover:scale-[1.01] transition-transform" />
            <div class="absolute bottom-2 left-2 bg-black/80 px-2.5 py-1 rounded text-[10.5px] text-white flex items-center gap-1.5">
              <i class="pi pi-map-marker text-sig-ok text-[10px]" /> Cap GPS Map Camera
            </div>
            <div class="absolute top-2 right-2 bg-black/75 px-2 py-1 rounded text-[10px] text-white flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
              <i class="pi pi-search-plus text-[10px]" /> Perbesar
            </div>
          </div>

          <!-- Interactive Map Strip if coordinates exist -->
          <div v-if="hasCoords" class="p-3 border-b text-[11.5px] flex flex-col gap-2"
               style="border-color: var(--line); background: var(--paper-1)">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-1.5 t-mono">
                <i class="pi pi-map-marker text-sig-ok" />
                <span class="font-semibold text-sig-ok">Lokasi GPS:</span>
                <span>{{ item.geo_lat.toFixed(5) }}, {{ item.geo_lng.toFixed(5) }}</span>
              </div>
              <div class="flex items-center gap-1">
                <Button
                  :label="showMap ? 'Tutup Peta' : 'Peta'"
                  :icon="showMap ? 'pi pi-times' : 'pi pi-map'"
                  size="small"
                  text
                  :severity="showMap ? 'warn' : 'secondary'"
                  class="!text-[11px] !py-0.5"
                  @click="showMap = !showMap"
                />
                <a :href="googleMapsUrl" target="_blank" rel="noopener noreferrer" class="inline-flex">
                  <Button icon="pi pi-external-link" size="small" text v-tooltip.top="'Buka di Google Maps'" class="!p-1" />
                </a>
              </div>
            </div>

            <!-- OpenStreetMap Embed Frame -->
            <div v-if="showMap" class="w-full h-44 rounded border overflow-hidden mt-1" style="border-color: var(--line)">
              <iframe
                :src="osmEmbedUrl"
                class="w-full h-full border-0"
                loading="lazy"
                title="Peta Lokasi OpenStreetMap"
              />
            </div>
          </div>

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

          <!-- Tabs Switcher: Detail vs Riwayat Servis -->
          <div class="flex border-b text-[12px] font-bold" style="border-color: var(--line); background: var(--paper-2)">
            <button
              class="flex-1 py-2.5 px-3 flex items-center justify-center gap-1.5 transition-colors cursor-pointer border-b-2"
              :class="activeTab === 'detail' ? 'border-acc-500 text-acc-500 bg-paper-1' : 'border-transparent text-ink-400 hover:text-ink-200'"
              @click="activeTab = 'detail'"
            >
              <i class="pi pi-info-circle text-[11px]" /> Detail Spesifikasi
            </button>
            <button
              class="flex-1 py-2.5 px-3 flex items-center justify-center gap-1.5 transition-colors cursor-pointer border-b-2"
              :class="activeTab === 'maintenance' ? 'border-acc-500 text-acc-500 bg-paper-1' : 'border-transparent text-ink-400 hover:text-ink-200'"
              @click="activeTab = 'maintenance'"
            >
              <i class="pi pi-wrench text-[11px]" />
              <span>Riwayat Servis</span>
              <span v-if="maintenances.length" class="text-[10px] px-1.5 py-0.2 rounded-full bg-acc-500 text-ink-950">
                {{ maintenances.length }}
              </span>
            </button>
          </div>

          <!-- Tab 1: Detail Rows -->
          <dl v-if="activeTab === 'detail'" class="divide-y" style="border-color: var(--line-soft)">
            <div v-for="([k, v]) in rows()" :key="String(k)"
                 class="flex items-start justify-between gap-4 px-5 py-2.5">
              <dt class="t-label shrink-0 pt-0.5">{{ k }}</dt>
              <dd class="text-[12.5px] font-medium text-right break-words">{{ v }}</dd>
            </div>
          </dl>

          <!-- Tab 2: Riwayat Servis & Kalibrasi (View Publik) -->
          <div v-else class="p-4 flex flex-col gap-3">
            <div v-if="!maintenances.length" class="text-center py-8 text-ink-400 text-[12px] flex flex-col items-center gap-2">
              <i class="pi pi-wrench text-2xl text-ink-500" />
              <span>Belum ada catatan servis atau kalibrasi untuk barang ini.</span>
            </div>

            <div v-else class="flex flex-col gap-2.5">
              <div
                v-for="(m, idx) in maintenances"
                :key="idx"
                class="rounded-lg p-3 border text-[12px] flex flex-col gap-1.5"
                style="background: var(--paper-1); border-color: var(--line)"
              >
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-1.5">
                    <span
                      class="px-2 py-0.5 rounded text-[10.5px] font-bold"
                      :class="m.service_type === 'Kalibrasi' ? 'bg-purple-900/60 text-purple-300 border border-purple-700/50' : 'bg-acc-500/20 text-acc-400 border border-acc-500/40'"
                    >
                      {{ m.service_type }}
                    </span>
                    <span class="font-semibold">{{ m.service_date }}</span>
                  </div>
                  <span v-if="m.vendor_or_technician" class="text-[11px] text-ink-400">
                    {{ m.vendor_or_technician }}
                  </span>
                </div>

                <p v-if="m.description" class="text-[11.5px] text-ink-300 leading-snug">
                  {{ m.description }}
                </p>

                <div v-if="m.next_service_date" class="flex items-center gap-1.5 text-[11px] font-semibold text-sig-ok pt-0.5">
                  <i class="pi pi-calendar text-[10px]" />
                  <span>Jatuh Tempo Berikutnya: {{ m.next_service_date }}</span>
                </div>

                <div v-if="m.photo_url" class="mt-1">
                  <button
                    class="text-[10.5px] text-acc-500 hover:underline flex items-center gap-1 cursor-pointer"
                    @click="selectedPhoto = m.photo_url; photoModal = true"
                  >
                    <i class="pi pi-image text-[10px]" /> Lihat Foto Bukti Nota
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="flex gap-2 mt-4">
          <Button label="Buka Panel" icon="pi pi-sign-in" size="small" class="flex-1"
                  @click="router.push('/')" />
          <Button label="Daftar Barang" icon="pi pi-box" size="small" severity="secondary" outlined
                  class="flex-1" @click="router.push('/items')" />
        </div>

        <p class="text-[11px] mt-4 text-ink-600 leading-relaxed text-center">
          Pembaruan data dan pencatatan servis resmi memerlukan login petugas di sistem.
        </p>
      </template>

      <!-- Lightbox Modal Foto Geotag & Nota -->
      <div
        v-if="photoModal && selectedPhoto"
        class="fixed inset-0 z-50 bg-black/85 flex items-center justify-center p-4 backdrop-blur-xs"
        @click.self="photoModal = false"
      >
        <div class="relative max-w-xl w-full bg-[var(--paper-1)] rounded-lg overflow-hidden border shadow-2xl flex flex-col"
             style="border-color: var(--line)">
          <div class="flex items-center justify-between px-3 py-2 border-b" style="border-color: var(--line)">
            <span class="text-[12px] font-bold flex items-center gap-1.5">
              <i class="pi pi-image text-acc-500" /> Foto Berstempel Geotag
            </span>
            <Button icon="pi pi-times" text rounded size="small" severity="secondary" @click="photoModal = false" />
          </div>
          <div class="bg-black p-2 flex items-center justify-center max-h-[70vh] overflow-hidden">
            <img :src="selectedPhoto" alt="Foto Geotag Asli" class="max-h-[65vh] object-contain rounded" />
          </div>
          <div class="flex items-center justify-between px-3 py-2 border-t text-[11px]" style="border-color: var(--line)">
            <span v-if="hasCoords" class="t-mono text-sig-ok">
              GPS: {{ item.geo_lat.toFixed(5) }}, {{ item.geo_lng.toFixed(5) }}
            </span>
            <div class="flex gap-1.5 ml-auto">
              <a v-if="hasCoords" :href="googleMapsUrl" target="_blank" rel="noopener noreferrer">
                <Button label="Google Maps" icon="pi pi-map-marker" size="small" text />
              </a>
              <a :href="selectedPhoto" target="_blank" download>
                <Button label="Unduh" icon="pi pi-download" size="small" text severity="secondary" />
              </a>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
