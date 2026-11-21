<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import Dialog from 'primevue/dialog'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Checkbox from 'primevue/checkbox'
import Tag from 'primevue/tag'
import { useToast } from 'primevue/usetoast'
import { useConfirm } from 'primevue/useconfirm'
import api from '@/api'
import {
  type SheetFormatId,
  type PrintMode,
  type LabelElement,
  type TableCellConfig,
  type TableRowConfig,
  type TableLayoutConfig,
  type LabelTemplate,
  SHEET_FORMAT_OPTIONS,
  PRINT_MODE_OPTIONS,
  inferSheetFormat,
  getSheetSpec,
  getPrintMode,
  getShowCutBorder,
  getRowCells,
  isLogoInHeader,
  ensureTableConfig,
  resolveFieldValue,
  printCustomLabels,
  openCustomPrintTab,
} from '@/utils/customLabelPrint'

import { AVAILABLE_FIELDS, DEFAULT_PRESETS } from '@/utils/labelPresets'
import LabelPrinterPanel from '@/components/LabelPrinterPanel.vue'

export type {
  SheetFormatId,
  LabelElement,
  TableCellConfig,
  TableRowConfig,
  TableLayoutConfig,
  LabelTemplate,
}

const props = withDefaults(
  defineProps<{
    visible: boolean
    items: any[]
    selectedIds: string[]
    initialTemplateId?: string
    initialCopies?: number
    initialOffset?: number
  }>(),
  {
    initialTemplateId: '',
    initialCopies: 1,
    initialOffset: 0,
  }
)

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
  (e: 'template-saved', template: LabelTemplate): void
}>()

const toast = useToast()
const confirm = useConfirm()
const MM_TO_PX = 3.7795
const zoom = ref(2.2)
const snapGrid = ref(true)
const gridSizeMm = ref(0.5)

const templates = ref<LabelTemplate[]>([])
const activeTemplate = ref<LabelTemplate>(JSON.parse(JSON.stringify(DEFAULT_PRESETS[0])))
const activeElementId = ref<string | null>(null)
const selectedTableRowIndex = ref<number | null>(0)
const selectedTableCellIndex = ref<number | null>(0)
const copies = ref(Math.max(1, props.initialCopies || 1))
const offset = ref(Math.max(0, props.initialOffset || 0))
const isPrinting = ref(false)
const paletteTab = ref<'data' | 'borders' | 'header_footer' | 'free_components' | 'printer'>('data')
const tableLogoFileInput = ref<HTMLInputElement | null>(null)
const isUploadingLogo = ref(false)
const saveTemplateDialog = ref(false)
const saveTemplateName = ref('')
const savingTemplate = ref(false)

function triggerTableLogoUpload() {
  tableLogoFileInput.value?.click()
}

async function onTableLogoSelected(e: Event) {
  const input = e.target as HTMLInputElement
  if (!input.files?.length) return
  const f = input.files[0]
  if (!f.type.startsWith('image/')) {
    toast.add({ severity: 'error', summary: 'Format File Salah', detail: 'Pilih file gambar (PNG, JPG, SVG, WebP)', life: 2500 })
    input.value = ''
    return
  }

  isUploadingLogo.value = true
  try {
    const formData = new FormData()
    formData.append('photo', f)
    const res = await api.post('/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })

    const uploadedUrl = res.data?.url
    if (!uploadedUrl) {
      throw new Error('Server tidak mengembalikan URL foto')
    }

    if (activeTemplate.value.tableConfig) {
      activeTemplate.value.tableConfig.logoUrl = uploadedUrl
      activeTemplate.value.tableConfig.showLogo = true
    }

    saveTemplates()

    // Jika sedang di template custom, langsung perbarui di database
    if (isCustomActive.value) {
      await api.post('/barcode/templates', activeTemplate.value)
      emit('template-saved', activeTemplate.value)
      toast.add({ severity: 'success', summary: 'Foto Logo Tersimpan di Database', detail: activeTemplate.value.name, life: 2500 })
    } else {
      toast.add({ severity: 'success', summary: 'Logo Berhasil Diunggah', detail: 'Klik "Simpan Template" untuk menyimpan perubahan secara permanen', life: 3500 })
    }
  } catch (err: any) {
    console.error('Upload logo gagal:', err)
    toast.add({
      severity: 'error',
      summary: 'Gagal Mengunggah Foto Logo',
      detail: err.response?.data?.error || err.message,
      life: 4000,
    })
  } finally {
    isUploadingLogo.value = false
    input.value = ''
  }
}

function resetTableLogo() {
  if (activeTemplate.value.tableConfig) {
    activeTemplate.value.tableConfig.logoUrl = ''
    saveTemplates()
    if (isCustomActive.value) {
      api.post('/barcode/templates', activeTemplate.value).catch(() => {})
    }
    toast.add({ severity: 'info', summary: 'Logo Direset ke Lambang Resmi', life: 1500 })
  }
}

function normalizeTemplateId(id?: string): string {
  return (id || '').replace(/^custom_/, '')
}

function findTemplateById(list: LabelTemplate[], targetId?: string): LabelTemplate | undefined {
  if (!targetId) return undefined
  const cleanTarget = normalizeTemplateId(targetId)
  return list.find((p) => p.id === targetId || normalizeTemplateId(p.id) === cleanTarget)
}

async function loadSavedTemplates(preferredId?: string) {
  let loadedList: LabelTemplate[] = []

  // 1. Coba muat dari backend database (PostgreSQL)
  try {
    const res = await api.get('/barcode/templates')
    if (res.data && Array.isArray(res.data) && res.data.length > 0) {
      loadedList = res.data.map((t: any) => {
        const item: LabelTemplate = {
          id: t.id,
          name: t.name,
          widthMm: t.widthMm || 64,
          heightMm: t.heightMm || 32,
          sheetFormat: t.sheetFormat || inferSheetFormat(t),
          config: t.config,
          layoutMode: t.layoutMode || 'table',
          tableConfig: t.tableConfig,
          elements: t.elements || [],
        }
        ensureTableConfig(item)
        return item
      })
      templates.value = loadedList
      try {
        localStorage.setItem('custom_label_templates_v6', JSON.stringify(loadedList))
      } catch {}
    }
  } catch (err) {
    console.warn('Gagal memuat template dari database, fallback ke localStorage:', err)
  }

  // 2. Fallback ke localStorage jika belum terisi dari server
  if (loadedList.length === 0) {
    const raw = localStorage.getItem('custom_label_templates_v6')
    if (raw) {
      try {
        const parsed = JSON.parse(raw)
        parsed.forEach((t: LabelTemplate) => {
          t.sheetFormat = t.sheetFormat || inferSheetFormat(t)
          ensureTableConfig(t)
        })
        templates.value = parsed
        loadedList = parsed
      } catch {
        templates.value = []
      }
    } else {
      templates.value = []
    }
  }

  // 3. Tentukan template yang harus aktif (activeTemplate)
  const targetId = preferredId || props.initialTemplateId
  if (targetId) {
    const match = findTemplateById(allPresets.value, targetId)
    if (match) {
      applyPreset(match)
      return
    }
  }

  // Cek apakah ada saved custom template dari database
  const customFromDB = loadedList.find((t) => !DEFAULT_PRESETS.some((dp) => dp.id === t.id))

  // Cek lastActive dari browser
  const lastActive = localStorage.getItem('active_custom_label_template_v6')
  if (lastActive) {
    try {
      const parsedActive = JSON.parse(lastActive)
      parsedActive.sheetFormat = parsedActive.sheetFormat || inferSheetFormat(parsedActive)
      ensureTableConfig(parsedActive)
      const stillExists = findTemplateById(allPresets.value, parsedActive.id)
      if (stillExists) {
        applyPreset(stillExists)
        return
      }
      activeTemplate.value = parsedActive
      return
    } catch {}
  }

  // Jika ada template custom dari DB, otomatis tampilkan template kustom tersebut
  if (customFromDB) {
    applyPreset(customFromDB)
    return
  }

  // Fallback ke preset 0
  applyPreset(DEFAULT_PRESETS[0])
}

function saveTemplates() {
  try {
    localStorage.setItem('custom_label_templates_v6', JSON.stringify(templates.value))
  } catch {}
}

const allPresets = computed(() => {
  const customList = templates.value.filter((t) => !DEFAULT_PRESETS.some((dp) => dp.id === t.id))
  return [...DEFAULT_PRESETS, ...customList]
})

const isCustomActive = computed(() => {
  return !DEFAULT_PRESETS.some((p) => p.id === activeTemplate.value.id)
})

const activeSheetFormat = computed<SheetFormatId>({
  get: () => activeTemplate.value.sheetFormat || inferSheetFormat(activeTemplate.value),
  set: (val) => {
    activeTemplate.value.sheetFormat = val
    const spec = getSheetSpec(val, activeTemplate.value.widthMm, activeTemplate.value.heightMm)
    if (val !== 'single' && val !== 'auto_a4') {
      activeTemplate.value.widthMm = spec.labelW
      activeTemplate.value.heightMm = spec.labelH
    }
    if (!spec.isSheet) {
      offset.value = 0
    } else if (offset.value >= spec.perSheet) {
      offset.value = Math.max(0, spec.perSheet - 1)
    }
  },
})

const activeSheetSpec = computed(() =>
  getSheetSpec(activeSheetFormat.value, activeTemplate.value.widthMm, activeTemplate.value.heightMm)
)

const activePrintMode = computed<PrintMode>({
  get: () => getPrintMode(activeTemplate.value),
  set: (val) => {
    if (!activeTemplate.value.config) activeTemplate.value.config = {}
    activeTemplate.value.config.printMode = val
  },
})

const activeShowCutBorder = computed<boolean>({
  get: () => getShowCutBorder(activeTemplate.value),
  set: (val) => {
    if (!activeTemplate.value.config) activeTemplate.value.config = {}
    activeTemplate.value.config.showCutBorder = val
  },
})

onMounted(() => {
  ensureTableConfig(activeTemplate.value)
  loadSavedTemplates(props.initialTemplateId)
})

watch(
  () => props.visible,
  (val) => {
    if (val) {
      copies.value = Math.max(1, props.initialCopies || 1)
      offset.value = Math.max(0, props.initialOffset || 0)
      loadSavedTemplates(props.initialTemplateId)
    }
  }
)

watch(
  () => props.initialTemplateId,
  (val) => {
    if (val) {
      const match = findTemplateById(allPresets.value, val)
      if (match) applyPreset(match)
    }
  }
)

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
      if (activeTemplate.value.tableConfig.logoPosition !== 'above_qr_only') {
        activeTemplate.value.tableConfig.logoPosition = 'header'
      }
    }
  }
)

