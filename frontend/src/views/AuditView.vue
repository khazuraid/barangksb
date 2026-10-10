<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import api from '@/api'
import { useToast } from 'primevue/usetoast'
import PageHeader from '@/components/PageHeader.vue'
import Panel from '@/components/Panel.vue'
import StatCard from '@/components/StatCard.vue'
import StatusChip from '@/components/StatusChip.vue'
import EmptyState from '@/components/EmptyState.vue'
import Tag from 'primevue/tag'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'

const toast = useToast()

const activeTab = ref<'db' | 'burst'>('db')

// --- Database Audit State ---
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

// --- Burst Attack & Security State ---
const securityLoading = ref(false)
const securityEvents = ref<any[]>([])
const blockedIPs = ref<any[]>([])
const securityStats = ref({
  total_incidents: 0,
  blocked_ips_count: 0,
  brute_force_count: 0,
  burst_attack_count: 0,
  protection_status: 'AKTIF',
  login_limit: '5 perc./menit',
  burst_limit: '35 req./5 detik',
})

async function fetchSecurity(showToast = false) {
  securityLoading.value = true
  try {
    const res = await api.get('/audit/security')
    securityEvents.value = res.data?.events || []
    blockedIPs.value = res.data?.blocked_ips || []
    if (res.data?.stats) {
      securityStats.value = res.data.stats
    }
    if (showToast) {
      toast.add({ severity: 'info', summary: 'Data Keamanan Diperbarui', life: 2000 })
    }
  } catch (err: any) {
    if (showToast) {
      toast.add({ severity: 'error', summary: 'Gagal memuat log keamanan', detail: err.message, life: 3000 })
    }
  } finally {
    securityLoading.value = false
  }
}

async function unblockIP(ip: string) {
  try {
    await api.post('/audit/security/unblock', { ip })
    toast.add({ severity: 'success', summary: 'IP Berhasil Dibuka', detail: `Blokir untuk IP ${ip} telah dicabut`, life: 3000 })
    fetchSecurity()
  } catch (err: any) {
    toast.add({ severity: 'error', summary: 'Gagal membuka blokir IP', detail: err.response?.data?.error || err.message, life: 3000 })
  }
}

async function clearEvents() {
  try {
    await api.delete('/audit/security/events')
    toast.add({ severity: 'info', summary: 'Log dibersihkan', detail: 'Riwayat insiden keamanan telah direset', life: 2500 })
    fetchSecurity()
  } catch (err: any) {
    toast.add({ severity: 'error', summary: 'Gagal membersihkan log', detail: err.message, life: 3000 })
  }
}

onMounted(() => {
  fetchAudit()
  fetchSecurity()
})

