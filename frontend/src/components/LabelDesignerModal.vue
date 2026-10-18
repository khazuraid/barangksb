<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import Dialog from 'primevue/dialog'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Checkbox from 'primevue/checkbox'
import Tag from 'primevue/tag'
import { useToast } from 'primevue/usetoast'

export interface LabelElement {
  id: string
  type: 'qr' | 'code128' | 'field' | 'static_text' | 'header' | 'line' | 'logo'
  field?: string
  prefix?: string
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

export interface TableRowConfig {
  id: string
  type: 'field' | 'two_fields' | 'custom' | 'barcode_strip'
  field: string
  label: string
  field2?: string
  label2?: string
  customText?: string
  bold?: boolean
  bold2?: boolean
  fontSize?: number // pt (opsional)
  align?: 'left' | 'center' | 'right'
  bgColor?: 'white' | 'light_gray' | 'dark'
  hasBottomBorder?: boolean
}

export interface TableLayoutConfig {
  // Border Global Controls
  borderWidth: number // 0.5, 1, 1.5, 2 pt
  borderStyle: 'solid' | 'dashed' | 'dotted' | 'double'
  borderColor: string // '#000000', '#333333', '#666666'
  showOuterBorder: boolean
  showGridLines: boolean
  showColBorder: boolean
  showLabelColBorder: boolean
  showHeaderBorder: boolean
  showFooterBorder: boolean

  // Density & Dimensions
  cellPaddingMm: number // 0.4 - 2.0 mm
  labelColWidthPercent: number // 20 - 55%

  // Kop / Header
  showHeader: boolean
  headerText: string
  headerSubText: string
  headerFontSize: number
  headerStyle: 'standard' | 'double_line' | 'inverted'
  showLogo: boolean
  logoUrl: string
  logoWidthMm: number

  // Barcode / QR
  qrPosition: 'left' | 'right' | 'none'
  qrType: 'qr' | 'code128'
  qrWidthMm: number

  // Footer / Catatan Kaki
  showFooter: boolean
  footerText: string
  footerFontSize: number

