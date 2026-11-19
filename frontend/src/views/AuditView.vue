<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import api from '@/api'
import { useToast } from 'primevue/usetoast'
import PageHeader from '@/components/PageHeader.vue'
import Panel from '@/components/Panel.vue'
import StatCard from '@/components/StatCard.vue'
import StatusChip from '@/components/StatusChip.vue'
import EmptyState from '@/components/EmptyState.vue'
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
      toast.add({ severity: 'info', summary: 'Data keamanan diperbarui', life: 2000 })
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
    toast.add({ severity: 'success', summary: 'IP dibuka', detail: `Blokir untuk IP ${ip} telah dicabut`, life: 3000 })
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
  <div>
    <PageHeader
      crumb="Administrasi"
      title="Audit Log &amp; Keamanan"
      sub="Pemantauan integritas data basis data dan sistem pertahanan anti-serangan (burst attack &amp; brute-force)."
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

    <!-- Tab Switcher -->
    <div class="inline-flex items-center gap-1.5 p-1 rounded-[9px] mb-5 border border-line bg-panel-2">
      <button
        type="button"
        class="px-3.5 py-1.5 rounded-[7px] text-[12.5px] flex items-center gap-2 transition-colors cursor-pointer"
        :class="activeTab === 'db'
          ? 'font-semibold border border-[var(--acc-border)] bg-panel text-acc-500 shadow-[var(--shadow-sm)]'
          : 'text-txt-dim hover:text-txt border border-transparent'"
        @click="activeTab = 'db'"
      >
        <i class="pi pi-database text-[11.5px]" aria-hidden="true" />
        <span>Log Mutasi Database</span>
        <span class="t-mono !text-[10.5px] px-1.5 py-0.5 rounded bg-panel-2 border border-line text-txt-dim">{{ total }}</span>
      </button>

      <button
        type="button"
        class="px-3.5 py-1.5 rounded-[7px] text-[12.5px] flex items-center gap-2 transition-colors cursor-pointer"
        :class="activeTab === 'burst'
          ? 'font-semibold border border-[var(--sig-bad-line)] bg-panel text-sig-bad shadow-[var(--shadow-sm)]'
          : 'text-txt-dim hover:text-txt border border-transparent'"
        @click="activeTab = 'burst'"
      >
        <i class="pi pi-shield text-[11.5px]" aria-hidden="true" />
        <span>Deteksi Burst Attack &amp; Keamanan</span>
        <span
          v-if="blockedIPs.length > 0"
          class="t-mono !text-[10.5px] px-1.5 py-0.5 rounded bg-[var(--sig-bad-bg)] border border-[var(--sig-bad-line)] text-sig-bad font-bold"
        >
          {{ blockedIPs.length }} diblokir
        </span>
        <span v-else class="t-mono !text-[10.5px] px-1.5 py-0.5 rounded bg-panel-2 border border-line text-txt-dim">
          {{ securityEvents.length }}
        </span>
      </button>
    </div>

    <!-- TAB 1: DATABASE AUDIT LOG -->
    <div v-if="activeTab === 'db'">
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-3.5 mb-5">
        <StatCard
          label="Total Log Aktivitas"
          :value="counts.total"
          icon="pi pi-shield"
          tone="accent"
          featured
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
          hint="Perubahan entri eksisting"
        />
        <StatCard
          label="Delete (Hapus)"
          :value="counts.del"
          icon="pi pi-trash"
          tone="bad"
          hint="Penghapusan data sistem"
        />
      </div>

      <!-- Filter Toolbar -->
      <div class="panel p-3 mb-4 flex flex-wrap items-center gap-2.5">
        <div class="relative flex-1 min-w-[220px]">
          <i class="pi pi-search absolute left-3 top-1/2 -translate-y-1/2 text-txt-dim text-[12px]" aria-hidden="true" />
          <InputText
            v-model="q"
            placeholder="Filter nama tabel, ID baris, atau aksi pada halaman ini…"
            class="w-full !pl-8 !text-[12.5px]"
          />
        </div>
        <Select
          v-model="opFilter"
          :options="opOptions"
          optionLabel="label"
          optionValue="value"
          class="w-full sm:w-48 !text-[12.5px]"
        />
      </div>

      <Panel title="Jejak Perubahan Basis Data" icon="pi pi-database" dense>
        <template #actions>
          <span class="t-mono !text-[11px] text-txt-dim">{{ range }}</span>
        </template>

        <EmptyState v-if="loading" loading title="Memuat log audit…" />

        <EmptyState
          v-else-if="!filtered.length"
          icon="pi pi-database"
          title="Tidak ada catatan audit"
          sub="Belum ada perubahan data yang sesuai dengan filter."
        />

        <div v-else class="overflow-x-auto">
          <table class="w-full text-[12.5px] border-collapse">
            <thead>
              <tr class="text-left border-b border-line bg-panel-2 text-[11px] uppercase tracking-[0.05em] text-txt-dim font-semibold">
                <th class="px-3.5 py-2.5 w-[170px]">Waktu</th>
                <th class="px-3.5 py-2.5 w-[120px]">Operasi</th>
                <th class="px-3.5 py-2.5 w-[150px]">Tabel</th>
                <th class="px-3.5 py-2.5">Row ID</th>
                <th class="px-3.5 py-2.5 w-[90px] text-right">Detail</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-line">
              <template v-for="(l, idx) in filtered" :key="l.id || idx">
                <tr
                  class="data-row cursor-pointer"
                  :class="{ 'data-row-active': expanded === (l.id || String(idx)) }"
                  @click="expanded = expanded === (l.id || String(idx)) ? null : (l.id || String(idx))"
                >
                  <td class="px-3.5 py-2.5 t-mono !text-[11.5px] text-txt-dim whitespace-nowrap">
                    {{ l.changed_at || l.created_at }}
                  </td>
                  <td class="px-3.5 py-2.5"><StatusChip :kind="l.op" /></td>
                  <td class="px-3.5 py-2.5 t-mono font-semibold text-txt">{{ l.table_name }}</td>
                  <td class="px-3.5 py-2.5 t-mono !text-[11.5px] text-txt-dim truncate max-w-[260px]">{{ l.row_id || '—' }}</td>
                  <td class="px-3.5 py-2.5 text-right">
                    <i
                      class="pi text-[11px] text-txt-dim"
                      :class="expanded === (l.id || String(idx)) ? 'pi-chevron-up' : 'pi-chevron-down'"
                      aria-hidden="true"
                    />
                  </td>
                </tr>
                <tr v-if="expanded === (l.id || String(idx))" class="bg-panel-2">
                  <td colspan="5" class="p-4">
                    <div class="grid md:grid-cols-2 gap-3">
                      <div>
                        <div class="t-mono !text-[10.5px] uppercase tracking-[0.06em] text-txt-dim mb-1.5">Nilai Sebelum (old_data)</div>
                        <pre class="p-3 rounded-[7px] border border-line bg-panel t-mono !text-[11px] text-txt overflow-x-auto max-h-60">{{ prettyJSON(l.old_data) }}</pre>
                      </div>
                      <div>
                        <div class="t-mono !text-[10.5px] uppercase tracking-[0.06em] text-txt-dim mb-1.5">Nilai Sesudah (new_data)</div>
                        <pre class="p-3 rounded-[7px] border border-line bg-panel t-mono !text-[11px] text-txt overflow-x-auto max-h-60">{{ prettyJSON(l.new_data) }}</pre>
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
            <span class="t-mono !text-[11.5px] text-txt-dim">{{ range }}</span>
            <div class="flex items-center gap-2">
              <Select v-model="perPage" :options="perPageOptions" class="!text-[12px] w-[96px]" />
              <Button icon="pi pi-angle-left" size="small" text severity="secondary" aria-label="Halaman sebelumnya" :disabled="page === 0" @click="page--" />
              <span class="t-num text-[12px] px-1 text-txt">Hal. {{ page + 1 }} / {{ lastPage + 1 }}</span>
              <Button icon="pi pi-angle-right" size="small" text severity="secondary" aria-label="Halaman berikutnya" :disabled="page >= lastPage" @click="page++" />
            </div>
          </div>
        </template>
      </Panel>
    </div>

    <!-- TAB 2: SECURITY & BURST PROTECTION -->
    <div v-else class="space-y-5">
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-3.5">
        <StatCard
          label="Status Proteksi"
          :value="securityStats.protection_status"
          icon="pi pi-shield"
          tone="ok"
          featured
          :hint="`Burst ${securityStats.burst_limit} · Login ${securityStats.login_limit}`"
        />
        <StatCard
          label="IP Sedang Diblokir"
          :value="securityStats.blocked_ips_count"
          icon="pi pi-ban"
          :tone="securityStats.blocked_ips_count > 0 ? 'bad' : 'neutral'"
          hint="Karantina aktif otomatis"
        />
        <StatCard
          label="Deteksi Burst Attack"
          :value="securityStats.burst_attack_count"
          icon="pi pi-bolt"
          tone="warn"
          hint="Lonjakan request abnormal"
        />
        <StatCard
          label="Brute-Force Login"
          :value="securityStats.brute_force_count"
          icon="pi pi-key"
          tone="bad"
          hint="Percobaan sandi berulang"
        />
      </div>

      <!-- Blocked IPs -->
      <Panel title="IP Dalam Karantina Blokir Aktif" icon="pi pi-ban" dense>
        <template #actions>
          <span class="t-mono !text-[11px] text-txt-dim">{{ blockedIPs.length }} IP</span>
        </template>

        <EmptyState
          v-if="!blockedIPs.length"
          icon="pi pi-check-circle"
          tone="ok"
          title="Tidak ada IP yang sedang diblokir"
          sub="Semua lalu lintas permintaan saat ini berada dalam batas wajar."
        />

        <div v-else class="overflow-x-auto">
          <table class="w-full text-[12.5px] border-collapse">
            <thead>
              <tr class="text-left border-b border-line bg-panel-2 text-[11px] uppercase tracking-[0.05em] text-txt-dim font-semibold">
                <th class="px-3.5 py-2.5">Alamat IP</th>
                <th class="px-3.5 py-2.5">Alasan</th>
                <th class="px-3.5 py-2.5">Percobaan</th>
                <th class="px-3.5 py-2.5">Terakhir Terlihat</th>
                <th class="px-3.5 py-2.5">Sisa Blokir</th>
                <th class="px-3.5 py-2.5 text-right">Aksi</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-line">
              <tr v-for="b in blockedIPs" :key="b.ip" class="data-row">
                <td class="px-3.5 py-2.5 t-mono font-bold text-sig-bad">{{ b.ip }}</td>
                <td class="px-3.5 py-2.5"><StatusChip :kind="b.reason || 'BLOCKED'" /></td>
                <td class="px-3.5 py-2.5 t-mono text-txt">{{ b.attempts }} request</td>
                <td class="px-3.5 py-2.5 t-mono !text-[11.5px] text-txt-dim">{{ b.last_seen }}</td>
                <td class="px-3.5 py-2.5 t-mono font-semibold text-sig-warn">{{ b.expires_in_sec }} detik</td>
                <td class="px-3.5 py-2.5 text-right">
                  <Button label="Buka Blokir" icon="pi pi-lock-open" size="small" severity="success" outlined @click="unblockIP(b.ip)" />
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </Panel>

      <!-- Security Events History -->
      <Panel title="Riwayat Insiden Serangan Terdeteksi" icon="pi pi-history" dense>
        <template #actions>
          <span class="t-mono !text-[11px] text-txt-dim">{{ securityEvents.length }} insiden</span>
        </template>

        <EmptyState
          v-if="!securityEvents.length"
          icon="pi pi-shield"
          title="Belum ada insiden serangan yang terekam"
          sub="Setiap deteksi burst request abnormal atau percobaan brute-force login dicatat otomatis di tabel ini."
        />

        <div v-else class="overflow-x-auto">
          <table class="w-full text-[12.5px] border-collapse">
            <thead>
              <tr class="text-left border-b border-line bg-panel-2 text-[11px] uppercase tracking-[0.05em] text-txt-dim font-semibold">
                <th class="px-3.5 py-2.5 w-[165px]">Waktu</th>
                <th class="px-3.5 py-2.5 w-[145px]">IP Penyerang</th>
                <th class="px-3.5 py-2.5 w-[160px]">Jenis Deteksi</th>
                <th class="px-3.5 py-2.5">Target Endpoint</th>
                <th class="px-3.5 py-2.5 w-[130px]">Status</th>
                <th class="px-3.5 py-2.5">Keterangan &amp; User Agent</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-line">
              <tr v-for="ev in securityEvents" :key="ev.id" class="data-row">
                <td class="px-3.5 py-2.5 whitespace-nowrap t-mono !text-[11.5px] text-txt-dim">{{ ev.time_str || ev.created_at }}</td>
                <td class="px-3.5 py-2.5 t-mono font-semibold text-sig-bad">{{ ev.ip }}</td>
                <td class="px-3.5 py-2.5">
                  <StatusChip :kind="ev.event_type === 'BURST_ATTACK' ? 'BURST_ATTACK' : 'BRUTE_FORCE'" />
                </td>
                <td class="px-3.5 py-2.5 t-mono !text-[11.5px]">
                  <span class="px-1.5 py-0.5 rounded bg-panel-2 border border-line font-bold mr-1.5 text-[10px] text-txt-dim">{{ ev.method }}</span>
                  <span class="text-txt">{{ ev.endpoint }}</span>
                </td>
                <td class="px-3.5 py-2.5">
                  <StatusChip kind="BLOCKED" :label="ev.status || 'BLOCKED (429)'" />
                </td>
                <td class="px-3.5 py-2.5 text-[11.5px]">
                  <div class="font-medium text-txt">{{ ev.details }}</div>
                  <div class="t-mono !text-[10.5px] truncate max-w-xs text-txt-dim">{{ ev.user_agent || '—' }}</div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </Panel>
    </div>
  </div>
</template>
