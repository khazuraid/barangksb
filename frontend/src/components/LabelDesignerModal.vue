<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import Dialog from 'primevue/dialog'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import InputNumber from 'primevue/inputnumber'
import Select from 'primevue/select'
import Checkbox from 'primevue/checkbox'
import Tag from 'primevue/tag'
import { useToast } from 'primevue/usetoast'

export interface LabelElement {
  id: string
  type: 'qr' | 'code128' | 'field' | 'static_text' | 'header' | 'line' | 'logo'
  field?: 'name' | 'sku' | 'location' | 'category' | 'register_number' | 'merk'
  text?: string
  logoUrl?: string
  x: number // mm
  y: number // mm
  w: number // mm
  h: number // mm
  fontSize?: number // pt
  bold?: boolean
  align?: 'left' | 'center' | 'right'
  border?: boolean
  lineOrientation?: 'horizontal' | 'vertical'
}

export interface LabelTemplate {
  id: string
  name: string
  widthMm: number
  heightMm: number
  elements: LabelElement[]
}

const props = defineProps<{
  visible: boolean
  items: any[]
  selectedIds: string[]
}>()

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
}>()

const toast = useToast()
const MM_TO_PX = 3.7795
const zoom = ref(2.5) // Screen display zoom factor
const snapGrid = ref(true)
const gridSizeMm = ref(1)

const DEFAULT_PRESETS: LabelTemplate[] = [
  {
    id: 'preset_50x30',
    name: 'Stiker Rol Standar (50 × 30 mm)',
    widthMm: 50,
    heightMm: 30,
    elements: [
      { id: 'qr_1', type: 'qr', x: 2, y: 2.5, w: 25, h: 25 },
      { id: 'hdr_1', type: 'header', text: 'INVENTARIS KANTOR', x: 28, y: 2.5, w: 20, h: 4, fontSize: 6.5, bold: true, align: 'left', border: true },
      { id: 'f_name', type: 'field', field: 'name', x: 28, y: 7.5, w: 20, h: 9, fontSize: 8, bold: true, align: 'left' },
      { id: 'f_sku', type: 'field', field: 'sku', x: 28, y: 17.5, w: 20, h: 4, fontSize: 7, bold: true, align: 'left' },
      { id: 'f_loc', type: 'field', field: 'location', x: 28, y: 22.5, w: 20, h: 4, fontSize: 6.5, bold: false, align: 'left' },
    ],
  },
  {
    id: 'preset_tj_103',
    name: 'Tom & Jerry No. 103 (64 × 32 mm)',
    widthMm: 64,
    heightMm: 32,
    elements: [
      { id: 'qr_1', type: 'qr', x: 3, y: 3, w: 26, h: 26 },
      { id: 'hdr_1', type: 'header', text: 'PEMERINTAH KABUPATEN KSB', x: 31, y: 2.5, w: 30, h: 4.5, fontSize: 6.5, bold: true, align: 'left', border: true },
      { id: 'f_name', type: 'field', field: 'name', x: 31, y: 8, w: 30, h: 10, fontSize: 8.5, bold: true, align: 'left' },
      { id: 'f_sku', type: 'field', field: 'sku', x: 31, y: 19, w: 30, h: 4.5, fontSize: 7.5, bold: true, align: 'left' },
      { id: 'f_loc', type: 'field', field: 'location', x: 31, y: 24.5, w: 30, h: 4, fontSize: 7, bold: false, align: 'left' },
    ],
  },
  {
    id: 'preset_40x30',
    name: 'Stiker Rol Mini (40 × 30 mm)',
    widthMm: 40,
    heightMm: 30,
    elements: [
      { id: 'qr_1', type: 'qr', x: 1.5, y: 4, w: 22, h: 22 },
      { id: 'f_name', type: 'field', field: 'name', x: 24.5, y: 4, w: 14.5, h: 10, fontSize: 7.5, bold: true, align: 'left' },
      { id: 'f_sku', type: 'field', field: 'sku', x: 24.5, y: 15, w: 14.5, h: 4, fontSize: 6.5, bold: true, align: 'left' },
      { id: 'f_loc', type: 'field', field: 'location', x: 24.5, y: 20, w: 14.5, h: 4, fontSize: 6, bold: false, align: 'left' },
    ],
  },
  {
    id: 'preset_70x50',
    name: 'Stiker Aset Besar (70 × 50 mm)',
    widthMm: 70,
    heightMm: 50,
    elements: [
      { id: 'hdr_1', type: 'header', text: 'KARTU INVENTARIS BARANG MILIK DAERAH', x: 3, y: 3, w: 64, h: 5, fontSize: 7.5, bold: true, align: 'center', border: true },
      { id: 'qr_1', type: 'qr', x: 4, y: 10, w: 32, h: 32 },
      { id: 'f_name', type: 'field', field: 'name', x: 38, y: 10, w: 29, h: 12, fontSize: 9.5, bold: true, align: 'left' },
      { id: 'f_sku', type: 'field', field: 'sku', x: 38, y: 23, w: 29, h: 5, fontSize: 8, bold: true, align: 'left' },
      { id: 'f_nup', type: 'field', field: 'register_number', x: 38, y: 29, w: 29, h: 5, fontSize: 7.5, bold: false, align: 'left' },
      { id: 'f_loc', type: 'field', field: 'location', x: 38, y: 35, w: 29, h: 5, fontSize: 7.5, bold: false, align: 'left' },
      { id: 'st_1', type: 'static_text', text: 'DILARANG DILEPAS / DIPINDAHKAN', x: 4, y: 44, w: 62, h: 4, fontSize: 6, bold: false, align: 'center' },
    ],
  },
]

