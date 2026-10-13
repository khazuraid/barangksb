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
  field?: string
  prefix?: string // Teks awalan otomatis, misal: "NUP: ", "Thn: ", "Merk: "
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
const zoom = ref(2.2) // Skala tampilan layar kanvas
const snapGrid = ref(true)
const gridSizeMm = ref(0.5)
const showGuideGrid = ref(false) // Grid garis visual di stiker (default mati agar stiker bersih)

// Daftar Lengkap Field dari Formulir Barang
const AVAILABLE_FIELDS = [
  { label: 'Nama Barang', value: 'name', icon: 'pi pi-box', defaultPrefix: '' },
  { label: 'Kode Barang (KIB / BMD)', value: 'item_code', icon: 'pi pi-tag', defaultPrefix: 'Kode: ' },
  { label: 'Kode Barcode / SKU', value: 'sku', icon: 'pi pi-hashtag', defaultPrefix: '' },
  { label: 'No. Register / NUP', value: 'register_number', icon: 'pi pi-id-card', defaultPrefix: 'NUP: ' },
  { label: 'Ruangan / Lokasi', value: 'location', icon: 'pi pi-map-marker', defaultPrefix: '📍 ' },
  { label: 'Kode Ruangan / Lokasi', value: 'location_code', icon: 'pi pi-compass', defaultPrefix: 'Ruang: ' },
  { label: 'Merk / Brand', value: 'merk', icon: 'pi pi-bookmark', defaultPrefix: 'Merk: ' },
  { label: 'Tipe / Model', value: 'type_model', icon: 'pi pi-cog', defaultPrefix: 'Tipe: ' },
  { label: 'Nomor Seri Pabrik (SN)', value: 'serial_number', icon: 'pi pi-barcode', defaultPrefix: 'SN: ' },
  { label: 'Tahun Pengadaan', value: 'procurement_year', icon: 'pi pi-calendar', defaultPrefix: 'Thn: ' },
  { label: 'Asal Usul / Sumber Dana', value: 'funding_source', icon: 'pi pi-dollar', defaultPrefix: 'Dana: ' },
  { label: 'Kategori Barang', value: 'category', icon: 'pi pi-folder', defaultPrefix: 'Kat: ' },
  { label: 'Kondisi Barang', value: 'condition_status', icon: 'pi pi-shield', defaultPrefix: 'Kondisi: ' },
  { label: 'Ukuran / Dimensi / CC', value: 'size', icon: 'pi pi-arrows-alt', defaultPrefix: 'Ukuran: ' },
  { label: 'Bahan / Material', value: 'material', icon: 'pi pi-database', defaultPrefix: 'Bahan: ' },
  { label: 'Harga Satuan (Rp)', value: 'price_per_unit', icon: 'pi pi-money-bill', defaultPrefix: 'Rp ' },
  { label: 'Keterangan', value: 'description', icon: 'pi pi-info-circle', defaultPrefix: 'Ket: ' },
]

