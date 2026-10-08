<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import api from '@/api'
import { useToast } from 'primevue/usetoast'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'

const props = withDefaults(
  defineProps<{
    modelValue?: string
    label?: string
    locationName?: string
    hint?: string
    geoLat?: number | null
    geoLng?: number | null
    geoAcc?: number | null
  }>(),
  {
    modelValue: '',
    label: 'Foto Barang',
    locationName: '',
    hint: 'Foto akan otomatis dicap watermark GPS, waktu WIB, dan nama petugas.',
    geoLat: null,
    geoLng: null,
    geoAcc: null,
  }
)

const emit = defineEmits<{
  (e: 'update:modelValue', val: string): void
  (e: 'update:geoLat', val: number | null): void
  (e: 'update:geoLng', val: number | null): void
  (e: 'update:geoAcc', val: number | null): void
  (e: 'update:geoName', val: string): void
}>()

const toast = useToast()

const uploading = ref(false)
const gettingGeo = ref(false)
const coords = ref<{ lat: number; lng: number; acc: number } | null>(null)
const geoError = ref('')
const geoLabel = ref(props.locationName || '')
const fileInput = ref<HTMLInputElement | null>(null)
const previewOpen = ref(false)
const showInlineMap = ref(false)
const modalTab = ref<'photo' | 'map'>('photo')
const localPreviewUrl = ref('')

const currentPhotoSrc = computed(() => props.modelValue || localPreviewUrl.value)

// Sync props if provided from existing item
watch(
  () => [props.geoLat, props.geoLng],
  ([lat, lng]) => {
    if (lat && lng && !coords.value) {
      coords.value = {
        lat: Number(lat),
        lng: Number(lng),
        acc: Number(props.geoAcc || 10),
      }
    }
  },
  { immediate: true }
)

function fetchGeo() {
  if (!navigator.geolocation) {
    geoError.value = 'Browser tidak mendukung GPS Geolocation'
    return
  }
  gettingGeo.value = true
  geoError.value = ''
  navigator.geolocation.getCurrentPosition(
    (pos) => {
      coords.value = {
        lat: pos.coords.latitude,
        lng: pos.coords.longitude,
        acc: Math.round(pos.coords.accuracy),
      }
      gettingGeo.value = false
      emit('update:geoLat', coords.value.lat)
      emit('update:geoLng', coords.value.lng)
      emit('update:geoAcc', coords.value.acc)
    },
    (err) => {
      gettingGeo.value = false
      switch (err.code) {
        case err.PERMISSION_DENIED:
          geoError.value = 'Izin lokasi ditolak di browser'
          break
        case err.POSITION_UNAVAILABLE:
          geoError.value = 'Sinyal GPS tidak tersedia'
          break
        case err.TIMEOUT:
          geoError.value = 'Waktu deteksi GPS habis'
          break
        default:
          geoError.value = 'Gagal mendeteksi lokasi'
      }
    },
    { enableHighAccuracy: true, timeout: 10000, maximumAge: 0 }
  )
}

onMounted(() => {
  if (!coords.value) {
    fetchGeo()
  }
})

const osmEmbedUrl = computed(() => {
  if (!coords.value) return ''
  const delta = 0.0035
  const minLng = coords.value.lng - delta
  const minLat = coords.value.lat - delta
  const maxLng = coords.value.lng + delta
  const maxLat = coords.value.lat + delta
  const bbox = `${minLng}%2C${minLat}%2C${maxLng}%2C${maxLat}`
  return `https://www.openstreetmap.org/export/embed.html?bbox=${bbox}&layer=mapnik&marker=${coords.value.lat}%2C${coords.value.lng}`
})

const googleMapsUrl = computed(() => {
  if (!coords.value) return ''
  return `https://www.google.com/maps?q=${coords.value.lat},${coords.value.lng}`
})

function triggerSelect(useCamera = false) {
  if (!fileInput.value) return
  if (useCamera) {
    fileInput.value.setAttribute('capture', 'environment')
  } else {
    fileInput.value.removeAttribute('capture')
  }
  fileInput.value.click()
}

