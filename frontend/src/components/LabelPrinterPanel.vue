<script setup lang="ts">
import Select from 'primevue/select'
import Button from 'primevue/button'
import Checkbox from 'primevue/checkbox'
import {
  type SheetFormatId,
  type PrintMode,
  SHEET_FORMAT_OPTIONS,
  PRINT_MODE_OPTIONS,
  getSheetSpec,
} from '@/utils/customLabelPrint'

defineProps<{
  sheetSpec: ReturnType<typeof getSheetSpec>
}>()

const sheetFormat = defineModel<SheetFormatId>('sheetFormat', { required: true })
const printMode = defineModel<PrintMode>('printMode', { required: true })
const showCutBorder = defineModel<boolean>('showCutBorder', { required: true })

const emit = defineEmits<{
  (e: 'preview'): void
  (e: 'print'): void
}>()
</script>

<template>
  <div class="flex-1 overflow-y-auto p-3 flex flex-col gap-3">
  <div class="flex flex-col gap-1.5">
    <label class="t-label text-txt">Format Kertas / Lembar Cetak</label>
    <Select
      v-model="sheetFormat"
      :options="SHEET_FORMAT_OPTIONS"
      optionLabel="label"
      optionValue="value"
      class="w-full !text-xs"
    />
    <span class="text-[10.5px] text-txt-dim leading-snug">
      Memilih format Tom &amp; Jerry / A4 otomatis menyelaraskan ukuran stiker &amp; susunan lembar saat dicetak.
    </span>
  </div>

  <div class="flex flex-col gap-1.5 pt-2 border-t border-line">
    <label class="t-label text-txt">Tampilan Printer saat Cetak</label>
    <Select
      v-model="printMode"
      :options="PRINT_MODE_OPTIONS"
      optionLabel="label"
      optionValue="value"
      class="w-full !text-xs"
    />
    <span class="text-[10.5px] text-txt-dim leading-snug">
      Mode <b>Lembar Cetak Tab Baru</b> menampilkan pratinjau lembar kuning/A4 &amp; tombol cetak selaras dengan layout bawaan.
    </span>
  </div>

  <div class="pt-2 border-t border-line flex flex-col gap-2">
    <label class="flex items-center gap-2 cursor-pointer text-xs text-txt">
      <Checkbox v-model="showCutBorder" binary />
      <span>Garis bantu potong putus-putus di tiap slot</span>
    </label>
  </div>

  <!-- Ringkasan & Mini Preview Lembar -->
  <div class="p-2.5 rounded-lg border border-line bg-panel-2 flex flex-col gap-2">
    <div class="flex items-center justify-between text-[11px]">
      <span class="font-semibold text-txt">{{ sheetSpec.shortLabel }}</span>
      <span class="t-mono text-acc-500 font-semibold">
        {{ sheetSpec.isSheet ? `${sheetSpec.perSheet}/lembar` : '1/halaman' }}
      </span>
    </div>
    <div class="grid grid-cols-2 gap-1.5 text-[10.5px] text-txt-dim t-mono">
      <div>Grid: <b class="text-txt">{{ sheetSpec.cols }}×{{ sheetSpec.rows }}</b></div>
      <div>Slot: <b class="text-txt">{{ sheetSpec.labelW }}×{{ sheetSpec.labelH }} mm</b></div>
      <div>Jarak: <b class="text-txt">{{ sheetSpec.rowGapMm }}/{{ sheetSpec.colGapMm }} mm</b></div>
      <div>Margin: <b class="text-txt">{{ sheetSpec.pageMargin }}</b></div>
    </div>

    <!-- Visual Mini Lembar -->
    <div
      class="p-2 rounded border grid place-items-center"
      :style="{
        background: sheetSpec.isYellowBacking ? '#fef08a' : '#ffffff',
        borderColor: sheetSpec.isYellowBacking ? '#facc15' : '#cbd5e1',
      }"
    >
      <div
        class="w-full grid gap-1"
        :style="{
          gridTemplateColumns: `repeat(${Math.min(sheetSpec.cols, 5)}, minmax(0, 1fr))`,
        }"
      >
        <div
          v-for="idx in Math.min(sheetSpec.perSheet, 15)"
          :key="idx"
          class="h-4 rounded-[2px] bg-white flex items-center justify-center text-[8px] font-mono text-slate-500"
          :style="{
            border: showCutBorder ? '1px dashed #94a3b8' : '1px solid #cbd5e1',
          }"
        >
          {{ idx }}
        </div>
      </div>
      <span v-if="sheetSpec.perSheet > 15" class="text-[9.5px] text-slate-500 mt-1">
        +{{ sheetSpec.perSheet - 15 }} slot lainnya per lembar
      </span>
    </div>

    <div class="flex items-center gap-1.5 pt-1">
      <Button
        label="Pratinjau Lembar"
        icon="pi pi-external-link"
        size="small"
        severity="secondary"
        outlined
        class="flex-1 !text-[11px] !py-1"
        @click="emit('preview')"
      />
      <Button
        label="Cetak"
        icon="pi pi-print"
        size="small"
        class="!text-[11px] !py-1"
        @click="emit('print')"
      />
    </div>
  </div>
  </div>
</template>