// Preset Desain Standar yang Rapi & Proporsional
const DEFAULT_PRESETS: LabelTemplate[] = [
  {
    id: 'preset_tj_103',
    name: '🌟 Tom & Jerry No. 103 (64 × 32 mm · Rekomendasi)',
    widthMm: 64,
    heightMm: 32,
    elements: [
      { id: 'hdr_1', type: 'header', text: 'PEMERINTAH KABUPATEN SUMBAWA BARAT', x: 2, y: 1.8, w: 60, h: 4, fontSize: 6.5, bold: true, align: 'center', border: true },
      { id: 'qr_1', type: 'qr', x: 2.5, y: 7, w: 23, h: 23 },
      { id: 'f_name', type: 'field', field: 'name', x: 27, y: 7, w: 35, h: 8, fontSize: 7.5, bold: true, align: 'left' },
      { id: 'f_sku', type: 'field', field: 'sku', prefix: '', x: 27, y: 15.5, w: 35, h: 4, fontSize: 7, bold: true, align: 'left' },
      { id: 'f_nup', type: 'field', field: 'register_number', prefix: 'NUP: ', x: 27, y: 20, w: 35, h: 3.5, fontSize: 6.5, bold: false, align: 'left' },
      { id: 'f_loc', type: 'field', field: 'location', prefix: '📍 ', x: 27, y: 24, w: 35, h: 3.5, fontSize: 6.5, bold: false, align: 'left' },
      { id: 'f_yr', type: 'field', field: 'procurement_year', prefix: 'Thn: ', x: 27, y: 27.5, w: 35, h: 3, fontSize: 6, bold: false, align: 'left' },
    ],
  },
  {
    id: 'preset_50x30',
    name: '🏷️ Stiker Rol Standar (50 × 30 mm)',
    widthMm: 50,
    heightMm: 30,
    elements: [
      { id: 'hdr_1', type: 'header', text: 'INVENTARIS KANTOR', x: 2, y: 2, w: 46, h: 3.5, fontSize: 6.5, bold: true, align: 'left', border: true },
      { id: 'qr_1', type: 'qr', x: 2, y: 6.5, w: 21, h: 21 },
      { id: 'f_name', type: 'field', field: 'name', x: 24.5, y: 6.5, w: 24, h: 8, fontSize: 7.5, bold: true, align: 'left' },
      { id: 'f_sku', type: 'field', field: 'sku', x: 24.5, y: 15, w: 24, h: 4, fontSize: 7, bold: true, align: 'left' },
      { id: 'f_nup', type: 'field', field: 'register_number', prefix: 'NUP: ', x: 24.5, y: 19.5, w: 24, h: 3.5, fontSize: 6.5, bold: false, align: 'left' },
      { id: 'f_loc', type: 'field', field: 'location', prefix: '📍 ', x: 24.5, y: 23.5, w: 24, h: 3.5, fontSize: 6.5, bold: false, align: 'left' },
    ],
  },
  {
    id: 'preset_40x30',
    name: '🏷️ Stiker Rol Mini (40 × 30 mm)',
    widthMm: 40,
    heightMm: 30,
    elements: [
      { id: 'qr_1', type: 'qr', x: 2, y: 3, w: 23, h: 23 },
      { id: 'f_name', type: 'field', field: 'name', x: 26, y: 3, w: 12.5, h: 10, fontSize: 7, bold: true, align: 'left' },
      { id: 'f_sku', type: 'field', field: 'sku', x: 26, y: 14, w: 12.5, h: 4, fontSize: 6.5, bold: true, align: 'left' },
      { id: 'f_loc', type: 'field', field: 'location', prefix: '', x: 26, y: 19, w: 12.5, h: 4, fontSize: 6, bold: false, align: 'left' },
    ],
  },
  {
    id: 'preset_70x50',
    name: '🏷️ Stiker Aset BMD Lengkap (70 × 50 mm)',
    widthMm: 70,
    heightMm: 50,
    elements: [
      { id: 'hdr_1', type: 'header', text: 'KARTU INVENTARIS BARANG MILIK DAERAH (BMD)', x: 3, y: 2.5, w: 64, h: 4.5, fontSize: 7.5, bold: true, align: 'center', border: true },
      { id: 'qr_1', type: 'qr', x: 3.5, y: 8.5, w: 28, h: 28 },
      { id: 'f_name', type: 'field', field: 'name', x: 33, y: 8.5, w: 34, h: 9, fontSize: 8.5, bold: true, align: 'left' },
      { id: 'f_code', type: 'field', field: 'item_code', prefix: 'Kode: ', x: 33, y: 18, w: 34, h: 4, fontSize: 7, bold: false, align: 'left' },
      { id: 'f_sku', type: 'field', field: 'sku', prefix: 'Barcode: ', x: 33, y: 22.5, w: 34, h: 4, fontSize: 7, bold: true, align: 'left' },
      { id: 'f_nup', type: 'field', field: 'register_number', prefix: 'No. Reg: ', x: 33, y: 27, w: 34, h: 4, fontSize: 7, bold: false, align: 'left' },
      { id: 'f_merk', type: 'field', field: 'merk', prefix: 'Merk/Tipe: ', x: 33, y: 31.5, w: 34, h: 4, fontSize: 6.5, bold: false, align: 'left' },
      { id: 'f_loc', type: 'field', field: 'location', prefix: 'Ruangan: ', x: 33, y: 36, w: 34, h: 4, fontSize: 6.5, bold: false, align: 'left' },
      { id: 'f_yr', type: 'field', field: 'procurement_year', prefix: 'Tahun: ', x: 3.5, y: 38, w: 28, h: 3.5, fontSize: 6.5, bold: false, align: 'center' },
      { id: 'line_1', type: 'line', x: 3, y: 42.5, w: 64, h: 1 },
      { id: 'st_1', type: 'static_text', text: 'DILARANG DILEPAS / DIPINDAHKAN TANPA IZIN PENGELOLA BMD', x: 3, y: 44, w: 64, h: 3.5, fontSize: 5.5, bold: false, align: 'center' },
    ],
  },
]

const templates = ref<LabelTemplate[]>([])
const activeTemplate = ref<LabelTemplate>(JSON.parse(JSON.stringify(DEFAULT_PRESETS[0])))
const activeElementId = ref<string | null>(null)
const copies = ref(1)
const isPrinting = ref(false)
const paletteTab = ref<'form' | 'barcode' | 'header' | 'text'>('form')

function loadSavedTemplates() {
  const raw = localStorage.getItem('custom_label_templates_v3')
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
  localStorage.setItem('custom_label_templates_v3', JSON.stringify(templates.value))
}

onMounted(() => {
  loadSavedTemplates()
})

const activeElement = computed(() => {
  if (!activeElementId.value) return null
  return activeTemplate.value.elements.find((el) => el.id === activeElementId.value) || null
})

// Sampel Barang Nyata untuk Pratinjau
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
      item_code: '02.06.01.02.001',
      register_number: '0001 / BMD / 2024',
      location: 'Ruang Server IT (Lantai 2)',
      location_code: 'R.02',
      category: 'Elektronik & IT',
      merk: 'Lenovo ThinkPad',
      type_model: 'E14 Core i5',
      serial_number: 'SN-PF41A92X',
      procurement_year: '2024',
      funding_source: 'APBD 2024',
      condition_status: 'Baik (B)',
      size: '14 Inch',
      material: 'Aluminium / Polimer',
      price_per_unit: 14500000,
      description: 'Kondisi lengkap dengan adaptor',
    }
  )
})

