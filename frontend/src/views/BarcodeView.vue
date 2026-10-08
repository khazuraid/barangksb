<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import api from '@/api'
import { useToast } from 'primevue/usetoast'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Checkbox from 'primevue/checkbox'
import PageHeader from '@/components/PageHeader.vue'
import Panel from '@/components/Panel.vue'
import EmptyState from '@/components/EmptyState.vue'
import Tag from 'primevue/tag'

const toast = useToast()
const items = ref<any[]>([])
const selected = ref<string[]>([])
const loading = ref(true)
const q = ref('')
const fmt = ref<'qr' | 'code128'>('qr')
const size = ref(160)
const categories = ref<any[]>([])
const locationFilter = ref('')
const categoriesRaw = ref<any[]>([])
const locationsRaw = ref<any[]>([])

async function load() {
  loading.value = true
  try {
    const res = await api.get('/items', { params: { per_page: 100 } })
    items.value = res.data.data
  } finally { loading.value = false }
}

onMounted(async () => {
  await load()
  const [c, l] = await Promise.all([api.get('/categories'), api.get('/locations')])
  categoriesRaw.value = c.data
  locationsRaw.value = l.data
})

const locFiltered = computed(() =>
  items.value.filter(i =>
    (!q.value || [i.name, i.sku, i.location].join(' ').toLowerCase().includes(q.value.toLowerCase())) &&
    (!locationFilter.value || i.location === locationFilter.value)
  )
)

function toggleAll() {
  selected.value = selected.value.length === locFiltered.value.length ? [] : locFiltered.value.map(i => i.id)
}

function print() {
  if (!selected.value.length) {
    toast.add({ severity: 'warn', summary: 'Belum ada label dipilih', detail: 'Centang minimal satu barang.', life: 2500 })
    return
  }
  window.open(`/api/barcode/sheet?ids=${selected.value.join(',')}`, '_blank')
}

function downloadOne(item: any) {
  const a = document.createElement('a')
  a.href = `/api/barcode/${item.id}.png?fmt=${fmt.value}`
  a.download = `${item.sku}.png`
  a.click()
}
</script>

<template>
  <div>
    <PageHeader crumb="Perangkat" title="Generator Label QR / Barcode"
      sub="Pilih barang, atur format, lalu cetak lembar label siap tempel">
      <template #actions>
        <Button label="Cetak Lembar Terpilih" icon="pi pi-print" size="small"
                :disabled="!selected.length" @click="print" />
      </template>
    </PageHeader>

    <!-- controls -->
    <div class="panel p-3.5 mb-4 flex flex-wrap gap-3 items-end">
      <label class="flex flex-col gap-1.5 flex-1 min-w-[220px]">
        <span class="text-[11.5px] font-semibold" style="color: var(--txt-dim)">Cari Barang</span>
        <InputText v-model="q" placeholder="Nama, SKU, atau lokasi…" class="w-full !text-[12.5px]" />
      </label>
      <label class="flex flex-col gap-1.5 min-w-[180px]">
        <span class="text-[11.5px] font-semibold" style="color: var(--txt-dim)">Lokasi</span>
        <Select v-model="locationFilter" :options="locationsRaw" optionLabel="name" optionValue="name"
                showClear placeholder="Semua lokasi" class="!text-[12.5px]" />
      </label>
      <label class="flex flex-col gap-1.5 min-w-[130px]">
        <span class="text-[11.5px] font-semibold" style="color: var(--txt-dim)">Format</span>
        <Select v-model="fmt" :options="[{label:'QR Code',value:'qr'},{label:'Barcode 128',value:'code128'}]"
                optionLabel="label" optionValue="value" class="!text-[12.5px]" />
      </label>
      <label class="flex flex-col gap-1.5 w-[120px]">
        <span class="text-[11.5px] font-semibold" style="color: var(--txt-dim)">Ukuran (px)</span>
        <InputText v-model.number="size" type="number" min="80" max="400" class="!text-[12.5px]" />
      </label>
      <div class="flex items-center gap-3 pb-1">
        <Button :label="selected.length === locFiltered.length && locFiltered.length ? 'Batal pilih semua' : 'Pilih semua'"
                icon="pi pi-check-square" text size="small" @click="toggleAll" />
        <Tag :value="selected.length + ' dipilih'" severity="warn" />
      </div>
    </div>

    <EmptyState v-if="loading" icon="pi pi-spin pi-spinner" title="Memuat barang…" />
    <EmptyState v-else-if="!locFiltered.length" icon="pi pi-qrcode" title="Tidak ada barang cocok"
                sub="Ubah kata kunci atau filter lokasi untuk menemukan barang." />

    <div v-else class="grid gap-3" style="grid-template-columns: repeat(auto-fill, minmax(180px, 1fr))">
      <div
        v-for="item in locFiltered"
        :key="item.id"
        class="panel relative overflow-hidden transition-all cursor-pointer"
        :class="selected.includes(item.id) ? 'ring-2 ring-acc-500' : 'hover:border-ink-300'"
        @click="selected.includes(item.id) ? selected.splice(selected.indexOf(item.id), 1) : selected.push(item.id)"
      >
        <span v-if="selected.includes(item.id)"
              class="absolute top-0 left-0 right-0 h-[3px] bg-acc-500" />

        <div class="p-3 flex items-start justify-between">
          <Checkbox :modelValue="selected.includes(item.id)" binary @click.stop
                    @update:modelValue="() => selected.includes(item.id) ? selected.splice(selected.indexOf(item.id), 1) : selected.push(item.id)" />
          <Button icon="pi pi-download" text rounded size="small" severity="secondary"
                  v-tooltip.top="'Unduh satu label'" @click.stop="downloadOne(item)" />
        </div>

        <div class="px-3 pb-3 text-center">
          <div class="grid place-items-center rounded-md border p-2 mb-2.5"
               style="border-color: var(--line-soft); background: var(--paper-1)">
            <img :src="`/api/barcode/${item.id}.png?fmt=${fmt}&size=${size}`"
                 class="max-w-full" :style="{ height: size * 0.62 + 'px' }" loading="lazy" />
          </div>
          <div class="text-[12.5px] font-semibold truncate" :title="item.name">{{ item.name }}</div>
          <div class="t-mono mt-0.5" style="color: var(--txt-dim)">{{ item.sku }}</div>
          <div class="mt-1.5 flex items-center justify-center gap-1.5">
            <Tag severity="secondary" :value="item.location" />
            <Tag severity="secondary" :value="item.unit" />
          </div>
        </div>
      </div>
    </div>

    <p v-if="!loading && locFiltered.length" class="text-[11.5px] mt-3" style="color: var(--txt-dim)">
      Lembar cetak dibuka di tab baru (format A4, 6 kolom). Untuk kualitas cetak terbaik pilih QR Code dan ukuran ≥ 160px.
    </p>
  </div>
</template>
