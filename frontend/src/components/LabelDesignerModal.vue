<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import Dialog from 'primevue/dialog'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Checkbox from 'primevue/checkbox'
import Tag from 'primevue/tag'
import { useToast } from 'primevue/usetoast'
import api from '@/api'

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

export interface TableCellConfig {
  id: string
  field: string
  label: string
  customText?: string
  bold?: boolean
  fontSize?: number // pt (opsional)
  align?: 'left' | 'center' | 'right'
  bgColor?: 'white' | 'light_gray' | 'dark'
}

export interface TableRowConfig {
  id: string
  type: 'cells' | 'field' | 'two_fields' | 'custom' | 'barcode_strip'
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
  cells?: TableCellConfig[]
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
  logoPosition?: 'above_qr' | 'header'
  showItemNameUnderLogo?: boolean

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
    id: 'preset_cirebon_stacked_qr',
    name: '🏛️ Kota Cirebon BMD (65 × 30 mm) - Lambang & QR Berdampingan Atas-Bawah',
    widthMm: 65,
    heightMm: 30,
    layoutMode: 'table',
    tableConfig: {
      borderWidth: 1.5,
      borderStyle: 'solid',
      borderColor: '#000000',
      showOuterBorder: true,
      showGridLines: true,
      showColBorder: true,
      showLabelColBorder: false,
      showHeaderBorder: true,
      showFooterBorder: false,
      cellPaddingMm: 0.6,
      labelColWidthPercent: 0,
      showHeader: false,
      headerText: '',
      headerSubText: '',
      headerFontSize: 6.5,
      headerStyle: 'standard',
      showLogo: true,
      logoUrl: '',
      logoWidthMm: 12,
      logoPosition: 'above_qr',
      showItemNameUnderLogo: true,
      qrPosition: 'left',
      qrType: 'qr',
      qrWidthMm: 20,
      showFooter: false,
      footerText: '',
      footerFontSize: 5.5,
      defaultFontSize: 8,
      rows: [
        {
          id: 'r1',
          type: 'cells',
          field: 'register_number',
          label: '',
          bold: true,
          fontSize: 8.5,
          bgColor: 'white',
          hasBottomBorder: true,
          cells: [{ id: 'c1', field: 'register_number', label: '', bold: true, fontSize: 8.5, align: 'center', bgColor: 'white' }],
        },
        {
          id: 'r2',
          type: 'cells',
          field: 'item_code',
          label: '',
          bold: true,
          fontSize: 8.5,
          bgColor: 'white',
          hasBottomBorder: true,
          cells: [{ id: 'c1', field: 'item_code', label: '', bold: true, fontSize: 8.5, align: 'center', bgColor: 'white' }],
        },
        {
          id: 'r3',
          type: 'custom',
          field: '',
          label: '',
          customText: 'PEMERINTAH KOTA CIREBON',
          bold: true,
          fontSize: 7.5,
          bgColor: 'white',
          hasBottomBorder: true,
          cells: [{ id: 'c1', field: '', label: '', customText: 'PEMERINTAH KOTA CIREBON', bold: true, fontSize: 7.5, align: 'center', bgColor: 'white' }],
        },
      ],
    },
    elements: [],
  },
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
        {
          id: 'r1',
          type: 'cells',
          field: 'name',
          label: 'Nama',
          bold: true,
          fontSize: 7,
          bgColor: 'white',
          cells: [{ id: 'c1', field: 'name', label: 'Nama', bold: true, fontSize: 7, bgColor: 'white' }],
        },
        {
          id: 'r2',
          type: 'cells',
          field: 'sku',
          label: 'SKU',
          bold: true,
          fontSize: 6.5,
          bgColor: 'white',
          cells: [{ id: 'c1', field: 'sku', label: 'SKU', bold: true, fontSize: 6.5, bgColor: 'white' }],
        },
        {
          id: 'r3',
          type: 'cells',
          field: 'register_number',
          label: 'No. Reg',
          bold: false,
          fontSize: 6,
          bgColor: 'white',
          cells: [{ id: 'c1', field: 'register_number', label: 'No. Reg', bold: false, fontSize: 6, bgColor: 'white' }],
        },
        {
          id: 'r4',
          type: 'cells',
          field: 'location',
          label: 'Ruang',
          field2: 'procurement_year',
          label2: 'Tahun',
          bold: false,
          fontSize: 6,
          bgColor: 'white',
          cells: [
            { id: 'c1', field: 'location', label: 'Ruang', bold: false, fontSize: 6, bgColor: 'white' },
            { id: 'c2', field: 'procurement_year', label: 'Tahun', bold: false, fontSize: 6, bgColor: 'white' },
          ],
        },
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
        {
          id: 'r1',
          type: 'cells',
          field: 'name',
          label: 'Barang',
          bold: true,
          fontSize: 7,
          bgColor: 'white',
          cells: [{ id: 'c1', field: 'name', label: 'Barang', bold: true, fontSize: 7, bgColor: 'white' }],
        },
        {
          id: 'r2',
          type: 'cells',
          field: 'sku',
          label: 'SKU',
          bold: true,
          fontSize: 6.5,
          bgColor: 'white',
          cells: [{ id: 'c1', field: 'sku', label: 'SKU', bold: true, fontSize: 6.5, bgColor: 'white' }],
        },
        {
          id: 'r3',
          type: 'cells',
          field: 'register_number',
          label: 'NUP',
          field2: 'location',
          label2: 'Ruang',
          bold: false,
          fontSize: 6,
          bgColor: 'white',
          cells: [
            { id: 'c1', field: 'register_number', label: 'NUP', bold: false, fontSize: 6, bgColor: 'white' },
            { id: 'c2', field: 'location', label: 'Ruang', bold: false, fontSize: 6, bgColor: 'white' },
          ],
        },
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
        {
          id: 'r1',
          type: 'cells',
          field: 'name',
          label: 'Nama Barang',
          bold: true,
          fontSize: 7.5,
          bgColor: 'white',
          cells: [{ id: 'c1', field: 'name', label: 'Nama Barang', bold: true, fontSize: 7.5, bgColor: 'white' }],
        },
        {
          id: 'r2',
          type: 'cells',
          field: 'item_code',
          label: 'Kode KIB',
          field2: 'register_number',
          label2: 'No. Reg',
          bold: false,
          fontSize: 6.5,
          bgColor: 'white',
          cells: [
            { id: 'c1', field: 'item_code', label: 'Kode KIB', bold: false, fontSize: 6.5, bgColor: 'white' },
            { id: 'c2', field: 'register_number', label: 'No. Reg', bold: false, fontSize: 6.5, bgColor: 'white' },
          ],
        },
        {
          id: 'r3',
          type: 'cells',
          field: 'sku',
          label: 'SKU / Barcode',
          bold: true,
          fontSize: 6.5,
          bgColor: 'white',
          cells: [{ id: 'c1', field: 'sku', label: 'SKU / Barcode', bold: true, fontSize: 6.5, bgColor: 'white' }],
        },
        {
          id: 'r4',
          type: 'cells',
          field: 'merk',
          label: 'Merk/Tipe',
          field2: 'procurement_year',
          label2: 'Tahun',
          bold: false,
          fontSize: 6.5,
          bgColor: 'white',
          cells: [
            { id: 'c1', field: 'merk', label: 'Merk/Tipe', bold: false, fontSize: 6.5, bgColor: 'white' },
            { id: 'c2', field: 'procurement_year', label: 'Tahun', bold: false, fontSize: 6.5, bgColor: 'white' },
          ],
        },
        {
          id: 'r5',
          type: 'cells',
          field: 'location',
          label: 'Ruangan',
          bold: false,
          fontSize: 6.5,
          bgColor: 'white',
          cells: [{ id: 'c1', field: 'location', label: 'Ruangan', bold: false, fontSize: 6.5, bgColor: 'white' }],
        },
        {
          id: 'r6',
          type: 'cells',
          field: 'funding_source',
          label: 'Sumber Dana',
          bold: false,
          fontSize: 6,
          bgColor: 'white',
          cells: [{ id: 'c1', field: 'funding_source', label: 'Sumber Dana', bold: false, fontSize: 6, bgColor: 'white' }],
        },
      ],
    },
    elements: [],
  },
  {
    id: 'preset_cirebon_bmd_qr',
    name: '🏛️ Kota Cirebon BMD (50 × 60 mm) - QR Kanan & Garis Dobel',
    widthMm: 50,
    heightMm: 60,
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
      cellPaddingMm: 0.8,
      labelColWidthPercent: 30,
      showHeader: true,
      headerText: 'PEMERINTAH KOTA CIREBON',
      headerSubText: 'BADAN PENGELOLAAN KEUANGAN DAN ASET DAERAH',
      headerFontSize: 6.5,
      headerStyle: 'double_line',
      showLogo: true,
      logoUrl: '',
      logoWidthMm: 8,
      qrPosition: 'right',
      qrType: 'qr',
      qrWidthMm: 22,
      showFooter: true,
      footerText: 'Scan QR untuk lapor kerusakan barang',
      footerFontSize: 5.5,
      defaultFontSize: 6.5,
      rows: [
        {
          id: 'r1',
          type: 'cells',
          field: 'name',
          label: 'Nama Barang',
          bold: true,
          fontSize: 7,
          bgColor: 'white',
          cells: [{ id: 'c1', field: 'name', label: 'Nama Barang', bold: true, fontSize: 7, bgColor: 'white' }],
        },
        {
          id: 'r2',
          type: 'cells',
          field: 'item_code',
          label: 'Kode Barang',
          bold: false,
          fontSize: 6.5,
          bgColor: 'white',
          cells: [{ id: 'c1', field: 'item_code', label: 'Kode Barang', bold: false, fontSize: 6.5, bgColor: 'white' }],
        },
        {
          id: 'r3',
          type: 'cells',
          field: 'register_number',
          label: 'NUP',
          bold: false,
          fontSize: 6.5,
          bgColor: 'white',
          cells: [{ id: 'c1', field: 'register_number', label: 'NUP', bold: false, fontSize: 6.5, bgColor: 'white' }],
        },
        {
          id: 'r4',
          type: 'cells',
          field: 'location',
          label: 'Ruangan',
          bold: false,
          fontSize: 6.5,
          bgColor: 'white',
          cells: [{ id: 'c1', field: 'location', label: 'Ruangan', bold: false, fontSize: 6.5, bgColor: 'white' }],
        },
      ],
    },
    elements: [],
  },
]

const templates = ref<LabelTemplate[]>([])
const activeTemplate = ref<LabelTemplate>(JSON.parse(JSON.stringify(DEFAULT_PRESETS[0])))
const activeElementId = ref<string | null>(null)
const selectedTableRowIndex = ref<number | null>(0)
const selectedTableCellIndex = ref<number | null>(0)
const copies = ref(1)
const isPrinting = ref(false)
const paletteTab = ref<'data' | 'borders' | 'header_footer' | 'free_components'>('data')
const tableLogoFileInput = ref<HTMLInputElement | null>(null)