// Nilai field yang dirender
function resolveFieldValue(item: any, fieldKey?: string, prefix?: string): string {
  if (!fieldKey) return ''
  let val = item[fieldKey]
  if (val === undefined || val === null || val === '') {
    // Fallback sample values if empty
    switch (fieldKey) {
      case 'name': val = item.name || 'Nama Aset'; break
      case 'sku': val = item.sku || 'AST-0001'; break
      case 'item_code': val = item.item_code || '02.06.01.001'; break
      case 'register_number': val = item.register_number || '0001'; break
      case 'location': val = item.location || 'Gudang Utama'; break
      case 'location_code': val = item.location_code || 'G.01'; break
      case 'merk': val = item.merk || 'Standar'; break
      case 'type_model': val = item.type_model || '-'; break
      case 'serial_number': val = item.serial_number || 'SN-12345'; break
      case 'procurement_year': val = item.procurement_year || '2024'; break
      case 'funding_source': val = item.funding_source || 'APBD'; break
      case 'condition_status': val = item.condition_status || 'Baik'; break
      case 'price_per_unit': val = item.price_per_unit ? `Rp ${Number(item.price_per_unit).toLocaleString('id-ID')}` : '-'; break
      default: val = '-'
    }
  } else if (fieldKey === 'price_per_unit') {
    val = `Rp ${Number(val).toLocaleString('id-ID')}`
  }

  const pfx = prefix || ''
  return `${pfx}${val}`
}

// Drag & Resize Engine
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

function onPointerUp() {
  isDragging = false
  isResizing = false
}

// Tambah Elemen Baru
function addFieldElement(fieldKey: string, defaultPrefix = '') {
  const id = `f_${fieldKey}_${Date.now().toString(36)}`
  const newEl: LabelElement = {
    id,
    type: 'field',
    field: fieldKey,
    prefix: defaultPrefix,
    x: 4,
    y: Math.min(activeTemplate.value.heightMm - 6, activeTemplate.value.elements.length * 4 + 4),
    w: 32,
    h: fieldKey === 'name' ? 8 : 4.5,
    fontSize: fieldKey === 'name' ? 8 : 7,
    bold: fieldKey === 'name' || fieldKey === 'sku',
    align: 'left',
  }
  activeTemplate.value.elements.push(newEl)
  activeElementId.value = id
  toast.add({ severity: 'info', summary: 'Field Ditambahkan', detail: fieldKey, life: 1200 })
}

function addOtherElement(type: 'qr' | 'code128' | 'header' | 'static_text' | 'line' | 'logo') {
  const id = `${type}_${Date.now().toString(36)}`
  let newEl: LabelElement

  switch (type) {
    case 'qr':
      newEl = { id, type: 'qr', x: 2.5, y: 4, w: 22, h: 22 }
      break
    case 'code128':
      newEl = { id, type: 'code128', x: 2.5, y: 4, w: 36, h: 12 }
      break
    case 'header':
      newEl = { id, type: 'header', text: 'PEMERINTAH KABUPATEN KSB', x: 2, y: 2, w: activeTemplate.value.widthMm - 4, h: 4, fontSize: 6.5, bold: true, align: 'center', border: true }
      break
    case 'static_text':
      newEl = { id, type: 'static_text', text: 'MILIK PEMERINTAH DAERAH', x: 2, y: 4, w: 28, h: 4, fontSize: 6.5, bold: false, align: 'left' }
      break
    case 'line':
      newEl = { id, type: 'line', x: 2, y: 10, w: activeTemplate.value.widthMm - 4, h: 1 }
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
      toast.add({ severity: 'success', summary: 'Logo Berhasil Dipasang', life: 1800 })
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

function applyPreset(tpl: LabelTemplate) {
  activeTemplate.value = JSON.parse(JSON.stringify(tpl))
  activeElementId.value = null
  toast.add({ severity: 'info', summary: 'Preset Dimuat', detail: tpl.name, life: 1800 })
}

function resetLayout() {
  const defaultTpl = DEFAULT_PRESETS.find((p) => p.id === activeTemplate.value.id) || DEFAULT_PRESETS[0]
  activeTemplate.value = JSON.parse(JSON.stringify(defaultTpl))
  activeElementId.value = null
  toast.add({ severity: 'info', summary: 'Tata Letak Direset ke Bawaan', life: 1500 })
}

function saveCurrentAsNewTemplate() {
  const name = prompt('Beri nama template label ini:', activeTemplate.value.name.replace(/^[🌟🏷️\s]+/, ''))
  if (!name) return
  const newTpl: LabelTemplate = {
    ...JSON.parse(JSON.stringify(activeTemplate.value)),
    id: `custom_${Date.now().toString(36)}`,
    name: `🏷️ ${name.trim()}`,
  }
  templates.value.push(newTpl)
  saveTemplates()
  toast.add({ severity: 'success', summary: 'Template Berhasil Disimpan', detail: newTpl.name, life: 2000 })
}

function exportJSON() {
  const blob = new Blob([JSON.stringify(activeTemplate.value, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `label_${activeTemplate.value.widthMm}x${activeTemplate.value.heightMm}mm.json`
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
        toast.add({ severity: 'success', summary: 'Template Diimpor', life: 2000 })
      } else {
        toast.add({ severity: 'error', summary: 'File JSON template tidak valid', life: 2500 })
      }
    } catch {
      toast.add({ severity: 'error', summary: 'Gagal membaca file JSON', life: 2500 })
    }
  }
  reader.readAsText(file)
}

// Navigasi Nudge dengan Keyboard Arrow Keys
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

// Data Barang yang Akan Dicetak
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
  }, 350)
}
</script>