  defaultFontSize: number
  rows: TableRowConfig[]
}

export interface LabelTemplate {
  id: string
  name: string
  widthMm: number
  heightMm: number
  layoutMode: 'table' | 'free'
  tableConfig: TableLayoutConfig
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
const zoom = ref(2.2)
const snapGrid = ref(true)
const gridSizeMm = ref(0.5)

// Field Lengkap dari Formulir Barang
const AVAILABLE_FIELDS = [
  { label: 'Nama Barang', value: 'name', icon: 'pi pi-box', defaultLabel: 'Nama' },
  { label: 'Kode Barang (KIB / BMD)', value: 'item_code', icon: 'pi pi-tag', defaultLabel: 'Kode KIB' },
  { label: 'Kode Barcode / SKU', value: 'sku', icon: 'pi pi-hashtag', defaultLabel: 'SKU' },
  { label: 'No. Register / NUP', value: 'register_number', icon: 'pi pi-id-card', defaultLabel: 'No. Reg' },
  { label: 'Ruangan / Lokasi', value: 'location', icon: 'pi pi-map-marker', defaultLabel: 'Ruang' },
  { label: 'Kode Ruangan / Lokasi', value: 'location_code', icon: 'pi pi-compass', defaultLabel: 'Kd. Ruang' },
  { label: 'Merk / Brand', value: 'merk', icon: 'pi pi-bookmark', defaultLabel: 'Merk' },
  { label: 'Tipe / Model', value: 'type_model', icon: 'pi pi-cog', defaultLabel: 'Tipe' },
  { label: 'Nomor Seri Pabrik (SN)', value: 'serial_number', icon: 'pi pi-barcode', defaultLabel: 'SN' },
  { label: 'Tahun Pengadaan', value: 'procurement_year', icon: 'pi pi-calendar', defaultLabel: 'Tahun' },
  { label: 'Asal Usul / Sumber Dana', value: 'funding_source', icon: 'pi pi-dollar', defaultLabel: 'Sumber' },
  { label: 'Kategori Barang', value: 'category', icon: 'pi pi-folder', defaultLabel: 'Kategori' },
  { label: 'Kondisi Barang', value: 'condition_status', icon: 'pi pi-shield', defaultLabel: 'Kondisi' },
  { label: 'Ukuran / Dimensi / CC', value: 'size', icon: 'pi pi-arrows-alt', defaultLabel: 'Ukuran' },
  { label: 'Bahan / Material', value: 'material', icon: 'pi pi-database', defaultLabel: 'Bahan' },
  { label: 'Harga Satuan (Rp)', value: 'price_per_unit', icon: 'pi pi-money-bill', defaultLabel: 'Harga' },
  { label: 'Keterangan', value: 'description', icon: 'pi pi-info-circle', defaultLabel: 'Ket' },
]

// Preset Desain Standar KSB Full Custom
const DEFAULT_PRESETS: LabelTemplate[] = [
  {
    id: 'preset_tj_103_table',
    name: '🌟 T&J 103 (64 × 32 mm) - Tabel Bergaris Lengkap',
    widthMm: 64,
    heightMm: 32,
    layoutMode: 'table',
    tableConfig: {
      borderWidth: 1.5,
      borderStyle: 'solid',
      borderColor: '#000000',
      showOuterBorder: true,
      showGridLines: true,
      showColBorder: true,
      showLabelColBorder: true,
      showHeaderBorder: true,
      showFooterBorder: false,
      cellPaddingMm: 0.8,
      labelColWidthPercent: 28,
      showHeader: true,
      headerText: 'PEMERINTAH KABUPATEN SUMBAWA BARAT',
      headerSubText: 'DINAS KESEHATAN — PUSKESMAS TALIWANG',
      headerFontSize: 6.5,
      headerStyle: 'double_line',
      showLogo: false,
      logoUrl: '',
      logoWidthMm: 8,
      qrPosition: 'left',
      qrType: 'qr',
      qrWidthMm: 22,
      showFooter: false,
      footerText: '',
      footerFontSize: 5.5,
      defaultFontSize: 6.5,
      rows: [
        { id: 'r1', type: 'field', field: 'name', label: 'Nama', bold: true, fontSize: 7, bgColor: 'white' },
        { id: 'r2', type: 'field', field: 'sku', label: 'SKU', bold: true, fontSize: 6.5, bgColor: 'white' },
        { id: 'r3', type: 'field', field: 'register_number', label: 'No. Reg', bold: false, fontSize: 6, bgColor: 'white' },
        { id: 'r4', type: 'two_fields', field: 'location', label: 'Ruang', field2: 'procurement_year', label2: 'Tahun', bold: false, fontSize: 6, bgColor: 'white' },
      ],
    },
    elements: [],
  },
  {
    id: 'preset_50x30_table',
    name: '🏷️ Stiker Rol (50 × 30 mm) - 2 Kolom Ringkas',
    widthMm: 50,
    heightMm: 30,
    layoutMode: 'table',
    tableConfig: {
      borderWidth: 1.2,
      borderStyle: 'solid',
      borderColor: '#000000',
      showOuterBorder: true,
      showGridLines: true,
      showColBorder: true,
      showLabelColBorder: true,
      showHeaderBorder: true,
      showFooterBorder: false,
      cellPaddingMm: 0.6,
      labelColWidthPercent: 30,
      showHeader: true,
      headerText: 'INVENTARIS PUSKESMAS KSB',
      headerSubText: '',
      headerFontSize: 6.5,
      headerStyle: 'standard',
      showLogo: false,
      logoUrl: '',
      logoWidthMm: 6,
      qrPosition: 'left',
      qrType: 'qr',
      qrWidthMm: 20,
      showFooter: false,
      footerText: '',
      footerFontSize: 5.5,
      defaultFontSize: 6.5,
      rows: [
        { id: 'r1', type: 'field', field: 'name', label: 'Barang', bold: true, fontSize: 7, bgColor: 'white' },
        { id: 'r2', type: 'field', field: 'sku', label: 'SKU', bold: true, fontSize: 6.5, bgColor: 'white' },
        { id: 'r3', type: 'two_fields', field: 'register_number', label: 'NUP', field2: 'location', label2: 'Ruang', bold: false, fontSize: 6, bgColor: 'white' },
      ],
    },
    elements: [],
  },
  {
    id: 'preset_70x50_table',
    name: '🏷️ Kartu Aset BMD (70 × 50 mm) - Full Custom & Dobel Kop',
    widthMm: 70,
    heightMm: 50,
    layoutMode: 'table',
    tableConfig: {
      borderWidth: 1.5,
      borderStyle: 'solid',
      borderColor: '#000000',
      showOuterBorder: true,
      showGridLines: true,
      showColBorder: true,
      showLabelColBorder: true,
      showHeaderBorder: true,
      showFooterBorder: true,
      cellPaddingMm: 0.9,
      labelColWidthPercent: 26,
      showHeader: true,
      headerText: 'KARTU INVENTARIS BARANG MILIK DAERAH (BMD)',
      headerSubText: 'BADAN PENGELOLAAN KEUANGAN DAN ASET DAERAH KSB',
      headerFontSize: 7,
      headerStyle: 'double_line',
      showLogo: false,
      logoUrl: '',
      logoWidthMm: 8,
      qrPosition: 'left',
      qrType: 'qr',
      qrWidthMm: 25,
      showFooter: true,
      footerText: 'DILARANG DILEPAS / DIPINDAHKAN TANPA IZIN PENGELOLA BMD',
      footerFontSize: 5.5,
      defaultFontSize: 6.5,
      rows: [
        { id: 'r1', type: 'field', field: 'name', label: 'Nama Barang', bold: true, fontSize: 7.5, bgColor: 'white' },
        { id: 'r2', type: 'two_fields', field: 'item_code', label: 'Kode KIB', field2: 'register_number', label2: 'No. Reg', bold: false, fontSize: 6.5, bgColor: 'white' },
        { id: 'r3', type: 'field', field: 'sku', label: 'SKU / Barcode', bold: true, fontSize: 6.5, bgColor: 'white' },
        { id: 'r4', type: 'two_fields', field: 'merk', label: 'Merk/Tipe', field2: 'procurement_year', label2: 'Tahun', bold: false, fontSize: 6.5, bgColor: 'white' },
        { id: 'r5', type: 'field', field: 'location', label: 'Ruangan', bold: false, fontSize: 6.5, bgColor: 'white' },
        { id: 'r6', type: 'field', field: 'funding_source', label: 'Sumber Dana', bold: false, fontSize: 6, bgColor: 'white' },
      ],
    },
    elements: [],
  },
]

const templates = ref<LabelTemplate[]>([])
const activeTemplate = ref<LabelTemplate>(JSON.parse(JSON.stringify(DEFAULT_PRESETS[0])))
const activeElementId = ref<string | null>(null)
const selectedTableRowIndex = ref<number | null>(null)
const copies = ref(1)
const isPrinting = ref(false)
const paletteTab = ref<'data' | 'borders' | 'header_footer' | 'free_components'>('data')

function ensureTableConfig(tpl: LabelTemplate) {
  if (!tpl.tableConfig) {
    tpl.tableConfig = {
      borderWidth: 1.5,
      borderStyle: 'solid',
      borderColor: '#000000',
      showOuterBorder: true,
      showGridLines: true,
      showColBorder: true,
      showLabelColBorder: true,
      showHeaderBorder: true,
      showFooterBorder: false,
      cellPaddingMm: 0.8,
      labelColWidthPercent: 28,
      showHeader: true,
      headerText: 'PEMERINTAH KABUPATEN SUMBAWA BARAT',
      headerSubText: '',
      headerFontSize: 6.5,
      headerStyle: 'standard',
      showLogo: false,
      logoUrl: '',
      logoWidthMm: 8,
      qrPosition: 'left',
      qrType: 'qr',
      qrWidthMm: 22,
      showFooter: false,
      footerText: '',
      footerFontSize: 5.5,
      defaultFontSize: 6.5,
      rows: [
        { id: 'r1', type: 'field', field: 'name', label: 'Nama', bold: true, fontSize: 7, bgColor: 'white' },
        { id: 'r2', type: 'field', field: 'sku', label: 'SKU', bold: true, fontSize: 6.5, bgColor: 'white' },
        { id: 'r3', type: 'field', field: 'location', label: 'Ruang', bold: false, fontSize: 6, bgColor: 'white' },
      ],
    }
  } else {
    // Migrasi field yang belum ada
    if (tpl.tableConfig.borderWidth === undefined) tpl.tableConfig.borderWidth = 1.5
    if (!tpl.tableConfig.borderStyle) tpl.tableConfig.borderStyle = 'solid'
    if (!tpl.tableConfig.borderColor) tpl.tableConfig.borderColor = '#000000'
    if (tpl.tableConfig.showLabelColBorder === undefined) tpl.tableConfig.showLabelColBorder = true
    if (tpl.tableConfig.cellPaddingMm === undefined) tpl.tableConfig.cellPaddingMm = 0.8
    if (tpl.tableConfig.labelColWidthPercent === undefined) tpl.tableConfig.labelColWidthPercent = 28
    if (tpl.tableConfig.headerSubText === undefined) tpl.tableConfig.headerSubText = ''
    if (!tpl.tableConfig.headerStyle) tpl.tableConfig.headerStyle = 'standard'
    if (tpl.tableConfig.showLogo === undefined) tpl.tableConfig.showLogo = false
    if (tpl.tableConfig.logoUrl === undefined) tpl.tableConfig.logoUrl = ''
    if (tpl.tableConfig.logoWidthMm === undefined) tpl.tableConfig.logoWidthMm = 8

    tpl.tableConfig.rows?.forEach((r) => {
      if (!r.type) r.type = 'field'
      if (!r.bgColor) r.bgColor = 'white'
    })
  }

  if (!tpl.layoutMode) {
    tpl.layoutMode = 'table'
  }
}

function loadSavedTemplates() {
  const raw = localStorage.getItem('custom_label_templates_v5')
  if (raw) {
    try {
      const parsed = JSON.parse(raw)
      parsed.forEach((t: LabelTemplate) => ensureTableConfig(t))
      templates.value = parsed
    } catch {
      templates.value = JSON.parse(JSON.stringify(DEFAULT_PRESETS))
    }
  } else {
    templates.value = JSON.parse(JSON.stringify(DEFAULT_PRESETS))
  }
}

function saveTemplates() {
  localStorage.setItem('custom_label_templates_v5', JSON.stringify(templates.value))
}

onMounted(() => {
  ensureTableConfig(activeTemplate.value)
  loadSavedTemplates()
})

const isTableMode = computed(() => activeTemplate.value.layoutMode === 'table')

const activeElement = computed(() => {
  if (!activeElementId.value) return null
  return activeTemplate.value.elements.find((el) => el.id === activeElementId.value) || null
})

const selectedTableRow = computed(() => {
  if (selectedTableRowIndex.value === null) return null
  return activeTemplate.value.tableConfig?.rows[selectedTableRowIndex.value] || null
})

// Sampel Barang Nyata
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

function resolveFieldValue(item: any, fieldKey?: string, prefix?: string): string {
  if (!fieldKey) return ''
  let val = item[fieldKey]
  if (val === undefined || val === null || val === '') {
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

// ==== OPERASI BARIS TABEL (FULL CUSTOM) ====
function addSingleFieldRow(fieldKey: string, defaultLabel: string) {
  ensureTableConfig(activeTemplate.value)
  const newRow: TableRowConfig = {
    id: `r_${Date.now().toString(36)}`,
    type: 'field',
    field: fieldKey,
    label: defaultLabel || fieldKey,
    bold: fieldKey === 'name' || fieldKey === 'sku',
    fontSize: fieldKey === 'name' ? 7 : activeTemplate.value.tableConfig.defaultFontSize,
    align: 'left',
    bgColor: 'white',
    hasBottomBorder: true,
  }
  activeTemplate.value.tableConfig.rows.push(newRow)
  selectedTableRowIndex.value = activeTemplate.value.tableConfig.rows.length - 1
  toast.add({ severity: 'success', summary: 'Baris Ditambahkan', detail: defaultLabel, life: 1200 })
}

function addTwoFieldsRow(f1: string, l1: string, f2: string, l2: string) {
  ensureTableConfig(activeTemplate.value)
  const newRow: TableRowConfig = {
    id: `r_${Date.now().toString(36)}`,
    type: 'two_fields',
    field: f1,
    label: l1,
    field2: f2,
    label2: l2,
    bold: false,
    bold2: false,
    fontSize: activeTemplate.value.tableConfig.defaultFontSize,
    align: 'left',
    bgColor: 'white',
    hasBottomBorder: true,
  }
  activeTemplate.value.tableConfig.rows.push(newRow)
  selectedTableRowIndex.value = activeTemplate.value.tableConfig.rows.length - 1
  toast.add({ severity: 'success', summary: 'Baris 2 Kolom Ditambahkan', detail: `${l1} & ${l2}`, life: 1400 })
}

function addCustomTextRow() {
  ensureTableConfig(activeTemplate.value)
  const newRow: TableRowConfig = {
    id: `r_${Date.now().toString(36)}`,
    type: 'custom',
    field: '',
    label: 'Catatan',
    customText: 'BARANG ASLI MILIK PEMERINTAH DAERAH',
    bold: true,
    fontSize: 6,
    align: 'center',
    bgColor: 'light_gray',
    hasBottomBorder: true,
  }
  activeTemplate.value.tableConfig.rows.push(newRow)
  selectedTableRowIndex.value = activeTemplate.value.tableConfig.rows.length - 1
  toast.add({ severity: 'info', summary: 'Baris Teks Bebas Ditambahkan', life: 1200 })
}

function addBarcodeStripRow() {
  ensureTableConfig(activeTemplate.value)
  const newRow: TableRowConfig = {
    id: `r_${Date.now().toString(36)}`,
    type: 'barcode_strip',
    field: 'sku',
    label: 'Barcode',
    bold: true,
    fontSize: 5.5,
    align: 'center',
    bgColor: 'white',
    hasBottomBorder: true,
  }
  activeTemplate.value.tableConfig.rows.push(newRow)
  selectedTableRowIndex.value = activeTemplate.value.tableConfig.rows.length - 1
  toast.add({ severity: 'info', summary: 'Baris Barcode Garis Ditambahkan', life: 1200 })
}

function duplicateTableRow(index: number) {
  if (!activeTemplate.value.tableConfig) return
  const original = activeTemplate.value.tableConfig.rows[index]
  const copy: TableRowConfig = JSON.parse(JSON.stringify(original))
  copy.id = `r_${Date.now().toString(36)}`
  activeTemplate.value.tableConfig.rows.splice(index + 1, 0, copy)
  selectedTableRowIndex.value = index + 1
  toast.add({ severity: 'info', summary: 'Baris Diduplikasi', life: 1000 })
}

function removeTableRow(index: number) {
  if (!activeTemplate.value.tableConfig) return
  activeTemplate.value.tableConfig.rows.splice(index, 1)
  if (selectedTableRowIndex.value === index) {
    selectedTableRowIndex.value = null
  } else if (selectedTableRowIndex.value !== null && selectedTableRowIndex.value > index) {
    selectedTableRowIndex.value--
  }
}

function moveRowUp(index: number) {
  if (index <= 0 || !activeTemplate.value.tableConfig) return
  const rows = activeTemplate.value.tableConfig.rows
  const temp = rows[index]
  rows[index] = rows[index - 1]
  rows[index - 1] = temp
  selectedTableRowIndex.value = index - 1
}

function moveRowDown(index: number) {
  if (!activeTemplate.value.tableConfig) return
  const rows = activeTemplate.value.tableConfig.rows
  if (index >= rows.length - 1) return
  const temp = rows[index]
  rows[index] = rows[index + 1]
  rows[index + 1] = temp
  selectedTableRowIndex.value = index + 1
}

// ==== DRAG & DROP ENGINE (MODE BEBAS) ====
let isDragging = false
let isResizing = false
let dragStartX = 0
let dragStartY = 0
let initialElemX = 0
let initialElemY = 0
let initialElemW = 0
let initialElemH = 0

function startDrag(e: PointerEvent, el: LabelElement) {
  if (isTableMode.value) return
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
  if (isTableMode.value || !activeElement.value) return
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

// Tambah Elemen Bebas
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

function addOtherElement(type: 'qr' | 'code128' | 'header' | 'static_text' | 'line') {
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
  }

  activeTemplate.value.elements.push(newEl)
  activeElementId.value = id
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
  const clone = JSON.parse(JSON.stringify(tpl))
  ensureTableConfig(clone)
  activeTemplate.value = clone
  activeElementId.value = null
  selectedTableRowIndex.value = null
  toast.add({ severity: 'info', summary: 'Preset Dimuat', detail: tpl.name, life: 1800 })
}

function resetLayout() {
  const defaultTpl = DEFAULT_PRESETS.find((p) => p.id === activeTemplate.value.id) || DEFAULT_PRESETS[0]
  const clone = JSON.parse(JSON.stringify(defaultTpl))
  ensureTableConfig(clone)
  activeTemplate.value = clone
  activeElementId.value = null
  selectedTableRowIndex.value = null
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
      if (parsed.widthMm && parsed.heightMm) {
        parsed.id = `imported_${Date.now().toString(36)}`
        ensureTableConfig(parsed)
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
    header="Desainer Label QR &amp; Barcode Aset (Kustom Penuh)"
    :style="{ width: '98vw', maxWidth: '1440px', height: '95vh' }"
    :contentStyle="{ padding: '0', display: 'flex', flexDirection: 'column', overflow: 'hidden' }"
    @update:visible="emit('update:visible', $event)"
  >
    <!-- TOP TOOLBAR: Layout Switcher, Presets, Dimensions & Print Action -->
    <div class="px-3.5 py-2 border-b border-slate-200 bg-white flex flex-wrap items-center justify-between gap-2.5 text-xs select-none shadow-xs">
      <!-- Left: Mode Switcher + Preset & Size -->
      <div class="flex items-center gap-2 flex-wrap">
        <!-- Layout Mode Toggle (Tabel vs Bebas) -->
        <div class="flex items-center bg-slate-100 p-0.5 rounded-lg border border-slate-200 shrink-0">
          <button
            type="button"
            class="px-2.5 py-1 rounded-md font-bold text-xs flex items-center gap-1.5 transition-all"
            :class="isTableMode ? 'bg-indigo-600 text-white shadow-xs' : 'text-slate-600 hover:text-slate-900'"
            @click="activeTemplate.layoutMode = 'table'"
          >
            <i class="pi pi-table" />
            <span>Mode Tabel (Kustom Full)</span>
          </button>
          <button
            type="button"
            class="px-2.5 py-1 rounded-md font-bold text-xs flex items-center gap-1.5 transition-all"
            :class="!isTableMode ? 'bg-indigo-600 text-white shadow-xs' : 'text-slate-600 hover:text-slate-900'"
            @click="activeTemplate.layoutMode = 'free'"
          >
            <i class="pi pi-th-large" />
            <span>Mode Bebas (Drag)</span>
          </button>
        </div>

        <div class="h-4 w-px bg-slate-200 mx-0.5 shrink-0" />

        <!-- Preset Selector -->
        <div class="flex items-center gap-1.5 shrink-0">
          <span class="text-slate-500 font-semibold">Preset:</span>
          <Select
            v-model="activeTemplate.id"
            :options="DEFAULT_PRESETS"
            optionLabel="name"
            optionValue="id"
            class="w-64 !text-xs"
            @change="(e) => {
              const p = DEFAULT_PRESETS.find((x) => x.id === e.value)
              if (p) applyPreset(p)
            }"
          />
        </div>

        <!-- Dimensi Stiker (mm) -->
        <div class="flex items-center gap-1 bg-slate-50 px-2 py-0.5 rounded-md border border-slate-200 shrink-0">
          <div class="flex items-center gap-1">
            <span class="text-slate-400 font-mono text-[11px]">P:</span>
            <input
              type="number"
              v-model.number="activeTemplate.widthMm"
              min="20"
              max="297"
              class="w-13 text-center font-mono font-bold text-xs bg-white text-slate-800 rounded border border-slate-300 py-0.5 px-1 focus:ring-1 focus:ring-indigo-500 focus:outline-hidden"
              title="Lebar stiker (mm)"
            />
          </div>
          <span class="text-slate-300 font-bold">&times;</span>
          <div class="flex items-center gap-1">
            <span class="text-slate-400 font-mono text-[11px]">L:</span>
            <input
              type="number"
              v-model.number="activeTemplate.heightMm"
              min="15"
              max="297"
              class="w-13 text-center font-mono font-bold text-xs bg-white text-slate-800 rounded border border-slate-300 py-0.5 px-1 focus:ring-1 focus:ring-indigo-500 focus:outline-hidden"
              title="Tinggi stiker (mm)"
            />
            <span class="text-slate-500 text-[11px] font-medium pl-0.5">mm</span>
          </div>
        </div>

        <Button
          icon="pi pi-refresh"
          label="Reset"
          size="small"
          text
          severity="secondary"
          v-tooltip.top="'Kembalikan tata letak ke preset awal'"
          class="!text-xs !py-1 shrink-0"
          @click="resetLayout"
        />
      </div>

      <!-- Right: View Controls & Print Action -->
      <div class="flex items-center gap-2 flex-wrap">
        <!-- Zoom Controller -->
        <div class="flex items-center bg-slate-100 rounded-md border border-slate-200 px-1 py-0.5 shrink-0">
          <button
            type="button"
            class="px-2 py-0.5 text-slate-600 hover:text-slate-900 font-bold"
            title="Perkecil Kanvas"
            @click="zoom = Math.max(1.2, zoom - 0.3)"
          >
            -
          </button>
          <span class="text-xs font-mono font-semibold text-slate-700 w-11 text-center">
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

        <!-- Snap Grid Toggle (Free Mode) -->
        <label v-if="!isTableMode" class="flex items-center gap-1.5 cursor-pointer text-slate-600 text-xs bg-slate-100 hover:bg-slate-200 px-2 py-1 rounded-md border border-slate-200 shrink-0">
          <Checkbox v-model="snapGrid" binary />
          <span>Snap 0.5mm</span>
        </label>

        <!-- Save Template -->
        <Button
          icon="pi pi-save"
          label="Simpan Template"
          size="small"
          severity="secondary"
          class="!text-xs !py-1 shrink-0"
          @click="saveCurrentAsNewTemplate"
        />

        <!-- Export / Import -->
        <Button icon="pi pi-download" text rounded size="small" severity="secondary" v-tooltip.top="'Ekspor JSON Template'" class="shrink-0" @click="exportJSON" />
        <label class="cursor-pointer p-1.5 rounded-md hover:bg-slate-100 text-slate-600 shrink-0" v-tooltip.top="'Impor File JSON'">
          <i class="pi pi-upload text-sm" />
          <input type="file" accept=".json" class="hidden" @change="importJSON" />
        </label>

        <span class="h-4 w-px bg-slate-200 mx-0.5 shrink-0" />

        <!-- Print Action: Copies + Cetak Button -->
        <div class="flex items-center gap-1.5 shrink-0">
          <div class="flex items-center gap-1 bg-slate-50 px-1.5 py-1 rounded-md border border-slate-200" title="Jumlah rangkap cetak per barang">
            <span class="text-slate-500 text-[11px] font-medium">&times;</span>
            <input
              type="number"
              v-model.number="copies"
              min="1"
              max="50"
              class="w-10 text-center font-bold text-xs bg-white text-slate-800 rounded border border-slate-300 py-0.5 focus:ring-1 focus:ring-indigo-500 focus:outline-hidden"
            />
          </div>
          <Button
            label="Cetak Label Ini"
            icon="pi pi-print"
            size="small"
            severity="success"
            class="!text-xs !py-1.5 !px-3 font-bold shadow-xs shrink-0"
            @click="printLabels"
          />
        </div>
      </div>
    </div>

    <!-- MAIN WORKSPACE: 3 COLUMNS -->
    <div class="flex-1 flex overflow-hidden bg-slate-100">
      <!-- LEFT COLUMN: PALETTE & TABLE ROW BUILDER -->
      <div class="w-80 border-r border-slate-200 bg-white flex flex-col overflow-hidden select-none">
        <!-- Palette Tab Headers -->
        <div class="p-2 border-b border-slate-200 bg-slate-50 flex items-center gap-1 overflow-x-auto">
          <button
            type="button"
            class="tab-btn"
            :class="paletteTab === 'data' ? 'tab-btn-active' : ''"
            @click="paletteTab = 'data'"
          >
            <i class="pi pi-list text-xs mr-1" />
            <span>{{ isTableMode ? 'Data & Baris' : 'Data Form' }}</span>
          </button>
          <button
            v-if="isTableMode"
            type="button"
            class="tab-btn"
            :class="paletteTab === 'borders' ? 'tab-btn-active' : ''"
            @click="paletteTab = 'borders'"
          >
            <i class="pi pi-th-large text-xs mr-1" />
            <span>Garis &amp; Jarak</span>
          </button>
          <button
            v-if="isTableMode"
            type="button"
            class="tab-btn"
            :class="paletteTab === 'header_footer' ? 'tab-btn-active' : ''"
            @click="paletteTab = 'header_footer'"
          >
            <i class="pi pi-building text-xs mr-1" />
            <span>Kop &amp; Barcode</span>
          </button>
          <button
            v-if="!isTableMode"
            type="button"
            class="tab-btn"
            :class="paletteTab === 'free_components' ? 'tab-btn-active' : ''"
            @click="paletteTab = 'free_components'"
          >
            <i class="pi pi-plus text-xs mr-1" />
            <span>Komponen</span>
          </button>
        </div>

        <!-- 1. TAB DATA & BARIS (MODE TABEL) -->
        <div v-if="paletteTab === 'data' && isTableMode" class="flex-1 overflow-y-auto p-2.5 flex flex-col gap-2.5">
          <!-- Tipe Baris Khusus Tambahan -->
          <div class="flex flex-col gap-1.5 p-2 bg-indigo-50/50 rounded-lg border border-indigo-100">
            <span class="text-[10px] font-bold text-indigo-900 uppercase tracking-wide">Tambah Struktur Baris Khusus:</span>
            <div class="grid grid-cols-2 gap-1.5">
              <button
                type="button"
                class="px-2 py-1.5 bg-white border border-indigo-200 hover:border-indigo-500 rounded text-left flex items-center gap-1.5 text-[11px] font-semibold text-slate-700 shadow-2xs transition-all"
                @click="addTwoFieldsRow('location', 'Ruang', 'procurement_year', 'Tahun')"
              >
                <i class="pi pi-table text-indigo-600 text-xs" />
                <span>2 Kolom Sebaris</span>
              </button>
              <button
                type="button"
                class="px-2 py-1.5 bg-white border border-indigo-200 hover:border-indigo-500 rounded text-left flex items-center gap-1.5 text-[11px] font-semibold text-slate-700 shadow-2xs transition-all"
                @click="addCustomTextRow"
              >
                <i class="pi pi-pencil text-purple-600 text-xs" />
                <span>Teks Bebas / Ket</span>
              </button>
              <button
                type="button"
                class="col-span-2 px-2 py-1.5 bg-white border border-indigo-200 hover:border-indigo-500 rounded text-left flex items-center gap-1.5 text-[11px] font-semibold text-slate-700 shadow-2xs transition-all"
                @click="addBarcodeStripRow"
              >
                <i class="pi pi-barcode text-cyan-600 text-xs" />
                <span>Barcode 128 Strip (Garis Barcode di Tabel)</span>
              </button>
            </div>
          </div>

          <!-- Tambah Baris dari Field Formulir -->
          <div class="flex flex-col gap-1">
            <div class="text-[11px] text-slate-500 px-1 pb-0.5 font-medium flex items-center justify-between">
              <span>Klik field untuk tambah ke tabel:</span>
              <span class="text-[10px] text-indigo-600 font-bold">{{ AVAILABLE_FIELDS.length }} Field</span>
            </div>

            <button
              v-for="f in AVAILABLE_FIELDS"
              :key="f.value"
              type="button"
              class="field-item-btn"
              @click="addSingleFieldRow(f.value, f.defaultLabel)"
            >
              <div class="flex items-center gap-2 truncate">
                <i :class="[f.icon, 'text-slate-500 text-xs shrink-0']" />
                <div class="truncate text-[11.5px] font-semibold text-slate-700">{{ f.label }}</div>
              </div>
              <i class="pi pi-plus text-slate-400 text-xs shrink-0" />
            </button>
          </div>
        </div>

        <!-- 2. TAB GARIS & JARAK (MODE TABEL - KUSTOM GARIS FULL) -->
        <div v-else-if="paletteTab === 'borders' && isTableMode" class="flex-1 overflow-y-auto p-3 flex flex-col gap-3">
          <!-- A. Ketebalan & Bentuk Garis Global -->
          <div class="flex flex-col gap-2 p-2.5 bg-slate-50 rounded-lg border border-slate-200">
            <span class="text-[11px] font-bold text-slate-800 uppercase tracking-wide">Ketebalan &amp; Jenis Garis</span>

            <div class="grid grid-cols-2 gap-2">
              <div class="flex flex-col gap-1">
                <label class="text-[10.5px] text-slate-600 font-semibold">Tebal Garis:</label>
                <Select
                  v-model="activeTemplate.tableConfig.borderWidth"
                  :options="[
                    { label: '0.5 pt (Tipis)', value: 0.5 },
                    { label: '1.0 pt (Standar)', value: 1 },
                    { label: '1.5 pt (Tegas)', value: 1.5 },
                    { label: '2.0 pt (Tebal)', value: 2 },
                  ]"
                  optionLabel="label"
                  optionValue="value"
                  class="w-full !text-xs"
                />
              </div>

              <div class="flex flex-col gap-1">
                <label class="text-[10.5px] text-slate-600 font-semibold">Jenis Garis:</label>
                <Select
                  v-model="activeTemplate.tableConfig.borderStyle"
                  :options="[
                    { label: 'Lurus (Solid)', value: 'solid' },
                    { label: 'Putus-putus', value: 'dashed' },
                    { label: 'Titik (Dotted)', value: 'dotted' },
                    { label: 'Garis Dobel', value: 'double' },
                  ]"
                  optionLabel="label"
                  optionValue="value"
                  class="w-full !text-xs"
                />
              </div>
            </div>

            <!-- Warna Garis Border -->
            <div class="flex items-center justify-between pt-1 border-t border-slate-200">
              <span class="text-[11px] text-slate-600 font-semibold">Warna Garis:</span>
              <div class="flex items-center gap-1.5">
                <button
                  type="button"
                  class="w-5 h-5 rounded-full border border-slate-300"
                  :class="activeTemplate.tableConfig.borderColor === '#000000' ? 'ring-2 ring-indigo-500 ring-offset-1' : ''"
                  style="background-color: #000000"
                  @click="activeTemplate.tableConfig.borderColor = '#000000'"
                />
                <button
                  type="button"
                  class="w-5 h-5 rounded-full border border-slate-300"
                  :class="activeTemplate.tableConfig.borderColor === '#334155' ? 'ring-2 ring-indigo-500 ring-offset-1' : ''"
                  style="background-color: #334155"
                  @click="activeTemplate.tableConfig.borderColor = '#334155'"
                />
                <button
                  type="button"
                  class="w-5 h-5 rounded-full border border-slate-300"
                  :class="activeTemplate.tableConfig.borderColor === '#64748b' ? 'ring-2 ring-indigo-500 ring-offset-1' : ''"
                  style="background-color: #64748b"
                  @click="activeTemplate.tableConfig.borderColor = '#64748b'"
                />
              </div>
            </div>
          </div>

          <!-- B. Toggle Garis Pembatas (Sisi-Sisi) -->
          <div class="flex flex-col gap-1.5 p-2.5 bg-slate-50 rounded-lg border border-slate-200">
            <span class="text-[11px] font-bold text-slate-800 uppercase tracking-wide">Tampilkan Garis Pembatas</span>

            <label class="flex items-center gap-2 cursor-pointer py-1 text-slate-700 text-xs">
              <Checkbox v-model="activeTemplate.tableConfig.showOuterBorder" binary />
              <span class="font-semibold">Bingkai Luar Stiker (Outer Box)</span>
            </label>

            <label class="flex items-center gap-2 cursor-pointer py-1 text-slate-700 text-xs">
              <Checkbox v-model="activeTemplate.tableConfig.showGridLines" binary />
              <span class="font-semibold">Garis Horizontal Antar Baris</span>
            </label>

            <label class="flex items-center gap-2 cursor-pointer py-1 text-slate-700 text-xs">
              <Checkbox v-model="activeTemplate.tableConfig.showLabelColBorder" binary />
              <span class="font-semibold">Garis Vertikal Pemisah Label &amp; Nilai</span>
            </label>

            <label class="flex items-center gap-2 cursor-pointer py-1 text-slate-700 text-xs">
              <Checkbox v-model="activeTemplate.tableConfig.showColBorder" binary />
              <span class="font-semibold">Garis Vertikal Pemisah Kolom QR</span>
            </label>

            <label class="flex items-center gap-2 cursor-pointer py-1 text-slate-700 text-xs">
              <Checkbox v-model="activeTemplate.tableConfig.showHeaderBorder" binary />
              <span class="font-semibold">Garis Bawah Kop / Header</span>
            </label>

            <label class="flex items-center gap-2 cursor-pointer py-1 text-slate-700 text-xs">
              <Checkbox v-model="activeTemplate.tableConfig.showFooterBorder" binary />
              <span class="font-semibold">Garis Atas Footer Catatan</span>
            </label>
          </div>

          <!-- C. Jarak / Padding & Proporsi Kolom -->
          <div class="flex flex-col gap-2 p-2.5 bg-slate-50 rounded-lg border border-slate-200">
            <span class="text-[11px] font-bold text-slate-800 uppercase tracking-wide">Kerapatan &amp; Lebar Kolom</span>

            <div class="flex items-center justify-between">
              <div class="flex flex-col">
                <span class="text-[11px] font-semibold text-slate-700">Kerapatan Baris (Padding):</span>
                <span class="text-[9.5px] text-slate-400">Jarak teks ke garis stiker</span>
              </div>
              <div class="flex items-center gap-1 bg-white border border-slate-300 rounded px-1.5 py-0.5">
                <input
                  type="number"
                  v-model.number="activeTemplate.tableConfig.cellPaddingMm"
                  min="0.3"
                  max="3"
                  step="0.1"
                  class="w-11 text-center font-bold text-xs bg-transparent focus:outline-hidden"
                />
                <span class="text-[10px] text-slate-400">mm</span>
              </div>
            </div>

            <div class="flex items-center justify-between pt-1 border-t border-slate-200">
              <div class="flex flex-col">
                <span class="text-[11px] font-semibold text-slate-700">Lebar Kolom Label:</span>
                <span class="text-[9.5px] text-slate-400">Persentase lebar nama field</span>
              </div>
              <div class="flex items-center gap-1 bg-white border border-slate-300 rounded px-1.5 py-0.5">
                <input
                  type="number"
                  v-model.number="activeTemplate.tableConfig.labelColWidthPercent"
                  min="18"
                  max="60"
                  step="1"
                  class="w-10 text-center font-bold text-xs bg-transparent focus:outline-hidden"
                />
                <span class="text-[10px] text-slate-400">%</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 3. TAB KOP, BARCODE & FOOTER (MODE TABEL) -->
        <div v-else-if="paletteTab === 'header_footer' && isTableMode" class="flex-1 overflow-y-auto p-3 flex flex-col gap-3">
          <!-- A. Pengaturan Barcode / QR Kolom -->
          <div class="flex flex-col gap-2 p-2.5 bg-slate-50 rounded-lg border border-slate-200">
            <span class="text-[11px] font-bold text-slate-800 uppercase tracking-wide">Kolom Barcode &amp; QR</span>

            <div class="flex flex-col gap-1">
              <label class="text-[11px] font-semibold text-slate-600">Posisi Barcode</label>
              <Select
                v-model="activeTemplate.tableConfig.qrPosition"
                :options="[
                  { label: 'Kolom Kiri', value: 'left' },
                  { label: 'Kolom Kanan', value: 'right' },
                  { label: 'Tanpa Barcode', value: 'none' },
                ]"
                optionLabel="label"
                optionValue="value"
                class="w-full !text-xs"
              />
            </div>

