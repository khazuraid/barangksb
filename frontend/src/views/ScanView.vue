<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api, { safeHref } from '@/api'
import { useThemeStore } from '@/stores/theme'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import StatusChip from '@/components/StatusChip.vue'
import AppLogo from '@/components/AppLogo.vue'
import ScanUnitCard from '@/components/ScanUnitCard.vue'

const route = useRoute()
const router = useRouter()
const theme = useThemeStore()
const item = ref<any>(null)
const error = ref('')
const loading = ref(true)
const showMap = ref(false)
const photoModal = ref(false)
const selectedPhoto = ref('')
const activeTab = ref<'detail' | 'units' | 'maintenance'>('detail')
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
const units = computed<any[]>(() => item.value?.units || [])
const scannedUnitId = computed<string>(() => item.value?.scanned_unit?.id || '')

const isStockItem = computed(() => item.value?.track_stock !== false)

// Total unit barang: barang habis pakai = stok saat ini, aset tetap (BMD) = jumlah unit fisik terdaftar.
const totalUnits = computed<number>(() =>
  isStockItem.value ? Number(item.value?.current_stock ?? 0) : units.value.length
)
const totalUnitsCaption = computed(() =>
  isStockItem.value ? item.value?.unit || 'unit' : 'unit fisik terdaftar'
)
const isLowStock = computed(
  () => isStockItem.value && Number(item.value?.current_stock ?? 0) <= Number(item.value?.min_stock ?? 0)
)
const scannedUnit = computed<any>(() => item.value?.scanned_unit || null)

const tabs = computed(() => {
  const list: { id: 'detail' | 'units' | 'maintenance'; label: string; count?: number }[] = [
    { id: 'detail', label: 'Spesifikasi' },
  ]
  if (units.value.length) list.push({ id: 'units', label: 'Unit Fisik', count: units.value.length })
  list.push({ id: 'maintenance', label: 'Servis', count: maintenances.value.length })
  return list
})

