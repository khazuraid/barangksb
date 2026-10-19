<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '@/api'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Textarea from 'primevue/textarea'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import Dialog from 'primevue/dialog'

const route = useRoute()
const router = useRouter()

const activeTab = ref<'create' | 'track'>('create')

// Form Laporan
const reporterName = ref('')
const reportType = ref('Kerusakan Alat / Barang')
const description = ref('')
const location = ref('')
const locations = ref<{ id: string; name: string; code: string }[]>([])
const photoUrl = ref('')
const uploadingPhoto = ref(false)
const photoPreview = ref('')

const reportTypeOptions = [
  { label: '🛠️ Kerusakan Alat / Barang', value: 'Kerusakan Alat / Barang' },
  { label: '🔍 Kehilangan / Hilang', value: 'Kehilangan / Tidak Ditemukan' },
  { label: '🚚 Perpindahan Ruangan / Lokasi', value: 'Perpindahan Ruangan / Lokasi' },
  { label: '🧹 Kondisi Fisik & Kebersihan', value: 'Kondisi Fisik & Kebersihan' },
  { label: '📋 Masalah Lainnya', value: 'Lainnya' },
]

// Item Info jika discan dari QR
const item = ref<any>(null)
const loadingItem = ref(false)

// Tracking
const trackQuery = ref('')
const trackResult = ref<any>(null)
const trackLoading = ref(false)
const trackError = ref('')

// Submission status
const submitting = ref(false)
const submitSuccess = ref(false)
const createdTicket = ref<any>(null)
const errorMessage = ref('')
const submitted = ref(false)