const templates = ref<LabelTemplate[]>([])
const activeTemplate = ref<LabelTemplate>(JSON.parse(JSON.stringify(DEFAULT_PRESETS[0])))
const activeElementId = ref<string | null>(null)
const copies = ref(1)
const isPrinting = ref(false)

function loadSavedTemplates() {
  const raw = localStorage.getItem('custom_label_templates_v2')
  if (raw) {
    try {
      templates.value = JSON.parse(raw)
    } catch {
      templates.value = JSON.parse(JSON.stringify(DEFAULT_PRESETS))
    }
  } else {
    templates.value = JSON.parse(JSON.stringify(DEFAULT_PRESETS))
  }
}

function saveTemplates() {
  localStorage.setItem('custom_label_templates_v2', JSON.stringify(templates.value))
}

onMounted(() => {
  loadSavedTemplates()
})

const activeElement = computed(() => {
  if (!activeElementId.value) return null
  return activeTemplate.value.elements.find((el) => el.id === activeElementId.value) || null
})

// Sample item for preview
const sampleItem = computed(() => {
  if (props.selectedIds.length > 0) {
    const found = props.items.find((i) => i.id === props.selectedIds[0])
    if (found) return found
  }
  return (
    props.items[0] || {
      id: 'sample',
      name: 'Laptop Lenovo ThinkPad E14 Gen 4',
      sku: 'ELK-2026-001',
      location: 'Ruang Server IT (Lantai 2)',
      category: 'Elektronik & IT',
      register_number: '0001 / BMD / 2024',
      merk: 'Lenovo ThinkPad',
    }
  )
})

// Drag and Resize Interaction
const canvasRef = ref<HTMLElement | null>(null)
let isDragging = false
let isResizing = false
let dragStartX = 0
let dragStartY = 0
let initialElemX = 0
let initialElemY = 0
let initialElemW = 0
let initialElemH = 0

function startDrag(e: PointerEvent, el: LabelElement) {
  activeElementId.value = el.id
  isDragging = true
  isResizing = false
  dragStartX = e.clientX
  dragStartY = e.clientY
  initialElemX = el.x
  initialElemY = el.y
  ;(e.target as HTMLElement).setPointerCapture(e.pointerId)
}

function startResize(e: PointerEvent, el: LabelElement) {
  e.stopPropagation()
  activeElementId.value = el.id
  isDragging = false
  isResizing = true
  dragStartX = e.clientX
  dragStartY = e.clientY
  initialElemW = el.w
  initialElemH = el.h
  ;(e.target as HTMLElement).setPointerCapture(e.pointerId)
}

function onPointerMove(e: PointerEvent) {
  if (!activeElement.value) return
  const currentScale = MM_TO_PX * zoom.value

  if (isDragging) {
    const deltaXmm = (e.clientX - dragStartX) / currentScale
    const deltaYmm = (e.clientY - dragStartY) / currentScale

    let newX = initialElemX + deltaXmm
    let newY = initialElemY + deltaYmm

    if (snapGrid.value) {
      newX = Math.round(newX / gridSizeMm.value) * gridSizeMm.value
      newY = Math.round(newY / gridSizeMm.value) * gridSizeMm.value
    }

    // Clamp within canvas boundaries
    newX = Math.max(0, Math.min(activeTemplate.value.widthMm - activeElement.value.w, newX))
    newY = Math.max(0, Math.min(activeTemplate.value.heightMm - activeElement.value.h, newY))

    activeElement.value.x = Number(newX.toFixed(1))
    activeElement.value.y = Number(newY.toFixed(1))
  } else if (isResizing) {
    const deltaWmm = (e.clientX - dragStartX) / currentScale
    const deltaHmm = (e.clientY - dragStartY) / currentScale

    let newW = initialElemW + deltaWmm
    let newH = initialElemH + deltaHmm

    if (snapGrid.value) {
      newW = Math.round(newW / gridSizeMm.value) * gridSizeMm.value
      newH = Math.round(newH / gridSizeMm.value) * gridSizeMm.value
    }

    newW = Math.max(3, Math.min(activeTemplate.value.widthMm - activeElement.value.x, newW))
    newH = Math.max(2, Math.min(activeTemplate.value.heightMm - activeElement.value.y, newH))

    activeElement.value.w = Number(newW.toFixed(1))
    activeElement.value.h = Number(newH.toFixed(1))
  }
}

function onPointerUp(e: PointerEvent) {
  isDragging = false
  isResizing = false
}

// Add New Elements
function addElement(type: LabelElement['type'], field?: LabelElement['field']) {
  const id = `${type}_${Date.now().toString(36)}`
  let newEl: LabelElement

  switch (type) {
    case 'qr':
      newEl = { id, type: 'qr', x: 2, y: 2, w: 20, h: 20 }
      break
    case 'code128':
      newEl = { id, type: 'code128', x: 2, y: 2, w: 35, h: 12 }
      break
    case 'header':
      newEl = { id, type: 'header', text: 'INVENTARIS KANTOR', x: 2, y: 2, w: 25, h: 4, fontSize: 7, bold: true, align: 'left', border: true }
      break
    case 'field':
      newEl = {
        id,
        type: 'field',
        field: field || 'name',
        x: 4,
        y: 4,
        w: 24,
        h: field === 'name' ? 8 : 4.5,
        fontSize: field === 'name' ? 8 : 7,
        bold: field === 'name' || field === 'sku',
        align: 'left',
      }
      break
    case 'static_text':
      newEl = { id, type: 'static_text', text: 'Teks Kustom', x: 2, y: 2, w: 25, h: 4, fontSize: 7, bold: false, align: 'left' }
      break
    case 'line':
      newEl = { id, type: 'line', x: 2, y: 2, w: 25, h: 1, lineOrientation: 'horizontal' }
      break
    case 'logo':
      newEl = { id, type: 'logo', x: 2, y: 2, w: 10, h: 10, logoUrl: '' }
      break
  }

  activeTemplate.value.elements.push(newEl)
  activeElementId.value = id
}

