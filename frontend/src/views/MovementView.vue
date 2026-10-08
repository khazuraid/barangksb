<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import api from '@/api'
import { useRouter } from 'vue-router'
import { useToast } from 'primevue/usetoast'
import Button from 'primevue/button'
import Select from 'primevue/select'
import InputNumber from 'primevue/inputnumber'
import InputText from 'primevue/inputtext'
import Textarea from 'primevue/textarea'
import PageHeader from '@/components/PageHeader.vue'
import Panel from '@/components/Panel.vue'
import Field from '@/components/Field.vue'
import EmptyState from '@/components/EmptyState.vue'
import Tag from 'primevue/tag'
import PhotoUploader from '@/components/PhotoUploader.vue'

type Mode = 'in' | 'out' | 'adjust'

const router = useRouter()
const toast = useToast()

const MODES = [
  { key: 'in' as Mode, label: 'Barang Masuk', icon: 'pi pi-arrow-down', desc: 'Penerimaan dari pengadaan, hibah, atau pengembalian.' },
  { key: 'out' as Mode, label: 'Barang Keluar', icon: 'pi pi-arrow-up', desc: 'Pengeluaran untuk pemakaian, peminjaman, atau pemusnahan.' },
  { key: 'adjust' as Mode, label: 'Opname / Koreksi', icon: 'pi pi-sliders-h', desc: 'Sesuaikan stok sistem dengan hasil hitung fisik.' },
]

const mode = ref<Mode>('in')
const items = ref<any[]>([])
const loadingItems = ref(true)
const submitting = ref(false)
const search = ref('')

const form = ref({ item_id: '', quantity: 1, received_by: '', notes: '', photo_url: '' })
const selectedItem = computed(() => items.value.find(i => i.id === form.value.item_id))

async function loadItems() {
  loadingItems.value = true
  try {
    const res = await api.get('/items', { params: { per_page: 100 } })
    items.value = res.data.data
  } finally { loadingItems.value = false }
}
onMounted(loadItems)

const options = computed(() =>
  items.value
    .filter(i => !search.value || `${i.name} ${i.sku} ${i.location}`.toLowerCase().includes(search.value.toLowerCase()))
    .map(i => ({ label: `${i.name} — ${i.sku} (stok ${i.current_stock} ${i.unit})`, value: i.id }))
)

const current = computed(() => MODES.find(m => m.key === mode.value)!)

const preview = computed(() => {
  if (!selectedItem.value) return null
  const cur = selectedItem.value.current_stock
  if (mode.value === 'adjust') return { from: cur, to: form.value.quantity, unit: selectedItem.value.unit }
  const to = mode.value === 'in' ? cur + form.value.quantity : cur - form.value.quantity
  return { from: cur, to, unit: selectedItem.value.unit }
})

function setMode(m: Mode) {
  mode.value = m
  form.value.quantity = m === 'adjust' ? (selectedItem.value?.current_stock ?? 0) : 1
}

async function submit() {
  if (!form.value.item_id) {
    toast.add({ severity: 'warn', summary: 'Barang belum dipilih', life: 2500 })
    return
  }
  if (mode.value !== 'adjust' && form.value.quantity <= 0) {
    toast.add({ severity: 'warn', summary: 'Jumlah harus lebih dari 0', life: 2500 })
    return
  }
  submitting.value = true
  const endpoint = mode.value === 'in' ? '/movement/in' : mode.value === 'out' ? '/movement/out' : '/movement/adjust'
  try {
    const res = await api.post(endpoint, form.value)
    const d = res.data
    toast.add({
      severity: 'success', life: 3500,
      summary: 'Mutasi tercatat',
      detail: `${d.name || ''} — ${d.previous} → ${d.new}`,
    })
    router.push('/history')
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Gagal mencatat', detail: e.response?.data?.error, life: 4000 })
  } finally { submitting.value = false }
}
</script>

