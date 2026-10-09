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
const perPage = ref(20)
const perPageOptions = [10, 20, 50]

const filtered = computed(() =>
  rows.value.filter(r => !q.value || [r.name, r.id].join(' ').toLowerCase().includes(q.value.toLowerCase()))
)
const lastPage = computed(() => Math.max(0, Math.ceil(filtered.value.length / perPage.value) - 1))
const range = computed(() => {
  if (!filtered.value.length) return '0 kategori'
  const from = page.value * perPage.value + 1
  const to = Math.min(filtered.value.length, (page.value + 1) * perPage.value)
  return `${from}–${to} dari ${filtered.value.length} kategori`
})
const pagedRows = computed(() =>
  filtered.value.slice(page.value * perPage.value, (page.value + 1) * perPage.value)
)
watch([q, perPage], () => { page.value = 0 })

async function load() {
  loading.value = true
  try { rows.value = (await api.get('/categories')).data } finally { loading.value = false }
}
onMounted(load)

async function add() {
  const name = newName.value.trim()
  if (!name) return
  busy.value = true
  try {
    await api.post('/categories', { name })
    newName.value = ''
    toast.add({ severity: 'success', summary: 'Kategori ditambah', detail: name, life: 2200 })
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
    // POST upserts by slug; delete the old slug row when the name changes slug.
    const res = await api.post('/categories', { name })
    if (res.data?.id && res.data.id !== r.id) await api.delete(`/categories/${r.id}`)
    toast.add({ severity: 'success', summary: 'Kategori diperbarui', life: 2200 })
    cancelEdit(); load()
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Gagal', detail: e.response?.data?.error, life: 3500 })
  } finally { busy.value = false }
}

function remove(r: any) {
  confirm.require({
    message: `Hapus kategori "${r.name}"? Barang yang memakai kategori ini tidak ikut terhapus, tetapi akan kehilangan acuan kategori.`,
    header: 'Konfirmasi hapus kategori',
    icon: 'pi pi-exclamation-triangle',
    acceptClass: 'p-button-danger',
    accept: async () => {
      try {
        await api.delete(`/categories/${r.id}`)
        toast.add({ severity: 'success', summary: 'Kategori dihapus', life: 2200 })
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
    <PageHeader crumb="Master Data" title="Kategori"
      sub="Pengelompokan barang — dipakai sebagai prefiks pembuatan SKU otomatis" />

    <Panel title="Tambah Kategori" icon="pi pi-plus">
      <form @submit.prevent="add" class="flex flex-wrap gap-2.5 items-end">
        <label class="flex flex-col gap-1.5 flex-1 min-w-[240px]">
          <span class="text-[11.5px] font-semibold" style="color: var(--txt-dim)">Nama Kategori</span>
          <InputText v-model="newName" placeholder="mis. Peralatan Medis &amp; Alkes" class="w-full" />
        </label>
        <Button type="submit" label="Tambah" icon="pi pi-plus" :loading="busy" />
      </form>
      <p class="text-[11.5px] mt-2.5" style="color: var(--txt-dim)">
        ID kategori dibuat otomatis dari nama (slug) dan menjadi prefiks SKU, mis. <span class="t-mono">peralatan-medis</span> → <span class="t-mono">PERA-2026-001</span>.
      </p>
    </Panel>

    <div class="mt-4">
      <Panel title="Daftar Kategori" icon="pi pi-tags" dense>
        <template #actions>
          <InputText v-model="q" placeholder="Cari kategori…" class="!text-[12px] !py-1.5 w-[200px]" />
          <Tag severity="secondary" :value="rows.length + ' kategori'" />
        </template>

        <EmptyState v-if="loading" icon="pi pi-spin pi-spinner" title="Memuat kategori…" />
        <EmptyState v-else-if="!filtered.length" icon="pi pi-tags" title="Tidak ada kategori cocok"
                    sub="Ubah kata kunci pencarian atau buat kategori baru." />

        <ul v-else class="divide-y" style="border-color: var(--line)">
          <li v-for="r in pagedRows" :key="r.id" class="flex items-center gap-3 px-4 py-3.5 hover:bg-slate-50 dark:hover:bg-slate-800/40 transition-colors">
            <template v-if="editingId === r.id">
              <InputText v-model="editName" class="flex-1 !text-[12.5px]" @keyup.enter="saveEdit(r)"
                         @keyup.esc="cancelEdit" autofocus />
              <Button icon="pi pi-check" size="small" severity="success" :loading="busy" @click="saveEdit(r)" />
              <Button icon="pi pi-times" size="small" text severity="secondary" @click="cancelEdit" />
            </template>
            <template v-else>
              <div class="w-8 h-8 rounded-lg grid place-items-center border shrink-0 text-indigo-600 dark:text-indigo-400 bg-indigo-50 dark:bg-indigo-950/40 border-indigo-200 dark:border-indigo-800">
                <i class="pi pi-tag text-xs" />
              </div>
              <div class="flex-1 min-w-0">
                <div class="text-[13px] font-semibold" style="color: var(--txt)">{{ r.name }}</div>
                <div class="t-mono text-[11px]" style="color: var(--txt-dim)">{{ r.id }}</div>
              </div>
              <div class="flex items-center gap-1 shrink-0">
                <Button icon="pi pi-pencil" text rounded size="small" severity="secondary" @click="startEdit(r)" />
                <Button icon="pi pi-trash" text rounded size="small" severity="danger" @click="remove(r)" />
              </div>
            </template>
          </li>
        </ul>

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