            <div v-if="activeTemplate.tableConfig.qrPosition !== 'none'" class="flex flex-col gap-1.5 pt-1">
              <div class="flex flex-col gap-1">
                <label class="text-[11px] font-semibold text-slate-600">Tipe Barcode</label>
                <Select
                  v-model="activeTemplate.tableConfig.qrType"
                  :options="[
                    { label: 'QR Code (Rekomendasi)', value: 'qr' },
                    { label: 'Barcode 128 (Garis)', value: 'code128' },
                  ]"
                  optionLabel="label"
                  optionValue="value"
                  class="w-full !text-xs"
                />
              </div>

              <div class="flex items-center justify-between pt-1 border-t border-slate-200">
                <span class="text-[11px] text-slate-600 font-medium">Lebar Kolom QR:</span>
                <div class="flex items-center gap-1 bg-white border border-slate-300 rounded px-1.5 py-0.5">
                  <input
                    type="number"
                    v-model.number="activeTemplate.tableConfig.qrWidthMm"
                    min="15"
                    max="50"
                    class="w-11 text-center font-bold text-xs bg-transparent focus:outline-hidden"
                  />
                  <span class="text-[10px] text-slate-400">mm</span>
                </div>
              </div>
            </div>
          </div>

          <!-- B. Kop / Header Instansi -->
          <div class="flex flex-col gap-2 p-2.5 bg-slate-50 rounded-lg border border-slate-200">
            <div class="flex items-center justify-between">
              <label class="flex items-center gap-1.5 cursor-pointer text-slate-800 font-bold text-xs">
                <Checkbox v-model="activeTemplate.tableConfig.showHeader" binary />
                <span>Kop / Header Instansi</span>
              </label>
            </div>

