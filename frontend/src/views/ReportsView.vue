<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import api from '@/api'
import { useToast } from 'primevue/usetoast'
import { useAuthStore } from '@/stores/auth'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Textarea from 'primevue/textarea'
import Tag from 'primevue/tag'
import Dialog from 'primevue/dialog'
import PageHeader from '@/components/PageHeader.vue'
import StatCard from '@/components/StatCard.vue'
import EmptyState from '@/components/EmptyState.vue'

const toast = useToast()
const auth = useAuthStore()

const reports = ref<any[]>([])
const total = ref(0)
const loading = ref(true)
const page = ref(1)
const perPage = ref(20)

const statusFilter = ref('ALL')
const typeFilter = ref('ALL')
const searchQuery = ref('')

const stats = ref({
  total: 0,
  pending: 0,
  in_progress: 0,
  resolved: 0,
  rejected: 0,
})

const statusOptions = [
  { label: 'Semua Status', value: 'ALL' },
  { label: '⏳ Menunggu Verifikasi', value: 'PENDING' },
  { label: '🔧 Sedang Ditindaklanjuti', value: 'IN_PROGRESS' },
  { label: '✅ Selesai Ditangani', value: 'RESOLVED' },
  { label: '❌ Ditolak', value: 'REJECTED' },
]

const typeOptions = [
  { label: 'Semua Kategori Aduan', value: 'ALL' },
  { label: '🛠️ Kerusakan Alat / Barang', value: 'Kerusakan Alat / Barang' },
  { label: '🔍 Kehilangan / Hilang', value: 'Kehilangan / Tidak Ditemukan' },
  { label: '🚚 Perpindahan Ruangan / Lokasi', value: 'Perpindahan Ruangan / Lokasi' },
  { label: '🧹 Kondisi Fisik & Kebersihan', value: 'Kondisi Fisik & Kebersihan' },
  { label: '📋 Masalah Lainnya', value: 'Lainnya' },
]

// Modal Detail / Tindak Lanjut
const selectedReport = ref<any>(null)
const detailModalVisible = ref(false)
const updatingStatus = ref(false)
const targetStatus = ref('IN_PROGRESS')
const adminNotes = ref('')
const updateItemCondition = ref('')
const photoZoomModal = ref(false)
const zoomedPhoto = ref('')

const conditionOptions = [
  { label: 'Jangan ubah kondisi barang', value: '' },
  { label: 'Ubah jadi: Rusak Ringan (RR)', value: 'Rusak Ringan' },
  { label: 'Ubah jadi: Rusak Berat (RB)', value: 'Rusak Berat' },
  { label: 'Ubah jadi: Berfungsi Baik (B)', value: 'Berfungsi' },
]

async function fetchReports(resetPage = false) {
  if (resetPage) page.value = 1
  loading.value = true
  try {
    const res = await api.get('/reports', {
      params: {
        status: statusFilter.value,
        type: typeFilter.value,
        q: searchQuery.value,
        page: page.value,
        per_page: perPage.value,
      },
    })
    reports.value = res.data?.data || []
    total.value = res.data?.total || 0
    if (res.data?.stats) {
      stats.value = res.data.stats
    }
  } catch (err: any) {
    toast.add({
      severity: 'error',
      summary: 'Gagal memuat laporan',
      detail: err?.response?.data?.error || err.message,
      life: 3000,
    })
  } finally {
    loading.value = false
  }
}

let searchDebounce: any = null
watch(searchQuery, () => {
  clearTimeout(searchDebounce)
  searchDebounce = setTimeout(() => {
    fetchReports(true)
  }, 350)
})

watch([statusFilter, typeFilter], () => {
  fetchReports(true)
})

onMounted(() => {
  fetchReports()
})

function openDetail(r: any) {
  selectedReport.value = r
  targetStatus.value = r.status || 'IN_PROGRESS'
  adminNotes.value = r.admin_notes || ''
  updateItemCondition.value = ''
  detailModalVisible.value = true
}

async function saveFollowUp() {
  if (!selectedReport.value) return
  updatingStatus.value = true
  try {
    const payload: any = {
      status: targetStatus.value,
      admin_notes: adminNotes.value,
    }
    if (updateItemCondition.value) {
      payload.update_item_condition = updateItemCondition.value
    }

    await api.put(`/reports/${selectedReport.value.id}/status`, payload)
    toast.add({
      severity: 'success',
      summary: 'Tindak Lanjut Disimpan',
      detail: `Status tiket ${selectedReport.value.report_number} diperbarui menjadi ${targetStatus.value}`,
      life: 2500,
    })
    detailModalVisible.value = false
    fetchReports()
  } catch (err: any) {
    toast.add({
      severity: 'error',
      summary: 'Gagal memperbarui',
      detail: err?.response?.data?.error || err.message,
      life: 3500,
    })
  } finally {
    updatingStatus.value = false
  }
}

