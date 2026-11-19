<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import api from '@/api'
import { useToast } from 'primevue/usetoast'
import { useConfirm } from 'primevue/useconfirm'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Textarea from 'primevue/textarea'
import Dialog from 'primevue/dialog'
import PageHeader from '@/components/PageHeader.vue'
import StatCard from '@/components/StatCard.vue'
import Panel from '@/components/Panel.vue'
import StatusChip from '@/components/StatusChip.vue'
import EmptyState from '@/components/EmptyState.vue'

const toast = useToast()
const confirm = useConfirm()

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
  { label: 'Menunggu Verifikasi', value: 'PENDING' },
  { label: 'Sedang Ditindaklanjuti', value: 'IN_PROGRESS' },
  { label: 'Selesai Ditangani', value: 'RESOLVED' },
  { label: 'Ditolak', value: 'REJECTED' },
]

const typeOptions = [
  { label: 'Semua Kategori Aduan', value: 'ALL' },
  { label: 'Kerusakan Alat / Barang', value: 'Kerusakan Alat / Barang' },
  { label: 'Kehilangan / Tidak Ditemukan', value: 'Kehilangan / Tidak Ditemukan' },
  { label: 'Perpindahan Ruangan / Lokasi', value: 'Perpindahan Ruangan / Lokasi' },
  { label: 'Kondisi Fisik & Kebersihan', value: 'Kondisi Fisik & Kebersihan' },
  { label: 'Masalah Lainnya', value: 'Lainnya' },
]

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
      summary: 'Tindak lanjut disimpan',
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

function deleteReport(r: any) {
  confirm.require({
    message: `Hapus laporan pengaduan ${r.report_number}? Tindakan ini tidak dapat dibatalkan.`,
    header: 'Konfirmasi hapus laporan',
    icon: 'pi pi-exclamation-triangle',
    acceptClass: 'p-button-danger',
    accept: async () => {
      try {
        await api.delete(`/reports/${r.id}`)
        toast.add({ severity: 'success', summary: 'Laporan dihapus', life: 2000 })
        fetchReports()
      } catch (err: any) {
        toast.add({ severity: 'error', summary: 'Gagal menghapus', detail: err?.response?.data?.error, life: 3000 })
      }
    },
  })
}

function copyPublicLink() {
  const url = `${window.location.origin}/lapor`
  navigator.clipboard.writeText(url)
  toast.add({
    severity: 'info',
    summary: 'Link publik disalin',
    detail: 'Link formulir pengaduan siap dibagikan ke publik atau ditempel pada QR.',
    life: 2500,
  })
}

