<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import api from '@/api'
import PageHeader from '@/components/PageHeader.vue'
import Panel from '@/components/Panel.vue'
import StatCard from '@/components/StatCard.vue'
import StatusChip from '@/components/StatusChip.vue'
import EmptyState from '@/components/EmptyState.vue'
import Tag from 'primevue/tag'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'

const auditLogs = ref<any[]>([])
const total = ref(0)
const page = ref(0)
const perPage = ref(25)
const perPageOptions = [25, 50, 100]

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

async function fetchAudit(reset = false) {
  if (reset) page.value = 0
  loading.value = true
  try {
    const res = await api.get('/audit', {
      params: { page: page.value + 1, per_page: perPage.value },
    })
    const list = res.data?.data || (Array.isArray(res.data) ? res.data : [])
    auditLogs.value = Array.isArray(list) ? list : []
    total.value = res.data?.total || auditLogs.value.length
  } catch {
    auditLogs.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

onMounted(() => fetchAudit())
watch([page, perPage], () => fetchAudit())

const lastPage = computed(() => Math.max(0, Math.ceil(total.value / perPage.value) - 1))
const range = computed(() => {
  if (!total.value) return '0 catatan'
  const from = page.value * perPage.value + 1
  const to = Math.min(total.value, (page.value + 1) * perPage.value)
  return `${from}–${to} dari ${total.value}`
})

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
    total: total.value,
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
  <div class="pb-16 w-full">
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
          @click="fetchAudit(true)"
        />
      </template>
    </PageHeader>

    <!-- Audit KPI Summary Strip -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-3 mb-4">
      <StatCard
        label="Total Log Aktivitas"
        :value="counts.total"
        icon="pi pi-shield"
        tone="accent"
        hint="Jejak mutasi database"
      />
      <StatCard
        label="Insert (Tambah)"
        :value="counts.ins"
        icon="pi pi-plus"
        tone="ok"
        hint="Data baru ditambahkan"
      />
      <StatCard
        label="Update (Ubah)"
        :value="counts.upd"
        icon="pi pi-sync"
        tone="info"
        hint="Modifikasi rekaman"
      />
      <StatCard
        label="Delete (Hapus)"
        :value="counts.del"
        icon="pi pi-trash"
        tone="bad"
        hint="Data dihapus dari sistem"
      />
    </div>

    <!-- Clean Unified Toolbar -->
    <div class="panel p-3 mb-4 flex flex-wrap items-center justify-between gap-2.5">
      <div class="flex flex-wrap items-center gap-2 flex-1 min-w-[280px]">
        <div class="relative flex-1 min-w-[220px] max-w-md">
          <i class="pi pi-search absolute left-3 top-1/2 -translate-y-1/2 text-ink-400 text-xs" />
          <InputText
            v-model="q"
            placeholder="Cari tabel, row ID, atau aksi…"
            class="w-full !pl-8 !text-[12px] !py-1.5"
          />
        </div>

        <Select
          v-model="opFilter"
          :options="opOptions"
          optionLabel="label"
          optionValue="value"
          placeholder="Semua Aksi"
          class="!text-[12px] !py-0.5 w-[180px]"
        />

        <Button
          v-if="q || opFilter"
          label="Reset"
          icon="pi pi-times"
          text
          size="small"
          severity="secondary"
          @click="q = ''; opFilter = ''"
        />
      </div>

      <div class="flex items-center gap-2">
        <Tag :value="filtered.length + ' ditampilkan'" severity="secondary" class="!text-[11px]" />
      </div>
    </div>

    <!-- Audit Table Panel with Pagination -->
    <Panel title="Jejak Mutasi Database &amp; Keamanan" icon="pi pi-shield" dense>
      <template #actions>
        <Tag :value="total + ' total log'" severity="secondary" class="!text-[11px]" />
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
            <tr class="text-left border-b" style="border-color: var(--line); background: var(--panel-2)">
              <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider w-[180px]" style="color: var(--txt-dim)">Waktu</th>
              <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider w-[170px]" style="color: var(--txt-dim)">Tabel Target</th>
              <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider w-[120px]" style="color: var(--txt-dim)">Aksi</th>
              <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider" style="color: var(--txt-dim)">Row ID / Kunci</th>
              <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider w-[80px] text-right" style="color: var(--txt-dim)">Detail</th>
            </tr>
          </thead>
          <tbody class="divide-y" style="border-color: var(--line)">
            <template v-for="(l, i) in filtered" :key="i">
              <tr
                class="hover:bg-slate-50 dark:hover:bg-slate-800/40 transition-colors cursor-pointer"
                @click="expanded = expanded === l.row_id + i ? null : l.row_id + i"
              >
                <td class="px-4 py-3 whitespace-nowrap text-[12px]" style="color: var(--txt-dim)">{{ l.at }}</td>
                <td class="px-4 py-3">
                  <span class="t-mono font-semibold text-indigo-600 dark:text-indigo-400 bg-indigo-50 dark:bg-indigo-950/40 border border-indigo-200 dark:border-indigo-800/60 px-2 py-0.5 rounded">
                    {{ l.table_name }}
                  </span>
                </td>
                <td class="px-4 py-3"><StatusChip :kind="l.op" /></td>
                <td class="px-4 py-3">
                  <span class="t-mono text-[11px]" style="color: var(--txt-dim)">{{ l.row_id || '—' }}</span>
                </td>
                <td class="px-4 py-3 text-right">
                  <i
                    class="pi text-[11px]"
                    :class="expanded === l.row_id + i ? 'pi-chevron-up' : 'pi-chevron-down'"
                    style="color: var(--txt-dim)"
                  />
                </td>
              </tr>

              <!-- JSON Diff Expanded Row -->
              <tr v-if="expanded === l.row_id + i">
                <td colspan="5" class="px-4 py-3.5" style="background: var(--panel-2)">
                  <div class="grid md:grid-cols-2 gap-3 pt-1">
                    <div class="flex flex-col gap-1.5">
                      <div class="text-[11px] font-semibold text-rose-600 dark:text-rose-400 uppercase tracking-wider flex items-center gap-1.5">
                        <i class="pi pi-minus-circle text-[10px]" />
                        Data Sebelumnya (Old Data)
                      </div>
                      <pre class="panel !rounded-lg p-3 t-mono text-[11px] overflow-x-auto max-h-[220px]" style="background: #172033; border-color: #27354a; color: #f8fafc">{{ prettyJSON(l.old_data) }}</pre>
                    </div>
                    <div class="flex flex-col gap-1.5">
                      <div class="text-[11px] font-semibold text-emerald-600 dark:text-emerald-400 uppercase tracking-wider flex items-center gap-1.5">
                        <i class="pi pi-plus-circle text-[10px]" />
                        Data Baru (New Data)
                      </div>
                      <pre class="panel !rounded-lg p-3 t-mono text-[11px] overflow-x-auto max-h-[220px]" style="background: #172033; border-color: #27354a; color: #f8fafc">{{ prettyJSON(l.new_data) }}</pre>
                    </div>
                  </div>
                </td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>

      <template #footer>
        <div class="flex flex-wrap items-center justify-between gap-3">
          <span class="text-[11.5px]" style="color: var(--txt-dim)">{{ range }}</span>
          <div class="flex items-center gap-2">
            <Select v-model="perPage" :options="perPageOptions" class="!text-[12px] !py-1 w-[95px]" />
            <Button
              icon="pi pi-angle-left"
              size="small"
              text
              severity="secondary"
              :disabled="page === 0"
              @click="page--"
            />
            <span class="t-num text-[12px] px-1 font-semibold" style="color: var(--txt)">Hal. {{ page + 1 }} / {{ lastPage + 1 }}</span>
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
      </template>
    </Panel>
  </div>
</template>