            <div v-if="activeTemplate.tableConfig.showHeader" class="flex flex-col gap-2 pt-1 border-t border-slate-200">
              <div class="flex flex-col gap-1">
                <label class="text-[10px] text-slate-500 font-semibold">Teks Baris 1 (Utama):</label>
                <InputText v-model="activeTemplate.tableConfig.headerText" placeholder="PEMERINTAH KABUPATEN..." class="w-full !text-xs !py-1" />
              </div>

              <div class="flex flex-col gap-1">
                <label class="text-[10px] text-slate-500 font-semibold">Teks Baris 2 (Sub-Kop / Dinas):</label>
                <InputText v-model="activeTemplate.tableConfig.headerSubText" placeholder="DINAS KESEHATAN..." class="w-full !text-xs !py-1" />
              </div>

              <div class="grid grid-cols-2 gap-2">
                <div class="flex flex-col gap-1">
                  <label class="text-[10px] text-slate-500 font-semibold">Gaya Kop:</label>
                  <Select
                    v-model="activeTemplate.tableConfig.headerStyle"
                    :options="[
                      { label: 'Standar', value: 'standard' },
                      { label: 'Dobel Garis Dinas', value: 'double_line' },
                      { label: 'Blok Hitam Teks Putih', value: 'inverted' },
                    ]"
                    optionLabel="label"
                    optionValue="value"
                    class="w-full !text-xs"
                  />
                </div>

                <div class="flex flex-col gap-1">
                  <label class="text-[10px] text-slate-500 font-semibold">Ukuran Font Kop:</label>
                  <div class="flex items-center gap-1 bg-white border border-slate-300 rounded px-1.5 py-0.5">
                    <input
                      type="number"
                      v-model.number="activeTemplate.tableConfig.headerFontSize"
                      min="5"
                      max="14"
                      step="0.5"
                      class="w-full text-center font-bold text-xs bg-transparent focus:outline-hidden"
                    />
                    <span class="text-[10px] text-slate-400">pt</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- C. Catatan Kaki / Footer -->
          <div class="flex flex-col gap-2 p-2.5 bg-slate-50 rounded-lg border border-slate-200">
            <div class="flex items-center justify-between">
              <label class="flex items-center gap-1.5 cursor-pointer text-slate-800 font-bold text-xs">
                <Checkbox v-model="activeTemplate.tableConfig.showFooter" binary />
                <span>Catatan Kaki (Footer)</span>
              </label>
            </div>

            <div v-if="activeTemplate.tableConfig.showFooter" class="flex flex-col gap-1.5 pt-1 border-t border-slate-200">
              <InputText v-model="activeTemplate.tableConfig.footerText" placeholder="DILARANG DILEPAS / DIPINDAHKAN..." class="w-full !text-xs !py-1" />
              <div class="flex items-center justify-between">
                <span class="text-[11px] text-slate-500">Ukuran Huruf Footer:</span>
                <div class="flex items-center gap-1 bg-white border border-slate-300 rounded px-1.5 py-0.5">
                  <input
                    type="number"
                    v-model.number="activeTemplate.tableConfig.footerFontSize"
                    min="4"
                    max="10"
                    step="0.5"
                    class="w-11 text-center font-bold text-xs bg-transparent focus:outline-hidden"
                  />
                  <span class="text-[10px] text-slate-400">pt</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 4. TAB KOMPONEN BEBAS (MODE BEBAS) -->
        <div v-else class="flex-1 overflow-y-auto p-2.5 flex flex-col gap-2">
          <button type="button" class="component-item-btn" @click="addOtherElement('qr')">
            <i class="pi pi-qrcode text-indigo-600 text-lg" />
            <div class="text-left flex-1 min-w-0">
              <div class="font-bold text-slate-800 text-[11.5px]">QR Code</div>
              <div class="text-[10px] text-slate-500">Scan link aset interaktif</div>
            </div>
            <i class="pi pi-plus text-slate-400 text-xs" />
          </button>

          <button type="button" class="component-item-btn" @click="addOtherElement('code128')">
            <i class="pi pi-barcode text-cyan-600 text-lg" />
            <div class="text-left flex-1 min-w-0">
              <div class="font-bold text-slate-800 text-[11.5px]">Barcode 128</div>
              <div class="text-[10px] text-slate-500">Barcode garis SKU aset</div>
            </div>
            <i class="pi pi-plus text-slate-400 text-xs" />
          </button>

          <button type="button" class="component-item-btn" @click="addOtherElement('header')">
            <i class="pi pi-building text-amber-600 text-lg" />
            <div class="text-left flex-1 min-w-0">
              <div class="font-bold text-slate-800 text-[11.5px]">Header / Kop Instansi</div>
              <div class="text-[10px] text-slate-500">Judul kop instansi atas</div>
            </div>
            <i class="pi pi-plus text-slate-400 text-xs" />
          </button>

          <button type="button" class="component-item-btn" @click="addOtherElement('line')">
            <i class="pi pi-minus text-slate-500 text-lg" />
            <div class="text-left flex-1 min-w-0">
              <div class="font-bold text-slate-800 text-[11.5px]">Garis Pembatas</div>
              <div class="text-[10px] text-slate-500">Garis pemisah horizontal</div>
            </div>
            <i class="pi pi-plus text-slate-400 text-xs" />
          </button>

          <button type="button" class="component-item-btn" @click="addOtherElement('static_text')">
            <i class="pi pi-pencil text-purple-600 text-lg" />
            <div class="text-left flex-1 min-w-0">
              <div class="font-bold text-slate-800 text-[11.5px]">Teks Bebas Statis</div>
              <div class="text-[10px] text-slate-500">Teks peringatan / keterangan</div>
            </div>
            <i class="pi pi-plus text-slate-400 text-xs" />
          </button>
        </div>

        <div class="p-2.5 border-t border-slate-200 bg-slate-50 text-[11px] text-slate-500 leading-tight">
          <span v-if="isTableMode">💡 <b>Mode Tabel Kustom:</b> Semua garis, kerapatan sel, 2 kolom sebaris, warna latar, dan kop instansi dapat dikustomisasi penuh.</span>
          <span v-else>💡 Geser elemen langsung di stiker. Tarik titik biru di pojok untuk ubah ukuran.</span>
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
            <span class="text-indigo-600 font-bold">{{ isTableMode ? 'Tabel Kustom Penuh' : 'Mode Bebas' }}</span>
            <span class="text-slate-300">|</span>
            <span class="text-slate-500 font-normal">Sampel: {{ sampleItem.name }}</span>
          </div>

          <!-- THE PHYSICAL STICKER -->
          <div
            class="relative bg-white shadow-xl rounded-[4px] text-black select-none overflow-hidden transition-all flex flex-col"
            :style="{
              width: `${activeTemplate.widthMm * MM_TO_PX * zoom}px`,
              height: `${activeTemplate.heightMm * MM_TO_PX * zoom}px`,
            }"
            @click.self="activeElementId = null; selectedTableRowIndex = null"
          >
            <!-- 1. MODE TABEL BERGARIS KUSTOM PENUH -->
            <div
              v-if="isTableMode"
              class="w-full h-full flex flex-col bg-white text-black box-border overflow-hidden"
              :style="{
                border: activeTemplate.tableConfig.showOuterBorder
                  ? `${activeTemplate.tableConfig.borderWidth * (zoom / 2.2)}px ${activeTemplate.tableConfig.borderStyle} ${activeTemplate.tableConfig.borderColor}`
                  : 'none',
              }"
            >
              <!-- Kop / Header Table Row -->
              <div
                v-if="activeTemplate.tableConfig.showHeader"
                class="w-full px-2 py-0.5 text-center font-extrabold uppercase tracking-tight shrink-0 flex flex-col items-center justify-center"
                :class="activeTemplate.tableConfig.headerStyle === 'inverted' ? 'bg-black text-white' : 'bg-slate-50/70 text-slate-900'"
                :style="{
                  borderBottom: activeTemplate.tableConfig.showHeaderBorder
                    ? activeTemplate.tableConfig.headerStyle === 'double_line'
                      ? `${activeTemplate.tableConfig.borderWidth * 2.2 * (zoom / 2.2)}px double ${activeTemplate.tableConfig.borderColor}`
                      : `${activeTemplate.tableConfig.borderWidth * (zoom / 2.2)}px ${activeTemplate.tableConfig.borderStyle} ${activeTemplate.tableConfig.borderColor}`
                    : 'none',
                }"
              >
                <span :style="{ fontSize: `${activeTemplate.tableConfig.headerFontSize * (zoom / 2.2)}pt` }">
                  {{ activeTemplate.tableConfig.headerText }}
                </span>
                <span
                  v-if="activeTemplate.tableConfig.headerSubText"
                  class="font-semibold text-[85%] opacity-90 tracking-normal"
                  :style="{ fontSize: `${(activeTemplate.tableConfig.headerFontSize - 1) * (zoom / 2.2)}pt` }"
                >
                  {{ activeTemplate.tableConfig.headerSubText }}
                </span>
              </div>

              <!-- Main Table Body: QR Column + Data Rows -->
              <div class="flex-1 flex overflow-hidden">
                <!-- QR Code Column (Left) -->
                <div
                  v-if="activeTemplate.tableConfig.qrPosition === 'left'"
                  class="shrink-0 flex flex-col items-center justify-center p-1"
                  :style="{
                    width: `${activeTemplate.tableConfig.qrWidthMm * MM_TO_PX * zoom}px`,
                    borderRight: activeTemplate.tableConfig.showColBorder
                      ? `${activeTemplate.tableConfig.borderWidth * (zoom / 2.2)}px ${activeTemplate.tableConfig.borderStyle} ${activeTemplate.tableConfig.borderColor}`
                      : 'none',
                  }"
                >
                  <img
                    v-if="activeTemplate.tableConfig.qrType === 'qr'"
                    :src="`/api/barcode/${sampleItem.id}.png?fmt=qr`"
                    class="w-full h-full object-contain pointer-events-none"
                    alt="QR"
                  />
                  <div v-else class="w-full h-full flex flex-col items-center justify-center">
                    <img :src="`/api/barcode/${sampleItem.id}.png?fmt=code128`" class="w-full h-[70%] object-contain pointer-events-none" alt="BC" />
                    <span class="font-mono font-bold leading-none text-center" :style="{ fontSize: `${5.5 * (zoom / 2.2)}pt` }">{{ sampleItem.sku }}</span>
                  </div>
                </div>

                <!-- Table Data Rows Subtable -->
                <div class="flex-1 flex flex-col justify-between overflow-hidden">
                  <div
                    v-for="(row, rIdx) in activeTemplate.tableConfig.rows"
                    :key="row.id"
                    class="flex-1 flex items-center cursor-pointer transition-colors relative"
                    :class="[
                      row.bgColor === 'light_gray' ? 'bg-slate-100' : row.bgColor === 'dark' ? 'bg-slate-800 text-white' : 'bg-white',
                      selectedTableRowIndex === rIdx ? 'ring-1 ring-indigo-500 bg-indigo-50/60' : 'hover:bg-indigo-50/30',
                    ]"
                    :style="{
                      padding: `${activeTemplate.tableConfig.cellPaddingMm * MM_TO_PX * (zoom / 2.2)}px 4px`,
                      borderBottom: (activeTemplate.tableConfig.showGridLines && rIdx < activeTemplate.tableConfig.rows.length - 1) || row.hasBottomBorder
                        ? `${activeTemplate.tableConfig.borderWidth * (zoom / 2.2)}px ${activeTemplate.tableConfig.borderStyle} ${activeTemplate.tableConfig.borderColor}`
                        : 'none',
                      fontSize: `${(row.fontSize || activeTemplate.tableConfig.defaultFontSize) * (zoom / 2.2)}pt`,
                    }"
                    @click.stop="selectedTableRowIndex = rIdx"
                  >
                    <!-- TIPE 1: Baris Field Tunggal Normal -->
                    <template v-if="!row.type || row.type === 'field'">
                      <div
                        class="font-bold shrink-0 truncate px-1"
                        :style="{
                          width: `${activeTemplate.tableConfig.labelColWidthPercent}%`,
                          borderRight: activeTemplate.tableConfig.showLabelColBorder
                            ? `${activeTemplate.tableConfig.borderWidth * (zoom / 2.2)}px ${activeTemplate.tableConfig.borderStyle} ${activeTemplate.tableConfig.borderColor}`
                            : 'none',
                        }"
                      >
                        {{ row.label }}:
                      </div>
                      <div
                        class="flex-1 truncate px-1.5"
                        :class="[
                          row.bold ? 'font-bold' : 'font-normal',
                          row.field === 'sku' ? 'font-mono' : '',
                          row.align === 'center' ? 'text-center' : row.align === 'right' ? 'text-right' : 'text-left',
                        ]"
                      >
                        {{ resolveFieldValue(sampleItem, row.field) }}
                      </div>
                    </template>

                    <!-- TIPE 2: Baris 2 Kolom Berdampingan (Two Fields) -->
                    <template v-else-if="row.type === 'two_fields'">
                      <!-- Kolom 1 -->
                      <div class="flex-1 flex items-center truncate min-w-0">
                        <span class="font-bold shrink-0 px-1">{{ row.label }}:</span>
                        <span class="truncate px-1" :class="row.bold ? 'font-bold' : ''">{{ resolveFieldValue(sampleItem, row.field) }}</span>
                      </div>
                      <!-- Pemisah Antar Kolom -->
                      <div
                        class="h-full w-px self-stretch"
                        :style="{
                          backgroundColor: activeTemplate.tableConfig.borderColor,
                          width: `${activeTemplate.tableConfig.borderWidth * (zoom / 2.2)}px`,
                        }"
                      />
                      <!-- Kolom 2 -->
                      <div class="flex-1 flex items-center truncate min-w-0 pl-1">
                        <span class="font-bold shrink-0 px-1">{{ row.label2 || 'Kol 2' }}:</span>
                        <span class="truncate px-1" :class="row.bold2 ? 'font-bold' : ''">{{ resolveFieldValue(sampleItem, row.field2) }}</span>
                      </div>
                    </template>

                    <!-- TIPE 3: Baris Teks Bebas / Manual -->
                    <template v-else-if="row.type === 'custom'">
                      <div
                        class="w-full text-center px-2 font-bold tracking-tight truncate"
                        :class="row.align === 'left' ? 'text-left' : row.align === 'right' ? 'text-right' : 'text-center'"
                      >
                        {{ row.customText || row.label }}
                      </div>
                    </template>

                    <!-- TIPE 4: Barcode Strip Memanjang -->
                    <template v-else-if="row.type === 'barcode_strip'">
                      <div class="w-full flex flex-col items-center justify-center p-0.5">
                        <img :src="`/api/barcode/${sampleItem.id}.png?fmt=code128`" class="w-full h-4 object-contain pointer-events-none" alt="BC" />
                        <span class="font-mono font-bold text-[7pt] leading-none">{{ sampleItem.sku }}</span>
                      </div>
                    </template>
                  </div>

                  <div
                    v-if="activeTemplate.tableConfig.rows.length === 0"
                    class="flex-1 flex items-center justify-center text-slate-400 text-xs p-4 text-center"
                  >
                    Tabel kosong. Tambah baris dari panel di sebelah kiri.
                  </div>
                </div>

                <!-- QR Code Column (Right) -->
                <div
                  v-if="activeTemplate.tableConfig.qrPosition === 'right'"
                  class="shrink-0 flex flex-col items-center justify-center p-1"
                  :style="{
                    width: `${activeTemplate.tableConfig.qrWidthMm * MM_TO_PX * zoom}px`,
                    borderLeft: activeTemplate.tableConfig.showColBorder
                      ? `${activeTemplate.tableConfig.borderWidth * (zoom / 2.2)}px ${activeTemplate.tableConfig.borderStyle} ${activeTemplate.tableConfig.borderColor}`
                      : 'none',
                  }"
                >
                  <img
                    v-if="activeTemplate.tableConfig.qrType === 'qr'"
                    :src="`/api/barcode/${sampleItem.id}.png?fmt=qr`"
                    class="w-full h-full object-contain pointer-events-none"
                    alt="QR"
                  />
                  <div v-else class="w-full h-full flex flex-col items-center justify-center">
                    <img :src="`/api/barcode/${sampleItem.id}.png?fmt=code128`" class="w-full h-[70%] object-contain pointer-events-none" alt="BC" />
                    <span class="font-mono font-bold leading-none text-center" :style="{ fontSize: `${5.5 * (zoom / 2.2)}pt` }">{{ sampleItem.sku }}</span>
                  </div>
                </div>
              </div>

              <!-- Footer Note Row -->
              <div
                v-if="activeTemplate.tableConfig.showFooter"
                class="w-full px-2 py-0.5 text-center text-slate-700 tracking-tight shrink-0 flex items-center justify-center font-medium"
                :style="{
                  borderTop: activeTemplate.tableConfig.showFooterBorder
                    ? `${activeTemplate.tableConfig.borderWidth * (zoom / 2.2)}px ${activeTemplate.tableConfig.borderStyle} ${activeTemplate.tableConfig.borderColor}`
                    : 'none',
                  fontSize: `${activeTemplate.tableConfig.footerFontSize * (zoom / 2.2)}pt`,
                }"
              >
                {{ activeTemplate.tableConfig.footerText }}
              </div>
            </div>

            <!-- 2. MODE BEBAS (Free Drag & Drop Elements) -->
            <template v-else>
              <div
                v-for="el in activeTemplate.elements"
                :key="el.id"
                class="absolute cursor-move select-none transition-shadow"
                :class="[
                  activeElementId === el.id ? 'ring-2 ring-indigo-600 shadow-md z-20' : 'hover:ring-1 hover:ring-slate-400 z-10',
                ]"
                :style="{
                  left: `${el.x * MM_TO_PX * zoom}px`,
                  top: `${el.y * MM_TO_PX * zoom}px`,
                  width: `${el.w * MM_TO_PX * zoom}px`,
                  height: `${el.h * MM_TO_PX * zoom}px`,
                }"
                @pointerdown="startDrag($event, el)"
              >
                <!-- QR Code -->
                <div v-if="el.type === 'qr'" class="w-full h-full p-0.5 bg-white pointer-events-none flex items-center justify-center">
                  <img :src="`/api/barcode/${sampleItem.id}.png?fmt=qr`" class="w-full h-full object-contain" alt="QR" />
                </div>

                <!-- Barcode 128 -->
                <div v-else-if="el.type === 'code128'" class="w-full h-full p-0.5 bg-white pointer-events-none flex flex-col items-center justify-center">
                  <img :src="`/api/barcode/${sampleItem.id}.png?fmt=code128`" class="w-full h-[70%] object-contain" alt="BC" />
                  <span class="text-[7pt] font-mono font-bold leading-none truncate">{{ sampleItem.sku }}</span>
                </div>

                <!-- Header -->
                <div
                  v-else-if="el.type === 'header'"
                  class="w-full h-full px-1 flex items-center uppercase tracking-wide leading-tight truncate pointer-events-none"
                  :class="[
                    el.border ? 'border-b border-black' : '',
                    el.align === 'center' ? 'justify-center text-center' : el.align === 'right' ? 'justify-end text-right' : 'justify-start text-left',
                  ]"
                  :style="{
                    fontSize: `${(el.fontSize || 6.5) * (zoom / 2)}pt`,
                    fontWeight: el.bold ? 800 : 600,
                  }"
                >
                  <span class="truncate">{{ el.text }}</span>
                </div>

                <!-- Field -->
                <div
                  v-else-if="el.type === 'field'"
                  class="w-full h-full px-1 flex items-center leading-tight truncate pointer-events-none"
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

                <!-- Static Text -->
                <div
                  v-else-if="el.type === 'static_text'"
                  class="w-full h-full px-1 flex items-center leading-tight truncate pointer-events-none"
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

                <!-- Line Divider -->
                <div v-else-if="el.type === 'line'" class="w-full h-full pointer-events-none border-t border-black my-auto" />

                <!-- Resize Handle -->
                <div
                  v-if="activeElementId === el.id"
                  class="absolute -right-1 -bottom-1 w-3.5 h-3.5 bg-indigo-600 border-2 border-white rounded-xs cursor-nwse-resize z-30 shadow-xs"
                  @pointerdown.stop="startResize($event, el)"
                />
              </div>
            </template>
          </div>
        </div>
      </div>

      <!-- RIGHT COLUMN: PROPERTIES INSPECTOR (FULL KUSTOM) -->
      <div class="w-88 shrink-0 border-l border-slate-200 bg-white p-3.5 flex flex-col gap-3 text-xs overflow-y-auto select-none">
        <div class="font-bold text-slate-700 uppercase tracking-wider text-[11px] flex items-center justify-between pb-2 border-b border-slate-200">
          <span>{{ isTableMode ? 'Inspektor Baris & Sel' : 'Pengaturan Elemen' }}</span>
          <Tag :value="isTableMode ? 'KUSTOM PENUH' : 'MODE BEBAS'" severity="info" class="!text-[9px] !py-0" />
        </div>

        <!-- 1. INSPECTOR UNTUK MODE TABEL -->
        <div v-if="isTableMode" class="flex flex-col gap-3">
          <!-- A. Pengaturan Baris Terpilih -->
          <div v-if="selectedTableRow" class="p-3 bg-indigo-50/60 rounded-lg border border-indigo-200 flex flex-col gap-2.5">
            <div class="flex items-center justify-between">
              <span class="font-bold text-indigo-900 text-xs">Edit Baris #{{ (selectedTableRowIndex ?? 0) + 1 }}</span>
              <div class="flex items-center gap-1">
                <button
                  type="button"
                  class="p-1 rounded bg-white border border-indigo-200 text-indigo-700 hover:bg-indigo-100 disabled:opacity-40"
                  :disabled="selectedTableRowIndex === 0"
                  title="Pindah Baris ke Atas"
                  @click="moveRowUp(selectedTableRowIndex!)"
                >
                  <i class="pi pi-arrow-up text-xs" />
                </button>
                <button
                  type="button"
                  class="p-1 rounded bg-white border border-indigo-200 text-indigo-700 hover:bg-indigo-100 disabled:opacity-40"
                  :disabled="selectedTableRowIndex === activeTemplate.tableConfig.rows.length - 1"
                  title="Pindah Baris ke Bawah"
                  @click="moveRowDown(selectedTableRowIndex!)"
                >
                  <i class="pi pi-arrow-down text-xs" />
                </button>
                <button
                  type="button"
                  class="p-1 rounded bg-white border border-indigo-200 text-indigo-700 hover:bg-indigo-100 ml-0.5"
                  title="Duplikat Baris Ini"
                  @click="duplicateTableRow(selectedTableRowIndex!)"
                >
                  <i class="pi pi-copy text-xs" />
                </button>
                <button
                  type="button"
                  class="p-1 rounded bg-rose-50 border border-rose-200 text-rose-600 hover:bg-rose-100 ml-0.5"
                  title="Hapus Baris Ini"
                  @click="removeTableRow(selectedTableRowIndex!)"
                >
                  <i class="pi pi-trash text-xs" />
                </button>
              </div>
            </div>

            <!-- Tipe Baris Selector -->
            <div class="flex flex-col gap-1">
              <label class="text-[10.5px] font-semibold text-slate-600">Tipe Baris:</label>
              <Select
                v-model="selectedTableRow.type"
                :options="[
                  { label: 'Field Tunggal (Label: Nilai)', value: 'field' },
                  { label: '2 Kolom Berdampingan', value: 'two_fields' },
                  { label: 'Teks Bebas / Peringatan', value: 'custom' },
                  { label: 'Barcode Garis Strip', value: 'barcode_strip' },
                ]"
                optionLabel="label"
                optionValue="value"
                class="w-full !text-xs"
              />
            </div>

            <!-- Pengaturan Berdasarkan Tipe Baris -->
            <!-- 1. FIELD TUNGGAL -->
            <template v-if="!selectedTableRow.type || selectedTableRow.type === 'field'">
              <div class="flex flex-col gap-1">
                <label class="text-[10.5px] font-semibold text-slate-600">Label Judul Baris:</label>
                <InputText v-model="selectedTableRow.label" class="w-full !text-xs !py-1" />
              </div>
              <div class="flex flex-col gap-1">
                <label class="text-[10.5px] font-semibold text-slate-600">Field Data Barang:</label>
                <Select
                  v-model="selectedTableRow.field"
                  :options="AVAILABLE_FIELDS"
                  optionLabel="label"
                  optionValue="value"
                  class="w-full !text-xs"
                />
              </div>
            </template>

            <!-- 2. TWO FIELDS -->
            <template v-else-if="selectedTableRow.type === 'two_fields'">
              <div class="grid grid-cols-2 gap-2 p-1.5 bg-white rounded border border-indigo-100">
                <div class="flex flex-col gap-1">
                  <label class="text-[10px] font-bold text-slate-600">Kolom Kiri</label>
                  <InputText v-model="selectedTableRow.label" placeholder="Label 1" class="w-full !text-xs !py-0.5" />
                  <Select v-model="selectedTableRow.field" :options="AVAILABLE_FIELDS" optionLabel="label" optionValue="value" class="w-full !text-[11px]" />
                </div>
                <div class="flex flex-col gap-1">
                  <label class="text-[10px] font-bold text-slate-600">Kolom Kanan</label>
                  <InputText v-model="selectedTableRow.label2" placeholder="Label 2" class="w-full !text-xs !py-0.5" />
                  <Select v-model="selectedTableRow.field2" :options="AVAILABLE_FIELDS" optionLabel="label" optionValue="value" class="w-full !text-[11px]" />
                </div>
              </div>
            </template>

            <!-- 3. CUSTOM TEXT -->
            <template v-else-if="selectedTableRow.type === 'custom'">
              <div class="flex flex-col gap-1">
                <label class="text-[10.5px] font-semibold text-slate-600">Teks Bebas Manual:</label>
                <InputText v-model="selectedTableRow.customText" placeholder="Contoh: ASLI BMD PEMDA KSB" class="w-full !text-xs !py-1" />
              </div>
            </template>

            <!-- Styling Sel / Baris Ini -->
            <div class="pt-2 border-t border-indigo-100 flex flex-col gap-2">
              <span class="text-[10.5px] font-bold text-indigo-900 uppercase">Gaya &amp; Warna Sel</span>

              <div class="grid grid-cols-2 gap-2">
                <div class="flex flex-col gap-1">
                  <label class="text-[10px] text-slate-600">Ukuran Font:</label>
                  <div class="flex items-center gap-1 bg-white border border-slate-300 rounded px-1.5 py-0.5">
                    <input
                      type="number"
                      v-model.number="selectedTableRow.fontSize"
                      min="4"
                      max="14"
                      step="0.5"
                      class="w-full text-center font-bold text-xs bg-transparent focus:outline-hidden"
                    />
                    <span class="text-[10px] text-slate-400">pt</span>
                  </div>
                </div>

                <div class="flex flex-col gap-1">
                  <label class="text-[10px] text-slate-600">Warna Latar Sel:</label>
                  <Select
                    v-model="selectedTableRow.bgColor"
                    :options="[
                      { label: 'Putih', value: 'white' },
                      { label: 'Abu-abu', value: 'light_gray' },
                      { label: 'Hitam (Teks Putih)', value: 'dark' },
                    ]"
                    optionLabel="label"
                    optionValue="value"
                    class="w-full !text-xs"
                  />
                </div>
              </div>

              <div class="flex items-center justify-between pt-1">
                <label class="flex items-center gap-1.5 cursor-pointer text-slate-700 text-xs">
                  <Checkbox v-model="selectedTableRow.bold" binary />
                  <span class="font-bold">Teks Tebal (Bold)</span>
                </label>

                <div class="flex items-center gap-1">
                  <button
                    type="button"
                    class="px-2 py-0.5 rounded border text-[11px]"
                    :class="selectedTableRow.align === 'left' ? 'bg-indigo-600 text-white font-bold' : 'bg-white text-slate-600'"
                    @click="selectedTableRow.align = 'left'"
                  >
                    Kiri
                  </button>
                  <button
                    type="button"
                    class="px-2 py-0.5 rounded border text-[11px]"
                    :class="selectedTableRow.align === 'center' ? 'bg-indigo-600 text-white font-bold' : 'bg-white text-slate-600'"
                    @click="selectedTableRow.align = 'center'"
                  >
                    Tengah
                  </button>
                  <button
                    type="button"
                    class="px-2 py-0.5 rounded border text-[11px]"
                    :class="selectedTableRow.align === 'right' ? 'bg-indigo-600 text-white font-bold' : 'bg-white text-slate-600'"
                    @click="selectedTableRow.align = 'right'"
                  >
                    Kanan
                  </button>
                </div>
              </div>

              <label class="flex items-center gap-1.5 cursor-pointer text-slate-700 text-xs pt-1">
                <Checkbox v-model="selectedTableRow.hasBottomBorder" binary />
                <span>Paksa Garis Bawah di Baris Ini</span>
              </label>
            </div>
          </div>

          <div v-else class="p-2.5 bg-slate-50 rounded-lg border border-slate-200 text-slate-500 text-xs text-center">
            👉 Klik salah satu baris pada kanvas stiker untuk mengatur field, kolom ganda, perataan teks, atau warna latarnya.
          </div>

          <!-- B. Daftar Urutan Baris Aktif -->
          <div class="flex flex-col gap-1.5">
            <span class="text-[11px] font-bold text-slate-700 uppercase tracking-wide">Urutan Baris Tabel ({{ activeTemplate.tableConfig.rows.length }})</span>
            <div class="flex flex-col gap-1 max-h-56 overflow-y-auto">
              <div
                v-for="(r, idx) in activeTemplate.tableConfig.rows"
                :key="r.id"
                class="flex items-center justify-between px-2 py-1 rounded border text-xs cursor-pointer transition-colors"
                :class="selectedTableRowIndex === idx ? 'bg-indigo-100 border-indigo-400 font-bold text-indigo-900' : 'bg-white border-slate-200 text-slate-700 hover:bg-slate-50'"
                @click="selectedTableRowIndex = idx"
              >
                <div class="flex items-center gap-1.5 truncate">
                  <span class="text-slate-400 text-[10px]">{{ idx + 1 }}.</span>
                  <span v-if="r.type === 'two_fields'" class="truncate text-indigo-700">⊞ {{ r.label }} | {{ r.label2 }}</span>
                  <span v-else-if="r.type === 'custom'" class="truncate text-purple-700">✎ {{ r.customText || r.label }}</span>
                  <span v-else-if="r.type === 'barcode_strip'" class="truncate text-cyan-700">|||| Barcode Garis</span>
                  <span v-else class="truncate">{{ r.label }} ({{ r.field }})</span>
                </div>
                <div class="flex items-center gap-1 shrink-0">
                  <button type="button" class="text-slate-400 hover:text-indigo-600 p-0.5" @click.stop="duplicateTableRow(idx)">
                    <i class="pi pi-copy text-[10px]" />
                  </button>
                  <button type="button" class="text-rose-500 hover:text-rose-700 p-0.5" @click.stop="removeTableRow(idx)">
                    <i class="pi pi-times text-[10px]" />
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 2. INSPECTOR UNTUK MODE BEBAS (Drag & Drop) -->
        <div v-else-if="activeElement" class="flex flex-col gap-3">
          <div v-if="activeElement.type === 'header' || activeElement.type === 'static_text'" class="flex flex-col gap-1">
            <label class="text-slate-600 font-semibold text-[11px]">Teks Tampilan</label>
            <InputText v-model="activeElement.text" class="w-full !text-xs !py-1" />
          </div>

          <div v-if="activeElement.type === 'field'" class="flex flex-col gap-2 p-2.5 bg-slate-50 rounded-lg border border-slate-200">
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
              <InputText v-model="activeElement.prefix" placeholder="Contoh: NUP: " class="w-full !text-xs !py-1" />
            </div>
          </div>

          <!-- Position & Dimension -->
          <div class="p-2.5 bg-slate-50 rounded-lg border border-slate-200 flex flex-col gap-2">
            <span class="text-[11px] font-bold text-slate-700">Posisi &amp; Ukuran (mm)</span>
            <div class="grid grid-cols-2 gap-2">
              <div class="flex items-center bg-white rounded border border-slate-300 px-2 py-1">
                <span class="text-slate-400 text-[10px] mr-1">X:</span>
                <input type="number" v-model.number="activeElement.x" step="0.5" class="w-full text-xs font-bold bg-transparent focus:outline-hidden" />
              </div>
              <div class="flex items-center bg-white rounded border border-slate-300 px-2 py-1">
                <span class="text-slate-400 text-[10px] mr-1">Y:</span>
                <input type="number" v-model.number="activeElement.y" step="0.5" class="w-full text-xs font-bold bg-transparent focus:outline-hidden" />
              </div>
              <div class="flex items-center bg-white rounded border border-slate-300 px-2 py-1">
                <span class="text-slate-400 text-[10px] mr-1">P:</span>
                <input type="number" v-model.number="activeElement.w" step="0.5" class="w-full text-xs font-bold bg-transparent focus:outline-hidden" />
              </div>
              <div class="flex items-center bg-white rounded border border-slate-300 px-2 py-1">
                <span class="text-slate-400 text-[10px] mr-1">L:</span>
                <input type="number" v-model.number="activeElement.h" step="0.5" class="w-full text-xs font-bold bg-transparent focus:outline-hidden" />
              </div>
            </div>
          </div>

          <div class="flex items-center gap-2 pt-1">
            <Button icon="pi pi-copy" label="Duplikat" size="small" severity="secondary" outlined class="w-1/2 !text-xs !py-1" @click="duplicateActiveElement" />
            <Button icon="pi pi-trash" label="Hapus" size="small" severity="danger" class="w-1/2 !text-xs !py-1" @click="removeActiveElement" />
          </div>
        </div>

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
        <!-- 1. Print Mode: Native HTML Table with Real Print Borders & Full Customization -->
        <table
          v-if="activeTemplate.layoutMode === 'table'"
          class="custom-print-table"
          :style="{
            border: activeTemplate.tableConfig.showOuterBorder
              ? `${activeTemplate.tableConfig.borderWidth}pt ${activeTemplate.tableConfig.borderStyle} ${activeTemplate.tableConfig.borderColor}`
              : 'none',
          }"
        >
          <!-- Kop Row -->
          <tr v-if="activeTemplate.tableConfig.showHeader">
            <th
              colspan="2"
              class="print-header-cell"
              :class="activeTemplate.tableConfig.headerStyle === 'inverted' ? 'print-inverted-kop' : ''"
              :style="{
                fontSize: `${activeTemplate.tableConfig.headerFontSize}pt`,
                borderBottom: activeTemplate.tableConfig.showHeaderBorder
                  ? activeTemplate.tableConfig.headerStyle === 'double_line'
                    ? `${activeTemplate.tableConfig.borderWidth * 2}pt double ${activeTemplate.tableConfig.borderColor}`
                    : `${activeTemplate.tableConfig.borderWidth}pt ${activeTemplate.tableConfig.borderStyle} ${activeTemplate.tableConfig.borderColor}`
                  : 'none',
              }"
            >
              <div>{{ activeTemplate.tableConfig.headerText }}</div>
              <div v-if="activeTemplate.tableConfig.headerSubText" style="font-weight: 600; font-size: 85%;">
                {{ activeTemplate.tableConfig.headerSubText }}
              </div>
            </th>
          </tr>

          <!-- Body Row: QR Column + Subtable of Rows -->
          <tr>
            <!-- Left QR Cell -->
            <td
              v-if="activeTemplate.tableConfig.qrPosition === 'left'"
              class="print-qr-cell"
              :style="{
                width: `${activeTemplate.tableConfig.qrWidthMm}mm`,
                borderRight: activeTemplate.tableConfig.showColBorder
                  ? `${activeTemplate.tableConfig.borderWidth}pt ${activeTemplate.tableConfig.borderStyle} ${activeTemplate.tableConfig.borderColor}`
                  : 'none',
              }"
            >
              <img
                v-if="activeTemplate.tableConfig.qrType === 'qr'"
                :src="`/api/barcode/${it.id}.png?fmt=qr`"
                class="print-qr-img"
                alt="QR"
              />
              <div v-else class="print-barcode-box">
                <img :src="`/api/barcode/${it.id}.png?fmt=code128`" class="print-barcode-img" alt="BC" />
                <span class="print-barcode-sku">{{ it.sku }}</span>
              </div>
            </td>

            <!-- Data Rows Subtable Cell -->
            <td class="print-data-cell">
              <table class="print-subtable">
                <tr
                  v-for="(r, rIdx) in activeTemplate.tableConfig.rows"
                  :key="r.id"
                  class="print-data-row"
                  :style="{
                    fontSize: `${r.fontSize || activeTemplate.tableConfig.defaultFontSize}pt`,
                    borderBottom: (activeTemplate.tableConfig.showGridLines && rIdx < activeTemplate.tableConfig.rows.length - 1) || r.hasBottomBorder
                      ? `${activeTemplate.tableConfig.borderWidth}pt ${activeTemplate.tableConfig.borderStyle} ${activeTemplate.tableConfig.borderColor}`
                      : 'none',
                    backgroundColor: r.bgColor === 'light_gray' ? '#f1f5f9' : r.bgColor === 'dark' ? '#1e293b' : 'transparent',
                    color: r.bgColor === 'dark' ? '#ffffff' : '#000000',
                  }"
                >
                  <!-- 1. FIELD TUNGGAL -->
                  <template v-if="!r.type || r.type === 'field'">
                    <td
                      class="print-label-col"
                      :style="{
                        width: `${activeTemplate.tableConfig.labelColWidthPercent}%`,
                        padding: `${activeTemplate.tableConfig.cellPaddingMm}mm 1mm`,
                        borderRight: activeTemplate.tableConfig.showLabelColBorder
                          ? `${activeTemplate.tableConfig.borderWidth}pt ${activeTemplate.tableConfig.borderStyle} ${activeTemplate.tableConfig.borderColor}`
                          : 'none',
                      }"
                    >
                      {{ r.label }}:
                    </td>
                    <td
                      class="print-val-col"
                      :class="[
                        r.bold ? 'font-bold' : '',
                        r.field === 'sku' ? 'font-mono' : '',
                      ]"
                      :style="{
                        padding: `${activeTemplate.tableConfig.cellPaddingMm}mm 1mm`,
                        textAlign: r.align || 'left',
                      }"
                    >
                      {{ resolveFieldValue(it, r.field) }}
                    </td>
                  </template>

                  <!-- 2. TWO FIELDS -->
                  <template v-else-if="r.type === 'two_fields'">
                    <td
                      colspan="2"
                      :style="{ padding: `${activeTemplate.tableConfig.cellPaddingMm}mm 1mm` }"
                    >
                      <div style="display: flex; width: 100%; align-items: center;">
                        <div style="flex: 1; display: flex; align-items: center; overflow: hidden;">
                          <b style="padding-right: 1mm;">{{ r.label }}:</b>
                          <span :style="{ fontWeight: r.bold ? 'bold' : 'normal' }">{{ resolveFieldValue(it, r.field) }}</span>
                        </div>
                        <div
                          :style="{
                            width: `${activeTemplate.tableConfig.borderWidth}pt`,
                            height: '100%',
                            backgroundColor: activeTemplate.tableConfig.borderColor,
                            margin: '0 1mm',
                          }"
                        />
                        <div style="flex: 1; display: flex; align-items: center; overflow: hidden; padding-left: 1mm;">
                          <b style="padding-right: 1mm;">{{ r.label2 || 'Kol 2' }}:</b>
                          <span :style="{ fontWeight: r.bold2 ? 'bold' : 'normal' }">{{ resolveFieldValue(it, r.field2) }}</span>
                        </div>
                      </div>
                    </td>
                  </template>

                  <!-- 3. CUSTOM TEXT -->
                  <template v-else-if="r.type === 'custom'">
                    <td
                      colspan="2"
                      :style="{
                        padding: `${activeTemplate.tableConfig.cellPaddingMm}mm 1mm`,
                        textAlign: r.align || 'center',
                        fontWeight: r.bold ? 'bold' : 'normal',
                      }"
                    >
                      {{ r.customText || r.label }}
                    </td>
                  </template>

                  <!-- 4. BARCODE STRIP -->
                  <template v-else-if="r.type === 'barcode_strip'">
                    <td colspan="2" style="text-align: center; padding: 0.5mm 1mm;">
                      <img :src="`/api/barcode/${it.id}.png?fmt=code128`" style="width: 90%; height: 6mm; object-fit: contain;" alt="BC" />
                      <div style="font-family: monospace; font-weight: bold; font-size: 5.5pt; line-height: 1;">{{ it.sku }}</div>
                    </td>
                  </template>
                </tr>
              </table>
            </td>

            <!-- Right QR Cell -->
            <td
              v-if="activeTemplate.tableConfig.qrPosition === 'right'"
              class="print-qr-cell"
              :style="{
                width: `${activeTemplate.tableConfig.qrWidthMm}mm`,
                borderLeft: activeTemplate.tableConfig.showColBorder
                  ? `${activeTemplate.tableConfig.borderWidth}pt ${activeTemplate.tableConfig.borderStyle} ${activeTemplate.tableConfig.borderColor}`
                  : 'none',
              }"
            >
              <img
                v-if="activeTemplate.tableConfig.qrType === 'qr'"
                :src="`/api/barcode/${it.id}.png?fmt=qr`"
                class="print-qr-img"
                alt="QR"
              />
              <div v-else class="print-barcode-box">
                <img :src="`/api/barcode/${it.id}.png?fmt=code128`" class="print-barcode-img" alt="BC" />
                <span class="print-barcode-sku">{{ it.sku }}</span>
              </div>
            </td>
          </tr>

          <!-- Footer Row -->
          <tr v-if="activeTemplate.tableConfig.showFooter">
            <td
              colspan="2"
              class="print-footer-cell"
              :style="{
                fontSize: `${activeTemplate.tableConfig.footerFontSize}pt`,
                borderTop: activeTemplate.tableConfig.showFooterBorder
                  ? `${activeTemplate.tableConfig.borderWidth}pt ${activeTemplate.tableConfig.borderStyle} ${activeTemplate.tableConfig.borderColor}`
                  : 'none',
              }"
            >
              {{ activeTemplate.tableConfig.footerText }}
            </td>
          </tr>
        </table>

        <!-- 2. Print Mode: Free Form Absolute Positions -->
        <template v-else>
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
            <img
              v-if="el.type === 'qr'"
              :src="`/api/barcode/${it.id}.png?fmt=qr`"
              class="w-full h-full object-contain"
              alt="QR"
            />
            <div v-else-if="el.type === 'code128'" class="w-full h-full flex flex-col items-center justify-center">
              <img :src="`/api/barcode/${it.id}.png?fmt=code128`" class="w-full h-[75%] object-contain" alt="BC" />
              <span class="text-[7pt] font-mono font-bold leading-none truncate">{{ it.sku }}</span>
            </div>
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
            <div v-else-if="el.type === 'line'" class="w-full h-full border-t border-black my-auto" />
          </div>
        </template>
      </div>
    </div>
  </teleport>