function triggerTableLogoUpload() {
  tableLogoFileInput.value?.click()
}

function onTableLogoSelected(e: Event) {
  const input = e.target as HTMLInputElement
  if (!input.files?.length) return
  const f = input.files[0]
  if (!f.type.startsWith('image/')) {
    toast.add({ severity: 'error', summary: 'Format File Salah', detail: 'Pilih file gambar (PNG, JPG, SVG, WebP)', life: 2500 })
    return
  }
  const reader = new FileReader()
  reader.onload = (ev) => {
    if (activeTemplate.value.tableConfig) {
      activeTemplate.value.tableConfig.logoUrl = ev.target?.result as string
      activeTemplate.value.tableConfig.showLogo = true
      toast.add({ severity: 'success', summary: 'Logo Berhasil Diganti', life: 1500 })
    }
  }
  reader.readAsDataURL(f)
}

function resetTableLogo() {
  if (activeTemplate.value.tableConfig) {
    activeTemplate.value.tableConfig.logoUrl = ''
    toast.add({ severity: 'info', summary: 'Logo Direset ke Lambang Resmi', life: 1500 })
  }
}

// Helper untuk normalisasi sel pada setiap baris
function getRowCells(row: TableRowConfig): TableCellConfig[] {
  if (row.cells && row.cells.length > 0) {
    return row.cells
  }
  if (row.type === 'two_fields') {
    row.cells = [
      { id: 'c1', field: row.field || 'name', label: row.label || 'Kolom 1', bold: row.bold, fontSize: row.fontSize, align: row.align, bgColor: row.bgColor },
      { id: 'c2', field: row.field2 || 'location', label: row.label2 || 'Kolom 2', bold: row.bold2, fontSize: row.fontSize, align: row.align, bgColor: row.bgColor },
    ]
  } else if (row.type === 'custom') {
    row.cells = [
      { id: 'c1', field: '', label: row.label || 'Catatan', customText: row.customText || '', bold: row.bold, fontSize: row.fontSize, align: row.align || 'center', bgColor: row.bgColor },
    ]
  } else {
    row.cells = [
      { id: 'c1', field: row.field || 'name', label: row.label || 'Nama', bold: row.bold, fontSize: row.fontSize, align: row.align, bgColor: row.bgColor },
    ]
  }
  return row.cells
}

function ensureTableConfig(tpl: LabelTemplate) {
  if (!tpl.tableConfig) {
    tpl.tableConfig = JSON.parse(JSON.stringify(DEFAULT_PRESETS[0].tableConfig))
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
    if (!tpl.tableConfig.logoPosition) tpl.tableConfig.logoPosition = 'above_qr'
    if (tpl.tableConfig.showItemNameUnderLogo === undefined) tpl.tableConfig.showItemNameUnderLogo = false
    if (tpl.tableConfig.showHeaderBorder === undefined || (tpl.tableConfig.showHeader && !tpl.tableConfig.showHeaderBorder)) {
      tpl.tableConfig.showHeaderBorder = true
    }

    tpl.tableConfig.rows?.forEach((r) => {
      if (!r.type) r.type = 'cells'
      if (!r.bgColor) r.bgColor = 'white'
      getRowCells(r)
    })
  }

  if (!tpl.layoutMode) {
    tpl.layoutMode = 'table'
  }
}

async function loadSavedTemplates() {
  // 1. Coba muat dari backend database (PostgreSQL)
  try {
    const res = await api.get('/barcode/templates')
    if (res.data && Array.isArray(res.data) && res.data.length > 0) {
      const serverTpls: LabelTemplate[] = res.data.map((t: any) => {
        const item: LabelTemplate = {
          id: t.id,
          name: t.name,
          widthMm: t.widthMm || 64,
          heightMm: t.heightMm || 32,
          layoutMode: t.layoutMode || 'table',
          tableConfig: t.tableConfig,
          elements: t.elements || [],
        }
        ensureTableConfig(item)
        return item
      })
      templates.value = serverTpls
      localStorage.setItem('custom_label_templates_v6', JSON.stringify(serverTpls))
      return
    }
  } catch (err) {
    console.warn('Gagal memuat template dari database, fallback ke localStorage:', err)
  }

  // 2. Fallback ke localStorage jika offline/backend belum terisi
  const raw = localStorage.getItem('custom_label_templates_v6')
  if (raw) {
    try {
      const parsed = JSON.parse(raw)
      parsed.forEach((t: LabelTemplate) => ensureTableConfig(t))
      templates.value = parsed
    } catch {
      templates.value = []
    }
  } else {
    templates.value = []
  }

  // Restore last active WIP design if present
  const lastActive = localStorage.getItem('active_custom_label_template_v6')
  if (lastActive) {
    try {
      const parsedActive = JSON.parse(lastActive)
      ensureTableConfig(parsedActive)
      activeTemplate.value = parsedActive
    } catch {}
  }
}

function saveTemplates() {
  localStorage.setItem('custom_label_templates_v6', JSON.stringify(templates.value))
}

const allPresets = computed(() => {
  const customList = templates.value.filter((t) => !DEFAULT_PRESETS.some((dp) => dp.id === t.id))
  return [...DEFAULT_PRESETS, ...customList]
})

const isCustomActive = computed(() => {
  return !DEFAULT_PRESETS.some((p) => p.id === activeTemplate.value.id)
})

onMounted(() => {
  ensureTableConfig(activeTemplate.value)
  loadSavedTemplates()
})

watch(
  activeTemplate,
  (val) => {
    try {
      localStorage.setItem('active_custom_label_template_v6', JSON.stringify(val))
    } catch {}
  },
  { deep: true }
)

watch(
  () => activeTemplate.value.tableConfig?.showHeader,
  (val) => {
    if (val && activeTemplate.value.tableConfig) {
      activeTemplate.value.tableConfig.showHeaderBorder = true
    }
  }
)

const isTableMode = computed(() => activeTemplate.value.layoutMode === 'table')

const activeElement = computed(() => {
  if (!activeElementId.value) return null
  return activeTemplate.value.elements.find((el) => el.id === activeElementId.value) || null
})

const selectedTableRow = computed(() => {
  if (selectedTableRowIndex.value === null || !activeTemplate.value.tableConfig?.rows) return null
  return activeTemplate.value.tableConfig.rows[selectedTableRowIndex.value] || null
})

