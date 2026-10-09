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
const total = ref(0)
const selected = ref<string[]>([])
const loading = ref(true)
const q = ref('')
const fmt = ref<'qr' | 'code128'>('qr')
const size = ref(160)
const locationFilter = ref('')
const locationsRaw = ref<any[]>([])

const page = ref(0)
const perPage = ref(24)
const perPageOptions = [12, 24, 48, 96]

const lastPage = computed(() => Math.max(0, Math.ceil(total.value / perPage.value) - 1))
const range = computed(() => {
  if (!total.value) return '0 barang'
  const from = page.value * perPage.value + 1
  const to = Math.min(total.value, (page.value + 1) * perPage.value)
  return `${from}–${to} dari ${total.value} barang`
})
const allCurrentSelected = computed(() =>
  items.value.length > 0 && items.value.every(i => selected.value.includes(i.id))
)

async function load() {
  loading.value = true
  try {
    const res = await api.get('/items', {
      params: {
        q: q.value,
        loc: locationFilter.value,
        page: page.value + 1,
        per_page: perPage.value,
      },
    })
    items.value = res.data.data || []
    total.value = res.data.total || 0
  } finally { loading.value = false }
}

onMounted(async () => {
  await load()
  const l = await api.get('/locations')
  locationsRaw.value = l.data
})

watch([page, perPage], load)
watch([q, locationFilter], () => {
  page.value = 0
  load()
})

function toggleAll() {
  const currentIds = items.value.map(i => i.id)
  if (allCurrentSelected.value) {
    selected.value = selected.value.filter(id => !currentIds.includes(id))
  } else {
    for (const id of currentIds) {
      if (!selected.value.includes(id)) selected.value.push(id)
    }
  }
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
        <Button :label="allCurrentSelected ? 'Batal pilih di halaman ini' : 'Pilih semua di halaman ini'"
                icon="pi pi-check-square" text size="small" @click="toggleAll" />
        <Tag :value="selected.length + ' dipilih'" severity="warn" />
      </div>
    </div>

    <EmptyState v-if="loading" icon="pi pi-spin pi-spinner" title="Memuat barang…" />
    <EmptyState v-else-if="!items.length" icon="pi pi-qrcode" title="Tidak ada barang cocok"
                sub="Ubah kata kunci atau filter lokasi untuk menemukan barang." />

    <div v-else class="grid gap-3.5" style="grid-template-columns: repeat(auto-fill, minmax(190px, 1fr))">
      <div
        v-for="item in items"
        :key="item.id"
        class="panel relative overflow-hidden transition-all duration-200 cursor-pointer hover:shadow-md"
        :class="selected.includes(item.id) ? '!border-indigo-500 ring-2 ring-indigo-500/20' : 'hover:border-slate-300 dark:hover:border-slate-600'"
        @click="selected.includes(item.id) ? selected.splice(selected.indexOf(item.id), 1) : selected.push(item.id)"
      >
        <span v-if="selected.includes(item.id)"
              class="absolute top-0 left-0 right-0 h-[3px] bg-indigo-600 dark:bg-indigo-400" />

        <div class="p-3 flex items-start justify-between">
          <Checkbox :modelValue="selected.includes(item.id)" binary @click.stop
                    @update:modelValue="() => selected.includes(item.id) ? selected.splice(selected.indexOf(item.id), 1) : selected.push(item.id)" />
          <Button icon="pi pi-download" text rounded size="small" severity="secondary"
                  v-tooltip.top="'Unduh satu label'" @click.stop="downloadOne(item)" />
        </div>

        <div class="px-3 pb-3 text-center">
          <!-- White sticker container for barcode scannability in all themes -->
          <div class="grid place-items-center rounded-lg border p-2 mb-2.5 shadow-xs"
               style="border-color: #e2e8f0; background: #ffffff">
            <img :src="`/api/barcode/${item.id}.png?fmt=${fmt}&size=${size}`"
                 class="max-w-full" :style="{ height: size * 0.62 + 'px' }" loading="lazy" />
          </div>
          <div class="text-[12.5px] font-semibold truncate" :title="item.name" style="color: var(--txt)">{{ item.name }}</div>
          <div class="t-mono text-[11px] mt-0.5" style="color: var(--txt-dim)">{{ item.sku }}</div>
          <div class="mt-2 flex items-center justify-center gap-1.5 flex-wrap">
            <Tag severity="secondary" :value="item.location" class="!text-[10px]" />
            <Tag severity="secondary" :value="item.unit" class="!text-[10px]" />
          </div>
        </div>
      </div>
    </div>

    <!-- pagination footer -->
    <div v-if="!loading && items.length" class="panel p-3 mt-4 flex flex-wrap items-center justify-between gap-3">
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

    <p v-if="!loading && items.length" class="text-[11.5px] mt-3" style="color: var(--txt-dim)">
      Lembar cetak dibuka di tab baru (format A4, 6 kolom). Untuk kualitas cetak terbaik pilih QR Code dan ukuran ≥ 160px.
    </p>
  </div>
</template>
