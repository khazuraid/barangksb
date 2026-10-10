<script setup lang="ts">
import { ref } from 'vue'
import api from '@/api'
import { useToast } from 'primevue/usetoast'
import { useRouter } from 'vue-router'
import Button from 'primevue/button'
import PageHeader from '@/components/PageHeader.vue'
import Panel from '@/components/Panel.vue'
import EmptyState from '@/components/EmptyState.vue'
import Tag from 'primevue/tag'

const toast = useToast()
const router = useRouter()

const fileInput = ref<HTMLInputElement | null>(null)
const selectedFile = ref<File | null>(null)
const uploading = ref(false)
const result = ref<{ ok: number; unchanged?: number; fail: number; total?: number; errors: string[] } | null>(null)

function onFileChange(e: Event) {
  const target = e.target as HTMLInputElement
  if (target.files?.[0]) {
    selectedFile.value = target.files[0]
    result.value = null
  }
}

function clearFile() {
  selectedFile.value = null
  if (fileInput.value) fileInput.value.value = ''
  result.value = null
}

async function upload() {
  if (!selectedFile.value) {
    toast.add({ severity: 'warn', summary: 'Pilih berkas Excel/CSV dulu', life: 2500 })
    return
  }
  uploading.value = true
  try {
    const fd = new FormData()
    fd.append('file', selectedFile.value)
    const res = await api.post('/adjust/bulk', fd, { headers: { 'Content-Type': 'multipart/form-data' } })
    result.value = res.data
    const d = res.data
    toast.add({
      severity: d.fail ? 'warn' : 'success',
      life: 4000,
      summary: 'Opname selesai diproses',
      detail: `${d.ok} disesuaikan, ${d.unchanged ?? 0} tanpa selisih, ${d.fail} gagal`,
    })
  } catch (e: any) {
    toast.add({
      severity: 'error',
      summary: 'Gagal memproses berkas',
      detail: e.response?.data?.error || e.message,
      life: 4000,
    })
  } finally {
    uploading.value = false
  }
}

function downloadData(fmt: 'xlsx' | 'csv') {
  const token = localStorage.getItem('token') || ''
  window.open(`/api/adjust/export?fmt=${fmt}&token=${encodeURIComponent(token)}`, '_blank')
}

function downloadTemplate(fmt: 'xlsx' | 'csv') {
  const token = localStorage.getItem('token') || ''
  window.open(`/api/adjust/template?fmt=${fmt}&token=${encodeURIComponent(token)}`, '_blank')
}
</script>