const selectedTableCell = computed(() => {
  if (!selectedTableRow.value) return null
  const cells = getRowCells(selectedTableRow.value)
  if (selectedTableCellIndex.value === null || selectedTableCellIndex.value >= cells.length) {
    return cells[0] || null
  }
  return cells[selectedTableCellIndex.value] || null
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

// ==== OPERASI TAMBAH DAN HAPUS BARIS (ROW) ====
function addNewRow(fieldKey = 'name', defaultLabel = 'Nama', insertAfterIndex?: number) {
  ensureTableConfig(activeTemplate.value)
  const newRow: TableRowConfig = {
    id: `r_${Date.now().toString(36)}`,
    type: 'cells',
    field: fieldKey,
    label: defaultLabel,
    bold: fieldKey === 'name' || fieldKey === 'sku',
    fontSize: fieldKey === 'name' ? 7 : activeTemplate.value.tableConfig.defaultFontSize,
    align: 'left',
    bgColor: 'white',
    hasBottomBorder: true,
    cells: [
      {
        id: `c_${Date.now().toString(36)}`,
        field: fieldKey,
        label: defaultLabel,
        bold: fieldKey === 'name' || fieldKey === 'sku',
        fontSize: fieldKey === 'name' ? 7 : activeTemplate.value.tableConfig.defaultFontSize,
        align: 'left',
        bgColor: 'white',
      },
    ],
  }

  if (insertAfterIndex !== undefined && insertAfterIndex >= 0 && insertAfterIndex < activeTemplate.value.tableConfig.rows.length) {
    activeTemplate.value.tableConfig.rows.splice(insertAfterIndex + 1, 0, newRow)
    selectedTableRowIndex.value = insertAfterIndex + 1
  } else {
    activeTemplate.value.tableConfig.rows.push(newRow)
    selectedTableRowIndex.value = activeTemplate.value.tableConfig.rows.length - 1
  }
  selectedTableCellIndex.value = 0
  toast.add({ severity: 'success', summary: 'Baris Ditambahkan', detail: `Baris baru #${selectedTableRowIndex.value + 1}`, life: 1200 })
}

function removeTableRow(index: number) {
  if (!activeTemplate.value.tableConfig || activeTemplate.value.tableConfig.rows.length === 0) return
  activeTemplate.value.tableConfig.rows.splice(index, 1)
  if (selectedTableRowIndex.value === index) {
    selectedTableRowIndex.value = activeTemplate.value.tableConfig.rows.length > 0 ? Math.min(index, activeTemplate.value.tableConfig.rows.length - 1) : null
    selectedTableCellIndex.value = 0
  } else if (selectedTableRowIndex.value !== null && selectedTableRowIndex.value > index) {
    selectedTableRowIndex.value--
  }
  toast.add({ severity: 'info', summary: 'Baris Dihapus', life: 1000 })
}

function duplicateTableRow(index: number) {
  if (!activeTemplate.value.tableConfig) return
  const original = activeTemplate.value.tableConfig.rows[index]
  const copy: TableRowConfig = JSON.parse(JSON.stringify(original))
  copy.id = `r_${Date.now().toString(36)}`
  if (copy.cells) {
    copy.cells.forEach((c) => (c.id = `c_${Date.now().toString(36)}_${Math.random().toString(36).substring(2, 5)}`))
  }
  activeTemplate.value.tableConfig.rows.splice(index + 1, 0, copy)
  selectedTableRowIndex.value = index + 1
  selectedTableCellIndex.value = 0
  toast.add({ severity: 'info', summary: 'Baris Diduplikasi', life: 1000 })
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

// ==== OPERASI TAMBAH DAN HAPUS SEL (CELL) ====
function addCellToRow(rowIndex: number, fieldKey?: string, label?: string) {
  if (!activeTemplate.value.tableConfig?.rows) return
  const row = activeTemplate.value.tableConfig.rows[rowIndex]
  if (!row) return

  const cells = getRowCells(row)
  const f = fieldKey || 'location'
  const l = label || (AVAILABLE_FIELDS.find((x) => x.value === f)?.defaultLabel || 'Kolom')

  const newCell: TableCellConfig = {
    id: `c_${Date.now().toString(36)}_${Math.random().toString(36).substring(2, 5)}`,
    field: f,
    label: l,
    bold: false,
    fontSize: row.fontSize || activeTemplate.value.tableConfig.defaultFontSize,
    align: 'left',
    bgColor: 'white',
  }

  cells.push(newCell)
  row.cells = cells
  row.type = 'cells'

  selectedTableRowIndex.value = rowIndex
  selectedTableCellIndex.value = cells.length - 1
  toast.add({ severity: 'success', summary: 'Sel Ditambahkan', detail: `Kolom ke-${cells.length} di Baris #${rowIndex + 1}`, life: 1200 })
}

function removeCellFromRow(rowIndex: number, cellIndex: number) {
  if (!activeTemplate.value.tableConfig?.rows) return
  const row = activeTemplate.value.tableConfig.rows[rowIndex]
  if (!row) return

  const cells = getRowCells(row)
  if (cells.length > 1) {
    cells.splice(cellIndex, 1)
    row.cells = cells
    if (selectedTableCellIndex.value !== null && selectedTableCellIndex.value >= cells.length) {
      selectedTableCellIndex.value = cells.length - 1
    }
    toast.add({ severity: 'info', summary: 'Sel Dihapus', life: 1000 })
  } else {
    // Jika hanya 1 sel tersisa di baris, hapus seluruh baris tersebut
    removeTableRow(rowIndex)
  }
}

// ==== OPERASI MERGE (GABUNG) DAN SPLIT (PECAH) ====
function mergeRowWithNext(rowIndex: number) {
  if (!activeTemplate.value.tableConfig?.rows) return
  const rows = activeTemplate.value.tableConfig.rows
  if (rowIndex < 0 || rowIndex >= rows.length - 1) return

  const currentRow = rows[rowIndex]
  const nextRow = rows[rowIndex + 1]
  const currentCells = getRowCells(currentRow)
  const nextCells = getRowCells(nextRow)

  nextCells.forEach((c) => {
    currentCells.push({
      ...JSON.parse(JSON.stringify(c)),
      id: `c_${Date.now().toString(36)}_${Math.random().toString(36).substring(2, 5)}`,
    })
  })

  currentRow.cells = currentCells
  currentRow.type = 'cells'

  rows.splice(rowIndex + 1, 1)
  selectedTableRowIndex.value = rowIndex
  selectedTableCellIndex.value = currentCells.length - 1
  toast.add({
    severity: 'success',
    summary: 'Baris Digabung (Merged)',
    detail: `Baris #${rowIndex + 1} kini digabung (${currentCells.length} kolom)`,
    life: 1800,
  })
}

function mergeCellsWithNext(rowIndex: number, cellIndex: number) {
  if (!activeTemplate.value.tableConfig?.rows) return
  const row = activeTemplate.value.tableConfig.rows[rowIndex]
  if (!row) return
  const cells = getRowCells(row)
  if (cellIndex < 0 || cellIndex >= cells.length - 1) return

  const c1 = cells[cellIndex]
  const c2 = cells[cellIndex + 1]

  const mergedLabel = [c1.label, c2.label].filter(Boolean).join(' / ')
  const val1 = c1.customText || (c1.field ? resolveFieldValue(sampleItem.value, c1.field) : '')
  const val2 = c2.customText || (c2.field ? resolveFieldValue(sampleItem.value, c2.field) : '')
  const mergedCustom = [val1, val2].filter(Boolean).join(' - ')

  c1.label = mergedLabel
  c1.customText = mergedCustom
  c1.field = ''

  cells.splice(cellIndex + 1, 1)
  row.cells = cells
  selectedTableCellIndex.value = cellIndex
  toast.add({
    severity: 'success',
    summary: 'Sel Berhasil Digabung (Merged)',
    detail: '2 sel digabung jadi satu kolom',
    life: 1500,
  })
}

function splitCell(rowIndex: number, cellIndex: number) {
  if (!activeTemplate.value.tableConfig?.rows) return
  const row = activeTemplate.value.tableConfig.rows[rowIndex]
  if (!row) return
  const cells = getRowCells(row)
  const currentCell = cells[cellIndex]

  const newCell: TableCellConfig = {
    id: `c_${Date.now().toString(36)}_${Math.random().toString(36).substring(2, 5)}`,
    field: 'location',
    label: 'Kolom Baru',
    bold: false,
    fontSize: currentCell?.fontSize || activeTemplate.value.tableConfig.defaultFontSize,
    align: 'left',
    bgColor: 'white',
  }

  cells.splice(cellIndex + 1, 0, newCell)
  row.cells = cells
  row.type = 'cells'
  selectedTableCellIndex.value = cellIndex + 1
  toast.add({
    severity: 'info',
    summary: 'Kolom Ditambah / Sel Dipecah (Split)',
    life: 1500,
  })
}

function clearAllRows() {
  if (!activeTemplate.value.tableConfig) return
  if (!confirm('Yakin ingin menghapus seluruh baris tabel ini? Anda dapat menambah baris baru tanpa batasan.')) return
  activeTemplate.value.tableConfig.rows = []
  selectedTableRowIndex.value = null
  selectedTableCellIndex.value = null
  toast.add({ severity: 'warn', summary: 'Seluruh Baris Dihapus', life: 1500 })
}

// Tambah baris preset
function addMultiCellRow(cellDefs: Array<{ field: string; label: string }>) {
  ensureTableConfig(activeTemplate.value)
  const cells: TableCellConfig[] = cellDefs.map((cd, idx) => ({
    id: `c_${Date.now().toString(36)}_${idx}`,
    field: cd.field,
    label: cd.label,
    bold: false,
    fontSize: activeTemplate.value.tableConfig.defaultFontSize,
    align: 'left',
    bgColor: 'white',
  }))

  const newRow: TableRowConfig = {
    id: `r_${Date.now().toString(36)}`,
    type: 'cells',
    field: cellDefs[0].field,
    label: cellDefs[0].label,
    bold: false,
    fontSize: activeTemplate.value.tableConfig.defaultFontSize,
    align: 'left',
    bgColor: 'white',
    hasBottomBorder: true,
    cells,
  }

  activeTemplate.value.tableConfig.rows.push(newRow)
  selectedTableRowIndex.value = activeTemplate.value.tableConfig.rows.length - 1
  selectedTableCellIndex.value = 0
  toast.add({ severity: 'success', summary: 'Baris Kolom Ditambahkan', detail: `${cells.length} Sel`, life: 1200 })
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
    cells: [
      {
        id: `c_${Date.now().toString(36)}`,
        field: '',
        label: 'Catatan',
        customText: 'BARANG ASLI MILIK PEMERINTAH DAERAH',
        bold: true,
        fontSize: 6,
        align: 'center',
        bgColor: 'light_gray',
      },
    ],
  }
  activeTemplate.value.tableConfig.rows.push(newRow)
  selectedTableRowIndex.value = activeTemplate.value.tableConfig.rows.length - 1
  selectedTableCellIndex.value = 0
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
  selectedTableCellIndex.value = 0
  toast.add({ severity: 'info', summary: 'Baris Barcode Garis Ditambahkan', life: 1200 })
}

// ==== KONVERSI MODE TABEL -> MODE BEBAS ====
function convertTableToFreeMode() {
  const tpl = activeTemplate.value
  const tc = tpl.tableConfig
  if (!tc || tpl.layoutMode !== 'table') return
  const W = tpl.widthMm
  const bw = tc.borderWidth || 1
  const els: LabelElement[] = []
  const uid = () => `el_${Math.random().toString(36).slice(2, 9)}`
  let y = 0

  // Header (Kop) - logo kiri + teks tengah
  if (tc.showHeader) {
    const headerH = tc.headerStyle === 'double_line' ? 12 : 8
    if (tc.showLogo) {
      els.push({ id: uid(), type: 'logo', logoUrl: tc.logoUrl || '', x: 1, y, w: 8, h: headerH, fontSize: tc.headerFontSize })
    }
    const textX = tc.showLogo ? 10 : 1
    const textW = W - textX - 1
    if (tc.headerStyle === 'double_line') {
      els.push({ id: uid(), type: 'header', text: tc.headerText || '', x: textX, y, w: textW, h: headerH / 2, fontSize: tc.headerFontSize, bold: true, align: 'center' })
      els.push({ id: uid(), type: 'header', text: tc.headerSubText || '', x: textX, y: y + headerH / 2, w: textW, h: headerH / 2, fontSize: (tc.headerFontSize || 6.5) - 0.5, bold: true, align: 'center' })
      els.push({ id: uid(), type: 'line', lineOrientation: 'horizontal', x: 0, y: y + headerH, w: W, h: 0.5, border: true })
    } else {
      els.push({ id: uid(), type: 'header', text: tc.headerText || '', x: textX, y, w: textW, h: headerH, fontSize: tc.headerFontSize, bold: true, align: 'center', border: true })
    }
    y += headerH
  }

  // Zona konten: QR kiri/kanan + baris field
  const qrW = tc.qrPosition !== 'none' ? tc.qrWidthMm || 20 : 0
  const rowsH = tc.rows.reduce((sum, r) => sum + 7, 0)
  const contentX = tc.qrPosition === 'left' ? qrW : 1
  const contentW = W - (qrW > 0 ? qrW + 2 : 2)
  const startY = y

  for (const r of tc.rows) {
    const rowH = 7
    const cells = r.type === 'cells' && r.cells?.length ? r.cells : [r]
    const cellH = rowH / cells.length
    cells.forEach((c, idx) => {
      if (r.type === 'barcode_strip') {
        els.push({ id: uid(), type: 'code128', x: contentX, y: y + idx * cellH, w: contentW, h: cellH, fontSize: c.fontSize || 5.5, bold: c.bold, align: c.align })
        return
      }
      const fieldEl: LabelElement = {
        id: uid(),
        type: 'field',
        field: c.field,
        prefix: c.customText ? `${c.customText} ` : undefined,
        x: contentX,
        y: y + idx * cellH,
        w: contentW,
        h: cellH,
        fontSize: c.fontSize || tc.defaultFontSize || 6.5,
        bold: c.bold,
        align: c.align,
      }
      // Baris pertama & terakhir pakai garis pemisah (pengganti grid)
      els.push(fieldEl)
      if (idx < cells.length - 1) {
        els.push({ id: uid(), type: 'line', lineOrientation: 'horizontal', x: contentX, y: y + (idx + 1) * cellH, w: contentW, h: 0.5, border: true })
      }
    })
    y += rowH
  }

  // QR / Barcode
  if (qrW > 0) {
    const qrType = tc.qrType === 'code128' ? 'code128' : 'qr'
    els.push({
      id: uid(),
      type: qrType,
      x: tc.qrPosition === 'left' ? 1 : W - qrW - 1,
      y: startY,
      w: qrW,
      h: Math.max(rowsH, 18),
      fontSize: 5.5,
    })
  }

  // Footer
  if (tc.showFooter) {
    els.push({ id: uid(), type: 'static_text', text: tc.footerText || '', x: 0, y: y + 0.5, w: W, h: 5, fontSize: tc.footerFontSize || 5.5, align: 'center', border: true })
  }

  // Bingkai luar
  if (tc.showOuterBorder) {
    els.push({ id: uid(), type: 'line', lineOrientation: 'horizontal', x: 0, y: 0, w: W, h: 0.5, border: true })
    els.push({ id: uid(), type: 'line', lineOrientation: 'horizontal', x: 0, y: tpl.heightMm - 0.5, w: W, h: 0.5, border: true })
    els.push({ id: uid(), type: 'line', lineOrientation: 'vertical', x: 0, y: 0, w: 0.5, h: tpl.heightMm, border: true })
    els.push({ id: uid(), type: 'line', lineOrientation: 'vertical', x: W - 0.5, y: 0, w: 0.5, h: tpl.heightMm, border: true })
  }

  tpl.elements = els
  tpl.layoutMode = 'free'
  activeElementId.value = null
  saveTemplates()
  toast.add({ severity: 'success', summary: 'Konversi Berhasil', detail: `${els.length} elemen dibuat dari mode tabel. Posisi bisa disesuaikan.`, life: 2500 })
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
  selectedTableRowIndex.value = 0
  selectedTableCellIndex.value = 0
  toast.add({ severity: 'info', summary: 'Preset Dimuat', detail: tpl.name, life: 1800 })
}

function resetLayout() {
  const defaultTpl = DEFAULT_PRESETS.find((p) => p.id === activeTemplate.value.id) || DEFAULT_PRESETS[0]
  const clone = JSON.parse(JSON.stringify(defaultTpl))
  ensureTableConfig(clone)
  activeTemplate.value = clone
  activeElementId.value = null
  selectedTableRowIndex.value = 0
  selectedTableCellIndex.value = 0
  toast.add({ severity: 'info', summary: 'Tata Letak Direset ke Bawaan', life: 1500 })
}

async function saveCurrentAsNewTemplate() {
  const name = prompt('Beri nama template label ini:', activeTemplate.value.name.replace(/^[🌟🏷️\s]+/, ''))
  if (!name) return
  const newTpl: LabelTemplate = {
    ...JSON.parse(JSON.stringify(activeTemplate.value)),
    id: `custom_${Date.now().toString(36)}`,
    name: `🏷️ ${name.trim()}`,
  }
  templates.value.push(newTpl)
  activeTemplate.value.id = newTpl.id
  activeTemplate.value.name = newTpl.name
  saveTemplates()

  // Sync ke database
  try {
    await api.post('/barcode/templates', newTpl)
    toast.add({ severity: 'success', summary: 'Tersimpan di Database', detail: newTpl.name, life: 2500 })
  } catch (err: any) {
    toast.add({ severity: 'warn', summary: 'Tersimpan Lokal', detail: 'Tersimpan di browser, gagal kirim ke database: ' + (err.response?.data?.error || err.message), life: 3500 })
  }
}

async function updateCurrentCustomTemplate() {
  const idx = templates.value.findIndex((t) => t.id === activeTemplate.value.id)
  if (idx !== -1) {
    const updated = JSON.parse(JSON.stringify(activeTemplate.value))
    templates.value[idx] = updated
    saveTemplates()

    // Sync ke database
    try {
      await api.post('/barcode/templates', updated)
      toast.add({ severity: 'success', summary: 'Database Diperbarui', detail: updated.name, life: 2500 })
    } catch (err: any) {
      toast.add({ severity: 'warn', summary: 'Lokal Diperbarui', detail: 'Gagal sinkron database: ' + (err.response?.data?.error || err.message), life: 3500 })
    }
  } else {
    saveCurrentAsNewTemplate()
  }
}

async function deleteCurrentCustomTemplate() {
  const target = activeTemplate.value
  if (!confirm(`Hapus template "${target.name}"?`)) return
  templates.value = templates.value.filter((t) => t.id !== target.id)
  saveTemplates()
  applyPreset(DEFAULT_PRESETS[0])

  // Sync ke database
  try {
    await api.delete(`/barcode/templates/${target.id}`)
    toast.add({ severity: 'info', summary: 'Template Dihapus dari Database', life: 2000 })
  } catch (err: any) {
    toast.add({ severity: 'info', summary: 'Template Dihapus Lokal', life: 1800 })
  }
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
    header="Desainer Label QR &amp; Barcode Aset (Tambah &amp; Hapus Baris / Sel Penuh)"
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
            <span>Mode Tabel (Tambah &amp; Hapus Baris/Sel)</span>
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

        <Button
          v-if="isTableMode"
          icon="pi pi-objects-column"
          label="Konversi ke Mode Bebas"
          size="small"
          severity="secondary"
          outlined
          class="!text-xs !py-1 shrink-0"
          @click="convertTableToFreeMode"
        />

        <div class="h-4 w-px bg-slate-200 mx-0.5 shrink-0" />

        <!-- Preset Selector -->
        <div class="flex items-center gap-1.5 shrink-0">
          <span class="text-slate-500 font-semibold">Preset:</span>
          <Select
            v-model="activeTemplate.id"
            :options="allPresets"
            optionLabel="name"
            optionValue="id"
            class="w-64 !text-xs"
            @change="(e) => {
              const p = allPresets.find((x) => x.id === e.value)
              if (p) applyPreset(p)
            }"
          />
          <Button
            v-if="isCustomActive"
            icon="pi pi-trash"
            text
            rounded
            size="small"
            severity="danger"
            v-tooltip.top="'Hapus Template Kustom Ini'"
            class="!p-1 shrink-0"
            @click="deleteCurrentCustomTemplate"
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

        <!-- Save / Update Template -->
        <div class="flex items-center gap-1 shrink-0">
          <Button
            v-if="isCustomActive"
            icon="pi pi-check"
            label="Perbarui Template"
            size="small"
            severity="info"
            class="!text-xs !py-1 shrink-0"
            v-tooltip.top="'Simpan perubahan ke template ini'"
            @click="updateCurrentCustomTemplate"
          />
          <Button
            icon="pi pi-save"
            :label="isCustomActive ? 'Simpan Baru' : 'Simpan Template'"
            size="small"
            severity="secondary"
            class="!text-xs !py-1 shrink-0"
            v-tooltip.top="'Simpan sebagai template kustom baru'"
            @click="saveCurrentAsNewTemplate"
          />
        </div>

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
      <div class="w-84 border-r border-slate-200 bg-white flex flex-col overflow-hidden select-none">
        <!-- Palette Tab Headers -->
        <div class="p-2 border-b border-slate-200 bg-slate-50 flex items-center gap-1 overflow-x-auto">
          <button
            type="button"
            class="tab-btn"
            :class="paletteTab === 'data' ? 'tab-btn-active' : ''"
            @click="paletteTab = 'data'"
          >
            <i class="pi pi-list text-xs mr-1" />
            <span>{{ isTableMode ? 'Data, Baris & Sel' : 'Data Form' }}</span>
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

        <!-- 1. TAB DATA & BARIS (MODE TABEL - FITUR TAMBAH & HAPUS BARIS/SEL) -->
        <div v-if="paletteTab === 'data' && isTableMode" class="flex-1 overflow-y-auto p-2.5 flex flex-col gap-2.5">
          <!-- A. KONTROL CEPAT TAMBAH & HAPUS BARIS / SEL -->
          <div class="flex flex-col gap-1.5 p-2 bg-indigo-50/70 rounded-lg border border-indigo-200 shadow-2xs">
            <span class="text-[10.5px] font-bold text-indigo-950 uppercase tracking-wide flex items-center justify-between">
              <span>Aksi Baris &amp; Sel:</span>
              <span class="text-[10px] text-indigo-700 font-semibold">Total {{ activeTemplate.tableConfig.rows.length }} Baris</span>
            </span>

            <div class="grid grid-cols-2 gap-1.5 pt-0.5">
              <!-- Tambah Baris Baru -->
              <button
                type="button"
                class="px-2 py-1.5 bg-indigo-600 hover:bg-indigo-700 text-white rounded text-left flex items-center gap-1.5 text-[11px] font-bold shadow-2xs transition-all"
                @click="addNewRow('name', 'Nama')"
              >
                <i class="pi pi-plus-circle text-xs" />
                <span>+ Baris Baru</span>
              </button>

              <!-- Tambah Sel di Baris Terpilih -->
              <button
                type="button"
                class="px-2 py-1.5 bg-white border border-indigo-300 hover:border-indigo-600 rounded text-left flex items-center gap-1.5 text-[11px] font-bold text-indigo-900 shadow-2xs transition-all disabled:opacity-50"
                :disabled="selectedTableRowIndex === null"
                @click="selectedTableRowIndex !== null && addCellToRow(selectedTableRowIndex)"
              >
                <i class="pi pi-table text-indigo-600 text-xs" />
                <span>+ Sel / Kolom</span>
              </button>

              <!-- Hapus Baris Terpilih -->
              <button
                type="button"
                class="px-2 py-1.5 bg-white border border-rose-200 hover:bg-rose-50 text-rose-700 rounded text-left flex items-center gap-1.5 text-[11px] font-semibold shadow-2xs transition-all disabled:opacity-50"
                :disabled="selectedTableRowIndex === null"
                @click="selectedTableRowIndex !== null && removeTableRow(selectedTableRowIndex)"
              >
                <i class="pi pi-trash text-xs text-rose-600" />
                <span>✕ Hapus Baris</span>
              </button>

              <!-- Hapus Sel Terpilih -->
              <button
                type="button"
                class="px-2 py-1.5 bg-white border border-amber-200 hover:bg-amber-50 text-amber-800 rounded text-left flex items-center gap-1.5 text-[11px] font-semibold shadow-2xs transition-all disabled:opacity-50"
                :disabled="selectedTableRowIndex === null || selectedTableCellIndex === null"
                @click="selectedTableRowIndex !== null && selectedTableCellIndex !== null && removeCellFromRow(selectedTableRowIndex, selectedTableCellIndex)"
              >
                <i class="pi pi-times-circle text-xs text-amber-600" />
                <span>✕ Hapus Sel</span>
              </button>
            </div>

            <!-- Shortcut Jenis Baris Multi-Kolom -->
            <div class="grid grid-cols-2 gap-1.5 pt-1 border-t border-indigo-100">
              <button
                type="button"
                class="px-2 py-1 bg-white border border-slate-200 hover:border-indigo-400 rounded text-left flex items-center gap-1.5 text-[10.5px] font-semibold text-slate-700"
                @click="addMultiCellRow([{ field: 'location', label: 'Ruang' }, { field: 'procurement_year', label: 'Tahun' }])"
              >
                <i class="pi pi-pause text-indigo-500 text-[10px]" />
                <span>Baris 2 Sel</span>
              </button>
              <button
                type="button"
                class="px-2 py-1 bg-white border border-slate-200 hover:border-indigo-400 rounded text-left flex items-center gap-1.5 text-[10.5px] font-semibold text-slate-700"
                @click="addMultiCellRow([{ field: 'location', label: 'Ruang' }, { field: 'procurement_year', label: 'Tahun' }, { field: 'condition_status', label: 'Kondisi' }])"
              >
                <i class="pi pi-bars text-indigo-500 text-[10px]" />
                <span>Baris 3 Sel</span>
              </button>
              <button
                type="button"
                class="px-2 py-1 bg-white border border-slate-200 hover:border-indigo-400 rounded text-left flex items-center gap-1.5 text-[10.5px] font-semibold text-slate-700"
                @click="addCustomTextRow"
              >
                <i class="pi pi-pencil text-purple-600 text-[10px]" />
                <span>Teks Catatan</span>
              </button>
              <button
                type="button"
                class="px-2 py-1 bg-white border border-slate-200 hover:border-indigo-400 rounded text-left flex items-center gap-1.5 text-[10.5px] font-semibold text-slate-700"
                @click="addBarcodeStripRow"
              >
                <i class="pi pi-barcode text-cyan-600 text-[10px]" />
                <span>Barcode 128 Strip</span>
              </button>
            </div>
          </div>

          <!-- B. DAFTAR 17 FIELD FORMULIR BARANG -->
          <div class="flex flex-col gap-1">
            <div class="text-[11px] text-slate-500 px-1 pb-0.5 font-medium flex items-center justify-between">
              <span>Field Formulir (Tambah Baris / Sel):</span>
              <span class="text-[10px] text-indigo-600 font-bold">{{ AVAILABLE_FIELDS.length }} Field</span>
            </div>

            <div
              v-for="f in AVAILABLE_FIELDS"
              :key="f.value"
              class="flex items-center justify-between p-1.5 bg-white border border-slate-200 hover:border-indigo-400 rounded-md transition-all gap-1 shadow-2xs"
            >
              <div class="flex items-center gap-1.5 truncate flex-1 min-w-0">
                <i :class="[f.icon, 'text-slate-500 text-xs shrink-0']" />
                <span class="truncate text-[11px] font-semibold text-slate-700">{{ f.label }}</span>
              </div>
              <div class="flex items-center gap-1 shrink-0">
                <!-- Tambah sebagai Baris Baru -->
                <button
                  type="button"
                  class="px-1.5 py-0.5 rounded bg-indigo-50 hover:bg-indigo-600 hover:text-white text-indigo-700 border border-indigo-200 text-[10px] font-bold transition-all"
                  title="Tambah sebagai Baris Baru"
                  @click="addNewRow(f.value, f.defaultLabel)"
                >
                  + Baris
                </button>
                <!-- Tambah sebagai Sel ke Baris Terpilih -->
                <button
                  type="button"
                  class="px-1.5 py-0.5 rounded bg-slate-100 hover:bg-emerald-600 hover:text-white text-slate-700 border border-slate-200 text-[10px] font-bold transition-all disabled:opacity-40"
                  :disabled="selectedTableRowIndex === null"
                  title="Tambah sebagai Sel Baru di Baris Terpilih"
                  @click="selectedTableRowIndex !== null && addCellToRow(selectedTableRowIndex, f.value, f.defaultLabel)"
                >
                  + Sel
                </button>
              </div>
            </div>
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
              <span class="font-semibold">Garis Vertikal Pemisah Label &amp; Sel</span>
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

              <!-- Lambang Kota / Logo di Atas Barcode / QR -->
              <div class="pt-2 border-t border-slate-200 flex flex-col gap-2">
                <label class="flex items-center gap-1.5 cursor-pointer text-slate-800 font-bold text-xs">
                  <Checkbox v-model="activeTemplate.tableConfig.showLogo" binary />
                  <span>Lambang Kota di Atas QR (Atas-Bawah)</span>
                </label>

                <div v-if="activeTemplate.tableConfig.showLogo" class="flex flex-col gap-2 pl-4 border-l-2 border-indigo-200">
                  <input type="file" ref="tableLogoFileInput" accept="image/*" class="hidden" @change="onTableLogoSelected" />
                  <div class="flex items-center gap-1.5">
                    <Button
                      type="button"
                      icon="pi pi-upload"
                      label="Unggah / Ganti Logo"
                      size="small"
                      class="!text-xs !py-1 flex-1"
                      @click="triggerTableLogoUpload"
                    />
                    <Button
                      v-if="activeTemplate.tableConfig.logoUrl"
                      type="button"
                      icon="pi pi-refresh"
                      label="Reset"
                      size="small"
                      severity="secondary"
                      outlined
                      class="!text-xs !py-1 shrink-0"
                      title="Kembalikan ke lambang resmi Kota Cirebon"
                      @click="resetTableLogo"
                    />
                  </div>
                  <div class="flex flex-col gap-1">
                    <label class="text-[10px] text-slate-500 font-semibold">Atau Masukkan URL Logo:</label>
                    <InputText v-model="activeTemplate.tableConfig.logoUrl" placeholder="Kosongkan untuk Lambang Kota Cirebon" class="w-full !text-xs !py-1" />
                  </div>
                  <label class="flex items-center gap-1.5 cursor-pointer text-slate-700 text-xs">
                    <Checkbox v-model="activeTemplate.tableConfig.showItemNameUnderLogo" binary />
                    <span>Tampilkan Nama Barang di Bawah (- Nama)</span>
                  </label>
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
              <label v-if="activeTemplate.tableConfig.showHeader" class="flex items-center gap-1.5 cursor-pointer text-indigo-700 text-xs font-semibold">
                <Checkbox v-model="activeTemplate.tableConfig.showHeaderBorder" binary />
                <span>Garis Bawah Kop</span>
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
          <span v-if="isTableMode">💡 <b>Mode Tabel Kustom:</b> Tambah atau hapus baris &amp; sel langsung melalui tombol hover di kanvas atau panel samping.</span>
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
          <!-- Sticker Label Badge Info + Quick Row Adder -->
          <div class="flex items-center gap-2 text-xs font-semibold text-slate-600 bg-white/80 px-3 py-1 rounded-full border border-slate-300 shadow-xs">
            <span>🏷️ {{ activeTemplate.widthMm }} &times; {{ activeTemplate.heightMm }} mm</span>
            <span class="text-slate-300">|</span>
            <span class="text-indigo-600 font-bold">{{ isTableMode ? 'Tabel Kustom Penuh' : 'Mode Bebas' }}</span>
            <span class="text-slate-300">|</span>
            <span class="text-slate-500 font-normal">Sampel: {{ sampleItem.name }}</span>
            <template v-if="isTableMode">
              <span class="text-slate-300">|</span>
              <button
                type="button"
                class="px-2 py-0.5 rounded bg-indigo-600 hover:bg-indigo-700 text-white font-bold text-[11px] flex items-center gap-1 shadow-2xs transition-all"
                @click="addNewRow('name', 'Nama')"
              >
                <i class="pi pi-plus text-[10px]" />
                <span>Tambah Baris</span>
              </button>
            </template>
          </div>

          <!-- THE PHYSICAL STICKER -->
          <div
            class="relative bg-white shadow-xl rounded-[4px] text-black select-none overflow-hidden transition-all flex flex-col"
            :style="{
              width: `${activeTemplate.widthMm * MM_TO_PX * zoom}px`,
              height: `${activeTemplate.heightMm * MM_TO_PX * zoom}px`,
            }"
            @click.self="activeElementId = null; selectedTableRowIndex = null; selectedTableCellIndex = null"
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
                  borderBottom: activeTemplate.tableConfig.showHeaderBorder !== false
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
                  class="shrink-0 flex flex-col items-center justify-between overflow-hidden"
                  :style="{
                    width: `${activeTemplate.tableConfig.qrWidthMm * MM_TO_PX * zoom}px`,
                    borderRight: activeTemplate.tableConfig.showColBorder
                      ? `${activeTemplate.tableConfig.borderWidth * (zoom / 2.2)}px ${activeTemplate.tableConfig.borderStyle} ${activeTemplate.tableConfig.borderColor}`
                      : 'none',
                  }"
                >
                  <!-- Lambang Kota di Atas jika showLogo aktif -->
                  <div
                    v-if="activeTemplate.tableConfig.showLogo"
                    class="w-full flex-1 flex flex-col items-center justify-center p-0.5 overflow-hidden"
                    :style="{
                      borderBottom: `${activeTemplate.tableConfig.borderWidth * (zoom / 2.2)}px ${activeTemplate.tableConfig.borderStyle} ${activeTemplate.tableConfig.borderColor}`,
                      minHeight: `${(activeTemplate.tableConfig.logoWidthMm || 10) * MM_TO_PX * (zoom / 2.2) * 0.6}px`,
                    }"
                  >
                    <img
                      v-if="activeTemplate.tableConfig.logoUrl"
                      :src="activeTemplate.tableConfig.logoUrl"
                      class="max-w-full max-h-full object-contain pointer-events-none"
                      alt="Logo"
                    />
                    <svg
                      v-else
                      viewBox="0 0 100 120"
                      class="w-full h-full max-h-12 object-contain pointer-events-none"
                    >
                      <path d="M14 10 H86 V72 C86 98 50 114 50 114 C50 114 14 98 14 72 Z" fill="#ffffff" stroke="#1e293b" stroke-width="3" />
                      <path d="M16 12 H84 V56 C84 56 50 64 50 64 C50 64 16 56 16 56 Z" fill="#38bdf8" />
                      <polygon points="50,17 53,24 61,24 54,29 57,36 50,31 43,36 46,29 39,24 47,24" fill="#facc15" stroke="#ca8a04" stroke-width="0.8" />
                      <rect x="22" y="38" width="56" height="14" fill="#1e293b" />
                      <rect x="25" y="34" width="8" height="4" fill="#1e293b" />
                      <rect x="37" y="34" width="8" height="4" fill="#1e293b" />
                      <rect x="49" y="34" width="8" height="4" fill="#1e293b" />
                      <rect x="61" y="34" width="8" height="4" fill="#1e293b" />
                      <path d="M16 56 Q33 50 50 58 Q67 66 84 56 V72 C84 96 50 112 50 112 C50 112 16 96 16 72 Z" fill="#0284c7" />
                      <path d="M16 66 Q33 60 50 68 Q67 76 84 66 V76 C84 96 50 112 50 112 C50 112 16 96 16 76 Z" fill="#0369a1" />
                      <path d="M38 72 C32 68 36 60 48 62 C58 64 63 70 58 76 C52 82 42 80 38 74" fill="none" stroke="#f97316" stroke-width="3" stroke-linecap="round" />
                      <circle cx="56" cy="71" r="2.5" fill="#f97316" />
                      <path d="M22 96 Q50 88 78 96" fill="none" stroke="#ffffff" stroke-width="3.5" stroke-linecap="round" />
                      <text x="50" y="98" font-size="5" font-weight="900" text-anchor="middle" fill="#0f172a" font-family="sans-serif">CIREBON</text>
                    </svg>
                  </div>

                  <!-- QR Code di Bawah Lambang -->
                  <div class="w-full flex-1 flex flex-col items-center justify-center p-1 overflow-hidden">
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

                  <!-- Nama Barang di Bawah Logo/QR jika aktif -->
                  <div
                    v-if="activeTemplate.tableConfig.showItemNameUnderLogo"
                    class="w-full py-0.5 px-0.5 text-center truncate font-bold text-slate-800 border-t border-slate-300"
                    :style="{ fontSize: `${5 * (zoom / 2.2)}pt` }"
                  >
                    - {{ sampleItem.name }}
                  </div>
                </div>

                <!-- Table Data Rows Subtable -->
                <div class="flex-1 flex flex-col justify-between overflow-hidden">
                  <div
                    v-for="(row, rIdx) in activeTemplate.tableConfig.rows"
                    :key="row.id"
                    class="group/row flex-1 flex items-stretch cursor-pointer transition-colors relative"
                    :class="[
                      row.bgColor === 'light_gray' ? 'bg-slate-100' : row.bgColor === 'dark' ? 'bg-slate-800 text-white' : 'bg-white',
                      selectedTableRowIndex === rIdx ? 'ring-1 ring-indigo-500 bg-indigo-50/50' : 'hover:bg-indigo-50/20',
                    ]"
                    :style="{
                      borderBottom: (activeTemplate.tableConfig.showGridLines && rIdx < activeTemplate.tableConfig.rows.length - 1) || row.hasBottomBorder
                        ? `${activeTemplate.tableConfig.borderWidth * (zoom / 2.2)}px ${activeTemplate.tableConfig.borderStyle} ${activeTemplate.tableConfig.borderColor}`
                        : 'none',
                    }"
                    @click.stop="selectedTableRowIndex = rIdx; selectedTableCellIndex = 0"
                  >
                    <!-- TIPE KHUSUS: Barcode Strip Memanjang -->
                    <template v-if="row.type === 'barcode_strip'">
                      <div class="w-full flex flex-col items-center justify-center p-0.5">
                        <img :src="`/api/barcode/${sampleItem.id}.png?fmt=code128`" class="w-full h-4 object-contain pointer-events-none" alt="BC" />
                        <span class="font-mono font-bold text-[7pt] leading-none">{{ sampleItem.sku }}</span>
                      </div>
                    </template>

                    <!-- TIPE STANDARD DENGAN MULTI-SEL (CELLS) -->
                    <template v-else>
                      <div
                        v-for="(cell, cIdx) in getRowCells(row)"
                        :key="cell.id"
                        class="group/cell flex-1 flex items-center min-w-0 relative transition-all"
                        :class="[
                          cell.bgColor === 'light_gray' ? 'bg-slate-100' : cell.bgColor === 'dark' ? 'bg-slate-800 text-white' : '',
                          selectedTableRowIndex === rIdx && selectedTableCellIndex === cIdx ? 'ring-1 ring-emerald-500 bg-emerald-50/60 font-semibold' : 'hover:bg-emerald-50/30',
                        ]"
                        :style="{
                          padding: `${activeTemplate.tableConfig.cellPaddingMm * MM_TO_PX * (zoom / 2.2)}px 4px`,
                          borderRight: activeTemplate.tableConfig.showLabelColBorder && cIdx < getRowCells(row).length - 1
                            ? `${activeTemplate.tableConfig.borderWidth * (zoom / 2.2)}px ${activeTemplate.tableConfig.borderStyle} ${activeTemplate.tableConfig.borderColor}`
                            : 'none',
                          fontSize: `${(cell.fontSize || row.fontSize || activeTemplate.tableConfig.defaultFontSize) * (zoom / 2.2)}pt`,
                        }"
                        @click.stop="selectedTableRowIndex = rIdx; selectedTableCellIndex = cIdx"
                      >
                        <!-- Label Sel -->
                        <span
                          v-if="cell.label"
                          class="font-bold shrink-0 truncate pr-1"
                          :style="{
                            maxWidth: getRowCells(row).length === 1 ? `${activeTemplate.tableConfig.labelColWidthPercent}%` : '45%',
                          }"
                        >
                          {{ cell.label }}:
                        </span>

                        <!-- Nilai Sel / Teks Bebas -->
                        <span
                          class="flex-1 truncate"
                          :class="[
                            cell.bold ? 'font-bold' : 'font-normal',
                            cell.field === 'sku' ? 'font-mono' : '',
                            cell.align === 'center' ? 'text-center' : cell.align === 'right' ? 'text-right' : 'text-left',
                          ]"
                        >
                          {{ cell.customText || resolveFieldValue(sampleItem, cell.field) }}
                        </span>

                        <!-- Tombol Cepat Aksi Sel (Pecah/Split, Gabung/Merge, Hapus Sel) -->
                        <div class="opacity-0 group-hover/cell:opacity-100 absolute right-0.5 top-0.5 flex items-center gap-0.5 z-10 bg-white/95 rounded px-0.5 shadow-2xs border border-slate-200">
                          <!-- Split / Pecah Kolom -->
                          <button
                            type="button"
                            class="p-0.5 rounded hover:bg-indigo-600 hover:text-white text-indigo-600 text-[8px]"
                            title="Pecah Sel / Tambah Kolom (Split)"
                            @click.stop="splitCell(rIdx, cIdx)"
                          >
                            <i class="pi pi-arrows-h text-[8px]" />
                          </button>
                          <!-- Merge / Gabung dengan Sel Kanan -->
                          <button
                            v-if="cIdx < getRowCells(row).length - 1"
                            type="button"
                            class="p-0.5 rounded hover:bg-emerald-600 hover:text-white text-emerald-600 text-[8px]"
                            title="Gabung dengan Sel Kanan (Merge)"
                            @click.stop="mergeCellsWithNext(rIdx, cIdx)"
                          >
                            <i class="pi pi-link text-[8px]" />
                          </button>
                          <!-- Hapus Sel Ini -->
                          <button
                            type="button"
                            class="p-0.5 rounded hover:bg-rose-600 hover:text-white text-rose-500 text-[8px]"
                            title="Hapus Sel Ini"
                            @click.stop="removeCellFromRow(rIdx, cIdx)"
                          >
                            <i class="pi pi-times text-[8px]" />
                          </button>
                        </div>
                      </div>
                    </template>

                    <!-- TOMBOL CEPAT BARIS (Hover Overlay di ujung kanan baris) -->
                    <div class="opacity-0 group-hover/row:opacity-100 absolute right-0 top-0 bottom-0 flex items-center gap-0.5 bg-white/95 px-1 border-l border-slate-300 shadow-xs transition-opacity z-20">
                      <!-- Tambah Sel ke Baris Ini -->
                      <button
                        type="button"
                        class="p-0.5 rounded bg-indigo-50 hover:bg-indigo-600 hover:text-white text-indigo-700 text-[10px]"
                        title="Tambah Sel / Kolom di Baris Ini"
                        @click.stop="addCellToRow(rIdx)"
                      >
                        <i class="pi pi-plus-circle text-[10px]" />
                      </button>
                      <!-- Sisip Baris Baru Tepat di Bawah Ini -->
                      <button
                        type="button"
                        class="p-0.5 rounded bg-slate-100 hover:bg-indigo-600 hover:text-white text-slate-700 text-[10px]"
                        title="Sisipkan Baris Baru di Bawah Ini"
                        @click.stop="addNewRow('location', 'Ruang', rIdx)"
                      >
                        <i class="pi pi-plus text-[10px]" />
                      </button>
                      <!-- Gabung (Merge) dengan Baris di Bawah Ini -->
                      <button
                        v-if="rIdx < activeTemplate.tableConfig.rows.length - 1"
                        type="button"
                        class="p-0.5 rounded bg-emerald-50 hover:bg-emerald-600 hover:text-white text-emerald-700 text-[10px]"
                        title="Gabung (Merge) dengan Baris di Bawahnya"
                        @click.stop="mergeRowWithNext(rIdx)"
                      >
                        <i class="pi pi-link text-[10px]" />
                      </button>
                      <!-- Hapus Baris Ini -->
                      <button
                        type="button"
                        class="p-0.5 rounded bg-rose-50 hover:bg-rose-600 hover:text-white text-rose-600 text-[10px]"
                        title="Hapus Seluruh Baris Ini"
                        @click.stop="removeTableRow(rIdx)"
                      >
                        <i class="pi pi-trash text-[10px]" />
                      </button>
                    </div>
                  </div>

                  <div
                    v-if="activeTemplate.tableConfig.rows.length === 0"
                    class="flex-1 flex flex-col items-center justify-center text-slate-400 text-xs p-4 text-center gap-2"
                  >
                    <span>Tabel masih kosong.</span>
                    <button
                      type="button"
                      class="px-2.5 py-1 bg-indigo-600 text-white rounded font-bold text-xs shadow-xs"
                      @click="addNewRow('name', 'Nama')"
                    >
                      + Tambah Baris Pertama
                    </button>
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
                    el.border ? 'border border-black' : '',
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
                <div
                  v-else-if="el.type === 'line'"
                  class="pointer-events-none border-black my-auto"
                  :class="[el.lineOrientation === 'vertical' ? 'h-full mx-auto border-l' : 'w-full border-t']"
                />

                <!-- Logo -->
                <div v-else-if="el.type === 'logo'" class="w-full h-full p-0.5 pointer-events-none flex items-center justify-center">
                  <img v-if="el.logoUrl" :src="el.logoUrl" class="max-w-full max-h-full object-contain" alt="Logo" />
                  <div v-else class="w-full h-full border border-dashed border-slate-400 rounded-xs flex items-center justify-center text-slate-400" :style="{ fontSize: `${6 * (zoom / 2)}pt` }">
                    Logo
                  </div>
                </div>

                <!-- Resize Handle -->
                <div
                  v-if="activeElementId === el.id"
                  class="absolute -right-1 -bottom-1 w-3.5 h-3.5 bg-indigo-600 border-2 border-white rounded-xs cursor-nwse-resize z-30 shadow-xs"
                  @pointerdown.stop="startResize($event, el)"
                />
              </div>
            </template>
          </div>

          <!-- Tombol Tambah Baris Cepat (di Luar Stiker) -->
          <div v-if="isTableMode" class="mt-3 flex items-center justify-center">
            <button
              type="button"
              class="px-3 py-1.5 bg-white hover:bg-indigo-50 text-indigo-700 hover:text-indigo-800 rounded-md border border-indigo-200 shadow-2xs text-[11px] font-bold transition-all flex items-center gap-1.5 cursor-pointer"
              @click="addNewRow('location', 'Ruang')"
            >
              <i class="pi pi-plus-circle text-xs" />
              <span>+ Tambah Baris Baru</span>
            </button>
          </div>
        </div>
      </div>

      <!-- RIGHT COLUMN: PROPERTIES INSPECTOR (FULL KUSTOM BARIS & SEL) -->
      <div class="w-88 shrink-0 border-l border-slate-200 bg-white p-3.5 flex flex-col gap-3 text-xs overflow-y-auto select-none">
        <div class="font-bold text-slate-700 uppercase tracking-wider text-[11px] flex items-center justify-between pb-2 border-b border-slate-200">
          <span>{{ isTableMode ? 'Inspektor Baris & Sel' : 'Pengaturan Elemen' }}</span>
          <Tag :value="isTableMode ? 'KUSTOM BARIS & SEL' : 'MODE BEBAS'" severity="info" class="!text-[9px] !py-0" />
        </div>

        <!-- 1. INSPECTOR UNTUK MODE TABEL -->
        <div v-if="isTableMode" class="flex flex-col gap-3">
          <!-- A. Pengaturan Baris Terpilih -->
          <div v-if="selectedTableRow" class="p-3 bg-indigo-50/60 rounded-lg border border-indigo-200 flex flex-col gap-2.5">
            <div class="flex items-center justify-between">
              <span class="font-bold text-indigo-900 text-xs">Baris #{{ (selectedTableRowIndex ?? 0) + 1 }}</span>
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

            <!-- Bagian Kelola Sel / Kolom di Baris Ini -->
            <div class="p-2 bg-white rounded border border-indigo-100 flex flex-col gap-2">
              <div class="flex items-center justify-between">
                <span class="text-[10.5px] font-bold text-slate-700 uppercase">
                  Kolom / Sel ({{ getRowCells(selectedTableRow).length }})
                </span>
                <div class="flex items-center gap-1">
                  <button
                    v-if="selectedTableRowIndex !== null && selectedTableRowIndex < activeTemplate.tableConfig.rows.length - 1"
                    type="button"
                    class="px-2 py-0.5 bg-emerald-50 hover:bg-emerald-600 hover:text-white text-emerald-800 font-bold text-[10px] rounded border border-emerald-200 transition-all flex items-center gap-1"
                    title="Gabung baris ini dengan baris di bawahnya"
                    @click="mergeRowWithNext(selectedTableRowIndex!)"
                  >
                    <i class="pi pi-link text-[9px]" />
                    <span>Merge Baris</span>
                  </button>
                  <button
                    type="button"
                    class="px-2 py-0.5 bg-indigo-50 hover:bg-indigo-600 hover:text-white text-indigo-700 font-bold text-[10px] rounded border border-indigo-200 transition-all flex items-center gap-1"
                    title="Tambah Sel / Kolom Baru di Baris Ini"
                    @click="addCellToRow(selectedTableRowIndex!)"
                  >
                    <i class="pi pi-plus text-[9px]" />
                    <span>+ Tambah Sel</span>
                  </button>
                </div>
              </div>

              <!-- Daftar Chips Sel di Baris Ini -->
              <div class="flex flex-wrap gap-1.5">
                <div
                  v-for="(c, cIdx) in getRowCells(selectedTableRow)"
                  :key="c.id"
                  class="flex items-center gap-1 px-2 py-1 rounded border text-[11px] cursor-pointer transition-all"
                  :class="selectedTableCellIndex === cIdx ? 'bg-emerald-500 border-emerald-600 text-white font-bold shadow-2xs' : 'bg-slate-50 border-slate-200 text-slate-700 hover:bg-slate-100'"
                  @click="selectedTableCellIndex = cIdx"
                >
                  <span>Sel {{ cIdx + 1 }}: {{ c.label || c.field }}</span>
                  <button
                    v-if="getRowCells(selectedTableRow).length > 1"
                    type="button"
                    class="ml-0.5 p-0.5 rounded-full hover:bg-black/20 text-current transition-colors"
                    title="Hapus Sel Ini"
                    @click.stop="removeCellFromRow(selectedTableRowIndex!, cIdx)"
                  >
                    <i class="pi pi-times text-[8px]" />
                  </button>
                </div>
              </div>
            </div>

            <!-- Detail Edit Sel Terpilih -->
            <div v-if="selectedTableCell" class="p-2.5 bg-white rounded border border-indigo-100 flex flex-col gap-2">
              <div class="flex items-center justify-between border-b border-slate-100 pb-1">
                <span class="font-bold text-emerald-800 text-[11px]">
                  Edit Sel #{{ (selectedTableCellIndex ?? 0) + 1 }}
                </span>
                <button
                  v-if="getRowCells(selectedTableRow).length > 1"
                  type="button"
                  class="text-rose-600 hover:text-rose-800 font-bold text-[10px] flex items-center gap-1"
                  @click="removeCellFromRow(selectedTableRowIndex!, selectedTableCellIndex!)"
                >
                  <i class="pi pi-trash text-[9px]" />
                  <span>Hapus Sel Ini</span>
                </button>
              </div>

              <div class="flex flex-col gap-1">
                <label class="text-[10px] font-semibold text-slate-600">Label Judul Sel:</label>
                <InputText v-model="selectedTableCell.label" placeholder="Contoh: Ruang" class="w-full !text-xs !py-1" />
              </div>

              <div class="flex flex-col gap-1">
                <label class="text-[10px] font-semibold text-slate-600">Field Data Barang:</label>
                <Select
                  v-model="selectedTableCell.field"
                  :options="AVAILABLE_FIELDS"
                  optionLabel="label"
                  optionValue="value"
                  class="w-full !text-xs"
                />
              </div>

              <div class="flex flex-col gap-1">
                <label class="text-[10px] font-semibold text-slate-600">Teks Manual (Opsional):</label>
                <InputText v-model="selectedTableCell.customText" placeholder="Kosongkan jika ambil otomatis dari field" class="w-full !text-xs !py-1" />
              </div>

              <div class="grid grid-cols-2 gap-2 pt-1 border-t border-slate-100">
                <div class="flex flex-col gap-1">
                  <label class="text-[10px] text-slate-600">Ukuran Font:</label>
                  <div class="flex items-center gap-1 bg-white border border-slate-300 rounded px-1.5 py-0.5">
                    <input
                      type="number"
                      v-model.number="selectedTableCell.fontSize"
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
                    v-model="selectedTableCell.bgColor"
                    :options="[
                      { label: 'Putih', value: 'white' },
                      { label: 'Abu-abu', value: 'light_gray' },
                      { label: 'Hitam (Inverted)', value: 'dark' },
                    ]"
                    optionLabel="label"
                    optionValue="value"
                    class="w-full !text-xs"
                  />
                </div>
              </div>

              <div class="flex items-center justify-between pt-1">
                <label class="flex items-center gap-1.5 cursor-pointer text-slate-700 text-xs">
                  <Checkbox v-model="selectedTableCell.bold" binary />
                  <span class="font-bold">Teks Tebal</span>
                </label>

                <div class="flex items-center gap-1">
                  <button
                    type="button"
                    class="px-2 py-0.5 rounded border text-[10.5px]"
                    :class="selectedTableCell.align === 'left' ? 'bg-indigo-600 text-white font-bold' : 'bg-white text-slate-600'"
                    @click="selectedTableCell.align = 'left'"
                  >
                    Kiri
                  </button>
                  <button
                    type="button"
                    class="px-2 py-0.5 rounded border text-[10.5px]"
                    :class="selectedTableCell.align === 'center' ? 'bg-indigo-600 text-white font-bold' : 'bg-white text-slate-600'"
                    @click="selectedTableCell.align = 'center'"
                  >
                    Tengah
                  </button>
                  <button
                    type="button"
                    class="px-2 py-0.5 rounded border text-[10.5px]"
                    :class="selectedTableCell.align === 'right' ? 'bg-indigo-600 text-white font-bold' : 'bg-white text-slate-600'"
                    @click="selectedTableCell.align = 'right'"
                  >
                    Kanan
                  </button>
                </div>
              </div>
              <!-- Tombol Cepat Aksi Sel (Pecah / Split & Gabung / Merge) -->
              <div class="flex items-center gap-1.5 pt-2 border-t border-slate-100 flex-wrap">
                <button
                  type="button"
                  class="flex-1 py-1 px-1.5 rounded bg-indigo-50 hover:bg-indigo-600 hover:text-white text-indigo-700 border border-indigo-200 text-[10px] font-bold flex items-center justify-center gap-1 transition-colors"
                  title="Pecah sel ini / tambah kolom baru di sebelahnya"
                  @click="splitCell(selectedTableRowIndex!, selectedTableCellIndex!)"
                >
                  <i class="pi pi-arrows-h text-xs" />
                  <span>Pecah (Split)</span>
                </button>
                <button
                  type="button"
                  :disabled="selectedTableCellIndex === null || selectedTableCellIndex >= getRowCells(selectedTableRow).length - 1"
                  class="flex-1 py-1 px-1.5 rounded bg-emerald-50 hover:bg-emerald-600 hover:text-white text-emerald-800 border border-emerald-200 text-[10px] font-bold flex items-center justify-center gap-1 disabled:opacity-40 transition-colors"
                  title="Gabung sel ini dengan sel sebelah kanan"
                  @click="mergeCellsWithNext(selectedTableRowIndex!, selectedTableCellIndex!)"
                >
                  <i class="pi pi-link text-xs" />
                  <span>Gabung (Merge)</span>
                </button>
              </div>
            </div>

            <!-- Tombol Aksi Hapus Baris -->
            <div class="pt-1 flex items-center justify-between">
              <label class="flex items-center gap-1.5 cursor-pointer text-slate-700 text-xs">
                <Checkbox v-model="selectedTableRow.hasBottomBorder" binary />
                <span>Paksa Garis Bawah</span>
              </label>

              <button
                type="button"
                class="px-2.5 py-1 rounded bg-rose-50 hover:bg-rose-600 hover:text-white text-rose-700 border border-rose-200 text-xs font-bold transition-all flex items-center gap-1"
                @click="removeTableRow(selectedTableRowIndex!)"
              >
                <i class="pi pi-trash text-xs" />
                <span>Hapus Seluruh Baris</span>
              </button>
            </div>
          </div>

          <div v-else class="p-2.5 bg-slate-50 rounded-lg border border-slate-200 text-slate-500 text-xs text-center">
            👉 Klik salah satu baris pada kanvas stiker untuk mengatur isi sel, menambah kolom, atau menghapusnya.
          </div>

          <!-- B. Daftar Urutan Baris Aktif -->
          <div class="flex flex-col gap-1.5">
            <div class="flex items-center justify-between">
              <span class="text-[11px] font-bold text-slate-700 uppercase tracking-wide">Urutan Baris ({{ activeTemplate.tableConfig.rows.length }})</span>
              <div class="flex items-center gap-2">
                <button
                  v-if="activeTemplate.tableConfig.rows.length > 0"
                  type="button"
                  class="text-rose-600 hover:text-rose-800 text-[10px] font-bold flex items-center gap-1"
                  title="Hapus seluruh baris stiker"
                  @click="clearAllRows"
                >
                  <i class="pi pi-trash text-[9px]" />
                  <span>Kosongkan</span>
                </button>
                <button
                  type="button"
                  class="text-indigo-600 hover:text-indigo-800 text-[11px] font-bold flex items-center gap-1"
                  @click="addNewRow('name', 'Nama')"
                >
                  <i class="pi pi-plus text-[10px]" />
                  <span>+ Tambah Baris</span>
                </button>
              </div>
            </div>

            <div class="flex flex-col gap-1 max-h-56 overflow-y-auto">
              <div
                v-for="(r, idx) in activeTemplate.tableConfig.rows"
                :key="r.id"
                class="flex items-center justify-between px-2 py-1 rounded border text-xs cursor-pointer transition-colors"
                :class="selectedTableRowIndex === idx ? 'bg-indigo-100 border-indigo-400 font-bold text-indigo-900' : 'bg-white border-slate-200 text-slate-700 hover:bg-slate-50'"
                @click="selectedTableRowIndex = idx; selectedTableCellIndex = 0"
              >
                <div class="flex items-center gap-1.5 truncate">
                  <span class="text-slate-400 text-[10px]">{{ idx + 1 }}.</span>
                  <span v-if="r.type === 'barcode_strip'" class="truncate text-cyan-700">|||| Barcode Garis</span>
                  <span v-else class="truncate">
                    {{ getRowCells(r).map((c) => c.label || c.field).join(' | ') }}
                    <span class="text-[10px] font-normal text-slate-400">({{ getRowCells(r).length }} sel)</span>
                  </span>
                </div>
                <div class="flex items-center gap-1 shrink-0">
                  <button
                    v-if="idx < activeTemplate.tableConfig.rows.length - 1"
                    type="button"
                    class="text-emerald-600 hover:text-emerald-800 p-0.5"
                    title="Gabung (Merge) dengan baris di bawahnya"
                    @click.stop="mergeRowWithNext(idx)"
                  >
                    <i class="pi pi-link text-[11px]" />
                  </button>
                  <button type="button" class="text-indigo-600 hover:text-indigo-800 p-0.5" title="Tambah Sel" @click.stop="addCellToRow(idx)">
                    <i class="pi pi-plus-circle text-[11px]" />
                  </button>
                  <button type="button" class="text-slate-400 hover:text-indigo-600 p-0.5" title="Duplikat" @click.stop="duplicateTableRow(idx)">
                    <i class="pi pi-copy text-[10px]" />
                  </button>
                  <button type="button" class="text-rose-500 hover:text-rose-700 p-0.5" title="Hapus Baris" @click.stop="removeTableRow(idx)">
                    <i class="pi pi-trash text-[10px]" />
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
        <!-- 1. Print Mode: Native HTML Table with Real Print Borders & Dynamic Cells -->
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
                borderBottom: activeTemplate.tableConfig.showHeaderBorder !== false
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
                padding: '0',
                verticalAlign: 'top',
              }"
            >
              <div style="display: flex; flex-direction: column; height: 100%; width: 100%; justify-content: space-between; box-sizing: border-box;">
                <!-- Lambang Kota / Logo Atas -->
                <div
                  v-if="activeTemplate.tableConfig.showLogo"
                  style="width: 100%; flex: 1; display: flex; align-items: center; justify-content: center; padding: 1mm 0.5mm; box-sizing: border-box;"
                  :style="{
                    borderBottom: `${activeTemplate.tableConfig.borderWidth}pt ${activeTemplate.tableConfig.borderStyle} ${activeTemplate.tableConfig.borderColor}`,
                  }"
                >
                  <img
                    v-if="activeTemplate.tableConfig.logoUrl"
                    :src="activeTemplate.tableConfig.logoUrl"
                    style="max-width: 95%; max-height: 12mm; object-fit: contain;"
                    alt="Logo"
                  />
                  <svg v-else viewBox="0 0 100 120" style="max-width: 90%; max-height: 12mm; object-fit: contain;">
                    <path d="M14 10 H86 V72 C86 98 50 114 50 114 C50 114 14 98 14 72 Z" fill="#ffffff" stroke="#1e293b" stroke-width="3" />
                    <path d="M16 12 H84 V56 C84 56 50 64 50 64 C50 64 16 56 16 56 Z" fill="#38bdf8" />
                    <polygon points="50,17 53,24 61,24 54,29 57,36 50,31 43,36 46,29 39,24 47,24" fill="#facc15" stroke="#ca8a04" stroke-width="0.8" />
                    <rect x="22" y="38" width="56" height="14" fill="#1e293b" />
                    <rect x="25" y="34" width="8" height="4" fill="#1e293b" />
                    <rect x="37" y="34" width="8" height="4" fill="#1e293b" />
                    <rect x="49" y="34" width="8" height="4" fill="#1e293b" />
                    <rect x="61" y="34" width="8" height="4" fill="#1e293b" />
                    <path d="M16 56 Q33 50 50 58 Q67 66 84 56 V72 C84 96 50 112 50 112 C50 112 16 96 16 72 Z" fill="#0284c7" />
                    <path d="M16 66 Q33 60 50 68 Q67 76 84 66 V76 C84 96 50 112 50 112 C50 112 16 96 16 76 Z" fill="#0369a1" />
                    <path d="M38 72 C32 68 36 60 48 62 C58 64 63 70 58 76 C52 82 42 80 38 74" fill="none" stroke="#f97316" stroke-width="3" stroke-linecap="round" />
                    <circle cx="56" cy="71" r="2.5" fill="#f97316" />
                    <path d="M22 96 Q50 88 78 96" fill="none" stroke="#ffffff" stroke-width="3.5" stroke-linecap="round" />
                    <text x="50" y="98" font-size="5" font-weight="900" text-anchor="middle" fill="#0f172a" font-family="sans-serif">CIREBON</text>
                  </svg>
                </div>

                <!-- QR Code di Bawah Lambang -->
                <div style="width: 100%; flex: 1; display: flex; align-items: center; justify-content: center; padding: 1mm 0.5mm; box-sizing: border-box;">
                  <img
                    v-if="activeTemplate.tableConfig.qrType === 'qr'"
                    :src="`/api/barcode/${it.id}.png?fmt=qr`"
                    class="print-qr-img"
                    style="width: 90%; height: 90%; object-fit: contain;"
                    alt="QR"
                  />
                  <div v-else class="print-barcode-box">
                    <img :src="`/api/barcode/${it.id}.png?fmt=code128`" class="print-barcode-img" alt="BC" />
                    <span class="print-barcode-sku">{{ it.sku }}</span>
                  </div>
                </div>

                <!-- Nama Barang di Bawah Kotak (- Nama) -->
                <div
                  v-if="activeTemplate.tableConfig.showItemNameUnderLogo"
                  style="width: 100%; font-size: 5pt; font-weight: 700; text-align: center; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; padding: 0.5mm 0; border-top: 0.5pt solid #cbd5e1;"
                >
                  - {{ it.name }}
                </div>
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
                    borderBottom: (activeTemplate.tableConfig.showGridLines && rIdx < activeTemplate.tableConfig.rows.length - 1) || r.hasBottomBorder
                      ? `${activeTemplate.tableConfig.borderWidth}pt ${activeTemplate.tableConfig.borderStyle} ${activeTemplate.tableConfig.borderColor}`
                      : 'none',
                  }"
                >
                  <!-- BARCODE STRIP ROW -->
                  <template v-if="r.type === 'barcode_strip'">
                    <td colspan="10" style="text-align: center; padding: 0.5mm 1mm;">
                      <img :src="`/api/barcode/${it.id}.png?fmt=code128`" style="width: 90%; height: 6mm; object-fit: contain;" alt="BC" />
                      <div style="font-family: monospace; font-weight: bold; font-size: 5.5pt; line-height: 1;">{{ it.sku }}</div>
                    </td>
                  </template>

                  <!-- DYNAMIC MULTI-CELL ROW -->
                  <template v-else>
                    <td
                      v-for="(cell, cIdx) in getRowCells(r)"
                      :key="cell.id || cIdx"
                      class="print-cell"
                      :style="{
                        fontSize: `${cell.fontSize || r.fontSize || activeTemplate.tableConfig.defaultFontSize}pt`,
                        borderRight: activeTemplate.tableConfig.showLabelColBorder && cIdx < getRowCells(r).length - 1
                          ? `${activeTemplate.tableConfig.borderWidth}pt ${activeTemplate.tableConfig.borderStyle} ${activeTemplate.tableConfig.borderColor}`
                          : 'none',
                        padding: `${activeTemplate.tableConfig.cellPaddingMm}mm 1mm`,
                        backgroundColor: cell.bgColor === 'light_gray' ? '#f1f5f9' : cell.bgColor === 'dark' ? '#1e293b' : 'transparent',
                        color: cell.bgColor === 'dark' ? '#ffffff' : '#000000',
                        textAlign: cell.align || 'left',
                      }"
                    >
                      <div style="display: flex; align-items: center; width: 100%; overflow: hidden;">
                        <b v-if="cell.label" style="padding-right: 1mm; shrink: 0; white-space: nowrap;">{{ cell.label }}:</b>
                        <span
                          :style="{
                            fontWeight: cell.bold ? 'bold' : 'normal',
                            fontFamily: cell.field === 'sku' ? 'monospace' : 'inherit',
                            overflow: 'hidden',
                            textOverflow: 'ellipsis',
                            whiteSpace: 'nowrap',
                          }"
                        >
                          {{ cell.customText || resolveFieldValue(it, cell.field) }}
                        </span>
                      </div>
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
                el.border ? 'border border-black' : '',
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
            <div
              v-else-if="el.type === 'line'"
              class="border-black my-auto"
              :class="[el.lineOrientation === 'vertical' ? 'h-full mx-auto border-l' : 'w-full border-t']"
            />
            <div v-else-if="el.type === 'logo'" class="w-full h-full flex items-center justify-center">
              <img v-if="el.logoUrl" :src="el.logoUrl" class="max-w-full max-h-full object-contain" alt="Logo" />
              <div v-else class="w-full h-full border border-dashed border-black flex items-center justify-center text-[5pt] text-black">Logo</div>
            </div>
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
  .print-cell {
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
