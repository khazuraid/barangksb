<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import api from '@/api'
import { useConfirm } from 'primevue/useconfirm'
import { useToast } from 'primevue/usetoast'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import PageHeader from '@/components/PageHeader.vue'
import Panel from '@/components/Panel.vue'
import EmptyState from '@/components/EmptyState.vue'
import Tag from 'primevue/tag'

const confirm = useConfirm()
const toast = useToast()
const rows = ref<any[]>([])
const loading = ref(true)
const newName = ref('')
const busy = ref(false)
const editingId = ref<string | null>(null)
const editName = ref('')
const q = ref('')

const page = ref(0)
const perPage = ref(18)
const perPageOptions = [12, 18, 30, 60]

const filtered = computed(() =>
  rows.value.filter(r => !q.value || [r.name, r.id].join(' ').toLowerCase().includes(q.value.toLowerCase()))
)
const lastPage = computed(() => Math.max(0, Math.ceil(filtered.value.length / perPage.value) - 1))
const range = computed(() => {
  if (!filtered.value.length) return '0 lokasi'
  const from = page.value * perPage.value + 1
  const to = Math.min(filtered.value.length, (page.value + 1) * perPage.value)
  return `${from}–${to} dari ${filtered.value.length} lokasi`
})
const pagedRows = computed(() =>
  filtered.value.slice(page.value * perPage.value, (page.value + 1) * perPage.value)
)
watch([q, perPage], () => { page.value = 0 })

async function load() {
  loading.value = true
  try { rows.value = (await api.get('/locations')).data } finally { loading.value = false }
}
onMounted(load)

async function add() {
  const name = newName.value.trim()
  if (!name) return
  busy.value = true
  try {
    await api.post('/locations', { name })
    newName.value = ''
    toast.add({ severity: 'success', summary: 'Lokasi ditambah', detail: name, life: 2200 })
    load()
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Gagal', detail: e.response?.data?.error, life: 3500 })
  } finally { busy.value = false }
}

function startEdit(r: any) { editingId.value = r.id; editName.value = r.name }
function cancelEdit() { editingId.value = null; editName.value = '' }

async function saveEdit(r: any) {
  const name = editName.value.trim()
  if (!name || name === r.name) return cancelEdit()
  busy.value = true
  try {
    const res = await api.post('/locations', { name })
    if (res.data?.id && res.data.id !== r.id) await api.delete(`/locations/${r.id}`)
    toast.add({ severity: 'success', summary: 'Lokasi diperbarui', life: 2200 })
    cancelEdit(); load()
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Gagal', detail: e.response?.data?.error, life: 3500 })
  } finally { busy.value = false }
}

function remove(r: any) {
  confirm.require({
    message: `Hapus lokasi "${r.name}"? Barang yang berada di lokasi ini tidak ikut terhapus, tetapi akan kehilangan acuan lokasi.`,
    header: 'Konfirmasi hapus lokasi',
    icon: 'pi pi-exclamation-triangle',
    acceptClass: 'p-button-danger',
    accept: async () => {
      try {
        await api.delete(`/locations/${r.id}`)
        toast.add({ severity: 'success', summary: 'Lokasi dihapus', life: 2200 })
        load()
      } catch (e: any) {
        toast.add({ severity: 'error', summary: 'Gagal', detail: e.response?.data?.error, life: 3500 })
      }
    },
  })
}
</script>

<template>
  <div>
    <PageHeader crumb="Master Data" title="Lokasi"
      sub="Ruangan, rak, dan titik penyimpanan barang" />

    <Panel title="Tambah Lokasi" icon="pi pi-plus">
      <form @submit.prevent="add" class="flex flex-wrap gap-2.5 items-end">
        <label class="flex flex-col gap-1.5 flex-1 min-w-[240px]">
          <span class="text-[11.5px] font-semibold" style="color: var(--txt-dim)">Nama Lokasi</span>
          <InputText v-model="newName" placeholder="mis. Gudang Lantai 2 / Ruang Farmasi" class="w-full" />
        </label>
        <Button type="submit" label="Tambah" icon="pi pi-plus" :loading="busy" />
      </form>
      <p class="text-[11.5px] mt-2.5" style="color: var(--txt-dim)">
        Gunakan penamaan hierarkis (Lantai → Ruangan → Rak) agar pencarian lokasi tetap rapi saat data bertambah.
      </p>
    </Panel>

    <div class="mt-4">
      <Panel title="Daftar Lokasi" icon="pi pi-map-marker" dense>
        <template #actions>
          <InputText v-model="q" placeholder="Cari lokasi…" class="!text-[12px] !py-1.5 w-[200px]" />
          <Tag severity="secondary" :value="rows.length + ' lokasi'" />
        </template>

        <EmptyState v-if="loading" icon="pi pi-spin pi-spinner" title="Memuat lokasi…" />
        <EmptyState v-else-if="!filtered.length" icon="pi pi-map-marker" title="Tidak ada lokasi cocok"
                    sub="Ubah kata kunci pencarian atau tambah lokasi baru." />

        <div v-else class="grid sm:grid-cols-2 lg:grid-cols-3 gap-3 p-3">
          <div v-for="r in pagedRows" :key="r.id"
               class="p-3.5 rounded-xl border flex items-start gap-3 transition-all hover:shadow-xs"
               style="border-color: var(--line); background: var(--panel)">
            <span class="w-8 h-8 shrink-0 grid place-items-center rounded-lg border text-sky-600 dark:text-sky-400 bg-sky-50 dark:bg-sky-950/40 border-sky-200 dark:border-sky-800">
              <i class="pi pi-map-marker text-[13px]" />
            </span>

            <template v-if="editingId === r.id">
              <div class="flex-1 min-w-0 flex items-center gap-1.5">
                <InputText v-model="editName" class="flex-1 !text-[12.5px]" @keyup.enter="saveEdit(r)"
                           @keyup.esc="cancelEdit" autofocus />
                <Button icon="pi pi-check" size="small" severity="success" :loading="busy" @click="saveEdit(r)" />
                <Button icon="pi pi-times" size="small" text severity="secondary" @click="cancelEdit" />
              </div>
            </template>
            <template v-else>
              <div class="flex-1 min-w-0">
                <div class="text-[13px] font-semibold truncate" style="color: var(--txt)">{{ r.name }}</div>
                <div class="t-mono text-[11px] truncate mt-0.5" style="color: var(--txt-dim)">{{ r.id }}</div>
              </div>
              <div class="flex gap-0.5 shrink-0">
                <Button icon="pi pi-pencil" text rounded size="small" severity="secondary" @click="startEdit(r)" />
                <Button icon="pi pi-trash" text rounded size="small" severity="danger" @click="remove(r)" />
              </div>
            </template>
          </div>
        </div>

        <template v-if="filtered.length" #footer>
          <div class="flex flex-wrap items-center justify-between gap-3">
            <span class="text-[11.5px]" style="color: var(--txt-dim)">{{ range }}</span>
            <div class="flex items-center gap-2">
              <Select v-model="perPage" :options="perPageOptions" class="!text-[12px] !py-1 w-[95px]" />
              <Button icon="pi pi-angle-left" size="small" text severity="secondary"
                      :disabled="page === 0" @click="page--" />
              <span class="t-num text-[12px] px-1">Hal. {{ page + 1 }} / {{ lastPage + 1 }}</span>
              <Button icon="pi pi-angle-right" size="small" text severity="secondary"
                      :disabled="page >= lastPage" @click="page++" />
            </div>
          </div>
        </template>
      </Panel>
    </div>
  </div>
</template>
