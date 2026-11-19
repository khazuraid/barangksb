<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api, { safeHref } from '@/api'
import { useThemeStore } from '@/stores/theme'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import StatusChip from '@/components/StatusChip.vue'
import AppLogo from '@/components/AppLogo.vue'

const route = useRoute()
const router = useRouter()
const theme = useThemeStore()
const item = ref<any>(null)
const error = ref('')
const loading = ref(true)
const showMap = ref(false)
const photoModal = ref(false)
const selectedPhoto = ref('')
const activeTab = ref<'detail' | 'maintenance'>('detail')
const isLoggedIn = ref(false)

onMounted(async () => {
  theme.apply()
  try {
    isLoggedIn.value = !!localStorage.getItem('token')
  } catch {
    isLoggedIn.value = false
  }
  try {
    item.value = (await api.get(`/scan/${route.params.id}`)).data
  } catch {
    error.value = 'Barang tidak ditemukan atau kode QR tidak valid.'
  } finally {
    loading.value = false
  }
})

const hasCoords = computed(() => item.value && item.value.geo_lat != null && item.value.geo_lng != null)

const osmEmbedUrl = computed(() => {
  if (!hasCoords.value) return ''
  const lat = item.value.geo_lat
  const lng = item.value.geo_lng
  const delta = 0.0035
  const bbox = `${lng - delta}%2C${lat - delta}%2C${lng + delta}%2C${lat + delta}`
  return `https://www.openstreetmap.org/export/embed.html?bbox=${bbox}&layer=mapnik&marker=${lat}%2C${lng}`
})

const googleMapsUrl = computed(() => {
  if (!hasCoords.value) return ''
  return `https://www.google.com/maps?q=${item.value.geo_lat},${item.value.geo_lng}`
})

async function downloadImage(url: string, filename = 'foto_geotag.jpg') {
  if (!url) return
  try {
    const res = await fetch(url)
    const blob = await res.blob()
    const blobUrl = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = blobUrl
    a.download = filename
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(blobUrl)
  } catch {
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    a.target = '_blank'
    a.rel = 'noopener noreferrer'
    a.click()
  }
}

const maintenances = computed(() => item.value?.maintenances || [])

function formatRupiah(val: number) {
  if (!val) return '—'
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(val)
}

function rows() {
  const it = item.value
  if (!it) return []
  const r: [string, any][] = [
    ['Kategori', it.category],
    ['Ruangan / Lokasi', it.location],
  ]
  if (it.location_code) r.push(['Kode Lokasi', it.location_code])
  if (it.track_stock !== false) {
    r.push(['Stok Tersedia', `${it.current_stock} ${it.unit}`])
    r.push(['Batas Minimum', `${it.min_stock} ${it.unit}`])
  } else {
    r.push(['Model Pengelolaan', 'Aset Tetap / Unit Mandiri (BMD)'])
    if (it.item_code) r.push(['No. Kode Barang', it.item_code])
    if (it.register_number) r.push(['No. Register', it.register_number])
  }
  r.push(
    ['Keadaan Barang', it.condition_status],
    ['Merk', it.merk],
    ['Model / Tipe', it.type_model],
    ['Serial Number (SN)', it.serial_number],
    ['Ukuran / Dimensi', it.size],
    ['Bahan / Material', it.material],
    ['Tahun Pengadaan', it.procurement_year],
    ['Sumber Dana', it.funding_source],
    ['Distributor / Vendor', it.distributor],
    ['Izin Edar (AKL/AKD)', it.akl_akd],
    ['Estimasi Nilai Satuan', it.price_per_unit ? formatRupiah(it.price_per_unit) : ''],
    ['Keterangan', it.notes]
  )
  return r.filter(([, v]) => v !== '' && v !== null && v !== undefined)
}
</script>

