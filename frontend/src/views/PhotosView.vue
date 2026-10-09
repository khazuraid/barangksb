<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import api from '@/api'
import { useToast } from 'primevue/usetoast'
import { useConfirm } from 'primevue/useconfirm'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import Dialog from 'primevue/dialog'
import PageHeader from '@/components/PageHeader.vue'
import StatCard from '@/components/StatCard.vue'
import EmptyState from '@/components/EmptyState.vue'

const toast = useToast()
const confirm = useConfirm()

const loading = ref(true)
const cleaning = ref(false)
const photos = ref<any[]>([])
const total = ref(0)
const stats = ref<any>({
  total_files: 0,
  total_size: 0,
  used_files: 0,
  used_size: 0,
  orphan_files: 0,
  orphan_size: 0,
})

const q = ref('')
const statusFilter = ref('all')
const page = ref(0)
const perPage = ref(24)
const perPageOptions = [12, 24, 48, 96]

const previewPhoto = ref<any | null>(null)
const previewVisible = ref(false)

function formatBytes(bytes: number) {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

function formatDate(dateStr: string) {
  if (!dateStr) return '-'
  try {
    const d = new Date(dateStr)
    return d.toLocaleDateString('id-ID', {
      day: '2-digit',
      month: 'short',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  } catch {
    return dateStr
  }
}

async function fetchPhotos(reset = false) {
  if (reset) page.value = 0
  loading.value = true
  try {
    const res = await api.get('/photos', {
      params: {
        q: q.value,
        status: statusFilter.value,
        page: page.value + 1,
        per_page: perPage.value,
      },
    })
    photos.value = res.data?.data || []
    total.value = res.data?.total || 0
    if (res.data?.stats) {
      stats.value = res.data.stats
    }
  } catch (err: any) {
    toast.add({
      severity: 'error',
      summary: 'Gagal memuat daftar foto',
      detail: err.response?.data?.error || err.message,
      life: 3500,
    })
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchPhotos()
})

watch([page, perPage, statusFilter], () => {
  fetchPhotos()
})

const lastPage = computed(() => Math.max(0, Math.ceil(total.value / perPage.value) - 1))
const range = computed(() => {
  if (!total.value) return '0 berkas foto'
  const from = page.value * perPage.value + 1
  const to = Math.min(total.value, (page.value + 1) * perPage.value)
  return `${from}–${to} dari ${total.value} berkas`
})

function openPreview(photo: any) {
  previewPhoto.value = photo
  previewVisible.value = true
}

function deleteSingle(photo: any) {
  const isOrphan = photo.is_orphan
  const warnText = isOrphan
    ? `Hapus foto yatim "${photo.filename}"? Berkas ini tidak terhubung ke item mana pun.`
    : `PERHATIAN: Foto "${photo.filename}" saat ini MASIH DIGUNAKAN oleh ${photo.references?.length || 1} catatan data. Tetap hapus dari penyimpanan fisik?`

  confirm.require({
    message: warnText,
    header: isOrphan ? 'Konfirmasi Hapus Foto Yatim' : 'Peringatan: Foto Sedang Digunakan',
    icon: isOrphan ? 'pi pi-trash' : 'pi pi-exclamation-triangle',
    acceptClass: 'p-button-danger',
    acceptLabel: 'Hapus Berkas',
    rejectLabel: 'Batal',
    accept: async () => {
      try {
        await api.delete(`/photos/${encodeURIComponent(photo.filename)}`)
        toast.add({
          severity: 'success',
          summary: 'Foto dihapus',
          detail: `${photo.filename} berhasil dihapus dari storage.`,
          life: 2500,
        })
        if (previewVisible.value && previewPhoto.value?.filename === photo.filename) {
          previewVisible.value = false
        }
        fetchPhotos()
      } catch (err: any) {
        toast.add({
          severity: 'error',
          summary: 'Gagal menghapus foto',
          detail: err.response?.data?.error || err.message,
          life: 3500,
        })
      }
    },
  })
}

function cleanAllOrphans() {
  if (stats.value.orphan_files === 0) {
    toast.add({
      severity: 'info',
      summary: 'Penyimpanan bersih',
      detail: 'Tidak ada foto yatim yang terdeteksi saat ini.',
      life: 2500,
    })
    return
  }

  confirm.require({
    message: `Bersihkan ${stats.value.orphan_files} foto yatim (${formatBytes(stats.value.orphan_size)}) dari penyimpanan fisik? Tindakan ini akan menghapus berkas foto yang tidak terhubung ke data barang/transaksi mana pun.`,
    header: 'Bersihkan Semua Foto Yatim',
    icon: 'pi pi-shield',
    acceptClass: 'p-button-danger',
    acceptLabel: 'Bersihkan Sekarang',
    rejectLabel: 'Batal',
    accept: async () => {
      cleaning.value = true
      try {
        const res = await api.post('/photos/clean-orphaned')
        toast.add({
          severity: 'success',
          summary: 'Pembersihan Selesai',
          detail: res.data?.message || 'Foto yatim berhasil dibersihkan.',
          life: 3500,
        })
        fetchPhotos(true)
      } catch (err: any) {
        toast.add({
          severity: 'error',
          summary: 'Gagal membersihkan foto yatim',
          detail: err.response?.data?.error || err.message,
          life: 4000,
        })
      } finally {
        cleaning.value = false
      }
    },
  })
}
</script>

<template>
  <div class="pb-16 w-full">
    <PageHeader
      crumb="Administrasi"
      title="Manajemen &amp; Audit Foto Fisik"
      sub="Pindai penyimpanan foto, deteksi berkas yatim yang tidak terhubung ke database, dan optimalkan ruang disk"
    >
      <template #actions>
        <Button
          label="Pindai Ulang"
          icon="pi pi-refresh"
          size="small"
          severity="secondary"
          outlined
          :loading="loading"
          @click="fetchPhotos(false)"
        />
        <Button
          :label="cleaning ? 'Membersihkan...' : 'Bersihkan Foto Yatim'"
          icon="pi pi-trash"
          size="small"
          severity="danger"
          :disabled="stats.orphan_files === 0 || cleaning"
          :loading="cleaning"
          @click="cleanAllOrphans"
        />
      </template>
    </PageHeader>

    <!-- KPI Metric Strip -->
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-3 mb-4">
      <StatCard
        label="Total Berkas Foto"
        :value="stats.total_files + ' foto'"
        icon="pi pi-images"
        tone="neutral"
        :hint="'Kapasitas: ' + formatBytes(stats.total_size)"
      />
      <StatCard
        label="Foto Terhubung Aktif"
        :value="stats.used_files + ' foto'"
        icon="pi pi-check-circle"
        tone="accent"
        :hint="'Terpakai: ' + formatBytes(stats.used_size) + ' di master & riwayat'"
      />
      <StatCard
        label="Foto Yatim (Tak Terhubung)"
        :value="stats.orphan_files + ' foto'"
        icon="pi pi-exclamation-circle"
        :tone="stats.orphan_files > 0 ? 'critical' : 'neutral'"
        :hint="stats.orphan_files > 0 ? 'Dapat dibersihkan: ' + formatBytes(stats.orphan_size) : 'Penyimpanan bersih tanpa foto yatim'"
      />
    </div>

    <!-- Unified Filter Toolbar -->
    <div class="panel p-3 mb-4 flex flex-wrap items-center justify-between gap-2.5">
      <div class="flex flex-wrap items-center gap-2 flex-1 min-w-[280px]">
        <div class="relative flex-1 min-w-[190px] max-w-sm">
          <i class="pi pi-search absolute left-3 top-1/2 -translate-y-1/2 text-ink-400 text-xs" />
          <InputText
            v-model="q"
            placeholder="Cari nama file, nama barang, atau SKU…"
            class="w-full !pl-8 !text-[12px] !py-1.5"
            @keyup.enter="fetchPhotos(true)"
          />
        </div>

        <Select
          v-model="statusFilter"
          :options="[
            { label: 'Semua Foto (' + stats.total_files + ')', value: 'all' },
            { label: 'Hanya Foto Yatim (' + stats.orphan_files + ')', value: 'orphan' },
            { label: 'Hanya Foto Terpakai (' + stats.used_files + ')', value: 'used' },
          ]"
          optionLabel="label"
          optionValue="value"
          class="!text-[12px] !py-0.5 w-[210px]"
        />

        <Button
          icon="pi pi-search"
          size="small"
          severity="secondary"
          outlined
          @click="fetchPhotos(true)"
        />

        <Button
          v-if="q || statusFilter !== 'all'"
          label="Reset"
          icon="pi pi-times"
          text
          size="small"
          severity="secondary"
          @click="q = ''; statusFilter = 'all'; fetchPhotos(true)"
        />
      </div>

      <div class="flex items-center gap-2">
        <Tag
          v-if="stats.orphan_files > 0"
          :value="stats.orphan_files + ' Foto Yatim Terdeteksi'"
          severity="danger"
          class="!text-[11px]"
        />
        <Tag
          v-else
          value="Penyimpanan Bersih"
          severity="success"
          class="!text-[11px]"
        />
      </div>
    </div>

    <!-- Empty State -->
    <EmptyState v-if="loading" icon="pi pi-spin pi-spinner" title="Memindai berkas foto…" />
    <EmptyState
      v-else-if="!photos.length"
      icon="pi pi-images"
      title="Tidak ada berkas foto ditemukan"
      sub="Tidak ada foto yang cocok dengan kata kunci atau filter status saat ini."
    />

    <!-- Photo Gallery Grid -->
    <div v-else class="flex flex-col gap-4">
      <div class="grid gap-3" style="grid-template-columns: repeat(auto-fill, minmax(260px, 1fr))">
        <div
          v-for="p in photos"
          :key="p.filename"
          class="panel overflow-hidden rounded-xl border flex flex-col justify-between transition-all hover:shadow-md"
          :class="p.is_orphan ? 'border-rose-300 dark:border-rose-900/60 bg-rose-50/20 dark:bg-rose-950/10' : ''"
          style="border-color: var(--line)"
        >
          <!-- Thumbnail Container -->
          <div class="relative bg-slate-900 aspect-video overflow-hidden cursor-pointer group" @click="openPreview(p)">
            <img
              :src="p.url"
              class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
              loading="lazy"
              :alt="p.filename"
            />
            <div class="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-2">
              <Button icon="pi pi-eye" rounded size="small" severity="secondary" text class="!bg-white/80 !text-slate-900" title="Perbesar" />
            </div>
            <!-- Status Badge in Thumbnail -->
            <div class="absolute top-2 left-2">
              <span
                v-if="p.is_orphan"
                class="px-2 py-0.5 rounded text-[10px] font-bold bg-rose-600 text-white shadow-sm flex items-center gap-1"
              >
                <i class="pi pi-exclamation-triangle text-[9px]" /> Yatim (Unused)
              </span>
              <span
                v-else
                class="px-2 py-0.5 rounded text-[10px] font-bold bg-emerald-600 text-white shadow-sm flex items-center gap-1"
              >
                <i class="pi pi-check text-[9px]" /> Terpakai
              </span>
            </div>
            <!-- Size Tag -->
            <div class="absolute bottom-2 right-2 bg-black/70 backdrop-blur-xs text-white text-[10.5px] px-1.5 py-0.5 rounded t-mono">
              {{ formatBytes(p.size) }}
            </div>
          </div>

          <!-- Metadata & Reference Card -->
          <div class="p-3 flex-1 flex flex-col justify-between gap-2.5">
            <div>
              <div class="flex items-start justify-between gap-1.5">
                <span class="font-mono text-[11px] text-slate-500 dark:text-slate-400 truncate max-w-[170px]" :title="p.filename">
                  {{ p.filename }}
                </span>
                <span class="text-[10px] text-slate-400 dark:text-slate-500 shrink-0">
                  {{ formatDate(p.mod_time) }}
                </span>
              </div>

              <!-- References List -->
              <div class="mt-2 pt-2 border-t text-[11px]" style="border-color: var(--line)">
                <div v-if="p.is_orphan" class="text-rose-600 dark:text-rose-400 flex items-start gap-1 leading-snug">
                  <i class="pi pi-info-circle text-[11px] mt-0.5 shrink-0" />
                  <span>Tidak terhubung ke barang atau transaksi. Aman untuk dihapus.</span>
                </div>
                <div v-else class="flex flex-col gap-1">
                  <div
                    v-for="(ref, idx) in p.references"
                    :key="idx"
                    class="p-1.5 rounded border text-[11px] flex flex-col gap-0.5"
                    style="background: var(--panel-2); border-color: var(--line)"
                  >
                    <div class="flex items-center gap-1 font-semibold truncate" style="color: var(--txt)">
                      <i :class="ref.type === 'item' ? 'pi pi-box' : ref.type === 'maintenance' ? 'pi pi-wrench' : 'pi pi-history'" class="text-[10px] text-indigo-500" />
                      <span class="truncate">{{ ref.title }}</span>
                    </div>
                    <span class="text-[10px]" style="color: var(--txt-dim)">{{ ref.sub }}</span>
                  </div>
                </div>
              </div>
            </div>

            <!-- Action buttons -->
            <div class="pt-2 border-t flex items-center justify-between gap-1" style="border-color: var(--line)">
              <div class="flex items-center gap-1">
                <a
                  :href="p.url + '?download=true'"
                  download
                  class="p-1.5 text-slate-500 hover:text-slate-800 dark:hover:text-slate-200 rounded hover:bg-slate-100 dark:hover:bg-slate-800 text-[11px] flex items-center gap-1 transition-colors"
                  title="Unduh Berkas"
                >
                  <i class="pi pi-download text-xs" />
                  <span>Unduh</span>
                </a>
              </div>
              <Button
                icon="pi pi-trash"
                size="small"
                text
                severity="danger"
                :label="p.is_orphan ? 'Hapus' : ''"
                class="!text-[11px] !py-1"
                :title="p.is_orphan ? 'Hapus foto tidak terpakai' : 'Hapus foto dari storage'"
                @click="deleteSingle(p)"
              />
            </div>
          </div>
        </div>
      </div>

      <!-- Pagination Footer -->
      <div
        class="panel p-3 rounded-xl border flex flex-wrap items-center justify-between gap-3 text-[12px]"
        style="background: var(--paper-1); border-color: var(--line)"
      >
        <span style="color: var(--txt-dim)">{{ range }}</span>
        <div class="flex items-center gap-2">
          <Select v-model="perPage" :options="perPageOptions" class="!text-[12px] !py-0.5 w-[95px]" />
          <Button
            icon="pi pi-angle-left"
            size="small"
            text
            severity="secondary"
            :disabled="page === 0"
            @click="page--"
          />
          <span class="t-num font-semibold px-1" style="color: var(--txt)">Hal. {{ page + 1 }} / {{ lastPage + 1 }}</span>
          <Button
            icon="pi pi-angle-right"
            size="small"
            text
            severity="secondary"
            :disabled="page >= lastPage"
            @click="page++"
          />
        </div>
      </div>
    </div>

    <!-- Preview Modal -->
    <Dialog
      v-model:visible="previewVisible"
      modal
      :header="previewPhoto?.filename || 'Pratinjau Foto'"
      :style="{ width: '92vw', maxWidth: '780px' }"
    >
      <div v-if="previewPhoto" class="flex flex-col gap-3">
        <div class="rounded-xl overflow-hidden bg-black flex items-center justify-center max-h-[65vh]">
          <img :src="previewPhoto.url" class="max-w-full max-h-[65vh] object-contain" :alt="previewPhoto.filename" />
        </div>
        <div class="p-3 rounded-lg border text-[12px] flex flex-wrap items-center justify-between gap-2" style="background: var(--panel-2); border-color: var(--line)">
          <div class="flex flex-col gap-0.5">
            <span class="font-bold" style="color: var(--txt)">Status: {{ previewPhoto.is_orphan ? 'Foto Yatim (Tidak Terpakai)' : 'Foto Terpakai' }}</span>
            <span style="color: var(--txt-dim)">Ukuran: {{ formatBytes(previewPhoto.size) }} · Diunggah: {{ formatDate(previewPhoto.mod_time) }}</span>
          </div>
          <div class="flex items-center gap-2">
            <a :href="previewPhoto.url + '?download=true'" download>
              <Button label="Unduh" icon="pi pi-download" size="small" severity="secondary" outlined />
            </a>
            <Button label="Hapus Foto" icon="pi pi-trash" size="small" severity="danger" @click="deleteSingle(previewPhoto)" />
          </div>
        </div>
      </div>
    </Dialog>
  </div>
</template>
