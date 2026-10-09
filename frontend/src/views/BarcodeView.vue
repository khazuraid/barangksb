<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import api from '@/api'
import { useToast } from 'primevue/usetoast'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Checkbox from 'primevue/checkbox'
import Tag from 'primevue/tag'
import Dialog from 'primevue/dialog'
import PageHeader from '@/components/PageHeader.vue'
import StatCard from '@/components/StatCard.vue'
import EmptyState from '@/components/EmptyState.vue'

const toast = useToast()
const items = ref<any[]>([])
const total = ref(0)
const selected = ref<string[]>([])
const loading = ref(true)

const page = ref(0)
const perPage = ref(24)
const perPageOptions = [12, 24, 48, 96]

const q = ref('')
const categoryFilter = ref('')
const locationFilter = ref('')
const fmt = ref<'qr' | 'code128'>('qr')
const size = ref(160)

// Thermal Printer & Sticker Label Profiles
const layout = ref('label_50x30')
const layoutOptions = [
  { label: '🏷️ Stiker Label 50 x 30 mm (Standar Printer Barcode)', value: 'label_50x30' },
  { label: '🏷️ Stiker Label 40 x 30 mm (Mini Aset)', value: 'label_40x30' },
  { label: '🏷️ Stiker Label 60 x 40 mm (Sedang)', value: 'label_60x40' },
  { label: '🏷️ Stiker Label 70 x 50 mm (Besar)', value: 'label_70x50' },
  { label: '🏷️ Stiker Label 100 x 50 mm (Stiker Aset Lebar)', value: 'label_100x50' },
  { label: '🖨️ Printer Thermal Roll 58 mm (Struk / Kasir 58mm)', value: 'thermal_58' },
  { label: '🖨️ Printer Thermal Roll 80 mm (Struk / Lebar 80mm)', value: 'thermal_80' },
  { label: '📄 Lembar Kertas A4 (Grid 3 x 7 · 21 Label)', value: 'a4_3x7' },
  { label: '📄 Lembar Kertas A4 (Grid 4 x 10 · 40 Label Mini)', value: 'a4_4x10' },
]

const showName = ref(true)
const showSKU = ref(true)
const showLoc = ref(true)
const labelTitle = ref('INVENTARIS KANTOR')
const printModalVisible = ref(false)

const categoriesRaw = ref<any[]>([])
const locationsRaw = ref<any[]>([])

async function fetchItems(reset = false) {
  if (reset) page.value = 0
  loading.value = true
  try {
    const res = await api.get('/items', {
      params: {
        q: q.value,
        cat: categoryFilter.value,
        loc: locationFilter.value,
        page: page.value + 1,
        per_page: perPage.value,
      },
    })
    items.value = res.data?.data || []
    total.value = res.data?.total || 0
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await fetchItems()
  const [c, l] = await Promise.all([api.get('/categories'), api.get('/locations')])
  categoriesRaw.value = c.data || []
  locationsRaw.value = l.data || []
})

watch([page, perPage], () => fetchItems())

const lastPage = computed(() => Math.max(0, Math.ceil(total.value / perPage.value) - 1))
const range = computed(() => {
  if (!total.value) return '0 barang'
  const from = page.value * perPage.value + 1
  const to = Math.min(total.value, (page.value + 1) * perPage.value)
  return `${from}–${to} dari ${total.value}`
})

function toggleAllOnPage() {
  const currentPageIds = items.value.map((i) => i.id)
  const allSelected = currentPageIds.every((id) => selected.value.includes(id))
  if (allSelected) {
    selected.value = selected.value.filter((id) => !currentPageIds.includes(id))
  } else {
    const toAdd = currentPageIds.filter((id) => !selected.value.includes(id))
    selected.value.push(...toAdd)
  }
}

const isAllOnPageSelected = computed(() => {
  if (!items.value.length) return false
  return items.value.every((i) => selected.value.includes(i.id))
})

function getPrintUrl(autoPrint = true) {
  const params = new URLSearchParams({
    ids: selected.value.join(','),
    layout: layout.value,
    fmt: fmt.value,
    title: labelTitle.value || 'INVENTARIS KANTOR',
    show_name: showName.value ? '1' : '0',
    show_sku: showSKU.value ? '1' : '0',
    show_loc: showLoc.value ? '1' : '0',
  })
  if (!autoPrint) {
    params.set('noprint', '1')
  }
  return `/api/barcode/sheet?${params.toString()}`
}

