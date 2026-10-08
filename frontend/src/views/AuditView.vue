<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import api from '@/api'
import PageHeader from '@/components/PageHeader.vue'
import Panel from '@/components/Panel.vue'
import StatusChip from '@/components/StatusChip.vue'
import EmptyState from '@/components/EmptyState.vue'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'

const auditLogs = ref<any[]>([])
const loading = ref(true)
const q = ref('')
const opFilter = ref('')
const expanded = ref<string | null>(null)

const opOptions = [
  { label: 'Semua aksi', value: '' },
  { label: 'INSERT (Tambah)', value: 'INSERT' },
  { label: 'UPDATE (Ubah)', value: 'UPDATE' },
  { label: 'DELETE (Hapus)', value: 'DELETE' },
]

async function fetchAudit() {
  loading.value = true
  try {
    const res = await api.get('/audit')
    const list = res.data?.data || (Array.isArray(res.data) ? res.data : [])
    auditLogs.value = Array.isArray(list) ? list : []
  } catch {
    auditLogs.value = []
  } finally {
    loading.value = false
  }
}

onMounted(() => fetchAudit())

const filtered = computed(() => {
  const safeList = Array.isArray(auditLogs.value) ? auditLogs.value : []
  return safeList.filter((l) => {
    const matchOp = !opFilter.value || l.op === opFilter.value
    const matchQ =
      !q.value ||
      [l.table_name, l.row_id, l.op].join(' ').toLowerCase().includes(q.value.toLowerCase())
    return matchOp && matchQ
  })
})

const counts = computed(() => {
  const safeList = Array.isArray(auditLogs.value) ? auditLogs.value : []
  return {
    total: safeList.length,
    ins: safeList.filter((l) => l.op === 'INSERT').length,
    upd: safeList.filter((l) => l.op === 'UPDATE').length,
    del: safeList.filter((l) => l.op === 'DELETE').length,
  }
})

function prettyJSON(raw: any) {
  if (!raw) return '—'
  try {
    return JSON.stringify(typeof raw === 'string' ? JSON.parse(raw) : raw, null, 2)
  } catch {
    return String(raw)
  }
}
</script>

<template>
  <div class="pb-16 max-w-6xl mx-auto">
    <PageHeader
      crumb="Administrasi"
      title="Audit Log &amp; Keamanan"
      sub="Rekaman otomatis setiap operasi penambahan, perubahan, dan penghapusan data database"
    >
      <template #actions>
        <Button
          label="Segarkan Log"
          icon="pi pi-refresh"
          size="small"
          severity="secondary"
          outlined
          :loading="loading"
          @click="fetchAudit"
        />
      </template>
    </PageHeader>

    <!-- Audit KPI Summary Strip -->
    <div
      class="grid grid-cols-2 sm:grid-cols-4 gap-px mb-4 rounded-xl overflow-hidden border shadow-xs"
      style="border-color: var(--line); background: var(--line)"
    >
      <div
        v-for="c in [
          { k: 'TOTAL LOG', v: counts.total, cls: 'text-ink-100' },
          { k: 'INSERT (TAMBAH)', v: counts.ins, cls: 'text-sig-ok' },
          { k: 'UPDATE (UBAH)', v: counts.upd, cls: 'text-sig-info' },
          { k: 'DELETE (HAPUS)', v: counts.del, cls: 'text-sig-bad' },
        ]"
        :key="c.k"
        class="px-4 py-3"
        style="background: var(--panel)"
      >
        <div class="t-label">{{ c.k }}</div>
        <div class="t-num text-[20px] font-bold mt-1" :class="c.cls">{{ c.v }}</div>
      </div>
    </div>

    <!-- Audit Table Panel -->
    <Panel title="Jejak Mutasi Database &amp; Keamanan" icon="pi pi-shield" dense>
      <template #actions>
        <InputText
          v-model="q"
          placeholder="Cari tabel / row id…"
          class="!text-[12px] !py-1.5 w-[190px]"
        />
        <Select
          v-model="opFilter"
          :options="opOptions"
          optionLabel="label"
          optionValue="value"
          class="!text-[12px] w-[160px]"
        />
      </template>

      <EmptyState v-if="loading" icon="pi pi-spin pi-spinner" title="Memuat audit log…" />
      <EmptyState
        v-else-if="!filtered.length"
        icon="pi pi-shield"
        title="Tidak ada catatan audit"
        sub="Jejak audit terekam otomatis saat ada modifikasi data barang atau pengguna."
      />

      <div v-else class="overflow-x-auto">
        <table class="w-full text-[12.5px]">
          <thead>
            <tr class="text-left" style="background: var(--paper-2)">
              <th class="t-label px-4 py-2.5 w-[180px]">Waktu</th>
              <th class="t-label px-4 py-2.5 w-[170px]">Tabel Target</th>
              <th class="t-label px-4 py-2.5 w-[110px]">Aksi</th>
              <th class="t-label px-4 py-2.5">Row ID / Kunci</th>
              <th class="t-label px-4 py-2.5 w-[80px] text-right">Detail</th>
            </tr>
          </thead>
          <tbody class="divide-y" style="border-color: var(--line-soft)">
            <template v-for="(l, i) in filtered" :key="i">
              <tr
                class="hover:bg-paper-2 transition-colors cursor-pointer"
                @click="expanded = expanded === l.row_id + i ? null : l.row_id + i"
              >
                <td class="px-4 py-2.5 whitespace-nowrap" style="color: var(--txt-dim)">{{ l.at }}</td>
                <td class="px-4 py-2.5">
                  <span class="t-mono font-semibold text-acc-400">{{ l.table_name }}</span>
                </td>
                <td class="px-4 py-2.5"><StatusChip :kind="l.op" /></td>
                <td class="px-4 py-2.5">
                  <span class="t-mono text-[11px]" style="color: var(--txt-dim)">{{ l.row_id || '—' }}</span>
                </td>
                <td class="px-4 py-2.5 text-right">
                  <i
                    class="pi text-[10px]"
                    :class="expanded === l.row_id + i ? 'pi-chevron-up' : 'pi-chevron-down'"
                    style="color: var(--txt-dim)"
                  />
                </td>
              </tr>

              <!-- JSON Diff Expanded Row -->
              <tr v-if="expanded === l.row_id + i">
                <td colspan="5" class="px-4 pb-3.5 pt-0" style="background: var(--paper-1)">
                  <div class="grid md:grid-cols-2 gap-3 pt-2">
                    <div class="flex flex-col gap-1">
                      <div class="t-label text-rose-400">Data Sebelumnya (Old Data)</div>
                      <pre class="panel !rounded-md p-3 t-mono text-[11px] overflow-x-auto max-h-[220px] bg-black/60 border border-line">{{ prettyJSON(l.old_data) }}</pre>
                    </div>
                    <div class="flex flex-col gap-1">
                      <div class="t-label text-emerald-400">Data Baru (New Data)</div>
                      <pre class="panel !rounded-md p-3 t-mono text-[11px] overflow-x-auto max-h-[220px] bg-black/60 border border-line">{{ prettyJSON(l.new_data) }}</pre>
                    </div>
                  </div>
                </td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>

      <template #footer>
        <div class="flex items-center justify-between text-[11.5px]" style="color: var(--txt-dim)">
          <span>Menampilkan {{ filtered.length }} dari {{ auditLogs.length }} catatan audit</span>
          <span>Klik baris untuk melihat perbedaan payload JSON</span>
        </div>
      </template>
    </Panel>
  </div>
</template>