onMounted(async () => {
  // Muat daftar lokasi dari database untuk dropdown
  try {
    const locRes = await api.get('/public/locations')
    locations.value = locRes.data || []
  } catch {
    // fallback: tetap bisa isi manual via InputText
  }

  const targetId = route.params.id || route.query.id || route.query.sku
  if (targetId) {
    loadingItem.value = true
    try {
      const res = await api.get(`/scan/${targetId}`)
      item.value = res.data
      if (item.value?.location) {
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

async function onPhotoSelected(e: Event) {
  const input = e.target as HTMLInputElement
  if (!input.files || input.files.length === 0) return

  const file = input.files[0]
  if (file.size > 10 * 1024 * 1024) {
    alert('Ukuran foto terlalu besar. Maksimal 10 MB.')
    return
  }

  // Local preview
  photoPreview.value = URL.createObjectURL(file)

  const formData = new FormData()
  formData.append('photo', file)

  uploadingPhoto.value = true
  try {
    const res = await api.post('/public/upload', formData)
    photoUrl.value = res.data.url
  } catch (err: any) {
    alert(err?.response?.data?.error || 'Gagal mengunggah foto. Pastikan format gambar JPG/PNG/WebP.')
    photoPreview.value = ''
    photoUrl.value = ''
  } finally {
    uploadingPhoto.value = false
  }
}

async function submitReport() {
  errorMessage.value = ''
  submitted.value = true
  if (!reporterName.value.trim() || !description.value.trim()) {
    errorMessage.value = 'Mohon isi Nama Lengkap dan Uraian Masalah.'
    return
  }
  if (!location.value.trim()) {
    errorMessage.value = 'Mohon pilih Lokasi / Ruangan barang saat ini.'
    return
  }

  submitting.value = true
  try {
    const payload = {
      reporter_name: reporterName.value.trim(),
      reporter_contact: '-',
      report_type: reportType.value,
      description: description.value.trim(),
      location: location.value.trim(),
      photo_url: photoUrl.value,
      item_id: item.value?.id || undefined,
      item_sku: item.value?.sku || (route.query.sku ? String(route.query.sku) : undefined),
      item_name: item.value?.name || undefined,
    }

    const res = await api.post('/public/reports', payload)
    createdTicket.value = res.data
    submitSuccess.value = true

    // Reset Form
    reporterName.value = ''
    description.value = ''
    photoUrl.value = ''
    photoPreview.value = ''
  } catch (err: any) {
    errorMessage.value = err?.response?.data?.error || 'Gagal mengirim laporan. Silakan coba kembali.'
  } finally {
    submitting.value = false
  }
}

async function fetchTracking() {
  if (!trackQuery.value.trim()) return
  trackLoading.value = true
  trackError.value = ''
  trackResult.value = null
  try {
    const res = await api.get(`/public/reports/${encodeURIComponent(trackQuery.value.trim())}`)
    trackResult.value = res.data
  } catch {
    trackError.value = 'Nomor tiket laporan tidak ditemukan. Pastikan nomor yang dimasukkan sudah benar.'
  } finally {
    trackLoading.value = false
  }
}

function copyTicketNumber() {
  if (createdTicket.value?.report_number) {
    navigator.clipboard.writeText(createdTicket.value.report_number)
    alert('Nomor tiket berhasil disalin ke papan klip!')
  }
}

function getStatusSeverity(status?: string) {
  switch (status) {
    case 'PENDING': return 'warn'
    case 'IN_PROGRESS': return 'info'
    case 'RESOLVED': return 'success'
    case 'REJECTED': return 'danger'
    default: return 'secondary'
  }
}

function getStatusLabel(status?: string) {
  switch (status) {
    case 'PENDING': return '⏳ Menunggu Verifikasi'
    case 'IN_PROGRESS': return '🔧 Sedang Ditindaklanjuti'
    case 'RESOLVED': return '✅ Selesai Ditangani'
    case 'REJECTED': return '❌ Ditolak / Tidak Valid'
    default: return status || '—'
  }
}

function formatDate(iso?: string) {
  if (!iso) return '—'
  return new Date(iso).toLocaleString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}
</script>

<template>
  <div class="min-h-screen bg-slate-50 text-slate-800 flex flex-col justify-between py-6 px-4 sm:px-6">
    <div class="w-full max-w-xl mx-auto flex flex-col gap-5">
      <!-- HEADER BRAND -->
      <div class="flex items-center justify-between border-b border-slate-200 pb-4">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-indigo-600 flex items-center justify-center font-black text-white text-base shadow-md shadow-indigo-600/20">
            IK
          </div>
          <div>
            <h1 class="text-sm font-black uppercase tracking-wider text-slate-900">Layanan Pengaduan Aset</h1>
            <p class="text-[11px] text-slate-500">Portal Pelaporan Kerusakan &amp; Kendala Barang Publik</p>
          </div>
        </div>

        <button
          type="button"
          @click="router.push('/login')"
          class="text-xs text-indigo-600 hover:text-indigo-700 font-medium flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg hover:bg-slate-100 transition-colors"
        >
          <i class="pi pi-lock text-[11px]" />
          <span>Login Petugas</span>
        </button>
      </div>

      <!-- TAB PILIHAN (BUAT LAPORAN / LACAK STATUS) -->
      <div class="grid grid-cols-2 p-1 bg-slate-200/80 rounded-xl border border-slate-200 text-xs font-bold text-center">
        <button
          type="button"
          class="py-2.5 rounded-lg transition-all flex items-center justify-center gap-2"
          :class="activeTab === 'create' ? 'bg-white text-indigo-600 shadow-sm' : 'text-slate-600 hover:text-slate-900'"
          @click="activeTab = 'create'"
        >
          <i class="pi pi-pencil text-xs" />
          <span>Buat Pengaduan</span>
        </button>
        <button
          type="button"
          class="py-2.5 rounded-lg transition-all flex items-center justify-center gap-2"
          :class="activeTab === 'track' ? 'bg-white text-indigo-600 shadow-sm' : 'text-slate-600 hover:text-slate-900'"
          @click="activeTab = 'track'"
        >
          <i class="pi pi-search text-xs" />
          <span>Lacak Aduan</span>
        </button>
      </div>

      <!-- ================= TAB 1: FORMULIR PENGADUAN ================= -->
      <div v-if="activeTab === 'create'" class="flex flex-col gap-4">
        <!-- Asset Context Card (Jika scan dari QR Code) -->
        <div v-if="item" class="p-3.5 bg-indigo-50/70 border border-indigo-100 rounded-xl flex items-start gap-3 shadow-sm">
          <div class="w-16 h-16 rounded-lg bg-white overflow-hidden shrink-0 border border-indigo-100 flex items-center justify-center">
            <img v-if="item.photo_url" :src="item.photo_url" alt="Foto Barang" class="w-full h-full object-cover" />
            <i v-else class="pi pi-box text-2xl text-indigo-400" />
          </div>
          <div class="flex-1 min-w-0">
            <div class="text-[10px] font-bold text-indigo-600 uppercase tracking-wider">Aset Teridentifikasi (QR Scan)</div>
            <h2 class="text-sm font-bold text-slate-900 truncate">{{ item.name }}</h2>
            <div class="text-xs font-mono text-slate-600 mt-0.5">{{ item.sku }}</div>
            <div class="text-[11px] text-slate-500 flex items-center gap-1.5 mt-1">
              <i class="pi pi-map-marker text-[10px] text-emerald-600" />
              <span class="truncate">{{ item.location || 'Lokasi belum diset' }}</span>
            </div>
          </div>
        </div>

        <!-- FORM BOX -->
        <div class="bg-white border border-slate-200 rounded-2xl p-5 shadow-sm flex flex-col gap-4">
          <div>
            <h3 class="text-sm font-bold text-slate-900 flex items-center gap-2">
              <i class="pi pi-file-edit text-indigo-600" />
              <span>Formulir Laporan Kerusakan / Kendala</span>
            </h3>
            <p class="text-[11px] text-slate-500 mt-0.5">
              Laporan Anda akan langsung masuk ke sistem dan ditindaklanjuti oleh petugas pengelola barang.
            </p>
          </div>

          <!-- Alert Error -->
          <div v-if="errorMessage" class="p-3 rounded-lg bg-rose-50 border border-rose-200 text-rose-700 text-xs flex items-center gap-2">
            <i class="pi pi-exclamation-circle text-rose-600 shrink-0" />
            <span>{{ errorMessage }}</span>
          </div>

          <!-- Form Fields -->
          <div class="flex flex-col gap-3.5 text-xs">
            <!-- 1. Nama Pelapor -->
            <div class="flex flex-col gap-1.5">
              <label class="text-xs font-semibold text-slate-700 uppercase tracking-wider flex items-center gap-1">
                <span>Nama Lengkap Pelapor</span>
                <span class="text-rose-500">*</span>
              </label>
              <InputText
                v-model="reporterName"
                placeholder="Contoh: Budi Santoso"
                class="w-full !bg-white !border-slate-300 !text-slate-900 !text-xs !py-2.5 focus:!border-indigo-600 focus:!ring-2 focus:!ring-indigo-100"
              />
            </div>

            <!-- 2. Kategori Laporan -->
            <div class="flex flex-col gap-1.5">
              <label class="text-xs font-semibold text-slate-700 uppercase tracking-wider flex items-center gap-1">
                <span>Jenis Kendala / Laporan</span>
                <span class="text-rose-500">*</span>
              </label>
              <Select
                v-model="reportType"
                :options="reportTypeOptions"
                optionLabel="label"
                optionValue="value"
                class="w-full !bg-white !border-slate-300 !text-slate-900 !text-xs"
              />
            </div>

            <!-- 3. Lokasi / Ruangan (wajib, dari database) -->
            <div class="flex flex-col gap-1.5">
              <label class="text-xs font-semibold text-slate-700 uppercase tracking-wider flex items-center gap-1">
                <span>Lokasi / Ruangan Saat Ini</span>
                <span class="text-rose-500">*</span>
              </label>
              <Select
                v-model="location"
                :options="locations"
                optionLabel="name"
                optionValue="name"
                :placeholder="locations.length ? 'Pilih ruangan / lokasi barang' : 'Lokasi dari database'"
                :filter="locations.length > 8"
                filterPlaceholder="Cari ruangan..."
                :editable="locations.length === 0"
                class="w-full !bg-white !border-slate-300 !text-slate-900 !text-xs"
                :class="{ '!border-rose-500': submitted && !location.trim() }"
              />
              <span v-if="submitted && !location.trim()" class="text-[11px] text-rose-500">
                Lokasi wajib dipilih
              </span>
            </div>

            <!-- 4. Uraian Masalah / Deskripsi -->
            <div class="flex flex-col gap-1.5">
              <label class="text-xs font-semibold text-slate-700 uppercase tracking-wider flex items-center gap-1">
                <span>Deskripsi Kerusakan / Kendala</span>
                <span class="text-rose-500">*</span>
              </label>
              <Textarea
                v-model="description"
                rows="4"
                placeholder="Jelaskan secara rinci kerusakan atau kendala yang terjadi, misalnya: AC mati total, lampu indikator merah berkedip, tidak mengeluarkan udara dingin..."
                class="w-full !bg-white !border-slate-300 !text-slate-900 !text-xs !py-2 focus:!border-indigo-600 focus:!ring-2 focus:!ring-indigo-100"
              />
            </div>

            <!-- 5. Foto Bukti Kerusakan -->
            <div class="flex flex-col gap-2">
              <label class="text-xs font-semibold text-slate-700 uppercase tracking-wider flex items-center justify-between">
                <span>Foto Bukti Kerusakan / Kondisi</span>
                <span class="text-[10px] text-slate-500 font-normal normal-case">Maks. 10 MB</span>
              </label>

              <!-- Upload Area -->
              <div
                v-if="!photoPreview"
                class="border-2 border-dashed border-slate-300 hover:border-indigo-500 bg-slate-50 hover:bg-slate-100/80 rounded-xl p-5 text-center flex flex-col items-center justify-center gap-2 cursor-pointer transition-colors"
                @click="($refs.photoInput as HTMLInputElement).click()"
              >
                <div class="w-10 h-10 rounded-full bg-indigo-50 flex items-center justify-center text-indigo-600">
                  <i class="pi pi-camera text-lg" />
                </div>
                <div class="text-xs font-semibold text-slate-700">Ambil Foto / Pilih Gambar</div>
                <div class="text-[10px] text-slate-500">Mendukung kamera HP langsung atau unggah file foto</div>
                <input
                  ref="photoInput"
                  type="file"
                  accept="image/*"
                  capture="environment"
                  class="hidden"
                  @change="onPhotoSelected"
                />
              </div>

              <!-- Preview Area -->
              <div v-else class="relative rounded-xl overflow-hidden border border-slate-200 bg-slate-100 max-h-56 flex items-center justify-center">
                <img :src="photoPreview" alt="Bukti Kerusakan" class="w-full max-h-56 object-contain" />
                <div v-if="uploadingPhoto" class="absolute inset-0 bg-slate-900/60 flex flex-col items-center justify-center gap-2 text-white text-xs">
                  <i class="pi pi-spin pi-spinner text-2xl" />
                  <span>Mengunggah foto...</span>
                </div>
                <button
                  v-if="!uploadingPhoto"
                  type="button"
                  @click="photoPreview = ''; photoUrl = ''"
                  class="absolute top-2 right-2 p-1.5 bg-rose-600 hover:bg-rose-700 text-white rounded-lg text-xs transition-colors shadow-sm"
                  title="Hapus Foto"
                >
                  <i class="pi pi-trash" />
                </button>
              </div>
            </div>

            <!-- Submit Button -->
            <Button
              type="button"
              label="Kirim Laporan Pengaduan"
              icon="pi pi-send"
              :loading="submitting || uploadingPhoto"
              class="w-full !py-3.5 !text-sm !font-bold !bg-indigo-600 hover:!bg-indigo-700 !text-white !border-0 !rounded-xl !mt-2 shadow-md shadow-indigo-600/20"
              @click="submitReport"
            />
          </div>
        </div>
      </div>

      <!-- ================= TAB 2: LACAK STATUS ADUAN ================= -->
      <div v-else class="flex flex-col gap-4">
        <div class="bg-white border border-slate-200 rounded-2xl p-5 shadow-sm flex flex-col gap-4">
          <div>
            <h3 class="text-sm font-bold text-slate-900 flex items-center gap-2">
              <i class="pi pi-search text-indigo-600" />
              <span>Cek Status Tindak Lanjut Pengaduan</span>
            </h3>
            <p class="text-[11px] text-slate-500 mt-0.5">
              Masukkan Nomor Tiket Pengaduan (contoh: LAP-20261019-XXXX) untuk melihat status penanganan oleh petugas.
            </p>
          </div>

          <div class="flex gap-2">
            <InputText
              v-model="trackQuery"
              placeholder="Masukkan nomor tiket..."
              class="flex-1 !bg-white !border-slate-300 !text-slate-900 !text-xs !py-2.5 focus:!border-indigo-600 focus:!ring-2 focus:!ring-indigo-100"
              @keyup.enter="fetchTracking"
            />
            <Button
              label="Cari"
              icon="pi pi-search"
              :loading="trackLoading"
              class="!bg-indigo-600 hover:!bg-indigo-700 !text-white !text-xs !font-bold !px-4 !border-0 !rounded-xl shadow-sm"
              @click="fetchTracking"
            />
          </div>

          <!-- Error Tracking -->
          <div v-if="trackError" class="p-3 rounded-lg bg-rose-50 border border-rose-200 text-rose-700 text-xs flex items-center gap-2">
            <i class="pi pi-exclamation-triangle text-rose-600 shrink-0" />
            <span>{{ trackError }}</span>
          </div>

          <!-- Result Tracking -->
          <div v-if="trackResult" class="border border-slate-200 rounded-xl bg-slate-50/70 p-4 flex flex-col gap-3">
            <div class="flex items-start justify-between gap-2 border-b border-slate-200 pb-3">
              <div>
                <div class="text-[10px] text-slate-500 uppercase font-semibold">Nomor Tiket</div>
                <div class="text-sm font-mono font-bold text-indigo-600">{{ trackResult.report_number }}</div>
                <div class="text-[11px] text-slate-500 mt-0.5">Dibuat: {{ formatDate(trackResult.created_at) }}</div>
              </div>
              <Tag :value="getStatusLabel(trackResult.status)" :severity="getStatusSeverity(trackResult.status)" class="!text-xs !py-1 !px-2.5" />
            </div>

            <!-- Detail Grid -->
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-2 text-xs">
              <div class="p-2.5 rounded-lg bg-white border border-slate-200">
                <span class="text-slate-500 text-[10px] block">Barang / Aset:</span>
                <span class="font-semibold text-slate-900">{{ trackResult.item_name || 'Aset Umum / Fasilitas' }}</span>
                <span v-if="trackResult.item_sku" class="text-slate-500 font-mono text-[11px] block">{{ trackResult.item_sku }}</span>
              </div>
              <div class="p-2.5 rounded-lg bg-white border border-slate-200">
                <span class="text-slate-500 text-[10px] block">Kategori Kendala:</span>
                <span class="font-semibold text-slate-900">{{ trackResult.report_type }}</span>
              </div>
              <div class="p-2.5 rounded-lg bg-white border border-slate-200">
                <span class="text-slate-500 text-[10px] block">Lokasi:</span>
                <span class="font-semibold text-slate-900">{{ trackResult.location || '—' }}</span>
              </div>
              <div class="p-2.5 rounded-lg bg-white border border-slate-200">
                <span class="text-slate-500 text-[10px] block">Pelapor:</span>
                <span class="font-semibold text-slate-900">{{ trackResult.reporter_name }}</span>
              </div>
            </div>

            <!-- Deskripsi -->
            <div class="p-3 rounded-lg bg-white border border-slate-200 text-xs">
              <span class="text-[10px] text-slate-500 uppercase font-bold block mb-1">Rincian Laporan:</span>
              <p class="text-slate-800 leading-relaxed">{{ trackResult.description }}</p>
            </div>

            <!-- Foto Lampiran jika ada -->
            <div v-if="trackResult.photo_url" class="rounded-lg overflow-hidden border border-slate-200 bg-white p-1">
              <img :src="trackResult.photo_url" alt="Foto Laporan" class="w-full max-h-48 object-contain rounded" />
            </div>

            <!-- Catatan Tindak Lanjut Petugas -->
            <div class="p-3 rounded-lg border flex flex-col gap-1.5"
                 :class="trackResult.status === 'RESOLVED' ? 'bg-emerald-50 border-emerald-200 text-emerald-900' : 'bg-indigo-50 border-indigo-200 text-indigo-900'">
              <div class="text-[10.5px] font-bold uppercase tracking-wide flex items-center justify-between">
                <span class="flex items-center gap-1.5">
                  <i class="pi pi-user-edit" />
                  <span>Tanggapan &amp; Tindak Lanjut Petugas:</span>
                </span>
                <span v-if="trackResult.handled_by" class="text-[10px] text-slate-500 font-normal">Oleh: {{ trackResult.handled_by }}</span>
              </div>
              <p class="text-xs leading-relaxed">
                {{ trackResult.admin_notes || 'Laporan telah diterima dan sedang menunggu giliran inspeksi / tindak lanjut oleh petugas pengelola barang.' }}
              </p>
              <div v-if="trackResult.handled_at" class="text-[10px] text-slate-500 text-right mt-1">
                Waktu pembaruan: {{ formatDate(trackResult.handled_at) }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- FOOTER -->
    <div class="w-full max-w-xl mx-auto text-center text-slate-500 text-[11px] mt-6">
      Sistem Informasi Inventaris &amp; Pengelolaan Aset Kantor &copy; 2026
    </div>

    <!-- DIALOG SUKSES KIRIM LAPORAN -->
    <Dialog
      v-model:visible="submitSuccess"
      modal
      header="Laporan Berhasil Terkirim"
      :style="{ width: '90vw', maxWidth: '440px' }"
      class="!bg-white !border-slate-200 !text-slate-900"
    >
      <div class="flex flex-col items-center text-center gap-3 p-2">
        <div class="w-14 h-14 rounded-full bg-emerald-50 text-emerald-600 border border-emerald-200 flex items-center justify-center text-2xl">
          <i class="pi pi-check" />
        </div>
        <div>
          <h4 class="text-base font-bold text-slate-900">Terima Kasih atas Laporan Anda!</h4>
          <p class="text-xs text-slate-500 mt-1">
            Laporan pengaduan telah dicatat ke sistem dan segera ditindaklanjuti oleh petugas pengelola barang.
          </p>
        </div>

        <div class="w-full p-3 rounded-xl bg-slate-50 border border-slate-200 flex flex-col items-center gap-1">
          <span class="text-[10px] text-slate-500 uppercase tracking-wider font-semibold">Nomor Tiket Anda:</span>
          <span class="text-lg font-mono font-black text-indigo-600 tracking-wider select-all">{{ createdTicket?.report_number }}</span>
          <Button
            label="Salin Nomor Tiket"
            icon="pi pi-copy"
            size="small"
            text
            class="!text-xs !py-1 !text-indigo-600 hover:!text-indigo-700"
            @click="copyTicketNumber"
          />
        </div>

        <div class="flex gap-2 w-full mt-2">
          <Button
            label="Tutup"
            severity="secondary"
            class="flex-1 !text-xs !py-2"
            @click="submitSuccess = false"
          />
          <Button
            label="Lacak Status"
            icon="pi pi-search"
            class="flex-1 !text-xs !py-2 !bg-indigo-600 hover:!bg-indigo-700 !text-white !border-0"
            @click="submitSuccess = false; activeTab = 'track'; trackQuery = createdTicket?.report_number; fetchTracking()"
          />
        </div>
      </div>
    </Dialog>
  </div>
</template>