<template>
  <div class="min-h-dvh shell grid place-items-center p-4 sm:p-6">
    <div class="w-full max-w-[480px]">
      <!-- Header -->
      <header class="flex items-center justify-between mb-4">
        <AppLogo size="md" />
        <div class="flex items-center gap-1.5">
          <Button
            :icon="theme.isDark ? 'pi pi-sun' : 'pi pi-moon'"
            text
            rounded
            size="small"
            severity="secondary"
            :aria-label="theme.isDark ? 'Ganti ke mode terang' : 'Ganti ke mode gelap'"
            @click="theme.toggle()"
          />
          <span class="t-mono !text-[10px] px-2 py-1 rounded-[6px] border border-line bg-panel text-txt-dim">
            QR Scan
          </span>
        </div>
      </header>

      <!-- Loading Skeleton -->
      <div v-if="loading" class="panel p-5 space-y-4">
        <div class="flex gap-3">
          <div class="skeleton w-14 h-14 rounded-[8px] shrink-0" />
          <div class="flex-1 space-y-2">
            <div class="skeleton h-3.5 w-24" />
            <div class="skeleton h-5 w-48" />
            <div class="skeleton h-3 w-32" />
          </div>
        </div>
        <div v-for="n in 5" :key="n" class="skeleton h-8 w-full" />
      </div>

      <!-- Error State -->
      <div v-else-if="error" class="panel p-7 text-center flex flex-col items-center gap-3">
        <div class="w-12 h-12 rounded-full bg-[var(--sig-bad-bg)] border border-[var(--sig-bad-line)] flex items-center justify-center text-sig-bad text-xl">
          <i class="pi pi-qrcode" aria-hidden="true" />
        </div>
        <div class="text-[16px] font-bold text-sig-bad">Kode QR Tidak Terdaftar</div>
        <p class="text-[12.5px] text-txt-dim leading-relaxed max-w-xs">
          Barang tidak ditemukan di sistem inventaris. Kemungkinan kode QR salah, barang telah dihapus, atau belum diinput ke master data.
        </p>
        <div class="px-3 py-1.5 rounded-[6px] t-mono !text-[11px] text-txt-dim bg-panel-2 border border-line">
          ID: {{ route.params.id }}
        </div>
        <div class="flex gap-2 w-full mt-2 pt-4 border-t border-line">
          <Button label="Dashboard" icon="pi pi-th-large" size="small" class="flex-1" @click="router.push('/')" />
          <Button label="Daftar Barang" icon="pi pi-box" size="small" severity="secondary" outlined class="flex-1" @click="router.push('/items')" />
        </div>
      </div>

      <!-- Asset Card -->
      <div v-else-if="item" class="panel overflow-hidden">
        <!-- Hero Band -->
        <div class="p-4 border-b border-line bg-panel-2">
          <div class="flex items-start gap-3.5">
            <button
              v-if="item.photo_url"
              type="button"
              class="w-16 h-16 rounded-[9px] border border-line overflow-hidden shrink-0 bg-panel cursor-pointer"
              aria-label="Perbesar foto barang"
              @click="selectedPhoto = item.photo_url; photoModal = true"
            >
              <img :src="item.photo_url" :alt="item.name" class="w-full h-full object-cover" />
            </button>
            <div v-else class="w-16 h-16 rounded-[9px] border border-line bg-panel grid place-items-center shrink-0">
              <i class="pi pi-box text-acc-500 text-xl" aria-hidden="true" />
            </div>

            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-1.5 flex-wrap mb-1">
                <span class="t-mono !text-[10.5px] font-semibold px-1.5 py-0.5 rounded border border-line bg-panel text-txt-dim">
                  {{ item.sku }}
                </span>
                <StatusChip v-if="item.condition_status" :kind="item.condition_status" />
              </div>
              <h1 class="text-[17px] font-bold tracking-[-0.02em] text-txt leading-snug">
                {{ item.name }}
              </h1>
              <div class="text-[12px] text-txt-dim mt-1 flex items-center gap-1.5">
                <i class="pi pi-map-marker text-[10.5px] text-acc-500" aria-hidden="true" />
                <span>{{ item.location || 'Lokasi belum ditentukan' }}</span>
              </div>
            </div>
          </div>

          <!-- Stock / Asset Strip -->
          <div class="mt-3.5 pt-3 border-t border-line flex items-center justify-between gap-2">
            <div v-if="item.track_stock !== false" class="flex items-baseline gap-1.5">
              <span class="text-[11px] uppercase tracking-[0.05em] text-txt-dim font-semibold">Stok Saat Ini</span>
              <span class="t-num text-[20px] font-bold" :class="item.current_stock <= item.min_stock ? 'text-sig-bad' : 'text-sig-ok'">
                {{ item.current_stock }}
              </span>
              <span class="text-[12px] text-txt-dim">{{ item.unit }}</span>
            </div>
            <div v-else class="flex items-center gap-2">
              <span class="text-[11px] font-semibold text-sig-info bg-[var(--sig-info-bg)] border border-[var(--sig-info-line)] px-2 py-0.5 rounded-[5px]">
                Aset Tetap (BMD)
              </span>
              <span v-if="item.item_code" class="t-mono !text-[11px] text-txt-dim">{{ item.item_code }}</span>
            </div>

            <Button
              label="Laporkan Masalah"
              icon="pi pi-megaphone"
              size="small"
              severity="warn"
              outlined
              @click="router.push(`/lapor/${item.id}`)"
            />
          </div>
        </div>

        <!-- Tabs -->
        <div class="grid grid-cols-2 border-b border-line bg-panel">
          <button
            type="button"
            class="py-2.5 text-[12.5px] font-semibold border-b-2 transition-colors cursor-pointer"
            :class="activeTab === 'detail' ? 'border-acc-500 text-acc-500' : 'border-transparent text-txt-dim hover:text-txt'"
            @click="activeTab = 'detail'"
          >
            Spesifikasi Aset
          </button>
          <button
            type="button"
            class="py-2.5 text-[12.5px] font-semibold border-b-2 transition-colors cursor-pointer flex items-center justify-center gap-1.5"
            :class="activeTab === 'maintenance' ? 'border-acc-500 text-acc-500' : 'border-transparent text-txt-dim hover:text-txt'"
            @click="activeTab = 'maintenance'"
          >
            <span>Riwayat Servis</span>
            <span class="t-mono !text-[10.5px] px-1.5 py-0.5 rounded bg-panel-2 border border-line text-txt-dim">
              {{ maintenances.length }}
            </span>
          </button>
        </div>

        <!-- Tab 1: Specification -->
        <div v-if="activeTab === 'detail'" class="p-4 space-y-4">
          <dl class="divide-y divide-line text-[12.5px]">
            <div v-for="[k, v] in rows()" :key="k" class="py-2 flex items-start justify-between gap-4">
              <dt class="text-txt-dim shrink-0">{{ k }}</dt>
              <dd class="font-medium text-txt text-right break-words">{{ v }}</dd>
            </div>
          </dl>

          <!-- Geotag Location Block -->
          <div v-if="hasCoords" class="p-3 rounded-[8px] border border-line bg-panel-2 space-y-2.5">
            <div class="flex items-center justify-between gap-2">
              <div class="flex items-center gap-1.5 text-[12px] font-semibold text-txt">
                <i class="pi pi-compass text-acc-500 text-[12px]" aria-hidden="true" />
                <span>Koordinat GPS Tercatat</span>
              </div>
              <span class="t-mono !text-[10.5px] text-txt-dim">
                {{ Number(item.geo_lat).toFixed(5) }}, {{ Number(item.geo_lng).toFixed(5) }}
              </span>
            </div>

            <div class="flex flex-wrap gap-2">
              <Button
                :label="showMap ? 'Sembunyikan Peta' : 'Lihat Peta Lokasi'"
                icon="pi pi-map"
                size="small"
                severity="secondary"
                outlined
                class="flex-1"
                @click="showMap = !showMap"
              />
              <a
                :href="googleMapsUrl"
                target="_blank"
                rel="noopener noreferrer"
                class="inline-flex items-center justify-center gap-1.5 px-3 py-1.5 rounded-[7px] border border-line bg-panel hover:border-[var(--acc-border)] text-[12px] font-semibold text-txt transition-colors"
              >
                <i class="pi pi-external-link text-[11px]" aria-hidden="true" />
                <span>Google Maps</span>
              </a>
            </div>

            <iframe
              v-if="showMap"
              :src="osmEmbedUrl"
              title="Peta lokasi aset"
              class="w-full h-48 rounded-[7px] border border-line"
              loading="lazy"
            />
          </div>

          <div class="flex items-center justify-between gap-2 pt-2 border-t border-line">
            <Button
              v-if="isLoggedIn"
              label="Buka di Panel Admin"
              icon="pi pi-arrow-right"
              iconPos="right"
              size="small"
              class="flex-1"
              @click="router.push(`/items/${item.id}`)"
            />
            <Button
              v-else
              label="Masuk Petugas"
              icon="pi pi-sign-in"
              size="small"
              severity="secondary"
              outlined
              class="flex-1"
              @click="router.push(`/login?redirect=/items/${item.id}`)"
            />
          </div>
        </div>

        <!-- Tab 2: Maintenance History -->
        <div v-else class="p-4">
          <div v-if="!maintenances.length" class="py-10 text-center">
            <i class="pi pi-wrench text-2xl text-txt-muted mb-2 block" aria-hidden="true" />
            <div class="text-[13px] font-semibold text-txt">Belum ada catatan pemeliharaan</div>
            <div class="text-[12px] text-txt-dim mt-1">Riwayat servis dan kalibrasi alat ini akan tampil di sini.</div>
          </div>

          <div v-else class="space-y-2.5">
            <div
              v-for="m in maintenances"
              :key="m.id"
              class="p-3 rounded-[8px] border border-line bg-panel-2 space-y-1.5 text-[12px]"
            >
              <div class="flex items-center justify-between gap-2">
                <span class="font-semibold text-acc-500">{{ m.service_type }}</span>
                <span class="t-mono !text-[11px] text-txt-dim">{{ m.service_date }}</span>
              </div>
              <div v-if="m.vendor_or_technician" class="text-txt-dim text-[11.5px]">
                Pelaksana: <strong class="text-txt">{{ m.vendor_or_technician }}</strong>
              </div>
              <p v-if="m.description" class="text-txt leading-relaxed">{{ m.description }}</p>
              <div class="flex items-center justify-between pt-1.5 border-t border-line-soft text-[11px]">
                <span v-if="m.next_service_date" class="text-sig-ok font-semibold">
                  Berikutnya: {{ m.next_service_date }}
                </span>
                <span v-else />
                <a
                  v-if="safeHref(m.photo_url)"
                  :href="safeHref(m.photo_url)"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="text-acc-500 hover:underline font-medium"
                >
                  Lihat Bukti
                </a>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Photo Lightbox -->
      <Dialog v-model:visible="photoModal" modal header="Foto Aset" :style="{ width: '92vw', maxWidth: '640px' }">
        <div class="flex flex-col items-center gap-3">
          <img :src="selectedPhoto" alt="Foto aset diperbesar" class="max-h-[70vh] w-auto rounded-[8px] border border-line object-contain" />
          <Button
            label="Unduh Foto"
            icon="pi pi-download"
            size="small"
            severity="secondary"
            outlined
            @click="downloadImage(selectedPhoto, `${item?.sku || 'aset'}.jpg`)"
          />
        </div>
      </Dialog>
    </div>
  </div>
</template>