function openPhoto(url: string) {
  if (!url) return
  selectedPhoto.value = url
  photoModal.value = true
}

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
  <div class="min-h-dvh shell flex flex-col">
    <!-- ============ Sticky Glass Header ============ -->
    <header class="sticky top-0 z-20 border-b border-line bg-[var(--topbar-bg)] backdrop-blur-md">
      <div class="max-w-3xl mx-auto px-4 sm:px-6 h-14 flex items-center justify-between gap-3">
        <AppLogo size="sm" />
        <div class="flex items-center gap-1.5 shrink-0">
          <span class="hidden sm:inline-flex items-center gap-1.5 t-mono !text-[11px] uppercase tracking-[0.06em] px-2.5 py-1 rounded-full border border-[var(--acc-border)] bg-acc-50 text-acc-600 dark:text-acc-300">
            <i class="pi pi-qrcode text-[11px]" aria-hidden="true" />QR Scan
          </span>
          <Button
            :icon="theme.isDark ? 'pi pi-sun' : 'pi pi-moon'"
            text
            rounded
            severity="secondary"
            :aria-label="theme.isDark ? 'Ganti ke mode terang' : 'Ganti ke mode gelap'"
            @click="theme.toggle()"
          />
        </div>
      </div>
    </header>

    <main class="flex-1 w-full max-w-3xl mx-auto px-4 sm:px-6 py-5 sm:py-8 space-y-4">
      <!-- ============ Loading ============ -->
      <div v-if="loading" class="panel p-5 space-y-4" aria-busy="true" aria-live="polite">
        <div class="flex gap-3">
          <div class="skeleton w-20 h-20 rounded-[10px] shrink-0" />
          <div class="flex-1 space-y-2.5">
            <div class="skeleton h-3.5 w-24" />
            <div class="skeleton h-6 w-3/4" />
            <div class="skeleton h-3.5 w-1/2" />
          </div>
        </div>
        <div class="grid grid-cols-3 gap-2.5">
          <div v-for="n in 3" :key="n" class="skeleton h-[74px]" />
        </div>
        <div v-for="n in 4" :key="'r' + n" class="skeleton h-8 w-full" />
      </div>

      <!-- ============ Error ============ -->
      <div v-else-if="error" class="panel p-7 text-center flex flex-col items-center gap-3" role="alert">
        <div class="w-12 h-12 rounded-full bg-[var(--sig-bad-bg)] border border-[var(--sig-bad-line)] flex items-center justify-center text-sig-bad text-xl">
          <i class="pi pi-qrcode" aria-hidden="true" />
        </div>
        <div class="text-[17px] font-bold text-sig-bad">Kode QR Tidak Terdaftar</div>
        <p class="text-[13px] text-txt-dim leading-relaxed max-w-sm">
          Barang tidak ditemukan di sistem inventaris. Kemungkinan kode QR salah, barang telah dihapus, atau belum diinput ke master data.
        </p>
        <div class="px-3 py-1.5 rounded-[6px] t-mono !text-[12px] text-txt-dim bg-panel-2 border border-line break-all">
          ID: {{ route.params.id }}
        </div>
        <div class="flex flex-col sm:flex-row gap-2 w-full mt-2 pt-4 border-t border-line">
          <Button label="Dashboard" icon="pi pi-th-large" class="flex-1" @click="router.push('/')" />
          <Button label="Daftar Barang" icon="pi pi-box" severity="secondary" outlined class="flex-1" @click="router.push('/items')" />
        </div>
      </div>

      <template v-else-if="item">
        <!-- ============ Hero ============ -->
        <section class="panel overflow-hidden relative">
          <div class="grid-bg absolute inset-0 opacity-40 pointer-events-none" aria-hidden="true" />
          <div
            class="absolute -top-20 -right-10 w-[280px] h-[280px] rounded-full pointer-events-none"
            style="background: radial-gradient(circle, var(--acc-wash) 0%, transparent 70%);"
            aria-hidden="true"
          />

          <div class="relative p-4 sm:p-5">
            <div class="inline-flex items-center gap-2 t-mono !text-[11px] uppercase tracking-[0.08em] text-acc-600 dark:text-acc-300 px-2.5 py-1 rounded-full border border-[var(--acc-border)] bg-acc-50 mb-3">
              <span class="w-1.5 h-1.5 rounded-full bg-acc-500" aria-hidden="true" />
              <span>{{ scannedUnit ? 'Unit Fisik Terpindai' : 'Hasil Pindai QR' }}</span>
            </div>

            <div class="flex items-start gap-3.5 sm:gap-4">
              <button
                v-if="item.photo_url"
                type="button"
                class="w-20 h-20 sm:w-24 sm:h-24 rounded-[12px] border border-line overflow-hidden shrink-0 bg-panel cursor-pointer"
                aria-label="Perbesar foto barang"
                @click="openPhoto(item.photo_url)"
              >
                <img :src="item.photo_url" :alt="item.name" class="w-full h-full object-cover" />
              </button>
              <div
                v-else
                class="w-20 h-20 sm:w-24 sm:h-24 rounded-[12px] border border-line bg-panel grid place-items-center shrink-0"
                aria-hidden="true"
              >
                <i class="pi pi-box text-acc-500 text-2xl" />
              </div>

              <div class="min-w-0 flex-1">
                <div class="flex items-center gap-1.5 flex-wrap mb-1.5">
                  <span class="t-mono !text-[11.5px] font-semibold px-2 py-0.5 rounded border border-line bg-panel text-txt-dim break-all">
                    {{ item.sku }}
                  </span>
                  <StatusChip v-if="item.condition_status" :kind="item.condition_status" />
                </div>
                <h1 class="text-[20px] sm:text-[24px] font-extrabold tracking-[-0.025em] text-txt leading-tight break-words">
                  {{ item.name }}
                </h1>
                <div class="text-[13px] text-txt-dim mt-1.5 flex items-center gap-1.5">
                  <i class="pi pi-map-marker text-[12px] text-acc-500" aria-hidden="true" />
                  <span>{{ item.location || 'Lokasi belum ditentukan' }}</span>
                  <span v-if="item.location_code" class="t-mono !text-[11.5px]">({{ item.location_code }})</span>
                </div>
              </div>
            </div>

            <!-- Metric tiles: Total Unit selalu tampil pertama -->
            <dl class="grid grid-cols-3 gap-2.5 mt-4">
              <div
                class="rounded-[10px] border p-3 flex flex-col justify-between min-h-[78px]"
                :class="isLowStock ? 'border-[var(--sig-bad-line)] bg-[var(--sig-bad-bg)]' : 'border-[var(--acc-border)] bg-acc-50'"
              >
                <dt class="text-[11.5px] font-semibold uppercase tracking-[0.04em] text-txt-dim">Total Unit</dt>
                <dd class="flex items-baseline gap-1.5 flex-wrap">
                  <span
                    class="t-num text-[28px] leading-none font-extrabold"
                    :class="isLowStock ? 'text-sig-bad' : 'text-acc-600 dark:text-acc-300'"
                    data-testid="total-units"
                  >{{ totalUnits }}</span>
                  <span class="text-[12px] text-txt-dim leading-tight">{{ totalUnitsCaption }}</span>
                </dd>
              </div>
              <div class="rounded-[10px] border border-line bg-panel p-3 flex flex-col justify-between min-h-[78px]">
                <dt class="text-[11.5px] font-semibold uppercase tracking-[0.04em] text-txt-dim">
                  {{ isStockItem ? 'Min. Stok' : 'Jenis' }}
                </dt>
                <dd class="text-[13px] font-semibold text-txt leading-snug">
                  <template v-if="isStockItem"><span class="t-num text-[20px] font-bold">{{ item.min_stock }}</span> <span class="text-txt-dim font-normal text-[12px]">{{ item.unit }}</span></template>
                  <template v-else>Aset Tetap (BMD)</template>
                </dd>
              </div>
              <div class="rounded-[10px] border border-line bg-panel p-3 flex flex-col justify-between min-h-[78px]">
                <dt class="text-[11.5px] font-semibold uppercase tracking-[0.04em] text-txt-dim">Kategori</dt>
                <dd class="text-[13px] font-semibold text-txt leading-snug break-words">{{ item.category || '—' }}</dd>
              </div>
            </dl>

            <p v-if="isLowStock" class="mt-2.5 text-[12.5px] text-sig-bad font-medium flex items-center gap-1.5" role="status">
              <i class="pi pi-exclamation-triangle text-[12px]" aria-hidden="true" />
              Stok berada di bawah atau sama dengan batas minimum.
            </p>

            <!-- Banner unit yang dipindai -->
            <div
              v-if="scannedUnit"
              class="mt-3 p-3 rounded-[10px] border border-acc-500 bg-panel flex items-center justify-between gap-3 flex-wrap"
            >
              <div class="text-[13px] text-txt">
                <span class="font-bold">Unit #{{ scannedUnit.unit_number }}</span>
                <span class="text-txt-dim"> dari {{ totalUnits }} unit terdaftar</span>
              </div>
              <span v-if="scannedUnit.register_number" class="t-mono !text-[12px] text-txt-dim">{{ scannedUnit.register_number }}</span>
            </div>

            <!-- Actions -->
            <div class="mt-4 flex flex-col sm:flex-row gap-2">
              <Button
                label="Laporkan Masalah"
                icon="pi pi-megaphone"
                severity="warn"
                class="flex-1 !min-h-[44px]"
                @click="router.push(`/lapor/${item.id}`)"
              />
              <Button
                v-if="isLoggedIn"
                label="Buka di Panel Admin"
                icon="pi pi-arrow-right"
                iconPos="right"
                class="flex-1 !min-h-[44px]"
                @click="router.push(`/items/${item.id}`)"
              />
              <Button
                v-else
                label="Masuk Petugas"
                icon="pi pi-sign-in"
                severity="secondary"
                outlined
                class="flex-1 !min-h-[44px]"
                @click="router.push(`/login?redirect=/items/${item.id}`)"
              />
            </div>
          </div>
        </section>

        <!-- ============ Tabs + Content ============ -->
        <section class="panel overflow-hidden">
          <div
            role="tablist"
            aria-label="Informasi barang"
            class="grid gap-1.5 p-1.5 border-b border-line bg-panel-2"
            :style="{ gridTemplateColumns: `repeat(${tabs.length}, minmax(0, 1fr))` }"
          >
            <button
              v-for="t in tabs"
              :id="'tab-' + t.id"
              :key="t.id"
              type="button"
              role="tab"
              :aria-selected="activeTab === t.id"
              :aria-controls="'panel-' + t.id"
              class="min-h-[44px] px-2 rounded-[8px] text-[13px] font-semibold transition-colors cursor-pointer flex items-center justify-center gap-1.5 focus-visible:outline-2 focus-visible:outline-acc-500"
              :class="
                activeTab === t.id
                  ? 'bg-acc-500 text-white dark:text-[#042f2e]'
                  : 'text-txt-dim hover:text-txt hover:bg-panel'
              "
              @click="activeTab = t.id"
            >
              <span>{{ t.label }}</span>
              <span
                v-if="t.count !== undefined"
                class="t-num !text-[11.5px] px-1.5 rounded-full"
                :class="activeTab === t.id ? 'bg-white/25' : 'bg-panel border border-line'"
              >{{ t.count }}</span>
            </button>
          </div>

          <!-- Spesifikasi -->
          <div
            v-if="activeTab === 'detail'"
            id="panel-detail"
            role="tabpanel"
            aria-labelledby="tab-detail"
            class="p-4 sm:p-5 space-y-5"
          >
            <dl class="divide-y divide-line text-[13px]">
              <div v-for="[k, v] in rows()" :key="k" class="py-2.5 flex items-start justify-between gap-4">
                <dt class="text-txt-dim shrink-0">{{ k }}</dt>
                <dd class="font-medium text-txt text-right break-words">{{ v }}</dd>
              </div>
            </dl>

            <div v-if="units.length" class="space-y-2.5 pt-4 border-t border-line">
              <div class="flex items-center justify-between gap-2">
                <h2 class="flex items-center gap-1.5 text-[13.5px] font-bold text-txt">
                  <i class="pi pi-th-large text-acc-500 text-[13px]" aria-hidden="true" />
                  Unit Fisik Terdaftar ({{ units.length }})
                </h2>
                <button
                  type="button"
                  class="min-h-[44px] px-2 text-[12.5px] font-semibold text-acc-500 hover:underline cursor-pointer"
                  @click="activeTab = 'units'"
                >
                  Lihat Tab Unit
                </button>
              </div>
              <div class="space-y-2.5">
                <ScanUnitCard
                  v-for="u in units"
                  :key="u.id"
                  :unit="u"
                  :scanned="scannedUnitId === u.id"
                  compact
                  @photo="openPhoto"
                />
              </div>
            </div>

            <div v-if="hasCoords" class="p-3.5 rounded-[10px] border border-line bg-panel-2 space-y-3">
              <div class="flex items-center justify-between gap-2 flex-wrap">
                <div class="flex items-center gap-1.5 text-[13px] font-semibold text-txt">
                  <i class="pi pi-compass text-acc-500 text-[13px]" aria-hidden="true" />
                  <span>Koordinat GPS Tercatat</span>
                </div>
                <span class="t-mono !text-[12px] text-txt-dim">
                  {{ Number(item.geo_lat).toFixed(5) }}, {{ Number(item.geo_lng).toFixed(5) }}
                </span>
              </div>
              <div class="flex flex-col sm:flex-row gap-2">
                <Button
                  :label="showMap ? 'Sembunyikan Peta' : 'Lihat Peta Lokasi'"
                  icon="pi pi-map"
                  severity="secondary"
                  outlined
                  class="flex-1 !min-h-[44px]"
                  @click="showMap = !showMap"
                />
                <a
                  :href="googleMapsUrl"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="min-h-[44px] inline-flex items-center justify-center gap-1.5 px-3 rounded-[8px] border border-line bg-panel hover:border-[var(--acc-border)] text-[13px] font-semibold text-txt transition-colors"
                >
                  <i class="pi pi-external-link text-[12px]" aria-hidden="true" />
                  <span>Google Maps</span>
                </a>
              </div>
              <iframe
                v-if="showMap"
                :src="osmEmbedUrl"
                title="Peta lokasi aset"
                class="w-full h-52 rounded-[8px] border border-line bg-slate-100"
                loading="lazy"
                referrerpolicy="no-referrer-when-downgrade"
              />
            </div>
          </div>

          <!-- Unit Fisik -->
          <div
            v-else-if="activeTab === 'units'"
            id="panel-units"
            role="tabpanel"
            aria-labelledby="tab-units"
            class="p-4 sm:p-5 space-y-3"
          >
            <div class="flex items-center justify-between gap-2 flex-wrap">
              <h2 class="text-[13.5px] font-bold text-txt">Semua Unit Fisik Terdaftar ({{ units.length }})</h2>
              <span class="text-[12px] text-txt-dim">Ketuk foto untuk memperbesar</span>
            </div>
            <div class="grid gap-3 sm:grid-cols-2">
              <ScanUnitCard
                v-for="u in units"
                :key="u.id"
                :unit="u"
                :scanned="scannedUnitId === u.id"
                @photo="openPhoto"
              />
            </div>
          </div>

          <!-- Servis -->
          <div
            v-else
            id="panel-maintenance"
            role="tabpanel"
            aria-labelledby="tab-maintenance"
            class="p-4 sm:p-5"
          >
            <div v-if="!maintenances.length" class="py-10 text-center">
              <i class="pi pi-wrench text-2xl text-txt-muted mb-2 block" aria-hidden="true" />
              <div class="text-[14px] font-semibold text-txt">Belum ada catatan pemeliharaan</div>
              <div class="text-[12.5px] text-txt-dim mt-1">Riwayat servis dan kalibrasi alat ini akan tampil di sini.</div>
            </div>

            <ol v-else class="space-y-2.5">
              <li
                v-for="m in maintenances"
                :key="m.id"
                class="p-3.5 rounded-[10px] border border-line bg-panel-2 space-y-1.5 text-[13px]"
              >
                <div class="flex items-center justify-between gap-2">
                  <span class="font-semibold text-acc-500">{{ m.service_type }}</span>
                  <span class="t-mono !text-[12px] text-txt-dim">{{ m.service_date }}</span>
                </div>
                <div v-if="m.vendor_or_technician" class="text-txt-dim text-[12.5px]">
                  Pelaksana: <strong class="text-txt">{{ m.vendor_or_technician }}</strong>
                </div>
                <p v-if="m.description" class="text-txt leading-relaxed">{{ m.description }}</p>
                <div class="flex items-center justify-between pt-1.5 border-t border-line-soft text-[12px]">
                  <span v-if="m.next_service_date" class="text-sig-ok font-semibold">Berikutnya: {{ m.next_service_date }}</span>
                  <span v-else />
                  <a
                    v-if="safeHref(m.photo_url)"
                    :href="safeHref(m.photo_url)"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="text-acc-500 hover:underline font-medium min-h-[32px] inline-flex items-center"
                  >
                    Lihat Bukti
                  </a>
                </div>
              </li>
            </ol>
          </div>
        </section>
      </template>
    </main>

    <!-- Photo Lightbox -->
    <Dialog v-model:visible="photoModal" modal header="Foto Aset" :style="{ width: '92vw', maxWidth: '640px' }">
      <div class="flex flex-col items-center gap-3">
        <img :src="selectedPhoto" alt="Foto aset diperbesar" class="max-h-[70vh] w-auto rounded-[8px] border border-line object-contain" />
        <Button
          label="Unduh Foto"
          icon="pi pi-download"
          severity="secondary"
          outlined
          @click="downloadImage(selectedPhoto, `${item?.sku || 'aset'}.jpg`)"
        />
      </div>
    </Dialog>
  </div>
</template>