</template>

<style scoped>
.tab-btn {
  padding: 6px 9px;
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
  white-space: nowrap;
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
  padding: 6px 9px;
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

  /* Table Mode Print Styling */
  .custom-print-table {
    width: 100% !important;
    height: 100% !important;
    border-collapse: collapse !important;
    table-layout: fixed !important;
    box-sizing: border-box !important;
    margin: 0 !important;
    padding: 0 !important;
  }
  .print-header-cell {
    padding: 0.8mm 1.5mm !important;
    text-align: center !important;
    font-weight: 800 !important;
    text-transform: uppercase !important;
    vertical-align: middle !important;
  }
  .print-inverted-kop {
    background-color: #000000 !important;
    color: #ffffff !important;
    -webkit-print-color-adjust: exact !important;
    print-color-adjust: exact !important;
  }
  .print-qr-cell {
    padding: 1mm !important;
    text-align: center !important;
    vertical-align: middle !important;
  }
  .print-qr-img {
    width: 100% !important;
    height: 100% !important;
    object-fit: contain !important;
  }
  .print-barcode-box {
    width: 100% !important;
    height: 100% !important;
    display: flex !important;
    flex-direction: column !important;
    align-items: center !important;
    justify-content: center !important;
  }
  .print-barcode-img {
    width: 100% !important;
    height: 70% !important;
    object-fit: contain !important;
  }
  .print-barcode-sku {
    font-family: monospace !important;
    font-size: 6pt !important;
    font-weight: bold !important;
  }
  .print-data-cell {
    padding: 0 !important;
    vertical-align: top !important;
  }
  .print-subtable {
    width: 100% !important;
    height: 100% !important;
    border-collapse: collapse !important;
    margin: 0 !important;
    padding: 0 !important;
  }
  .print-data-row {
    height: 1% !important;
    box-sizing: border-box !important;
    -webkit-print-color-adjust: exact !important;
    print-color-adjust: exact !important;
  }
  .print-label-col {
    font-weight: bold !important;
    white-space: nowrap !important;
    vertical-align: middle !important;
    box-sizing: border-box !important;
  }
  .print-val-col {
    vertical-align: middle !important;
    overflow: hidden !important;
    box-sizing: border-box !important;
  }
  .print-footer-cell {
    padding: 0.8mm 1mm !important;
    text-align: center !important;
    vertical-align: middle !important;
    box-sizing: border-box !important;
  }
}
</style>
