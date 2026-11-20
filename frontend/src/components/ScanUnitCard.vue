<script setup lang="ts">
import StatusChip from '@/components/StatusChip.vue'

defineProps<{
  unit: any
  scanned?: boolean
  /** compact = ringkasan horizontal (tab Spesifikasi), default = kartu lengkap (tab Unit Fisik) */
  compact?: boolean
}>()

const emit = defineEmits<{ (e: 'photo', url: string): void }>()
</script>

<template>
  <!-- ============ Compact: ringkasan horizontal ============ -->
  <article
    v-if="compact"
    class="p-3 rounded-[10px] border bg-panel-2 flex items-start gap-3"
    :class="scanned ? 'border-acc-500 ring-1 ring-acc-500/30' : 'border-line'"
  >
    <button
      v-if="unit.photo_url"
      type="button"
      class="w-16 h-16 rounded-[8px] border border-line overflow-hidden shrink-0 bg-panel cursor-pointer"
      :aria-label="'Perbesar foto Unit #' + unit.unit_number"
      @click="emit('photo', unit.photo_url)"
    >
      <img :src="unit.photo_url" :alt="'Unit #' + unit.unit_number" class="w-full h-full object-cover" loading="lazy" />
    </button>
    <div v-else class="w-16 h-16 rounded-[8px] border border-line bg-panel grid place-items-center shrink-0 text-txt-dim">
      <i class="pi pi-image text-base" aria-hidden="true" />
    </div>

    <div class="min-w-0 flex-1 space-y-1.5">
      <div class="flex items-center gap-1.5 flex-wrap">
        <span class="px-2 py-0.5 rounded text-[12px] font-bold bg-panel border border-line text-txt">
          Unit #{{ unit.unit_number }}
        </span>
        <span v-if="scanned" class="px-2 py-0.5 rounded text-[11.5px] font-semibold bg-acc-500 text-white dark:text-[#042f2e]">
          Dipindai
        </span>
        <StatusChip :kind="unit.condition_status || 'Baik (B)'" />
      </div>
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-x-3 gap-y-0.5 text-[12px] text-txt-dim">
        <div v-if="unit.register_number">
          No. Reg: <span class="t-mono text-txt font-medium">{{ unit.register_number }}</span>
        </div>
        <div v-if="unit.barcode">
          Barcode: <span class="t-mono text-txt font-medium">{{ unit.barcode }}</span>
        </div>
        <div v-if="unit.serial_number">
          SN: <span class="t-mono text-txt font-medium">{{ unit.serial_number }}</span>
        </div>
        <div v-if="unit.location">
          Lokasi: <span class="text-txt font-medium">
            {{ unit.location }}<span v-if="unit.location_code" class="t-mono text-txt-dim"> ({{ unit.location_code }})</span>
          </span>
        </div>
      </div>
      <div v-if="unit.geo_name" class="text-[12px] text-txt-dim flex items-center gap-1">
        <i class="pi pi-map-marker text-[11px] text-acc-500" aria-hidden="true" />
        <span class="truncate">{{ unit.geo_name }}</span>
      </div>
      <p v-if="unit.notes" class="text-[12px] text-txt-dim leading-snug break-words">{{ unit.notes }}</p>
    </div>
  </article>

  <!-- ============ Full: kartu lengkap ============ -->
  <article
    v-else
    class="rounded-[12px] border bg-panel-2 overflow-hidden"
    :class="scanned ? 'border-acc-500 ring-1 ring-acc-500/30' : 'border-line'"
  >
    <button
      v-if="unit.photo_url"
      type="button"
      class="w-full h-44 bg-panel border-b border-line overflow-hidden block cursor-pointer relative"
      :aria-label="'Perbesar foto Unit #' + unit.unit_number"
      @click="emit('photo', unit.photo_url)"
    >
      <img :src="unit.photo_url" :alt="'Unit #' + unit.unit_number" class="w-full h-full object-cover" loading="lazy" />
      <span class="absolute bottom-2 right-2 px-2 py-1 rounded bg-black/65 text-white text-[12px] font-medium">
        <i class="pi pi-search-plus mr-1 text-[11px]" aria-hidden="true" />Perbesar
      </span>
    </button>

    <div class="p-3.5 space-y-2.5">
      <div class="flex items-center justify-between gap-2 flex-wrap">
        <div class="flex items-center gap-1.5 flex-wrap">
          <span class="px-2 py-0.5 rounded text-[12.5px] font-bold bg-panel border border-line text-txt">
            Unit #{{ unit.unit_number }}
          </span>
          <span v-if="scanned" class="px-2 py-0.5 rounded text-[11.5px] font-semibold bg-acc-500 text-white dark:text-[#042f2e]">
            Unit Dipindai
          </span>
        </div>
        <StatusChip :kind="unit.condition_status || 'Baik (B)'" />
      </div>

      <dl class="divide-y divide-line-soft text-[12.5px]">
        <div v-if="unit.register_number" class="py-2 flex items-center justify-between gap-3">
          <dt class="text-txt-dim">No. Register / NUP</dt>
          <dd class="t-mono font-semibold text-txt">{{ unit.register_number }}</dd>
        </div>
        <div v-if="unit.barcode" class="py-2 flex items-center justify-between gap-3">
          <dt class="text-txt-dim">Barcode Unit</dt>
          <dd class="t-mono font-semibold text-txt">{{ unit.barcode }}</dd>
        </div>
        <div v-if="unit.serial_number" class="py-2 flex items-center justify-between gap-3">
          <dt class="text-txt-dim">Nomor Seri (SN)</dt>
          <dd class="t-mono font-semibold text-txt">{{ unit.serial_number }}</dd>
        </div>
        <div v-if="unit.location" class="py-2 flex items-center justify-between gap-3">
          <dt class="text-txt-dim">Lokasi / Ruangan</dt>
          <dd class="font-medium text-txt text-right">
            {{ unit.location }}<span v-if="unit.location_code" class="t-mono text-txt-dim"> ({{ unit.location_code }})</span>
          </dd>
        </div>
        <div v-if="unit.geo_name" class="py-2 flex items-start justify-between gap-3">
          <dt class="text-txt-dim shrink-0">Lokasi Foto</dt>
          <dd class="text-txt text-right break-words">{{ unit.geo_name }}</dd>
        </div>
        <div v-if="unit.notes" class="py-2 flex items-start justify-between gap-3">
          <dt class="text-txt-dim shrink-0">Catatan</dt>
          <dd class="text-txt text-right break-words">{{ unit.notes }}</dd>
        </div>
      </dl>
    </div>
  </article>
</template>