watch([page, perPage], () => fetchAudit())
watch(activeTab, (tab) => {
  if (tab === 'burst') fetchSecurity()
})

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
      sub="Pemantauan integritas data basis data dan sistem pertahanan anti-serangan (burst attack &amp; brute-force)"
    >
      <template #actions>
        <Button
          v-if="activeTab === 'db'"
          label="Segarkan Log"
          icon="pi pi-refresh"
          size="small"
          severity="secondary"
          outlined
          :loading="loading"
          @click="fetchAudit(true)"
        />
        <div v-else class="flex items-center gap-2">
          <Button
            label="Segarkan Status"
            icon="pi pi-refresh"
            size="small"
            severity="secondary"
            outlined
            :loading="securityLoading"
            @click="fetchSecurity(true)"
          />
          <Button
            v-if="securityEvents.length > 0"
            label="Bersihkan Riwayat"
            icon="pi pi-trash"
            size="small"
            severity="danger"
            text
            @click="clearEvents"
          />
        </div>
      </template>
    </PageHeader>

    <!-- Tab Switcher Navigation -->
    <div class="flex items-center gap-2 p-1 rounded-xl mb-4 w-fit border" style="background: var(--panel-2); border-color: var(--line)">
      <button
        @click="activeTab = 'db'"
        class="px-4 py-2 rounded-lg text-[13px] flex items-center gap-2 transition-all cursor-pointer"
        :class="activeTab === 'db'
          ? 'font-bold shadow-xs border text-indigo-600 dark:text-indigo-400 bg-white dark:bg-slate-900 border-indigo-200 dark:border-indigo-800'
          : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 border-transparent'"
      >
        <i class="pi pi-database text-xs" />
        <span>Log Mutasi Database</span>
        <Tag :value="total" severity="secondary" class="!text-[10.5px] !py-0.5" />
      </button>

      <button
        @click="activeTab = 'burst'"
        class="px-4 py-2 rounded-lg text-[13px] flex items-center gap-2 transition-all cursor-pointer relative"
        :class="activeTab === 'burst'
          ? 'font-bold shadow-xs border text-rose-600 dark:text-rose-400 bg-white dark:bg-slate-900 border-rose-200 dark:border-rose-800'
          : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 border-transparent'"
      >
        <i class="pi pi-shield text-xs" />
        <span>Deteksi Burst Attack &amp; Keamanan</span>
        <span
          v-if="blockedIPs.length > 0"
          class="px-1.5 py-0.5 rounded-full text-[10.5px] font-bold bg-rose-500 text-white animate-pulse"
        >
          {{ blockedIPs.length }} Diblokir
        </span>
        <Tag v-else :value="securityEvents.length" severity="secondary" class="!text-[10.5px] !py-0.5" />
      </button>
    </div>

    <!-- TAB 1: DATABASE AUDIT LOG -->
    <div v-if="activeTab === 'db'">
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
      <Panel title="Jejak Mutasi Database" icon="pi pi-database" dense>
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

    <!-- TAB 2: BURST ATTACK & SECURITY DEFENSE -->
    <div v-else class="space-y-4">
      <!-- Security Status Banner -->
      <div class="panel p-4 flex flex-wrap items-center justify-between gap-4 border-l-4 !border-l-emerald-500" style="background: var(--panel-2)">
        <div class="flex items-center gap-3.5">
          <div class="w-10 h-10 rounded-xl grid place-items-center bg-emerald-50 dark:bg-emerald-950/40 border border-emerald-200 dark:border-emerald-800 text-emerald-600 dark:text-emerald-400 text-base">
            <i class="pi pi-shield" />
          </div>
          <div>
            <div class="flex items-center gap-2">
              <span class="text-[14px] font-bold" style="color: var(--txt)">Perlindungan Burst Attack &amp; Anti Brute-Force</span>
              <span class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-emerald-500/20 text-emerald-600 dark:text-emerald-400 border border-emerald-500/30">
                ● AKTIF
              </span>
            </div>
            <p class="text-[12px] mt-0.5 leading-relaxed" style="color: var(--txt-dim)">
              Membatasi flood request berlebih (ambang: <strong>{{ securityStats.burst_limit }}</strong>) dan menahan serangan brute-force login (ambang: <strong>{{ securityStats.login_limit }}</strong>).
            </p>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <Button
            label="Segarkan Status"
            icon="pi pi-refresh"
            size="small"
            severity="secondary"
            outlined
            :loading="securityLoading"
            @click="fetchSecurity(true)"
          />
        </div>
      </div>

      <!-- KPI Summary Cards -->
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-3">
        <StatCard
          label="Total Insiden Terdeteksi"
          :value="securityStats.total_incidents"
          icon="pi pi-exclamation-triangle"
          :tone="securityStats.total_incidents > 0 ? 'bad' : 'accent'"
          hint="Serangan burst & brute force"
        />
        <StatCard
          label="IP Sedang Dibatasi"
          :value="securityStats.blocked_ips_count"
          icon="pi pi-ban"
          :tone="securityStats.blocked_ips_count > 0 ? 'bad' : 'ok'"
          hint="IP yang ditahan sementara"
        />
        <StatCard
          label="Serangan Brute Force"
          :value="securityStats.brute_force_count"
          icon="pi pi-key"
          tone="info"
          hint="Percobaan login berulang"
        />
        <StatCard
          label="Lonjakan Burst Request"
          :value="securityStats.burst_attack_count"
          icon="pi pi-bolt"
          tone="accent"
          hint="Flood request abnormal"
        />
      </div>

      <!-- Section: Currently Blocked / Throttled IPs -->
      <Panel title="Daftar IP Sedang Ditahan / Diblokir Sementara" icon="pi pi-ban" dense>
        <template #actions>
          <Tag :value="blockedIPs.length + ' IP'" :severity="blockedIPs.length ? 'danger' : 'secondary'" class="!text-[11px]" />
        </template>

        <EmptyState
          v-if="!blockedIPs.length"
          icon="pi pi-check-circle"
          title="Tidak ada IP yang sedang diblokir"
          sub="Seluruh lalu lintas jaringan berada dalam ambang batas normal. Sistem terus memantau setiap request secara otomatis."
        />

        <div v-else class="overflow-x-auto">
          <table class="w-full text-[12.5px]">
            <thead>
              <tr class="text-left border-b" style="border-color: var(--line); background: var(--panel-2)">
                <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider w-[180px]" style="color: var(--txt-dim)">IP Address</th>
                <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider w-[180px]" style="color: var(--txt-dim)">Tipe Pelanggaran</th>
                <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider w-[140px]" style="color: var(--txt-dim)">Jumlah Hit</th>
                <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider w-[140px]" style="color: var(--txt-dim)">Terakhir Dilihat</th>
                <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider w-[140px]" style="color: var(--txt-dim)">Sisa Waktu Blokir</th>
                <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider text-right" style="color: var(--txt-dim)">Tindakan Admin</th>
              </tr>
            </thead>
            <tbody class="divide-y" style="border-color: var(--line)">
              <tr v-for="b in blockedIPs" :key="b.ip" class="hover:bg-slate-50 dark:hover:bg-slate-800/40">
                <td class="px-4 py-3 font-mono font-bold text-rose-600 dark:text-rose-400">{{ b.ip }}</td>
                <td class="px-4 py-3">
                  <Tag :value="b.type" severity="danger" class="!text-[10.5px]" />
                </td>
                <td class="px-4 py-3 font-mono">{{ b.attempts }} request</td>
                <td class="px-4 py-3 text-[12px]" style="color: var(--txt-dim)">{{ b.last_seen }}</td>
                <td class="px-4 py-3">
                  <span class="font-mono text-[11.5px] font-bold text-amber-500">
                    {{ b.expires_in_sec }} detik
                  </span>
                </td>
                <td class="px-4 py-3 text-right">
                  <Button
                    label="Buka Blokir"
                    icon="pi pi-lock-open"
                    size="small"
                    severity="success"
                    outlined
                    @click="unblockIP(b.ip)"
                  />
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </Panel>

      <!-- Section: Security Events Incident History -->
      <Panel title="Riwayat Insiden Serangan Terdeteksi" icon="pi pi-history" dense>
        <template #actions>
          <Tag :value="securityEvents.length + ' insiden'" severity="secondary" class="!text-[11px]" />
        </template>

        <EmptyState
          v-if="!securityEvents.length"
          icon="pi pi-shield"
          title="Belum ada insiden serangan yang terekam"
          sub="Setiap deteksi burst request abnormal atau percobaan brute-force login akan dicatat secara otomatis di tabel ini."
        />

        <div v-else class="overflow-x-auto">
          <table class="w-full text-[12.5px]">
            <thead>
              <tr class="text-left border-b" style="border-color: var(--line); background: var(--panel-2)">
                <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider w-[170px]" style="color: var(--txt-dim)">Waktu</th>
                <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider w-[150px]" style="color: var(--txt-dim)">IP Penyerang</th>
                <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider w-[180px]" style="color: var(--txt-dim)">Jenis Deteksi</th>
                <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider" style="color: var(--txt-dim)">Target Endpoint</th>
                <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider w-[130px]" style="color: var(--txt-dim)">Status</th>
                <th class="px-4 py-3 font-semibold text-[11px] uppercase tracking-wider" style="color: var(--txt-dim)">Keterangan &amp; User Agent</th>
              </tr>
            </thead>
            <tbody class="divide-y" style="border-color: var(--line)">
              <tr v-for="ev in securityEvents" :key="ev.id" class="hover:bg-slate-50 dark:hover:bg-slate-800/40">
                <td class="px-4 py-3 whitespace-nowrap text-[12px]" style="color: var(--txt-dim)">{{ ev.time_str || ev.created_at }}</td>
                <td class="px-4 py-3 font-mono font-semibold text-rose-600 dark:text-rose-400">{{ ev.ip }}</td>
                <td class="px-4 py-3">
                  <Tag
                    :value="ev.event_type === 'BURST_ATTACK' ? 'BURST ATTACK' : 'BRUTE FORCE LOGIN'"
                    :severity="ev.event_type === 'BURST_ATTACK' ? 'warn' : 'danger'"
                    class="!text-[10px]"
                  />
                </td>
                <td class="px-4 py-3 font-mono text-[11.5px]">
                  <span class="px-1.5 py-0.5 rounded bg-slate-100 dark:bg-slate-800 font-bold mr-1.5 text-[10px]">{{ ev.method }}</span>
                  <span style="color: var(--txt)">{{ ev.endpoint }}</span>
                </td>
                <td class="px-4 py-3">
                  <Tag :value="ev.status || 'BLOCKED (429)'" severity="danger" class="!text-[10px]" />
                </td>
                <td class="px-4 py-3 text-[11.5px]">
                  <div class="font-medium" style="color: var(--txt)">{{ ev.details }}</div>
                  <div class="text-[10.5px] font-mono truncate max-w-xs" style="color: var(--txt-dim)">
                    {{ ev.user_agent || '—' }}
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </Panel>
    </div>
  </div>
</template>
