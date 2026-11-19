<script setup lang="ts">
import { ref, onMounted } from 'vue'
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

const route = useRoute()
const router = useRouter()
const theme = useThemeStore()

const activeTab = ref<'create' | 'track'>('create')

const reporterName = ref('')
const reportType = ref('Kerusakan Alat / Barang')
const description = ref('')
const location = ref('')
const locations = ref<{ id: string; name: string; code: string }[]>([])
const photoUrl = ref('')
const uploadingPhoto = ref(false)
const photoPreview = ref('')
const photoError = ref('')

const reportTypeOptions = [
  { label: 'Kerusakan Alat / Barang', value: 'Kerusakan Alat / Barang' },
  { label: 'Kehilangan / Tidak Ditemukan', value: 'Kehilangan / Tidak Ditemukan' },
  { label: 'Perpindahan Ruangan / Lokasi', value: 'Perpindahan Ruangan / Lokasi' },
  { label: 'Kondisi Fisik & Kebersihan', value: 'Kondisi Fisik & Kebersihan' },
  { label: 'Masalah Lainnya', value: 'Lainnya' },
]

const item = ref<any>(null)
const loadingItem = ref(false)

const trackQuery = ref('')
const trackResult = ref<any>(null)
const trackLoading = ref(false)
const trackError = ref('')

const submitting = ref(false)
const submitSuccess = ref(false)
const createdTicket = ref<any>(null)
const copiedTicket = ref(false)
const errorMessage = ref('')
const submitted = ref(false)