<template>
  <Dialog
    :visible="visible"
    modal
    header="Desainer Label QR &amp; Barcode (Drag &amp; Drop)"
    :style="{ width: '96vw', maxWidth: '1360px', height: '94vh' }"
    :contentStyle="{ padding: '0', display: 'flex', flexDirection: 'column', overflow: 'hidden' }"
    @update:visible="emit('update:visible', $event)"
  >
    <!-- TOP TOOLBAR: Clean, Professional, Light -->
    <div class="px-4 py-2.5 border-b border-slate-200 bg-white flex flex-wrap items-center justify-between gap-3 text-xs select-none shadow-xs">
      <!-- Left: Preset & Size Controls -->
      <div class="flex items-center gap-2.5 flex-wrap">
        <div class="flex items-center gap-1.5">
          <span class="text-slate-500 font-semibold">Preset:</span>
          <Select
            :modelValue="activeTemplate.id"
            :options="templates"
            optionLabel="name"
            optionValue="id"
            placeholder="Pilih Preset..."
            class="!w-[280px] !text-xs !py-0.5"
            @update:modelValue="(val) => {
              const found = templates.find((t) => t.id === val)
              if (found) applyPreset(found)
            }"
          />
        </div>

        <div class="flex items-center gap-1 bg-slate-100 p-1 rounded-md border border-slate-200">
          <span class="text-slate-500 text-[11px] font-medium px-1">Ukuran:</span>
          <InputNumber
            v-model="activeTemplate.widthMm"
            :min="20"
            :max="210"
            suffix=" mm"
            class="!w-[76px] !text-xs"
            inputClass="!py-0.5 !text-xs !bg-white !text-slate-800 text-center font-mono font-semibold"
          />
          <span class="text-slate-400 font-bold">&times;</span>
          <InputNumber
            v-model="activeTemplate.heightMm"
            :min="10"
            :max="297"
            suffix=" mm"
            class="!w-[76px] !text-xs"
            inputClass="!py-0.5 !text-xs !bg-white !text-slate-800 text-center font-mono font-semibold"
          />
        </div>

        <Button
          icon="pi pi-refresh"
          label="Reset"
          size="small"
          text
          severity="secondary"
          v-tooltip.top="'Kembalikan tata letak ke preset awal'"
          class="!text-xs !py-1"
          @click="resetLayout"
        />
      </div>

      <!-- Right: View Controls & Print Action -->
      <div class="flex items-center gap-2 flex-wrap">
        <!-- Zoom Controller -->
        <div class="flex items-center bg-slate-100 rounded-md border border-slate-200 px-1 py-0.5">
          <button
            type="button"
            class="px-2 py-0.5 text-slate-600 hover:text-slate-900 font-bold"
            title="Perkecil Kanvas"
            @click="zoom = Math.max(1.2, zoom - 0.3)"
          >
            -
          </button>
          <span class="text-xs font-mono font-semibold text-slate-700 w-12 text-center">
            {{ Math.round(zoom * 45) }}%
          </span>
          <button
            type="button"
            class="px-2 py-0.5 text-slate-600 hover:text-slate-900 font-bold"
            title="Perbesar Kanvas"
            @click="zoom = Math.min(3.5, zoom + 0.3)"
          >
            +
          </button>
        </div>

        <!-- Snap Grid Toggle -->
        <label class="flex items-center gap-1.5 cursor-pointer text-slate-600 text-xs bg-slate-100 hover:bg-slate-200 px-2 py-1.5 rounded-md border border-slate-200">
          <Checkbox v-model="snapGrid" binary />
          <span>Snap 0.5mm</span>
        </label>

        <!-- Save Template -->
        <Button
          icon="pi pi-save"
          label="Simpan Template"
          size="small"
          severity="secondary"
          class="!text-xs !py-1"
          @click="saveCurrentAsNewTemplate"
        />

        <!-- Export / Import -->
        <Button icon="pi pi-download" text rounded size="small" severity="secondary" v-tooltip.top="'Ekspor JSON Template'" @click="exportJSON" />
        <label class="cursor-pointer p-1.5 rounded-md hover:bg-slate-100 text-slate-600" v-tooltip.top="'Impor File JSON'">
          <i class="pi pi-upload text-sm" />
          <input type="file" accept=".json" class="hidden" @change="importJSON" />
        </label>

        <span class="h-4 w-px bg-slate-200 mx-1" />

        <!-- Print Action -->
        <div class="flex items-center gap-1.5">
          <InputNumber
            v-model="copies"
            :min="1"
            :max="50"
            prefix="&times; "
            class="!w-[65px] !text-xs"
            inputClass="!py-1 !text-xs !bg-slate-50 !text-slate-800 text-center font-bold"
            v-tooltip.top="'Jumlah rangkap cetak per barang'"
          />
          <Button
            label="Cetak Label Ini"
            icon="pi pi-print"
            size="small"
            severity="success"
            class="!text-xs !py-1 font-bold shadow-xs"
            @click="printLabels"
          />
        </div>
      </div>
    </div>

    <!-- MAIN WORKSPACE: 3 COLUMNS -->
    <div class="flex-1 flex overflow-hidden bg-slate-100">
      <!-- LEFT COLUMN: PALETTE OF ELEMENTS FROM FORM & BARCODES -->
      <div class="w-68 border-r border-slate-200 bg-white flex flex-col overflow-hidden select-none">
        <!-- Palette Tab Headers -->
        <div class="p-2 border-b border-slate-200 bg-slate-50 flex items-center gap-1">
          <button
            type="button"
            class="tab-btn"
            :class="paletteTab === 'form' ? 'tab-btn-active' : ''"
            @click="paletteTab = 'form'"
          >
            <i class="pi pi-list text-xs mr-1" />
            <span>Data Formulir</span>
          </button>
          <button
            type="button"
            class="tab-btn"
            :class="paletteTab === 'barcode' ? 'tab-btn-active' : ''"
            @click="paletteTab = 'barcode'"
          >
            <i class="pi pi-qrcode text-xs mr-1" />
            <span>Barcode &amp; Kop</span>
          </button>
        </div>

        <!-- Palette Content: TAB 1 - ALL FIELDS FROM FORM -->
        <div v-if="paletteTab === 'form'" class="flex-1 overflow-y-auto p-3 flex flex-col gap-2">
          <div class="text-[11px] font-bold text-slate-500 uppercase tracking-wider mb-1">
            Klik field untuk taruh di stiker:
          </div>

          <div class="flex flex-col gap-1.5">
            <button
              v-for="fld in AVAILABLE_FIELDS"
              :key="fld.value"
              type="button"
              class="field-item-btn"
              @click="addFieldElement(fld.value, fld.defaultPrefix)"
            >
              <div class="flex items-center gap-2 min-w-0">
                <i :class="[fld.icon, 'text-indigo-600 text-xs shrink-0']" />
                <span class="font-semibold text-slate-800 truncate text-[11.5px]">{{ fld.label }}</span>
              </div>
              <span class="text-[10px] text-slate-400 font-mono shrink-0">+ Taruh</span>
            </button>
          </div>
        </div>

        <!-- Palette Content: TAB 2 - BARCODE, HEADER, LOGO, DIVIDER, TEXT -->
        <div v-else class="flex-1 overflow-y-auto p-3 flex flex-col gap-2.5">
          <div class="text-[11px] font-bold text-slate-500 uppercase tracking-wider">
            Elemen Visual:
          </div>

          <button
            type="button"
            class="component-item-btn"
            @click="addOtherElement('qr')"
          >
            <i class="pi pi-qrcode text-indigo-600 text-lg" />
            <div class="text-left flex-1 min-w-0">
              <div class="font-bold text-slate-800 text-[11.5px]">QR Code</div>
              <div class="text-[10px] text-slate-500">Scan link aset interaktif</div>
            </div>
            <i class="pi pi-plus text-slate-400 text-xs" />
          </button>

          <button
            type="button"
            class="component-item-btn"
            @click="addOtherElement('code128')"
          >
            <i class="pi pi-barcode text-cyan-600 text-lg" />
            <div class="text-left flex-1 min-w-0">
              <div class="font-bold text-slate-800 text-[11.5px]">Barcode 128</div>
              <div class="text-[10px] text-slate-500">Barcode garis SKU aset</div>
            </div>
            <i class="pi pi-plus text-slate-400 text-xs" />
          </button>

          <button
            type="button"
            class="component-item-btn"
            @click="addOtherElement('header')"
          >
            <i class="pi pi-building text-amber-600 text-lg" />
            <div class="text-left flex-1 min-w-0">
              <div class="font-bold text-slate-800 text-[11.5px]">Header / Kop Instansi</div>
              <div class="text-[10px] text-slate-500">Judul kop instansi atas</div>
            </div>
            <i class="pi pi-plus text-slate-400 text-xs" />
          </button>

          <button
            type="button"
            class="component-item-btn"
            @click="addOtherElement('logo')"
          >
            <i class="pi pi-image text-emerald-600 text-lg" />
            <div class="text-left flex-1 min-w-0">
              <div class="font-bold text-slate-800 text-[11.5px]">Logo Instansi / Kantor</div>
              <div class="text-[10px] text-slate-500">Unggah file logo PNG/JPG</div>
            </div>
            <i class="pi pi-plus text-slate-400 text-xs" />
          </button>

          <button
            type="button"
            class="component-item-btn"
            @click="addOtherElement('line')"
          >
            <i class="pi pi-minus text-slate-500 text-lg" />
            <div class="text-left flex-1 min-w-0">
              <div class="font-bold text-slate-800 text-[11.5px]">Garis Pembatas</div>
              <div class="text-[10px] text-slate-500">Garis pemisah horizontal</div>
            </div>
            <i class="pi pi-plus text-slate-400 text-xs" />
          </button>

          <button
            type="button"
            class="component-item-btn"
            @click="addOtherElement('static_text')"
          >
            <i class="pi pi-pencil text-purple-600 text-lg" />
            <div class="text-left flex-1 min-w-0">
              <div class="font-bold text-slate-800 text-[11.5px]">Teks Bebas Statis</div>
              <div class="text-[10px] text-slate-500">Teks peringatan / keterangan</div>
            </div>
            <i class="pi pi-plus text-slate-400 text-xs" />
          </button>
        </div>

        <div class="p-2.5 border-t border-slate-200 bg-slate-50 text-[11px] text-slate-500 leading-tight">
          💡 Geser elemen langsung di stiker. Tarik titik biru di pojok untuk ubah ukuran.
        </div>
      </div>

      <!-- CENTER COLUMN: VISUAL STICKER CANVAS -->
      <div
        class="flex-1 overflow-auto p-8 grid place-items-center relative bg-slate-200/60"
        @pointermove="onPointerMove"
        @pointerup="onPointerUp"
      >
        <div class="flex flex-col items-center gap-2">
          <!-- Sticker Label Badge Info -->
          <div class="flex items-center gap-2 text-xs font-semibold text-slate-600 bg-white/80 px-3 py-1 rounded-full border border-slate-300 shadow-xs">
            <span>🏷️ {{ activeTemplate.widthMm }} &times; {{ activeTemplate.heightMm }} mm</span>
            <span class="text-slate-300">|</span>
            <span class="text-slate-500 font-normal">Sampel: {{ sampleItem.name }}</span>
          </div>

          <!-- THE PHYSICAL STICKER (White card with clean border and drop shadow) -->
          <div
            class="relative bg-white shadow-xl rounded-[4px] border border-slate-300 text-black select-none overflow-hidden transition-all"
            :style="{
              width: `${activeTemplate.widthMm * MM_TO_PX * zoom}px`,
              height: `${activeTemplate.heightMm * MM_TO_PX * zoom}px`,
            }"
            @click.self="activeElementId = null"
          >
            <!-- Optional Faint Grid Guides -->
            <div
              v-if="showGuideGrid"
              class="absolute inset-0 pointer-events-none opacity-10"
              :style="{
                backgroundImage: 'linear-gradient(to right, #000 1px, transparent 1px), linear-gradient(to bottom, #000 1px, transparent 1px)',
                backgroundSize: `${gridSizeMm * 2 * MM_TO_PX * zoom}px ${gridSizeMm * 2 * MM_TO_PX * zoom}px`,
              }"
            />

            <!-- Placed Elements on Sticker -->
            <div
              v-for="el in activeTemplate.elements"
              :key="el.id"
              class="absolute cursor-move transition-shadow"
              :class="[
                activeElementId === el.id
                  ? 'ring-2 ring-indigo-600 ring-offset-1 bg-indigo-50/20 z-20'
                  : 'hover:ring-1 hover:ring-indigo-300 z-10',
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
              <!-- 1. QR Code -->
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

              <!-- 2. Barcode 128 -->
              <div
                v-else-if="el.type === 'code128'"
                class="w-full h-full p-0.5 flex flex-col items-center justify-center pointer-events-none"
              >
                <img
                  :src="`/api/barcode/${sampleItem.id}.png?fmt=code128`"
                  class="w-full h-[75%] object-contain pointer-events-none"
                  alt="BC"
                />
                <span class="text-[8px] font-mono font-bold leading-tight truncate max-w-full">{{ sampleItem.sku }}</span>
              </div>

              <!-- 3. Header / Kop -->
              <div
                v-else-if="el.type === 'header'"
                class="w-full h-full flex items-center pointer-events-none uppercase tracking-wide leading-tight truncate px-0.5"
                :class="[
                  el.border ? 'border-b border-black' : '',
                  el.align === 'center' ? 'justify-center text-center' : el.align === 'right' ? 'justify-end text-right' : 'justify-start text-left',
                ]"
                :style="{
                  fontSize: `${(el.fontSize || 6.5) * (zoom / 2)}pt`,
                  fontWeight: el.bold ? 800 : 600,
                }"
              >
                {{ el.text || 'PEMERINTAH KABUPATEN KSB' }}
              </div>

              <!-- 4. Dynamic Field from Form -->
              <div
                v-else-if="el.type === 'field'"
                class="w-full h-full flex items-center pointer-events-none leading-tight px-0.5 overflow-hidden"
                :class="[
                  el.align === 'center' ? 'justify-center text-center' : el.align === 'right' ? 'justify-end text-right' : 'justify-start text-left',
                ]"
                :style="{
                  fontSize: `${(el.fontSize || 7) * (zoom / 2)}pt`,
                  fontWeight: el.bold ? 700 : 400,
                  fontFamily: el.field === 'sku' ? 'monospace' : 'inherit',
                }"
              >
                <span class="truncate">{{ resolveFieldValue(sampleItem, el.field, el.prefix) }}</span>
              </div>

              <!-- 5. Static Text -->
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

              <!-- 6. Line Divider -->
              <div
                v-else-if="el.type === 'line'"
                class="w-full h-full pointer-events-none border-t border-black my-auto"
              />

              <!-- 7. Logo -->
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
                  class="w-full h-full border border-dashed border-slate-300 bg-slate-50 flex flex-col items-center justify-center text-slate-400 p-1"
                >
                  <i class="pi pi-image text-sm" />
                </div>
              </div>

              <!-- Resize Handle -->
              <div
                v-if="activeElementId === el.id"
                class="absolute -right-1 -bottom-1 w-3.5 h-3.5 bg-indigo-600 border-2 border-white rounded-xs cursor-nwse-resize z-30 shadow-xs"
                @pointerdown.stop="startResize($event, el)"
              />
            </div>
          </div>
        </div>
      </div>

      <!-- RIGHT COLUMN: PROPERTIES INSPECTOR -->
      <div class="w-72 border-l border-slate-200 bg-white p-3.5 flex flex-col gap-3 text-xs overflow-y-auto select-none">
        <div class="font-bold text-slate-700 uppercase tracking-wider text-[11px] flex items-center justify-between pb-2 border-b border-slate-200">
          <span>Pengaturan Elemen</span>
          <Tag v-if="activeElement" :value="activeElement.type.toUpperCase()" severity="info" class="!text-[9px] !py-0" />
        </div>

        <div v-if="activeElement" class="flex flex-col gap-3">
          <!-- 1. Text Field Inputs -->
          <div v-if="activeElement.type === 'header' || activeElement.type === 'static_text'" class="flex flex-col gap-1">
            <label class="text-slate-600 font-semibold text-[11px]">Teks Tampilan</label>
            <InputText v-model="activeElement.text" class="w-full !text-xs !py-1" />
          </div>

          <!-- 2. Form Field Selector & Prefix -->
          <div v-if="activeElement.type === 'field'" class="flex flex-col gap-2 p-2 bg-slate-50 rounded-md border border-slate-200">
            <div class="flex flex-col gap-1">
              <label class="text-slate-600 font-semibold text-[11px]">Field dari Formulir</label>
              <Select
                v-model="activeElement.field"
                :options="AVAILABLE_FIELDS"
                optionLabel="label"
                optionValue="value"
                class="w-full !text-xs"
              />
            </div>

            <div class="flex flex-col gap-1">
              <label class="text-slate-600 font-semibold text-[11px]">Teks Awalan (Prefix)</label>
              <InputText
                v-model="activeElement.prefix"
                placeholder="Contoh: NUP: , Thn: , dll"
                class="w-full !text-xs !py-1"
              />
              <span class="text-[10px] text-slate-400">Contoh hasil: "{{ resolveFieldValue(sampleItem, activeElement.field, activeElement.prefix) }}"</span>
            </div>
          </div>

          <!-- 3. Logo Controls -->
          <div v-if="activeElement.type === 'logo'" class="flex flex-col gap-2 p-2.5 bg-slate-50 rounded-md border border-slate-200">
            <label class="text-[11px] text-slate-700 font-bold">File Gambar Logo</label>
            <div v-if="activeElement.logoUrl" class="flex items-center gap-2">
              <img :src="activeElement.logoUrl" class="w-12 h-12 object-contain bg-white rounded p-1 border border-slate-300" alt="Preview" />
              <div class="flex flex-col gap-1 flex-1">
                <span class="text-[10.5px] text-emerald-600 font-semibold">✓ Logo terpasang</span>
                <button
                  type="button"
                  class="text-[10px] text-rose-600 hover:text-rose-700 text-left font-semibold"
                  @click="activeElement.logoUrl = ''"
                >
                  Hapus Logo
                </button>
              </div>
            </div>

            <label class="cursor-pointer w-full flex items-center justify-center gap-2 py-1.5 px-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded text-xs font-semibold transition-colors">
              <i class="pi pi-upload text-xs" />
              <span>{{ activeElement.logoUrl ? 'Ganti Logo' : 'Unggah File Logo (PNG/JPG)' }}</span>
              <input type="file" accept="image/*" class="hidden" @change="handleLogoUpload" />
            </label>
          </div>

          <!-- 4. Dimension & Position (mm) -->
          <div class="p-2.5 bg-slate-50 rounded-md border border-slate-200 flex flex-col gap-2">
            <span class="text-[11px] font-bold text-slate-700">Posisi &amp; Ukuran (mm)</span>
            <div class="grid grid-cols-2 gap-2">
              <div class="flex flex-col gap-1">
                <label class="text-[10px] text-slate-500 font-mono">Posisi X</label>
                <InputNumber v-model="activeElement.x" :min="0" :step="0.5" suffix=" mm" class="w-full !text-xs" inputClass="!py-0.5 !text-xs font-mono" />
              </div>
              <div class="flex flex-col gap-1">
                <label class="text-[10px] text-slate-500 font-mono">Posisi Y</label>
                <InputNumber v-model="activeElement.y" :min="0" :step="0.5" suffix=" mm" class="w-full !text-xs" inputClass="!py-0.5 !text-xs font-mono" />
              </div>
              <div class="flex flex-col gap-1">
                <label class="text-[10px] text-slate-500 font-mono">Lebar (W)</label>
                <InputNumber v-model="activeElement.w" :min="2" :step="0.5" suffix=" mm" class="w-full !text-xs" inputClass="!py-0.5 !text-xs font-mono" />
              </div>
              <div class="flex flex-col gap-1">
                <label class="text-[10px] text-slate-500 font-mono">Tinggi (H)</label>
                <InputNumber v-model="activeElement.h" :min="1" :step="0.5" suffix=" mm" class="w-full !text-xs" inputClass="!py-0.5 !text-xs font-mono" />
              </div>
            </div>
          </div>

          <!-- 5. Typography -->
          <div v-if="!['qr', 'code128', 'line', 'logo'].includes(activeElement.type)" class="p-2.5 bg-slate-50 rounded-md border border-slate-200 flex flex-col gap-2">
            <div class="flex items-center justify-between">
              <label class="text-[11px] font-semibold text-slate-700">Ukuran Huruf</label>
              <InputNumber v-model="activeElement.fontSize" :min="4" :max="36" suffix=" pt" class="!w-24 !text-xs" inputClass="!py-0.5 !text-xs text-center font-bold" />
            </div>

            <div class="flex items-center justify-between pt-1.5 border-t border-slate-200">
              <label class="flex items-center gap-1.5 cursor-pointer text-slate-700 text-[11px]">
                <Checkbox v-model="activeElement.bold" binary />
                <span class="font-bold">Tebal (Bold)</span>
              </label>

              <label v-if="activeElement.type === 'header'" class="flex items-center gap-1.5 cursor-pointer text-slate-700 text-[11px]">
                <Checkbox v-model="activeElement.border" binary />
                <span>Garis Bawah</span>
              </label>
            </div>

            <div class="flex items-center justify-between pt-1.5 border-t border-slate-200">
              <span class="text-[11px] text-slate-600">Perataan:</span>
              <div class="flex items-center gap-1 bg-white p-0.5 rounded border border-slate-200">
                <button
                  type="button"
                  class="align-btn"
                  :class="activeElement.align === 'left' ? 'bg-indigo-600 text-white font-bold' : 'text-slate-600 hover:bg-slate-100'"
                  title="Rata Kiri"
                  @click="activeElement.align = 'left'"
                >
                  <i class="pi pi-align-left text-xs" />
                </button>
                <button
                  type="button"
                  class="align-btn"
                  :class="activeElement.align === 'center' ? 'bg-indigo-600 text-white font-bold' : 'text-slate-600 hover:bg-slate-100'"
                  title="Rata Tengah"
                  @click="activeElement.align = 'center'"
                >
                  <i class="pi pi-align-center text-xs" />
                </button>
                <button
                  type="button"
                  class="align-btn"
                  :class="activeElement.align === 'right' ? 'bg-indigo-600 text-white font-bold' : 'text-slate-600 hover:bg-slate-100'"
                  title="Rata Kanan"
                  @click="activeElement.align = 'right'"
                >
                  <i class="pi pi-align-right text-xs" />
                </button>
              </div>
            </div>
          </div>

          <!-- 6. Actions (Duplicate & Delete) -->
          <div class="flex items-center gap-2 pt-1">
            <Button
              icon="pi pi-copy"
              label="Duplikat"
              size="small"
              severity="secondary"
              outlined
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

        <!-- Empty State in Inspector -->
        <div v-else class="flex flex-col items-center justify-center p-8 text-center text-slate-400">
          <i class="pi pi-mouse text-3xl mb-2 text-slate-300" />
          <p class="text-xs">Klik salah satu elemen pada stiker untuk mengatur ukuran, teks, atau posisinya.</p>
        </div>
      </div>
    </div>
  </Dialog>

  <!-- PURE PRINT ENGINE (Direnter khusus saat window.print) -->
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
          <!-- QR Code -->
          <img
            v-if="el.type === 'qr'"
            :src="`/api/barcode/${it.id}.png?fmt=qr`"
            class="w-full h-full object-contain"
            alt="QR"
          />

          <!-- Barcode 128 -->
          <div v-else-if="el.type === 'code128'" class="w-full h-full flex flex-col items-center justify-center">
            <img :src="`/api/barcode/${it.id}.png?fmt=code128`" class="w-full h-[75%] object-contain" alt="BC" />
            <span class="text-[7pt] font-mono font-bold leading-none truncate">{{ it.sku }}</span>
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
              fontSize: `${el.fontSize || 6.5}pt`,
              fontWeight: el.bold ? 800 : 600,
            }"
          >
            {{ el.text }}
          </div>

          <!-- Dynamic Field from Form -->
          <div
            v-else-if="el.type === 'field'"
            class="w-full h-full flex items-center leading-tight truncate"
            :class="[
              el.align === 'center' ? 'justify-center text-center' : el.align === 'right' ? 'justify-end text-right' : 'justify-start text-left',
            ]"
            :style="{
              fontSize: `${el.fontSize || 7}pt`,
              fontWeight: el.bold ? 700 : 400,
              fontFamily: el.field === 'sku' ? 'monospace' : 'inherit',
            }"
          >
            {{ resolveFieldValue(it, el.field, el.prefix) }}
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

          <!-- Line Divider -->
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
.tab-btn {
  flex: 1;
  padding: 6px 10px;
  font-size: 11px;
  font-weight: 600;
  border-radius: 6px;
  color: #64748b;
  background: transparent;
  border: none;
  cursor: pointer;
  transition: all 0.15s ease;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}
