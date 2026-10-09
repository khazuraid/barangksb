<script setup lang="ts">
import { ref, computed } from 'vue'
import api from '@/api'
import { useToast } from 'primevue/usetoast'
import { useRouter } from 'vue-router'
import Button from 'primevue/button'
import FileUpload from 'primevue/fileupload'
import PageHeader from '@/components/PageHeader.vue'
import Panel from '@/components/Panel.vue'
import EmptyState from '@/components/EmptyState.vue'
import Tag from 'primevue/tag'

const toast = useToast()
const router = useRouter()

const file = ref<File | null>(null)
const uploading = ref(false)
const result = ref<{ ok: number; fail: number; errors: string[] } | null>(null)

function onSelect(e: any) {
  file.value = e.files?.[0] ?? null
  result.value = null
}

async function upload() {
  if (!file.value) {
    toast.add({ severity: 'warn', summary: 'Pilih berkas CSV dulu', life: 2500 })
    return
  }
  uploading.value = true
  try {
    const fd = new FormData()
    fd.append('file', file.value)
    const res = await api.post('/adjust/bulk', fd, { headers: { 'Content-Type': 'multipart/form-data' } })
    result.value = res.data
    const d = res.data
    toast.add({
      severity: d.fail ? 'warn' : 'success', life: 4000,
      summary: 'Opname selesai',
      detail: `${d.ok} baris berhasil, ${d.fail} gagal`,
    })
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Upload gagal', detail: e.response?.data?.error, life: 4000 })
  } finally { uploading.value = false }
}

function downloadTemplate() {
  const csv = 'sku,stok_fisik\nELKO-2026-001,12\nELKO-2026-002,0\n'
  const blob = new Blob(['\uFEFF' + csv], { type: 'text/csv;charset=utf-8' })
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = 'template_opname_massal.csv'
  a.click()
  URL.revokeObjectURL(a.href)
}

function exportCurrent() {
  window.open('/api/export/items.csv', '_blank')
}

const preview = computed(() => {
  if (!file.value) return []
  return []
})
</script>

<template>
  <div>
    <PageHeader crumb="Perangkat" title="Opname Massal"
      sub="Sesuaikan stok banyak barang sekaligus dengan mengunggah berkas CSV">
      <template #actions>
        <Button label="Unduh Data Barang" icon="pi pi-download" size="small" severity="secondary" outlined
                @click="exportCurrent" />
      </template>
    </PageHeader>

    <!-- steps -->
    <div class="grid sm:grid-cols-3 gap-3 mb-5">
      <div v-for="(s, i) in [
        { n: '01', t: 'Unduh data barang', d: 'CSV memuat seluruh SKU dan stok sistem saat ini.' },
        { n: '02', t: 'Isi stok fisik', d: 'Cukup ubah kolom stok_fisik sesuai hasil hitung.' },
        { n: '03', t: 'Unggah kembali', d: 'Sistem mencatat selisih sebagai transaksi opname.' },
      ]" :key="s.n" class="panel p-4 flex flex-col justify-between">
        <div>
          <span class="t-mono font-bold text-xs px-2 py-0.5 rounded border text-indigo-600 dark:text-indigo-400 bg-indigo-50 dark:bg-indigo-950/40 border-indigo-200 dark:border-indigo-800">{{ s.n }}</span>
          <div class="text-[13px] font-semibold mt-2.5" style="color: var(--txt)">{{ s.t }}</div>
          <p class="text-[11.5px] mt-1 leading-relaxed" style="color: var(--txt-dim)">{{ s.d }}</p>
        </div>
      </div>
    </div>

    <div class="grid lg:grid-cols-2 gap-4 items-start">
      <!-- upload -->
      <Panel title="Unggah Berkas Opname" icon="pi pi-upload">
        <FileUpload
          mode="basic"
          accept=".csv,text/csv"
          :auto="false"
          chooseLabel="Pilih File CSV"
          :customUpload="true"
          @select="onSelect"
        />

        <div v-if="file" class="mt-3.5 panel p-3 flex items-center gap-3"
             style="background: var(--panel-2)">
          <div class="w-8 h-8 rounded-lg grid place-items-center text-indigo-600 dark:text-indigo-400 bg-indigo-50 dark:bg-indigo-950/40 border border-indigo-200 dark:border-indigo-800 text-xs shrink-0">
            <i class="pi pi-file" />
          </div>
          <div class="flex-1 min-w-0">
            <div class="text-[12.5px] font-semibold truncate" style="color: var(--txt)">{{ file.name }}</div>
            <div class="t-mono text-[11px]" style="color: var(--txt-dim)">{{ (file.size / 1024).toFixed(1) }} KB</div>
          </div>
          <Button icon="pi pi-times" text rounded size="small" severity="secondary" @click="file = null" />
        </div>

        <div class="flex flex-wrap gap-2.5 mt-4">
          <Button label="Proses Opname" icon="pi pi-check" :loading="uploading"
                  :disabled="!file" severity="warn" @click="upload" />
          <Button label="Unduh Template" icon="pi pi-file-edit" text severity="secondary"
                  @click="downloadTemplate" />
        </div>

        <div class="mt-4 pt-4 border-t" style="border-color: var(--line)">
          <div class="text-[11px] font-semibold uppercase tracking-wider mb-2" style="color: var(--txt-dim)">Format Berkas</div>
          <pre class="panel p-3 t-mono text-[11px] overflow-x-auto" style="background: var(--panel-2)">sku,stok_fisik