<template>
  <div>
    <PageHeader crumb="Perangkat" title="Opname Massal"
      sub="Sesuaikan stok banyak barang sekaligus sesuai format data master inventaris">
      <template #actions>
        <div class="flex flex-wrap items-center gap-2">
          <Button label="Unduh Data Barang (.xlsx)" icon="pi pi-file-excel" size="small" severity="success"
                  @click="downloadData('xlsx')" />
          <Button label="Unduh CSV" icon="pi pi-download" size="small" severity="secondary" outlined
                  @click="downloadData('csv')" />
        </div>
      </template>
    </PageHeader>

    <!-- 3 Steps -->
    <div class="grid sm:grid-cols-3 gap-3 mb-5">
      <div v-for="s in [
        { n: '01', t: 'Unduh data barang', d: 'Unduh berkas Excel/CSV berisi seluruh SKU, nama, dan stok sistem saat ini.' },
        { n: '02', t: 'Isi stok fisik', d: 'Cukup masukkan angka hasil hitung nyata pada kolom Stok Fisik.' },
        { n: '03', t: 'Unggah berkas', d: 'Sistem mencatat transaksi penyesuaian (ADJUST) otomatis pada barang berselisih.' },
      ]" :key="s.n" class="panel p-4 flex flex-col justify-between">
        <div>
          <span class="t-mono font-bold text-xs px-2 py-0.5 rounded border text-indigo-600 dark:text-indigo-400 bg-indigo-50 dark:bg-indigo-950/40 border-indigo-200 dark:border-indigo-800">{{ s.n }}</span>
          <div class="text-[13px] font-semibold mt-2.5" style="color: var(--txt)">{{ s.t }}</div>
          <p class="text-[11.5px] mt-1 leading-relaxed" style="color: var(--txt-dim)">{{ s.d }}</p>
        </div>
      </div>
    </div>

    <div class="grid lg:grid-cols-2 gap-4 items-start">
      <!-- Upload panel -->
      <Panel title="Unggah Berkas Opname" icon="pi pi-upload">
        <div class="space-y-4">
          <!-- File chooser -->
          <div>
            <label class="block text-[12px] font-semibold mb-1.5" style="color: var(--txt)">
              Pilih Berkas Spreadsheet (.xlsx, .xls, .csv)
            </label>
            <input
              ref="fileInput"
              type="file"
              accept=".xlsx,.xls,.csv,text/csv,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
              class="block w-full text-[12px] text-slate-500 file:mr-3 file:py-2 file:px-3 file:rounded-md file:border-0 file:text-[12px] file:font-semibold file:bg-indigo-600 file:text-white hover:file:bg-indigo-500 cursor-pointer border rounded-lg p-1.5"
              style="border-color: var(--line); background: var(--panel-2)"
              @change="onFileChange"
            />
          </div>

          <div v-if="selectedFile" class="panel p-3 flex items-center gap-3"
               style="background: var(--panel-2)">
            <div class="w-8 h-8 rounded-lg grid place-items-center text-indigo-600 dark:text-indigo-400 bg-indigo-50 dark:bg-indigo-950/40 border border-indigo-200 dark:border-indigo-800 text-xs shrink-0">
              <i class="pi pi-file" />
            </div>
            <div class="flex-1 min-w-0">
              <div class="text-[12.5px] font-semibold truncate" style="color: var(--txt)">{{ selectedFile.name }}</div>
              <div class="t-mono text-[11px]" style="color: var(--txt-dim)">{{ (selectedFile.size / 1024).toFixed(1) }} KB</div>
            </div>
            <Button icon="pi pi-times" text rounded size="small" severity="secondary" @click="clearFile" />
          </div>

          <div class="flex flex-wrap items-center gap-2.5 pt-1">
            <Button label="Proses Opname" icon="pi pi-check" :loading="uploading"
                    :disabled="!selectedFile" severity="warn" @click="upload" />
            <Button label="Unduh Template Excel" icon="pi pi-file-excel" size="small" text severity="secondary"
                    @click="downloadTemplate('xlsx')" />
            <Button label="Template CSV" icon="pi pi-file" size="small" text severity="secondary"
                    @click="downloadTemplate('csv')" />
          </div>

          <!-- Format guide table -->
          <div class="pt-4 border-t" style="border-color: var(--line)">
            <div class="text-[11px] font-semibold uppercase tracking-wider mb-2" style="color: var(--txt-dim)">
              Format Kolom Data Barang
            </div>
            <div class="panel p-3 t-mono text-[11px] overflow-x-auto space-y-1" style="background: var(--panel-2)">
              <div class="font-bold text-indigo-600 dark:text-indigo-400">
                SKU, Nama Barang, Kategori, Lokasi, Stok Sistem, Satuan, Stok Fisik, Keterangan
              </div>
              <div style="color: var(--txt-dim)">
                ALAT-2026-001, Tensimeter Digital, Alkes, Rawat Inap, 5, unit, 5, Sesuai fisik
              </div>
              <div style="color: var(--txt-dim)">
                ELEK-2026-002, Laptop Admin, IT, Ruang TU, 3, unit, 2, 1 unit servis
              </div>
            </div>
            <ul class="text-[11.5px] mt-2.5 space-y-1 leading-relaxed" style="color: var(--txt-dim)">
              <li>• Kolom wajib: <strong>SKU</strong> dan <strong>Stok Fisik</strong> (atau kolom <strong>Stok</strong>).</li>
              <li>• Kolom <strong>Nama, Kategori, Lokasi, Satuan</strong> disediakan sebagai referensi dan tidak akan mengubah data master.</li>
              <li>• Mendukung berkas <strong>Excel (.xlsx)</strong> dan <strong>CSV</strong> (pemisah koma maupun titik koma).</li>
              <li>• Baris dengan stok fisik sama dengan stok sistem tidak menghasilkan selisih mutasi.</li>
              <li>• Baris dengan stok fisik kosong atau tanda (-) akan dilewati otomatis.</li>
            </ul>
          </div>
        </div>
      </Panel>

      <!-- Result panel -->
      <Panel title="Hasil Proses" icon="pi pi-list-check">
        <EmptyState v-if="!result" icon="pi pi-clipboard"
                    title="Belum ada proses dijalankan"
                    sub="Hasil per baris akan tampil di sini setelah berkas diproses." />

        <template v-else>
          <div class="grid grid-cols-3 gap-2.5 mb-4">
            <div class="panel p-3 flex flex-col justify-between">
              <div class="text-[10.5px] font-semibold uppercase tracking-wider" style="color: var(--txt-dim)">Disesuaikan</div>
              <div class="t-num text-[20px] font-bold text-emerald-600 dark:text-emerald-400 mt-1">{{ result.ok }}</div>
            </div>
            <div class="panel p-3 flex flex-col justify-between">
              <div class="text-[10.5px] font-semibold uppercase tracking-wider" style="color: var(--txt-dim)">Tanpa Selisih</div>
              <div class="t-num text-[20px] font-bold text-slate-500 mt-1">{{ result.unchanged ?? 0 }}</div>
            </div>
            <div class="panel p-3 flex flex-col justify-between">
              <div class="text-[10.5px] font-semibold uppercase tracking-wider" style="color: var(--txt-dim)">Gagal</div>
              <div class="t-num text-[20px] font-bold mt-1" :class="result.fail ? 'text-rose-600 dark:text-rose-400' : 'text-slate-500'">{{ result.fail }}</div>
            </div>
          </div>

          <div v-if="result.errors?.length" class="space-y-2">
            <div class="text-[11px] font-semibold uppercase tracking-wider text-rose-500">Baris Bermasalah ({{ result.errors.length }})</div>
            <ul class="divide-y max-h-[260px] overflow-y-auto border rounded-md"
                style="border-color: var(--line)">
              <li v-for="(err, i) in result.errors" :key="i"
                  class="px-3 py-2 text-[11.5px] t-mono flex items-start gap-2 text-rose-600 dark:text-rose-400"
                  style="border-color: var(--line)">
                <i class="pi pi-exclamation-circle text-[11px] mt-0.5 shrink-0" />
                <span>{{ err }}</span>
              </li>
            </ul>
          </div>
          <Tag v-else severity="success" class="w-full justify-center !py-2.5"
               icon="pi pi-check-circle" value="Semua baris valid berhasil diproses tanpa kendala" />

          <div class="flex gap-2.5 mt-4">
            <Button label="Lihat Riwayat Mutasi" icon="pi pi-history" size="small" outlined
                    @click="router.push('/history')" />
            <Button label="Daftar Barang" icon="pi pi-box" size="small" outlined
                    @click="router.push('/items')" />
          </div>
        </template>
      </Panel>
    </div>
  </div>
</template>
