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
    itemName?: string
    locationName?: string
    hint?: string
    geoLat?: number | null
    geoLng?: number | null
    geoAcc?: number | null
  }>(),
  {
    modelValue: '',
    label: 'Foto Barang',
    itemName: '',
    locationName: '',
    hint: 'Foto akan otomatis dikompresi dan dicap watermark GPS Map Camera.',
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
const lastRawFile = ref<File | null>(null)

// Animated Upload State
const isUploadingModal = ref(false)
const uploadProgress = ref(0)
const uploadStep = ref<'gps' | 'stamp' | 'compress' | 'upload' | 'success'>('gps')
const currentStepText = ref('')
let uploadAbortCtrl: AbortController | null = null

function cancelUpload() {
  if (uploadAbortCtrl) {
    uploadAbortCtrl.abort()
    uploadAbortCtrl = null
  }
  uploading.value = false
  isUploadingModal.value = false
  previewOpen.value = false
  removePhoto()
  toast.add({ severity: 'info', summary: 'Foto Dibatalkan', detail: 'Unggahan dibatalkan dan foto dihapus.', life: 2500 })
}

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

function fetchGeo(): Promise<{ lat: number; lng: number; acc: number } | null> {
  return new Promise((resolve) => {
    if (!navigator.geolocation) {
      geoError.value = 'Browser tidak mendukung GPS Geolocation'
      resolve(null)
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
        resolve(coords.value)
      },
      (err) => {
        gettingGeo.value = false
        switch (err.code) {
          case err.PERMISSION_DENIED:
            geoError.value = 'Izin GPS ditolak di browser'
            break
          case err.POSITION_UNAVAILABLE:
            geoError.value = 'Sinyal GPS tidak tersedia'
            break
          case err.TIMEOUT:
            geoError.value = 'Waktu deteksi GPS habis'
            break
          default:
            geoError.value = 'Gagal mendeteksi lokasi GPS'
        }
        resolve(null)
      },
      { enableHighAccuracy: true, timeout: 8000, maximumAge: 0 }
    )
  })
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

// Reverse geocoding helper (OpenStreetMap Nominatim)
async function getReverseGeocode(lat: number, lng: number): Promise<{ title: string; fullAddress: string }> {
  try {
    const controller = new AbortController()
    const timer = setTimeout(() => controller.abort(), 3500)
    const res = await fetch(
      `https://nominatim.openstreetmap.org/reverse?format=json&lat=${lat}&lon=${lng}&zoom=18&addressdetails=1`,
      {
        headers: { 'Accept-Language': 'id,en' },
        signal: controller.signal,
      }
    )
    clearTimeout(timer)
    if (!res.ok) throw new Error('Status ' + res.status)
    const data = await res.json()
    const addr = data.address || {}

    const district = addr.suburb || addr.municipality || addr.district || addr.city_district || addr.city || addr.town || addr.village || 'Lokasi'
    const region = addr.state || addr.region || ''
    const country = addr.country || 'Indonesia'

    const titleParts = [district]
    if (region && region !== district) titleParts.push(region)
    titleParts.push(country)

    const title = titleParts.join(', ')
    const fullAddress = data.display_name || `${lat}, ${lng}`
    return { title, fullAddress }
  } catch (e: any) {
    return {
      title: 'Indonesia',
      fullAddress: `Koordinat GPS: Lat ${lat.toFixed(6)}°, Long ${lng.toFixed(6)}°`,
    }
  }
}

// Fetch map tile as local blob to prevent tainted canvas
function latLngToTile(lat: number, lng: number, zoom: number) {
  const n = Math.pow(2, zoom)
  const x = Math.floor(((lng + 180) / 360) * n)
  const latRad = (lat * Math.PI) / 180
  const y = Math.floor(((1 - Math.log(Math.tan(latRad) + 1 / Math.cos(latRad)) / Math.PI) / 2) * n)
  return { x, y }
}

async function loadTileBlobImage(lat: number, lng: number): Promise<HTMLImageElement | null> {
  try {
    const { x, y } = latLngToTile(lat, lng, 16)
    const tileUrl = `https://tile.openstreetmap.org/16/${x}/${y}.png`
    const controller = new AbortController()
    const timer = setTimeout(() => controller.abort(), 2000)
    const res = await fetch(tileUrl, { signal: controller.signal })
    clearTimeout(timer)
    if (!res.ok) return null
    const blob = await res.blob()
    const blobUrl = URL.createObjectURL(blob)
    return new Promise((resolve) => {
      const img = new Image()
      img.onload = () => resolve(img)
      img.onerror = () => resolve(null)
      img.src = blobUrl
    })
  } catch {
    return null
  }
}

function formatGpsDate(d: Date): string {
  const days = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday']
  const dayName = days[d.getDay()]
  const day = String(d.getDate()).padStart(2, '0')
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const year = d.getFullYear()

  let hours = d.getHours()
  const minutes = String(d.getMinutes()).padStart(2, '0')
  const ampm = hours >= 12 ? 'PM' : 'AM'
  hours = hours % 12
  hours = hours ? hours : 12
  const hourStr = String(hours).padStart(2, '0')

  const offsetMin = -d.getTimezoneOffset()
  const sign = offsetMin >= 0 ? '+' : '-'
  const offH = String(Math.floor(Math.abs(offsetMin) / 60)).padStart(2, '0')
  const offM = String(Math.abs(offsetMin) % 60).padStart(2, '0')
  const tzStr = `GMT ${sign}${offH}:${offM}`

  return `${dayName}, ${day}/${month}/${year} ${hourStr}:${minutes} ${ampm} ${tzStr}`
}

function wrapText(ctx: CanvasRenderingContext2D, text: string, maxWidth: number): string[] {
  const words = text.split(' ')
  const lines: string[] = []
  let currentLine = ''

  for (const word of words) {
    const testLine = currentLine ? `${currentLine} ${word}` : word
    const metrics = ctx.measureText(testLine)
    if (metrics.width > maxWidth && currentLine) {
      lines.push(currentLine)
      currentLine = word
    } else {
      currentLine = testLine
    }
  }
  if (currentLine) {
    lines.push(currentLine)
  }
  return lines
}

function drawPin(ctx: CanvasRenderingContext2D, cx: number, cy: number, size: number) {
  ctx.save()
  ctx.fillStyle = 'rgba(0, 0, 0, 0.45)'
  ctx.beginPath()
  ctx.ellipse(cx, cy + size * 0.15, size * 0.35, size * 0.15, 0, 0, Math.PI * 2)
  ctx.fill()

  ctx.fillStyle = '#ea4335'
  ctx.beginPath()
  ctx.arc(cx, cy - size * 0.65, size * 0.45, Math.PI * 0.8, Math.PI * 0.2, false)
  ctx.lineTo(cx, cy)
  ctx.closePath()
  ctx.fill()

  ctx.fillStyle = '#ffffff'
  ctx.beginPath()
  ctx.arc(cx, cy - size * 0.65, size * 0.16, 0, Math.PI * 2)
  ctx.fill()

  ctx.restore()
}

function drawProceduralMap(ctx: CanvasRenderingContext2D, x: number, y: number, w: number, h: number) {
  ctx.save()
  ctx.fillStyle = '#2d3748'
  ctx.fillRect(x, y, w, h)

  ctx.fillStyle = '#22543d'
  ctx.fillRect(x + w * 0.1, y + h * 0.1, w * 0.35, h * 0.4)
  ctx.fillRect(x + w * 0.6, y + h * 0.5, w * 0.3, h * 0.35)

  ctx.strokeStyle = '#4a5568'
  ctx.lineWidth = Math.max(2, w * 0.04)
  ctx.beginPath()
  ctx.moveTo(x, y + h * 0.35)
  ctx.lineTo(x + w, y + h * 0.35)
  ctx.moveTo(x + w * 0.45, y)
  ctx.lineTo(x + w * 0.45, y + h)
  ctx.stroke()

  ctx.strokeStyle = '#d97706'
  ctx.lineWidth = Math.max(3, w * 0.05)
  ctx.beginPath()
  ctx.moveTo(x, y + h * 0.7)
  ctx.lineTo(x + w * 0.45, y + h * 0.35)
  ctx.lineTo(x + w, y + h * 0.2)
  ctx.stroke()

  ctx.restore()
}

function drawIndonesianFlag(ctx: CanvasRenderingContext2D, x: number, y: number, w: number, h: number) {
  ctx.save()
  ctx.shadowColor = 'transparent'
  // Top red half
  ctx.fillStyle = '#dc2626'
  ctx.fillRect(x, y, w, h / 2)
  // Bottom white half
  ctx.fillStyle = '#ffffff'
  ctx.fillRect(x, y + h / 2, w, h / 2)
  // Subtle border
  ctx.strokeStyle = 'rgba(255, 255, 255, 0.4)'
  ctx.lineWidth = 1
  ctx.strokeRect(x, y, w, h)
  ctx.restore()
}

// Compact GPS Map Camera Stamping + Compression
async function stampGpsMapCamera(
  file: File,
  geoCoords: { lat: number; lng: number; acc: number } | null,
  locName: string,
  itemNameText?: string
): Promise<{ blob: Blob; previewUrl: string; originalSizeKB: number; compressedSizeKB: number }> {
  return new Promise(async (resolve) => {
    currentStepText.value = 'Membaca berkas gambar...'

    const originalSizeKB = Math.round(file.size / 1024)
    const reader = new FileReader()

    reader.onload = async () => {
      const img = new Image()
      img.onload = async () => {
        const canvas = document.createElement('canvas')
        let width = img.width
        let height = img.height

        // Compress resolution: max 1920px (Full HD clarity for crisp text & map details)
        const maxDim = 1920
        if (width > maxDim || height > maxDim) {
          if (width > height) {
            height = Math.round((height * maxDim) / width)
            width = maxDim
          } else {
            width = Math.round((width * maxDim) / height)
            height = maxDim
          }
        }

        canvas.width = width
        canvas.height = height
        const ctx = canvas.getContext('2d')
        if (!ctx) {
          resolve({ blob: file, previewUrl: reader.result as string, originalSizeKB, compressedSizeKB: originalSizeKB })
          return
        }

        // Draw base photo
        ctx.drawImage(img, 0, 0, width, height)

        if (!geoCoords) {
          canvas.toBlob(
            (b) => {
              const compressedSizeKB = Math.round((b?.size || file.size) / 1024)
              resolve({ blob: b || file, previewUrl: b ? URL.createObjectURL(b) : (reader.result as string), originalSizeKB, compressedSizeKB })
            },
            'image/jpeg',
            0.80
          )
          return
        }

        // Reverse Geocoding
        currentStepText.value = 'Mengidentifikasi alamat wilayah lokasi...'
        const geoData = await getReverseGeocode(geoCoords.lat, geoCoords.lng)
        const trimmedItemName = (itemNameText || '').trim()
        const titleText = locName ? `${locName}, ${geoData.title}` : geoData.title
        const addressText = geoData.fullAddress

        // Prepare map tile
        currentStepText.value = 'Menyiapkan thumbnail peta koordinat...'
        const tileImg = await loadTileBlobImage(geoCoords.lat, geoCoords.lng)

        // Large, bold, highly legible GPS Map Camera fonts (matches real GPS Map Camera overlay)
        currentStepText.value = 'Mengecap banner GPS Map Camera...'

        // Dynamic ratio-adaptive scaling matching real GPS Map Camera apps
        // Automatically adapts proportions, typography, and card height based on aspect ratio (portrait, landscape, square)
        const aspectRatio = width / height
        const isLandscape = aspectRatio > 1.05
        const isPortrait = aspectRatio < 0.95

        // Base dimension tailored to orientation so watermark never overwhelms the subject photo
        const baseDim = isLandscape ? height : width
        const scale = Math.max(0.9, baseDim / 720)

        // Flush bottom banner edge-to-edge
        const cardX = 0
        const cardW = width

        // Ratio-adapted paddings
        const padX = Math.round(isPortrait ? 18 * scale : 24 * scale)
        const padY = Math.round(16 * scale)

        // Dynamic font sizes strictly proportional to the active ratio base dimension
        const titleSize = Math.max(26, Math.round(baseDim * (isLandscape ? 0.046 : 0.044)))
        const bodySize = Math.max(18, Math.round(baseDim * (isLandscape ? 0.030 : 0.028)))
        const metaSize = Math.max(16, Math.round(baseDim * (isLandscape ? 0.026 : 0.024)))
        const badgeSize = Math.max(15, Math.round(baseDim * (isLandscape ? 0.024 : 0.022)))

        const titleLineH = Math.round(titleSize * 1.25)
        const bodyLineH = Math.round(bodySize * 1.34)
        const metaLineH = Math.round(metaSize * 1.35)

        // Measure Top-Right Badge: [📷 GPS Map Camera]
        ctx.font = `600 ${badgeSize}px system-ui, -apple-system, sans-serif`
        const badgeText = 'GPS Map Camera'
        const badgeTextW = ctx.measureText(badgeText).width
        const badgeIconSize = Math.round(badgeSize * 1.3)
        const badgeTotalW = badgeIconSize + badgeTextW + Math.round(10 * scale)

        // Map thumbnail dimensions dynamically adapted to aspect ratio
        const approxMapSize = Math.round(isLandscape ? height * 0.22 : width * 0.23)

        // Available width for right-side text column
        const textX = padX + approxMapSize + padX
        // Title area width guarantees zero overlap with top-right badge
        const titleAreaW = cardW - textX - badgeTotalW - padX - Math.round(14 * scale)
        // Full text width for address, coords, and timestamp
        const textAreaW = cardW - textX - padX

        let textContentH = 0
        let itemLines: string[] = []
        let subLines: string[] = []
        let titleLines: string[] = []
        let addressLines: string[] = []
        const flagW = Math.round(titleSize * 1.15)
        const flagH = Math.round(titleSize * 0.75)

        if (trimmedItemName) {
          // 1. Item Name (Prominent bold header)
          ctx.font = `700 ${titleSize}px system-ui, -apple-system, sans-serif`
          itemLines = wrapText(ctx, trimmedItemName, titleAreaW).slice(0, 2)

          // 2. Subtitle: Location / District (Cyan) + Indonesian Flag
          const subTitleText = locName ? `${locName} · ${geoData.title}` : geoData.title
          ctx.font = `600 ${bodySize}px system-ui, -apple-system, sans-serif`
          const subAreaW = textAreaW - Math.round(bodySize * 1.2) - Math.round(12 * scale)
          subLines = wrapText(ctx, subTitleText, subAreaW).slice(0, 2)

          // 3. Street Address
          ctx.font = `400 ${metaSize}px system-ui, -apple-system, sans-serif`
          addressLines = wrapText(ctx, addressText, textAreaW).slice(0, 2)

          const itemBlockH = itemLines.length * titleLineH
          const subBlockH = subLines.length * bodyLineH
          const addressBlockH = addressLines.length * metaLineH
          textContentH = itemBlockH + Math.round(4 * scale) + subBlockH + Math.round(4 * scale) + addressBlockH + Math.round(5 * scale) + metaLineH * 2 + Math.round(4 * scale)
        } else {
          // Fallback if no item name is provided
          ctx.font = `700 ${titleSize}px system-ui, -apple-system, sans-serif`
          titleLines = wrapText(ctx, titleText, titleAreaW).slice(0, 2)

          ctx.font = `400 ${bodySize}px system-ui, -apple-system, sans-serif`
          addressLines = wrapText(ctx, addressText, textAreaW).slice(0, 2)

          const titleBlockH = titleLines.length * titleLineH
          const addressBlockH = addressLines.length * bodyLineH
          textContentH = titleBlockH + Math.round(6 * scale) + addressBlockH + Math.round(6 * scale) + metaLineH * 2 + Math.round(6 * scale)
        }

        const mapSize = Math.max(approxMapSize, textContentH)
        const cardH = mapSize + padY * 2
        // Flush at bottom edge
        const cardY = height - cardH

        // Draw Card Background (semi-translucent deep charcoal with subtle top divider)
        ctx.save()
        ctx.fillStyle = 'rgba(15, 18, 22, 0.88)'
        ctx.fillRect(cardX, cardY, cardW, cardH)
        ctx.strokeStyle = 'rgba(255, 255, 255, 0.14)'
        ctx.lineWidth = Math.max(1, Math.round(scale))
        ctx.beginPath()
        ctx.moveTo(cardX, cardY)
        ctx.lineTo(cardX + cardW, cardY)
        ctx.stroke()
        ctx.restore()

        // Draw Top-Right Badge: [📷 GPS Map Camera]
        ctx.save()
        const badgeX = cardW - padX - badgeTotalW
        const badgeY = cardY + padY

        // Cyan camera icon background
        ctx.fillStyle = '#0284c7'
        ctx.fillRect(badgeX, badgeY - 1, badgeIconSize, badgeIconSize)
        ctx.fillStyle = '#38bdf8'
        ctx.beginPath()
        ctx.arc(badgeX + badgeIconSize / 2, badgeY - 1 + badgeIconSize / 2, badgeIconSize * 0.32, 0, Math.PI * 2)
        ctx.fill()
        ctx.fillStyle = '#ffffff'
        ctx.beginPath()
        ctx.arc(badgeX + badgeIconSize / 2, badgeY - 1 + badgeIconSize / 2, badgeIconSize * 0.16, 0, Math.PI * 2)
        ctx.fill()

        // Badge Text
        ctx.font = `600 ${badgeSize}px system-ui, -apple-system, sans-serif`
        ctx.fillStyle = '#f8fafc'
        ctx.textBaseline = 'top'
        ctx.fillText(badgeText, badgeX + badgeIconSize + Math.round(6 * scale), badgeY)
        ctx.restore()

        // Draw Map Box (Left column)
        const mapX = padX
        const mapY = cardY + padY
        const mapRadius = Math.round(10 * scale)

        ctx.save()
        ctx.beginPath()
        ctx.moveTo(mapX + mapRadius, mapY)
        ctx.lineTo(mapX + mapSize - mapRadius, mapY)
        ctx.arcTo(mapX + mapSize, mapY, mapX + mapSize, mapY + mapRadius, mapRadius)
        ctx.lineTo(mapX + mapSize, mapY + mapSize - mapRadius)
        ctx.arcTo(mapX + mapSize, mapY + mapSize, mapX + mapSize - mapRadius, mapY + mapSize, mapRadius)
        ctx.lineTo(mapX + mapRadius, mapY + mapSize)
        ctx.arcTo(mapX, mapY + mapSize, mapX, mapY + mapSize - mapRadius, mapRadius)
        ctx.lineTo(mapX, mapY + mapRadius)
        ctx.arcTo(mapX, mapY, mapX + mapRadius, mapY, mapRadius)
        ctx.closePath()
        ctx.clip()

        if (tileImg) {
          ctx.drawImage(tileImg, mapX, mapY, mapSize, mapSize)
        } else {
          drawProceduralMap(ctx, mapX, mapY, mapSize, mapSize)
        }

        // Draw Red Google Map Pin in center
        drawPin(ctx, mapX + mapSize / 2, mapY + mapSize / 2 + Math.round(mapSize * 0.05), Math.round(mapSize * 0.22))

        // Google watermark at bottom of map
        ctx.font = `bold ${Math.max(13, Math.round(15 * scale))}px system-ui, sans-serif`
        ctx.fillStyle = 'rgba(0, 0, 0, 0.7)'
        ctx.fillText('Google', mapX + 8, mapY + mapSize - 6)
        ctx.fillStyle = '#ffffff'
        ctx.fillText('Google', mapX + 7, mapY + mapSize - 7)
        ctx.restore()

        // Draw Text Block (Right column, vertically centered with map)
        const textYOffset = Math.max(0, Math.floor((mapSize - textContentH) / 2))
        let textY = cardY + padY + textYOffset

        ctx.save()
        ctx.shadowColor = 'rgba(0, 0, 0, 0.95)'
        ctx.shadowBlur = Math.round(4 * scale)
        ctx.shadowOffsetX = 0
        ctx.shadowOffsetY = Math.round(2 * scale)
        ctx.textBaseline = 'top'

        if (trimmedItemName) {
          // 1. Item Name Header (Bold White)
          ctx.font = `700 ${titleSize}px system-ui, -apple-system, sans-serif`
          ctx.fillStyle = '#ffffff'
          for (const line of itemLines) {
            ctx.fillText(line, textX, textY)
            textY += titleLineH
          }
          textY += Math.round(4 * scale)

          // 2. Subtitle: Location / District (Cyan) + Indonesian Flag
          ctx.font = `600 ${bodySize}px system-ui, -apple-system, sans-serif`
          ctx.fillStyle = '#38bdf8'
          for (let i = 0; i < subLines.length; i++) {
            const line = subLines[i]
            ctx.fillText(line, textX, textY)
            if (i === subLines.length - 1) {
              const lw = ctx.measureText(line).width
              const subFlagW = Math.round(bodySize * 1.15)
              const subFlagH = Math.round(bodySize * 0.75)
              drawIndonesianFlag(ctx, textX + lw + Math.round(10 * scale), textY + Math.round((bodySize - subFlagH) / 2), subFlagW, subFlagH)
            }
            textY += bodyLineH
          }
          textY += Math.round(4 * scale)

          // 3. Street Address
          ctx.font = `400 ${metaSize}px system-ui, -apple-system, sans-serif`
          ctx.fillStyle = '#f1f5f9'
          for (const line of addressLines) {
            ctx.fillText(line, textX, textY)
            textY += metaLineH
          }
          textY += Math.round(5 * scale)
        } else {
          // 1. Title Header + Indonesian Flag
          ctx.font = `700 ${titleSize}px system-ui, -apple-system, sans-serif`
          ctx.fillStyle = '#ffffff'
          for (let i = 0; i < titleLines.length; i++) {
            const line = titleLines[i]
            ctx.fillText(line, textX, textY)
            if (i === titleLines.length - 1) {
              const lw = ctx.measureText(line).width
              drawIndonesianFlag(ctx, textX + lw + Math.round(12 * scale), textY + Math.round((titleSize - flagH) / 2), flagW, flagH)
            }
            textY += titleLineH
          }
          textY += Math.round(6 * scale)

          // 2. Full Detailed Address (wrapped 1-2 lines)
          ctx.font = `400 ${bodySize}px system-ui, -apple-system, sans-serif`
          ctx.fillStyle = '#f1f5f9'
          for (const line of addressLines) {
            ctx.fillText(line, textX, textY)
            textY += bodyLineH
          }
          textY += Math.round(6 * scale)
        }

        // Monospace Coordinates
        ctx.font = `600 ${metaSize}px ui-monospace, SFMono-Regular, system-ui, monospace`
        ctx.fillStyle = '#ffffff'
        ctx.fillText(`Lat ${geoCoords.lat.toFixed(6)}°  Long ${geoCoords.lng.toFixed(6)}°`, textX, textY)
        textY += metaLineH + Math.round(5 * scale)

        // Timestamp (Single, crisp, clean)
        const dateStr = formatGpsDate(new Date())
        ctx.font = `400 ${metaSize}px system-ui, -apple-system, sans-serif`
        ctx.fillStyle = '#cbd5e1'
        ctx.fillText(dateStr, textX, textY)
        ctx.restore()

        // Output Blob with compression (quality: 0.80)
        currentStepText.value = 'Mengompresi ke JPEG (q=0.80)...'
        canvas.toBlob(
          (blob) => {
            if (blob) {
              const compressedSizeKB = Math.round(blob.size / 1024)
              const previewUrl = URL.createObjectURL(blob)
              resolve({ blob, previewUrl, originalSizeKB, compressedSizeKB })
            } else {
              resolve({ blob: file, previewUrl: reader.result as string, originalSizeKB, compressedSizeKB: originalSizeKB })
            }
          },
          'image/jpeg',
          0.80
        )
      }
      img.onerror = () => resolve({ blob: file, previewUrl: reader.result as string, originalSizeKB, compressedSizeKB: originalSizeKB })
      img.src = reader.result as string
    }
    reader.onerror = () => resolve({ blob: file, previewUrl: '', originalSizeKB: 0, compressedSizeKB: 0 })
    reader.readAsDataURL(file)
  })
}

async function onFileSelected(e: Event) {
  const target = e.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  lastRawFile.value = file
  uploading.value = true
  isUploadingModal.value = true
  uploadProgress.value = 15
  uploadStep.value = 'gps'
  currentStepText.value = 'Mengunci koordinat GPS presisi tinggi...'

  try {
    // 1. Lock GPS
    if (!coords.value && navigator.geolocation) {
      await fetchGeo()
    }
    uploadProgress.value = 35

    const locName = geoLabel.value || props.locationName
    if (locName) {
      emit('update:geoName', locName)
    }

    // 2. Stamp photo with compact GPS Map Camera template & compression
    uploadStep.value = 'stamp'
    currentStepText.value = 'Mengecap banner GPS Map Camera & peta...'
    uploadProgress.value = 50

    const stamped = await stampGpsMapCamera(file, coords.value, locName, props.itemName)
    if (localPreviewUrl.value) URL.revokeObjectURL(localPreviewUrl.value)
    localPreviewUrl.value = stamped.previewUrl

    uploadStep.value = 'upload'
    currentStepText.value = 'Mengunggah foto berstempel ke server...'
    uploadProgress.value = 70

    // 3. Upload to server
    const fd = new FormData()
    fd.append('photo', stamped.blob, 'photo.jpg')
    fd.append('client_stamped', 'true')
    if (coords.value) {
      fd.append('geo_lat', coords.value.lat.toString())
      fd.append('geo_lng', coords.value.lng.toString())
      fd.append('geo_acc', coords.value.acc.toString())
    }
    if (locName) {
      fd.append('geo_name', locName)
    }

    uploadAbortCtrl = new AbortController()
    const res = await api.post('/upload', fd, {
      signal: uploadAbortCtrl.signal,
      onUploadProgress: (evt) => {
        if (evt.total) {
          const pct = Math.round(70 + (evt.loaded / evt.total) * 28)
          uploadProgress.value = Math.min(98, pct)
        }
      },
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

    uploadProgress.value = 100
    uploadStep.value = 'success'
    currentStepText.value = 'Foto berhasil diproses & disimpan!'

    toast.add({
      severity: 'success',
      summary: 'Foto Geotag Berhasil',
      detail: coords.value
        ? `Watermark GPS (${coords.value.lat.toFixed(4)}, ${coords.value.lng.toFixed(4)}) tercetak di dalam foto`
        : 'Foto berhasil disimpan',
      life: 3000,
    })
  } catch (err: any) {
    if (err.name === 'CanceledError' || err.name === 'AbortError') return
    const errorMsg = err.response?.data?.error || err.message || 'Gagal mengunggah foto'
    currentStepText.value = 'Gagal memproses foto'
    isUploadingModal.value = false
    toast.add({
      severity: 'error',
      summary: 'Gagal mengunggah foto',
      detail: errorMsg,
      life: 4500,
    })
  } finally {
    uploading.value = false
    uploadAbortCtrl = null
    if (target) target.value = ''
  }
}

async function reStampPhoto() {
  if (!lastRawFile.value || uploading.value) return
  uploading.value = true
  isUploadingModal.value = true
  uploadProgress.value = 25
  uploadStep.value = 'stamp'
  currentStepText.value = 'Mencetak ulang cap GPS Map Camera...'

  try {
    const locName = geoLabel.value || props.locationName
    const stamped = await stampGpsMapCamera(lastRawFile.value, coords.value, locName, props.itemName)
    if (localPreviewUrl.value) URL.revokeObjectURL(localPreviewUrl.value)
    localPreviewUrl.value = stamped.previewUrl

    uploadStep.value = 'upload'
    uploadProgress.value = 70
    currentStepText.value = 'Mengunggah foto berstempel terbaru...'

    const fd = new FormData()
    fd.append('photo', stamped.blob, 'photo.jpg')
    fd.append('client_stamped', 'true')
    if (coords.value) {
      fd.append('geo_lat', coords.value.lat.toString())
      fd.append('geo_lng', coords.value.lng.toString())
      fd.append('geo_acc', coords.value.acc.toString())
    }
    if (locName) {
      fd.append('geo_name', locName)
    }

    uploadAbortCtrl = new AbortController()
    const res = await api.post('/upload', fd, { signal: uploadAbortCtrl.signal })
    const url = res.data.url
    emit('update:modelValue', url)

    uploadProgress.value = 100
    uploadStep.value = 'success'
    currentStepText.value = 'Cap watermark berhasil diperbarui!'
    toast.add({
      severity: 'success',
      summary: 'Cap Berhasil Diperbarui',
      detail: props.itemName ? `Nama "${props.itemName}" tercetak pada foto` : 'Cap geotag diperbarui',
      life: 2500,
    })
  } catch (err: any) {
    if (err.name === 'CanceledError' || err.name === 'AbortError') return
    const errorMsg = err.response?.data?.error || err.message || 'Gagal memperbarui cap'
    isUploadingModal.value = false
    toast.add({
      severity: 'error',
      summary: 'Gagal memperbarui cap',
      detail: errorMsg,
      life: 3500,
    })
  } finally {
    uploading.value = false
    uploadAbortCtrl = null
  }
}

let debounceTimer: any = null
watch(
  () => [props.itemName, props.locationName],
  () => {
    if (lastRawFile.value && !uploading.value && currentPhotoSrc.value) {
      if (debounceTimer) clearTimeout(debounceTimer)
      debounceTimer = setTimeout(() => {
        reStampPhoto()
      }, 1400)
    }
  }
)

function removePhoto() {
  if (localPreviewUrl.value) {
    URL.revokeObjectURL(localPreviewUrl.value)
    localPreviewUrl.value = ''
  }
  lastRawFile.value = null
  emit('update:modelValue', '')
}

async function downloadPhoto() {
  if (!currentPhotoSrc.value) return
  try {
    const res = await fetch(currentPhotoSrc.value)
    const blob = await res.blob()
    const blobUrl = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = blobUrl
    a.download = 'foto_geotag.jpg'
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(blobUrl)
  } catch {
    const a = document.createElement('a')
    a.href = currentPhotoSrc.value
    a.download = 'foto_geotag.jpg'
    a.target = '_blank'
    a.click()
  }
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
    <div v-if="!currentPhotoSrc" class="flex items-center gap-2">
      <InputText
        v-model="geoLabel"
        size="small"
        placeholder="Nama ruangan / keterangan lokasi (opsional)"
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
          <i class="pi pi-check-circle" /> Foto Berstempel GPS Map Camera
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
            label="Batal / Hapus"
            icon="pi pi-trash"
            size="small"
            text
            severity="danger"
            v-tooltip.top="'Batalkan dan hapus foto'"
            @click="cancelUpload"
          />
        </div>
      </div>

      <!-- Stamped Item Name & Location Banner indicator -->
      <div v-if="props.itemName || props.locationName" class="flex items-center justify-between gap-2 px-2.5 py-1.5 rounded text-[11.5px] border" style="background: var(--panel-2); border-color: var(--line)">
        <div class="flex items-center gap-1.5 overflow-hidden">
          <i class="pi pi-tag text-indigo-500 text-[10px] shrink-0" />
          <span style="color: var(--txt-dim)">Nama di Cap:</span>
          <span class="font-semibold truncate" style="color: var(--txt)">{{ props.itemName || '(Belum ada nama)' }}</span>
          <span v-if="props.locationName" class="text-indigo-400 truncate">· {{ props.locationName }}</span>
        </div>
        <Button
          v-if="lastRawFile"
          label="Perbarui Cap"
          icon="pi pi-sync"
          size="small"
          text
          class="!py-0 !px-1.5 !text-[10px] shrink-0"
          :loading="uploading"
          v-tooltip.top="'Cetak ulang cap dengan nama barang / lokasi terbaru'"
          @click="reStampPhoto"
        />
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
        <span>Stempel peta, alamat lengkap, dan koordinat GPS tercetak langsung di dalam foto</span>
        <span v-if="coords" class="t-mono text-[10.5px] text-sig-ok flex items-center gap-1">
          <i class="pi pi-map-marker text-[9px]" /> GPS Aktif
        </span>
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
          {{ uploading ? 'Memproses cap GPS Map Camera…' : 'Ambil foto berkamera atau unggah berkas' }}
        </span>
        <span class="text-[11px]" style="color: var(--txt-dim)">
          Format JPG / PNG, cap peta dan alamat otomatis dicetak di dalam foto
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

    <!-- Futuristic Upload & Stamping Radar Animation Modal -->
    <Teleport to="body">
      <div
        v-if="isUploadingModal"
        class="fixed inset-0 z-[9999] bg-black/85 backdrop-blur-md flex items-center justify-center p-4 transition-all duration-300"
      >
        <div
          class="relative max-w-sm w-full bg-slate-900 border border-slate-700/80 rounded-2xl p-6 shadow-2xl flex flex-col items-center text-center overflow-hidden"
          style="background: linear-gradient(180deg, #0f172a 0%, #090d16 100%)"
        >
          <!-- Background Ambient Glow -->
          <div
            class="absolute -top-20 -left-20 w-44 h-44 rounded-full bg-cyan-500/20 blur-3xl pointer-events-none"
          />
          <div
            class="absolute -bottom-20 -right-20 w-44 h-44 rounded-full bg-indigo-500/20 blur-3xl pointer-events-none"
          />

          <!-- Cool Radar / Satellite Orbit Visual -->
          <div class="relative w-28 h-28 my-3 flex items-center justify-center">
            <!-- Pulsing outer wave -->
            <div
              class="absolute inset-0 rounded-full border border-cyan-500/30 animate-ping opacity-75"
              style="animation-duration: 2.4s"
            />
            <!-- Outer rotating dashed orbit -->
            <div
              class="absolute inset-1 rounded-full border border-dashed border-cyan-400/50 animate-spin"
              style="animation-duration: 9s"
            />
            <!-- Middle counter-rotating dashed orbit -->
            <div
              class="absolute inset-3 rounded-full border border-dashed border-indigo-400/60 animate-spin"
              style="animation-duration: 6s; animation-direction: reverse"
            />
            <!-- Inner glowing core circle with step icon -->
            <div
              class="relative w-16 h-16 rounded-full flex items-center justify-center shadow-lg transition-all duration-500"
              :class="{
                'bg-cyan-500/20 border border-cyan-400 text-cyan-300 shadow-cyan-500/40': uploadStep === 'gps',
                'bg-indigo-500/20 border border-indigo-400 text-indigo-300 shadow-indigo-500/40': uploadStep === 'stamp',
                'bg-purple-500/20 border border-purple-400 text-purple-300 shadow-purple-500/40': uploadStep === 'compress',
                'bg-sky-500/20 border border-sky-400 text-sky-300 shadow-sky-500/40': uploadStep === 'upload',
                'bg-emerald-500/25 border border-emerald-400 text-emerald-300 shadow-emerald-500/50': uploadStep === 'success',
              }"
            >
              <i
                v-if="uploadStep === 'gps'"
                class="pi pi-compass text-2xl animate-pulse"
              />
              <i
                v-else-if="uploadStep === 'stamp'"
                class="pi pi-camera text-2xl animate-pulse"
              />
              <i
                v-else-if="uploadStep === 'compress'"
                class="pi pi-sliders-h text-2xl animate-pulse"
              />
              <i
                v-else-if="uploadStep === 'upload'"
                class="pi pi-cloud-upload text-2xl animate-bounce"
              />
              <i
                v-else
                class="pi pi-check text-2xl scale-110 font-bold"
              />
            </div>
          </div>

          <!-- Title & Step Text -->
          <div class="flex flex-col gap-1 z-10 w-full mt-1">
            <div class="text-[15px] font-bold text-white flex items-center justify-center gap-1.5">
              <span>{{ uploadStep === 'success' ? 'Foto Siap Digunakan' : 'Memproses Geotagging' }}</span>
            </div>

            <!-- Item Name Pill -->
            <div v-if="props.itemName" class="inline-flex items-center justify-center gap-1.5 px-2.5 py-0.5 rounded-full bg-slate-800/80 border border-slate-700 text-slate-300 text-[11px] font-medium mx-auto max-w-[260px] truncate">
              <i class="pi pi-tag text-[9px] text-indigo-400 shrink-0" />
              <span class="truncate">{{ props.itemName }}</span>
            </div>

            <p class="text-[12px] text-slate-400 mt-1 min-h-[32px] flex items-center justify-center gap-1.5 leading-snug">
              <span v-if="uploadStep !== 'success'" class="w-1.5 h-1.5 rounded-full bg-cyan-400 animate-ping inline-block" />
              <span>{{ currentStepText }}</span>
            </p>
          </div>

          <!-- Gradient Progress Bar -->
          <div class="w-full flex flex-col gap-1.5 mt-2 z-10">
            <div class="flex justify-between items-center text-[11px]">
              <span class="text-slate-400 font-mono text-[10px] uppercase tracking-wider">
                {{ uploadStep === 'success' ? 'Selesai 100%' : 'Sedang Diproses' }}
              </span>
              <span class="font-mono font-bold text-cyan-300">{{ uploadProgress }}%</span>
            </div>
            <div class="w-full h-2 rounded-full bg-slate-800 border border-slate-700/60 overflow-hidden p-0.5">
              <div
                class="h-full rounded-full bg-gradient-to-r from-cyan-500 via-indigo-500 to-emerald-400 transition-all duration-300 shadow-sm"
                :style="{ width: `${uploadProgress}%` }"
              />
            </div>
          </div>

          <!-- Actions -->
          <div class="w-full flex flex-col gap-2 mt-5 z-10">
            <!-- If Upload Finished: Options to Cancel/Undo, View, or Close -->
            <div v-if="uploadStep === 'success'" class="flex items-center gap-2 w-full">
              <Button
                label="Batal / Hapus"
                icon="pi pi-trash"
                severity="danger"
                outlined
                size="small"
                class="flex-1 !text-[12px]"
                @click="cancelUpload"
              />
              <Button
                label="Lihat Foto"
                icon="pi pi-eye"
                severity="primary"
                size="small"
                class="flex-1 !text-[12px]"
                @click="isUploadingModal = false; modalTab = 'photo'; previewOpen = true"
              />
              <Button
                icon="pi pi-check"
                severity="success"
                size="small"
                v-tooltip.top="'Selesai'"
                @click="isUploadingModal = false"
              />
            </div>

            <!-- While in progress: Abort / Cancel Button -->
            <div v-else class="flex justify-center">
              <Button
                label="Batal Unggah"
                icon="pi pi-times"
                severity="danger"
                text
                size="small"
                class="!text-[11.5px]"
                @click="cancelUpload"
              />
            </div>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Preview Modal with Tab (Foto Berstempel vs Peta Geotag) Teleported to Body -->
    <Teleport to="body">
      <div
        v-if="previewOpen"
        class="fixed inset-0 z-[9999] bg-black/85 flex items-center justify-center p-4 backdrop-blur-xs"
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
                class="px-3 py-1 rounded text-[12px] font-bold flex items-center gap-1.5 transition-colors cursor-pointer"
                :class="modalTab === 'photo' ? 'bg-acc-500 text-ink-950' : 'text-ink-300 hover:bg-paper-2'"
                @click="modalTab = 'photo'"
              >
                <i class="pi pi-image text-[11px]" /> Foto Berstempel GPS Map Camera
              </button>
              <button
                v-if="coords"
                class="px-3 py-1 rounded text-[12px] font-bold flex items-center gap-1.5 transition-colors cursor-pointer"
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
              <Button
                label="Batal / Hapus Foto"
                icon="pi pi-trash"
                size="small"
                severity="danger"
                text
                @click="cancelUpload"
              />
              <a
                v-if="coords"
                :href="googleMapsUrl"
                target="_blank"
                rel="noopener noreferrer"
                class="inline-flex"
              >
                <Button label="Google Maps" icon="pi pi-map-marker" size="small" text />
              </a>
              <Button
                label="Unduh Foto (JPG)"
                icon="pi pi-download"
                size="small"
                severity="success"
                text
                @click="downloadPhoto"
              />
              <Button label="Tutup" size="small" severity="secondary" @click="previewOpen = false" />
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
