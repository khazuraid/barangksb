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
import Tag from 'primevue/tag'

const logs = ref<any[]>([])
const loading = ref(true)
const q = ref('')
const opFilter = ref('')
const expanded = ref<string | null>(null)

const opOptions = [
  { label: 'Semua aksi', value: '' },
  { label: 'INSERT', value: 'INSERT' },
  { label: 'UPDATE', value: 'UPDATE' },
  { label: 'DELETE', value: 'DELETE' },
]

onMounted(async () => {
  try { logs.value = (await api.get('/audit')).data } finally { loading.value = false }
})

const filtered = computed(() =>
  logs.value.filter(l =>
    (!opFilter.value || l.op === opFilter.value) &&
    (!q.value || [l.table_name, l.row_id, l.op].join(' ').toLowerCase().includes(q.value.toLowerCase()))
  )
)

const counts = computed(() => ({
  total: logs.value.length,
  ins: logs.value.filter(l => l.op === 'INSERT').length,
  upd: logs.value.filter(l => l.op === 'UPDATE').length,
  del: logs.value.filter(l => l.op === 'DELETE').length,
}))

function pretty(raw: any) {
  if (!raw) return '—'
  try { return JSON.stringify(typeof raw === 'string' ? JSON.parse(raw) : raw, null, 2) } catch { return String(raw) }
}
</script>

<template>
  <div>
    <PageHeader crumb="Administrasi" title="Audit Log"
      sub="Rekaman otomatis setiap perubahan pada tabel barang dan transaksi (maksimal 200 terbaru)">
      <template #actions>
        <Button label="Muat ulang" icon="pi pi-refresh" size="small" severity="secondary" outlined
                @click="() => api.get('/audit').then(r => logs = r.data)" />
      </template>
    </PageHeader>

    <!-- summary strip -->
    <div class="grid grid-cols-2 sm:grid-cols-4 gap-px mb-4 rounded-lg overflow-hidden border"
         style="border-color: var(--line); background: var(--line)">
      <div v-for="c in [
        { k: 'TOTAL', v: counts.total, cls: '' },
        { k: 'INSERT', v: counts.ins, cls: 'text-sig-ok' },
        { k: 'UPDATE', v: counts.upd, cls: 'text-sig-info' },
        { k: 'DELETE', v: counts.del, cls: 'text-sig-bad' },
      ]" :key="c.k" class="px-4 py-3" style="background: var(--panel)">
        <div class="t-label">{{ c.k }}</div>
        <div class="t-num text-[20px] font-bold mt-1" :class="c.cls">{{ c.v }}</div>
      </div>
    </div>

    <Panel title="Jejak Perubahan" icon="pi pi-shield" dense>
      <template #actions>
        <InputText v-model="q" placeholder="Cari tabel atau row id…" class="!text-[12px] !py-1.5 w-[190px]" />
        <Select v-model="opFilter" :options="opOptions" optionLabel="label" optionValue="value"
                class="!text-[12px] w-[140px]" />
      </template>

      <EmptyState v-if="loading" icon="pi pi-spin pi-spinner" title="Memuat audit log…" />
      <EmptyState v-else-if="!filtered.length" icon="pi pi-shield" title="Tidak ada catatan"
                  sub="Audit log terisi otomatis saat data barang atau transaksi berubah." />

      <div v-else class="overflow-x-auto">
        <table class="w-full text-[12.5px]">
          <thead>
            <tr class="text-left" style="background: var(--paper-2)">
              <th class="t-label px-4 py-2.5 w-[180px]">Waktu</th>
              <th class="t-label px-4 py-2.5 w-[170px]">Tabel</th>
              <th class="t-label px-4 py-2.5 w-[110px]">Aksi</th>
              <th class="t-label px-4 py-2.5">Row ID</th>
              <th class="t-label px-4 py-2.5 w-[90px]"></th>
            </tr>
          </thead>
          <tbody class="divide-y" style="border-color: var(--line-soft)">
            <template v-for="(l, i) in filtered" :key="i">
              <tr class="hover:bg-paper-2 transition-colors cursor-pointer" @click="expanded = expanded === l.row_id + i ? null : l.row_id + i">
                <td class="px-4 py-2.5 whitespace-nowrap" style="color: var(--txt-dim)">{{ l.at }}</td>
                <td class="px-4 py-2.5">
                  <span class="t-mono font-semibold">{{ l.table_name }}</span>
                </td>
                <td class="px-4 py-2.5"><StatusChip :kind="l.op" /></td>
                <td class="px-4 py-2.5"><span class="t-mono" style="color: var(--txt-dim)">{{ l.row_id || '—' }}</span></td>
                <td class="px-4 py-2.5 text-right">
                  <i class="pi text-[10px]" :class="expanded === l.row_id + i ? 'pi-chevron-up' : 'pi-chevron-down'"
                     style="color: var(--txt-dim)" />
                </td>
              </tr>
              <tr v-if="expanded === l.row_id + i">
                <td colspan="5" class="px-4 pb-3 pt-0" style="background: var(--paper-1)">
                  <div class="grid md:grid-cols-2 gap-3">
                    <div>
                      <div class="t-label mb-1.5">Data Lama</div>
                      <pre class="panel !rounded-md p-2.5 t-mono overflow-x-auto max-h-[220px]">{{ pretty(l.old_data) }}</pre>
                    </div>
                    <div>
                      <div class="t-label mb-1.5">Data Baru</div>
                      <pre class="panel !rounded-md p-2.5 t-mono overflow-x-auto max-h-[220px]">{{ pretty(l.new_data) }}</pre>
                    </div>
                  </div>
                </td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>

      <template #footer>
        <div class="flex items-center justify-between">
          <span class="text-[11.5px]" style="color: var(--txt-dim)">
            Menampilkan {{ filtered.length }} dari {{ logs.length }} catatan
          </span>
          <span class="text-[11.5px]" style="color: var(--txt-dim)">
            Klik baris untuk melihat payload JSON
          </span>
        </div>
      </template>
    </Panel>
  </div>
</template>