.tab-btn-active {
  background: #ffffff;
  color: #4f46e5;
  box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
}

.field-item-btn {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 7px 9px;
  background-color: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s ease;
  text-align: left;
}
.field-item-btn:hover {
  background-color: #f8fafc;
  border-color: #6366f1;
  box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.04);
}

.component-item-btn {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  background-color: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s ease;
}
.component-item-btn:hover {
  background-color: #f8fafc;
  border-color: #6366f1;
  box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.04);
}

.align-btn {
  padding: 3px 6px;
  border-radius: 4px;
  border: none;
  cursor: pointer;
  transition: all 0.1s ease;
}
</style>

<style>
@media print {
  body * {
    visibility: hidden !important;
  }
  #custom-print-sheet,
  #custom-print-sheet * {
    visibility: visible !important;
  }
  #custom-print-sheet {
    position: absolute !important;
    left: 0 !important;
    top: 0 !important;
    margin: 0 !important;
    padding: 0 !important;
    background: transparent !important;
  }
  .custom-print-label {
    position: relative !important;
    box-sizing: border-box !important;
    overflow: hidden !important;
    background: #ffffff !important;
    page-break-after: always !important;
    break-after: page !important;
  }
  .custom-print-el {
    position: absolute !important;
    box-sizing: border-box !important;
    display: flex !important;
    overflow: hidden !important;
  }
}
</style>