function handleLogoUpload(e: Event) {
  const input = e.target as HTMLInputElement
  if (!input.files?.length || !activeElement.value) return
  const file = input.files[0]
  const reader = new FileReader()
  reader.onload = (ev) => {
    if (activeElement.value && ev.target?.result) {
      activeElement.value.logoUrl = ev.target.result as string
      toast.add({ severity: 'success', summary: 'Logo Berhasil Diunggah', life: 1800 })
    }
  }
  reader.readAsDataURL(file)
}

function removeActiveElement() {
  if (!activeElementId.value) return
  activeTemplate.value.elements = activeTemplate.value.elements.filter((el) => el.id !== activeElementId.value)
  activeElementId.value = null
}

function duplicateActiveElement() {
  if (!activeElement.value) return
  const copy: LabelElement = JSON.parse(JSON.stringify(activeElement.value))
  copy.id = `${copy.type}_${Date.now().toString(36)}`
  copy.x = Math.min(activeTemplate.value.widthMm - copy.w, copy.x + 2)
  copy.y = Math.min(activeTemplate.value.heightMm - copy.h, copy.y + 2)
  activeTemplate.value.elements.push(copy)
  activeElementId.value = copy.id
}

// Preset management
function applyPreset(tpl: LabelTemplate) {
  activeTemplate.value = JSON.parse(JSON.stringify(tpl))
  activeElementId.value = null
  toast.add({ severity: 'info', summary: 'Template Dimuat', detail: tpl.name, life: 1800 })
}

function saveCurrentAsNewTemplate() {
  const name = prompt('Masukkan nama template kustom:', activeTemplate.value.name)
  if (!name) return
  const newTpl: LabelTemplate = {
    ...JSON.parse(JSON.stringify(activeTemplate.value)),
    id: `custom_${Date.now().toString(36)}`,
    name: name.trim(),
  }
  templates.value.push(newTpl)
  saveTemplates()
  toast.add({ severity: 'success', summary: 'Template Disimpan', detail: newTpl.name, life: 2000 })
}