ELKO-2026-001,12
ELKO-2026-002,0</pre>
          <ul class="text-[11.5px] mt-2.5 space-y-1" style="color: var(--txt-dim)">
            <li>• Baris header opsional — akan dilewati otomatis.</li>
            <li>• Kolom 1 = SKU persis seperti di sistem (peka huruf besar/kecil).</li>
            <li>• Kolom 2 = jumlah stok nyata hasil hitung fisik.</li>
            <li>• Selisih dicatat sebagai transaksi ADJUST+ / ADJUST−.</li>
          </ul>
        </div>
      </Panel>

      <!-- result -->
      <Panel title="Hasil Proses" icon="pi pi-list-check">
        <EmptyState v-if="!result" icon="pi pi-clipboard"
                    title="Belum ada proses dijalankan"
                    sub="Hasil per baris akan tampil di sini setelah berkas diproses." />

        <template v-else>
          <div class="grid grid-cols-2 gap-3 mb-4">
            <div class="panel p-3.5 flex items-center justify-between">
              <div>
                <div class="text-[11px] font-semibold uppercase tracking-wider" style="color: var(--txt-dim)">Berhasil</div>
                <div class="t-num text-[22px] font-bold text-emerald-600 dark:text-emerald-400 mt-0.5">{{ result.ok }}</div>
              </div>
              <div class="w-8 h-8 rounded-lg grid place-items-center bg-emerald-50 dark:bg-emerald-950/40 border border-emerald-200 dark:border-emerald-800 text-emerald-600 dark:text-emerald-400 text-xs">
                <i class="pi pi-check" />
              </div>
            </div>
            <div class="panel p-3.5 flex items-center justify-between">
              <div>
                <div class="text-[11px] font-semibold uppercase tracking-wider" style="color: var(--txt-dim)">Gagal</div>
                <div class="t-num text-[22px] font-bold mt-0.5" :class="result.fail ? 'text-rose-600 dark:text-rose-400' : 'text-slate-500'">{{ result.fail }}</div>
              </div>
              <div class="w-8 h-8 rounded-lg grid place-items-center bg-rose-50 dark:bg-rose-950/40 border border-rose-200 dark:border-rose-800 text-rose-600 dark:text-rose-400 text-xs">
                <i class="pi pi-times" />
              </div>
            </div>
          </div>

          <div v-if="result.errors?.length">
            <div class="t-label mb-2">Baris Bermasalah</div>
            <ul class="divide-y max-h-[260px] overflow-y-auto border rounded-md"
                style="border-color: var(--line); border-color: var(--line)">
              <li v-for="(err, i) in result.errors" :key="i"
                  class="px-3.5 py-2.5 t-mono flex items-start gap-2.5" style="border-color: var(--line)">
                <i class="pi pi-exclamation-circle text-sig-bad text-[11px] mt-0.5" />
                <span>{{ err }}</span>
              </li>
            </ul>
          </div>
          <Tag v-else severity="success" class="w-full justify-center !py-2.5"
               icon="pi pi-check-circle" value="Semua baris berhasil diproses" />

          <div class="flex gap-2.5 mt-4">
            <Button label="Lihat Riwayat" icon="pi pi-history" size="small" outlined
                    @click="router.push('/history')" />
            <Button label="Barang" icon="pi pi-box" size="small" outlined
                    @click="router.push('/items')" />
          </div>
        </template>
      </Panel>
    </div>
  </div>
</template>
