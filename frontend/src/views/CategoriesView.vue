<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api'
import { useConfirm } from 'primevue/useconfirm'
import { useToast } from 'primevue/usetoast'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
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
          <Tag severity="secondary" value="{{ rows.length }} kategori" />
        </template>

        <EmptyState v-if="loading" icon="pi pi-spin pi-spinner" title="Memuat kategori…" />
        <EmptyState v-else-if="!rows.length" icon="pi pi-tags" title="Belum ada kategori"
                    sub="Buat kategori pertama untuk mulai mendata barang." />

        <ul v-else class="divide-y" style="border-color: var(--line-soft)">
          <li v-for="r in rows" :key="r.id" class="flex items-center gap-3 px-4 py-3 hover:bg-paper-2 transition-colors">
            <template v-if="editingId === r.id">
              <InputText v-model="editName" class="flex-1 !text-[12.5px]" @keyup.enter="saveEdit(r)"
                         @keyup.esc="cancelEdit" autofocus />
              <Button icon="pi pi-check" size="small" severity="success" :loading="busy" @click="saveEdit(r)" />
              <Button icon="pi pi-times" size="small" text severity="secondary" @click="cancelEdit" />
            </template>
            <template v-else>
              <div class="flex-1 min-w-0">
                <div class="text-[13px] font-semibold">{{ r.name }}</div>
                <div class="t-mono" style="color: var(--txt-dim)">{{ r.id }}</div>
              </div>
              <Button icon="pi pi-pencil" text rounded size="small" severity="secondary" @click="startEdit(r)" />
              <Button icon="pi pi-trash" text rounded size="small" severity="danger" @click="remove(r)" />
            </template>
          </li>
        </ul>
      </Panel>
    </div>
  </div>
</template>
