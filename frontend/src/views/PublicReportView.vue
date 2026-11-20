<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '@/api'
import { useThemeStore } from '@/stores/theme'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Textarea from 'primevue/textarea'
import Select from 'primevue/select'
import Dialog from 'primevue/dialog'
import AppLogo from '@/components/AppLogo.vue'
import StatusChip from '@/components/StatusChip.vue'

interface RecentTicket {
  number: string
  type: string
  location: string
  createdAt: string
}

const RECENT_TICKETS_KEY = 'lapor_recent_tickets_v1'

const route = useRoute()
const router = useRouter()
const theme = useThemeStore()

const activeTab = ref<'create' | 'track'>('create')

const reporterName = ref('')
const reporterContact = ref('')
const reportType = ref('Kerusakan Alat / Barang')
const description = ref('')
const location = ref('')
const locations = ref<{ id: string; name: string; code: string }[]>([])
const photoUrl = ref('')
const uploadingPhoto = ref(false)
const photoPreview = ref('')
const photoError = ref('')
const dragOver = ref(false)
const fileInputRef = ref<HTMLInputElement | null>(null)

const reportTypeOptions = [
  {
    value: 'Kerusakan Alat / Barang',
    title: 'Kerusakan Alat',
    desc: 'Mati, eror, pecah, atau tidak layak pakai',
    icon: 'pi pi-wrench',
  },
  {
    value: 'Kehilangan / Tidak Ditemukan',
    title: 'Kehilangan',
    desc: 'Barang tidak berada di tempat semestinya',
    icon: 'pi pi-search',
  },
  {
    value: 'Perpindahan Ruangan / Lokasi',
    title: 'Pindah Ruangan',
    desc: 'Mutasi atau peminjaman antar ruangan',
    icon: 'pi pi-arrows-h',
  },
  {
    value: 'Kondisi Fisik & Kebersihan',
    title: 'Fisik & Kebersihan',
    desc: 'Kotor, berkarat, atau perlu perawatan rutin',
    icon: 'pi pi-sparkles',
  },
  {
    value: 'Lainnya',
    title: 'Kendala Lainnya',
    desc: 'Keluhan atau permintaan fasilitas lain',
    icon: 'pi pi-ellipsis-h',
  },
]

const item = ref<any>(null)
const loadingItem = ref(false)

const trackQuery = ref('')
const trackResult = ref<any>(null)
const trackLoading = ref(false)
const trackError = ref('')
const recentTickets = ref<RecentTicket[]>([])

const submitting = ref(false)
const submitSuccess = ref(false)
const createdTicket = ref<any>(null)
const copiedTicket = ref(false)
const copiedTracked = ref(false)
const errorMessage = ref('')
const submitted = ref(false)

const lightboxVisible = ref(false)
const lightboxUrl = ref('')

function loadRecentTickets() {
  try {
    const raw = localStorage.getItem(RECENT_TICKETS_KEY)
    if (!raw) return
    const parsed = JSON.parse(raw)
    if (Array.isArray(parsed)) {
      recentTickets.value = parsed.slice(0, 5)
    }
  } catch {
    recentTickets.value = []
  }
}

function saveRecentTicket(ticket: RecentTicket) {
  const next = [ticket, ...recentTickets.value.filter((t) => t.number !== ticket.number)].slice(0, 5)
  recentTickets.value = next
  try {
    localStorage.setItem(RECENT_TICKETS_KEY, JSON.stringify(next))
  } catch {
    // ignore storage errors in private browsing
  }
}

function revokePreview() {
  if (photoPreview.value && photoPreview.value.startsWith('blob:')) {
    URL.revokeObjectURL(photoPreview.value)
  }
}

onBeforeUnmount(() => {
  revokePreview()
})

onMounted(async () => {
  theme.apply()
  loadRecentTickets()

  try {
    const locRes = await api.get('/public/locations')
    locations.value = locRes.data || []
  } catch {
    // fallback: tetap bisa isi manual
  }

  const targetId = route.params.id || route.query.id || route.query.sku
  if (targetId) {
    loadingItem.value = true
    try {
      const res = await api.get(`/scan/${encodeURIComponent(String(targetId))}`)
      item.value = res.data
      if (item.value?.scanned_unit?.location) {
        location.value = item.value.scanned_unit.location
      } else if (item.value?.location) {
        location.value = item.value.location
      }
    } catch {
      // not found is ok, user can still submit report
    } finally {
      loadingItem.value = false
    }
  }

  if (route.query.track) {
    activeTab.value = 'track'
    trackQuery.value = String(route.query.track)
    fetchTracking()
  }
})

async function handlePhotoFile(file: File) {
  photoError.value = ''
  if (!file.type.startsWith('image/')) {
    photoError.value = 'File harus berupa gambar (JPG, PNG, atau WebP).'
    return
  }
  if (file.size > 10 * 1024 * 1024) {
    photoError.value = 'Ukuran foto terlalu besar. Maksimal 10 MB.'
    return
  }

  revokePreview()
  photoPreview.value = URL.createObjectURL(file)

  const formData = new FormData()
  formData.append('photo', file)

  uploadingPhoto.value = true
  try {
    const res = await api.post('/public/upload', formData)
    photoUrl.value = res.data.url
  } catch (err: any) {
    photoError.value =
      err?.response?.data?.error || 'Gagal mengunggah foto. Pastikan format gambar JPG, PNG, atau WebP.'
    revokePreview()
    photoPreview.value = ''
    photoUrl.value = ''
  } finally {
    uploadingPhoto.value = false
  }
}

function onPhotoSelected(e: Event) {
  const input = e.target as HTMLInputElement
  if (!input.files || input.files.length === 0) return
  handlePhotoFile(input.files[0])
  input.value = ''
}

function onPhotoDrop(e: DragEvent) {
  dragOver.value = false
  const file = e.dataTransfer?.files?.[0]
  if (file) handlePhotoFile(file)
}

function clearPhoto() {
  revokePreview()
  photoPreview.value = ''
  photoUrl.value = ''
  photoError.value = ''
}