async function deleteReport(r: any) {
  if (!confirm(`Hapus laporan pengaduan ${r.report_number}? Tindakan ini tidak dapat dibatalkan.`)) return
  try {
    await api.delete(`/reports/${r.id}`)
    toast.add({ severity: 'success', summary: 'Laporan Dihapus', life: 2000 })
    fetchReports()
  } catch (err: any) {
    toast.add({ severity: 'error', summary: 'Gagal menghapus', detail: err?.response?.data?.error, life: 3000 })
  }
}

function copyPublicLink() {
  const url = `${window.location.origin}/lapor`
  navigator.clipboard.writeText(url)
  toast.add({
    severity: 'info',
    summary: 'Link Publik Disalin',
    detail: 'Link formulir pengaduan siap dibagikan ke publik atau ditempel pada QR.',
    life: 2500,
  })
}

function getWhatsAppUrl(phone?: string, ticketNo?: string) {
  if (!phone) return '#'
  let cleaned = phone.replace(/[^0-9]/g, '')
  if (cleaned.startsWith('0')) {
    cleaned = '62' + cleaned.slice(1)
  }
  const text = encodeURIComponent(`Halo, menindaklanjuti laporan Anda di sistem inventaris dengan nomor tiket *${ticketNo || ''}*...`)
  return `https://wa.me/${cleaned}?text=${text}`
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
    case 'PENDING': return '⏳ Menunggu'
    case 'IN_PROGRESS': return '🔧 Diproses'
    case 'RESOLVED': return '✅ Selesai'
    case 'REJECTED': return '❌ Ditolak'
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
  <div class="shell max-w-7xl mx-auto space-y-6">
    <!-- PAGE HEADER -->
    <PageHeader
      title="Pengaduan &amp; Pelaporan Aset"
      subtitle="Kelola laporan kerusakan, kehilangan, dan pengaduan aset publik secara terintegrasi"
    >
      <template #actions>
        <div class="flex items-center gap-2">
          <Button
            label="Salin Link Publik"
            icon="pi pi-link"
            size="small"
            severity="secondary"
            outlined
            class="!text-xs"
            @click="copyPublicLink"
          />
          <a :href="'/lapor'" target="_blank" rel="noopener noreferrer">
            <Button
              label="Buka Form Publik"
              icon="pi pi-external-link"
              size="small"
              class="!text-xs !bg-indigo-600 hover:!bg-indigo-500 !border-0"
            />
          </a>
        </div>
      </template>
    </PageHeader>

    <!-- STATS ROW -->
    <div class="grid grid-cols-2 md:grid-cols-5 gap-3">
      <StatCard
        title="Total Laporan"
        :value="stats.total"
        icon="pi pi-inbox"
        color="indigo"
      />
      <StatCard
        title="Menunggu Verifikasi"
        :value="stats.pending"
        icon="pi pi-clock"
        color="amber"
      />
      <StatCard
        title="Sedang Ditangani"
        :value="stats.in_progress"
        icon="pi pi-cog"
        color="sky"
      />
      <StatCard
        title="Selesai Diperbaiki"
        :value="stats.resolved"
        icon="pi pi-check-circle"
        color="emerald"
      />
      <StatCard
        title="Ditolak / Batal"
        :value="stats.rejected"
        icon="pi pi-times-circle"
        color="rose"
      />
    </div>

    <!-- FILTER TOOLBAR -->
    <div class="panel shell-panel p-3.5 flex flex-wrap items-center justify-between gap-3 text-xs">
      <div class="flex items-center gap-2 flex-1 min-w-[280px]">
        <div class="relative flex-1">
          <i class="pi pi-search absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 text-xs" />
          <InputText
            v-model="searchQuery"
            placeholder="Cari no. tiket, nama barang, SKU, atau nama pelapor..."
            class="w-full !pl-8 !py-1.5 !text-xs !bg-transparent"
          />
        </div>
      </div>

      <div class="flex items-center gap-2 flex-wrap">
        <Select
          v-model="statusFilter"
          :options="statusOptions"
          optionLabel="label"
          optionValue="value"
          class="!text-xs !min-w-[160px]"
        />

        <Select
          v-model="typeFilter"
          :options="typeOptions"
          optionLabel="label"
          optionValue="value"
          class="!text-xs !min-w-[180px]"
        />

        <Button
          icon="pi pi-refresh"
          size="small"
          text
          rounded
          :loading="loading"
          v-tooltip.top="'Segarkan Data'"
          @click="fetchReports(false)"
        />
      </div>
    </div>

    <!-- DATA TABLE -->
    <div class="panel shell-panel overflow-hidden">
      <div v-if="loading" class="p-12 text-center text-slate-400">
        <i class="pi pi-spin pi-spinner text-2xl text-indigo-500 mb-2" />
        <p class="text-xs">Memuat data laporan pengaduan...</p>
      </div>

      <div v-else-if="reports.length === 0" class="p-10">
        <EmptyState
          icon="pi pi-inbox"
          title="Tidak Ada Laporan Pengaduan"
          description="Belum ada laporan pengaduan yang sesuai dengan filter pencarian saat ini."
        />
      </div>

      <div v-else class="overflow-x-auto">
        <table class="w-full text-left text-xs border-collapse">
          <thead>
            <tr class="border-b border-line bg-slate-500/5 text-slate-400 font-bold uppercase tracking-wider text-[10px]">
              <th class="py-3 px-3.5">Tiket &amp; Tanggal</th>
              <th class="py-3 px-3.5">Barang / Aset</th>
              <th class="py-3 px-3.5">Pelapor &amp; Kontak</th>
              <th class="py-3 px-3.5">Jenis Aduan</th>
              <th class="py-3 px-3.5">Status</th>
              <th class="py-3 px-3.5 text-center">Foto</th>
              <th class="py-3 px-3.5 text-right">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-line">
            <tr
              v-for="r in reports"
              :key="r.id"
              class="hover:bg-slate-500/5 transition-colors cursor-pointer"
              @click="openDetail(r)"
            >
              <!-- Tiket & Waktu -->
              <td class="py-3 px-3.5">
                <div class="font-mono font-bold text-indigo-400 text-xs">{{ r.report_number }}</div>
                <div class="text-[10.5px] text-slate-400 mt-0.5">{{ formatDate(r.created_at) }}</div>
              </td>

              <!-- Aset / Barang -->
              <td class="py-3 px-3.5 max-w-[200px]">
                <div class="font-bold text-slate-200 truncate" :title="r.item_name || 'Aset Umum'">
                  {{ r.item_name || 'Aset Umum / Fasilitas' }}
                </div>
                <div v-if="r.item_sku" class="font-mono text-[10.5px] text-slate-400">{{ r.item_sku }}</div>
                <div v-if="r.location" class="text-[10px] text-slate-400 flex items-center gap-1 mt-0.5">
                  <i class="pi pi-map-marker text-[9px] text-emerald-400" />
                  <span class="truncate">{{ r.location }}</span>
                </div>
              </td>

              <!-- Pelapor & Kontak -->
              <td class="py-3 px-3.5" @click.stop>
                <div class="font-semibold text-slate-200">{{ r.reporter_name }}</div>
                <div class="flex items-center gap-1.5 mt-0.5">
                  <span class="text-[11px] text-slate-400">{{ r.reporter_contact }}</span>
                  <a
                    :href="getWhatsAppUrl(r.reporter_contact, r.report_number)"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="text-emerald-400 hover:text-emerald-300 p-0.5"
                    title="Chat via WhatsApp"
                  >
                    <i class="pi pi-whatsapp text-xs" />
                  </a>
                </div>
              </td>

              <!-- Jenis Aduan -->
              <td class="py-3 px-3.5">
                <span class="font-medium text-slate-300">{{ r.report_type }}</span>
                <p class="text-[10.5px] text-slate-400 truncate max-w-[180px] mt-0.5" :title="r.description">
                  {{ r.description }}
                </p>
              </td>

              <!-- Status Tag -->
              <td class="py-3 px-3.5">
                <Tag
                  :value="getStatusLabel(r.status)"
                  :severity="getStatusSeverity(r.status)"
                  class="!text-[10.5px] !py-0.5 !px-2"
                />
                <div v-if="r.handled_by" class="text-[9.5px] text-slate-400 mt-1">
                  Oleh: {{ r.handled_by }}
                </div>
              </td>

              <!-- Foto Thumbnail -->
              <td class="py-3 px-3.5 text-center" @click.stop>
                <img
                  v-if="r.photo_url"
                  :src="r.photo_url"
                  alt="Bukti"
                  class="w-10 h-10 object-cover rounded border border-line mx-auto cursor-pointer hover:scale-105 transition-transform"
                  @click="zoomedPhoto = r.photo_url; photoZoomModal = true"
                />
                <span v-else class="text-slate-500 text-[10px]">—</span>
              </td>

              <!-- Aksi -->
              <td class="py-3 px-3.5 text-right whitespace-nowrap" @click.stop>
                <div class="flex items-center justify-end gap-1">
                  <Button
                    icon="pi pi-pencil"
                    label="Tindak Lanjut"
                    size="small"
                    class="!text-xs !py-1 !px-2.5 !bg-indigo-600 hover:!bg-indigo-500 !border-0"
                    @click="openDetail(r)"
                  />
                  <Button
                    v-if="auth.isAdmin"
                    icon="pi pi-trash"
                    size="small"
                    severity="danger"
                    text
                    class="!p-1 text-rose-400 hover:text-rose-300"
                    v-tooltip.top="'Hapus Laporan'"
                    @click="deleteReport(r)"
                  />
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- PAGINATION -->
      <div v-if="total > perPage" class="p-3 border-t border-line flex items-center justify-between text-xs text-slate-400">
        <div>Menampilkan {{ reports.length }} dari {{ total }} laporan</div>
        <div class="flex items-center gap-1">
          <Button
            label="Sebelumnya"
            icon="pi pi-chevron-left"
            size="small"
            text
            :disabled="page <= 1"
            class="!text-xs"
            @click="page--; fetchReports(false)"
          />
          <span class="px-2 font-bold text-slate-300">Hal. {{ page }}</span>
          <Button
            label="Berikutnya"
            icon="pi pi-chevron-right"
            iconPos="right"
            size="small"
            text
            :disabled="page * perPage >= total"
            class="!text-xs"
            @click="page++; fetchReports(false)"
          />
        </div>
      </div>
    </div>

    <!-- MODAL TINDAK LANJUT & DETAIL LAPORAN -->
    <Dialog
      v-model:visible="detailModalVisible"
      modal
      :header="`Tindak Lanjut Pengaduan ${selectedReport?.report_number || ''}`"
      :style="{ width: '92vw', maxWidth: '640px' }"
    >
      <div v-if="selectedReport" class="flex flex-col gap-4 text-xs pt-1">
        <!-- Status Banner -->
        <div class="p-3 rounded-xl bg-slate-500/10 border border-line flex items-center justify-between">
          <div>
            <span class="text-[10px] text-slate-400 uppercase font-semibold">Status Saat Ini:</span>
            <div class="font-bold text-sm text-slate-200 mt-0.5">{{ getStatusLabel(selectedReport.status) }}</div>
          </div>
          <div class="text-right">
            <span class="text-[10px] text-slate-400 uppercase font-semibold">Waktu Masuk:</span>
            <div class="text-slate-300 text-xs mt-0.5">{{ formatDate(selectedReport.created_at) }}</div>
          </div>
        </div>

        <!-- Info Barang & Pelapor -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <div class="p-3 rounded-xl bg-slate-500/5 border border-line space-y-1">
            <span class="text-[10px] font-bold text-indigo-400 uppercase tracking-wider block">Barang / Aset Terkait</span>
            <div class="font-bold text-sm text-slate-100">{{ selectedReport.item_name || 'Aset Umum / Fasilitas' }}</div>
            <div v-if="selectedReport.item_sku" class="font-mono text-xs text-slate-400">SKU: {{ selectedReport.item_sku }}</div>
            <div v-if="selectedReport.location" class="text-xs text-slate-400 flex items-center gap-1">
              <i class="pi pi-map-marker text-emerald-400 text-[10px]" />
              <span>{{ selectedReport.location }}</span>
            </div>
          </div>

          <div class="p-3 rounded-xl bg-slate-500/5 border border-line space-y-1">
            <span class="text-[10px] font-bold text-indigo-400 uppercase tracking-wider block">Data Pelapor</span>
            <div class="font-bold text-sm text-slate-100">{{ selectedReport.reporter_name }}</div>
            <div class="text-xs text-slate-400 flex items-center gap-2">
              <span>{{ selectedReport.reporter_contact }}</span>
              <a
                :href="getWhatsAppUrl(selectedReport.reporter_contact, selectedReport.report_number)"
                target="_blank"
                rel="noopener noreferrer"
                class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded bg-emerald-600/20 text-emerald-400 hover:bg-emerald-600/30 font-semibold text-[10px]"
              >
                <i class="pi pi-whatsapp" /> Chat WA
              </a>
            </div>
            <div class="text-xs text-slate-400">Kategori: <b>{{ selectedReport.report_type }}</b></div>
          </div>
        </div>

        <!-- Uraian Keluhan & Bukti Foto -->
        <div class="p-3 rounded-xl bg-slate-500/5 border border-line space-y-2">
          <span class="text-[10px] font-bold text-slate-400 uppercase tracking-wider block">Deskripsi Masalah dari Pelapor:</span>
          <p class="text-slate-200 leading-relaxed whitespace-pre-wrap">{{ selectedReport.description }}</p>

          <div v-if="selectedReport.photo_url" class="pt-2 border-t border-line">
            <span class="text-[10px] font-bold text-slate-400 uppercase tracking-wider block mb-1.5">Foto Lampiran Bukti:</span>
            <img
              :src="selectedReport.photo_url"
              alt="Bukti Kerusakan"
              class="max-h-56 rounded-lg border border-line object-contain bg-black/30 cursor-pointer hover:opacity-90 transition-opacity"
              @click="zoomedPhoto = selectedReport.photo_url; photoZoomModal = true"
            />
          </div>
        </div>

        <!-- FORM UPDATE TINDAK LANJUT -->
        <div class="p-4 rounded-xl bg-indigo-500/5 border border-indigo-500/20 space-y-3">
          <span class="text-xs font-bold text-indigo-400 uppercase tracking-wider flex items-center gap-1.5">
            <i class="pi pi-user-edit" />
            <span>Formulir Tindak Lanjut Petugas</span>
          </span>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div class="space-y-1">
              <label class="font-bold text-slate-300">Ubah Status Tiket</label>
              <Select
                v-model="targetStatus"
                :options="[
                  { label: '⏳ Menunggu Verifikasi', value: 'PENDING' },
                  { label: '🔧 Sedang Ditindaklanjuti', value: 'IN_PROGRESS' },
                  { label: '✅ Selesai Ditangani', value: 'RESOLVED' },
                  { label: '❌ Ditolak / Tidak Valid', value: 'REJECTED' },
                ]"
                optionLabel="label"
                optionValue="value"
                class="w-full !text-xs"
              />
            </div>

            <div class="space-y-1">
              <label class="font-bold text-slate-300">Perbarui Kondisi Aset di Master Data</label>
              <Select
                v-model="updateItemCondition"
                :options="conditionOptions"
                optionLabel="label"
                optionValue="value"
                class="w-full !text-xs"
                :disabled="!selectedReport.item_id"
              />
            </div>
          </div>

          <div class="space-y-1">
            <label class="font-bold text-slate-300">Catatan Tindakan / Umpan Balik Petugas</label>
            <Textarea
              v-model="adminNotes"
              rows="3"
              placeholder="Tuliskan tindakan yang diambil, jadwal teknisi, atau keterangan penyelesaian masalah (dapat dilihat pelapor saat melacak nomor tiket)..."
              class="w-full !text-xs"
            />
          </div>
        </div>

        <!-- Action Buttons -->
        <div class="flex items-center justify-end gap-2 pt-2 border-t border-line">
          <Button
            label="Tutup"
            severity="secondary"
            class="!text-xs"
            @click="detailModalVisible = false"
          />
          <Button
            label="Simpan Tindak Lanjut"
            icon="pi pi-check"
            :loading="updatingStatus"
            class="!text-xs !bg-indigo-600 hover:!bg-indigo-500 !border-0"
            @click="saveFollowUp"
          />
        </div>
      </div>
    </Dialog>

    <!-- PHOTO ZOOM MODAL -->
    <Dialog
      v-model:visible="photoZoomModal"
      modal
      header="Foto Bukti Kerusakan"
      :style="{ width: '90vw', maxWidth: '800px' }"
    >
      <div class="flex items-center justify-center p-2 bg-black/60 rounded-xl overflow-hidden">
        <img :src="zoomedPhoto" alt="Zoom Foto" class="max-h-[75vh] w-auto object-contain rounded" />
      </div>
    </Dialog>
  </div>
</template>
