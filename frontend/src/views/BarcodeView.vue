<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import api from '@/api'
import { useToast } from 'primevue/usetoast'
import { useConfirm } from 'primevue/useconfirm'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import InputNumber from 'primevue/inputnumber'
import Select from 'primevue/select'
import Checkbox from 'primevue/checkbox'
import Tag from 'primevue/tag'
import Dialog from 'primevue/dialog'
import PageHeader from '@/components/PageHeader.vue'
import StatCard from '@/components/StatCard.vue'
import EmptyState from '@/components/EmptyState.vue'
import LabelDesignerModal from '@/components/LabelDesignerModal.vue'
import {
  type LabelTemplate,
  type SheetFormatId,
  inferSheetFormat,
  getSheetSpec,
  ensureTableConfig,
  renderCustomLabelHtml,
  printCustomLabels,
  openCustomPrintTab,
} from '@/utils/customLabelPrint'

const toast = useToast()
const confirm = useConfirm()
const items = ref<any[]>([])
const selectedItemsMap = ref<Record<string, any>>({})
const total = ref(0)
const selected = ref<string[]>([])
const loading = ref(true)
const printingCustom = ref(false)
const showCustomDesigner = ref(false)
const saveTemplateDialog = ref(false)
const saveTemplateName = ref('')
const savingTemplate = ref(false)

const page = ref(0)
const perPage = ref(24)
const perPageOptions = [
  { label: '24 / hal', value: 24 },
  { label: '48 / hal', value: 48 },
  { label: '96 / hal', value: 96 },
]

const q = ref('')
const categoryFilter = ref('')
const locationFilter = ref('')
const fmt = ref<'qr' | 'code128'>('qr')
const size = ref(160)

// Custom Label Templates dari Database & Penyimpanan Lokal
const customTemplates = ref<any[]>([])
const selectedCustomTemplateId = ref<string>('')

async function loadCustomTemplates(newTpl?: any) {
  const map = new Map<string, any>()

  // 1. Muat dari localStorage (instan, tahan reload, tahan offline)
  const localKeys = ['custom_label_templates_v6', 'custom_label_templates', 'active_custom_label_template_v6']
  for (const k of localKeys) {
    try {
      const raw = localStorage.getItem(k)
      if (raw) {
        const parsed = JSON.parse(raw)
        if (Array.isArray(parsed)) {
          parsed.forEach((t) => {
            if (t && t.id && !t.id.startsWith('preset_')) {
              map.set(t.id, t)
            }
          })
        } else if (parsed && parsed.id && !parsed.id.startsWith('preset_')) {
          map.set(parsed.id, parsed)
        }
      }
    } catch {}
  }

  // 2. Muat dari backend database (PostgreSQL)
  try {
    const res = await api.get('/barcode/templates')
    if (res.data && Array.isArray(res.data) && res.data.length > 0) {
      res.data.forEach((t: any) => {
        if (t && t.id) {
          map.set(t.id, t)
        }
      })
    }
  } catch (err) {
    console.warn('Gagal memuat template kustom dari server:', err)
  }

  // 3. Gabungkan template yang baru saja di-save jika ada
  if (newTpl && newTpl.id && !newTpl.id.startsWith('preset_')) {
    map.set(newTpl.id, newTpl)
  }

  customTemplates.value = Array.from(map.values()).map((t: any) => {
    const normalized: LabelTemplate = {
      ...t,
      widthMm: t.widthMm || 64,
      heightMm: t.heightMm || 32,
      layoutMode: t.layoutMode || 'table',
      sheetFormat: t.sheetFormat || inferSheetFormat(t),
      elements: Array.isArray(t.elements) ? t.elements : [],
      tableConfig: t.tableConfig,
    }
    ensureTableConfig(normalized)
    return normalized
  })

  // Simpan kembali list lengkap ke localStorage
  try {
    localStorage.setItem('custom_label_templates_v6', JSON.stringify(customTemplates.value))
  } catch {}

  // Jika ada template baru disimpan, otomatis pilih di dropdown
  if (newTpl && newTpl.id) {
    const targetVal = newTpl.id.startsWith('custom_') ? newTpl.id : `custom_${newTpl.id}`
    layout.value = targetVal
    try {
      localStorage.setItem('barcode_selected_layout', targetVal)
    } catch {}
  }
}

function onTemplateSaved(savedTpl: any) {
  loadCustomTemplates(savedTpl)
  toast.add({
    severity: 'success',
    summary: 'Template Berhasil Disimpan',
    detail: `Template "${(savedTpl?.name || '').replace(/^[🌟🏷️\s]+/, '')}" kini aktif di dropdown`,
    life: 3000,
  })
}

async function openCustomDesignerWithTemplate(templateId?: string) {
  if (templateId) {
    selectedCustomTemplateId.value = templateId
  } else if (selectedCustomTemplate.value?.id) {
    selectedCustomTemplateId.value = selectedCustomTemplate.value.id
  } else {
    selectedCustomTemplateId.value = ''
  }
  if (selected.value.length > 0) {
    await ensureSelectedItemsLoaded()
  }
  showCustomDesigner.value = true
}

// Thermal Printer, Sticker Sheets & Label Profiles
let savedLayout: string | null = null
try {
  savedLayout = localStorage.getItem('barcode_selected_layout')
} catch {}
const layout = ref(savedLayout || 'tj_103')

watch(layout, (newVal) => {
  if (newVal) {
    try {
      localStorage.setItem('barcode_selected_layout', newVal)
    } catch {}
  }
})

const selectedCustomTemplate = computed(() => {
  if (!layout.value.startsWith('custom_')) return null
  const rawId = layout.value.replace(/^custom_/, '')
  return customTemplates.value.find((t) => {
    const tId = t.id || ''
    const cleanTId = tId.replace(/^custom_/, '')
    return tId === layout.value || tId === rawId || cleanTId === rawId
  }) || null
})