async function submitReport() {
  errorMessage.value = ''
  submitted.value = true
  if (!reporterName.value.trim() || !description.value.trim()) {
    errorMessage.value = 'Mohon lengkapi Nama Lengkap Pelapor dan Uraian Masalah.'
    return
  }
  if (!location.value.trim()) {
    errorMessage.value = 'Mohon pilih atau ketik Lokasi / Ruangan barang saat ini.'
    return
  }

  submitting.value = true
  try {
    const scannedUnit = item.value?.scanned_unit
    const unitNote = scannedUnit
      ? `[Unit #${scannedUnit.unit_number}${scannedUnit.register_number ? ` · Reg ${scannedUnit.register_number}` : ''}] `
      : ''

    const payload = {
      reporter_name: reporterName.value.trim(),
      reporter_contact: reporterContact.value.trim() || '-',
      report_type: reportType.value,
      description: `${unitNote}${description.value.trim()}`,
      location: location.value.trim(),
      photo_url: photoUrl.value,
      item_id: item.value?.id || undefined,
      item_sku: item.value?.sku || (route.query.sku ? String(route.query.sku) : undefined),
      item_name: item.value?.name || undefined,
    }

    const res = await api.post('/public/reports', payload)
    createdTicket.value = res.data
    copiedTicket.value = false
    submitSuccess.value = true

    if (res.data?.report_number) {
      saveRecentTicket({
        number: res.data.report_number,
        type: reportType.value,
        location: location.value.trim(),
        createdAt: new Date().toISOString(),
      })
    }

    reporterName.value = ''
    reporterContact.value = ''
    description.value = ''
    clearPhoto()
    submitted.value = false
  } catch (err: any) {
    errorMessage.value = err?.response?.data?.error || 'Gagal mengirim laporan. Silakan coba kembali.'
  } finally {
    submitting.value = false
  }
}

async function fetchTracking(ticketNum?: string) {
  if (ticketNum) trackQuery.value = ticketNum
  const q = trackQuery.value.trim()
  if (!q) return
  trackLoading.value = true
  trackError.value = ''
  trackResult.value = null
  try {
    const res = await api.get(`/public/reports/${encodeURIComponent(q)}`)
    trackResult.value = res.data
  } catch {
    trackError.value = 'Nomor tiket tidak ditemukan. Periksa kembali format LAP-YYYYMMDD-XXXXXX Anda.'
  } finally {
    trackLoading.value = false
  }
}

async function copyText(val: string, target: 'created' | 'tracked') {
  if (!val) return
  try {
    await navigator.clipboard.writeText(val)
    if (target === 'created') {
      copiedTicket.value = true
      setTimeout(() => (copiedTicket.value = false), 2200)
    } else {
      copiedTracked.value = true
      setTimeout(() => (copiedTracked.value = false), 2200)
    }
  } catch {
    // ignore clipboard failures
  }
}

function trackCreatedTicket() {
  if (!createdTicket.value?.report_number) return
  submitSuccess.value = false
  activeTab.value = 'track'
  fetchTracking(createdTicket.value.report_number)
}

function openLightbox(url: string) {
  lightboxUrl.value = url
  lightboxVisible.value = true
}