<template>
  <div>
    <PageHeader crumb="Operasional" title="Mutasi Barang"
      sub="Setiap pencatatan langsung mengubah stok dan terekam pada riwayat serta audit log" />

    <div class="grid lg:grid-cols-[260px_1fr] gap-4 items-start">
      <!-- mode rail -->
      <div class="flex lg:flex-col gap-2 overflow-x-auto lg:overflow-visible">
        <button
          v-for="m in MODES" :key="m.key"
          @click="setMode(m.key)"
          class="text-left p-3.5 rounded-lg border transition-all shrink-0 lg:shrink lg:w-full min-w-[210px]"
          :class="mode === m.key
            ? 'border-acc-500 bg-acc-50'
            : 'hover:border-ink-300'"
          :style="mode === m.key ? '' : 'border-color: var(--line); background: var(--panel)'"
        >
          <div class="flex items-center gap-2">
            <i :class="m.icon" class="text-[13px]" :style="{ color: mode === m.key ? 'var(--p-primary-600)' : 'var(--txt-dim)' }" />
            <span class="text-[12.5px] font-bold">{{ m.label }}</span>
          </div>
          <p class="text-[11.5px] mt-1.5 leading-snug" style="color: var(--txt-dim)">{{ m.desc }}</p>
        </button>
      </div>

      <!-- form -->
      <div class="grid xl:grid-cols-[1fr_300px] gap-4 items-start">
        <Panel :title="current.label" :icon="current.icon">
          <form @submit.prevent="submit" class="grid sm:grid-cols-2 gap-4">
            <Field label="Barang" required span>
              <Select v-model="form.item_id" :options="options" optionLabel="label" optionValue="value"
                      filter :filterFields="['label']" :loading="loadingItems" showClear
                      placeholder="Pilih barang…" class="w-full" @filter="(e: any) => search = e.value" />
            </Field>

            <Field :label="mode === 'adjust' ? 'Stok Fisik Hasil Hitung' : (mode === 'in' ? 'Jumlah Masuk' : 'Jumlah Keluar')"
                   :required="mode !== 'adjust'" :hint="mode === 'adjust' ? 'Isi jumlah nyata hasil penghitungan fisik.' : undefined">
              <InputNumber v-model="form.quantity" :min="mode === 'adjust' ? 0 : 1" showButtons class="w-full" />
            </Field>

            <Field :label="mode === 'out' ? 'Diberikan Kepada' : 'Petugas / Penerima'">
              <InputText v-model="form.received_by" placeholder="Nama petugas" class="w-full" />
            </Field>

            <Field label="Keterangan" span>
              <Textarea v-model="form.notes" rows="2" autoResize class="w-full"
                        :placeholder="mode === 'adjust' ? 'Alasan koreksi (mis. selisih opname)' : 'Catatan tambahan'" />
            </Field>

            <div class="sm:col-span-2">
              <PhotoUploader
                v-model="form.photo_url"
                :location-name="selectedItem?.location || ''"
                label="Foto Bukti &amp; Cap Geotag"
                hint="Cap GPS lokasi, waktu, dan nama petugas otomatis tertempel pada foto"
              />
            </div>

            <div class="sm:col-span-2 flex gap-2.5 pt-1">
              <Button type="submit" :label="mode === 'in' ? 'Catat Masuk' : mode === 'out' ? 'Catat Keluar' : 'Simpan Opname'"
                      :icon="current.icon" :loading="submitting"
                      :severity="mode === 'in' ? 'success' : mode === 'out' ? 'danger' : 'warn'" />
              <Button label="Lihat Riwayat" icon="pi pi-history" text severity="secondary"
                      @click="router.push('/history')" />
            </div>
          </form>
        </Panel>

        <!-- preview -->
        <Panel title="Pratinjau Perubahan" icon="pi pi-eye">
          <EmptyState v-if="!preview" icon="pi pi-arrow-left" title="Belum ada barang dipilih"
                      sub="Perubahan stok akan ditampilkan di sini sebelum disimpan." />
          <template v-else>
            <div class="text-[12.5px] font-semibold mb-3">{{ selectedItem?.name }}</div>
            <div class="flex items-center justify-between gap-3">
              <div class="text-center">
                <div class="t-label">Sebelum</div>
                <div class="t-num text-[24px] font-bold mt-1">{{ preview.from }}</div>
              </div>
              <i class="pi pi-arrow-right text-[13px]" style="color: var(--txt-dim)" />
              <div class="text-center">
                <div class="t-label">Sesudah</div>
                <div class="t-num text-[24px] font-bold mt-1"
                     :class="preview.to < preview.from ? 'text-sig-bad' : preview.to > preview.from ? 'text-sig-ok' : ''">
                  {{ preview.to }}
                </div>
              </div>
            </div>
            <div class="mt-3.5 pt-3.5 border-t flex items-center justify-between text-[12px]"
                 style="border-color: var(--line)">
              <span style="color: var(--txt-dim)">Selisih</span>
              <span class="t-num font-bold"
                    :class="preview.to - preview.from < 0 ? 'text-sig-bad' : 'text-sig-ok'">
                {{ preview.to - preview.from > 0 ? '+' : '' }}{{ preview.to - preview.from }} {{ preview.unit }}
              </span>
            </div>

            <Tag v-if="mode === 'out' && preview.to < 0" severity="danger"
                 class="w-full justify-center mt-3.5"
                 icon="pi pi-exclamation-triangle" value="Stok tidak mencukupi" />
            <Tag v-else-if="mode === 'out' && preview.to <= (selectedItem?.min_stock ?? 0)"
                 severity="warn" class="w-full justify-center mt-3.5"
                 icon="pi pi-exclamation-triangle" value="Akan menyentuh batas minimum" />
          </template>
        </Panel>
      </div>
    </div>
  </div>
</template>
