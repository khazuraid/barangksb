<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import api from '@/api'
import { useToast } from 'primevue/usetoast'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Checkbox from 'primevue/checkbox'
import PageHeader from '@/components/PageHeader.vue'
import EmptyState from '@/components/EmptyState.vue'
import Tag from 'primevue/tag'

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

function print() {
  if (!selected.value.length) {
    toast.add({
      severity: 'warn',
      summary: 'Belum ada label dipilih',
      detail: 'Centang minimal satu barang.',
      life: 2500,
    })
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
  <div class="pb-16 max-w-7xl mx-auto">
    <PageHeader
      crumb="Perangkat"
      title="Generator Label QR / Barcode"
      sub="Pilih barang, atur format barcode, lalu cetak lembar label siap tempel"
    >
      <template #actions>
        <Button
          label="Cetak Lembar Terpilih"
          icon="pi pi-print"
          size="small"
          :disabled="!selected.length"
          @click="print"
        />
      </template>
    </PageHeader>

    <!-- Toolbar Filter -->
    <div
      class="panel p-3.5 mb-4 flex flex-wrap gap-3 items-end rounded-xl border"
      style="border-color: var(--line); background: var(--paper-1)"
    >
      <label class="flex flex-col gap-1.5 flex-1 min-w-[200px]">
        <span class="text-[11.5px] font-semibold text-ink-400">Cari Barang</span>
        <InputText
          v-model="q"
          placeholder="Nama atau SKU…"
          class="w-full !text-[12px] !py-1.5"
          @keyup.enter="fetchItems(true)"
        />
      </label>

      <label class="flex flex-col gap-1.5 min-w-[150px]">
        <span class="text-[11.5px] font-semibold text-ink-400">Kategori</span>
        <Select
          v-model="categoryFilter"
          :options="categoriesRaw"
          optionLabel="name"
          optionValue="name"
          showClear
          placeholder="Semua"
          class="!text-[12px]"
          @change="fetchItems(true)"
          filter
        />
      </label>

      <label class="flex flex-col gap-1.5 min-w-[150px]">
        <span class="text-[11.5px] font-semibold text-ink-400">Lokasi</span>
        <Select
          v-model="locationFilter"
          :options="locationsRaw"
          optionLabel="name"
          optionValue="name"
          showClear
          placeholder="Semua"
          class="!text-[12px]"
          @change="fetchItems(true)"
          filter
        />
      </label>

      <label class="flex flex-col gap-1.5 min-w-[130px]">
        <span class="text-[11.5px] font-semibold text-ink-400">Format</span>
        <Select
          v-model="fmt"
          :options="[
            { label: 'QR Code', value: 'qr' },
            { label: 'Barcode 128', value: 'code128' },
          ]"
          optionLabel="label"
          optionValue="value"
          class="!text-[12px]"
        />
      </label>

      <div class="flex items-center gap-2 pb-0.5">
        <Button icon="pi pi-search" size="small" @click="fetchItems(true)" />
        <Button
          :label="isAllOnPageSelected ? 'Batal pilih hal. ini' : 'Pilih semua di hal. ini'"
          icon="pi pi-check-square"
          text
          size="small"
          severity="secondary"
          @click="toggleAllOnPage"
        />
        <Tag :value="selected.length + ' dipilih'" severity="warn" />
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
      <div class="grid gap-3" style="grid-template-columns: repeat(auto-fill, minmax(185px, 1fr))">
        <div
          v-for="it in items"
          :key="it.id"
          class="panel relative overflow-hidden transition-all cursor-pointer rounded-xl border flex flex-col justify-between"
          :class="selected.includes(it.id) ? 'ring-2 ring-acc-500 border-acc-500' : 'hover:border-ink-400'"
          style="background: var(--paper-1); border-color: var(--line)"
          @click="selected.includes(it.id) ? selected.splice(selected.indexOf(it.id), 1) : selected.push(it.id)"
        >
          <span
            v-if="selected.includes(it.id)"
            class="absolute top-0 left-0 right-0 h-[3px] bg-acc-500"
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
              class="grid place-items-center rounded-lg border p-2 mb-2 bg-white"
              style="border-color: var(--line-soft)"
            >
              <img
                :src="`/api/barcode/${it.id}.png?fmt=${fmt}&size=${size}`"
                class="max-w-full"
                :style="{ height: size * 0.58 + 'px' }"
                loading="lazy"
                :alt="it.sku"
              />
            </div>
            <div class="text-[12.5px] font-semibold truncate" :title="it.name">{{ it.name }}</div>
            <div class="t-mono text-[11px] text-ink-400 mt-0.5">{{ it.sku }}</div>
            <div class="mt-1.5 flex items-center justify-center gap-1 flex-wrap">
              <Tag severity="secondary" :value="it.location" class="!text-[10px]" />
            </div>
          </div>
        </div>
      </div>

      <!-- Pagination Footer -->
      <div
        class="panel p-3 rounded-xl border flex flex-wrap items-center justify-between gap-3 text-[12px]"
        style="background: var(--paper-1); border-color: var(--line)"
      >
        <span class="text-ink-400">{{ range }}</span>
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
          <span class="t-num font-semibold px-1">Hal. {{ page + 1 }} / {{ lastPage + 1 }}</span>
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
  </div>
</template>