const layoutOptions = computed(() => {
  const list: any[] = []
  if (customTemplates.value.length > 0) {
    customTemplates.value.forEach((t) => {
      const rawId = t.id || ''
      const customVal = rawId.startsWith('custom_') ? rawId : `custom_${rawId}`
      const cleanName = (t.name || 'Desain Kustom').replace(/^[🌟🏷️🖨️📄⭐\s]+/, '')
      list.push({
        label: `[Tersimpan] ${cleanName} (${t.widthMm || 64}x${t.heightMm || 32} mm)`,
        value: customVal,
      })
    })
  }
  list.push(
    { label: 'Tom & Jerry No. 103 (12 Label / Lembar · 3x4 · 64x32 mm - Rekomendasi)', value: 'tj_103' },
    { label: 'Tom & Jerry No. 107 (24 Label / Lembar · 3x8 · 50x18 mm)', value: 'tj_107' },
    { label: 'Tom & Jerry No. 108 (40 Label / Lembar · 5x8 · 38x18 mm)', value: 'tj_108' },
    { label: 'Tom & Jerry No. 121 (10 Label / Lembar · 2x5 · 75x38 mm)', value: 'tj_121' },
    { label: 'Kertas Stiker A4 (21 Label / Lembar · 3x7 · 65x38 mm)', value: 'a4_3x7' },
    { label: 'Kertas Stiker A4 (40 Label / Lembar · 4x10 · 48x25.5 mm)', value: 'a4_4x10' },
    { label: 'Printer Thermal Roll 58 mm (Struk Kasir)', value: 'thermal_58' },
    { label: 'Printer Thermal Roll 80 mm (Struk Lebar)', value: 'thermal_80' },
    { label: 'Stiker Rol 50 x 30 mm (Standar Printer Barcode)', value: 'label_50x30' },
    { label: 'Stiker Rol 40 x 30 mm (Mini Aset)', value: 'label_40x30' },
    { label: 'Stiker Rol 60 x 40 mm (Sedang)', value: 'label_60x40' },
    { label: 'Stiker Rol 70 x 50 mm (Besar)', value: 'label_70x50' },
    { label: 'Stiker Rol 100 x 50 mm (Stiker Aset Lebar)', value: 'label_100x50' },
  )
  return list
})

function saveCurrentLayoutAsTemplate() {
  const currentOption = layoutOptions.value.find((o) => o.value === layout.value)
  const defaultName = currentOption
    ? currentOption.label.replace(/^(\[Tersimpan\]|[🌟📄🏷️🖨️⭐])+\s*/g, '').split('(')[0].trim()
    : 'Template Label'
  saveTemplateName.value = defaultName || 'Template Label'
  saveTemplateDialog.value = true
}

async function confirmSaveLayoutAsTemplate() {
  const cleanName = saveTemplateName.value.trim()
  if (!cleanName) {
    toast.add({ severity: 'warn', summary: 'Nama template wajib diisi', life: 2500 })
    return
  }
  savingTemplate.value = true
  const customId = `custom_${Date.now().toString(36)}`

  let widthMm = 64
  let heightMm = 32
  let sheetFormat: SheetFormatId = 'single'
  if (layout.value === 'tj_103') {
    widthMm = 64; heightMm = 32; sheetFormat = 'tj_103'
  } else if (layout.value === 'tj_107') {
    widthMm = 50; heightMm = 18; sheetFormat = 'tj_107'
  } else if (layout.value === 'tj_108') {
    widthMm = 38; heightMm = 18; sheetFormat = 'tj_108'
  } else if (layout.value === 'tj_121') {
    widthMm = 75; heightMm = 38; sheetFormat = 'tj_121'
  } else if (layout.value === 'a4_3x7') {
    widthMm = 65; heightMm = 38; sheetFormat = 'a4_3x7'
  } else if (layout.value === 'a4_4x10') {
    widthMm = 48; heightMm = 25.5; sheetFormat = 'a4_4x10'
  } else if (layout.value.includes('50x30') || layout.value === 'label_50x30') {
    widthMm = 50; heightMm = 30
  } else if (layout.value.includes('40x30') || layout.value === 'label_40x30') {
    widthMm = 40; heightMm = 30
  } else if (layout.value.includes('60x40') || layout.value === 'label_60x40') {
    widthMm = 60; heightMm = 40
  } else if (layout.value.includes('70x50') || layout.value === 'label_70x50') {
    widthMm = 70; heightMm = 50
  } else if (layout.value.includes('100x50') || layout.value === 'label_100x50') {
    widthMm = 100; heightMm = 50
  } else if (layout.value === 'thermal_58') {
    widthMm = 58; heightMm = 40
  } else if (layout.value === 'thermal_80') {
    widthMm = 80; heightMm = 50
  } else if (selectedCustomTemplate.value) {
    widthMm = selectedCustomTemplate.value.widthMm || 64
    heightMm = selectedCustomTemplate.value.heightMm || 32
    sheetFormat = selectedCustomTemplate.value.sheetFormat || inferSheetFormat(selectedCustomTemplate.value)
  }

  let newTpl: LabelTemplate
  if (selectedCustomTemplate.value) {
    const cloned = JSON.parse(JSON.stringify(selectedCustomTemplate.value))
    newTpl = {
      ...cloned,
      id: customId,
      name: cleanName,
      widthMm,
      heightMm,
      sheetFormat,
    }
  } else {
    const rows: any[] = []
    if (showName.value) {
      rows.push({
        id: 'r_name',
        type: 'cells',
        field: 'name',
        label: '',
        cells: [{ id: 'c_name', field: 'name', label: '', bold: true, fontSize: 7.5, align: 'left', bgColor: 'white' }],
      })
    }
    if (showSKU.value) {
      rows.push({
        id: 'r_sku',
        type: 'cells',
        field: 'sku',
        label: '',
        cells: [{ id: 'c_sku', field: 'sku', label: '', bold: true, fontSize: 6.5, align: 'left', bgColor: 'white' }],
      })
    }
    if (showLoc.value) {
      rows.push({
        id: 'r_loc',
        type: 'cells',
        field: 'location',
        label: '',
        cells: [{ id: 'c_loc', field: 'location', label: '', bold: false, fontSize: 6, align: 'left', bgColor: 'white' }],
      })
    }
    if (rows.length === 0) {
      rows.push({
        id: 'r_name',
        type: 'cells',
        field: 'name',
        label: '',
        cells: [{ id: 'c_name', field: 'name', label: '', bold: true, fontSize: 7.5, align: 'left', bgColor: 'white' }],
      })
    }

    newTpl = {
      id: customId,
      name: cleanName,
      widthMm,
      heightMm,
      sheetFormat,
      layoutMode: 'table',
      elements: [],
      config: {
        baseLayout: layout.value,
        fmt: fmt.value,
        labelTitle: labelTitle.value || 'INVENTARIS KANTOR',
        showName: showName.value,
        showSKU: showSKU.value,
        showLoc: showLoc.value,
        showBorder: showBorder.value,
        copies: copies.value,
      },
      tableConfig: {
        borderWidth: 1,
        borderStyle: 'solid',
        borderColor: '#cbd5e1',
        showOuterBorder: showBorder.value,
        showGridLines: false,
        showColBorder: false,
        showLabelColBorder: false,
        showHeaderBorder: true,
        showFooterBorder: false,
        cellPaddingMm: 0.7,
        labelColWidthPercent: 0,
        showHeader: true,
        headerText: labelTitle.value || 'INVENTARIS KANTOR',
        headerSubText: '',
        headerFontSize: 6.5,
        headerStyle: 'standard',
        showLogo: false,
        logoUrl: '',
        logoWidthMm: 10,
        logoPosition: 'above_qr',
        showItemNameUnderLogo: false,
        qrPosition: 'left',
        qrType: fmt.value,
        qrWidthMm: Math.min(20, Math.max(12, Math.round(heightMm * 0.62))),
        showFooter: false,
        footerText: '',
        footerFontSize: 6,
        defaultFontSize: 6.5,
        rows,
      },
    }
  }
  ensureTableConfig(newTpl)

  try {
    try {
      await api.post('/barcode/templates', newTpl)
    } catch (err) {
      console.warn('Disimpan di penyimpanan lokal browser:', err)
    }

    await loadCustomTemplates(newTpl)
    saveTemplateDialog.value = false
    toast.add({
      severity: 'success',
      summary: 'Template Disimpan',
      detail: `"${cleanName}" berhasil ditambahkan ke dropdown`,
      life: 3000,
    })
  } finally {
    savingTemplate.value = false
  }
}