async function onFileSelected(e: Event) {
  const target = e.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  if (localPreviewUrl.value) URL.revokeObjectURL(localPreviewUrl.value)
  localPreviewUrl.value = URL.createObjectURL(file)

  uploading.value = true
  try {
    // Pastikan koordinat GPS didapatkan sebelum unggah agar cap geotag selalu tercetak
    if (!coords.value && navigator.geolocation) {
      await new Promise<void>((resolve) => {
        navigator.geolocation.getCurrentPosition(
          (pos) => {
            coords.value = {
              lat: pos.coords.latitude,
              lng: pos.coords.longitude,
              acc: Math.round(pos.coords.accuracy),
            }
            emit('update:geoLat', coords.value.lat)
            emit('update:geoLng', coords.value.lng)
            emit('update:geoAcc', coords.value.acc)
            resolve()
          },
          () => resolve(),
          { enableHighAccuracy: true, timeout: 6000, maximumAge: 0 }
        )
      })
    }

    const fd = new FormData()
    fd.append('photo', file)
    if (coords.value) {
      fd.append('geo_lat', coords.value.lat.toString())
      fd.append('geo_lng', coords.value.lng.toString())
      fd.append('geo_acc', coords.value.acc.toString())
    }
    const locName = geoLabel.value || props.locationName
    if (locName) {
      fd.append('geo_name', locName)
      emit('update:geoName', locName)
    }

    const res = await api.post('/upload', fd, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    const url = res.data.url
    emit('update:modelValue', url)
    if (res.data.geo_lat && res.data.geo_lng) {
      coords.value = {
        lat: res.data.geo_lat,
        lng: res.data.geo_lng,
        acc: res.data.geo_acc || 10,
      }
      emit('update:geoLat', res.data.geo_lat)
      emit('update:geoLng', res.data.geo_lng)
      emit('update:geoAcc', res.data.geo_acc)
    }

    modalTab.value = 'photo'
    previewOpen.value = true
    toast.add({
      severity: 'success',
      summary: 'Foto & Cap Geotag Berhasil',
      detail: coords.value
        ? `Watermark GPS (${coords.value.lat.toFixed(4)}, ${coords.value.lng.toFixed(4)}) tercetak pada foto`
        : 'Foto tersimpan (tanpa koordinat GPS)',
      life: 3500,
    })
  } catch (err: any) {
    toast.add({
      severity: 'error',
      summary: 'Gagal mengunggah foto',
      detail: err.response?.data?.error || err.message,
      life: 4000,
    })
  } finally {
    uploading.value = false
    if (target) target.value = ''
  }
}

function removePhoto() {
  if (localPreviewUrl.value) {
    URL.revokeObjectURL(localPreviewUrl.value)
    localPreviewUrl.value = ''
  }
  emit('update:modelValue', '')
}
</script>

<template>
  <div class="flex flex-col gap-2">
    <div class="flex items-center justify-between">
      <span class="text-[11.5px] font-semibold" style="color: var(--txt-dim)">
        {{ label }}
      </span>
      <span v-if="hint" class="text-[10.5px]" style="color: var(--txt-dim)">
        {{ hint }}
      </span>
    </div>

    <!-- Hidden file input -->
    <input
      ref="fileInput"
      type="file"
      accept="image/*"
      class="hidden"
      @change="onFileSelected"
    />

    <!-- GPS Bar -->
    <div
      class="flex flex-wrap items-center justify-between gap-2 px-3 py-2 rounded border text-[11.5px]"
      style="background: var(--paper-1); border-color: var(--line)"
    >
      <div class="flex items-center gap-2">
        <i
          class="pi"
          :class="gettingGeo ? 'pi-spin pi-spinner text-acc-500' : coords ? 'pi-map-marker text-sig-ok' : 'pi-map-marker text-sig-bad'"
        />
        <div v-if="gettingGeo" class="t-mono text-[11px]" style="color: var(--txt-dim)">
          Mendeteksi GPS…
        </div>
        <div v-else-if="coords" class="flex flex-wrap items-center gap-1.5">
          <span class="font-semibold text-sig-ok">GPS Terkunci:</span>
          <span class="t-mono font-medium">{{ coords.lat.toFixed(6) }}, {{ coords.lng.toFixed(6) }}</span>
          <span class="t-mono text-[10.5px]" style="color: var(--txt-dim)">(±{{ coords.acc }}m)</span>
        </div>
        <div v-else class="text-sig-bad">
          {{ geoError || 'GPS tidak terdeteksi' }}
        </div>
      </div>

      <div class="flex items-center gap-1">
        <Button
          v-if="coords"
          :label="showInlineMap ? 'Tutup Peta' : 'Peta'"
          :icon="showInlineMap ? 'pi pi-map-marker' : 'pi pi-map'"
          text
          size="small"
          :severity="showInlineMap ? 'warn' : 'secondary'"
          @click="showInlineMap = !showInlineMap"
        />
        <Button
          icon="pi pi-refresh"
          text
          rounded
          size="small"
          severity="secondary"
          :loading="gettingGeo"
          v-tooltip.top="'Segarkan koordinat GPS'"
          @click="fetchGeo"
        />
      </div>
    </div>

    <!-- Mini Inline Map Box (OpenStreetMap) -->
    <div
      v-if="showInlineMap && coords"
      class="rounded-lg overflow-hidden border flex flex-col"
      style="border-color: var(--line); background: var(--paper-1)"
    >
      <div class="relative w-full h-44 bg-stone-900">
        <iframe
          :src="osmEmbedUrl"
          class="w-full h-full border-0"
          loading="lazy"
          title="OpenStreetMap Pin"
        />
      </div>
      <div class="flex items-center justify-between px-3 py-1.5 text-[11px] border-t" style="border-color: var(--line)">
        <span class="t-mono text-sig-ok font-semibold flex items-center gap-1">
          <i class="pi pi-compass text-[10px]" /> {{ coords.lat.toFixed(5) }}, {{ coords.lng.toFixed(5) }}
        </span>
        <a
          :href="googleMapsUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="text-acc-500 hover:underline flex items-center gap-1 font-medium"
        >
          <i class="pi pi-external-link text-[10px]" /> Buka di Google Maps
        </a>
      </div>
    </div>

    <!-- Optional Location label input when setting up -->
    <div v-if="!modelValue" class="flex items-center gap-2">
      <InputText
        v-model="geoLabel"
        size="small"
        placeholder="Nama ruangan / lokasi stempel (opsional)"
        class="w-full !text-[12px]"
      />
    </div>

    <!-- Preview / Upload Container -->
    <div
      v-if="currentPhotoSrc"
      class="rounded-lg overflow-hidden border flex flex-col gap-2.5 p-3"
      style="background: var(--paper-1); border-color: var(--line)"
    >
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-1.5 text-[12px] font-bold text-sig-ok">
          <i class="pi pi-check-circle" /> Foto Berstempel Geotag
        </div>
        <div class="flex items-center gap-1">
          <Button
            label="Perbesar"
            icon="pi pi-search-plus"
            size="small"
            text
            severity="secondary"
            @click="modalTab = 'photo'; previewOpen = true"
          />
          <Button
            v-if="coords"
            label="Peta"
            icon="pi pi-map"
            size="small"
            text
            severity="secondary"
            @click="modalTab = 'map'; previewOpen = true"
          />
          <Button
            label="Ganti"
            icon="pi pi-camera"
            size="small"
            text
            severity="secondary"
            :loading="uploading"
            @click="triggerSelect(false)"
          />
          <Button
            icon="pi pi-trash"
            size="small"
            text
            severity="danger"
            v-tooltip.top="'Hapus Foto'"
            @click="removePhoto"
          />
        </div>
      </div>

      <!-- Prominent Preview Image with Cap Geotag clearly visible -->
      <div
        class="relative w-full rounded-md overflow-hidden bg-black/40 border flex items-center justify-center cursor-pointer group"
        style="border-color: var(--line); min-height: 180px; max-height: 280px"
        @click="modalTab = 'photo'; previewOpen = true"
      >
        <img
          :src="currentPhotoSrc"
          @error="(e) => { if (localPreviewUrl && (e.target as HTMLImageElement).src !== localPreviewUrl) (e.target as HTMLImageElement).src = localPreviewUrl }"
          alt="Foto Berstempel Geotag"
          class="w-full h-auto max-h-[280px] object-contain group-hover:scale-[1.01] transition-transform"
        />
        <div
          class="absolute inset-0 bg-black/30 opacity-0 group-hover:opacity-100 flex items-center justify-center text-white text-[12px] font-semibold transition-opacity"
        >
          <span class="bg-black/75 px-3 py-1.5 rounded-full flex items-center gap-1.5">
            <i class="pi pi-eye" /> Klik untuk melihat ukuran penuh
          </span>
        </div>
      </div>

      <div class="flex items-center justify-between text-[11px] px-1" style="color: var(--txt-dim)">
        <span>Watermark GPS, tanggal WIB, dan identitas petugas tercetak langsung di dalam foto</span>
      </div>
    </div>

    <!-- Empty Upload State -->
    <div
      v-else
      class="border-2 border-dashed rounded-lg p-4 text-center flex flex-col items-center justify-center gap-2.5 transition-colors"
      style="border-color: var(--line); background: var(--paper-1)"
    >
      <div
        class="w-10 h-10 rounded-full flex items-center justify-center text-acc-500"
        style="background: var(--paper-2)"
      >
        <i v-if="uploading" class="pi pi-spin pi-spinner text-lg" />
        <i v-else class="pi pi-camera text-lg" />
      </div>

      <div class="flex flex-col gap-0.5">
        <span class="text-[12.5px] font-semibold">
          {{ uploading ? 'Memproses & mengecap geotag…' : 'Ambil foto berkamera atau unggah berkas' }}
        </span>
        <span class="text-[11px]" style="color: var(--txt-dim)">
          Format JPG / PNG, ukuran otomatis disesuaikan
        </span>
      </div>

      <div class="flex flex-wrap gap-2 mt-1">
        <Button
          label="Buka Kamera"
          icon="pi pi-camera"
          size="small"
          severity="primary"
          :loading="uploading"
          @click="triggerSelect(true)"
        />
        <Button
          label="Pilih Berkas"
          icon="pi pi-upload"
          size="small"
          severity="secondary"
          outlined
          :loading="uploading"
          @click="triggerSelect(false)"
        />
      </div>
    </div>

    <!-- Preview Modal with Tab (Foto Berstempel vs Peta Geotag) -->
    <div
      v-if="previewOpen"
      class="fixed inset-0 z-50 bg-black/85 flex items-center justify-center p-4 backdrop-blur-xs"
      @click.self="previewOpen = false"
    >
      <div
        class="relative max-w-3xl w-full bg-[var(--paper-1)] rounded-lg overflow-hidden border shadow-2xl flex flex-col"
        style="border-color: var(--line)"
      >
        <!-- Modal Header with Tabs -->
        <div class="flex items-center justify-between px-4 py-2.5 border-b gap-3" style="border-color: var(--line)">
          <div class="flex items-center gap-2">
            <button
              class="px-3 py-1 rounded text-[12px] font-bold flex items-center gap-1.5 transition-colors"
              :class="modalTab === 'photo' ? 'bg-acc-500 text-ink-950' : 'text-ink-300 hover:bg-paper-2'"
              @click="modalTab = 'photo'"
            >
              <i class="pi pi-image text-[11px]" /> Foto Berstempel
            </button>
            <button
              v-if="coords"
              class="px-3 py-1 rounded text-[12px] font-bold flex items-center gap-1.5 transition-colors"
              :class="modalTab === 'map' ? 'bg-acc-500 text-ink-950' : 'text-ink-300 hover:bg-paper-2'"
              @click="modalTab = 'map'"
            >
              <i class="pi pi-map text-[11px]" /> Peta Lokasi (OpenStreetMap)
            </button>
          </div>
          <Button
            icon="pi pi-times"
            text
            rounded
            size="small"
            severity="secondary"
            @click="previewOpen = false"
          />
        </div>

        <!-- Modal Body -->
        <div class="bg-black flex items-center justify-center min-h-[50vh] max-h-[72vh] overflow-hidden">
          <!-- Tab 1: Foto Geotag Penuh -->
          <div v-if="modalTab === 'photo'" class="p-3 w-full h-full flex items-center justify-center overflow-auto">
            <img
              :src="currentPhotoSrc"
              @error="(e) => { if (localPreviewUrl && (e.target as HTMLImageElement).src !== localPreviewUrl) (e.target as HTMLImageElement).src = localPreviewUrl }"
              alt="Foto Geotag Penuh"
              class="max-h-[68vh] object-contain rounded"
            />
          </div>

          <!-- Tab 2: Peta Lokasi Geotag Interaktif -->
          <div v-else-if="modalTab === 'map' && coords" class="relative w-full h-[65vh]">
            <iframe
              :src="osmEmbedUrl"
              class="w-full h-full border-0"
              title="Peta Lokasi OpenStreetMap"
            />
          </div>
        </div>

        <!-- Modal Footer -->
        <div class="flex items-center justify-between px-4 py-2.5 border-t text-[11.5px]" style="border-color: var(--line)">
          <div v-if="coords" class="flex items-center gap-2 t-mono">
            <i class="pi pi-map-marker text-sig-ok" />
            <span>{{ coords.lat.toFixed(6) }}, {{ coords.lng.toFixed(6) }}</span>
            <span style="color: var(--txt-dim)">(±{{ coords.acc }}m)</span>
          </div>
          <div v-else></div>

          <div class="flex items-center gap-2">
            <a
              v-if="coords"
              :href="googleMapsUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex"
            >
              <Button label="Google Maps" icon="pi pi-map-marker" size="small" text />
            </a>
            <a :href="modelValue" target="_blank" download class="inline-flex">
              <Button label="Unduh Foto" icon="pi pi-download" size="small" text severity="secondary" />
            </a>
            <Button label="Tutup" size="small" severity="secondary" @click="previewOpen = false" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
