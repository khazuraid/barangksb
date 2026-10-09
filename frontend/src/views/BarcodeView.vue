<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import api from '@/api'
import { useToast } from 'primevue/usetoast'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Checkbox from 'primevue/checkbox'
import Tag from 'primevue/tag'
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
  if (typeof window !== 'undefined') {
    window.open(`/api/barcode/sheet?ids=${selected.value.join(',')}`, '_blank')
  }
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
      sub="Pilih barang inventaris, atur format barcode, lalu cetak lembar label siap tempel"
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
        :hint="selected.length ? 'Siap dicetak di lembar' : 'Belum ada barang dicentang'"
      />
      <StatCard
        label="Format Label Aktif"
        :value="fmt === 'qr' ? 'QR Code' : 'Barcode 128'"
        icon="pi pi-qrcode"
        tone="info"
        hint="Resolusi tajam siap tempel"
      />
    </div>

    <!-- Clean Unified Toolbar -->
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
          class="!text-[12px] !py-0.5 w-[140px]"
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
      <div class="grid gap-3" style="grid-template-columns: repeat(auto-fill, minmax(190px, 1fr))">
        <div
          v-for="it in items"
          :key="it.id"
          class="panel relative overflow-hidden transition-all cursor-pointer rounded-xl border flex flex-col justify-between hover:shadow-md"
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
              class="grid place-items-center rounded-lg border p-2.5 mb-2.5 bg-white shadow-sm"
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
  </div>
</template>