function deleteSelectedCustomTemplate() {
  if (!selectedCustomTemplate.value) return
  const tpl = selectedCustomTemplate.value
  const cleanName = (tpl.name || '').replace(/^[🌟🏷️🖨️📄⭐\s]+/, '')
  confirm.require({
    message: `Hapus template tersimpan "${cleanName}" dari dropdown?`,
    header: 'Hapus Template Label',
    icon: 'pi pi-trash',
    acceptClass: 'p-button-danger',
    acceptLabel: 'Hapus',
    rejectLabel: 'Batal',
    accept: async () => {
      customTemplates.value = customTemplates.value.filter((t) => t.id !== tpl.id)
      try {
        localStorage.setItem('custom_label_templates_v6', JSON.stringify(customTemplates.value))
      } catch {}

      try {
        await api.delete(`/barcode/templates/${tpl.id}`)
      } catch (err) {
        console.warn('Gagal hapus dari database server:', err)
      }

      layout.value = 'tj_103'
      toast.add({
        severity: 'info',
        summary: 'Template Dihapus',
        detail: `Template "${cleanName}" telah dihapus`,
        life: 2500,
      })
    },
  })
}

const showName = ref(true)
const showSKU = ref(true)
const showLoc = ref(true)
const showBorder = ref(true)
const copies = ref(1)
const offset = ref(0)
const labelTitle = ref('INVENTARIS KANTOR')
const printModalVisible = ref(false)

const customSheetSpec = computed(() => {
  if (!selectedCustomTemplate.value) return null
  return getSheetSpec(
    inferSheetFormat(selectedCustomTemplate.value),
    selectedCustomTemplate.value.widthMm || 64,
    selectedCustomTemplate.value.heightMm || 32
  )
})

const isSheetLayout = computed(() => {
  if (selectedCustomTemplate.value && customSheetSpec.value) {
    return customSheetSpec.value.isSheet
  }
  return ['tj_103', 'tj_107', 'tj_108', 'tj_121', 'a4_3x7', 'a4_4x10'].includes(layout.value)
})

const labelsPerSheet = computed(() => {
  if (selectedCustomTemplate.value && customSheetSpec.value) {
    return customSheetSpec.value.perSheet
  }
  switch (layout.value) {
    case 'tj_103': return 12
    case 'tj_107': return 24
    case 'tj_108': return 40
    case 'tj_121': return 10
    case 'a4_3x7': return 21
    case 'a4_4x10': return 40
    default: return 1
  }
})

const printerProfileLabel = computed(() => {
  if (selectedCustomTemplate.value && customSheetSpec.value) {
    return customSheetSpec.value.isSheet ? customSheetSpec.value.shortLabel : 'Desain Kustom'
  }
  if (layout.value.startsWith('tj_')) return 'Lembar Tom & Jerry'
  if (layout.value.startsWith('thermal')) return 'Printer Thermal'
  if (layout.value.startsWith('a4')) return 'Kertas A4'
  return 'Stiker Rol'
})

watch(labelsPerSheet, (maxPerSheet) => {
  if (offset.value >= maxPerSheet) {
    offset.value = Math.max(0, maxPerSheet - 1)
  }
})
const totalLabelsToPrint = computed(() => selected.value.length * Math.max(1, copies.value || 1))
const estimatedSheets = computed(() => {
  if (!isSheetLayout.value) return null
  const totalWithOffset = totalLabelsToPrint.value + (offset.value || 0)
  return Math.ceil(totalWithOffset / labelsPerSheet.value)
})

const categoriesRaw = ref<any[]>([])
const locationsRaw = ref<any[]>([])

function rememberItems(list: any[]) {
  if (!Array.isArray(list) || list.length === 0) return
  const next = { ...selectedItemsMap.value }
  for (const it of list) {
    if (it && it.id) {
      next[it.id] = it
    }
  }
  selectedItemsMap.value = next
}

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
    rememberItems(items.value)
    total.value = res.data?.total || 0
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await Promise.all([fetchItems(), loadCustomTemplates()])
  const [c, l] = await Promise.all([api.get('/categories'), api.get('/locations')])
  categoriesRaw.value = c.data || []
  locationsRaw.value = l.data || []
})

// Debounced live search
let searchDebounce: any = null
watch(q, () => {
  clearTimeout(searchDebounce)
  searchDebounce = setTimeout(() => {
    fetchItems(true)
  }, 350)
})

watch([categoryFilter, locationFilter, perPage], () => fetchItems(true))
watch(page, () => fetchItems())

