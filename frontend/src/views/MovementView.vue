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
const photoUploaderRef = ref<any>(null)

const form = ref({ item_id: '', quantity: 1, received_by: '', notes: '', photo_url: '' })
const selectedItem = computed(() => items.value.find(i => i.id === form.value.item_id))

async function loadItems() {
  loadingItems.value = true
  try {
    const res = await api.get('/items', { params: { per_page: 5000 } })
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
    if (photoUploaderRef.value?.hasPendingPhoto) {
      const uploadedUrl = await photoUploaderRef.value.uploadPending()
      if (uploadedUrl) {
        form.value.photo_url = uploadedUrl
      }
    }
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
    <PageHeader
      crumb="Operasional"
      title="Mutasi Barang"
      sub="Setiap pencatatan langsung mengubah stok dan terekam pada riwayat serta audit log."
    />

    <div class="grid lg:grid-cols-[260px_1fr] gap-4 items-start">
      <!-- mode rail -->
      <div class="flex lg:flex-col gap-2 overflow-x-auto lg:overflow-visible" role="tablist" aria-label="Jenis mutasi">
        <button
          v-for="m in MODES"
          :key="m.key"
          type="button"
          role="tab"
          :aria-selected="mode === m.key"
          class="text-left p-3.5 rounded-[10px] border transition-colors shrink-0 lg:shrink lg:w-full min-w-[216px] relative overflow-hidden cursor-pointer"
          :class="mode === m.key
            ? 'border-[var(--acc-border)] bg-[var(--acc-wash)]'
            : 'border-line bg-panel hover:bg-panel-2'"
          @click="setMode(m.key)"
        >
          <span
            v-if="mode === m.key"
            class="absolute top-0 left-0 bottom-0 w-[3px] bg-acc-500"
            aria-hidden="true"
          />
          <div class="flex items-center gap-2.5">
            <span
              class="w-7 h-7 rounded-[7px] grid place-items-center text-[12px] border shrink-0 transition-colors"
              :class="mode === m.key
                ? (m.key === 'in'
                    ? 'bg-[var(--sig-ok-bg)] text-sig-ok border-[var(--sig-ok-line)]'
                    : m.key === 'out'
                      ? 'bg-[var(--sig-bad-bg)] text-sig-bad border-[var(--sig-bad-line)]'
                      : 'bg-[var(--sig-warn-bg)] text-sig-warn border-[var(--sig-warn-line)]')
                : 'bg-panel-2 text-txt-dim border-line'"
            >
              <i :class="m.icon" aria-hidden="true" />
            </span>
            <span
              class="text-[13px] font-semibold"
              :class="mode === m.key ? 'text-acc-500' : 'text-txt'"
            >
              {{ m.label }}
            </span>
          </div>
          <p class="text-[11.5px] text-txt-dim mt-2 leading-relaxed">{{ m.desc }}</p>
        </button>
      </div>

      <!-- form + preview -->
      <div class="grid xl:grid-cols-[1fr_320px] gap-4 items-start">
        <Panel :title="current.label" :icon="current.icon">
          <form @submit.prevent="submit" class="grid sm:grid-cols-2 gap-4">
            <Field label="Barang" required span>
              <Select
                v-model="form.item_id"
                :options="options"
                optionLabel="label"
                optionValue="value"
                filter
                :filterFields="['label']"
                :loading="loadingItems"
                showClear
                placeholder="Pilih barang…"
                class="w-full"
                @filter="(e: any) => search = e.value"
              />
            </Field>

            <Field
              :label="mode === 'adjust' ? 'Stok Fisik Hasil Hitung' : (mode === 'in' ? 'Jumlah Masuk' : 'Jumlah Keluar')"
              :required="mode !== 'adjust'"
              :hint="mode === 'adjust' ? 'Isi jumlah nyata hasil penghitungan fisik.' : undefined"
            >
              <InputNumber v-model="form.quantity" :min="mode === 'adjust' ? 0 : 1" showButtons class="w-full" />
            </Field>

            <Field :label="mode === 'out' ? 'Diberikan Kepada' : 'Petugas / Penerima'">
              <InputText v-model="form.received_by" placeholder="Nama petugas" class="w-full" />
            </Field>

            <Field label="Keterangan" span>
              <Textarea
                v-model="form.notes"
                rows="2"
                autoResize
                class="w-full"
                :placeholder="mode === 'adjust' ? 'Alasan koreksi (mis. selisih opname)' : 'Catatan tambahan'"
              />
            </Field>

            <div class="sm:col-span-2">
              <PhotoUploader
                ref="photoUploaderRef"
                :defer-upload="true"
                v-model="form.photo_url"
                :location-name="selectedItem?.location || ''"
                label="Foto Bukti &amp; Cap Geotag"
                hint="Cap GPS lokasi, waktu, dan nama petugas otomatis tertempel pada foto"
              />
            </div>

            <div class="sm:col-span-2 flex flex-wrap gap-2.5 pt-3 border-t border-line">
              <Button
                type="submit"
                :label="mode === 'in' ? 'Catat Masuk' : mode === 'out' ? 'Catat Keluar' : 'Simpan Opname'"
                :icon="current.icon"
                :loading="submitting"
                :severity="mode === 'in' ? 'success' : mode === 'out' ? 'danger' : 'warn'"
              />
              <Button
                type="button"
                label="Lihat Riwayat"
                icon="pi pi-history"
                text
                severity="secondary"
                @click="router.push('/history')"
              />
            </div>
          </form>
        </Panel>

        <!-- preview -->
        <Panel title="Pratinjau Perubahan" icon="pi pi-eye">
          <EmptyState
            v-if="!preview"
            icon="pi pi-arrow-left"
            title="Belum ada barang dipilih"
            sub="Pilih barang dan jumlah untuk melihat simulasi perubahan stok."
          />
          <template v-else>
            <div class="p-3 rounded-[8px] border border-line bg-panel-2 mb-3">
              <div class="text-[13px] font-semibold text-txt truncate">{{ selectedItem?.name }}</div>
              <div class="t-mono !text-[11px] text-txt-dim mt-0.5 truncate">
                {{ selectedItem?.sku }} · {{ selectedItem?.location }}
              </div>
            </div>

            <div class="grid grid-cols-2 gap-2 p-3 rounded-[8px] border border-line bg-panel-2 text-center">
              <div>
                <div class="t-label">Sebelum</div>
                <div class="t-num text-[22px] font-bold text-txt mt-0.5">{{ preview.from }}</div>
                <div class="text-[11px] text-txt-dim">{{ preview.unit }}</div>
              </div>
              <div class="border-l border-line">
                <div class="t-label">Sesudah</div>
                <div
                  class="t-num text-[22px] font-bold mt-0.5"
                  :class="preview.to < preview.from ? 'text-sig-bad' : preview.to > preview.from ? 'text-sig-ok' : 'text-txt'"
                >
                  {{ preview.to }}
                </div>
                <div class="text-[11px] text-txt-dim">{{ preview.unit }}</div>
              </div>
            </div>

            <div class="mt-3 p-3 rounded-[8px] border border-line bg-panel-2 flex items-center justify-between text-[12.5px]">
              <span class="text-txt-dim">Estimasi Selisih</span>
              <span
                class="t-num font-bold text-[13px]"
                :class="preview.to - preview.from < 0 ? 'text-sig-bad' : preview.to - preview.from > 0 ? 'text-sig-ok' : 'text-txt-dim'"
              >
                {{ preview.to - preview.from > 0 ? '+' : '' }}{{ preview.to - preview.from }} {{ preview.unit }}
              </span>
            </div>

            <Tag
              v-if="mode === 'out' && preview.to < 0"
              severity="danger"
              class="w-full justify-center mt-3"
              icon="pi pi-exclamation-triangle"
              value="Peringatan: Stok tidak mencukupi"
            />
            <Tag
              v-else-if="mode === 'out' && preview.to <= (selectedItem?.min_stock ?? 0)"
              severity="warn"
              class="w-full justify-center mt-3"
              icon="pi pi-exclamation-triangle"
              value="Perhatian: Menembus batas minimum"
            />
          </template>
        </Panel>
      </div>
    </div>
  </div>
</template>