function getWhatsAppUrl(phone?: string, ticketNo?: string) {
  if (!phone || phone === '-') return '#'
  let cleaned = phone.replace(/[^0-9]/g, '')
  if (cleaned.startsWith('0')) {
    cleaned = '62' + cleaned.slice(1)
  }
  const text = encodeURIComponent(`Halo, menindaklanjuti laporan Anda di sistem inventaris dengan nomor tiket *${ticketNo || ''}*...`)
  return `https://wa.me/${cleaned}?text=${text}`
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
  <div>
    <PageHeader
      crumb="Operasional"
      title="Pengaduan &amp; Pelaporan Aset"
      sub="Kelola laporan kerusakan, kehilangan, dan pengaduan aset publik secara terintegrasi."
    >
      <template #actions>
        <Button
          label="Salin Link Publik"
          icon="pi pi-link"
          size="small"
          severity="secondary"
          outlined
          @click="copyPublicLink"
        />
        <Button
          label="Buka Portal Publik"
          icon="pi pi-external-link"
          size="small"
          @click="$router.push('/lapor')"
        />
      </template>
    </PageHeader>

    <!-- KPI Row -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-3.5 mb-5">
      <StatCard
        label="Total Laporan"
        :value="stats.total"
        icon="pi pi-inbox"
        tone="accent"
        featured
        hint="Seluruh tiket masuk"
      />
      <StatCard
        label="Menunggu Verifikasi"
        :value="stats.pending"
        icon="pi pi-clock"
        :tone="stats.pending > 0 ? 'warn' : 'neutral'"
        hint="Perlu ditinjau petugas"
      />
      <StatCard
        label="Sedang Diproses"
        :value="stats.in_progress"
        icon="pi pi-wrench"
        tone="info"
        hint="Dalam penanganan teknisi"
      />
      <StatCard
        label="Selesai Ditangani"
        :value="stats.resolved"
        icon="pi pi-check-circle"
        tone="ok"
        hint="Tiket tuntas"
      />
    </div>

    <!-- Filter Toolbar -->
    <div class="panel p-3 mb-4 flex flex-wrap items-center gap-2.5">
      <div class="relative flex-1 min-w-[220px]">
        <i class="pi pi-search absolute left-3 top-1/2 -translate-y-1/2 text-txt-dim text-[12px]" aria-hidden="true" />
        <InputText
          v-model="searchQuery"
          placeholder="Cari nomor tiket, nama pelapor, barang, atau lokasi…"
          class="w-full !pl-8 !text-[12.5px]"
        />
      </div>

      <Select
        v-model="statusFilter"
        :options="statusOptions"
        optionLabel="label"
        optionValue="value"
        class="w-full sm:w-48 !text-[12.5px]"
      />

      <Select
        v-model="typeFilter"
        :options="typeOptions"
        optionLabel="label"
        optionValue="value"
        class="w-full sm:w-56 !text-[12.5px]"
      />

      <Button
        icon="pi pi-refresh"
        size="small"
        severity="secondary"
        outlined
        aria-label="Muat ulang laporan"
        v-tooltip.top="'Muat Ulang'"
        @click="fetchReports(false)"
      />
    </div>

    <!-- Reports Table Panel -->
    <Panel title="Daftar Tiket Pengaduan" icon="pi pi-megaphone" dense>
      <template #actions>
        <span class="t-mono !text-[11px] text-txt-dim">{{ total }} tiket</span>
      </template>

      <EmptyState v-if="loading" loading title="Memuat daftar pengaduan…" />

      <EmptyState
        v-else-if="reports.length === 0"
        icon="pi pi-inbox"
        title="Belum ada laporan pengaduan"
        sub="Laporan kerusakan atau kendala aset yang dikirim melalui portal publik akan muncul di sini."
      />

      <div v-else class="overflow-x-auto">
        <table class="w-full text-left border-collapse text-[13px]">
          <thead>
            <tr class="border-b border-line bg-panel-2 text-[11px] uppercase tracking-[0.05em] text-txt-dim font-semibold">
              <th class="px-3.5 py-2.5">Tiket &amp; Waktu</th>
              <th class="px-3.5 py-2.5">Barang / Aset</th>
              <th class="px-3.5 py-2.5">Pelapor &amp; Kategori</th>
              <th class="px-3.5 py-2.5">Uraian Masalah</th>
              <th class="px-3.5 py-2.5">Status</th>
              <th class="px-3.5 py-2.5 text-right">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-line">
            <tr
              v-for="r in reports"
              :key="r.id"
              class="data-row cursor-pointer"
              @click="openDetail(r)"
            >
              <td class="px-3.5 py-2.5 whitespace-nowrap">
                <div class="t-mono font-bold text-acc-500 text-[12px]">{{ r.report_number }}</div>
                <div class="t-mono !text-[10.5px] text-txt-dim mt-0.5">{{ formatDate(r.created_at) }}</div>
              </td>

              <td class="px-3.5 py-2.5">
                <div class="font-semibold text-txt text-[13px]">{{ r.item_name || 'Aset Umum / Fasilitas' }}</div>
                <div class="flex items-center gap-2 mt-0.5 text-[11px] text-txt-dim">
                  <span v-if="r.item_sku" class="t-mono px-1.5 py-0.5 rounded border border-line bg-panel-2">{{ r.item_sku }}</span>
                  <span v-if="r.location" class="flex items-center gap-1">
                    <i class="pi pi-map-marker text-[10px]" aria-hidden="true" />
                    {{ r.location }}
                  </span>
                </div>
              </td>

              <td class="px-3.5 py-2.5">
                <div class="font-medium text-txt text-[12.5px]">{{ r.reporter_name }}</div>
                <div class="text-[11px] text-txt-dim mt-0.5">{{ r.report_type }}</div>
              </td>

              <td class="px-3.5 py-2.5 max-w-xs">
                <p class="text-[12px] text-txt-dim line-clamp-2 leading-relaxed">{{ r.description }}</p>
                <div v-if="r.photo_url" class="mt-1 inline-flex items-center gap-1 text-[10.5px] text-acc-500 font-medium">
                  <i class="pi pi-image text-[10px]" aria-hidden="true" />
                  <span>Ada lampiran foto</span>
                </div>
              </td>

              <td class="px-3.5 py-2.5 whitespace-nowrap">
                <StatusChip :kind="r.status" />
              </td>

              <td class="px-3.5 py-2.5 text-right whitespace-nowrap" @click.stop>
                <div class="flex items-center justify-end gap-1">
                  <Button
                    icon="pi pi-eye"
                    size="small"
                    text
                    rounded
                    severity="secondary"
                    aria-label="Tindak lanjuti laporan"
                    v-tooltip.top="'Tindak Lanjuti'"
                    @click="openDetail(r)"
                  />
                  <Button
                    icon="pi pi-trash"
                    size="small"
                    text
                    rounded
                    severity="danger"
                    aria-label="Hapus laporan"
                    v-tooltip.top="'Hapus Laporan'"
                    @click="deleteReport(r)"
                  />
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <template v-if="total > perPage" #footer>
        <div class="flex items-center justify-between text-[12px] text-txt-dim">
          <span class="t-mono !text-[11.5px]">Menampilkan {{ reports.length }} dari {{ total }} laporan</span>
          <div class="flex items-center gap-1">
            <Button
              label="Sebelumnya"
              icon="pi pi-chevron-left"
              size="small"
              text
              severity="secondary"
              :disabled="page <= 1"
              @click="page--; fetchReports(false)"
            />
            <span class="t-num px-2 font-semibold text-txt">Hal. {{ page }}</span>
            <Button
              label="Berikutnya"
              icon="pi pi-chevron-right"
              iconPos="right"
              size="small"
              text
              severity="secondary"
              :disabled="page * perPage >= total"
              @click="page++; fetchReports(false)"
            />
          </div>
        </div>
      </template>
    </Panel>

    <!-- Modal Tindak Lanjut & Detail Laporan -->
    <Dialog
      v-model:visible="detailModalVisible"
      modal
      :header="`Tindak Lanjut Pengaduan ${selectedReport?.report_number || ''}`"
      :style="{ width: '92vw', maxWidth: '640px' }"
    >
      <div v-if="selectedReport" class="flex flex-col gap-4 text-[12.5px] pt-1">
        <div class="p-3 rounded-[8px] bg-panel-2 border border-line flex items-center justify-between gap-3">
          <div>
            <span class="text-[10.5px] text-txt-dim uppercase tracking-[0.05em] font-semibold block mb-1">Status Saat Ini</span>
            <StatusChip :kind="selectedReport.status" />
          </div>
          <div class="text-right">
            <span class="text-[10.5px] text-txt-dim uppercase tracking-[0.05em] font-semibold block">Waktu Masuk</span>
            <div class="t-mono !text-[11.5px] text-txt mt-0.5">{{ formatDate(selectedReport.created_at) }}</div>
          </div>
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <div class="p-3 rounded-[8px] bg-panel-2 border border-line space-y-1">
            <span class="text-[10.5px] font-semibold text-acc-500 uppercase tracking-[0.05em] block">Barang / Aset Terkait</span>
            <div class="font-bold text-[13.5px] text-txt">{{ selectedReport.item_name || 'Aset Umum / Fasilitas' }}</div>
            <div v-if="selectedReport.item_sku" class="t-mono !text-[11.5px] text-txt-dim">SKU: {{ selectedReport.item_sku }}</div>
            <div v-if="selectedReport.location" class="text-[12px] text-txt-dim flex items-center gap-1">
              <i class="pi pi-map-marker text-sig-ok text-[10px]" aria-hidden="true" />
              <span>{{ selectedReport.location }}</span>
            </div>
          </div>

          <div class="p-3 rounded-[8px] bg-panel-2 border border-line space-y-1">
            <span class="text-[10.5px] font-semibold text-acc-500 uppercase tracking-[0.05em] block">Data Pelapor</span>
            <div class="font-bold text-[13.5px] text-txt">{{ selectedReport.reporter_name }}</div>
            <div v-if="selectedReport.reporter_contact && selectedReport.reporter_contact !== '-'" class="text-[12px] text-txt-dim flex items-center gap-2">
              <span>{{ selectedReport.reporter_contact }}</span>
              <a
                :href="getWhatsAppUrl(selectedReport.reporter_contact, selectedReport.report_number)"
                target="_blank"
                rel="noopener noreferrer"
                class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded bg-[var(--sig-ok-bg)] border border-[var(--sig-ok-line)] text-sig-ok font-semibold text-[10.5px]"
              >
                <i class="pi pi-whatsapp" aria-hidden="true" /> Chat WA
              </a>
            </div>
            <div class="text-[12px] text-txt-dim">Kategori: <strong class="text-txt">{{ selectedReport.report_type }}</strong></div>
          </div>
        </div>

        <div class="p-3 rounded-[8px] bg-panel-2 border border-line space-y-2">
          <span class="text-[10.5px] font-semibold text-txt-dim uppercase tracking-[0.05em] block">Deskripsi Masalah dari Pelapor</span>
          <p class="text-txt leading-relaxed whitespace-pre-wrap max-w-[65ch]">{{ selectedReport.description }}</p>

          <div v-if="selectedReport.photo_url" class="pt-2 border-t border-line">
            <span class="text-[10.5px] font-semibold text-txt-dim uppercase tracking-[0.05em] block mb-1.5">Foto Lampiran Bukti</span>
            <img
              :src="selectedReport.photo_url"
              alt="Foto bukti kerusakan aset"
              class="max-h-56 rounded-[7px] border border-line object-contain bg-ink-950 cursor-pointer hover:opacity-90 transition-opacity"
              @click="zoomedPhoto = selectedReport.photo_url; photoZoomModal = true"
            />
          </div>
        </div>

        <div class="p-4 rounded-[8px] bg-[var(--acc-wash)] border border-[var(--acc-border)] space-y-3">
          <span class="text-[11.5px] font-bold text-acc-500 uppercase tracking-[0.05em] flex items-center gap-1.5">
            <i class="pi pi-user-edit" aria-hidden="true" />
            <span>Formulir Tindak Lanjut Petugas</span>
          </span>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <label class="flex flex-col gap-1">
              <span class="text-[11.5px] font-semibold text-txt">Ubah Status Tiket</span>
              <Select
                v-model="targetStatus"
                :options="[
                  { label: 'Menunggu Verifikasi', value: 'PENDING' },
                  { label: 'Sedang Ditindaklanjuti', value: 'IN_PROGRESS' },
                  { label: 'Selesai Ditangani', value: 'RESOLVED' },
                  { label: 'Ditolak / Tidak Valid', value: 'REJECTED' },
                ]"
                optionLabel="label"
                optionValue="value"
                class="w-full !text-[12px]"
              />
            </label>

            <label class="flex flex-col gap-1">
              <span class="text-[11.5px] font-semibold text-txt">Perbarui Kondisi Aset di Master Data</span>
              <Select
                v-model="updateItemCondition"
                :options="conditionOptions"
                optionLabel="label"
                optionValue="value"
                class="w-full !text-[12px]"
                :disabled="!selectedReport.item_id"
              />
            </label>
          </div>

          <label class="flex flex-col gap-1">
            <span class="text-[11.5px] font-semibold text-txt">Catatan Tindakan / Umpan Balik Petugas</span>
            <Textarea
              v-model="adminNotes"
              rows="3"
              placeholder="Tuliskan tindakan yang diambil, jadwal teknisi, atau keterangan penyelesaian masalah…"
              class="w-full !text-[12px]"
            />
          </label>
        </div>

        <div class="flex items-center justify-end gap-2 pt-2 border-t border-line">
          <Button label="Tutup" size="small" severity="secondary" text @click="detailModalVisible = false" />
          <Button label="Simpan Tindak Lanjut" icon="pi pi-check" size="small" :loading="updatingStatus" @click="saveFollowUp" />
        </div>
      </div>
    </Dialog>

    <!-- Photo Zoom Modal -->
    <Dialog
      v-model:visible="photoZoomModal"
      modal
      header="Foto Bukti Kerusakan"
      :style="{ width: '90vw', maxWidth: '800px' }"
    >
      <div class="flex items-center justify-center p-2 bg-ink-950 border border-line rounded-[8px] overflow-hidden">
        <img :src="zoomedPhoto" alt="Foto bukti kerusakan diperbesar" class="max-h-[75vh] w-auto object-contain rounded" />
      </div>
    </Dialog>
  </div>
</template>