const lastPage = computed(() => Math.max(0, Math.ceil(total.value / perPage.value) - 1))
const range = computed(() => {
  if (!total.value) return '0 barang'
  const from = page.value * perPage.value + 1
  const to = Math.min(total.value, (page.value + 1) * perPage.value)
  return `${from}–${to} dari ${total.value} barang`
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

const selectingAllGlobal = ref(false)
async function selectAllGlobal() {
  selectingAllGlobal.value = true
  try {
    const batchSize = 500
    const collected: any[] = []
    let currentPage = 1
    let expectedTotal = Infinity

    while (collected.length < expectedTotal) {
      const res = await api.get('/items', {
        params: {
          q: q.value,
          cat: categoryFilter.value,
          loc: locationFilter.value,
          page: currentPage,
          per_page: batchSize,
        },
      })
      const batch: any[] = res.data?.data || []
      const resTotal = Number(res.data?.total ?? batch.length)
      if (Number.isFinite(resTotal)) expectedTotal = resTotal
      collected.push(...batch)
      if (batch.length < batchSize) break
      currentPage++
    }

    rememberItems(collected)
    const allIds = collected.map((i: any) => i.id)
    selected.value = Array.from(new Set([...selected.value, ...allIds]))
    toast.add({
      severity: 'success',
      summary: 'Seluruh Label Dipilih',
      detail: `${selected.value.length} label dipilih untuk siap dicetak`,
      life: 2500,
    })
  } catch {
    toast.add({ severity: 'error', summary: 'Gagal memilih semua barang', life: 2500 })
  } finally {
    selectingAllGlobal.value = false
  }
}

async function ensureSelectedItemsLoaded(): Promise<any[]> {
  const missingIds = selected.value.filter((id) => !selectedItemsMap.value[id])
  if (missingIds.length > 0) {
    try {
      const batchSize = 500
      let currentPage = 1
      let expectedTotal = Infinity
      let fetched = 0
      while (fetched < expectedTotal) {
        const res = await api.get('/items', {
          params: { page: currentPage, per_page: batchSize },
        })
        const batch: any[] = res.data?.data || []
        const resTotal = Number(res.data?.total ?? batch.length)
        if (Number.isFinite(resTotal)) expectedTotal = resTotal
        rememberItems(batch)
        fetched += batch.length
        if (batch.length < batchSize || selected.value.every((id) => Boolean(selectedItemsMap.value[id]))) {
          break
        }
        currentPage++
      }
    } catch (err) {
      console.warn('Gagal melengkapi data barang terpilih:', err)
    }
  }

  return selected.value.map((id) => {
    return (
      selectedItemsMap.value[id] ||
      items.value.find((i) => i.id === id) || {
        id,
        name: id,
        sku: id,
      }
    )
  })
}

const designerItems = computed(() => {
  const byId = new Map<string, any>()
  for (const it of items.value) {
    if (it?.id) byId.set(it.id, it)
  }
  for (const id of selected.value) {
    const cached = selectedItemsMap.value[id]
    if (cached && !byId.has(id)) {
      byId.set(id, cached)
    }
  }
  return Array.from(byId.values())
})

function clearSelection() {
  selected.value = []
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
    show_border: showBorder.value ? '1' : '0',
    copies: String(Math.max(1, copies.value || 1)),
    offset: String(Math.max(0, offset.value || 0)),
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

async function printNow() {
  if (layout.value.startsWith('custom_')) {
    const tpl = selectedCustomTemplate.value
    if (!tpl) {
      toast.add({
        severity: 'error',
        summary: 'Template Kustom Tidak Ditemukan',
        detail: 'Pilih ulang template dari dropdown atau buka Desainer Kustom.',
        life: 3000,
      })
      return
    }
    printingCustom.value = true
    try {
      const targetItems = await ensureSelectedItemsLoaded()
      await printCustomLabels(tpl, targetItems, {
        copies: Math.max(1, copies.value || 1),
        offset: isSheetLayout.value ? Math.max(0, offset.value || 0) : 0,
      })
    } catch (err: any) {
      toast.add({
        severity: 'error',
        summary: 'Gagal Mencetak Template Kustom',
        detail: err?.message || 'Tidak dapat menyiapkan dokumen cetak.',
        life: 3000,
      })
    } finally {
      printingCustom.value = false
    }
    return
  }
  const url = getPrintUrl(true)
  window.open(url, '_blank', 'noopener,noreferrer')
}

async function openSheetTab() {
  if (layout.value.startsWith('custom_')) {
    const tpl = selectedCustomTemplate.value
    if (!tpl) {
      toast.add({
        severity: 'error',
        summary: 'Template Kustom Tidak Ditemukan',
        detail: 'Pilih ulang template dari dropdown atau buka Desainer Kustom.',
        life: 3000,
      })
      return
    }
    const win = window.open('', '_blank')
    if (!win) {
      toast.add({
        severity: 'warn',
        summary: 'Popup Diblokir Browser',
        detail: 'Izinkan popup untuk membuka pratinjau cetak di tab baru.',
        life: 3500,
      })
      return
    }
    win.document.write('<p style="font-family:sans-serif;padding:24px;color:#475569">Menyiapkan pratinjau label...</p>')
    printingCustom.value = true
    try {
      const targetItems = await ensureSelectedItemsLoaded()
      openCustomPrintTab(win, tpl, targetItems, {
        copies: Math.max(1, copies.value || 1),
        offset: isSheetLayout.value ? Math.max(0, offset.value || 0) : 0,
      })
    } catch (err: any) {
      try { win.close() } catch {}
      toast.add({
        severity: 'error',
        summary: 'Gagal Membuka Pratinjau',
        detail: err?.message || 'Tidak dapat menyiapkan pratinjau cetak.',
        life: 3000,
      })
    } finally {
      printingCustom.value = false
    }
    return
  }
  const url = getPrintUrl(false)
  window.open(url, '_blank', 'noopener,noreferrer')
}

const customPreviewHtml = computed(() => {
  if (!selectedCustomTemplate.value) return ''
  return renderCustomLabelHtml(selectedCustomTemplate.value, sampleItem.value, '')
})

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
        <div class="flex items-center gap-2">
          <Button
            label="Desain Kustom (Drag &amp; Drop)"
            icon="pi pi-palette"
            size="small"
            severity="info"
            outlined
            @click="openCustomDesignerWithTemplate()"
          />
          <Tag
            v-if="customTemplates.length > 0"
            severity="info"
            class="cursor-pointer !text-xs !py-1 !px-2 font-semibold shadow-2xs hover:opacity-90"
            v-tooltip.top="'Buka desain kustom tersimpan'"
            @click="openCustomDesignerWithTemplate(customTemplates[0].id)"
          >
            <i class="pi pi-check-circle mr-1 text-[10px]" />
            {{ customTemplates.length }} Desain Kustom
          </Tag>
          <Button
            label="Pratinjau &amp; Cetak Label"
            icon="pi pi-print"
            size="small"
            :disabled="!selected.length"
            @click="openPrintDialog"
          />
        </div>
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
        featured
        :hint="selected.length ? selected.length + ' label siap dicetak' : 'Centang barang di bawah'"
      />
      <StatCard
        label="Profil Printer / Label"
        :value="printerProfileLabel"
        icon="pi pi-print"
        tone="info"
        :hint="layoutOptions.find(o => o.value === layout)?.label || layout"
      />
    </div>

    <!-- Unified Toolbar -->
    <div class="panel p-3 mb-4 flex flex-wrap items-center justify-between gap-2.5">
      <div class="flex flex-wrap items-center gap-2 flex-1 min-w-[280px]">
        <div class="relative flex-1 min-w-[190px] max-w-sm">
          <i class="pi pi-search absolute left-3 top-1/2 -translate-y-1/2 text-txt-dim text-xs" aria-hidden="true" />
          <InputText
            v-model="q"
            placeholder="Cari nama, SKU, atau merk…"
            class="w-full !pl-8 !pr-7 !text-[12px] !py-1.5"
          />
          <button
            v-if="q"
            type="button"
            aria-label="Hapus pencarian"
            class="absolute right-2 top-1/2 -translate-y-1/2 text-txt-dim hover:text-txt text-xs p-1 cursor-pointer"
            @click="q = ''"
          >
            <i class="pi pi-times" aria-hidden="true" />
          </button>
        </div>

        <Select
          v-model="categoryFilter"
          :options="categoriesRaw"
          optionLabel="name"
          optionValue="name"
          showClear
          placeholder="Semua Kategori"
          class="!text-[12px] !py-0.5 w-[160px]"
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

        <div class="flex items-center gap-1">
          <Select
            v-model="layout"
            :options="layoutOptions"
            optionLabel="label"
            optionValue="value"
            class="!text-[12px] !py-0.5 w-[250px] sm:w-[280px]"
            v-tooltip.top="'Pilih ukuran label / template tersimpan'"
          />
          <Button
            icon="pi pi-bookmark"
            size="small"
            severity="secondary"
            outlined
            class="!p-1.5 !text-xs shrink-0"
            v-tooltip.top="'Simpan profil label saat ini ke dropdown'"
            @click="saveCurrentLayoutAsTemplate"
          />
          <Button
            v-if="selectedCustomTemplate"
            icon="pi pi-pencil"
            size="small"
            severity="info"
            outlined
            class="!p-1.5 !text-xs shrink-0"
            aria-label="Edit template di Desainer Kustom"
            v-tooltip.top="'Buka & edit di Desainer Kustom'"
            @click="openCustomDesignerWithTemplate(selectedCustomTemplate.id)"
          />
          <Button
            v-if="selectedCustomTemplate"
            icon="pi pi-trash"
            size="small"
            severity="danger"
            outlined
            class="!p-1.5 !text-xs shrink-0"
            aria-label="Hapus template dari dropdown"
            v-tooltip.top="'Hapus template ini dari dropdown'"
            @click="deleteSelectedCustomTemplate"
          />
        </div>

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

      <div class="flex items-center gap-2 flex-wrap">
        <Button
          :label="isAllOnPageSelected ? 'Batal pilih hal. ini' : 'Pilih semua hal. ini'"
          icon="pi pi-check-square"
          text
          size="small"
          severity="secondary"
          @click="toggleAllOnPage"
        />
        <Button
          v-if="total > items.length"
          :label="`Pilih Semua (${total} barang)`"
          icon="pi pi-check-circle"
          text
          size="small"
          severity="info"
          :loading="selectingAllGlobal"
          @click="selectAllGlobal"
        />
        <Button
          v-if="selected.length"
          label="Kosongkan"
          icon="pi pi-times"
          text
          size="small"
          severity="danger"
          @click="clearSelection"
        />
        <Tag :value="selected.length + ' dipilih'" severity="warn" class="!text-[11px]" />
        <Button
          label="Desain Kustom"
          icon="pi pi-palette"
          size="small"
          severity="secondary"
          outlined
          @click="openCustomDesignerWithTemplate()"
        />
        <Button
          label="Cetak"
          icon="pi pi-print"
          size="small"
          :disabled="!selected.length"
          @click="openPrintDialog"
        />
      </div>
    </div>

    <EmptyState v-if="loading" loading title="Memuat barcode barang…" />
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
          class="panel relative overflow-hidden transition-colors cursor-pointer flex flex-col justify-between"
          :class="selected.includes(it.id) ? 'ring-2 ring-acc-500 !border-acc-500' : 'hover:border-[var(--acc-border)]'"
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
              :aria-label="'Pilih ' + it.name"
              @click.stop
              @update:modelValue="() => selected.includes(it.id) ? selected.splice(selected.indexOf(it.id), 1) : selected.push(it.id)"
            />
            <Button
              icon="pi pi-download"
              text
              rounded
              size="small"
              severity="secondary"
              aria-label="Unduh satu label"
              v-tooltip.top="'Unduh satu label'"
              @click.stop="downloadOne(it)"
            />
          </div>

          <div class="px-3 pb-3 text-center">
            <div class="grid place-items-center rounded-[8px] border border-line-soft p-2 mb-2 bg-white">
              <img
                :src="`/api/barcode/${it.id}.png?fmt=${fmt}&size=${size}`"
                class="max-w-full"
                :style="{ height: fmt === 'code128' ? '48px' : '76px' }"
                loading="lazy"
                :alt="it.sku"
              />
            </div>
            <div class="text-[12.5px] font-semibold truncate text-txt" :title="it.name">{{ it.name }}</div>
            <div class="t-mono text-[11px] mt-0.5 text-txt-dim">{{ it.sku }}</div>
            <div class="mt-2 flex items-center justify-center gap-1 flex-wrap">
              <Tag severity="secondary" :value="it.location || 'Tanpa Lokasi'" class="!text-[10px]" />
            </div>
          </div>
        </div>
      </div>

      <!-- Pagination Footer -->
      <div class="panel p-3 flex flex-wrap items-center justify-between gap-3 text-[12px]">
        <div class="flex items-center gap-2">
          <span class="t-mono !text-[11.5px] text-txt-dim">{{ range }}</span>
        </div>
        <div class="flex items-center gap-2">
          <Select
            v-model="perPage"
            :options="perPageOptions"
            optionLabel="label"
            optionValue="value"
            class="!text-[12px] !py-0.5 w-[135px]"
          />
          <Button
            icon="pi pi-angle-left"
            size="small"
            text
            severity="secondary"
            aria-label="Halaman sebelumnya"
            :disabled="page === 0"
            @click="page--"
          />
          <span class="t-num font-semibold px-1 text-txt">Hal. {{ page + 1 }} / {{ lastPage + 1 }}</span>
          <Button
            icon="pi pi-angle-right"
            size="small"
            text
            severity="secondary"
            aria-label="Halaman berikutnya"
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
        <div class="grid sm:grid-cols-2 gap-3.5 p-3 rounded-[10px] border border-line bg-panel-2">
          <div class="flex flex-col gap-1.5 sm:col-span-2">
            <div class="flex items-center justify-between gap-2 flex-wrap">
              <label class="text-[11.5px] font-semibold text-txt">Pilihan Ukuran Stiker / Jenis Printer</label>
              <button
                type="button"
                class="text-[11px] text-acc-500 hover:underline flex items-center gap-1 font-medium cursor-pointer"
                @click="saveCurrentLayoutAsTemplate"
              >
                <i class="pi pi-bookmark text-[10px]" aria-hidden="true" />
                <span>Simpan Pengaturan Ini ke Dropdown</span>
              </button>
            </div>
            <div class="flex items-center gap-1.5">
              <Select
                v-model="layout"
                :options="layoutOptions"
                optionLabel="label"
                optionValue="value"
                class="w-full !text-[12.5px]"
              />
              <Button
                v-if="selectedCustomTemplate"
                icon="pi pi-pencil"
                size="small"
                severity="info"
                outlined
                class="!p-2 shrink-0"
                aria-label="Edit template di Desainer Kustom"
                v-tooltip.top="'Edit template ini di Desainer Kustom'"
                @click="printModalVisible = false; openCustomDesignerWithTemplate(selectedCustomTemplate.id)"
              />
              <Button
                v-if="selectedCustomTemplate"
                icon="pi pi-trash"
                size="small"
                severity="danger"
                outlined
                class="!p-2 shrink-0"
                aria-label="Hapus template dari dropdown"
                v-tooltip.top="'Hapus template ini dari dropdown'"
                @click="deleteSelectedCustomTemplate"
              />
            </div>
          </div>

          <div v-if="!selectedCustomTemplate" class="flex flex-col gap-1.5">
            <label class="text-[11.5px] font-semibold text-txt">Format Barcode</label>
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

          <div v-if="!selectedCustomTemplate" class="flex flex-col gap-1.5">
            <label class="text-[11.5px] font-semibold text-txt">Teks Header Label</label>
            <InputText
              v-model="labelTitle"
              placeholder="INVENTARIS KANTOR"
              class="w-full !text-[12.5px]"
            />
          </div>

          <div class="flex flex-col gap-1.5">
            <label class="text-[11.5px] font-semibold text-txt">Jumlah Rangkap per Barang (Copies)</label>
            <InputNumber
              v-model="copies"
              :min="1"
              :max="50"
              showButtons
              class="w-full !text-[12.5px]"
              inputClass="!py-1.5 !text-[12.5px]"
            />
          </div>

          <div v-if="isSheetLayout" class="flex flex-col gap-1.5">
            <div class="flex items-center justify-between">
              <label class="text-[11.5px] font-semibold text-txt">Lewati Stiker Terpakai (Offset)</label>
              <span class="text-[10px] text-txt-dim">Slot kosong di awal</span>
            </div>
            <InputNumber
              v-model="offset"
              :min="0"
              :max="labelsPerSheet - 1"
              showButtons
              class="w-full !text-[12.5px]"
              inputClass="!py-1.5 !text-[12.5px]"
            />
          </div>

          <!-- Element Toggles (Built-in layouts only) -->
          <div v-if="!selectedCustomTemplate" class="flex flex-wrap items-center gap-4 sm:col-span-2 pt-2 border-t border-line">
            <label class="flex items-center gap-2 cursor-pointer text-[12px] text-txt">
              <Checkbox v-model="showName" binary />
              <span>Tampilkan Nama Barang</span>
            </label>
            <label class="flex items-center gap-2 cursor-pointer text-[12px] text-txt">
              <Checkbox v-model="showSKU" binary />
              <span>Tampilkan Kode SKU</span>
            </label>
            <label class="flex items-center gap-2 cursor-pointer text-[12px] text-txt">
              <Checkbox v-model="showLoc" binary />
              <span>Tampilkan Ruangan / Lokasi</span>
            </label>
            <label class="flex items-center gap-2 cursor-pointer text-[12px] text-txt">
              <Checkbox v-model="showBorder" binary />
              <span>Garis Batas Stiker (Border)</span>
            </label>
          </div>

          <!-- Bulk Summary Banner -->
          <div class="sm:col-span-2 p-2.5 rounded-[8px] border border-line bg-panel text-[11.5px] flex flex-wrap items-center justify-between gap-2">
            <div class="flex items-center gap-2 text-txt">
              <i class="pi pi-print text-acc-500 text-sm" aria-hidden="true" />
              <span><strong class="t-num">{{ selected.length }}</strong> barang &times; <strong class="t-num">{{ copies }}</strong> rangkap = <strong class="t-num">{{ totalLabelsToPrint }}</strong> label stiker</span>
            </div>
            <div v-if="isSheetLayout" class="font-semibold text-acc-500 t-num">
              Kebutuhan kertas: &plusmn; {{ estimatedSheets }} lembar ({{ labelsPerSheet }} label/lembar)
            </div>
          </div>
        </div>

        <!-- Live Visual Preview of Labels -->
        <div class="flex flex-col gap-2">
          <div class="flex items-center justify-between">
            <span class="t-label">
              {{
                selectedCustomTemplate
                  ? `Pratinjau Desain Kustom · ${customSheetSpec?.shortLabel || '1 Label / Halaman'}`
                  : layout === 'tj_103'
                    ? 'Pratinjau Lembar Tom & Jerry No. 103 (3 Kolom × 4 Baris)'
                    : 'Pratinjau Nyata Label'
              }}
            </span>
            <span class="t-mono !text-[11px] text-txt-dim">
              {{ totalLabelsToPrint }} label akan dicetak
            </span>
          </div>

          <div class="p-4 rounded-[10px] border border-line grid place-items-center bg-panel-2 overflow-x-auto">
            <!-- 0. Pratinjau Desain Kustom (Live HTML Render) -->
            <div
              v-if="layout.startsWith('custom_') && selectedCustomTemplate"
              class="w-full flex flex-col items-center gap-3"
            >
              <div class="w-full max-w-[560px] flex items-center justify-between gap-2 text-xs font-bold text-txt">
                <span class="flex items-center gap-1.5 truncate">
                  <i class="pi pi-palette text-acc-500" aria-hidden="true" />
                  <span class="truncate">{{ (selectedCustomTemplate.name || 'Desain Kustom').replace(/^[🌟🏷️🖨️📄⭐\s]+/, '') }}</span>
                </span>
                <div class="flex items-center gap-1.5 shrink-0">
                  <span class="t-mono !text-[11px] text-acc-500 bg-panel px-2 py-0.5 rounded border border-line">
                    {{ selectedCustomTemplate.widthMm }} &times; {{ selectedCustomTemplate.heightMm }} mm
                  </span>
                  <Button
                    label="Edit Template"
                    icon="pi pi-pencil"
                    size="small"
                    severity="secondary"
                    outlined
                    class="!text-[11px] !py-0.5 !px-2"
                    @click="printModalVisible = false; openCustomDesignerWithTemplate(selectedCustomTemplate.id)"
                  />
                </div>
              </div>

              <div
                v-if="customSheetSpec?.isSheet"
                class="rounded-[10px] shadow-sm border p-4 max-w-full"
                :class="customSheetSpec.isYellowBacking ? 'border-amber-300' : 'border-line bg-panel'"
                :style="customSheetSpec.isYellowBacking ? { background: '#fef08a', width: '560px' } : { width: '560px' }"
              >
                <div
                  class="flex items-center justify-between pb-2.5 mb-2.5 border-b text-[11px] font-extrabold uppercase tracking-wider"
                  :class="customSheetSpec.isYellowBacking ? 'border-amber-300/80 text-amber-950' : 'border-line text-txt'"
                >
                  <span>{{ customSheetSpec.shortLabel }} ({{ selectedCustomTemplate.widthMm }} &times; {{ selectedCustomTemplate.heightMm }} mm)</span>
                  <span class="font-normal text-[10px] opacity-85">{{ customSheetSpec.perSheet }} Label / Lembar</span>
                </div>
                <div
                  class="grid gap-2"
                  :style="{ gridTemplateColumns: `repeat(${customSheetSpec.cols}, minmax(0, 1fr))` }"
                >
                  <div
                    v-for="idx in Math.min(customSheetSpec.perSheet, 12)"
                    :key="idx"
                    class="rounded-[5px] overflow-hidden flex items-center justify-center"
                    :style="{
                      aspectRatio: `${selectedCustomTemplate.widthMm || 64} / ${selectedCustomTemplate.heightMm || 32}`,
                    }"
                    :class="
                      idx - 1 < offset
                        ? 'border-2 border-dashed border-amber-400/70 bg-amber-100/50 flex-col text-amber-800'
                        : idx - 1 - offset < totalLabelsToPrint
                          ? 'bg-white shadow-sm border border-slate-300 text-black'
                          : 'border border-dashed border-amber-400/40 bg-amber-100/20 text-amber-900/50'
                    "
                  >
                    <template v-if="idx - 1 < offset">
                      <span class="text-[9px] font-bold">Slot #{{ idx }}</span>
                      <span class="text-[8px] opacity-80">(Dilewati)</span>
                    </template>
                    <template v-else-if="idx - 1 - offset < totalLabelsToPrint">
                      <div
                        class="w-full h-full flex items-center justify-center overflow-hidden"
                        :style="{
                          zoom: String(
                            Math.min(
                              1,
                              ((528 - (customSheetSpec.cols - 1) * 8) / customSheetSpec.cols) /
                                ((selectedCustomTemplate.widthMm || 64) * 3.7795)
                            )
                          ),
                        }"
                        v-html="customPreviewHtml"
                      />
                    </template>
                    <template v-else>
                      <span class="text-[8px]">Kosong</span>
                    </template>
                  </div>
                </div>
              </div>
              <div
                v-else
                class="rounded-[10px] shadow-sm border border-line bg-panel p-3 flex flex-col items-center gap-2 max-w-full"
              >
                <div
                  class="bg-white text-black shadow-sm overflow-hidden"
                  style="display: inline-block; line-height: 1;"
                  v-html="customPreviewHtml"
                />
              </div>

              <div class="text-[11px] text-txt-dim text-center">
                Klik <strong>Cetak Sekarang</strong> untuk langsung mencetak {{ totalLabelsToPrint }} label sesuai template &amp; format kertas ini.
              </div>
            </div>

            <!-- 1. Simulasi Lembar Tom & Jerry No. 103 (Yellow Backing Paper 3x4 Grid) -->
            <div
              v-else-if="layout === 'tj_103'"
              class="rounded-[10px] shadow-sm border border-amber-300 p-4 max-w-full"
              style="background: #fef08a; width: 560px;"
            >
              <div class="flex items-center justify-between pb-2.5 mb-2.5 border-b border-amber-300/80 text-[11px] font-extrabold text-amber-950 uppercase tracking-wider">
                <span>Tom &amp; Jerry No. 103 (64 &times; 32 mm)</span>
                <span class="font-normal text-[10px] text-amber-900">12 Label / Lembar</span>
              </div>
              <div class="grid grid-cols-3 gap-2">
                <div
                  v-for="idx in 12"
                  :key="idx"
                  class="h-[76px] rounded-[5px] transition-all overflow-hidden"
                  :class="idx - 1 < offset ? 'border-2 border-dashed border-amber-400/70 bg-amber-100/50 flex flex-col items-center justify-center text-amber-800' : 'bg-white shadow-sm border border-slate-300 p-1.5 flex items-center justify-between text-black'"
                >
                  <!-- Skipped / Offset Slot -->
                  <template v-if="idx - 1 < offset">
                    <span class="text-[9px] font-bold">Slot #{{ idx }}</span>
                    <span class="text-[8px] opacity-80">(Dilewati)</span>
                  </template>
                  <!-- Printed Slot -->
                  <template v-else-if="idx - 1 - offset < totalLabelsToPrint">
                    <div class="w-[38%] flex items-center justify-center p-0.5">
                      <img :src="`/api/barcode/${sampleItem.id}.png?fmt=${fmt}`" class="max-w-full max-h-[60px] object-contain" alt="code" />
                    </div>
                    <div class="w-[62%] flex flex-col justify-center pl-1.5 border-l border-slate-200">
                      <div class="text-[7.5px] font-extrabold uppercase border-b border-black pb-0.5 leading-tight truncate">{{ labelTitle }}</div>
                      <div v-if="showName" class="text-[8.5px] font-bold mt-0.5 line-clamp-1 leading-tight">{{ sampleItem.name }}</div>
                      <div v-if="showSKU" class="text-[7.5px] font-mono font-bold mt-0.5">{{ sampleItem.sku }}</div>
                      <div v-if="showLoc && sampleItem.location" class="text-[7px] text-slate-700 truncate">{{ sampleItem.location }}</div>
                    </div>
                  </template>
                  <!-- Unused Slot on sheet -->
                  <template v-else>
                    <div class="w-full h-full flex flex-col items-center justify-center text-[8.5px] text-slate-400">
                      <span>Slot #{{ idx }}</span>
                      <span class="text-[7.5px] opacity-70">(Kosong)</span>
                    </div>
                  </template>
                </div>
              </div>
            </div>

            <!-- 2. Stiker Label 50x30 mm -->
            <div
              v-else-if="layout === 'label_50x30'"
              class="bg-white text-black p-2 rounded shadow-sm border border-slate-300 flex items-center justify-between"
              style="width: 280px; height: 168px;"
            >
              <div class="w-1/2 flex items-center justify-center p-1">
                <img :src="`/api/barcode/${sampleItem.id}.png?fmt=${fmt}`" class="max-w-full max-h-full object-contain" alt="code" />
              </div>
              <div class="w-1/2 flex flex-col justify-center pl-2 border-l border-slate-200">
                <div class="text-[9px] font-extrabold uppercase border-b border-black pb-0.5 tracking-wider">{{ labelTitle }}</div>
                <div v-if="showName" class="text-[11px] font-bold mt-1 line-clamp-2 leading-tight">{{ sampleItem.name }}</div>
                <div v-if="showSKU" class="text-[10px] font-mono font-bold mt-1">{{ sampleItem.sku }}</div>
                <div v-if="showLoc && sampleItem.location" class="text-[9px] text-slate-700 mt-0.5 truncate">{{ sampleItem.location }}</div>
              </div>
            </div>

            <!-- 3. Stiker Label 40x30 mm -->
            <div
              v-else-if="layout === 'label_40x30'"
              class="bg-white text-black p-1.5 rounded shadow-sm border border-slate-300 flex items-center justify-between"
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

            <!-- 4. Thermal Roll 58 mm -->
            <div
              v-else-if="layout === 'thermal_58'"
              class="bg-white text-black p-3 rounded shadow-sm border border-slate-300 text-center flex flex-col items-center"
              style="width: 260px;"
            >
              <div class="text-[10px] font-bold uppercase tracking-wider border-b border-black pb-1 w-full">{{ labelTitle }}</div>
              <div class="my-2">
                <img :src="`/api/barcode/${sampleItem.id}.png?fmt=${fmt}`" :style="{ height: fmt === 'code128' ? '40px' : '65px' }" alt="code" />
              </div>
              <div v-if="showName" class="text-[11px] font-bold leading-tight">{{ sampleItem.name }}</div>
              <div v-if="showSKU" class="text-[10px] font-mono font-bold mt-0.5">{{ sampleItem.sku }}</div>
              <div v-if="showLoc && sampleItem.location" class="text-[9px] text-slate-700 mt-0.5">{{ sampleItem.location }}</div>
              <div class="w-full border-b border-dashed border-slate-400 mt-2" />
            </div>

            <!-- 5. Thermal Roll 80 mm -->
            <div
              v-else-if="layout === 'thermal_80'"
              class="bg-white text-black p-4 rounded shadow-sm border border-slate-300 text-center flex flex-col items-center"
              style="width: 320px;"
            >
              <div class="text-[11px] font-extrabold uppercase tracking-wider border-b-2 border-black pb-1 w-full">{{ labelTitle }}</div>
              <div class="my-2.5">
                <img :src="`/api/barcode/${sampleItem.id}.png?fmt=${fmt}`" :style="{ height: fmt === 'code128' ? '50px' : '85px' }" alt="code" />
              </div>
              <div v-if="showName" class="text-[12.5px] font-bold leading-tight">{{ sampleItem.name }}</div>
              <div v-if="showSKU" class="text-[11px] font-mono font-bold mt-1">{{ sampleItem.sku }}</div>
              <div v-if="showLoc && sampleItem.location" class="text-[10px] text-slate-700 mt-0.5">{{ sampleItem.location }}</div>
              <div class="w-full border-b border-dashed border-slate-400 mt-3" />
            </div>

            <!-- 6. Stiker Lebar 60x40 / 70x50 / 100x50 / A4 / TJ lainnya -->
            <div
              v-else
              class="bg-white text-black p-3 rounded shadow-sm border border-slate-300 flex items-center justify-between"
              style="width: 340px; height: 180px;"
            >
              <div class="w-1/2 flex items-center justify-center p-2">
                <img :src="`/api/barcode/${sampleItem.id}.png?fmt=${fmt}`" class="max-w-full max-h-full object-contain" alt="code" />
              </div>
              <div class="w-1/2 flex flex-col justify-center pl-3 border-l border-slate-200">
                <div class="text-[10px] font-extrabold uppercase border-b border-black pb-1 tracking-wider">{{ labelTitle }}</div>
                <div v-if="showName" class="text-[12px] font-bold mt-1.5 line-clamp-2 leading-tight">{{ sampleItem.name }}</div>
                <div v-if="showSKU" class="text-[11px] font-mono font-bold mt-1">{{ sampleItem.sku }}</div>
                <div v-if="showLoc && sampleItem.location" class="text-[10px] text-slate-700 mt-0.5 truncate">{{ sampleItem.location }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="flex items-center justify-between w-full pt-2 flex-wrap gap-2">
          <Button
            label="Batal"
            text
            severity="secondary"
            size="small"
            @click="printModalVisible = false"
          />
          <div class="flex items-center gap-2 flex-wrap">
            <Button
              label="Desain Kustom"
              icon="pi pi-palette"
              size="small"
              severity="secondary"
              outlined
              @click="printModalVisible = false; openCustomDesignerWithTemplate(selectedCustomTemplate?.id)"
            />
            <Button
              label="Buka di Tab Baru"
              icon="pi pi-external-link"
              size="small"
              severity="secondary"
              outlined
              :loading="printingCustom"
              @click="openSheetTab"
            />
            <Button
              label="Cetak Sekarang"
              icon="pi pi-print"
              size="small"
              :loading="printingCustom"
              @click="printNow"
            />
          </div>
        </div>
      </template>
    </Dialog>

    <!-- Save Template Name Dialog (replaces native prompt) -->
    <Dialog
      v-model:visible="saveTemplateDialog"
      modal
      header="Simpan Pengaturan ke Dropdown"
      :style="{ width: '420px' }"
    >
      <form @submit.prevent="confirmSaveLayoutAsTemplate" class="flex flex-col gap-3 pt-1">
        <label class="text-[12px] text-txt-dim">
          Beri nama template label agar mudah dipilih kembali dari dropdown cetak:
        </label>
        <InputText
          v-model="saveTemplateName"
          placeholder="mis. Stiker Aset Ruangan 64x32"
          class="w-full"
          autofocus
        />
      </form>
      <template #footer>
        <Button label="Batal" text severity="secondary" size="small" @click="saveTemplateDialog = false" />
        <Button
          label="Simpan Template"
          icon="pi pi-bookmark"
          size="small"
          :loading="savingTemplate"
          @click="confirmSaveLayoutAsTemplate"
        />
      </template>
    </Dialog>

    <!-- Custom Visual Drag-and-Drop Label Designer -->
    <LabelDesignerModal
      v-model:visible="showCustomDesigner"
      :items="designerItems"
      :selectedIds="selected"
      :initialTemplateId="selectedCustomTemplateId"
      :initialCopies="copies"
      :initialOffset="offset"
      @template-saved="onTemplateSaved"
    />
  </div>
</template>