function openPrintDialog() {
  if (!selected.value.length) {
    toast.add({
      severity: 'warn',
      summary: 'Belum ada label dipilih',
      detail: 'Centang minimal satu barang inventaris untuk dicetak.',
      life: 2500,
    })
    return
  }
  printModalVisible.value = true
}

function printNow() {
  const url = getPrintUrl(true)
  window.open(url, '_blank')
}

function openSheetTab() {
  const url = getPrintUrl(false)
  window.open(url, '_blank')
}

function downloadOne(item: any) {
  const a = document.createElement('a')
  a.href = `/api/barcode/${item.id}.png?fmt=${fmt.value}`
  a.download = `${item.sku}.png`
  a.click()
}

// First selected item for live preview in dialog
const sampleItem = computed(() => {
  if (selected.value.length) {
    const found = items.value.find((i) => i.id === selected.value[0])
    if (found) return found
  }
  return items.value[0] || {
    id: 'sample',
    name: 'Laptop Lenovo ThinkPad E14',
    sku: 'ELK-2026-001',
    location: 'Ruang Server IT',
  }
})
</script>

<template>
  <div class="pb-16 w-full">
    <PageHeader
      crumb="Perangkat"
      title="Generator Label QR &amp; Barcode"
      sub="Cetak label aset siap tempel untuk printer thermal roll (58mm/80mm), stiker label berbagai ukuran (50x30, 40x30, 60x40, 100x50), dan kertas lembar A4"
    >
      <template #actions>
        <Button
          label="Pratinjau &amp; Cetak Label"
          icon="pi pi-print"
          size="small"
          :disabled="!selected.length"
          @click="openPrintDialog"
        />
      </template>
    </PageHeader>

    <!-- KPI Metric Strip -->
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-3 mb-4">
      <StatCard
        label="Total Master Barang"
        :value="total"
        icon="pi pi-box"
        tone="neutral"
        hint="Barang inventaris tercatat"
      />
      <StatCard
        label="Label Dipilih"
        :value="selected.length"
        icon="pi pi-check-square"
        tone="accent"
        :hint="selected.length ? selected.length + ' label siap dicetak' : 'Centang barang di bawah'"
      />
      <StatCard
        label="Profil Printer / Label"
        :value="layout.startsWith('thermal') ? 'Printer Thermal' : layout.startsWith('a4') ? 'Kertas A4' : 'Stiker Label'"
        icon="pi pi-print"
        tone="info"
        :hint="layoutOptions.find(o => o.value === layout)?.label?.slice(2) || layout"
      />
    </div>

    <!-- Unified Toolbar -->
    <div class="panel p-3 mb-4 flex flex-wrap items-center justify-between gap-2.5">
      <div class="flex flex-wrap items-center gap-2 flex-1 min-w-[280px]">
        <div class="relative flex-1 min-w-[190px] max-w-sm">
          <i class="pi pi-search absolute left-3 top-1/2 -translate-y-1/2 text-ink-400 text-xs" />
          <InputText
            v-model="q"
            placeholder="Cari nama atau SKU…"
            class="w-full !pl-8 !text-[12px] !py-1.5"
            @keyup.enter="fetchItems(true)"
          />
        </div>

        <Select
          v-model="categoryFilter"
          :options="categoriesRaw"
          optionLabel="name"
          optionValue="name"
          showClear
          placeholder="Semua Kategori"
          class="!text-[12px] !py-0.5 w-[160px]"
          @change="fetchItems(true)"
          filter
        />

        <Select
          v-model="locationFilter"
          :options="locationsRaw"
          optionLabel="name"
          optionValue="name"
          showClear
          placeholder="Semua Lokasi"
          class="!text-[12px] !py-0.5 w-[160px]"
          @change="fetchItems(true)"
          filter
        />

        <Select
          v-model="fmt"
          :options="[
            { label: 'QR Code', value: 'qr' },
            { label: 'Barcode 128', value: 'code128' },
          ]"
          optionLabel="label"
          optionValue="value"
          class="!text-[12px] !py-0.5 w-[130px]"
        />

        <Select
          v-model="layout"
          :options="layoutOptions"
          optionLabel="label"
          optionValue="value"
          class="!text-[12px] !py-0.5 w-[230px]"
          v-tooltip.top="'Pilih ukuran label / jenis printer'"
        />

        <Button
          icon="pi pi-search"
          size="small"
          severity="secondary"
          outlined
          @click="fetchItems(true)"
        />

        <Button
          v-if="q || categoryFilter || locationFilter"
          label="Reset"
          icon="pi pi-times"
          text
          size="small"
          severity="secondary"
          @click="q = ''; categoryFilter = ''; locationFilter = ''; fetchItems(true)"
        />
      </div>

      <div class="flex items-center gap-2">
        <Button
          :label="isAllOnPageSelected ? 'Batal pilih hal. ini' : 'Pilih semua di hal. ini'"
          icon="pi pi-check-square"
          text
          size="small"
          severity="secondary"
          @click="toggleAllOnPage"
        />
        <Tag :value="selected.length + ' dipilih'" severity="warn" class="!text-[11px]" />
        <Button
          label="Cetak"
          icon="pi pi-print"
          size="small"
          :disabled="!selected.length"
          @click="openPrintDialog"
        />
      </div>
    </div>

    <EmptyState v-if="loading" icon="pi pi-spin pi-spinner" title="Memuat barcode barang…" />
    <EmptyState
      v-else-if="!items.length"
      icon="pi pi-qrcode"
      title="Tidak ada barang cocok"
      sub="Ubah kata kunci atau filter lokasi untuk menemukan barang."
    />

    <!-- Items Barcode Grid -->
    <div v-else class="flex flex-col gap-4">
      <div class="grid gap-3" style="grid-template-columns: repeat(auto-fill, minmax(200px, 1fr))">
        <div
          v-for="it in items"
          :key="it.id"
          class="panel relative overflow-hidden transition-all cursor-pointer rounded-xl border flex flex-col justify-between hover:shadow-md"
          :class="selected.includes(it.id) ? 'ring-2 ring-indigo-500 border-indigo-500' : 'hover:border-slate-400 dark:hover:border-slate-600'"
          style="background: var(--paper-1); border-color: var(--line)"
          @click="selected.includes(it.id) ? selected.splice(selected.indexOf(it.id), 1) : selected.push(it.id)"
        >
          <span
            v-if="selected.includes(it.id)"
            class="absolute top-0 left-0 right-0 h-[3px] bg-indigo-500"
          />

          <div class="p-2.5 flex items-start justify-between">
            <Checkbox
              :modelValue="selected.includes(it.id)"
              binary
              @click.stop
              @update:modelValue="() => selected.includes(it.id) ? selected.splice(selected.indexOf(it.id), 1) : selected.push(it.id)"
            />
            <Button
              icon="pi pi-download"
              text
              rounded
              size="small"
              severity="secondary"
              v-tooltip.top="'Unduh satu label'"
              @click.stop="downloadOne(it)"
            />
          </div>

          <div class="px-3 pb-3 text-center">
            <div
              class="grid place-items-center rounded-lg border p-2 mb-2 bg-white shadow-xs"
              style="border-color: var(--line-soft)"
            >
              <img
                :src="`/api/barcode/${it.id}.png?fmt=${fmt}&size=${size}`"
                class="max-w-full"
                :style="{ height: fmt === 'code128' ? '48px' : '76px' }"
                loading="lazy"
                :alt="it.sku"
              />
            </div>
            <div class="text-[12.5px] font-semibold truncate" :title="it.name" style="color: var(--txt)">{{ it.name }}</div>
            <div class="t-mono text-[11px] mt-0.5" style="color: var(--txt-dim)">{{ it.sku }}</div>
            <div class="mt-2 flex items-center justify-center gap-1 flex-wrap">
              <Tag severity="secondary" :value="it.location || 'Tanpa Lokasi'" class="!text-[10px]" />
            </div>
          </div>
        </div>
      </div>

      <!-- Pagination Footer -->
      <div
        class="panel p-3 rounded-xl border flex flex-wrap items-center justify-between gap-3 text-[12px]"
        style="background: var(--paper-1); border-color: var(--line)"
      >
        <span style="color: var(--txt-dim)">{{ range }}</span>
        <div class="flex items-center gap-2">
          <Select v-model="perPage" :options="perPageOptions" class="!text-[12px] !py-0.5 w-[95px]" />
          <Button
            icon="pi pi-angle-left"
            size="small"
            text
            severity="secondary"
            :disabled="page === 0"
            @click="page--"
          />
          <span class="t-num font-semibold px-1" style="color: var(--txt)">Hal. {{ page + 1 }} / {{ lastPage + 1 }}</span>
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
    </div>

    <!-- Print Settings & Live Preview Dialog -->
    <Dialog
      v-model:visible="printModalVisible"
      modal
      header="Pengaturan &amp; Pratinjau Cetak Label"
      :style="{ width: '92vw', maxWidth: '780px' }"
    >
      <div class="flex flex-col gap-4">
        <!-- Settings Form Grid -->
        <div class="grid sm:grid-cols-2 gap-3.5 p-3 rounded-xl border" style="background: var(--panel-2); border-color: var(--line)">
          <div class="flex flex-col gap-1.5 sm:col-span-2">
            <label class="text-[11.5px] font-semibold" style="color: var(--txt)">Pilihan Ukuran Stiker / Jenis Printer</label>
            <Select
              v-model="layout"
              :options="layoutOptions"
              optionLabel="label"
              optionValue="value"
              class="w-full !text-[12.5px]"
            />
          </div>

          <div class="flex flex-col gap-1.5">
            <label class="text-[11.5px] font-semibold" style="color: var(--txt)">Format Barcode</label>
            <Select
              v-model="fmt"
              :options="[
                { label: 'QR Code (Rekomendasi)', value: 'qr' },
                { label: 'Barcode 128 (Garis Horizontal)', value: 'code128' },
              ]"
              optionLabel="label"
              optionValue="value"
              class="w-full !text-[12.5px]"
            />
          </div>

          <div class="flex flex-col gap-1.5">
            <label class="text-[11.5px] font-semibold" style="color: var(--txt)">Teks Header Label</label>
            <InputText
              v-model="labelTitle"
              placeholder="INVENTARIS KANTOR"
              class="w-full !text-[12.5px]"
            />
          </div>

          <!-- Element Toggles -->
          <div class="flex flex-wrap items-center gap-4 sm:col-span-2 pt-2 border-t" style="border-color: var(--line)">
            <label class="flex items-center gap-2 cursor-pointer text-[12px]" style="color: var(--txt)">
              <Checkbox v-model="showName" binary />
              <span>Tampilkan Nama Barang</span>
            </label>
            <label class="flex items-center gap-2 cursor-pointer text-[12px]" style="color: var(--txt)">
              <Checkbox v-model="showSKU" binary />
              <span>Tampilkan Kode SKU</span>
            </label>
            <label class="flex items-center gap-2 cursor-pointer text-[12px]" style="color: var(--txt)">
              <Checkbox v-model="showLoc" binary />
              <span>Tampilkan Ruangan / Lokasi</span>
            </label>
          </div>
        </div>

        <!-- Live Visual Preview of Single Label -->
        <div class="flex flex-col gap-2">
          <div class="flex items-center justify-between">
            <span class="text-[11px] font-semibold uppercase tracking-wider" style="color: var(--txt-dim)">
              Pratinjau Nyata Label (Skala 1:1)
            </span>
            <span class="text-[11px] font-mono" style="color: var(--txt-dim)">
              {{ selected.length }} label akan dicetak
            </span>
          </div>

          <div class="p-6 rounded-xl border grid place-items-center bg-slate-100 dark:bg-slate-900 overflow-x-auto" style="border-color: var(--line)">
            <!-- Actual Visual Card Simulated per Layout -->
            <!-- 1. Stiker Label 50x30 mm -->
            <div
              v-if="layout === 'label_50x30'"
              class="bg-white text-black p-2 rounded shadow-md border border-slate-300 flex items-center justify-between"
              style="width: 280px; height: 168px;"
            >
              <div class="w-1/2 flex items-center justify-center p-1">
                <img :src="`/api/barcode/${sampleItem.id}.png?fmt=${fmt}`" class="max-w-full max-h-full object-contain" alt="code" />
              </div>
              <div class="w-1/2 flex flex-col justify-center pl-2 border-l border-slate-200">
                <div class="text-[9px] font-extrabold uppercase border-b border-black pb-0.5 tracking-wider">{{ labelTitle }}</div>
                <div v-if="showName" class="text-[11px] font-bold mt-1 line-clamp-2 leading-tight">{{ sampleItem.name }}</div>
                <div v-if="showSKU" class="text-[10px] font-mono font-bold mt-1">{{ sampleItem.sku }}</div>
                <div v-if="showLoc && sampleItem.location" class="text-[9px] text-slate-700 mt-0.5 truncate">📍 {{ sampleItem.location }}</div>
              </div>
            </div>

            <!-- 2. Stiker Label 40x30 mm -->
            <div
              v-else-if="layout === 'label_40x30'"
              class="bg-white text-black p-1.5 rounded shadow-md border border-slate-300 flex items-center justify-between"
              style="width: 240px; height: 180px;"
            >
              <div class="w-1/2 flex items-center justify-center">
                <img :src="`/api/barcode/${sampleItem.id}.png?fmt=${fmt}`" class="max-w-full max-h-full object-contain" alt="code" />
              </div>
              <div class="w-1/2 flex flex-col justify-center pl-1.5">
                <div class="text-[8px] font-extrabold uppercase border-b border-black pb-0.5">{{ labelTitle }}</div>
                <div v-if="showName" class="text-[10px] font-bold mt-1 line-clamp-2 leading-tight">{{ sampleItem.name }}</div>
                <div v-if="showSKU" class="text-[9px] font-mono font-bold mt-0.5">{{ sampleItem.sku }}</div>
                <div v-if="showLoc && sampleItem.location" class="text-[8px] text-slate-700 mt-0.5 truncate">{{ sampleItem.location }}</div>
              </div>
            </div>

            <!-- 3. Thermal Roll 58 mm -->
            <div
              v-else-if="layout === 'thermal_58'"
              class="bg-white text-black p-3 rounded shadow-md border border-slate-300 text-center flex flex-col items-center"
              style="width: 260px;"
            >
              <div class="text-[10px] font-bold uppercase tracking-wider border-b border-black pb-1 w-full">{{ labelTitle }}</div>
              <div class="my-2">
                <img :src="`/api/barcode/${sampleItem.id}.png?fmt=${fmt}`" :style="{ height: fmt === 'code128' ? '40px' : '65px' }" alt="code" />
              </div>
              <div v-if="showName" class="text-[11px] font-bold leading-tight">{{ sampleItem.name }}</div>
              <div v-if="showSKU" class="text-[10px] font-mono font-bold mt-0.5">{{ sampleItem.sku }}</div>
              <div v-if="showLoc && sampleItem.location" class="text-[9px] text-slate-700 mt-0.5">📍 {{ sampleItem.location }}</div>
              <div class="w-full border-b border-dashed border-slate-400 mt-2" />
            </div>

            <!-- 4. Thermal Roll 80 mm -->
            <div
              v-else-if="layout === 'thermal_80'"
              class="bg-white text-black p-4 rounded shadow-md border border-slate-300 text-center flex flex-col items-center"
              style="width: 320px;"
            >
              <div class="text-[11px] font-extrabold uppercase tracking-wider border-b-2 border-black pb-1 w-full">{{ labelTitle }}</div>
              <div class="my-2.5">
                <img :src="`/api/barcode/${sampleItem.id}.png?fmt=${fmt}`" :style="{ height: fmt === 'code128' ? '50px' : '85px' }" alt="code" />
              </div>
              <div v-if="showName" class="text-[12.5px] font-bold leading-tight">{{ sampleItem.name }}</div>
              <div v-if="showSKU" class="text-[11px] font-mono font-bold mt-1">{{ sampleItem.sku }}</div>
              <div v-if="showLoc && sampleItem.location" class="text-[10px] text-slate-700 mt-0.5">📍 {{ sampleItem.location }}</div>
              <div class="w-full border-b border-dashed border-slate-400 mt-3" />
            </div>

            <!-- 5. Stiker Lebar 60x40 / 70x50 / 100x50 / A4 -->
            <div
              v-else
              class="bg-white text-black p-3 rounded shadow-md border border-slate-300 flex items-center justify-between"
              style="width: 340px; height: 180px;"
            >
              <div class="w-1/2 flex items-center justify-center p-2">
                <img :src="`/api/barcode/${sampleItem.id}.png?fmt=${fmt}`" class="max-w-full max-h-full object-contain" alt="code" />
              </div>
              <div class="w-1/2 flex flex-col justify-center pl-3 border-l border-slate-200">
                <div class="text-[10px] font-extrabold uppercase border-b border-black pb-1 tracking-wider">{{ labelTitle }}</div>
                <div v-if="showName" class="text-[12px] font-bold mt-1.5 line-clamp-2 leading-tight">{{ sampleItem.name }}</div>
                <div v-if="showSKU" class="text-[11px] font-mono font-bold mt-1">{{ sampleItem.sku }}</div>
                <div v-if="showLoc && sampleItem.location" class="text-[10px] text-slate-700 mt-0.5 truncate">📍 {{ sampleItem.location }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="flex items-center justify-between w-full pt-2">
          <Button
            label="Batal"
            text
            severity="secondary"
            size="small"
            @click="printModalVisible = false"
          />
          <div class="flex items-center gap-2">
            <Button
              label="Buka di Tab Baru"
              icon="pi pi-external-link"
              size="small"
              severity="secondary"
              outlined
              @click="openSheetTab"
            />
            <Button
              label="Cetak Sekarang (Print)"
              icon="pi pi-print"
              size="small"
              @click="printNow"
            />
          </div>
        </div>
      </template>
    </Dialog>
  </div>
</template>