function formatDate(d: string) {
  if (!d) return '-'
  return new Date(d).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

const trackStepIndex = computed(() => {
  const st = trackResult.value?.status
  if (st === 'PENDING') return 0
  if (st === 'IN_PROGRESS') return 1
  if (st === 'RESOLVED' || st === 'REJECTED') return 2
  return 0
})

const trackSteps = computed(() => {
  const isRejected = trackResult.value?.status === 'REJECTED'
  return [
    { key: 'PENDING', label: 'Laporan Diterima', sub: 'Tercatat di sistem', icon: 'pi pi-inbox' },
    { key: 'IN_PROGRESS', label: 'Diproses Petugas', sub: 'Inspeksi & tindak lanjut', icon: 'pi pi-wrench' },
    {
      key: isRejected ? 'REJECTED' : 'RESOLVED',
      label: isRejected ? 'Laporan Ditolak' : 'Selesai Ditangani',
      sub: isRejected ? 'Lihat catatan petugas' : 'Kendala diselesaikan',
      icon: isRejected ? 'pi pi-times-circle' : 'pi pi-check-circle',
    },
  ]
})
</script>

<template>
  <div class="min-h-dvh shell flex flex-col">
    <!-- ============ Sticky Glass Header ============ -->
    <header class="sticky top-0 z-20 border-b border-line bg-[var(--topbar-bg)] backdrop-blur-md">
      <div class="max-w-5xl mx-auto px-4 sm:px-6 h-14 flex items-center justify-between gap-3">
        <button
          type="button"
          class="flex items-center gap-2.5 cursor-pointer text-left min-w-0"
          @click="activeTab = 'create'"
        >
          <AppLogo size="sm" />
        </button>

        <!-- Segmented Switcher (Desktop) -->
        <nav
          aria-label="Navigasi portal laporan"
          class="hidden sm:flex items-center p-1 rounded-[9px] border border-line bg-panel-2"
        >
          <button
            type="button"
            class="px-3 py-1 rounded-[6px] text-[12px] font-semibold transition-all cursor-pointer flex items-center gap-1.5"
            :class="
              activeTab === 'create'
                ? 'bg-panel text-txt shadow-sm border border-line-soft'
                : 'text-txt-dim hover:text-txt border border-transparent'
            "
            @click="activeTab = 'create'"
          >
            <i class="pi pi-file-edit text-[11px]" :class="activeTab === 'create' ? 'text-acc-500' : ''" aria-hidden="true" />
            <span>Buat Laporan</span>
          </button>
          <button
            type="button"
            class="px-3 py-1 rounded-[6px] text-[12px] font-semibold transition-all cursor-pointer flex items-center gap-1.5"
            :class="
              activeTab === 'track'
                ? 'bg-panel text-txt shadow-sm border border-line-soft'
                : 'text-txt-dim hover:text-txt border border-transparent'
            "
            @click="activeTab = 'track'"
          >
            <i class="pi pi-search text-[11px]" :class="activeTab === 'track' ? 'text-acc-500' : ''" aria-hidden="true" />
            <span>Lacak Tiket</span>
            <span
              v-if="recentTickets.length"
              class="t-mono !text-[10px] px-1.5 py-0.2 rounded-full bg-acc-50 text-acc-600 dark:text-acc-300"
            >
              {{ recentTickets.length }}
            </span>
          </button>
        </nav>

        <div class="flex items-center gap-1.5 shrink-0">
          <Button
            :icon="theme.isDark ? 'pi pi-sun' : 'pi pi-moon'"
            text
            rounded
            size="small"
            severity="secondary"
            :aria-label="theme.isDark ? 'Ganti ke mode terang' : 'Ganti ke mode gelap'"
            @click="theme.toggle()"
          />
          <Button
            label="Login Petugas"
            icon="pi pi-sign-in"
            size="small"
            outlined
            severity="secondary"
            @click="router.push('/login')"
          />
        </div>
      </div>
    </header>

    <!-- ============ Architectural Hero Band ============ -->
    <section class="relative border-b border-line bg-[var(--sidebar-bg)] overflow-hidden">
      <div class="grid-bg absolute inset-0 opacity-50 pointer-events-none" aria-hidden="true" />
      <div
        class="absolute -top-24 right-0 w-[380px] h-[380px] rounded-full pointer-events-none"
        style="background: radial-gradient(circle, var(--acc-wash) 0%, transparent 70%);"
        aria-hidden="true"
      />
      <div class="relative max-w-5xl mx-auto px-4 sm:px-6 py-6 sm:py-8">
        <div class="flex flex-col md:flex-row md:items-end justify-between gap-4">
          <div class="max-w-2xl">
            <div class="inline-flex items-center gap-2 t-mono !text-[10.5px] uppercase tracking-[0.08em] text-acc-600 dark:text-acc-300 px-2.5 py-1 rounded-full border border-[var(--acc-border)] bg-acc-50 mb-2.5">
              <span class="w-1.5 h-1.5 rounded-full bg-acc-500 animate-pulse" aria-hidden="true" />
              <span>Portal Pengaduan Aset &amp; Sarpras</span>
            </div>
            <h1 class="text-[24px] sm:text-[30px] leading-[1.12] font-extrabold tracking-[-0.03em] text-txt">
              {{ activeTab === 'create' ? 'Laporkan kendala inventaris dengan cepat.' : 'Pantau progres tindak lanjut laporan Anda.' }}
            </h1>
            <p class="text-[13px] sm:text-[13.5px] text-txt-dim mt-2 max-w-[60ch] leading-relaxed">
              {{
                activeTab === 'create'
                  ? 'Sampaikan kerusakan alat, kehilangan, atau perpindahan ruangan tanpa perlu login. Setiap laporan langsung mendapat nomor tiket pelacakan.'
                  : 'Masukkan nomor tiket untuk melihat status penanganan terkini beserta catatan resmi dari petugas sarpras.'
              }}
            </p>
          </div>

          <!-- Compact Metric Pill -->
          <div class="flex items-center gap-2.5 shrink-0">
            <div class="panel px-3.5 py-2.5 flex items-center gap-2.5">
              <div class="w-8 h-8 rounded-[8px] bg-acc-50 border border-[var(--acc-border)] grid place-items-center text-acc-500">
                <i class="pi pi-bolt text-[13px]" aria-hidden="true" />
              </div>
              <div class="leading-tight">
                <div class="text-[11px] font-semibold text-txt">Tanpa Login</div>
                <div class="text-[10.5px] text-txt-dim">Tiket instan &amp; transparan</div>
              </div>
            </div>
          </div>
        </div>

        <!-- Mobile Segmented Switcher -->
        <div class="grid grid-cols-2 gap-1.5 p-1 mt-5 rounded-[10px] border border-line bg-panel sm:hidden">
          <button
            type="button"
            class="py-2 rounded-[7px] text-[12.5px] font-semibold transition-all cursor-pointer flex items-center justify-center gap-1.5"
            :class="activeTab === 'create' ? 'bg-acc-500 text-white dark:text-[#042f2e]' : 'text-txt-dim'"
            @click="activeTab = 'create'"
          >
            <i class="pi pi-file-edit text-[12px]" aria-hidden="true" />
            <span>Buat Laporan</span>
          </button>
          <button
            type="button"
            class="py-2 rounded-[7px] text-[12.5px] font-semibold transition-all cursor-pointer flex items-center justify-center gap-1.5"
            :class="activeTab === 'track' ? 'bg-acc-500 text-white dark:text-[#042f2e]' : 'text-txt-dim'"
            @click="activeTab = 'track'"
          >
            <i class="pi pi-search text-[12px]" aria-hidden="true" />
            <span>Lacak Tiket</span>
          </button>
        </div>
      </div>
    </section>

    <!-- ============ Main Workspace (2-Column Desktop / Single-Column Mobile) ============ -->
    <main class="flex-1 w-full max-w-5xl mx-auto px-4 sm:px-6 py-6 sm:py-7">
      <div class="grid grid-cols-1 lg:grid-cols-[1fr_320px] gap-6 items-start">
        <!-- ============ LEFT: Primary Column ============ -->
        <div class="min-w-0 space-y-4">
          <!-- ==================== TAB 1: CREATE REPORT ==================== -->
          <template v-if="activeTab === 'create'">
            <!-- Loading skeleton for QR-linked item -->
            <div v-if="loadingItem" class="panel p-4 flex items-center gap-3 animate-pulse">
              <div class="w-12 h-12 rounded-[9px] bg-panel-2 border border-line shrink-0" />
              <div class="flex-1 space-y-2">
                <div class="h-3 w-32 rounded bg-panel-2" />
                <div class="h-4 w-56 rounded bg-panel-2" />
              </div>
            </div>

            <!-- Detected QR Asset Context Card -->
            <div
              v-else-if="item"
              class="panel overflow-hidden border-[var(--acc-border)]"
            >
              <div class="px-4 py-2 bg-acc-50 border-b border-[var(--acc-border)] flex items-center justify-between gap-2">
                <div class="flex items-center gap-2 text-[11px] font-bold uppercase tracking-[0.06em] text-acc-600 dark:text-acc-300">
                  <i class="pi pi-qrcode text-[12px]" aria-hidden="true" />
                  <span>Barang Terdeteksi via QR</span>
                </div>
                <span class="t-mono !text-[10.5px] text-acc-600 dark:text-acc-300 font-semibold">
                  {{ item.sku }}
                </span>
              </div>
              <div class="p-4 flex items-start gap-3.5">
                <button
                  v-if="item.scanned_unit?.photo_url || item.photo_url"
                  type="button"
                  class="w-14 h-14 rounded-[9px] border border-line overflow-hidden bg-panel-2 shrink-0 cursor-pointer"
                  aria-label="Lihat foto barang"
                  @click="openLightbox(item.scanned_unit?.photo_url || item.photo_url)"
                >
                  <img
                    :src="item.scanned_unit?.photo_url || item.photo_url"
                    :alt="item.name"
                    class="w-full h-full object-cover"
                  />
                </button>
                <div
                  v-else
                  class="w-14 h-14 rounded-[9px] border border-line bg-panel-2 grid place-items-center shrink-0 text-acc-500"
                >
                  <i class="pi pi-box text-lg" aria-hidden="true" />
                </div>

                <div class="min-w-0 flex-1">
                  <div class="flex items-center gap-2 flex-wrap">
                    <div class="text-[14.5px] font-bold text-txt leading-snug">{{ item.name }}</div>
                    <StatusChip
                      v-if="item.scanned_unit?.condition_status || item.condition_status"
                      :kind="item.scanned_unit?.condition_status || item.condition_status"
                    />
                  </div>
                  <div class="mt-1.5 flex items-center gap-x-3 gap-y-1 flex-wrap text-[12px] text-txt-dim">
                    <span class="inline-flex items-center gap-1">
                      <i class="pi pi-map-marker text-[10.5px] text-acc-500" aria-hidden="true" />
                      {{ item.scanned_unit?.location || item.location || 'Lokasi belum tercatat' }}
                    </span>
                    <span v-if="item.scanned_unit" class="t-mono !text-[11px] px-1.5 py-0.5 rounded bg-panel-2 border border-line text-txt">
                      Unit #{{ item.scanned_unit.unit_number }}
                      <template v-if="item.scanned_unit.register_number"> · Reg {{ item.scanned_unit.register_number }}</template>
                    </span>
                  </div>
                </div>
              </div>
            </div>

            <!-- Main Form Card -->
            <div class="panel p-5 sm:p-6">
              <div class="flex items-center justify-between gap-2 pb-4 border-b border-line">
                <div>
                  <h2 class="text-[16px] font-bold tracking-[-0.02em] text-txt">Detail Pengaduan</h2>
                  <p class="text-[12px] text-txt-dim mt-0.5">
                    Kolom bertanda <span class="text-sig-bad font-semibold">*</span> wajib diisi agar petugas dapat bertindak cepat.
                  </p>
                </div>
                <span class="hidden sm:inline-flex t-mono !text-[10.5px] px-2 py-0.5 rounded border border-line bg-panel-2 text-txt-dim">
                  FORM-PUB-01
                </span>
              </div>

              <div
                v-if="errorMessage"
                role="alert"
                class="mt-4 p-3 rounded-[9px] bg-[var(--sig-bad-bg)] border border-[var(--sig-bad-line)] text-sig-bad text-[12.5px] flex items-start gap-2.5"
              >
                <i class="pi pi-exclamation-circle mt-0.5 shrink-0" aria-hidden="true" />
                <span>{{ errorMessage }}</span>
              </div>

              <form @submit.prevent="submitReport" class="mt-5 flex flex-col gap-5">
                <!-- Step 1: Visual Category Picker -->
                <fieldset class="flex flex-col gap-2">
                  <legend class="text-[12px] font-semibold text-txt mb-1">
                    1. Pilih Kategori Masalah <span class="text-sig-bad">*</span>
                  </legend>
                  <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
                    <button
                      v-for="opt in reportTypeOptions"
                      :key="opt.value"
                      type="button"
                      :aria-pressed="reportType === opt.value"
                      class="group text-left p-3 rounded-[9px] border transition-all cursor-pointer flex items-start gap-2.5"
                      :class="
                        reportType === opt.value
                          ? 'border-acc-500 bg-acc-50 ring-1 ring-[var(--acc-ring)]'
                          : 'border-line bg-panel-2 hover:border-acc-500/50'
                      "
                      @click="reportType = opt.value"
                    >
                      <div
                        class="w-8 h-8 rounded-[7px] grid place-items-center shrink-0 transition-colors"
                        :class="
                          reportType === opt.value
                            ? 'bg-acc-500 text-white dark:text-[#042f2e]'
                            : 'bg-panel border border-line text-txt-dim group-hover:text-acc-500'
                        "
                      >
                        <i :class="opt.icon" class="text-[13px]" aria-hidden="true" />
                      </div>
                      <div class="min-w-0 flex-1">
                        <div class="text-[12.5px] font-semibold text-txt leading-tight">{{ opt.title }}</div>
                        <div class="text-[11px] text-txt-dim mt-0.5 leading-snug">{{ opt.desc }}</div>
                      </div>
                    </button>
                  </div>
                </fieldset>

                <!-- Step 2: Reporter Identity + Optional Contact -->
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <label class="flex flex-col gap-1.5">
                    <span class="text-[12px] font-semibold text-txt">
                      2. Nama Lengkap Pelapor <span class="text-sig-bad">*</span>
                    </span>
                    <InputText
                      v-model="reporterName"
                      placeholder="Contoh: Budi Santoso (Poli Umum)"
                      class="w-full"
                      :invalid="submitted && !reporterName.trim()"
                    />
                  </label>

                  <label class="flex flex-col gap-1.5">
                    <span class="text-[12px] font-semibold text-txt">
                      Kontak / WhatsApp <span class="text-[11px] font-normal text-txt-dim">(opsional)</span>
                    </span>
                    <InputText
                      v-model="reporterContact"
                      placeholder="08xx-xxxx-xxxx (untuk konfirmasi)"
                      class="w-full"
                    />
                  </label>
                </div>

                <!-- Step 3: Location -->
                <label class="flex flex-col gap-1.5">
                  <div class="flex items-center justify-between gap-2">
                    <span class="text-[12px] font-semibold text-txt">
                      3. Lokasi / Ruangan Barang <span class="text-sig-bad">*</span>
                    </span>
                    <span
                      v-if="item && location"
                      class="text-[10.5px] font-medium text-acc-600 dark:text-acc-300 flex items-center gap-1"
                    >
                      <i class="pi pi-check-circle text-[10px]" aria-hidden="true" />
                      Terisi otomatis dari QR
                    </span>
                  </div>
                  <Select
                    v-if="locations.length"
                    v-model="location"
                    :options="locations"
                    optionLabel="name"
                    optionValue="name"
                    filter
                    editable
                    placeholder="Pilih atau ketik nama ruangan..."
                    class="w-full"
                    :invalid="submitted && !location.trim()"
                  />
                  <InputText
                    v-else
                    v-model="location"
                    placeholder="Contoh: Poli Umum Lt. 1, Ruang IGD, Gudang Farmasi"
                    class="w-full"
                    :invalid="submitted && !location.trim()"
                  />
                </label>

                <!-- Step 4: Problem Description -->
                <label class="flex flex-col gap-1.5">
                  <div class="flex items-center justify-between gap-2">
                    <span class="text-[12px] font-semibold text-txt">
                      4. Uraian Masalah / Kronologi <span class="text-sig-bad">*</span>
                    </span>
                    <span class="t-mono !text-[10.5px] text-txt-dim">{{ description.length }} karakter</span>
                  </div>
                  <Textarea
                    v-model="description"
                    rows="4"
                    autoResize
                    placeholder="Jelaskan gejala kerusakan, bagian yang bermasalah, atau informasi tambahan yang membantu petugas..."
                    class="w-full"
                    :invalid="submitted && !description.trim()"
                  />
                </label>

                <!-- Step 5: Evidence Photo Dropzone -->
                <div class="flex flex-col gap-1.5">
                  <div class="flex items-center justify-between gap-2">
                    <span class="text-[12px] font-semibold text-txt">
                      5. Foto Bukti Kondisi <span class="text-[11px] font-normal text-txt-dim">(opsional, maks. 10 MB)</span>
                    </span>
                    <span v-if="photoUrl && !uploadingPhoto" class="text-[11px] font-semibold text-sig-ok flex items-center gap-1">
                      <i class="pi pi-check-circle text-[11px]" aria-hidden="true" />
                      Terunggah
                    </span>
                  </div>

                  <input
                    ref="fileInputRef"
                    type="file"
                    accept="image/*"
                    class="hidden"
                    @change="onPhotoSelected"
                  />

                  <div
                    v-if="!photoPreview"
                    role="button"
                    tabindex="0"
                    class="rounded-[10px] border-2 border-dashed p-5 text-center transition-all cursor-pointer flex flex-col items-center justify-center gap-2"
                    :class="
                      dragOver
                        ? 'border-acc-500 bg-acc-50'
                        : 'border-line bg-panel-2 hover:border-acc-500/60 hover:bg-acc-50/40'
                    "
                    @click="fileInputRef?.click()"
                    @keydown.enter.prevent="fileInputRef?.click()"
                    @keydown.space.prevent="fileInputRef?.click()"
                    @dragover.prevent="dragOver = true"
                    @dragleave.prevent="dragOver = false"
                    @drop.prevent="onPhotoDrop"
                  >
                    <div class="w-11 h-11 rounded-[10px] bg-panel border border-line grid place-items-center text-acc-500">
                      <i class="pi pi-camera text-lg" aria-hidden="true" />
                    </div>
                    <div>
                      <div class="text-[13px] font-semibold text-txt">
                        Ambil Foto atau Seret Gambar ke Sini
                      </div>
                      <div class="text-[11.5px] text-txt-dim mt-0.5">
                        Mendukung kamera HP · JPG, PNG, atau WebP
                      </div>
                    </div>
                  </div>

                  <div
                    v-else
                    class="rounded-[10px] border border-line bg-panel-2 p-3 flex flex-col sm:flex-row items-start sm:items-center gap-3.5"
                  >
                    <div class="relative w-full sm:w-28 h-36 sm:h-20 rounded-[8px] overflow-hidden border border-line bg-panel shrink-0">
                      <img :src="photoPreview" alt="Preview foto laporan" class="w-full h-full object-cover" />
                      <div
                        v-if="uploadingPhoto"
                        class="absolute inset-0 bg-black/55 grid place-items-center text-white text-[11px] font-semibold"
                      >
                        <i class="pi pi-spin pi-spinner text-base" aria-hidden="true" />
                      </div>
                    </div>
                    <div class="flex-1 min-w-0">
                      <div class="text-[12.5px] font-semibold text-txt">
                        {{ uploadingPhoto ? 'Mengunggah foto ke server...' : 'Foto bukti siap dilampirkan' }}
                      </div>
                      <div class="text-[11.5px] text-txt-dim mt-0.5">
                        Klik pratinjau atau ganti jika gambar kurang jelas.
                      </div>
                      <div class="mt-2.5 flex items-center gap-2 flex-wrap">
                        <Button
                          type="button"
                          label="Ganti Foto"
                          icon="pi pi-refresh"
                          size="small"
                          outlined
                          severity="secondary"
                          :disabled="uploadingPhoto"
                          @click="fileInputRef?.click()"
                        />
                        <Button
                          type="button"
                          label="Hapus"
                          icon="pi pi-trash"
                          size="small"
                          text
                          severity="danger"
                          :disabled="uploadingPhoto"
                          @click="clearPhoto"
                        />
                      </div>
                    </div>
                  </div>

                  <div v-if="photoError" role="alert" class="text-[12px] text-sig-bad flex items-center gap-1.5 mt-1">
                    <i class="pi pi-exclamation-triangle text-[11px]" aria-hidden="true" />
                    <span>{{ photoError }}</span>
                  </div>
                </div>

                <!-- Submit Action Bar -->
                <div class="pt-3 border-t border-line flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                  <p class="text-[11.5px] text-txt-dim flex items-center gap-1.5">
                    <i class="pi pi-shield text-acc-500 text-[12px]" aria-hidden="true" />
                    <span>Nomor kontak disamarkan otomatis pada halaman lacak publik.</span>
                  </p>
                  <Button
                    type="submit"
                    label="Kirim Laporan Sekarang"
                    icon="pi pi-send"
                    :loading="submitting"
                    :disabled="uploadingPhoto"
                    class="w-full sm:w-auto shrink-0"
                  />
                </div>
              </form>
            </div>
          </template>

          <!-- ==================== TAB 2: TRACK REPORT ==================== -->
          <template v-else>
            <!-- Search Card -->
            <div class="panel p-5 sm:p-6">
              <h2 class="text-[16px] font-bold tracking-[-0.02em] text-txt">Cari Nomor Tiket Laporan</h2>
              <p class="text-[12.5px] text-txt-dim mt-0.5">
                Masukkan nomor tiket yang Anda terima setelah mengirim laporan (contoh:
                <span class="t-mono text-txt font-medium">LAP-20261120-A1B2C3</span>).
              </p>

              <form @submit.prevent="fetchTracking()" class="mt-4 flex flex-col sm:flex-row gap-2">
                <InputText
                  v-model="trackQuery"
                  placeholder="LAP-YYYYMMDD-XXXXXXXXXXXX"
                  class="flex-1 t-mono"
                />
                <Button
                  type="submit"
                  label="Lacak Status"
                  icon="pi pi-search"
                  :loading="trackLoading"
                  class="shrink-0"
                />
              </form>

              <!-- Recent Tickets Quick Chips -->
              <div v-if="recentTickets.length" class="mt-4 pt-3.5 border-t border-line">
                <div class="text-[11px] font-semibold uppercase tracking-[0.06em] text-txt-dim mb-2">
                  Tiket Terakhir dari Perangkat Ini
                </div>
                <div class="flex flex-wrap gap-1.5">
                  <button
                    v-for="rt in recentTickets"
                    :key="rt.number"
                    type="button"
                    class="px-2.5 py-1.5 rounded-[7px] border border-line bg-panel-2 hover:border-acc-500 text-left transition-colors cursor-pointer flex items-center gap-2"
                    @click="fetchTracking(rt.number)"
                  >
                    <span class="t-mono !text-[11px] font-semibold text-acc-600 dark:text-acc-300">{{ rt.number }}</span>
                    <span class="text-[11px] text-txt-dim truncate max-w-[140px]">· {{ rt.location }}</span>
                  </button>
                </div>
              </div>

              <div
                v-if="trackError"
                role="alert"
                class="mt-4 p-3 rounded-[9px] bg-[var(--sig-bad-bg)] border border-[var(--sig-bad-line)] text-sig-bad text-[12.5px] flex items-center gap-2"
              >
                <i class="pi pi-exclamation-circle shrink-0" aria-hidden="true" />
                <span>{{ trackError }}</span>
              </div>
            </div>

            <!-- Tracking Result Dossier -->
            <div v-if="trackResult" class="panel overflow-hidden">
              <!-- Ticket Header -->
              <div class="p-4 sm:p-5 bg-panel-2 border-b border-line flex flex-wrap items-center justify-between gap-3">
                <div>
                  <div class="text-[10.5px] font-semibold uppercase tracking-[0.07em] text-txt-dim">
                    Nomor Tiket Resmi
                  </div>
                  <div class="flex items-center gap-2 mt-0.5">
                    <span class="t-mono text-[16px] sm:text-[18px] font-bold text-txt">
                      {{ trackResult.report_number }}
                    </span>
                    <button
                      type="button"
                      class="p-1.5 rounded-[6px] border border-line bg-panel hover:border-acc-500 text-txt-dim hover:text-acc-500 cursor-pointer transition-colors"
                      :aria-label="copiedTracked ? 'Nomor tiket tersalin' : 'Salin nomor tiket'"
                      @click="copyText(trackResult.report_number, 'tracked')"
                    >
                      <i :class="copiedTracked ? 'pi pi-check text-sig-ok' : 'pi pi-copy'" class="text-[11px]" aria-hidden="true" />
                    </button>
                  </div>
                </div>
                <StatusChip :kind="trackResult.status" />
              </div>

              <!-- Visual 3-Step Progress Bar -->
              <div class="p-4 sm:p-5 border-b border-line bg-panel">
                <div class="grid grid-cols-3 gap-2">
                  <div
                    v-for="(step, idx) in trackSteps"
                    :key="step.key"
                    class="relative flex flex-col items-center text-center px-1"
                  >
                    <div
                      v-if="idx < trackSteps.length - 1"
                      class="hidden sm:block absolute top-4 left-[calc(50%+18px)] right-[calc(-50%+18px)] h-[2px]"
                      :class="idx < trackStepIndex ? 'bg-acc-500' : 'bg-line'"
                      aria-hidden="true"
                    />
                    <div
                      class="relative z-10 w-8 h-8 rounded-full grid place-items-center border text-[12px] transition-colors"
                      :class="
                        idx <= trackStepIndex
                          ? trackResult.status === 'REJECTED' && idx === 2
                            ? 'bg-sig-bad text-white border-sig-bad'
                            : 'bg-acc-500 text-white dark:text-[#042f2e] border-acc-500'
                          : 'bg-panel-2 text-txt-dim border-line'
                      "
                    >
                      <i :class="step.icon" aria-hidden="true" />
                    </div>
                    <div
                      class="mt-2 text-[11.5px] font-semibold leading-tight"
                      :class="idx <= trackStepIndex ? 'text-txt' : 'text-txt-dim'"
                    >
                      {{ step.label }}
                    </div>
                    <div class="text-[10.5px] text-txt-dim mt-0.5 hidden sm:block">{{ step.sub }}</div>
                  </div>
                </div>
              </div>

              <!-- Structured Metadata Grid -->
              <div class="p-4 sm:p-5 space-y-4">
                <dl class="grid grid-cols-1 sm:grid-cols-2 gap-3 text-[12.5px]">
                  <div class="p-3 rounded-[8px] border border-line bg-panel-2">
                    <dt class="text-[10.5px] uppercase tracking-[0.05em] text-txt-dim font-semibold">Pelapor</dt>
                    <dd class="font-semibold text-txt mt-0.5">{{ trackResult.reporter_name }}</dd>
                    <dd v-if="trackResult.reporter_contact && trackResult.reporter_contact !== '-'" class="t-mono !text-[11px] text-txt-dim mt-0.5">
                      {{ trackResult.reporter_contact }}
                    </dd>
                  </div>
                  <div class="p-3 rounded-[8px] border border-line bg-panel-2">
                    <dt class="text-[10.5px] uppercase tracking-[0.05em] text-txt-dim font-semibold">Kategori &amp; Waktu</dt>
                    <dd class="font-semibold text-txt mt-0.5">{{ trackResult.report_type }}</dd>
                    <dd class="text-[11px] text-txt-dim mt-0.5">{{ formatDate(trackResult.created_at) }}</dd>
                  </div>
                  <div class="p-3 rounded-[8px] border border-line bg-panel-2">
                    <dt class="text-[10.5px] uppercase tracking-[0.05em] text-txt-dim font-semibold">Lokasi / Ruangan</dt>
                    <dd class="font-semibold text-txt mt-0.5">{{ trackResult.location || '-' }}</dd>
                  </div>
                  <div v-if="trackResult.item_name" class="p-3 rounded-[8px] border border-line bg-panel-2">
                    <dt class="text-[10.5px] uppercase tracking-[0.05em] text-txt-dim font-semibold">Barang Terkait</dt>
                    <dd class="font-semibold text-txt mt-0.5">{{ trackResult.item_name }}</dd>
                    <dd v-if="trackResult.item_sku" class="t-mono !text-[11px] text-txt-dim mt-0.5">{{ trackResult.item_sku }}</dd>
                  </div>
                </dl>

                <div class="p-3.5 rounded-[9px] border border-line bg-panel-2">
                  <div class="text-[11px] uppercase tracking-[0.05em] font-semibold text-txt-dim mb-1.5">
                    Uraian Masalah
                  </div>
                  <p class="text-[13px] text-txt whitespace-pre-line leading-relaxed">
                    {{ trackResult.description }}
                  </p>
                </div>

                <!-- Attached Photo -->
                <div v-if="trackResult.photo_url">
                  <div class="text-[11px] uppercase tracking-[0.05em] font-semibold text-txt-dim mb-1.5">
                    Lampiran Foto Bukti
                  </div>
                  <button
                    type="button"
                    class="group relative block rounded-[9px] overflow-hidden border border-line bg-panel-2 cursor-pointer"
                    aria-label="Perbesar foto lampiran"
                    @click="openLightbox(trackResult.photo_url)"
                  >
                    <img
                      :src="trackResult.photo_url"
                      alt="Foto bukti laporan"
                      class="max-h-56 w-auto object-cover"
                    />
                    <span class="absolute bottom-2 right-2 px-2 py-1 rounded-[6px] bg-black/65 text-white text-[10.5px] font-medium flex items-center gap-1">
                      <i class="pi pi-search-plus text-[10px]" aria-hidden="true" />
                      Perbesar
                    </span>
                  </button>
                </div>

                <!-- Official Admin Response -->
                <div
                  v-if="trackResult.admin_notes"
                  class="p-4 rounded-[10px] border border-[var(--acc-border)] bg-acc-50"
                >
                  <div class="flex items-center justify-between gap-2 mb-1.5">
                    <div class="text-[11.5px] font-bold uppercase tracking-[0.05em] text-acc-600 dark:text-acc-300 flex items-center gap-1.5">
                      <i class="pi pi-verified text-[12px]" aria-hidden="true" />
                      <span>Tanggapan &amp; Tindak Lanjut Petugas</span>
                    </div>
                    <span v-if="trackResult.handled_at" class="text-[11px] text-txt-dim">
                      {{ formatDate(trackResult.handled_at) }}
                    </span>
                  </div>
                  <p class="text-[13px] text-txt whitespace-pre-line leading-relaxed">
                    {{ trackResult.admin_notes }}
                  </p>
                </div>
              </div>
            </div>
          </template>
        </div>

        <!-- ============ RIGHT: Context & Process Sidebar ============ -->
        <aside class="space-y-4 lg:sticky lg:top-20">
          <!-- Process Timeline Card -->
          <div class="panel p-4 sm:p-5">
            <div class="t-mono !text-[10.5px] uppercase tracking-[0.08em] text-acc-500 font-semibold">
              Alur Penanganan
            </div>
            <h3 class="text-[14.5px] font-bold text-txt mt-1">Bagaimana laporan diproses?</h3>
            <ol class="mt-4 space-y-3.5">
              <li
                v-for="s in [
                  { n: '01', t: 'Kirim & Terima Tiket', d: 'Sistem menerbitkan kode unik LAP-... sebagai bukti laporan resmi.' },
                  { n: '02', t: 'Verifikasi Petugas', d: 'Tim sarpras memeriksa laporan, foto bukti, dan lokasi barang.' },
                  { n: '03', t: 'Perbaikan & Update', d: 'Status dan catatan tindak lanjut dapat dipantau kapan saja.' },
                ]"
                :key="s.n"
                class="flex items-start gap-3"
              >
                <span class="t-mono !text-[11px] font-bold px-2 py-1 rounded-[6px] bg-panel-2 border border-line text-acc-600 dark:text-acc-300 shrink-0">
                  {{ s.n }}
                </span>
                <div class="min-w-0">
                  <div class="text-[12.5px] font-semibold text-txt leading-snug">{{ s.t }}</div>
                  <div class="text-[11.5px] text-txt-dim mt-0.5 leading-relaxed">{{ s.d }}</div>
                </div>
              </li>
            </ol>
          </div>

          <!-- Practical Tips Card -->
          <div class="panel p-4 sm:p-5 bg-panel-2">
            <div class="flex items-center gap-2 text-[12.5px] font-bold text-txt">
              <i class="pi pi-info-circle text-acc-500" aria-hidden="true" />
              <span>Tips Laporan Cepat Ditangani</span>
            </div>
            <ul class="mt-3 space-y-2 text-[12px] text-txt-dim leading-relaxed">
              <li class="flex items-start gap-2">
                <i class="pi pi-check text-acc-500 text-[10px] mt-1 shrink-0" aria-hidden="true" />
                <span>Pindai <strong class="text-txt">QR Label</strong> pada barang bila tersedia agar kode &amp; lokasi terisi otomatis.</span>
              </li>
              <li class="flex items-start gap-2">
                <i class="pi pi-check text-acc-500 text-[10px] mt-1 shrink-0" aria-hidden="true" />
                <span>Lampirkan foto kondisi kerusakan dengan pencahayaan jelas.</span>
              </li>
              <li class="flex items-start gap-2">
                <i class="pi pi-check text-acc-500 text-[10px] mt-1 shrink-0" aria-hidden="true" />
                <span>Simpan nomor tiket untuk mengecek perkembangan penanganan.</span>
              </li>
            </ul>
          </div>

          <!-- Quick Switch CTA -->
          <div class="panel p-4 flex items-center justify-between gap-3">
            <div class="min-w-0">
              <div class="text-[12.5px] font-semibold text-txt">
                {{ activeTab === 'create' ? 'Sudah punya nomor tiket?' : 'Ingin melaporkan kendala baru?' }}
              </div>
              <div class="text-[11px] text-txt-dim mt-0.5">
                {{ activeTab === 'create' ? 'Cek progres tindak lanjut di sini.' : 'Kirim laporan tanpa perlu akun.' }}
              </div>
            </div>
            <Button
              :label="activeTab === 'create' ? 'Lacak' : 'Lapor'"
              :icon="activeTab === 'create' ? 'pi pi-search' : 'pi pi-plus'"
              size="small"
              outlined
              @click="activeTab = activeTab === 'create' ? 'track' : 'create'"
            />
          </div>
        </aside>
      </div>
    </main>

    <!-- ============ Footer ============ -->
    <footer class="border-t border-line bg-panel py-4">
      <div class="max-w-5xl mx-auto px-4 sm:px-6 flex flex-wrap items-center justify-between gap-2 text-[11.5px] text-txt-dim">
        <span class="t-mono !text-[11px]">© 2026 · Inventaris &amp; Sarpras Puskesmas</span>
        <span>Portal Pengaduan Publik</span>
      </div>
    </footer>

    <!-- ============ Success Ticket Dialog ============ -->
    <Dialog
      v-model:visible="submitSuccess"
      modal
      header="Laporan Berhasil Dikirim"
      :style="{ width: '440px', maxWidth: '94vw' }"
    >
      <div class="flex flex-col items-center text-center py-2">
        <div class="w-14 h-14 rounded-2xl bg-[var(--sig-ok-bg)] border border-[var(--sig-ok-line)] text-sig-ok grid place-items-center mb-3">
          <i class="pi pi-check-circle text-2xl" aria-hidden="true" />
        </div>
        <div class="text-[16px] font-bold text-txt">Terima kasih atas laporan Anda!</div>
        <p class="text-[12.5px] text-txt-dim mt-1 max-w-[38ch]">
          Simpan nomor tiket berikut untuk memantau status perbaikan dan catatan petugas.
        </p>

        <div class="w-full mt-4 p-4 rounded-[10px] border border-[var(--acc-border)] bg-acc-50">
          <div class="text-[10.5px] uppercase tracking-[0.07em] font-semibold text-acc-600 dark:text-acc-300">
            Nomor Tiket Laporan
          </div>
          <div class="mt-1.5 flex items-center justify-center gap-2">
            <span class="t-mono text-[17px] font-bold text-txt select-all">
              {{ createdTicket?.report_number }}
            </span>
            <Button
              :icon="copiedTicket ? 'pi pi-check' : 'pi pi-copy'"
              size="small"
              text
              rounded
              :severity="copiedTicket ? 'success' : 'secondary'"
              :aria-label="copiedTicket ? 'Tersalin' : 'Salin nomor tiket'"
              @click="copyText(createdTicket?.report_number, 'created')"
            />
          </div>
          <div v-if="copiedTicket" class="text-[11px] font-semibold text-sig-ok mt-1">
            Nomor tiket berhasil disalin!
          </div>
        </div>

        <div class="mt-5 w-full flex flex-col sm:flex-row gap-2">
          <Button
            label="Lacak Status Sekarang"
            icon="pi pi-search"
            class="flex-1"
            @click="trackCreatedTicket"
          />
          <Button
            label="Tutup"
            severity="secondary"
            outlined
            @click="submitSuccess = false"
          />
        </div>
      </div>
    </Dialog>

    <!-- ============ Photo Lightbox Dialog ============ -->
    <Dialog
      v-model:visible="lightboxVisible"
      modal
      header="Foto Lampiran"
      :style="{ width: '640px', maxWidth: '94vw' }"
    >
      <div class="grid place-items-center bg-panel-2 rounded-[8px] border border-line p-2">
        <img
          v-if="lightboxUrl"
          :src="lightboxUrl"
          alt="Pratinjau foto ukuran penuh"
          class="max-h-[70vh] w-auto rounded-[6px] object-contain"
        />
      </div>
    </Dialog>
  </div>
</template>