function exportJSON() {
  const blob = new Blob([JSON.stringify(activeTemplate.value, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `template_label_${activeTemplate.value.widthMm}x${activeTemplate.value.heightMm}mm.json`
  a.click()
  URL.revokeObjectURL(url)
}

function importJSON(e: Event) {
  const input = e.target as HTMLInputElement
  if (!input.files?.length) return
  const file = input.files[0]
  const reader = new FileReader()
  reader.onload = (ev) => {
    try {
      const parsed = JSON.parse(ev.target?.result as string)
      if (parsed.widthMm && parsed.heightMm && Array.isArray(parsed.elements)) {
        parsed.id = `imported_${Date.now().toString(36)}`
        activeTemplate.value = parsed
        templates.value.push(parsed)
        saveTemplates()
        toast.add({ severity: 'success', summary: 'Template Berhasil Diimpor', life: 2000 })
      } else {
        toast.add({ severity: 'error', summary: 'Format file JSON tidak valid', life: 2500 })
      }
    } catch {
      toast.add({ severity: 'error', summary: 'Gagal membaca file JSON', life: 2500 })
    }
  }
  reader.readAsText(file)
}

// Keydown nudging
function onKeyDown(e: KeyboardEvent) {
  if (!activeElement.value) return
  if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return

  const step = e.shiftKey ? 1 : 0.5
  if (e.key === 'ArrowLeft') {
    activeElement.value.x = Math.max(0, activeElement.value.x - step)
    e.preventDefault()
  } else if (e.key === 'ArrowRight') {
    activeElement.value.x = Math.min(activeTemplate.value.widthMm - activeElement.value.w, activeElement.value.x + step)
    e.preventDefault()
  } else if (e.key === 'ArrowUp') {
    activeElement.value.y = Math.max(0, activeElement.value.y - step)
    e.preventDefault()
  } else if (e.key === 'ArrowDown') {
    activeElement.value.y = Math.min(activeTemplate.value.heightMm - activeElement.value.h, activeElement.value.y + step)
    e.preventDefault()
  } else if (e.key === 'Delete' || e.key === 'Backspace') {
    removeActiveElement()
    e.preventDefault()
  }
}

onMounted(() => {
  window.addEventListener('keydown', onKeyDown)
})
onUnmounted(() => {
  window.removeEventListener('keydown', onKeyDown)
})

// Get Items for printing
const itemsToPrint = computed(() => {
  const targetIds = props.selectedIds.length > 0 ? props.selectedIds : [sampleItem.value.id]
  const list: any[] = []
  targetIds.forEach((id) => {
    const item = props.items.find((i) => i.id === id) || sampleItem.value
    for (let c = 0; c < Math.max(1, copies.value); c++) {
      list.push(item)
    }
  })
  return list
})

function printLabels() {
  isPrinting.value = true
  setTimeout(() => {
    window.print()
    isPrinting.value = false
  }, 300)
}

function resolveFieldValue(item: any, field?: string) {
  switch (field) {
    case 'name': return item.name || '-'
    case 'sku': return item.sku || '-'
    case 'location': return item.location ? `📍 ${item.location}` : '-'
    case 'category': return item.category || '-'
    case 'register_number': return item.register_number ? `No. Reg: ${item.register_number}` : '-'
    case 'merk': return item.merk || '-'
    default: return '-'
  }
}
</script>

<template>
  <Dialog
    :visible="visible"
    modal
    header="🎨 Desainer Label QR &amp; Barcode (Full Drag &amp; Drop)"
    :style="{ width: '96vw', maxWidth: '1240px', height: '92vh' }"
    :contentStyle="{ padding: '0', display: 'flex', flexDirection: 'column', overflow: 'hidden' }"
    @update:visible="emit('update:visible', $event)"
  >
    <!-- Top Action Bar -->
    <div class="p-3 border-b flex flex-wrap items-center justify-between gap-2.5 bg-slate-900 text-white text-xs select-none">
      <div class="flex items-center gap-2 flex-wrap">
        <span class="font-bold text-slate-300">Ukuran Label:</span>
        <div class="flex items-center gap-1">
          <InputNumber
            v-model="activeTemplate.widthMm"
            :min="20"
            :max="210"
            suffix=" mm"
            class="!w-[88px] !text-xs"
            inputClass="!py-1 !text-xs !bg-slate-800 !text-white !border-slate-700"
          />
          <span class="text-slate-400">&times;</span>
          <InputNumber
            v-model="activeTemplate.heightMm"
            :min="10"
            :max="297"
            suffix=" mm"
            class="!w-[88px] !text-xs"
            inputClass="!py-1 !text-xs !bg-slate-800 !text-white !border-slate-700"
          />
        </div>

        <span class="h-4 w-px bg-slate-700 mx-1" />

        <!-- Preset Templates Selector -->
        <Select
          :modelValue="activeTemplate.id"
          :options="templates"
          optionLabel="name"
          optionValue="id"
          placeholder="Pilih Preset..."
          class="!w-[230px] !text-xs"
          @update:modelValue="(val) => {
            const found = templates.find((t) => t.id === val)
            if (found) applyPreset(found)
          }"
        />

        <Button
          icon="pi pi-bookmark"
          label="Simpan Template"
          size="small"
          severity="secondary"
          class="!text-xs !py-1"
          @click="saveCurrentAsNewTemplate"
        />
      </div>

      <div class="flex items-center gap-2">
        <!-- Zoom Controller -->
        <div class="flex items-center gap-1 bg-slate-800 rounded px-2 py-0.5 border border-slate-700">
          <i class="pi pi-search-plus text-slate-400 text-xs" />
          <button type="button" class="px-1 text-slate-300 hover:text-white" @click="zoom = Math.max(1, zoom - 0.5)">-</button>
          <span class="text-xs font-mono w-10 text-center">{{ Math.round(zoom * 40) }}%</span>
          <button type="button" class="px-1 text-slate-300 hover:text-white" @click="zoom = Math.min(4, zoom + 0.5)">+</button>
        </div>

        <!-- Snap Grid Toggle -->
        <label class="flex items-center gap-1.5 cursor-pointer text-slate-300 text-xs bg-slate-800 px-2 py-1 rounded border border-slate-700">
          <Checkbox v-model="snapGrid" binary />
          <span>Snap 1mm</span>
        </label>

        <!-- Import / Export -->
        <Button icon="pi pi-download" text rounded size="small" severity="secondary" v-tooltip.top="'Ekspor JSON'" @click="exportJSON" />
        <label class="cursor-pointer p-1.5 rounded hover:bg-slate-800 text-slate-300" v-tooltip.top="'Impor JSON'">
          <i class="pi pi-upload text-sm" />
          <input type="file" accept=".json" class="hidden" @change="importJSON" />
        </label>

        <span class="h-4 w-px bg-slate-700 mx-1" />

        <div class="flex items-center gap-1.5">
          <InputNumber
            v-model="copies"
            :min="1"
            :max="50"
            prefix="&times; "
            class="!w-[65px] !text-xs"
            inputClass="!py-1 !text-xs !bg-slate-800 !text-white !border-slate-700 text-center"
            v-tooltip.top="'Jumlah rangkap cetak'"
          />
          <Button
            label="Cetak Label Ini"
            icon="pi pi-print"
            size="small"
            severity="success"
            class="!text-xs !py-1 font-bold"
            @click="printLabels"
          />
        </div>
      </div>
    </div>

    <!-- Main Workspace (Sidebar Palette + Canvas + Property Inspector) -->
    <div class="flex-1 flex overflow-hidden bg-slate-950">
      <!-- Left Sidebar: Elements Palette -->
      <div class="w-60 border-r border-slate-800 bg-slate-900/90 p-3 flex flex-col gap-3 text-xs overflow-y-auto select-none">
        <div class="font-bold text-slate-300 uppercase tracking-wider text-[10.5px]">Tambah Elemen Baru</div>

        <div class="flex flex-col gap-1.5">
          <button
            type="button"
            class="palette-btn"
            @click="addElement('qr')"
          >
            <i class="pi pi-qrcode text-indigo-400 text-base" />
            <div class="text-left">
              <div class="font-bold">QR Code</div>
              <div class="text-[10px] text-slate-400">Link scan detail aset</div>
            </div>
          </button>

          <button
            type="button"
            class="palette-btn"
            @click="addElement('code128')"
          >
            <i class="pi pi-barcode text-sky-400 text-base" />
            <div class="text-left">
              <div class="font-bold">Barcode 128</div>
              <div class="text-[10px] text-slate-400">Barcode garis SKU</div>
            </div>
          </button>

          <button
            type="button"
            class="palette-btn"
            @click="addElement('header')"
          >
            <i class="pi pi-building text-amber-400 text-base" />
            <div class="text-left">
              <div class="font-bold">Header / Instansi</div>
              <div class="text-[10px] text-slate-400">Judul kop stiker</div>
            </div>
          </button>

          <button
            type="button"
            class="palette-btn"
            @click="addElement('field', 'name')"
          >
            <i class="pi pi-box text-emerald-400 text-base" />
            <div class="text-left">
              <div class="font-bold">Nama Barang</div>
              <div class="text-[10px] text-slate-400">Field dinamis nama</div>
            </div>
          </button>

          <button
            type="button"
            class="palette-btn"
            @click="addElement('field', 'sku')"
          >
            <i class="pi pi-hashtag text-cyan-400 text-base" />
            <div class="text-left">
              <div class="font-bold">Kode SKU</div>
              <div class="text-[10px] text-slate-400">Kode inventaris barang</div>
            </div>
          </button>

          <button
            type="button"
            class="palette-btn"
            @click="addElement('field', 'register_number')"
          >
            <i class="pi pi-id-card text-purple-400 text-base" />
            <div class="text-left">
              <div class="font-bold">No. Register / NUP</div>
              <div class="text-[10px] text-slate-400">Nomor urut pencatatan</div>
            </div>
          </button>

          <button
            type="button"
            class="palette-btn"
            @click="addElement('field', 'location')"
          >
            <i class="pi pi-map-marker text-rose-400 text-base" />
            <div class="text-left">
              <div class="font-bold">Ruangan / Lokasi</div>
              <div class="text-[10px] text-slate-400">Letak penempatan fisik</div>
            </div>
          </button>

          <button
            type="button"
            class="palette-btn"
            @click="addElement('static_text')"
          >
            <i class="pi pi-pencil text-slate-300 text-base" />
            <div class="text-left">
              <div class="font-bold">Teks Statis Bebas</div>
              <div class="text-[10px] text-slate-400">Catatan peringatan / tahun</div>
            </div>
          </button>

          <button
            type="button"
            class="palette-btn"
            @click="addElement('line')"
          >
            <i class="pi pi-minus text-slate-400 text-base" />
            <div class="text-left">
              <div class="font-bold">Garis Pembatas</div>
              <div class="text-[10px] text-slate-400">Divider horizontal</div>
            </div>
          </button>

          <button
            type="button"
            class="palette-btn"
            @click="addElement('logo')"
          >
            <i class="pi pi-image text-amber-400 text-base" />
            <div class="text-left">
              <div class="font-bold">Logo Instansi / Gambar</div>
              <div class="text-[10px] text-slate-400">Logo Pemda / kantor</div>
            </div>
          </button>
        </div>

        <div class="mt-auto p-2 rounded bg-slate-800/80 border border-slate-700/60 text-[10.5px] text-slate-300 leading-relaxed">
          💡 <strong>Tips:</strong> Klik elemen di kanvas untuk menggeser (drag). Tarik kotak biru di pojok kanan-bawah untuk mengubah ukuran. Gunakan tombol panah keyboard untuk pergeseran presisi 0.5 mm.
        </div>
      </div>

      <!-- Center: Visual WYSIWYG Canvas Area -->
      <div
        class="flex-1 overflow-auto p-8 grid place-items-center relative bg-[radial-gradient(#334155_1px,transparent_1px)] [background-size:16px_16px]"
        @pointermove="onPointerMove"
        @pointerup="onPointerUp"
      >
        <!-- Actual Label Container in Millimeters -->
        <div
          ref="canvasRef"
          class="relative bg-white shadow-2xl rounded-sm border border-slate-400 text-black select-none overflow-hidden"
          :style="{
            width: `${activeTemplate.widthMm * MM_TO_PX * zoom}px`,
            height: `${activeTemplate.heightMm * MM_TO_PX * zoom}px`,
          }"
          @click.self="activeElementId = null"
        >
          <!-- Grid Overlay when active -->
          <div
            v-if="snapGrid"
            class="absolute inset-0 pointer-events-none opacity-20"
            :style="{
              backgroundImage: 'linear-gradient(to right, #000 1px, transparent 1px), linear-gradient(to bottom, #000 1px, transparent 1px)',
              backgroundSize: `${gridSizeMm * MM_TO_PX * zoom}px ${gridSizeMm * MM_TO_PX * zoom}px`,
            }"
          />

          <!-- Placed Elements on Canvas -->
          <div
            v-for="el in activeTemplate.elements"
            :key="el.id"
            class="absolute cursor-move transition-shadow"
            :class="[
              activeElementId === el.id ? 'ring-2 ring-indigo-600 bg-indigo-50/20 z-20' : 'hover:ring-1 hover:ring-slate-400 z-10',
            ]"
            :style="{
              left: `${el.x * MM_TO_PX * zoom}px`,
              top: `${el.y * MM_TO_PX * zoom}px`,
              width: `${el.w * MM_TO_PX * zoom}px`,
              height: `${el.h * MM_TO_PX * zoom}px`,
            }"
            @pointerdown="startDrag($event, el)"
            @click.stop="activeElementId = el.id"
          >
            <!-- 1. QR Code Element -->
            <div
              v-if="el.type === 'qr'"
              class="w-full h-full p-0.5 flex items-center justify-center pointer-events-none"
            >
              <img
                :src="`/api/barcode/${sampleItem.id}.png?fmt=qr`"
                class="w-full h-full object-contain pointer-events-none"
                alt="QR"
              />
            </div>

            <!-- 2. Barcode 128 Element -->
            <div
              v-else-if="el.type === 'code128'"
              class="w-full h-full p-0.5 flex flex-col items-center justify-center pointer-events-none"
            >
              <img
                :src="`/api/barcode/${sampleItem.id}.png?fmt=code128`"
                class="w-full h-[75%] object-contain pointer-events-none"
                alt="Barcode"
              />
              <span class="text-[8px] font-mono font-bold leading-tight truncate max-w-full">{{ sampleItem.sku }}</span>
            </div>

            <!-- 3. Header / Instansi Element -->
            <div
              v-else-if="el.type === 'header'"
              class="w-full h-full flex items-center pointer-events-none uppercase tracking-wide leading-tight truncate px-0.5"
              :class="[
                el.border ? 'border-b border-black' : '',
                el.align === 'center' ? 'justify-center text-center' : el.align === 'right' ? 'justify-end text-right' : 'justify-start text-left',
              ]"
              :style="{
                fontSize: `${(el.fontSize || 7) * (zoom / 2)}pt`,
                fontWeight: el.bold ? 800 : 600,
              }"
            >
              {{ el.text || 'INVENTARIS KANTOR' }}
            </div>

            <!-- 4. Field Element -->
            <div
              v-else-if="el.type === 'field'"
              class="w-full h-full flex items-center pointer-events-none leading-tight px-0.5 overflow-hidden"
              :class="[
                el.align === 'center' ? 'justify-center text-center' : el.align === 'right' ? 'justify-end text-right' : 'justify-start text-left',
              ]"
              :style="{
                fontSize: `${(el.fontSize || 7.5) * (zoom / 2)}pt`,
                fontWeight: el.bold ? 700 : 400,
                fontFamily: el.field === 'sku' ? 'monospace' : 'inherit',
              }"
            >
              <span class="truncate">{{ resolveFieldValue(sampleItem, el.field) }}</span>
            </div>

            <!-- 5. Static Text Element -->
            <div
              v-else-if="el.type === 'static_text'"
              class="w-full h-full flex items-center pointer-events-none leading-tight px-0.5 overflow-hidden"
              :class="[
                el.align === 'center' ? 'justify-center text-center' : el.align === 'right' ? 'justify-end text-right' : 'justify-start text-left',
              ]"
              :style="{
                fontSize: `${(el.fontSize || 7) * (zoom / 2)}pt`,
                fontWeight: el.bold ? 700 : 400,
              }"
            >
              <span class="truncate">{{ el.text }}</span>
            </div>

            <!-- 6. Line Divider Element -->
            <div
              v-else-if="el.type === 'line'"
              class="w-full h-full pointer-events-none border-t border-black my-auto"
            />

            <!-- 7. Logo / Image Element -->
            <div
              v-else-if="el.type === 'logo'"
              class="w-full h-full p-0.5 flex items-center justify-center pointer-events-none"
            >
              <img
                v-if="el.logoUrl"
                :src="el.logoUrl"
                class="w-full h-full object-contain pointer-events-none"
                alt="Logo"
              />
              <div
                v-else
                class="w-full h-full border border-dashed border-slate-400 bg-slate-50 flex flex-col items-center justify-center text-slate-400 p-1"
              >
                <i class="pi pi-image text-sm mb-0.5" />
                <span class="text-[7px] text-center font-bold">Upload Logo</span>
              </div>
            </div>

            <!-- Resize Handle for Active Element -->
            <div
              v-if="activeElementId === el.id"
              class="absolute -right-1 -bottom-1 w-3 h-3 bg-indigo-600 border border-white rounded-xs cursor-nwse-resize z-30"
              @pointerdown.stop="startResize($event, el)"
            />
          </div>
        </div>
      </div>

      <!-- Right Sidebar: Properties Inspector -->
      <div class="w-64 border-l border-slate-800 bg-slate-900/90 p-3.5 flex flex-col gap-3 text-xs overflow-y-auto select-none">
        <div class="font-bold text-slate-300 uppercase tracking-wider text-[10.5px] flex items-center justify-between">
          <span>Properti Elemen</span>
          <Tag v-if="activeElement" :value="activeElement.type" severity="info" class="!text-[9px] !py-0" />
        </div>

        <div v-if="activeElement" class="flex flex-col gap-3">
          <!-- Text Value for Header & Static -->
          <div v-if="activeElement.type === 'header' || activeElement.type === 'static_text'" class="flex flex-col gap-1">
            <label class="text-slate-400 text-[11px]">Teks Tampilan</label>
            <InputText v-model="activeElement.text" class="w-full !text-xs !py-1 !bg-slate-800 !text-white !border-slate-700" />
          </div>

          <!-- Dynamic Field Selector -->
          <div v-if="activeElement.type === 'field'" class="flex flex-col gap-1">
            <label class="text-slate-400 text-[11px]">Field Data Barang</label>
            <Select
              v-model="activeElement.field"
              :options="[
                { label: 'Nama Barang', value: 'name' },
                { label: 'Kode SKU', value: 'sku' },
                { label: 'Lokasi / Ruangan', value: 'location' },
                { label: 'Kategori', value: 'category' },
                { label: 'No. Register / NUP', value: 'register_number' },
                { label: 'Merk / Tipe', value: 'merk' },
              ]"
              optionLabel="label"
              optionValue="value"
              class="w-full !text-xs"
            />
          </div>

          <!-- Dimensions & Coordinates Grid (mm) -->
          <div class="grid grid-cols-2 gap-2 p-2 bg-slate-800/80 rounded border border-slate-700/80">
            <div class="flex flex-col gap-1">
              <label class="text-[10px] text-slate-400 font-mono">Posisi X (mm)</label>
              <InputNumber v-model="activeElement.x" :min="0" :step="0.5" class="w-full !text-xs" inputClass="!py-0.5 !text-xs !bg-slate-900 !text-white" />
            </div>
            <div class="flex flex-col gap-1">
              <label class="text-[10px] text-slate-400 font-mono">Posisi Y (mm)</label>
              <InputNumber v-model="activeElement.y" :min="0" :step="0.5" class="w-full !text-xs" inputClass="!py-0.5 !text-xs !bg-slate-900 !text-white" />
            </div>
            <div class="flex flex-col gap-1">
              <label class="text-[10px] text-slate-400 font-mono">Lebar (mm)</label>
              <InputNumber v-model="activeElement.w" :min="2" :step="0.5" class="w-full !text-xs" inputClass="!py-0.5 !text-xs !bg-slate-900 !text-white" />
            </div>
            <div class="flex flex-col gap-1">
              <label class="text-[10px] text-slate-400 font-mono">Tinggi (mm)</label>
              <InputNumber v-model="activeElement.h" :min="1" :step="0.5" class="w-full !text-xs" inputClass="!py-0.5 !text-xs !bg-slate-900 !text-white" />
            </div>
          </div>

          <!-- Logo Image Controls -->
          <div v-if="activeElement.type === 'logo'" class="flex flex-col gap-2 p-2.5 bg-slate-800/80 rounded border border-slate-700/80">
            <label class="text-[11px] text-slate-300 font-bold">Gambar Logo</label>
            <div v-if="activeElement.logoUrl" class="flex items-center gap-2">
              <img :src="activeElement.logoUrl" class="w-12 h-12 object-contain bg-white rounded p-1 border border-slate-600" alt="Preview" />
              <div class="flex flex-col gap-1 flex-1">
                <span class="text-[10px] text-emerald-400">✓ Logo terpasang</span>
                <button
                  type="button"
                  class="text-[10.5px] text-rose-400 hover:text-rose-300 text-left font-semibold"
                  @click="activeElement.logoUrl = ''"
                >
                  Hapus Gambar
                </button>
              </div>
            </div>

            <label class="cursor-pointer w-full flex items-center justify-center gap-2 py-1.5 px-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded text-xs font-semibold transition-colors">
              <i class="pi pi-upload text-xs" />
              <span>{{ activeElement.logoUrl ? 'Ganti File Logo' : 'Unggah File Logo (PNG/JPG)' }}</span>
              <input type="file" accept="image/*" class="hidden" @change="handleLogoUpload" />
            </label>
            <span class="text-[9.5px] text-slate-400">Rekomendasi file PNG transparan untuk hasil cetak bersih.</span>
          </div>

          <!-- Typography (Font Size, Bold, Alignment) -->
          <div v-if="!['qr', 'code128', 'line', 'logo'].includes(activeElement.type)" class="flex flex-col gap-2 p-2 bg-slate-800/80 rounded border border-slate-700/80">
            <div class="flex items-center justify-between">
              <label class="text-[10.5px] text-slate-400">Ukuran Font</label>
              <InputNumber v-model="activeElement.fontSize" :min="4" :max="36" suffix=" pt" class="!w-24 !text-xs" inputClass="!py-0.5 !text-xs !bg-slate-900 !text-white text-center" />
            </div>

            <div class="flex items-center justify-between pt-1 border-t border-slate-700">
              <label class="flex items-center gap-1.5 cursor-pointer text-slate-300 text-[11px]">
                <Checkbox v-model="activeElement.bold" binary />
                <span>Huruf Tebal (Bold)</span>
              </label>

              <label v-if="activeElement.type === 'header'" class="flex items-center gap-1.5 cursor-pointer text-slate-300 text-[11px]">
                <Checkbox v-model="activeElement.border" binary />
                <span>Garis Bawah</span>
              </label>
            </div>

            <div class="flex items-center justify-between pt-1 border-t border-slate-700">
              <span class="text-[10.5px] text-slate-400">Rata Teks:</span>
              <div class="flex items-center gap-1">
                <button
                  type="button"
                  class="align-btn"
                  :class="activeElement.align === 'left' ? 'bg-indigo-600 text-white' : 'text-slate-400 hover:bg-slate-700'"
                  @click="activeElement.align = 'left'"
                >
                  <i class="pi pi-align-left text-xs" />
                </button>
                <button
                  type="button"
                  class="align-btn"
                  :class="activeElement.align === 'center' ? 'bg-indigo-600 text-white' : 'text-slate-400 hover:bg-slate-700'"
                  @click="activeElement.align = 'center'"
                >
                  <i class="pi pi-align-center text-xs" />
                </button>
                <button
                  type="button"
                  class="align-btn"
                  :class="activeElement.align === 'right' ? 'bg-indigo-600 text-white' : 'text-slate-400 hover:bg-slate-700'"
                  @click="activeElement.align = 'right'"
                >
                  <i class="pi pi-align-right text-xs" />
                </button>
              </div>
            </div>
          </div>

          <!-- Actions: Duplicate & Delete -->
          <div class="flex items-center gap-2 pt-2">
            <Button
              icon="pi pi-copy"
              label="Duplikat"
              size="small"
              severity="secondary"
              class="w-1/2 !text-xs !py-1"
              @click="duplicateActiveElement"
            />
            <Button
              icon="pi pi-trash"
              label="Hapus"
              size="small"
              severity="danger"
              class="w-1/2 !text-xs !py-1"
              @click="removeActiveElement"
            />
          </div>
        </div>

        <div v-else class="flex flex-col items-center justify-center p-8 text-center text-slate-500">
          <i class="pi pi-mouse text-2xl mb-2 opacity-50" />
          <p>Klik salah satu elemen pada kanvas stiker untuk mengatur propertinya.</p>
        </div>
      </div>
    </div>
  </Dialog>

  <!-- PURE PRINT ENGINE (Rendered only on window.print) -->
  <teleport to="body">
    <div v-if="isPrinting" id="custom-print-sheet">
      <div
        v-for="(it, iIdx) in itemsToPrint"
        :key="iIdx"
        class="custom-print-label"
        :style="{
          width: `${activeTemplate.widthMm}mm`,
          height: `${activeTemplate.heightMm}mm`,
        }"
      >
        <div
          v-for="el in activeTemplate.elements"
          :key="el.id"
          class="custom-print-el"
          :style="{
            left: `${el.x}mm`,
            top: `${el.y}mm`,
            width: `${el.w}mm`,
            height: `${el.h}mm`,
          }"
        >
          <!-- QR -->
          <img
            v-if="el.type === 'qr'"
            :src="`/api/barcode/${it.id}.png?fmt=qr`"
            class="w-full h-full object-contain"
            alt="QR"
          />

          <!-- Barcode 128 -->
          <div v-else-if="el.type === 'code128'" class="w-full h-full flex flex-col items-center justify-center">
            <img :src="`/api/barcode/${it.id}.png?fmt=code128`" class="w-full h-[75%] object-contain" alt="BC" />
            <span class="text-[7.5pt] font-mono font-bold leading-none truncate">{{ it.sku }}</span>
          </div>

          <!-- Header -->
          <div
            v-else-if="el.type === 'header'"
            class="w-full h-full flex items-center uppercase tracking-wide leading-tight truncate"
            :class="[
              el.border ? 'border-b border-black' : '',
              el.align === 'center' ? 'justify-center text-center' : el.align === 'right' ? 'justify-end text-right' : 'justify-start text-left',
            ]"
            :style="{
              fontSize: `${el.fontSize || 7}pt`,
              fontWeight: el.bold ? 800 : 600,
            }"
          >
            {{ el.text }}
          </div>

          <!-- Field -->
          <div
            v-else-if="el.type === 'field'"
            class="w-full h-full flex items-center leading-tight truncate"
            :class="[
              el.align === 'center' ? 'justify-center text-center' : el.align === 'right' ? 'justify-end text-right' : 'justify-start text-left',
            ]"
            :style="{
              fontSize: `${el.fontSize || 7.5}pt`,
              fontWeight: el.bold ? 700 : 400,
              fontFamily: el.field === 'sku' ? 'monospace' : 'inherit',
            }"
          >
            {{ resolveFieldValue(it, el.field) }}
          </div>

          <!-- Static Text -->
          <div
            v-else-if="el.type === 'static_text'"
            class="w-full h-full flex items-center leading-tight truncate"
            :class="[
              el.align === 'center' ? 'justify-center text-center' : el.align === 'right' ? 'justify-end text-right' : 'justify-start text-left',
            ]"
            :style="{
              fontSize: `${el.fontSize || 7}pt`,
              fontWeight: el.bold ? 700 : 400,
            }"
          >
            {{ el.text }}
          </div>

          <!-- Line -->
          <div
            v-else-if="el.type === 'line'"
            class="w-full h-full border-t border-black my-auto"
          />

          <!-- Logo -->
          <img
            v-else-if="el.type === 'logo' && el.logoUrl"
            :src="el.logoUrl"
            class="w-full h-full object-contain"
            alt="Logo"
          />
        </div>
      </div>
    </div>
  </teleport>
</template>

<style scoped>
.palette-btn {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  background-color: rgba(30, 41, 59, 0.7);
  border: 1px solid rgba(51, 65, 85, 0.8);
  border-radius: 6px;
  color: #f1f5f9;
  cursor: pointer;
  transition: all 0.15s ease;
}
.palette-btn:hover {
  background-color: #334155;
  border-color: #6366f1;
}

.align-btn {
  padding: 3px 6px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.1s ease;
}
</style>

<style>
@media print {
  body * {
    visibility: hidden;
  }
  #custom-print-sheet,
  #custom-print-sheet * {
    visibility: visible;
  }
  #custom-print-sheet {
    position: absolute;
    left: 0;
    top: 0;
    margin: 0;
    padding: 0;
    background: transparent;
  }
  .custom-print-label {
    position: relative;
    box-sizing: border-box;
    overflow: hidden;
    background: #ffffff;
    page-break-after: always;
    break-after: page;
  }
  .custom-print-el {
    position: absolute;
    box-sizing: border-box;
    display: flex;
    overflow: hidden;
  }
}
</style>