onMounted(async () => {
  theme.apply()
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

  photoError.value = ''
  const file = input.files[0]
  if (file.size > 10 * 1024 * 1024) {
    photoError.value = 'Ukuran foto terlalu besar. Maksimal 10 MB.'
    return
  }

  photoPreview.value = URL.createObjectURL(file)

  const formData = new FormData()
  formData.append('photo', file)

  uploadingPhoto.value = true
  try {
    const res = await api.post('/public/upload', formData)
    photoUrl.value = res.data.url
  } catch (err: any) {
    photoError.value = err?.response?.data?.error || 'Gagal mengunggah foto. Pastikan format gambar JPG, PNG, atau WebP.'
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
    copiedTicket.value = false
    submitSuccess.value = true

    reporterName.value = ''
    description.value = ''
    photoUrl.value = ''
    photoPreview.value = ''
    photoError.value = ''
    submitted.value = false
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

async function copyTicketNumber() {
  if (!createdTicket.value?.report_number) return
  try {
    await navigator.clipboard.writeText(createdTicket.value.report_number)
    copiedTicket.value = true
    setTimeout(() => { copiedTicket.value = false }, 2500)
  } catch {
    copiedTicket.value = false
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
  <div class="min-h-dvh shell flex flex-col justify-between py-6 px-4 sm:px-6">
    <div class="w-full max-w-xl mx-auto flex flex-col gap-5">
      <!-- Header Brand -->
      <header class="flex items-center justify-between border-b border-line pb-4 gap-3">
        <div class="flex items-center gap-3 min-w-0">
          <AppLogo size="md" />
          <span class="hidden sm:inline-block h-5 w-px bg-line" aria-hidden="true" />
          <div class="hidden sm:block min-w-0">
            <div class="text-[12.5px] font-bold text-txt truncate">Layanan Pengaduan Aset</div>
            <div class="text-[11px] text-txt-dim truncate">Portal Pelaporan Kerusakan &amp; Kendala</div>
          </div>
        </div>

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
            severity="secondary"
            outlined
            @click="router.push('/login')"
          />
        </div>
      </header>

      <!-- Tab Switcher -->
      <div class="grid grid-cols-2 gap-1.5 p-1 rounded-[9px] bg-panel-2 border border-line">
        <button
          type="button"
          class="py-2 px-3 rounded-[7px] text-[12.5px] font-semibold transition-colors flex items-center justify-center gap-2 cursor-pointer"
          :class="activeTab === 'create'
            ? 'bg-panel text-acc-500 border border-[var(--acc-border)] shadow-[var(--shadow-sm)]'
            : 'text-txt-dim hover:text-txt border border-transparent'"
          @click="activeTab = 'create'"
        >
          <i class="pi pi-megaphone text-[12px]" aria-hidden="true" />
          <span>Buat Laporan Baru</span>
        </button>
        <button
          type="button"
          class="py-2 px-3 rounded-[7px] text-[12.5px] font-semibold transition-colors flex items-center justify-center gap-2 cursor-pointer"
          :class="activeTab === 'track'
            ? 'bg-panel text-acc-500 border border-[var(--acc-border)] shadow-[var(--shadow-sm)]'
            : 'text-txt-dim hover:text-txt border border-transparent'"
          @click="activeTab = 'track'"
        >
          <i class="pi pi-search text-[12px]" aria-hidden="true" />
          <span>Lacak Status Tiket</span>
        </button>
      </div>

      <!-- TAB 1: BUAT LAPORAN -->
      <div v-if="activeTab === 'create'" class="flex flex-col gap-4">
        <div v-if="loadingItem" class="panel p-4 space-y-2">
          <div class="skeleton h-3.5 w-32" />
          <div class="skeleton h-5 w-56" />
        </div>

        <div v-else-if="item" class="panel p-4 flex items-center gap-3.5 border-[var(--acc-border)] bg-[var(--acc-wash)]">
          <div class="w-12 h-12 rounded-[8px] bg-panel border border-line overflow-hidden shrink-0 grid place-items-center">
            <img v-if="item.photo_url" :src="item.photo_url" :alt="item.name" class="w-full h-full object-cover" />
            <i v-else class="pi pi-box text-acc-500 text-lg" aria-hidden="true" />
          </div>
          <div class="min-w-0 flex-1">
            <div class="t-mono !text-[10px] uppercase tracking-[0.06em] text-acc-500 font-semibold">Barang Terdeteksi dari QR</div>
            <div class="text-[14px] font-bold text-txt truncate">{{ item.name }}</div>
            <div class="flex items-center gap-2 mt-0.5 text-[11.5px] text-txt-dim flex-wrap">
              <span class="t-mono px-1.5 py-0.5 rounded border border-line bg-panel">{{ item.sku }}</span>
              <span v-if="item.location">· {{ item.location }}</span>
            </div>
          </div>
        </div>

        <div class="panel p-5 sm:p-6 flex flex-col gap-4">
          <div>
            <h1 class="text-[17px] font-bold tracking-[-0.02em] text-txt">Formulir Pengaduan &amp; Kerusakan</h1>
            <p class="text-[12.5px] text-txt-dim mt-0.5 max-w-[65ch]">
              Laporkan kerusakan alat, kehilangan, atau kendala fasilitas agar segera ditindaklanjuti petugas.
            </p>
          </div>

          <div
            v-if="errorMessage"
            role="alert"
            class="p-3 rounded-[8px] bg-[var(--sig-bad-bg)] border border-[var(--sig-bad-line)] text-sig-bad text-[12.5px] flex items-center gap-2"
          >
            <i class="pi pi-exclamation-circle shrink-0" aria-hidden="true" />
            <span>{{ errorMessage }}</span>
          </div>

          <form @submit.prevent="submitReport" class="flex flex-col gap-4">
            <label class="flex flex-col gap-1.5">
              <span class="text-[12px] font-semibold text-txt">Nama Lengkap Pelapor <span class="text-sig-bad">*</span></span>
              <InputText
                v-model="reporterName"
                placeholder="Contoh: Budi Santoso (Poli Umum)"
                class="w-full"
                :invalid="submitted && !reporterName.trim()"
              />
            </label>

            <label class="flex flex-col gap-1.5">
              <span class="text-[12px] font-semibold text-txt">Kategori Masalah <span class="text-sig-bad">*</span></span>
              <Select
                v-model="reportType"
                :options="reportTypeOptions"
                optionLabel="label"
                optionValue="value"
                class="w-full"
              />
            </label>

            <label class="flex flex-col gap-1.5">
              <span class="text-[12px] font-semibold text-txt">Lokasi / Ruangan Barang <span class="text-sig-bad">*</span></span>
              <Select
                v-if="locations.length > 0"
                v-model="location"
                :options="locations"
                optionLabel="name"
                optionValue="name"
                placeholder="Pilih ruangan / lokasi barang…"
                filter
                editable
                class="w-full"
                :invalid="submitted && !location.trim()"
              />
              <InputText
                v-else
                v-model="location"
                placeholder="Contoh: Ruang Poli Gigi / Gudang Farmasi"
                class="w-full"
                :invalid="submitted && !location.trim()"
              />
            </label>

            <label class="flex flex-col gap-1.5">
              <span class="text-[12px] font-semibold text-txt">Uraian Masalah / Kerusakan <span class="text-sig-bad">*</span></span>
              <Textarea
                v-model="description"
                rows="4"
                placeholder="Jelaskan kondisi kerusakan, gejala masalah, atau detail kendala yang ditemukan…"
                class="w-full"
                :invalid="submitted && !description.trim()"
              />
            </label>

            <div class="flex flex-col gap-1.5">
              <span class="text-[12px] font-semibold text-txt">Foto Bukti Kondisi (Opsional)</span>
              <div
                v-if="photoError"
                role="alert"
                class="p-2.5 rounded-[7px] bg-[var(--sig-bad-bg)] border border-[var(--sig-bad-line)] text-sig-bad text-[12px] flex items-center gap-2"
              >
                <i class="pi pi-exclamation-circle shrink-0" aria-hidden="true" />
                <span>{{ photoError }}</span>
              </div>

              <div class="flex items-center gap-3">
                <label class="flex-1 cursor-pointer">
                  <div class="border border-dashed border-line hover:border-[var(--acc-border)] rounded-[8px] p-4 text-center transition-colors bg-panel-2">
                    <i v-if="uploadingPhoto" class="pi pi-spin pi-spinner text-acc-500 text-lg mb-1 block" aria-hidden="true" />
                    <i v-else class="pi pi-camera text-txt-dim text-lg mb-1 block" aria-hidden="true" />
                    <span class="text-[12px] text-txt-dim block">
                      {{ uploadingPhoto ? 'Mengunggah foto…' : 'Klik untuk ambil foto atau pilih gambar (maks. 10 MB)' }}
                    </span>
                  </div>
                  <input type="file" accept="image/*" capture="environment" class="hidden" @change="onPhotoSelected" />
                </label>

                <div v-if="photoPreview" class="relative w-20 h-20 rounded-[8px] overflow-hidden border border-line shrink-0 bg-panel-2">
                  <img :src="photoPreview" alt="Pratinjau foto bukti laporan" class="w-full h-full object-cover" />
                  <button
                    type="button"
                    aria-label="Hapus foto lampiran"
                    class="absolute top-1 right-1 w-5 h-5 rounded-full bg-ink-950/80 text-white flex items-center justify-center text-[10px] cursor-pointer"
                    @click="photoPreview = ''; photoUrl = ''; photoError = ''"
                  >
                    <i class="pi pi-times" aria-hidden="true" />
                  </button>
                </div>
              </div>
            </div>

            <Button
              type="submit"
              label="Kirim Laporan Pengaduan"
              icon="pi pi-send"
              :loading="submitting"
              class="w-full mt-1"
            />
          </form>
        </div>
      </div>

      <!-- TAB 2: LACAK STATUS TIKET -->
      <div v-else class="flex flex-col gap-4">
        <div class="panel p-5 sm:p-6 flex flex-col gap-4">
          <div>
            <h1 class="text-[17px] font-bold tracking-[-0.02em] text-txt">Lacak Status Pengaduan</h1>
            <p class="text-[12.5px] text-txt-dim mt-0.5">
              Masukkan nomor tiket laporan Anda (contoh: <span class="t-mono text-acc-500 font-semibold">LPR-202604-0001</span>).
            </p>
          </div>

          <div class="flex gap-2">
            <InputText
              v-model="trackQuery"
              placeholder="Masukkan nomor tiket…"
              class="flex-1 t-mono"
              @keyup.enter="fetchTracking"
            />
            <Button
              label="Cek Status"
              icon="pi pi-search"
              :loading="trackLoading"
              @click="fetchTracking"
            />
          </div>

          <div
            v-if="trackError"
            role="alert"
            class="p-3 rounded-[8px] bg-[var(--sig-bad-bg)] border border-[var(--sig-bad-line)] text-sig-bad text-[12.5px] flex items-center gap-2"
          >
            <i class="pi pi-exclamation-triangle shrink-0" aria-hidden="true" />
            <span>{{ trackError }}</span>
          </div>

          <div v-if="trackResult" class="border border-line rounded-[10px] bg-panel-2 p-4 flex flex-col gap-3">
            <div class="flex items-start justify-between gap-2 border-b border-line pb-3">
              <div>
                <div class="text-[10.5px] text-txt-dim uppercase tracking-[0.05em] font-semibold">Nomor Tiket</div>
                <div class="text-[15px] t-mono font-bold text-acc-500">{{ trackResult.report_number }}</div>
                <div class="t-mono !text-[11px] text-txt-dim mt-0.5">Dibuat: {{ formatDate(trackResult.created_at) }}</div>
              </div>
              <StatusChip :kind="trackResult.status" />
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-2 text-[12.5px]">
              <div class="p-2.5 rounded-[7px] bg-panel border border-line">
                <span class="text-txt-dim text-[10.5px] block">Barang / Aset</span>
                <span class="font-semibold text-txt">{{ trackResult.item_name || 'Aset Umum / Fasilitas' }}</span>
                <span v-if="trackResult.item_sku" class="text-txt-dim t-mono !text-[11px] block">{{ trackResult.item_sku }}</span>
              </div>
              <div class="p-2.5 rounded-[7px] bg-panel border border-line">
                <span class="text-txt-dim text-[10.5px] block">Kategori Kendala</span>
                <span class="font-semibold text-txt">{{ trackResult.report_type }}</span>
              </div>
              <div class="p-2.5 rounded-[7px] bg-panel border border-line">
                <span class="text-txt-dim text-[10.5px] block">Lokasi</span>
                <span class="font-semibold text-txt">{{ trackResult.location || '—' }}</span>
              </div>
              <div class="p-2.5 rounded-[7px] bg-panel border border-line">
                <span class="text-txt-dim text-[10.5px] block">Pelapor</span>
                <span class="font-semibold text-txt">{{ trackResult.reporter_name }}</span>
              </div>
            </div>

            <div class="p-3 rounded-[7px] bg-panel border border-line text-[12.5px]">
              <span class="text-[10.5px] text-txt-dim uppercase tracking-[0.05em] font-semibold block mb-1">Rincian Laporan</span>
              <p class="text-txt leading-relaxed max-w-[65ch]">{{ trackResult.description }}</p>
            </div>

            <div v-if="trackResult.photo_url" class="rounded-[7px] overflow-hidden border border-line bg-panel p-1">
              <img :src="trackResult.photo_url" alt="Foto lampiran laporan" class="w-full max-h-48 object-contain rounded" />
            </div>

            <div
              class="p-3 rounded-[7px] border flex flex-col gap-1.5"
              :class="trackResult.status === 'RESOLVED'
                ? 'bg-[var(--sig-ok-bg)] border-[var(--sig-ok-line)] text-txt'
                : 'bg-[var(--acc-wash)] border-[var(--acc-border)] text-txt'"
            >
              <div class="text-[10.5px] font-bold uppercase tracking-[0.05em] flex items-center justify-between">
                <span class="flex items-center gap-1.5" :class="trackResult.status === 'RESOLVED' ? 'text-sig-ok' : 'text-acc-500'">
                  <i class="pi pi-user-edit" aria-hidden="true" />
                  <span>Tanggapan &amp; Tindak Lanjut Petugas</span>
                </span>
                <span v-if="trackResult.handled_by" class="text-[10.5px] text-txt-dim font-normal">Oleh: {{ trackResult.handled_by }}</span>
              </div>
              <p class="text-[12.5px] leading-relaxed">
                {{ trackResult.admin_notes || 'Laporan telah diterima dan sedang menunggu giliran inspeksi oleh petugas pengelola barang.' }}
              </p>
              <div v-if="trackResult.handled_at" class="t-mono !text-[10.5px] text-txt-dim text-right mt-1">
                Pembaruan: {{ formatDate(trackResult.handled_at) }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <footer class="w-full max-w-xl mx-auto text-center text-txt-dim text-[11.5px] mt-6">
      Sistem Informasi Inventaris &amp; Pengelolaan Aset Puskesmas &copy; 2026
    </footer>

    <!-- Dialog Sukses Kirim Laporan -->
    <Dialog
      v-model:visible="submitSuccess"
      modal
      header="Laporan Berhasil Terkirim"
      :style="{ width: '90vw', maxWidth: '440px' }"
    >
      <div class="flex flex-col items-center text-center gap-3 p-2">
        <div class="w-12 h-12 rounded-full bg-[var(--sig-ok-bg)] text-sig-ok border border-[var(--sig-ok-line)] flex items-center justify-center text-xl">
          <i class="pi pi-check" aria-hidden="true" />
        </div>
        <div>
          <h4 class="text-[15px] font-bold text-txt">Laporan Anda telah dicatat</h4>
          <p class="text-[12.5px] text-txt-dim mt-1 leading-relaxed">
            Simpan nomor tiket di bawah ini untuk memantau perkembangan tindak lanjut oleh petugas.
          </p>
        </div>

        <div class="w-full p-3.5 rounded-[8px] bg-panel-2 border border-line flex flex-col items-center gap-1.5">
          <span class="text-[10.5px] text-txt-dim uppercase tracking-[0.06em] font-semibold">Nomor Tiket Anda</span>
          <span class="text-[18px] t-mono font-bold text-acc-500 tracking-wider select-all">{{ createdTicket?.report_number }}</span>
          <Button
            :label="copiedTicket ? 'Nomor Tiket Disalin' : 'Salin Nomor Tiket'"
            :icon="copiedTicket ? 'pi pi-check' : 'pi pi-copy'"
            size="small"
            text
            :severity="copiedTicket ? 'success' : 'primary'"
            @click="copyTicketNumber"
          />
        </div>

        <div class="flex gap-2 w-full mt-2">
          <Button label="Tutup" size="small" severity="secondary" outlined class="flex-1" @click="submitSuccess = false" />
          <Button
            label="Lacak Status"
            icon="pi pi-search"
            size="small"
            class="flex-1"
            @click="submitSuccess = false; activeTab = 'track'; trackQuery = createdTicket?.report_number; fetchTracking()"
          />
        </div>
      </div>
    </Dialog>
  </div>
</template>