watch(
  () => activeTemplate.value.tableConfig?.logoPosition,
  (pos) => {
    if (pos === 'header' && activeTemplate.value.tableConfig && !activeTemplate.value.tableConfig.showHeader) {
      activeTemplate.value.tableConfig.showHeader = true
      if (!activeTemplate.value.tableConfig.headerText?.trim()) {
        activeTemplate.value.tableConfig.headerText = 'PEMERINTAH KOTA CIREBON'
      }
      if (!activeTemplate.value.tableConfig.headerSubText?.trim()) {
        activeTemplate.value.tableConfig.headerSubText = 'PUSKESMAS KESAMBI'
      }
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
  confirm.require({
    header: 'Kosongkan Semua Baris',
    message: 'Yakin ingin menghapus seluruh baris tabel ini? Anda dapat menambah baris baru tanpa batasan.',
    icon: 'pi pi-exclamation-triangle',
    rejectLabel: 'Batal',
    acceptLabel: 'Hapus Semua',
    acceptClass: 'p-button-danger',
    accept: () => {
      activeTemplate.value.tableConfig.rows = []
      selectedTableRowIndex.value = null
      selectedTableCellIndex.value = null
      toast.add({ severity: 'warn', summary: 'Seluruh Baris Dihapus', life: 1500 })
    },
  })
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

function saveCurrentAsNewTemplate() {
  saveTemplateName.value = activeTemplate.value.name.replace(/^[🌟🏷️🏛️\s]+/, '').trim()
  saveTemplateDialog.value = true
}

async function confirmSaveNewTemplate() {
  const trimmed = saveTemplateName.value.trim()
  if (!trimmed) {
    toast.add({ severity: 'warn', summary: 'Nama template wajib diisi', life: 2000 })
    return
  }
  savingTemplate.value = true
  try {
    const newTpl: LabelTemplate = {
      ...JSON.parse(JSON.stringify(activeTemplate.value)),
      id: `custom_${Date.now().toString(36)}`,
      name: trimmed,
      sheetFormat: activeSheetFormat.value,
    }
    templates.value.push(newTpl)
    activeTemplate.value.id = newTpl.id
    activeTemplate.value.name = newTpl.name
    saveTemplates()
    emit('template-saved', newTpl)
    saveTemplateDialog.value = false

    // Sync ke database
    try {
      await api.post('/barcode/templates', newTpl)
      toast.add({ severity: 'success', summary: 'Tersimpan di Database', detail: newTpl.name, life: 2500 })
    } catch (err: any) {
      toast.add({ severity: 'warn', summary: 'Tersimpan Lokal', detail: 'Tersimpan di browser, gagal kirim ke database: ' + (err.response?.data?.error || err.message), life: 3500 })
    }
  } finally {
    savingTemplate.value = false
  }
}

async function updateCurrentCustomTemplate() {
  const idx = templates.value.findIndex((t) => t.id === activeTemplate.value.id)
  if (idx !== -1) {
    const updated: LabelTemplate = {
      ...JSON.parse(JSON.stringify(activeTemplate.value)),
      sheetFormat: activeSheetFormat.value,
    }
    templates.value[idx] = updated
    saveTemplates()
    emit('template-saved', updated)

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

function deleteCurrentCustomTemplate() {
  const target = activeTemplate.value
  confirm.require({
    header: 'Hapus Template Kustom',
    message: `Hapus template "${target.name}"?`,
    icon: 'pi pi-trash',
    rejectLabel: 'Batal',
    acceptLabel: 'Hapus',
    acceptClass: 'p-button-danger',
    accept: async () => {
      templates.value = templates.value.filter((t) => t.id !== target.id)
      saveTemplates()
      applyPreset(DEFAULT_PRESETS[0])
      emit('template-saved', activeTemplate.value)

      // Sync ke database
      try {
        await api.delete(`/barcode/templates/${target.id}`)
        toast.add({ severity: 'info', summary: 'Template Dihapus dari Database', life: 2000 })
      } catch {
        toast.add({ severity: 'info', summary: 'Template Dihapus Lokal', life: 1800 })
      }
    },
  })
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
const resolvedItemsToPrint = computed(() => {
  if (props.selectedIds.length === 0) {
    return [sampleItem.value]
  }
  const byId = new Map(props.items.map((i) => [i.id, i]))
  return props.selectedIds.map((id) => byId.get(id) || { id, name: id, sku: id })
})

function openSheetPreview(autoPrint = false) {
  const win = window.open('', '_blank')
  if (!win) {
    if (autoPrint) {
      // Fallback to hidden iframe when browser blocks popups
      void printCustomLabels(activeTemplate.value, resolvedItemsToPrint.value, {
        copies: copies.value,
        offset: offset.value,
        sheetFormat: activeSheetFormat.value,
        showCutBorder: activeShowCutBorder.value,
      })
      return
    }
    toast.add({
      severity: 'warn',
      summary: 'Popup Diblokir Browser',
      detail: 'Izinkan popup untuk membuka lembar cetak di tab baru.',
      life: 3500,
    })
    return
  }
  openCustomPrintTab(win, activeTemplate.value, resolvedItemsToPrint.value, {
    copies: copies.value,
    offset: offset.value,
    sheetFormat: activeSheetFormat.value,
    showCutBorder: activeShowCutBorder.value,
    autoPrint,
  })
}

async function printLabels() {
  if (activePrintMode.value === 'sheet_tab') {
    openSheetPreview(true)
    return
  }
  isPrinting.value = true
  try {
    await printCustomLabels(activeTemplate.value, resolvedItemsToPrint.value, {
      copies: copies.value,
      offset: offset.value,
      sheetFormat: activeSheetFormat.value,
      showCutBorder: activeShowCutBorder.value,
    })
  } catch (err: any) {
    toast.add({
      severity: 'error',
      summary: 'Gagal Mencetak Label',
      detail: err?.message || 'Tidak dapat menyiapkan dokumen cetak.',
      life: 3000,
    })
  } finally {
    isPrinting.value = false
  }
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
    <div class="px-3.5 py-2 border-b border-line bg-panel flex flex-wrap items-center justify-between gap-2.5 text-xs select-none">
      <!-- Left: Mode Switcher + Preset & Size -->
      <div class="flex items-center gap-2 flex-wrap">
        <!-- Layout Mode Toggle (Tabel vs Bebas) -->
        <div class="flex items-center bg-panel-2 p-0.5 rounded-lg border border-line shrink-0">
          <button
            type="button"
            class="px-2.5 py-1 rounded-md font-semibold text-xs flex items-center gap-1.5 transition-all"
            :class="isTableMode ? 'bg-acc-500 text-white shadow-xs' : 'text-txt-dim hover:text-txt'"
            @click="activeTemplate.layoutMode = 'table'"
          >
            <i class="pi pi-table" aria-hidden="true" />
            <span>Mode Tabel (Tambah &amp; Hapus Baris/Sel)</span>
          </button>
          <button
            type="button"
            class="px-2.5 py-1 rounded-md font-semibold text-xs flex items-center gap-1.5 transition-all"
            :class="!isTableMode ? 'bg-acc-500 text-white shadow-xs' : 'text-txt-dim hover:text-txt'"
            @click="activeTemplate.layoutMode = 'free'"
          >
            <i class="pi pi-th-large" aria-hidden="true" />
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

        <div class="h-4 w-px bg-line mx-0.5 shrink-0" />

        <!-- Preset Selector -->
        <div class="flex items-center gap-1.5 shrink-0">
          <span class="text-txt-dim font-semibold">Preset:</span>
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
            aria-label="Hapus template kustom ini"
            v-tooltip.top="'Hapus Template Kustom Ini'"
            class="!p-1 shrink-0"
            @click="deleteCurrentCustomTemplate"
          />
        </div>

        <!-- Format Kertas / Lembar -->
        <div class="flex items-center gap-1.5 shrink-0">
          <span class="text-txt-dim font-semibold">Kertas:</span>
          <Select
            v-model="activeSheetFormat"
            :options="SHEET_FORMAT_OPTIONS"
            optionLabel="label"
            optionValue="value"
            class="w-56 !text-xs"
          />
        </div>

        <!-- Dimensi Stiker (mm) -->
        <div class="flex items-center gap-1 bg-panel-2 px-2 py-0.5 rounded-md border border-line shrink-0">
          <div class="flex items-center gap-1">
            <span class="text-txt-muted t-mono text-[11px]">P:</span>
            <input
              type="number"
              v-model.number="activeTemplate.widthMm"
              min="20"
              max="297"
              aria-label="Lebar stiker dalam milimeter"
              class="w-13 text-center t-mono t-num font-bold text-xs bg-panel text-txt rounded border border-line py-0.5 px-1 focus:ring-1 focus:ring-acc-500 focus:outline-hidden"
              title="Lebar stiker (mm)"
            />
          </div>
          <span class="text-txt-muted font-bold">&times;</span>
          <div class="flex items-center gap-1">
            <span class="text-txt-muted t-mono text-[11px]">L:</span>
            <input
              type="number"
              v-model.number="activeTemplate.heightMm"
              min="15"
              max="297"
              aria-label="Tinggi stiker dalam milimeter"
              class="w-13 text-center t-mono t-num font-bold text-xs bg-panel text-txt rounded border border-line py-0.5 px-1 focus:ring-1 focus:ring-acc-500 focus:outline-hidden"
              title="Tinggi stiker (mm)"
            />
            <span class="text-txt-dim text-[11px] font-medium pl-0.5">mm</span>
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
        <div class="flex items-center bg-panel-2 rounded-md border border-line px-1 py-0.5 shrink-0">
          <button
            type="button"
            aria-label="Perkecil kanvas"
            class="px-2 py-0.5 text-txt-dim hover:text-txt font-bold"
            title="Perkecil Kanvas"
            @click="zoom = Math.max(1.2, zoom - 0.3)"
          >
            -
          </button>
          <span class="text-xs t-mono t-num font-semibold text-txt w-11 text-center">
            {{ Math.round(zoom * 45) }}%
          </span>
          <button
            type="button"
            aria-label="Perbesar kanvas"
            class="px-2 py-0.5 text-txt-dim hover:text-txt font-bold"
            title="Perbesar Kanvas"
            @click="zoom = Math.min(3.5, zoom + 0.3)"
          >
            +
          </button>
        </div>

        <!-- Snap Grid Toggle (Free Mode) -->
        <label v-if="!isTableMode" class="flex items-center gap-1.5 cursor-pointer text-txt-dim text-xs bg-panel-2 hover:bg-panel px-2 py-1 rounded-md border border-line shrink-0">
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
        <Button icon="pi pi-download" text rounded size="small" severity="secondary" aria-label="Ekspor JSON template" v-tooltip.top="'Ekspor JSON Template'" class="shrink-0" @click="exportJSON" />
        <label class="cursor-pointer p-1.5 rounded-md hover:bg-panel-2 text-txt-dim hover:text-txt shrink-0" aria-label="Impor file JSON template" v-tooltip.top="'Impor File JSON'">
          <i class="pi pi-upload text-sm" aria-hidden="true" />
          <input type="file" accept=".json" class="hidden" @change="importJSON" />
        </label>

        <span class="h-4 w-px bg-line mx-0.5 shrink-0" />

        <!-- Print Action: Copies + Offset + Cetak Button -->
        <div class="flex items-center gap-1.5 shrink-0">
          <div class="flex items-center gap-1 bg-panel-2 px-1.5 py-1 rounded-md border border-line" title="Jumlah rangkap cetak per barang">
            <span class="text-txt-dim text-[11px] font-medium">&times;</span>
            <input
              type="number"
              v-model.number="copies"
              min="1"
              max="50"
              aria-label="Jumlah rangkap cetak per barang"
              class="w-10 text-center t-num font-bold text-xs bg-panel text-txt rounded border border-line py-0.5 focus:ring-1 focus:ring-acc-500 focus:outline-hidden"
            />
          </div>
          <div
            v-if="activeSheetSpec.isSheet"
            class="flex items-center gap-1 bg-panel-2 px-1.5 py-1 rounded-md border border-line"
            title="Lewati slot stiker yang sudah terpakai di lembar pertama"
          >
            <span class="text-txt-dim text-[11px] font-medium">Offset:</span>
            <input
              type="number"
              v-model.number="offset"
              min="0"
              :max="Math.max(0, activeSheetSpec.perSheet - 1)"
              aria-label="Lewati slot stiker terpakai"
              class="w-10 text-center t-num font-bold text-xs bg-panel text-txt rounded border border-line py-0.5 focus:ring-1 focus:ring-acc-500 focus:outline-hidden"
            />
          </div>
          <Button
            label="Pratinjau Lembar"
            icon="pi pi-external-link"
            size="small"
            severity="secondary"
            outlined
            class="!text-xs !py-1.5 !px-2.5 font-semibold shrink-0"
            v-tooltip.top="'Buka tampilan lembar cetak T&J / A4 di tab baru'"
            @click="openSheetPreview(false)"
          />
          <Button
            label="Cetak Label Ini"
            icon="pi pi-print"
            size="small"
            :loading="isPrinting"
            class="!text-xs !py-1.5 !px-3 font-semibold shrink-0"
            @click="printLabels"
          />
        </div>
      </div>
    </div>

    <!-- MAIN WORKSPACE: 3 COLUMNS -->
    <div class="flex-1 flex overflow-hidden bg-canvas">
      <!-- LEFT COLUMN: PALETTE & TABLE ROW BUILDER -->
      <div class="w-84 border-r border-line bg-panel flex flex-col overflow-hidden select-none">
        <!-- Palette Tab Headers -->
        <div class="p-2 border-b border-line bg-panel-2 flex items-center gap-1 overflow-x-auto">
          <button
            type="button"
            class="tab-btn"
            :class="paletteTab === 'data' ? 'tab-btn-active' : ''"
            @click="paletteTab = 'data'"
          >
            <i class="pi pi-list text-xs mr-1" aria-hidden="true" />
            <span>{{ isTableMode ? 'Data, Baris & Sel' : 'Data Form' }}</span>
          </button>
          <button
            v-if="isTableMode"
            type="button"
            class="tab-btn"
            :class="paletteTab === 'borders' ? 'tab-btn-active' : ''"
            @click="paletteTab = 'borders'"
          >
            <i class="pi pi-th-large text-xs mr-1" aria-hidden="true" />
            <span>Garis &amp; Jarak</span>
          </button>
          <button
            v-if="isTableMode"
            type="button"
            class="tab-btn"
            :class="paletteTab === 'header_footer' ? 'tab-btn-active' : ''"
            @click="paletteTab = 'header_footer'"
          >
            <i class="pi pi-building text-xs mr-1" aria-hidden="true" />
            <span>Kop &amp; Barcode</span>
          </button>
          <button
            v-if="!isTableMode"
            type="button"
            class="tab-btn"
            :class="paletteTab === 'free_components' ? 'tab-btn-active' : ''"
            @click="paletteTab = 'free_components'"
          >
            <i class="pi pi-plus text-xs mr-1" aria-hidden="true" />
            <span>Komponen</span>
          </button>
          <button
            type="button"
            class="tab-btn"
            :class="paletteTab === 'printer' ? 'tab-btn-active' : ''"
            @click="paletteTab = 'printer'"
          >
            <i class="pi pi-print text-xs mr-1" aria-hidden="true" />
            <span>Printer &amp; Lembar</span>
          </button>
        </div>

        <!-- 1. TAB DATA & BARIS (MODE TABEL - FITUR TAMBAH & HAPUS BARIS/SEL) -->
        <div v-if="paletteTab === 'data' && isTableMode" class="flex-1 overflow-y-auto p-2.5 flex flex-col gap-2.5">
          <!-- A. KONTROL CEPAT TAMBAH & HAPUS BARIS / SEL -->
          <div class="flex flex-col gap-1.5 p-2 bg-[var(--acc-wash)] rounded-lg border border-[var(--acc-border)]">
            <span class="t-label text-txt flex items-center justify-between">
              <span>Aksi Baris &amp; Sel:</span>
              <span class="text-[10px] text-acc-500 t-num font-semibold">Total {{ activeTemplate.tableConfig.rows.length }} Baris</span>
            </span>

            <div class="grid grid-cols-2 gap-1.5 pt-0.5">
              <!-- Tambah Baris Baru -->
              <button
                type="button"
                class="px-2 py-1.5 bg-acc-500 hover:bg-acc-600 text-white rounded text-left flex items-center gap-1.5 text-[11px] font-semibold transition-all"
                @click="addNewRow('name', 'Nama')"
              >
                <i class="pi pi-plus-circle text-xs" aria-hidden="true" />
                <span>+ Baris Baru</span>
              </button>

              <!-- Tambah Sel di Baris Terpilih -->
              <button
                type="button"
                class="px-2 py-1.5 bg-panel border border-[var(--acc-border)] hover:border-acc-500 rounded text-left flex items-center gap-1.5 text-[11px] font-semibold text-txt transition-all disabled:opacity-50"
                :disabled="selectedTableRowIndex === null"
                @click="selectedTableRowIndex !== null && addCellToRow(selectedTableRowIndex)"
              >
                <i class="pi pi-table text-acc-500 text-xs" aria-hidden="true" />
                <span>+ Sel / Kolom</span>
              </button>

              <!-- Hapus Baris Terpilih -->
              <button
                type="button"
                class="px-2 py-1.5 bg-panel border border-[var(--sig-bad-line)] hover:bg-[var(--sig-bad-bg)] text-[var(--sig-bad)] rounded text-left flex items-center gap-1.5 text-[11px] font-semibold transition-all disabled:opacity-50"
                :disabled="selectedTableRowIndex === null"
                @click="selectedTableRowIndex !== null && removeTableRow(selectedTableRowIndex)"
              >
                <i class="pi pi-trash text-xs" aria-hidden="true" />
                <span>Hapus Baris</span>
              </button>

              <!-- Hapus Sel Terpilih -->
              <button
                type="button"
                class="px-2 py-1.5 bg-panel border border-[var(--sig-warn-line)] hover:bg-[var(--sig-warn-bg)] text-[var(--sig-warn)] rounded text-left flex items-center gap-1.5 text-[11px] font-semibold transition-all disabled:opacity-50"
                :disabled="selectedTableRowIndex === null || selectedTableCellIndex === null"
                @click="selectedTableRowIndex !== null && selectedTableCellIndex !== null && removeCellFromRow(selectedTableRowIndex, selectedTableCellIndex)"
              >
                <i class="pi pi-times-circle text-xs" aria-hidden="true" />
                <span>Hapus Sel</span>
              </button>
            </div>

            <!-- Shortcut Jenis Baris Multi-Kolom -->
            <div class="grid grid-cols-2 gap-1.5 pt-1 border-t border-[var(--acc-border)]">
              <button
                type="button"
                class="px-2 py-1 bg-panel border border-line hover:border-acc-500 rounded text-left flex items-center gap-1.5 text-[10.5px] font-semibold text-txt"
                @click="addMultiCellRow([{ field: 'location', label: 'Ruang' }, { field: 'procurement_year', label: 'Tahun' }])"
              >
                <i class="pi pi-pause text-acc-500 text-[10px]" aria-hidden="true" />
                <span>Baris 2 Sel</span>
              </button>
              <button
                type="button"
                class="px-2 py-1 bg-panel border border-line hover:border-acc-500 rounded text-left flex items-center gap-1.5 text-[10.5px] font-semibold text-txt"
                @click="addMultiCellRow([{ field: 'location', label: 'Ruang' }, { field: 'procurement_year', label: 'Tahun' }, { field: 'condition_status', label: 'Kondisi' }])"
              >
                <i class="pi pi-bars text-acc-500 text-[10px]" aria-hidden="true" />
                <span>Baris 3 Sel</span>
              </button>
              <button
                type="button"
                class="px-2 py-1 bg-panel border border-line hover:border-acc-500 rounded text-left flex items-center gap-1.5 text-[10.5px] font-semibold text-txt"
                @click="addCustomTextRow"
              >
                <i class="pi pi-pencil text-acc-500 text-[10px]" aria-hidden="true" />
                <span>Teks Catatan</span>
              </button>
              <button
                type="button"
                class="px-2 py-1 bg-panel border border-line hover:border-acc-500 rounded text-left flex items-center gap-1.5 text-[10.5px] font-semibold text-txt"
                @click="addBarcodeStripRow"
              >
                <i class="pi pi-barcode text-[var(--sig-info)] text-[10px]" aria-hidden="true" />
                <span>Barcode 128 Strip</span>
              </button>
            </div>
          </div>

          <!-- B. DAFTAR 17 FIELD FORMULIR BARANG -->
          <div class="flex flex-col gap-1">
            <div class="text-[11px] text-txt-dim px-1 pb-0.5 font-medium flex items-center justify-between">
              <span>Field Formulir (Tambah Baris / Sel):</span>
              <span class="text-[10px] text-acc-500 t-num font-bold">{{ AVAILABLE_FIELDS.length }} Field</span>
            </div>

            <div
              v-for="f in AVAILABLE_FIELDS"
              :key="f.value"
              class="flex items-center justify-between p-1.5 bg-panel border border-line hover:border-acc-500 rounded-md transition-all gap-1"
            >
              <div class="flex items-center gap-1.5 truncate flex-1 min-w-0">
                <i :class="[f.icon, 'text-txt-dim text-xs shrink-0']" aria-hidden="true" />
                <span class="truncate text-[11px] font-semibold text-txt">{{ f.label }}</span>
              </div>
              <div class="flex items-center gap-1 shrink-0">
                <!-- Tambah sebagai Baris Baru -->
                <button
                  type="button"
                  class="px-1.5 py-0.5 rounded bg-[var(--acc-50)] hover:bg-acc-500 hover:text-white text-acc-500 border border-[var(--acc-border)] text-[10px] font-bold transition-all"
                  title="Tambah sebagai Baris Baru"
                  @click="addNewRow(f.value, f.defaultLabel)"
                >
                  + Baris
                </button>
                <!-- Tambah sebagai Sel ke Baris Terpilih -->
                <button
                  type="button"
                  class="px-1.5 py-0.5 rounded bg-panel-2 hover:bg-[var(--sig-ok)] hover:text-white text-txt-dim border border-line text-[10px] font-bold transition-all disabled:opacity-40"
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

        <!-- 1B. TAB DATA FORM (MODE BEBAS) -->
        <div v-else-if="paletteTab === 'data' && !isTableMode" class="flex-1 overflow-y-auto p-2.5 flex flex-col gap-1.5">
          <div class="text-[11px] text-txt-dim px-1 pb-0.5 font-medium flex items-center justify-between">
            <span>Klik untuk tambah ke kanvas bebas:</span>
            <span class="text-[10px] text-acc-500 t-num font-bold">{{ AVAILABLE_FIELDS.length }} Field</span>
          </div>
          <button
            v-for="f in AVAILABLE_FIELDS"
            :key="f.value"
            type="button"
            class="component-item-btn text-left"
            @click="addFieldElement(f.value, f.defaultLabel ? `${f.defaultLabel}: ` : '')"
          >
            <i :class="[f.icon, 'text-acc-500 text-sm shrink-0']" aria-hidden="true" />
            <span class="truncate text-[11.5px] font-semibold text-txt flex-1">{{ f.label }}</span>
            <i class="pi pi-plus text-txt-muted text-xs" aria-hidden="true" />
          </button>
        </div>

        <!-- 2. TAB GARIS & JARAK (MODE TABEL - KUSTOM GARIS FULL) -->
        <div v-else-if="paletteTab === 'borders' && isTableMode" class="flex-1 overflow-y-auto p-3 flex flex-col gap-3">
          <!-- A. Ketebalan & Bentuk Garis Global -->
          <div class="flex flex-col gap-2 p-2.5 bg-panel-2 rounded-lg border border-line">
            <span class="t-label text-txt">Ketebalan &amp; Jenis Garis</span>

            <div class="grid grid-cols-2 gap-2">
              <div class="flex flex-col gap-1">
                <label class="text-[10.5px] text-txt-dim font-semibold">Tebal Garis:</label>
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
                <label class="text-[10.5px] text-txt-dim font-semibold">Jenis Garis:</label>
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
            <div class="flex items-center justify-between pt-1 border-t border-line">
              <span class="text-[11px] text-txt-dim font-semibold">Warna Garis:</span>
              <div class="flex items-center gap-1.5">
                <button
                  type="button"
                  aria-label="Warna garis hitam pekat"
                  class="w-5 h-5 rounded-full border border-line"
                  :class="activeTemplate.tableConfig.borderColor === '#000000' ? 'ring-2 ring-acc-500 ring-offset-1' : ''"
                  style="background-color: #000000"
                  @click="activeTemplate.tableConfig.borderColor = '#000000'"
                />
                <button
                  type="button"
                  aria-label="Warna garis abu gelap"
                  class="w-5 h-5 rounded-full border border-line"
                  :class="activeTemplate.tableConfig.borderColor === '#334155' ? 'ring-2 ring-acc-500 ring-offset-1' : ''"
                  style="background-color: #334155"
                  @click="activeTemplate.tableConfig.borderColor = '#334155'"
                />
                <button
                  type="button"
                  aria-label="Warna garis abu sedang"
                  class="w-5 h-5 rounded-full border border-line"
                  :class="activeTemplate.tableConfig.borderColor === '#64748b' ? 'ring-2 ring-acc-500 ring-offset-1' : ''"
                  style="background-color: #64748b"
                  @click="activeTemplate.tableConfig.borderColor = '#64748b'"
                />
              </div>
            </div>
          </div>

          <!-- B. Toggle Garis Pembatas (Sisi-Sisi) -->
          <div class="flex flex-col gap-1.5 p-2.5 bg-panel-2 rounded-lg border border-line">
            <span class="t-label text-txt">Tampilkan Garis Pembatas</span>

            <label class="flex items-center gap-2 cursor-pointer py-1 text-txt text-xs">
              <Checkbox v-model="activeTemplate.tableConfig.showOuterBorder" binary />
              <span class="font-semibold">Bingkai Luar Stiker (Outer Box)</span>
            </label>

            <label class="flex items-center gap-2 cursor-pointer py-1 text-txt text-xs">
              <Checkbox v-model="activeTemplate.tableConfig.showGridLines" binary />
              <span class="font-semibold">Garis Horizontal Antar Baris</span>
            </label>

            <label class="flex items-center gap-2 cursor-pointer py-1 text-txt text-xs">
              <Checkbox v-model="activeTemplate.tableConfig.showLabelColBorder" binary />
              <span class="font-semibold">Garis Vertikal Pemisah Label &amp; Sel</span>
            </label>

            <label class="flex items-center gap-2 cursor-pointer py-1 text-txt text-xs">
              <Checkbox v-model="activeTemplate.tableConfig.showColBorder" binary />
              <span class="font-semibold">Garis Vertikal Pemisah Kolom QR</span>
            </label>

            <label class="flex items-center gap-2 cursor-pointer py-1 text-txt text-xs">
              <Checkbox v-model="activeTemplate.tableConfig.showHeaderBorder" binary />
              <span class="font-semibold">Garis Bawah Kop / Header</span>
            </label>

            <label class="flex items-center gap-2 cursor-pointer py-1 text-txt text-xs">
              <Checkbox v-model="activeTemplate.tableConfig.showFooterBorder" binary />
              <span class="font-semibold">Garis Atas Footer Catatan</span>
            </label>
          </div>

          <!-- C. Jarak / Padding & Proporsi Kolom -->
          <div class="flex flex-col gap-2 p-2.5 bg-panel-2 rounded-lg border border-line">
            <span class="t-label text-txt">Kerapatan &amp; Lebar Kolom</span>

            <div class="flex items-center justify-between">
              <div class="flex flex-col">
                <span class="text-[11px] font-semibold text-txt">Kerapatan Baris (Padding):</span>
                <span class="text-[9.5px] text-txt-muted">Jarak teks ke garis stiker</span>
              </div>
              <div class="flex items-center gap-1 bg-panel border border-line rounded px-1.5 py-0.5">
                <input
                  type="number"
                  v-model.number="activeTemplate.tableConfig.cellPaddingMm"
                  min="0.3"
                  max="3"
                  step="0.1"
                  aria-label="Kerapatan baris dalam milimeter"
                  class="w-11 text-center t-num font-bold text-xs text-txt bg-transparent focus:outline-hidden"
                />
                <span class="text-[10px] text-txt-muted">mm</span>
              </div>
            </div>

            <div class="flex items-center justify-between pt-1 border-t border-line">
              <div class="flex flex-col">
                <span class="text-[11px] font-semibold text-txt">Lebar Kolom Label:</span>
                <span class="text-[9.5px] text-txt-muted">Persentase lebar nama field</span>
              </div>
              <div class="flex items-center gap-1 bg-panel border border-line rounded px-1.5 py-0.5">
                <input
                  type="number"
                  v-model.number="activeTemplate.tableConfig.labelColWidthPercent"
                  min="18"
                  max="60"
                  step="1"
                  aria-label="Lebar kolom label dalam persen"
                  class="w-10 text-center t-num font-bold text-xs text-txt bg-transparent focus:outline-hidden"
                />
                <span class="text-[10px] text-txt-muted">%</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 3. TAB KOP, BARCODE & FOOTER (MODE TABEL) -->
        <div v-else-if="paletteTab === 'header_footer' && isTableMode" class="flex-1 overflow-y-auto p-3 flex flex-col gap-3">
          <!-- A. Pengaturan Barcode / QR Kolom -->
          <div class="flex flex-col gap-2 p-2.5 bg-panel-2 rounded-lg border border-line">
            <span class="t-label text-txt">Kolom Barcode &amp; QR</span>

            <div class="flex flex-col gap-1">
              <label class="text-[11px] font-semibold text-txt-dim">Posisi Barcode</label>
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
                <label class="text-[11px] font-semibold text-txt-dim">Tipe Barcode</label>
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

              <div class="flex items-center justify-between pt-1 border-t border-line">
                <span class="text-[11px] text-txt-dim font-medium">Lebar Kolom QR:</span>
                <div class="flex items-center gap-1 bg-panel border border-line rounded px-1.5 py-0.5">
                  <input
                    type="number"
                    v-model.number="activeTemplate.tableConfig.qrWidthMm"
                    min="15"
                    max="50"
                    aria-label="Lebar kolom QR dalam milimeter"
                    class="w-11 text-center t-num font-bold text-xs text-txt bg-transparent focus:outline-hidden"
                  />
                  <span class="text-[10px] text-txt-muted">mm</span>
                </div>
              </div>

              <!-- Lambang Kota / Logo Instansi -->
              <div class="pt-2 border-t border-line flex flex-col gap-2">
                <label class="flex items-center gap-1.5 cursor-pointer text-txt font-bold text-xs">
                  <Checkbox v-model="activeTemplate.tableConfig.showLogo" binary />
                  <span>Tampilkan Lambang Kota Cirebon / Logo</span>
                </label>

                <div v-if="activeTemplate.tableConfig.showLogo" class="flex flex-col gap-2 pl-4 border-l-2 border-[var(--acc-border)]">
                  <div class="flex flex-col gap-1">
                    <label class="text-[10px] text-txt-dim font-semibold">Posisi Lambang / Logo:</label>
                    <Select
                      v-model="activeTemplate.tableConfig.logoPosition"
                      :options="[
                        { label: 'Di Sebelah Tulisan Kop / Header (Barcode Besar)', value: 'header' },
                        { label: 'Di Atas Barcode / QR (Kolom Samping)', value: 'above_qr_only' },
                      ]"
                      optionLabel="label"
                      optionValue="value"
                      class="w-full !text-xs"
                    />
                  </div>
                  <input type="file" ref="tableLogoFileInput" accept="image/*" class="hidden" @change="onTableLogoSelected" />
                  <div class="flex items-center gap-1.5">
                    <Button
                      type="button"
                      icon="pi pi-upload"
                      :label="isUploadingLogo ? 'Mengunggah...' : 'Unggah / Ganti Foto Logo'"
                      :loading="isUploadingLogo"
                      :disabled="isUploadingLogo"
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
                    <label class="text-[10px] text-txt-dim font-semibold">Atau Masukkan URL Logo:</label>
                    <InputText v-model="activeTemplate.tableConfig.logoUrl" placeholder="Kosongkan untuk Lambang Kota Cirebon" class="w-full !text-xs !py-1" />
                  </div>
                </div>

                <label class="flex items-center gap-1.5 cursor-pointer text-txt text-xs pt-0.5">
                  <Checkbox v-model="activeTemplate.tableConfig.showItemNameUnderLogo" binary />
                  <span>Tampilkan Nama Barang di Bawah QR (- Nama)</span>
                </label>
              </div>
            </div>
          </div>

          <!-- B. Kop / Header Instansi -->
          <div class="flex flex-col gap-2 p-2.5 bg-panel-2 rounded-lg border border-line">
            <div class="flex items-center justify-between">
              <label class="flex items-center gap-1.5 cursor-pointer text-txt font-bold text-xs">
                <Checkbox v-model="activeTemplate.tableConfig.showHeader" binary />
                <span>Kop / Header Instansi</span>
              </label>
              <label v-if="activeTemplate.tableConfig.showHeader" class="flex items-center gap-1.5 cursor-pointer text-acc-500 text-xs font-semibold">
                <Checkbox v-model="activeTemplate.tableConfig.showHeaderBorder" binary />
                <span>Garis Bawah Kop</span>
              </label>
            </div>

            <div v-if="activeTemplate.tableConfig.showHeader" class="flex flex-col gap-2 pt-1 border-t border-line">
              <div class="flex flex-col gap-1">
                <label class="text-[10px] text-txt-dim font-semibold">Teks Baris 1 (Utama):</label>
                <InputText v-model="activeTemplate.tableConfig.headerText" placeholder="PEMERINTAH KABUPATEN..." class="w-full !text-xs !py-1" />
              </div>

              <div class="flex flex-col gap-1">
                <label class="text-[10px] text-txt-dim font-semibold">Teks Baris 2 (Sub-Kop / Dinas):</label>
                <InputText v-model="activeTemplate.tableConfig.headerSubText" placeholder="DINAS KESEHATAN..." class="w-full !text-xs !py-1" />
              </div>

              <div class="grid grid-cols-2 gap-2">
                <div class="flex flex-col gap-1">
                  <label class="text-[10px] text-txt-dim font-semibold">Gaya Kop:</label>
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
                  <label class="text-[10px] text-txt-dim font-semibold">Ukuran Font Kop:</label>
                  <div class="flex items-center gap-1 bg-panel border border-line rounded px-1.5 py-0.5">
                    <input
                      type="number"
                      v-model.number="activeTemplate.tableConfig.headerFontSize"
                      min="5"
                      max="14"
                      step="0.5"
                      aria-label="Ukuran font kop dalam poin"
                      class="w-full text-center t-num font-bold text-xs text-txt bg-transparent focus:outline-hidden"
                    />
                    <span class="text-[10px] text-txt-muted">pt</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- C. Catatan Kaki / Footer -->
          <div class="flex flex-col gap-2 p-2.5 bg-panel-2 rounded-lg border border-line">
            <div class="flex items-center justify-between">
              <label class="flex items-center gap-1.5 cursor-pointer text-txt font-bold text-xs">
                <Checkbox v-model="activeTemplate.tableConfig.showFooter" binary />
                <span>Catatan Kaki (Footer)</span>
              </label>
            </div>

            <div v-if="activeTemplate.tableConfig.showFooter" class="flex flex-col gap-1.5 pt-1 border-t border-line">
              <InputText v-model="activeTemplate.tableConfig.footerText" placeholder="DILARANG DILEPAS / DIPINDAHKAN..." class="w-full !text-xs !py-1" />
              <div class="flex items-center justify-between">
                <span class="text-[11px] text-txt-dim">Ukuran Huruf Footer:</span>
                <div class="flex items-center gap-1 bg-panel border border-line rounded px-1.5 py-0.5">
                  <input
                    type="number"
                    v-model.number="activeTemplate.tableConfig.footerFontSize"
                    min="4"
                    max="10"
                    step="0.5"
                    aria-label="Ukuran huruf footer dalam poin"
                    class="w-11 text-center t-num font-bold text-xs text-txt bg-transparent focus:outline-hidden"
                  />
                  <span class="text-[10px] text-txt-muted">pt</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 4. TAB KOMPONEN BEBAS (MODE BEBAS) -->
        <div v-else class="flex-1 overflow-y-auto p-2.5 flex flex-col gap-2">
          <button type="button" class="component-item-btn" @click="addOtherElement('qr')">
            <i class="pi pi-qrcode text-acc-500 text-lg" aria-hidden="true" />
            <div class="text-left flex-1 min-w-0">
              <div class="font-bold text-txt text-[11.5px]">QR Code</div>
              <div class="text-[10px] text-txt-dim">Scan link aset interaktif</div>
            </div>
            <i class="pi pi-plus text-txt-muted text-xs" aria-hidden="true" />
          </button>

          <button type="button" class="component-item-btn" @click="addOtherElement('code128')">
            <i class="pi pi-barcode text-[var(--sig-info)] text-lg" aria-hidden="true" />
            <div class="text-left flex-1 min-w-0">
              <div class="font-bold text-txt text-[11.5px]">Barcode 128</div>
              <div class="text-[10px] text-txt-dim">Barcode garis SKU aset</div>
            </div>
            <i class="pi pi-plus text-txt-muted text-xs" aria-hidden="true" />
          </button>

          <button type="button" class="component-item-btn" @click="addOtherElement('header')">
            <i class="pi pi-building text-[var(--sig-warn)] text-lg" aria-hidden="true" />
            <div class="text-left flex-1 min-w-0">
              <div class="font-bold text-txt text-[11.5px]">Header / Kop Instansi</div>
              <div class="text-[10px] text-txt-dim">Judul kop instansi atas</div>
            </div>
            <i class="pi pi-plus text-txt-muted text-xs" aria-hidden="true" />
          </button>

          <button type="button" class="component-item-btn" @click="addOtherElement('line')">
            <i class="pi pi-minus text-txt-dim text-lg" aria-hidden="true" />
            <div class="text-left flex-1 min-w-0">
              <div class="font-bold text-txt text-[11.5px]">Garis Pembatas</div>
              <div class="text-[10px] text-txt-dim">Garis pemisah horizontal</div>
            </div>
            <i class="pi pi-plus text-txt-muted text-xs" aria-hidden="true" />
          </button>

          <button type="button" class="component-item-btn" @click="addOtherElement('static_text')">
            <i class="pi pi-pencil text-acc-500 text-lg" aria-hidden="true" />
            <div class="text-left flex-1 min-w-0">
              <div class="font-bold text-txt text-[11.5px]">Teks Bebas Statis</div>
              <div class="text-[10px] text-txt-dim">Teks peringatan / keterangan</div>
            </div>
            <i class="pi pi-plus text-txt-muted text-xs" aria-hidden="true" />
          </button>
        </div>

        <!-- 5. TAB PRINTER & LEMBAR CETAK (MODE TABEL & BEBAS) -->
        <!-- 5. TAB PRINTER & LEMBAR CETAK (MODE TABEL & BEBAS) -->
        <LabelPrinterPanel
          v-if="paletteTab === 'printer'"
          v-model:sheet-format="activeSheetFormat"
          v-model:print-mode="activePrintMode"
          v-model:show-cut-border="activeShowCutBorder"
          :sheet-spec="activeSheetSpec"
          @preview="openSheetPreview(false)"
          @print="printLabels"
        />

        <div class="p-2.5 border-t border-line bg-panel-2 text-[11px] text-txt-dim leading-tight flex items-start gap-1.5">
          <i class="pi pi-info-circle text-acc-500 text-xs mt-0.5 shrink-0" aria-hidden="true" />
          <span v-if="isTableMode"><b>Mode Tabel Kustom:</b> Tambah atau hapus baris &amp; sel langsung melalui tombol hover di kanvas atau panel samping.</span>
          <span v-else>Geser elemen langsung di stiker. Tarik titik di pojok untuk ubah ukuran.</span>
        </div>
      </div>

      <!-- CENTER COLUMN: VISUAL STICKER CANVAS -->
      <div
        class="flex-1 overflow-auto p-8 grid place-items-center relative bg-canvas"
        @pointermove="onPointerMove"
        @pointerup="onPointerUp"
      >
        <div class="flex flex-col items-center gap-2">
          <!-- Sticker Label Badge Info + Quick Row Adder -->
          <div class="flex items-center gap-2 text-xs font-semibold text-txt-dim bg-panel px-3 py-1 rounded-full border border-line shadow-xs">
            <span class="inline-flex items-center gap-1.5 t-num">
              <i class="pi pi-tag text-acc-500 text-[11px]" aria-hidden="true" />
              {{ activeTemplate.widthMm }} &times; {{ activeTemplate.heightMm }} mm
            </span>
            <span class="text-line">|</span>
            <span class="text-acc-500 font-bold">{{ isTableMode ? 'Tabel Kustom Penuh' : 'Mode Bebas' }}</span>
            <span class="text-line">|</span>
            <span class="text-txt-dim font-normal">Sampel: {{ sampleItem.name }}</span>
            <template v-if="isTableMode">
              <span class="text-line">|</span>
              <button
                type="button"
                class="px-2 py-0.5 rounded bg-acc-500 hover:bg-acc-600 text-white font-semibold text-[11px] flex items-center gap-1 transition-all"
                @click="addNewRow('name', 'Nama')"
              >
                <i class="pi pi-plus text-[10px]" aria-hidden="true" />
                <span>Tambah Baris</span>
              </button>
            </template>
          </div>

          <!-- THE PHYSICAL STICKER (wrapped in Tom & Jerry yellow backing pad when active) -->
          <div
            :class="
              activeSheetSpec.isYellowBacking
                ? 'p-4 rounded-[10px] border border-amber-300 shadow-md flex flex-col items-center gap-2'
                : 'contents'
            "
            :style="activeSheetSpec.isYellowBacking ? { background: '#fef08a' } : undefined"
          >
            <div
              v-if="activeSheetSpec.isYellowBacking"
              class="w-full flex items-center justify-between text-[10px] font-extrabold uppercase tracking-wider text-amber-950 border-b border-amber-300/80 pb-1.5"
            >
              <span>{{ activeSheetSpec.shortLabel }}</span>
              <span class="font-mono font-semibold opacity-80">{{ activeSheetSpec.perSheet }} Label / Lembar</span>
            </div>
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
                class="w-full px-2 py-0.5 text-center font-extrabold uppercase tracking-tight shrink-0 flex flex-row items-center justify-center"
                :class="activeTemplate.tableConfig.headerStyle === 'inverted' ? 'bg-black text-white' : 'bg-slate-50/70 text-slate-900'"
                :style="{
                  gap: `${1.4 * MM_TO_PX * zoom}px`,
                  borderBottom: activeTemplate.tableConfig.showHeaderBorder !== false
                    ? activeTemplate.tableConfig.headerStyle === 'double_line'
                      ? `${activeTemplate.tableConfig.borderWidth * 2.2 * (zoom / 2.2)}px double ${activeTemplate.tableConfig.borderColor}`
                      : `${activeTemplate.tableConfig.borderWidth * (zoom / 2.2)}px ${activeTemplate.tableConfig.borderStyle} ${activeTemplate.tableConfig.borderColor}`
                    : 'none',
                }"
              >
                <!-- Lambang Kota Cirebon di Sebelah Tulisan Kop -->
                <div
                  v-if="isLogoInHeader(activeTemplate.tableConfig)"
                  class="shrink-0 flex items-center justify-center overflow-hidden"
                  :style="{
                    width: `${Math.max(6.2, Math.min(8.5, (activeTemplate.tableConfig.headerFontSize || 7) * 0.3528 * (activeTemplate.tableConfig.headerSubText ? 2.65 : 1.75))) * MM_TO_PX * zoom}px`,
                    height: `${Math.max(6.2, Math.min(8.5, (activeTemplate.tableConfig.headerFontSize || 7) * 0.3528 * (activeTemplate.tableConfig.headerSubText ? 2.65 : 1.75))) * MM_TO_PX * zoom}px`,
                  }"
                >
                  <img
                    v-if="activeTemplate.tableConfig.logoUrl"
                    :src="activeTemplate.tableConfig.logoUrl"
                    class="w-full h-full object-contain pointer-events-none"
                    alt="Logo"
                  />
                  <svg
                    v-else
                    viewBox="0 0 100 120"
                    class="w-full h-full object-contain pointer-events-none"
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

                <div class="flex flex-col items-center justify-center min-w-0 text-center">
                  <span
                    class="leading-[1.12]"
                    :style="{ fontSize: `${activeTemplate.tableConfig.headerFontSize * (zoom / 2.2)}pt` }"
                  >
                    {{ activeTemplate.tableConfig.headerText }}
                  </span>
                  <span
                    v-if="activeTemplate.tableConfig.headerSubText"
                    class="font-bold opacity-90 tracking-normal leading-[1.1]"
                    :style="{ fontSize: `${Math.max(4.5, activeTemplate.tableConfig.headerFontSize - 0.5) * (zoom / 2.2)}pt` }"
                  >
                    {{ activeTemplate.tableConfig.headerSubText }}
                  </span>
                </div>
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
                  <!-- Lambang Kota di Atas jika showLogo aktif dan bukan di Kop -->
                  <div
                    v-if="activeTemplate.tableConfig.showLogo && !isLogoInHeader(activeTemplate.tableConfig)"
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

                  <!-- QR Code (Diperbesar Penuh saat Logo di Kop) -->
                  <div class="w-full flex-1 min-h-0 flex flex-col items-center justify-center p-0.5 overflow-hidden">
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
                      selectedTableRowIndex === rIdx ? 'ring-1 ring-acc-500 bg-teal-50/50' : 'hover:bg-teal-50/20',
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
                          selectedTableRowIndex === rIdx && selectedTableCellIndex === cIdx ? 'ring-1 ring-acc-500 bg-teal-50/60 font-semibold' : 'hover:bg-teal-50/30',
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
                        <div class="opacity-0 group-hover/cell:opacity-100 absolute right-0.5 top-0.5 flex items-center gap-0.5 z-10 bg-panel/95 rounded px-0.5 shadow-2xs border border-line">
                          <!-- Split / Pecah Kolom -->
                          <button
                            type="button"
                            aria-label="Pecah sel atau tambah kolom"
                            class="p-0.5 rounded hover:bg-acc-500 hover:text-white text-acc-500 text-[8px]"
                            title="Pecah Sel / Tambah Kolom (Split)"
                            @click.stop="splitCell(rIdx, cIdx)"
                          >
                            <i class="pi pi-arrows-h text-[8px]" aria-hidden="true" />
                          </button>
                          <!-- Merge / Gabung dengan Sel Kanan -->
                          <button
                            v-if="cIdx < getRowCells(row).length - 1"
                            type="button"
                            aria-label="Gabung dengan sel kanan"
                            class="p-0.5 rounded hover:bg-[var(--sig-ok)] hover:text-white text-[var(--sig-ok)] text-[8px]"
                            title="Gabung dengan Sel Kanan (Merge)"
                            @click.stop="mergeCellsWithNext(rIdx, cIdx)"
                          >
                            <i class="pi pi-link text-[8px]" aria-hidden="true" />
                          </button>
                          <!-- Hapus Sel Ini -->
                          <button
                            type="button"
                            aria-label="Hapus sel ini"
                            class="p-0.5 rounded hover:bg-[var(--sig-bad)] hover:text-white text-[var(--sig-bad)] text-[8px]"
                            title="Hapus Sel Ini"
                            @click.stop="removeCellFromRow(rIdx, cIdx)"
                          >
                            <i class="pi pi-times text-[8px]" aria-hidden="true" />
                          </button>
                        </div>
                      </div>
                    </template>

                    <!-- TOMBOL CEPAT BARIS (Hover Overlay di ujung kanan baris) -->
                    <div class="opacity-0 group-hover/row:opacity-100 absolute right-0 top-0 bottom-0 flex items-center gap-0.5 bg-panel/95 px-1 border-l border-line shadow-xs transition-opacity z-20">
                      <!-- Tambah Sel ke Baris Ini -->
                      <button
                        type="button"
                        aria-label="Tambah sel atau kolom di baris ini"
                        class="p-0.5 rounded bg-[var(--acc-50)] hover:bg-acc-500 hover:text-white text-acc-500 text-[10px]"
                        title="Tambah Sel / Kolom di Baris Ini"
                        @click.stop="addCellToRow(rIdx)"
                      >
                        <i class="pi pi-plus-circle text-[10px]" aria-hidden="true" />
                      </button>
                      <!-- Sisip Baris Baru Tepat di Bawah Ini -->
                      <button
                        type="button"
                        aria-label="Sisipkan baris baru di bawah ini"
                        class="p-0.5 rounded bg-panel-2 hover:bg-acc-500 hover:text-white text-txt-dim text-[10px]"
                        title="Sisipkan Baris Baru di Bawah Ini"
                        @click.stop="addNewRow('location', 'Ruang', rIdx)"
                      >
                        <i class="pi pi-plus text-[10px]" aria-hidden="true" />
                      </button>
                      <!-- Gabung (Merge) dengan Baris di Bawah Ini -->
                      <button
                        v-if="rIdx < activeTemplate.tableConfig.rows.length - 1"
                        type="button"
                        aria-label="Gabung dengan baris di bawahnya"
                        class="p-0.5 rounded bg-[var(--sig-ok-bg)] hover:bg-[var(--sig-ok)] hover:text-white text-[var(--sig-ok)] text-[10px]"
                        title="Gabung (Merge) dengan Baris di Bawahnya"
                        @click.stop="mergeRowWithNext(rIdx)"
                      >
                        <i class="pi pi-link text-[10px]" aria-hidden="true" />
                      </button>
                      <!-- Hapus Baris Ini -->
                      <button
                        type="button"
                        aria-label="Hapus seluruh baris ini"
                        class="p-0.5 rounded bg-[var(--sig-bad-bg)] hover:bg-[var(--sig-bad)] hover:text-white text-[var(--sig-bad)] text-[10px]"
                        title="Hapus Seluruh Baris Ini"
                        @click.stop="removeTableRow(rIdx)"
                      >
                        <i class="pi pi-trash text-[10px]" aria-hidden="true" />
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
                      class="px-2.5 py-1 bg-acc-500 hover:bg-acc-600 text-white rounded font-semibold text-xs shadow-xs"
                      @click="addNewRow('name', 'Nama')"
                    >
                      + Tambah Baris Pertama
                    </button>
                  </div>
                </div>

                <!-- QR Code Column (Right) -->
                <div
                  v-if="activeTemplate.tableConfig.qrPosition === 'right'"
                  class="shrink-0 flex flex-col items-center justify-between overflow-hidden"
                  :style="{
                    width: `${activeTemplate.tableConfig.qrWidthMm * MM_TO_PX * zoom}px`,
                    borderLeft: activeTemplate.tableConfig.showColBorder
                      ? `${activeTemplate.tableConfig.borderWidth * (zoom / 2.2)}px ${activeTemplate.tableConfig.borderStyle} ${activeTemplate.tableConfig.borderColor}`
                      : 'none',
                  }"
                >
                  <div
                    v-if="activeTemplate.tableConfig.showLogo && !isLogoInHeader(activeTemplate.tableConfig)"
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

                  <div class="w-full flex-1 min-h-0 flex flex-col items-center justify-center p-0.5 overflow-hidden">
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

                  <div
                    v-if="activeTemplate.tableConfig.showItemNameUnderLogo"
                    class="w-full py-0.5 px-0.5 text-center truncate font-bold text-slate-800 border-t border-slate-300"
                    :style="{ fontSize: `${5 * (zoom / 2.2)}pt` }"
                  >
                    - {{ sampleItem.name }}
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
                  activeElementId === el.id ? 'ring-2 ring-acc-500 shadow-md z-20' : 'hover:ring-1 hover:ring-slate-400 z-10',
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
                  class="absolute -right-1 -bottom-1 w-3.5 h-3.5 bg-acc-500 border-2 border-white rounded-xs cursor-nwse-resize z-30 shadow-xs"
                  @pointerdown.stop="startResize($event, el)"
                />
              </div>
            </template>
            </div>
          </div>

          <!-- Tombol Tambah Baris Cepat (di Luar Stiker) -->
          <div v-if="isTableMode" class="mt-3 flex items-center justify-center">
            <button
              type="button"
              class="px-3 py-1.5 bg-[var(--panel)] hover:bg-[var(--acc-wash)] text-[var(--acc-600)] rounded-md border border-[var(--acc-border)] shadow-2xs text-[11px] font-semibold transition-colors flex items-center gap-1.5 cursor-pointer"
              @click="addNewRow('location', 'Ruang')"
            >
              <i class="pi pi-plus-circle text-xs" aria-hidden="true" />
              <span>Tambah Baris Baru</span>
            </button>
          </div>
        </div>
      </div>

      <!-- RIGHT COLUMN: PROPERTIES INSPECTOR (FULL KUSTOM BARIS & SEL) -->
      <div class="w-88 shrink-0 border-l border-[var(--line)] bg-[var(--panel)] p-3.5 flex flex-col gap-3 text-xs overflow-y-auto select-none">
        <div class="t-label flex items-center justify-between pb-2 border-b border-[var(--line)]">
          <span>{{ isTableMode ? 'Inspektor Baris & Sel' : 'Pengaturan Elemen' }}</span>
          <Tag :value="isTableMode ? 'KUSTOM BARIS & SEL' : 'MODE BEBAS'" severity="info" class="!text-[9px] !py-0" />
        </div>

        <!-- 1. INSPECTOR UNTUK MODE TABEL -->
        <div v-if="isTableMode" class="flex flex-col gap-3">
          <!-- A. Pengaturan Baris Terpilih -->
          <div v-if="selectedTableRow" class="p-3 bg-[var(--acc-wash)] rounded-lg border border-[var(--acc-border)] flex flex-col gap-2.5">
            <div class="flex items-center justify-between">
              <span class="font-semibold text-[var(--txt)] text-xs t-num">Baris #{{ (selectedTableRowIndex ?? 0) + 1 }}</span>
              <div class="flex items-center gap-1">
                <button
                  type="button"
                  class="p-1 rounded bg-[var(--panel)] border border-[var(--line)] text-[var(--txt-dim)] hover:text-[var(--acc-600)] hover:border-[var(--acc-border)] disabled:opacity-40 transition-colors"
                  :disabled="selectedTableRowIndex === 0"
                  title="Pindah Baris ke Atas"
                  aria-label="Pindah baris ke atas"
                  @click="moveRowUp(selectedTableRowIndex!)"
                >
                  <i class="pi pi-arrow-up text-xs" aria-hidden="true" />
                </button>
                <button
                  type="button"
                  class="p-1 rounded bg-[var(--panel)] border border-[var(--line)] text-[var(--txt-dim)] hover:text-[var(--acc-600)] hover:border-[var(--acc-border)] disabled:opacity-40 transition-colors"
                  :disabled="selectedTableRowIndex === activeTemplate.tableConfig.rows.length - 1"
                  title="Pindah Baris ke Bawah"
                  aria-label="Pindah baris ke bawah"
                  @click="moveRowDown(selectedTableRowIndex!)"
                >
                  <i class="pi pi-arrow-down text-xs" aria-hidden="true" />
                </button>
                <button
                  type="button"
                  class="p-1 rounded bg-[var(--panel)] border border-[var(--line)] text-[var(--txt-dim)] hover:text-[var(--acc-600)] hover:border-[var(--acc-border)] ml-0.5 transition-colors"
                  title="Duplikat Baris Ini"
                  aria-label="Duplikat baris ini"
                  @click="duplicateTableRow(selectedTableRowIndex!)"
                >
                  <i class="pi pi-copy text-xs" aria-hidden="true" />
                </button>
                <button
                  type="button"
                  class="p-1 rounded bg-[var(--sig-bad-bg)] border border-[var(--sig-bad-line)] text-[var(--sig-bad)] hover:opacity-85 ml-0.5 transition-opacity"
                  title="Hapus Baris Ini"
                  aria-label="Hapus baris ini"
                  @click="removeTableRow(selectedTableRowIndex!)"
                >
                  <i class="pi pi-trash text-xs" aria-hidden="true" />
                </button>
              </div>
            </div>

            <!-- Bagian Kelola Sel / Kolom di Baris Ini -->
            <div class="p-2 bg-[var(--panel)] rounded border border-[var(--line)] flex flex-col gap-2">
              <div class="flex items-center justify-between">
                <span class="t-label t-num">
                  Kolom / Sel ({{ getRowCells(selectedTableRow).length }})
                </span>
                <div class="flex items-center gap-1">
                  <button
                    v-if="selectedTableRowIndex !== null && selectedTableRowIndex < activeTemplate.tableConfig.rows.length - 1"
                    type="button"
                    class="px-2 py-0.5 bg-[var(--sig-ok-bg)] hover:bg-[var(--sig-ok)] hover:text-white text-[var(--sig-ok)] font-semibold text-[10px] rounded border border-[var(--sig-ok-line)] transition-colors flex items-center gap-1"
                    title="Gabung baris ini dengan baris di bawahnya"
                    @click="mergeRowWithNext(selectedTableRowIndex!)"
                  >
                    <i class="pi pi-link text-[9px]" aria-hidden="true" />
                    <span>Merge Baris</span>
                  </button>
                  <button
                    type="button"
                    class="px-2 py-0.5 bg-[var(--acc-wash)] hover:bg-acc-500 hover:text-white text-[var(--acc-600)] font-semibold text-[10px] rounded border border-[var(--acc-border)] transition-colors flex items-center gap-1"
                    title="Tambah Sel / Kolom Baru di Baris Ini"
                    @click="addCellToRow(selectedTableRowIndex!)"
                  >
                    <i class="pi pi-plus text-[9px]" aria-hidden="true" />
                    <span>Tambah Sel</span>
                  </button>
                </div>
              </div>

              <!-- Daftar Chips Sel di Baris Ini -->
              <div class="flex flex-wrap gap-1.5">
                <div
                  v-for="(c, cIdx) in getRowCells(selectedTableRow)"
                  :key="c.id"
                  class="flex items-center gap-1 px-2 py-1 rounded border text-[11px] cursor-pointer transition-colors"
                  :class="selectedTableCellIndex === cIdx ? 'bg-acc-500 border-acc-600 text-white font-semibold shadow-2xs' : 'bg-[var(--panel-2)] border-[var(--line)] text-[var(--txt-dim)] hover:text-[var(--txt)]'"
                  @click="selectedTableCellIndex = cIdx"
                >
                  <span class="t-num">Sel {{ cIdx + 1 }}: {{ c.label || c.field }}</span>
                  <button
                    v-if="getRowCells(selectedTableRow).length > 1"
                    type="button"
                    class="ml-0.5 p-0.5 rounded-full hover:bg-black/20 text-current transition-colors"
                    title="Hapus Sel Ini"
                    aria-label="Hapus sel ini"
                    @click.stop="removeCellFromRow(selectedTableRowIndex!, cIdx)"
                  >
                    <i class="pi pi-times text-[8px]" aria-hidden="true" />
                  </button>
                </div>
              </div>
            </div>

            <!-- Detail Edit Sel Terpilih -->
            <div v-if="selectedTableCell" class="p-2.5 bg-[var(--panel)] rounded border border-[var(--line)] flex flex-col gap-2">
              <div class="flex items-center justify-between border-b border-[var(--line-soft)] pb-1">
                <span class="font-semibold text-[var(--acc-600)] text-[11px] t-num">
                  Edit Sel #{{ (selectedTableCellIndex ?? 0) + 1 }}
                </span>
                <button
                  v-if="getRowCells(selectedTableRow).length > 1"
                  type="button"
                  class="text-[var(--sig-bad)] hover:underline font-semibold text-[10px] flex items-center gap-1"
                  @click="removeCellFromRow(selectedTableRowIndex!, selectedTableCellIndex!)"
                >
                  <i class="pi pi-trash text-[9px]" aria-hidden="true" />
                  <span>Hapus Sel Ini</span>
                </button>
              </div>

              <div class="flex flex-col gap-1">
                <label class="text-[10px] font-semibold text-[var(--txt-dim)]">Label Judul Sel:</label>
                <InputText v-model="selectedTableCell.label" placeholder="Contoh: Ruang" class="w-full !text-xs !py-1" />
              </div>

              <div class="flex flex-col gap-1">
                <label class="text-[10px] font-semibold text-[var(--txt-dim)]">Field Data Barang:</label>
                <Select
                  v-model="selectedTableCell.field"
                  :options="AVAILABLE_FIELDS"
                  optionLabel="label"
                  optionValue="value"
                  class="w-full !text-xs"
                />
              </div>

              <div class="flex flex-col gap-1">
                <label class="text-[10px] font-semibold text-[var(--txt-dim)]">Teks Manual (Opsional):</label>
                <InputText v-model="selectedTableCell.customText" placeholder="Kosongkan jika ambil otomatis dari field" class="w-full !text-xs !py-1" />
              </div>

              <div class="grid grid-cols-2 gap-2 pt-1 border-t border-[var(--line-soft)]">
                <div class="flex flex-col gap-1">
                  <label class="text-[10px] text-[var(--txt-dim)]">Ukuran Font:</label>
                  <div class="flex items-center gap-1 bg-[var(--panel-2)] border border-[var(--line)] rounded px-1.5 py-0.5">
                    <input
                      type="number"
                      v-model.number="selectedTableCell.fontSize"
                      min="4"
                      max="14"
                      step="0.5"
                      class="w-full text-center font-semibold text-xs t-num text-[var(--txt)] bg-transparent focus:outline-hidden"
                    />
                    <span class="text-[10px] text-[var(--txt-muted)]">pt</span>
                  </div>
                </div>

                <div class="flex flex-col gap-1">
                  <label class="text-[10px] text-[var(--txt-dim)]">Warna Latar Sel:</label>
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
                <label class="flex items-center gap-1.5 cursor-pointer text-[var(--txt)] text-xs">
                  <Checkbox v-model="selectedTableCell.bold" binary />
                  <span class="font-semibold">Teks Tebal</span>
                </label>

                <div class="flex items-center gap-1">
                  <button
                    type="button"
                    class="px-2 py-0.5 rounded border text-[10.5px] transition-colors"
                    :class="selectedTableCell.align === 'left' ? 'bg-acc-500 border-acc-600 text-white font-semibold' : 'bg-[var(--panel-2)] border-[var(--line)] text-[var(--txt-dim)]'"
                    @click="selectedTableCell.align = 'left'"
                  >
                    Kiri
                  </button>
                  <button
                    type="button"
                    class="px-2 py-0.5 rounded border text-[10.5px] transition-colors"
                    :class="selectedTableCell.align === 'center' ? 'bg-acc-500 border-acc-600 text-white font-semibold' : 'bg-[var(--panel-2)] border-[var(--line)] text-[var(--txt-dim)]'"
                    @click="selectedTableCell.align = 'center'"
                  >
                    Tengah
                  </button>
                  <button
                    type="button"
                    class="px-2 py-0.5 rounded border text-[10.5px] transition-colors"
                    :class="selectedTableCell.align === 'right' ? 'bg-acc-500 border-acc-600 text-white font-semibold' : 'bg-[var(--panel-2)] border-[var(--line)] text-[var(--txt-dim)]'"
                    @click="selectedTableCell.align = 'right'"
                  >
                    Kanan
                  </button>
                </div>
              </div>
              <!-- Tombol Cepat Aksi Sel (Pecah / Split & Gabung / Merge) -->
              <div class="flex items-center gap-1.5 pt-2 border-t border-[var(--line-soft)] flex-wrap">
                <button
                  type="button"
                  class="flex-1 py-1 px-1.5 rounded bg-[var(--acc-wash)] hover:bg-acc-500 hover:text-white text-[var(--acc-600)] border border-[var(--acc-border)] text-[10px] font-semibold flex items-center justify-center gap-1 transition-colors"
                  title="Pecah sel ini / tambah kolom baru di sebelahnya"
                  @click="splitCell(selectedTableRowIndex!, selectedTableCellIndex!)"
                >
                  <i class="pi pi-arrows-h text-xs" aria-hidden="true" />
                  <span>Pecah (Split)</span>
                </button>
                <button
                  type="button"
                  :disabled="selectedTableCellIndex === null || selectedTableCellIndex >= getRowCells(selectedTableRow).length - 1"
                  class="flex-1 py-1 px-1.5 rounded bg-[var(--sig-ok-bg)] hover:bg-[var(--sig-ok)] hover:text-white text-[var(--sig-ok)] border border-[var(--sig-ok-line)] text-[10px] font-semibold flex items-center justify-center gap-1 disabled:opacity-40 transition-colors"
                  title="Gabung sel ini dengan sel sebelah kanan"
                  @click="mergeCellsWithNext(selectedTableRowIndex!, selectedTableCellIndex!)"
                >
                  <i class="pi pi-link text-xs" aria-hidden="true" />
                  <span>Gabung (Merge)</span>
                </button>
              </div>
            </div>

            <!-- Tombol Aksi Hapus Baris -->
            <div class="pt-1 flex items-center justify-between">
              <label class="flex items-center gap-1.5 cursor-pointer text-[var(--txt)] text-xs">
                <Checkbox v-model="selectedTableRow.hasBottomBorder" binary />
                <span>Paksa Garis Bawah</span>
              </label>

              <button
                type="button"
                class="px-2.5 py-1 rounded bg-[var(--sig-bad-bg)] hover:bg-[var(--sig-bad)] hover:text-white text-[var(--sig-bad)] border border-[var(--sig-bad-line)] text-xs font-semibold transition-colors flex items-center gap-1"
                @click="removeTableRow(selectedTableRowIndex!)"
              >
                <i class="pi pi-trash text-xs" aria-hidden="true" />
                <span>Hapus Seluruh Baris</span>
              </button>
            </div>
          </div>

          <div v-else class="p-3 bg-[var(--panel-2)] rounded-lg border border-[var(--line)] text-[var(--txt-dim)] text-xs text-center flex flex-col items-center gap-1.5">
            <i class="pi pi-arrow-left text-sm text-[var(--txt-muted)]" aria-hidden="true" />
            <span>Klik salah satu baris pada kanvas stiker untuk mengatur isi sel, menambah kolom, atau menghapusnya.</span>
          </div>

          <!-- B. Daftar Urutan Baris Aktif -->
          <div class="flex flex-col gap-1.5">
            <div class="flex items-center justify-between">
              <span class="t-label t-num">Urutan Baris ({{ activeTemplate.tableConfig.rows.length }})</span>
              <div class="flex items-center gap-2">
                <button
                  v-if="activeTemplate.tableConfig.rows.length > 0"
                  type="button"
                  class="text-[var(--sig-bad)] hover:underline text-[10px] font-semibold flex items-center gap-1"
                  title="Hapus seluruh baris stiker"
                  @click="clearAllRows"
                >
                  <i class="pi pi-trash text-[9px]" aria-hidden="true" />
                  <span>Kosongkan</span>
                </button>
                <button
                  type="button"
                  class="text-[var(--acc-600)] hover:underline text-[11px] font-semibold flex items-center gap-1"
                  @click="addNewRow('name', 'Nama')"
                >
                  <i class="pi pi-plus text-[10px]" aria-hidden="true" />
                  <span>Tambah Baris</span>
                </button>
              </div>
            </div>

            <div class="flex flex-col gap-1 max-h-56 overflow-y-auto">
              <div
                v-for="(r, idx) in activeTemplate.tableConfig.rows"
                :key="r.id"
                class="flex items-center justify-between px-2 py-1 rounded border text-xs cursor-pointer transition-colors"
                :class="selectedTableRowIndex === idx ? 'bg-[var(--acc-wash)] border-[var(--acc-border)] font-semibold text-[var(--txt)]' : 'bg-[var(--panel)] border-[var(--line)] text-[var(--txt-dim)] hover:bg-[var(--panel-2)]'"
                @click="selectedTableRowIndex = idx; selectedTableCellIndex = 0"
              >
                <div class="flex items-center gap-1.5 truncate">
                  <span class="text-[var(--txt-muted)] text-[10px] t-num">{{ idx + 1 }}.</span>
                  <span v-if="r.type === 'barcode_strip'" class="truncate text-[var(--acc-600)] t-mono">|||| Barcode Garis</span>
                  <span v-else class="truncate">
                    {{ getRowCells(r).map((c) => c.label || c.field).join(' | ') }}
                    <span class="text-[10px] font-normal text-[var(--txt-muted)] t-num">({{ getRowCells(r).length }} sel)</span>
                  </span>
                </div>
                <div class="flex items-center gap-1 shrink-0">
                  <button
                    v-if="idx < activeTemplate.tableConfig.rows.length - 1"
                    type="button"
                    class="text-[var(--sig-ok)] hover:opacity-80 p-0.5"
                    title="Gabung (Merge) dengan baris di bawahnya"
                    aria-label="Gabung dengan baris di bawahnya"
                    @click.stop="mergeRowWithNext(idx)"
                  >
                    <i class="pi pi-link text-[11px]" aria-hidden="true" />
                  </button>
                  <button
                    type="button"
                    class="text-[var(--acc-600)] hover:opacity-80 p-0.5"
                    title="Tambah Sel"
                    aria-label="Tambah sel ke baris"
                    @click.stop="addCellToRow(idx)"
                  >
                    <i class="pi pi-plus-circle text-[11px]" aria-hidden="true" />
                  </button>
                  <button
                    type="button"
                    class="text-[var(--txt-muted)] hover:text-[var(--acc-600)] p-0.5"
                    title="Duplikat"
                    aria-label="Duplikat baris"
                    @click.stop="duplicateTableRow(idx)"
                  >
                    <i class="pi pi-copy text-[10px]" aria-hidden="true" />
                  </button>
                  <button
                    type="button"
                    class="text-[var(--sig-bad)] hover:opacity-80 p-0.5"
                    title="Hapus Baris"
                    aria-label="Hapus baris"
                    @click.stop="removeTableRow(idx)"
                  >
                    <i class="pi pi-trash text-[10px]" aria-hidden="true" />
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 2. INSPECTOR UNTUK MODE BEBAS (Drag & Drop) -->
        <div v-else-if="activeElement" class="flex flex-col gap-3">
          <div v-if="activeElement.type === 'header' || activeElement.type === 'static_text'" class="flex flex-col gap-1">
            <label class="text-[var(--txt-dim)] font-semibold text-[11px]">Teks Tampilan</label>
            <InputText v-model="activeElement.text" class="w-full !text-xs !py-1" />
          </div>

          <div v-if="activeElement.type === 'field'" class="flex flex-col gap-2 p-2.5 bg-[var(--panel-2)] rounded-lg border border-[var(--line)]">
            <div class="flex flex-col gap-1">
              <label class="text-[var(--txt-dim)] font-semibold text-[11px]">Field dari Formulir</label>
              <Select
                v-model="activeElement.field"
                :options="AVAILABLE_FIELDS"
                optionLabel="label"
                optionValue="value"
                class="w-full !text-xs"
              />
            </div>
            <div class="flex flex-col gap-1">
              <label class="text-[var(--txt-dim)] font-semibold text-[11px]">Teks Awalan (Prefix)</label>
              <InputText v-model="activeElement.prefix" placeholder="Contoh: NUP: " class="w-full !text-xs !py-1" />
            </div>
          </div>

          <!-- Position & Dimension -->
          <div class="p-2.5 bg-[var(--panel-2)] rounded-lg border border-[var(--line)] flex flex-col gap-2">
            <span class="text-[11px] font-semibold text-[var(--txt)]">Posisi &amp; Ukuran (mm)</span>
            <div class="grid grid-cols-2 gap-2">
              <div class="flex items-center bg-[var(--panel)] rounded border border-[var(--line)] px-2 py-1">
                <span class="text-[var(--txt-muted)] text-[10px] mr-1">X:</span>
                <input type="number" v-model.number="activeElement.x" step="0.5" class="w-full text-xs font-semibold t-num text-[var(--txt)] bg-transparent focus:outline-hidden" />
              </div>
              <div class="flex items-center bg-[var(--panel)] rounded border border-[var(--line)] px-2 py-1">
                <span class="text-[var(--txt-muted)] text-[10px] mr-1">Y:</span>
                <input type="number" v-model.number="activeElement.y" step="0.5" class="w-full text-xs font-semibold t-num text-[var(--txt)] bg-transparent focus:outline-hidden" />
              </div>
              <div class="flex items-center bg-[var(--panel)] rounded border border-[var(--line)] px-2 py-1">
                <span class="text-[var(--txt-muted)] text-[10px] mr-1">P:</span>
                <input type="number" v-model.number="activeElement.w" step="0.5" class="w-full text-xs font-semibold t-num text-[var(--txt)] bg-transparent focus:outline-hidden" />
              </div>
              <div class="flex items-center bg-[var(--panel)] rounded border border-[var(--line)] px-2 py-1">
                <span class="text-[var(--txt-muted)] text-[10px] mr-1">L:</span>
                <input type="number" v-model.number="activeElement.h" step="0.5" class="w-full text-xs font-semibold t-num text-[var(--txt)] bg-transparent focus:outline-hidden" />
              </div>
            </div>
          </div>

          <div class="flex items-center gap-2 pt-1">
            <Button icon="pi pi-copy" label="Duplikat" size="small" severity="secondary" outlined class="w-1/2 !text-xs !py-1" @click="duplicateActiveElement" />
            <Button icon="pi pi-trash" label="Hapus" size="small" severity="danger" class="w-1/2 !text-xs !py-1" @click="removeActiveElement" />
          </div>
        </div>

        <div v-else class="flex flex-col items-center justify-center p-8 text-center text-[var(--txt-muted)]">
          <i class="pi pi-mouse text-3xl mb-2 opacity-50" aria-hidden="true" />
          <p class="text-xs">Klik salah satu elemen pada stiker untuk mengatur ukuran, teks, atau posisinya.</p>
        </div>
      </div>
    </div>
  </Dialog>

  <!-- Dialog Simpan Template Baru -->
  <Dialog
    v-model:visible="saveTemplateDialog"
    modal
    header="Simpan Template Label Baru"
    :style="{ width: '26rem' }"
    :breakpoints="{ '640px': '92vw' }"
  >
    <form class="flex flex-col gap-3 pt-1" @submit.prevent="confirmSaveNewTemplate">
      <div class="flex flex-col gap-1.5">
        <label for="save-template-name" class="text-xs font-semibold text-[var(--txt)]">Nama Template</label>
        <InputText
          id="save-template-name"
          v-model="saveTemplateName"
          placeholder="Contoh: Stiker BMN Lab Komputer"
          autofocus
          class="w-full"
        />
        <p class="text-[11px] text-[var(--txt-muted)]">
          Template tersimpan di database dan langsung muncul di dropdown pilihan template.
        </p>
      </div>
      <div class="flex justify-end gap-2 pt-2 border-t border-[var(--line-soft)]">
        <Button type="button" label="Batal" severity="secondary" outlined size="small" @click="saveTemplateDialog = false" />
        <Button type="submit" label="Simpan Template" icon="pi pi-check" size="small" :loading="savingTemplate" />
      </div>
    </form>
  </Dialog>
</template>

<style scoped>
.tab-btn {
  padding: 6px 9px;
  font-size: 11px;
  font-weight: 600;
  border-radius: 6px;
  color: var(--txt-dim);
  background: transparent;
  border: none;
  cursor: pointer;
  transition: all 0.15s ease;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  white-space: nowrap;
}
.tab-btn:hover {
  color: var(--txt);
}
.tab-btn-active {
  background: var(--panel);
  color: var(--acc-500);
  box-shadow: 0 1px 2px 0 rgba(15, 23, 42, 0.06);
}

.component-item-btn {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  background-color: var(--panel);
  border: 1px solid var(--line);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s ease;
}
.component-item-btn:hover {
  background-color: var(--panel-2);
  border-color: var(--acc-500);
  box-shadow: 0 1px 2px 0 rgba(15, 23, 42, 0.05);
}
</style>
