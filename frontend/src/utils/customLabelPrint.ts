export type SheetFormatId =
  | 'single'
  | 'tj_103'
  | 'tj_107'
  | 'tj_108'
  | 'tj_121'
  | 'a4_3x7'
  | 'a4_4x10'
  | 'auto_a4'

export interface LabelElement {
  id: string
  type: 'qr' | 'code128' | 'field' | 'static_text' | 'header' | 'line' | 'logo'
  field?: string
  prefix?: string
  text?: string
  logoUrl?: string
  x: number
  y: number
  w: number
  h: number
  fontSize?: number
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
  fontSize?: number
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
  fontSize?: number
  align?: 'left' | 'center' | 'right'
  bgColor?: 'white' | 'light_gray' | 'dark'
  hasBottomBorder?: boolean
  cells?: TableCellConfig[]
}

export interface TableLayoutConfig {
  borderWidth: number
  borderStyle: 'solid' | 'dashed' | 'dotted' | 'double'
  borderColor: string
  showOuterBorder: boolean
  showGridLines: boolean
  showColBorder: boolean
  showLabelColBorder: boolean
  cellPaddingMm: number
  labelColWidthPercent: number
  showHeader: boolean
  headerText: string
  headerSubText?: string
  headerFontSize: number
  headerStyle: 'standard' | 'inverted' | 'double_line'
  showHeaderBorder?: boolean
  showLogo: boolean
  logoUrl?: string
  logoWidthMm: number
  logoPosition?: 'above_qr' | 'above_qr_only' | 'top_left' | 'header'
  showItemNameUnderLogo?: boolean
  qrPosition: 'left' | 'right' | 'none'
  qrType: 'qr' | 'code128'
  qrWidthMm: number
  showFooter: boolean
  footerText: string
  footerFontSize: number
  showFooterBorder?: boolean
  defaultFontSize: number
  rows: TableRowConfig[]
}

export interface LabelTemplate {
  id: string
  name: string
  widthMm: number
  heightMm: number
  layoutMode: 'table' | 'free'
  sheetFormat?: SheetFormatId
  config?: Record<string, any>
  tableConfig: TableLayoutConfig
  elements: LabelElement[]
}

export interface SheetSpec {
  id: SheetFormatId
  label: string
  shortLabel: string
  isSheet: boolean
  cols: number
  rows: number
  perSheet: number
  labelW: number
  labelH: number
  rowGapMm: number
  colGapMm: number
  pageMargin: string
  isYellowBacking: boolean
}

export type PrintMode = 'sheet_tab' | 'direct'

export const PRINT_MODE_OPTIONS: { label: string; value: PrintMode }[] = [
  { label: 'Lembar Cetak Tab Baru (Selaras T&J / A4)', value: 'sheet_tab' },
  { label: 'Langsung Dialog Print Browser', value: 'direct' },
]

export function getPrintMode(tpl?: Partial<LabelTemplate> | null): PrintMode {
  return tpl?.config?.printMode === 'direct' ? 'direct' : 'sheet_tab'
}

export function getShowCutBorder(tpl?: Partial<LabelTemplate> | null): boolean {
  if (typeof tpl?.config?.showCutBorder === 'boolean') return tpl.config.showCutBorder
  // Default off when table already draws its own outer border; on otherwise
  if (tpl?.layoutMode === 'table' && tpl?.tableConfig?.showOuterBorder) return false
  return true
}

export const SHEET_FORMAT_OPTIONS: { label: string; value: SheetFormatId }[] = [
  { label: 'Tom & Jerry No. 103 (3×4 · 12/lembar · 64×32 mm)', value: 'tj_103' },
  { label: 'Tom & Jerry No. 107 (3×8 · 24/lembar · 50×18 mm)', value: 'tj_107' },
  { label: 'Tom & Jerry No. 108 (5×8 · 40/lembar · 38×18 mm)', value: 'tj_108' },
  { label: 'Tom & Jerry No. 121 (2×5 · 10/lembar · 75×38 mm)', value: 'tj_121' },
  { label: 'Stiker A4 3×7 (21/lembar · 65×38 mm)', value: 'a4_3x7' },
  { label: 'Stiker A4 4×10 (40/lembar · 48×25.5 mm)', value: 'a4_4x10' },
  { label: 'Grid A4 Otomatis (Sesuai Ukuran P×L)', value: 'auto_a4' },
  { label: '1 Label / Halaman (Thermal / Rol)', value: 'single' },
]

const VALID_SHEET_FORMATS = new Set<SheetFormatId>([
  'single',
  'tj_103',
  'tj_107',
  'tj_108',
  'tj_121',
  'a4_3x7',
  'a4_4x10',
  'auto_a4',
])

export function inferSheetFormat(tpl?: Partial<LabelTemplate> | null): SheetFormatId {
  if (!tpl) return 'single'
  if (tpl.sheetFormat && VALID_SHEET_FORMATS.has(tpl.sheetFormat)) {
    return tpl.sheetFormat
  }
  const cfgFormat = tpl.config?.sheetFormat || tpl.config?.baseLayout
  if (typeof cfgFormat === 'string') {
    if (VALID_SHEET_FORMATS.has(cfgFormat as SheetFormatId)) {
      return cfgFormat as SheetFormatId
    }
    if (cfgFormat.startsWith('thermal') || cfgFormat.startsWith('label_')) {
      return 'single'
    }
  }
  const w = Number(tpl.widthMm || 0)
  const h = Number(tpl.heightMm || 0)
  const hint = `${tpl.id || ''} ${tpl.name || ''}`
  if ((w === 64 && h === 32) || /tom\s*&?\s*jerry.*103|t&j\s*103|tj[\s_-]*103/i.test(hint)) return 'tj_103'
  if ((w === 50 && h === 18) || /tom\s*&?\s*jerry.*107|t&j\s*107|tj[\s_-]*107/i.test(hint)) return 'tj_107'
  if ((w === 38 && h === 18) || /tom\s*&?\s*jerry.*108|t&j\s*108|tj[\s_-]*108/i.test(hint)) return 'tj_108'
  if ((w === 75 && h === 38) || /tom\s*&?\s*jerry.*121|t&j\s*121|tj[\s_-]*121/i.test(hint)) return 'tj_121'
  if (w === 65 && h === 38) return 'a4_3x7'
  if (w === 48 && h === 25.5) return 'a4_4x10'
  if (/thermal|rol\b|roll\b|struk/i.test(hint)) return 'single'
  // Multi-label A4 grid fallback so saved sticker templates don't print 1 label per A4 sheet
  if (w > 0 && h > 0 && w <= 105 && h <= 140) return 'auto_a4'
  return 'single'
}

export function getSheetSpec(format: SheetFormatId, widthMm = 64, heightMm = 32): SheetSpec {
  const w = Math.max(15, Number(widthMm) || 64)
  const h = Math.max(10, Number(heightMm) || 32)
  switch (format) {
    case 'tj_103':
      return {
        id: 'tj_103',
        label: 'Tom & Jerry No. 103 (3×4 · 12/lembar)',
        shortLabel: 'Tom & Jerry No. 103 (3×4)',
        isSheet: true,
        cols: 3,
        rows: 4,
        perSheet: 12,
        labelW: 64,
        labelH: 32,
        rowGapMm: 3.5,
        colGapMm: 4,
        pageMargin: '6mm 5mm',
        isYellowBacking: true,
      }
    case 'tj_107':
      return {
        id: 'tj_107',
        label: 'Tom & Jerry No. 107 (3×8 · 24/lembar)',
        shortLabel: 'Tom & Jerry No. 107 (3×8)',
        isSheet: true,
        cols: 3,
        rows: 8,
        perSheet: 24,
        labelW: 50,
        labelH: 18,
        rowGapMm: 2,
        colGapMm: 3,
        pageMargin: '5mm',
        isYellowBacking: true,
      }
    case 'tj_108':
      return {
        id: 'tj_108',
        label: 'Tom & Jerry No. 108 (5×8 · 40/lembar)',
        shortLabel: 'Tom & Jerry No. 108 (5×8)',
        isSheet: true,
        cols: 5,
        rows: 8,
        perSheet: 40,
        labelW: 38,
        labelH: 18,
        rowGapMm: 2,
        colGapMm: 2,
        pageMargin: '5mm',
        isYellowBacking: true,
      }
    case 'tj_121':
      return {
        id: 'tj_121',
        label: 'Tom & Jerry No. 121 (2×5 · 10/lembar)',
        shortLabel: 'Tom & Jerry No. 121 (2×5)',
        isSheet: true,
        cols: 2,
        rows: 5,
        perSheet: 10,
        labelW: 75,
        labelH: 38,
        rowGapMm: 3,
        colGapMm: 4,
        pageMargin: '6mm 5mm',
        isYellowBacking: true,
      }
    case 'a4_3x7':
      return {
        id: 'a4_3x7',
        label: 'Stiker A4 3×7 (21/lembar)',
        shortLabel: 'Stiker A4 (3×7)',
        isSheet: true,
        cols: 3,
        rows: 7,
        perSheet: 21,
        labelW: 65,
        labelH: 38,
        rowGapMm: 2,
        colGapMm: 3,
        pageMargin: '8mm 5mm',
        isYellowBacking: false,
      }
    case 'a4_4x10':
      return {
        id: 'a4_4x10',
        label: 'Stiker A4 4×10 (40/lembar)',
        shortLabel: 'Stiker A4 (4×10)',
        isSheet: true,
        cols: 4,
        rows: 10,
        perSheet: 40,
        labelW: 48,
        labelH: 25.5,
        rowGapMm: 2,
        colGapMm: 2,
        pageMargin: '8mm 5mm',
        isYellowBacking: false,
      }
    case 'auto_a4': {
      const gap = 3
      const cols = Math.max(1, Math.floor((200 + gap) / (w + gap)))
      const rows = Math.max(1, Math.floor((285 + gap) / (h + gap)))
      return {
        id: 'auto_a4',
        label: `Grid A4 Otomatis (${cols}×${rows} · ${cols * rows}/lembar)`,
        shortLabel: `Grid A4 (${cols}×${rows})`,
        isSheet: true,
        cols,
        rows,
        perSheet: cols * rows,
        labelW: w,
        labelH: h,
        rowGapMm: gap,
        colGapMm: gap,
        pageMargin: '6mm 5mm',
        isYellowBacking: false,
      }
    }
    default:
      return {
        id: 'single',
        label: `1 Label / Halaman (${w}×${h} mm)`,
        shortLabel: `Thermal / Rol (${w}×${h} mm)`,
        isSheet: false,
        cols: 1,
        rows: 1,
        perSheet: 1,
        labelW: w,
        labelH: h,
        rowGapMm: 0,
        colGapMm: 0,
        pageMargin: '0',
        isYellowBacking: false,
      }
  }
}

export function getRowCells(row: TableRowConfig): TableCellConfig[] {
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
      { id: 'c1', field: row.field || 'name', label: row.label ?? '', bold: row.bold, fontSize: row.fontSize, align: row.align, bgColor: row.bgColor },
    ]
  }
  return row.cells
}

export function isLogoInHeader(tc?: Partial<TableLayoutConfig> | null): boolean {
  if (!tc || !tc.showLogo || !tc.showHeader) return false
  return tc.logoPosition !== 'above_qr_only'
}

export function ensureTableConfig(tpl: LabelTemplate, fallbackTableConfig?: TableLayoutConfig) {
  if (!tpl.tableConfig) {
    tpl.tableConfig = fallbackTableConfig
      ? JSON.parse(JSON.stringify(fallbackTableConfig))
      : {
          borderWidth: 1.5,
          borderStyle: 'solid',
          borderColor: '#000000',
          showOuterBorder: true,
          showGridLines: true,
          showColBorder: true,
          showLabelColBorder: true,
          cellPaddingMm: 0.8,
          labelColWidthPercent: 28,
          showHeader: true,
          headerText: 'PEMERINTAH KOTA CIREBON',
          headerSubText: 'PUSKESMAS KESAMBI',
          headerFontSize: 7,
          headerStyle: 'standard',
          showHeaderBorder: true,
          showLogo: true,
          logoUrl: '',
          logoWidthMm: 8,
          logoPosition: 'header',
          showItemNameUnderLogo: false,
          qrPosition: 'left',
          qrType: 'qr',
          qrWidthMm: 21,
          showFooter: false,
          footerText: '',
          footerFontSize: 5.5,
          showFooterBorder: true,
          defaultFontSize: 7,
          rows: [
            {
              id: 'r1',
              type: 'cells',
              field: 'name',
              label: '',
              bold: true,
              fontSize: 7.5,
              bgColor: 'white',
              cells: [{ id: 'c1', field: 'name', label: '', bold: true, fontSize: 7.5, align: 'left', bgColor: 'white' }],
            },
            {
              id: 'r2',
              type: 'cells',
              field: 'sku',
              label: 'SKU',
              bold: false,
              fontSize: 6.5,
              bgColor: 'white',
              cells: [{ id: 'c1', field: 'sku', label: 'SKU', bold: false, fontSize: 6.5, align: 'left', bgColor: 'white' }],
            },
            {
              id: 'r3',
              type: 'cells',
              field: 'location',
              label: 'Ruang',
              bold: false,
              fontSize: 6.5,
              bgColor: 'white',
              cells: [{ id: 'c1', field: 'location', label: 'Ruang', bold: false, fontSize: 6.5, align: 'left', bgColor: 'white' }],
            },
          ],
        }
  } else {
    const tc = tpl.tableConfig
    if (tc.borderWidth === undefined) tc.borderWidth = 1.5
    if (!tc.borderStyle) tc.borderStyle = 'solid'
    if (!tc.borderColor) tc.borderColor = '#000000'
    if (tc.showOuterBorder === undefined) tc.showOuterBorder = true
    if (tc.showGridLines === undefined) tc.showGridLines = true
    if (tc.showColBorder === undefined) tc.showColBorder = true
    if (tc.showLabelColBorder === undefined) tc.showLabelColBorder = true
    if (tc.cellPaddingMm === undefined) tc.cellPaddingMm = 0.8
    if (tc.labelColWidthPercent === undefined) tc.labelColWidthPercent = 28
    if (tc.headerSubText === undefined) tc.headerSubText = ''
    if (tc.headerSubText && tc.headerSubText.trim().toUpperCase() === 'PUSKESAMS KESAMBI') {
      tc.headerSubText = 'PUSKESMAS KESAMBI'
    }
    if (tc.headerFontSize === undefined) tc.headerFontSize = 7.5
    if (!tc.headerStyle) tc.headerStyle = 'standard'
    if (tc.showHeaderBorder === undefined) tc.showHeaderBorder = true
    if (tc.logoWidthMm === undefined) tc.logoWidthMm = 8
    if (!tc.logoPosition || (tc.showHeader && tc.logoPosition === 'above_qr')) {
      tc.logoPosition = tc.showHeader ? 'header' : 'above_qr'
    }
    if (tc.showItemNameUnderLogo === undefined) tc.showItemNameUnderLogo = false
    if (!tc.qrPosition) tc.qrPosition = 'left'
    if (!tc.qrType) tc.qrType = 'qr'
    if (tc.qrWidthMm === undefined) tc.qrWidthMm = 21
    if (tc.footerFontSize === undefined) tc.footerFontSize = 5.5
    if (tc.showFooterBorder === undefined) tc.showFooterBorder = true
    if (tc.defaultFontSize === undefined) tc.defaultFontSize = 7
    if (!Array.isArray(tc.rows)) tc.rows = []

    tc.rows.forEach((r) => {
      if (!r.type) r.type = 'cells'
      if (!r.bgColor) r.bgColor = 'white'
      getRowCells(r)
    })
  }

  if (!tpl.layoutMode) {
    tpl.layoutMode = 'table'
  }
  if (!tpl.sheetFormat) {
    tpl.sheetFormat = inferSheetFormat(tpl)
  }
}

export function resolveFieldValue(item: any, fieldKey?: string, prefix?: string): string {
  if (!fieldKey) return ''
  const isSample = !item || item.id === 'sample'
  let val = item?.[fieldKey]
  if (val === undefined || val === null || val === '') {
    if (isSample) {
      switch (fieldKey) {
        case 'name': val = item?.name || 'Nama Aset'; break
        case 'sku': val = item?.sku || 'AST-0001'; break
        case 'item_code': val = item?.item_code || '02.06.01.001'; break
        case 'register_number': val = item?.register_number || '0001'; break
        case 'location': val = item?.location || 'Gudang Utama'; break
        case 'location_code': val = item?.location_code || 'G.01'; break
        case 'merk': val = item?.merk || 'Standar'; break
        case 'type_model': val = item?.type_model || '-'; break
        case 'serial_number': val = item?.serial_number || 'SN-12345'; break
        case 'procurement_year': val = item?.procurement_year || '2024'; break
        case 'funding_source': val = item?.funding_source || 'APBD'; break
        case 'condition_status': val = item?.condition_status || 'Baik'; break
        case 'price_per_unit': val = item?.price_per_unit ? `Rp ${Number(item.price_per_unit).toLocaleString('id-ID')}` : '-'; break
        default: val = '-'
      }
    } else {
      switch (fieldKey) {
        case 'name': val = item?.name || '-'; break
        case 'sku': val = item?.sku || '-'; break
        case 'location': val = item?.location || '-'; break
        case 'condition_status': val = item?.condition_status || '-'; break
        default: val = '-'
      }
    }
  } else if (fieldKey === 'price_per_unit') {
    val = `Rp ${Number(val).toLocaleString('id-ID')}`
  }

  const pfx = prefix || ''
  return `${pfx}${val}`
}

function escapeHtml(raw: unknown): string {
  return String(raw ?? '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

function safeColor(raw: unknown, fallback = '#000000'): string {
  const s = String(raw ?? '').trim()
  if (/^#[0-9a-fA-F]{3,8}$/.test(s) || /^[a-zA-Z]{3,20}$/.test(s)) return s
  return fallback
}

function safeBorderStyle(raw: unknown): string {
  return raw === 'dashed' || raw === 'dotted' || raw === 'double' ? raw : 'solid'
}

function resolveAssetUrl(raw: string, origin: string): string {
  const s = String(raw || '').trim()
  if (!s) return ''
  if (s.startsWith('data:image/')) return s
  if (s.startsWith('/')) return `${origin}${s}`
  if (/^https?:\/\//i.test(s)) return s
  return ''
}

const DEFAULT_CREST_SVG = `<svg viewBox="0 0 100 120" style="max-width: 90%; max-height: 100%; object-fit: contain;">
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
</svg>`

function bgStyleFor(bg?: string): { bg: string; color: string } {
  if (bg === 'light_gray') return { bg: '#f1f5f9', color: '#000000' }
  if (bg === 'dark') return { bg: '#1e293b', color: '#ffffff' }
  return { bg: 'transparent', color: '#000000' }
}

function renderQrColumnHtml(tc: TableLayoutConfig, item: any, origin: string, side: 'left' | 'right'): string {
  const bw = Number(tc.borderWidth) || 1.5
  const bs = safeBorderStyle(tc.borderStyle)
  const bc = safeColor(tc.borderColor)
  const borderSide = side === 'left' ? 'border-right' : 'border-left'
  const colBorder = tc.showColBorder ? `${borderSide}: ${bw}pt ${bs} ${bc};` : ''
  const qrUrl = resolveAssetUrl(`/api/barcode/${encodeURIComponent(item.id)}.png?fmt=${tc.qrType === 'code128' ? 'code128' : 'qr'}`, origin)
  const logoUrl = tc.logoUrl ? resolveAssetUrl(tc.logoUrl, origin) : ''
  const logoMaxH = Math.max(5, (Number(tc.logoWidthMm) || 10) * 0.65)
  const showLogoAboveQr = tc.showLogo && !isLogoInHeader(tc)

  const logoHtml = showLogoAboveQr
    ? `<div style="width:100%;flex:0 0 auto;max-height:${logoMaxH}mm;display:flex;align-items:center;justify-content:center;padding:0.6mm 0.5mm;box-sizing:border-box;border-bottom:${bw}pt ${bs} ${bc};overflow:hidden;">
        ${logoUrl ? `<img src="${escapeHtml(logoUrl)}" style="max-width:95%;max-height:${logoMaxH - 1}mm;object-fit:contain;" alt="Logo" />` : DEFAULT_CREST_SVG}
      </div>`
    : ''

  const codeHtml =
    tc.qrType === 'qr'
      ? `<img src="${escapeHtml(qrUrl)}" style="width:98%;height:98%;max-width:100%;max-height:100%;object-fit:contain;display:block;" alt="QR" />`
      : `<div style="width:100%;height:100%;display:flex;flex-direction:column;align-items:center;justify-content:center;overflow:hidden;">
          <img src="${escapeHtml(qrUrl)}" style="width:96%;height:74%;object-fit:contain;display:block;" alt="BC" />
          <span style="font-family:monospace;font-weight:700;font-size:5.5pt;line-height:1;margin-top:0.4mm;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;max-width:100%;">${escapeHtml(item.sku || '')}</span>
        </div>`

  const nameUnderHtml = tc.showItemNameUnderLogo
    ? `<div style="width:100%;flex:0 0 auto;font-size:5pt;font-weight:700;text-align:center;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;padding:0.35mm 0.5mm;border-top:0.5pt solid #cbd5e1;box-sizing:border-box;">- ${escapeHtml(item.name || '')}</div>`
    : ''

  return `<div style="width:${Number(tc.qrWidthMm) || 20}mm;flex:0 0 auto;display:flex;flex-direction:column;align-items:center;justify-content:space-between;overflow:hidden;box-sizing:border-box;${colBorder}">
    ${logoHtml}
    <div style="width:100%;flex:1;min-height:0;display:flex;align-items:center;justify-content:center;padding:0.4mm;box-sizing:border-box;overflow:hidden;">
      ${codeHtml}
    </div>
    ${nameUnderHtml}
  </div>`
}

export function renderCustomLabelHtml(rawTpl: LabelTemplate, item: any, origin = typeof window !== 'undefined' ? window.location.origin : ''): string {
  const tpl: LabelTemplate = JSON.parse(JSON.stringify(rawTpl))
  ensureTableConfig(tpl)
  const w = Number(tpl.widthMm) || 64
  const h = Number(tpl.heightMm) || 32

  if (tpl.layoutMode === 'free') {
    const elsHtml = (tpl.elements || [])
      .map((el) => {
        const align = el.align || 'left'
        const justify = align === 'center' ? 'center' : align === 'right' ? 'flex-end' : 'flex-start'
        const basePos = `position:absolute;left:${Number(el.x) || 0}mm;top:${Number(el.y) || 0}mm;width:${Number(el.w) || 10}mm;height:${Number(el.h) || 5}mm;box-sizing:border-box;overflow:hidden;display:flex;align-items:center;`

        if (el.type === 'qr') {
          const src = resolveAssetUrl(`/api/barcode/${encodeURIComponent(item.id)}.png?fmt=qr`, origin)
          return `<div style="${basePos}justify-content:center;padding:0.3mm;background:#fff;"><img src="${escapeHtml(src)}" style="width:100%;height:100%;object-fit:contain;display:block;" alt="QR" /></div>`
        }
        if (el.type === 'code128') {
          const src = resolveAssetUrl(`/api/barcode/${encodeURIComponent(item.id)}.png?fmt=code128`, origin)
          return `<div style="${basePos}justify-content:center;padding:0.3mm;background:#fff;"><img src="${escapeHtml(src)}" style="width:100%;height:100%;object-fit:contain;display:block;" alt="BC" /></div>`
        }
        if (el.type === 'logo') {
          const logoSrc = el.logoUrl ? resolveAssetUrl(el.logoUrl, origin) : ''
          return `<div style="${basePos}justify-content:center;">${logoSrc ? `<img src="${escapeHtml(logoSrc)}" style="width:100%;height:100%;object-fit:contain;display:block;" alt="Logo" />` : DEFAULT_CREST_SVG}</div>`
        }
        if (el.type === 'header') {
          return `<div style="${basePos}background:#000;color:#fff;font-weight:700;justify-content:center;text-align:center;padding:0 0.5mm;font-size:${Number(el.fontSize) || 7}pt;line-height:1.15;-webkit-print-color-adjust:exact;print-color-adjust:exact;"><span style="width:100%;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;">${escapeHtml(el.text || 'INVENTARIS')}</span></div>`
        }
        if (el.type === 'line') {
          const lineCss =
            el.lineOrientation === 'vertical'
              ? 'height:100%;border-left:1pt solid #000;margin:0 auto;'
              : 'width:100%;border-top:1pt solid #000;'
          return `<div style="${basePos}justify-content:center;"><div style="${lineCss}"></div></div>`
        }
        const textVal = el.type === 'field' ? resolveFieldValue(item, el.field, el.prefix) : el.text || ''
        const borderCss = el.border ? 'border:0.5pt solid #000;' : ''
        const fontFam = el.field === 'sku' ? 'font-family:monospace;' : ''
        return `<div style="${basePos}padding:0 0.5mm;justify-content:${justify};text-align:${align};font-size:${Number(el.fontSize) || 8}pt;font-weight:${el.bold ? 700 : 400};line-height:1.15;${borderCss}${fontFam}"><span style="width:100%;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;">${escapeHtml(textVal)}</span></div>`
      })
      .join('')

    return `<div class="custom-label-box" style="position:relative;width:${w}mm;height:${h}mm;background:#fff;color:#000;box-sizing:border-box;overflow:hidden;">${elsHtml}</div>`
  }

  const tc = tpl.tableConfig
  const bw = Number(tc.borderWidth) || 1.5
  const bs = safeBorderStyle(tc.borderStyle)
  const bc = safeColor(tc.borderColor)
  const outerBorder = tc.showOuterBorder ? `border:${bw}pt ${bs} ${bc};` : 'border:none;'

  let headerHtml = ''
  if (tc.showHeader) {
    const isInv = tc.headerStyle === 'inverted'
    const hdrBg = isInv ? 'background:#000000;color:#ffffff;' : 'background:#f8fafc;color:#0f172a;'
    const hdrBorder =
      tc.showHeaderBorder !== false
        ? tc.headerStyle === 'double_line'
          ? `border-bottom:${(bw * 2.2).toFixed(2)}pt double ${bc};`
          : `border-bottom:${bw}pt ${bs} ${bc};`
        : ''
    const hdrFs = Number(tc.headerFontSize) || 7.5
    const subHtml = tc.headerSubText
      ? `<span style="font-weight:700;font-size:${Math.max(4.5, hdrFs - 0.5)}pt;opacity:0.92;line-height:1.1;">${escapeHtml(tc.headerSubText)}</span>`
      : ''
    const logoInHdr = isLogoInHeader(tc)
    const logoUrl = tc.logoUrl ? resolveAssetUrl(tc.logoUrl, origin) : ''
    const hdrLogoH = Number(
      Math.max(6.2, Math.min(8.5, hdrFs * 0.3528 * (tc.headerSubText ? 2.65 : 1.75))).toFixed(2)
    )
    const hdrLogoHtml = logoInHdr
      ? `<div style="width:${hdrLogoH}mm;height:${hdrLogoH}mm;flex:0 0 auto;display:flex;align-items:center;justify-content:center;overflow:hidden;">
          ${
            logoUrl
              ? `<img src="${escapeHtml(logoUrl)}" style="width:100%;height:100%;object-fit:contain;display:block;" alt="Logo" />`
              : DEFAULT_CREST_SVG
          }
        </div>`
      : ''

    headerHtml = `<div style="width:100%;flex:0 0 auto;padding:0.55mm 1.5mm;text-align:center;font-weight:800;text-transform:uppercase;letter-spacing:-0.01em;display:flex;flex-direction:row;align-items:center;justify-content:center;gap:1.4mm;box-sizing:border-box;-webkit-print-color-adjust:exact;print-color-adjust:exact;${hdrBg}${hdrBorder}">
      ${hdrLogoHtml}
      <div style="display:flex;flex-direction:column;align-items:center;justify-content:center;min-width:0;text-align:center;">
        <span style="font-size:${hdrFs}pt;line-height:1.12;">${escapeHtml(tc.headerText || '')}</span>
        ${subHtml}
      </div>
    </div>`
  }

  const leftQrHtml = tc.qrPosition === 'left' ? renderQrColumnHtml(tc, item, origin, 'left') : ''
  const rightQrHtml = tc.qrPosition === 'right' ? renderQrColumnHtml(tc, item, origin, 'right') : ''

  const rows = tc.rows || []
  const rowsHtml = rows
    .map((row, rIdx) => {
      const rowBg = bgStyleFor(row.bgColor)
      const hasBottom = (tc.showGridLines && rIdx < rows.length - 1) || row.hasBottomBorder
      const rowBorder = hasBottom ? `border-bottom:${bw}pt ${bs} ${bc};` : ''

      if (row.type === 'barcode_strip') {
        const bcUrl = resolveAssetUrl(`/api/barcode/${encodeURIComponent(item.id)}.png?fmt=code128`, origin)
        return `<div style="flex:1;min-height:0;display:flex;flex-direction:column;align-items:center;justify-content:center;padding:0.4mm 1mm;box-sizing:border-box;overflow:hidden;background:${rowBg.bg};color:${rowBg.color};-webkit-print-color-adjust:exact;print-color-adjust:exact;${rowBorder}">
          <img src="${escapeHtml(bcUrl)}" style="width:92%;height:4.5mm;object-fit:contain;display:block;" alt="BC" />
          <span style="font-family:monospace;font-weight:700;font-size:5.5pt;line-height:1;">${escapeHtml(item.sku || '')}</span>
        </div>`
      }

      const cells = getRowCells(row)
      const cellsHtml = cells
        .map((cell, cIdx) => {
          const cellBg = cell.bgColor && cell.bgColor !== 'white' ? bgStyleFor(cell.bgColor) : rowBg
          const cellBorder =
            tc.showLabelColBorder && cIdx < cells.length - 1 ? `border-right:${bw}pt ${bs} ${bc};` : ''
          const fs = Number(cell.fontSize || row.fontSize || tc.defaultFontSize) || 7
          const padV = Number(tc.cellPaddingMm ?? 0.8)
          const labelMaxW = cells.length === 1 ? `${Number(tc.labelColWidthPercent) || 28}%` : '45%'
          const align = cell.align || 'left'
          const valText = cell.customText || resolveFieldValue(item, cell.field)
          const labelHtml = cell.label
            ? `<span style="font-weight:700;flex-shrink:0;max-width:${labelMaxW};overflow:hidden;text-overflow:ellipsis;white-space:nowrap;padding-right:1mm;">${escapeHtml(cell.label)}:</span>`
            : ''
          return `<div style="flex:1;min-width:0;display:flex;align-items:center;overflow:hidden;box-sizing:border-box;padding:${padV}mm 1mm;font-size:${fs}pt;line-height:1.15;background:${cellBg.bg};color:${cellBg.color};-webkit-print-color-adjust:exact;print-color-adjust:exact;${cellBorder}">
            ${labelHtml}
            <span style="flex:1;min-width:0;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;text-align:${align};font-weight:${cell.bold ? 700 : 400};font-family:${cell.field === 'sku' ? 'monospace' : 'inherit'};">${escapeHtml(valText)}</span>
          </div>`
        })
        .join('')

      return `<div style="flex:1;min-height:0;display:flex;align-items:stretch;overflow:hidden;box-sizing:border-box;background:${rowBg.bg};color:${rowBg.color};-webkit-print-color-adjust:exact;print-color-adjust:exact;${rowBorder}">${cellsHtml}</div>`
    })
    .join('')

  let footerHtml = ''
  if (tc.showFooter) {
    const ftrBorder = tc.showFooterBorder ? `border-top:${bw}pt ${bs} ${bc};` : ''
    footerHtml = `<div style="width:100%;flex:0 0 auto;padding:0.4mm 1.5mm;text-align:center;color:#334155;font-weight:500;font-size:${Number(tc.footerFontSize) || 5.5}pt;line-height:1.15;display:flex;align-items:center;justify-content:center;box-sizing:border-box;${ftrBorder}">${escapeHtml(tc.footerText || '')}</div>`
  }

  return `<div class="custom-label-box" style="width:${w}mm;height:${h}mm;display:flex;flex-direction:column;background:#ffffff;color:#000000;box-sizing:border-box;overflow:hidden;${outerBorder}">
    ${headerHtml}
    <div style="flex:1;min-height:0;display:flex;overflow:hidden;">
      ${leftQrHtml}
      <div style="flex:1;min-width:0;display:flex;flex-direction:column;justify-content:space-between;overflow:hidden;">
        ${rowsHtml}
      </div>
      ${rightQrHtml}
    </div>
    ${footerHtml}
  </div>`
}

export interface CustomPrintOptions {
  copies?: number
  offset?: number
  sheetFormat?: SheetFormatId
  showCutBorder?: boolean
  showToolbar?: boolean
  autoPrint?: boolean
}

export function expandItemsForPrint(items: any[], copies = 1, offset = 0, isSheet = false): (any | null)[] {
  const safeCopies = Math.max(1, Math.min(50, Number(copies) || 1))
  const safeOffset = isSheet ? Math.max(0, Math.min(100, Number(offset) || 0)) : 0
  const expanded: (any | null)[] = []
  for (let o = 0; o < safeOffset; o++) {
    expanded.push(null)
  }
  for (const it of items) {
    for (let c = 0; c < safeCopies; c++) {
      expanded.push(it)
    }
  }
  return expanded
}

export function fitTemplateToSheetSpec(rawTpl: LabelTemplate, format?: SheetFormatId): LabelTemplate {
  const tpl: LabelTemplate = JSON.parse(JSON.stringify(rawTpl))
  ensureTableConfig(tpl)
  const fmt = format || inferSheetFormat(tpl)
  const spec = getSheetSpec(fmt, tpl.widthMm, tpl.heightMm)
  if (!spec.isSheet || fmt === 'auto_a4') return tpl

  const origW = Number(tpl.widthMm) || spec.labelW
  const origH = Number(tpl.heightMm) || spec.labelH
  const sx = spec.labelW / origW
  const sy = spec.labelH / origH
  const sMin = Math.min(sx, sy)

  if (Math.abs(sx - 1) > 0.01 || Math.abs(sy - 1) > 0.01) {
    const scaleFont = (pt?: number, fallback = 7) =>
      Math.max(4.5, Number(((Number(pt) || fallback) * sMin).toFixed(1)))

    if (Array.isArray(tpl.elements)) {
      tpl.elements = tpl.elements.map((el) => ({
        ...el,
        x: Number(((Number(el.x) || 0) * sx).toFixed(2)),
        y: Number(((Number(el.y) || 0) * sy).toFixed(2)),
        w: Number(((Number(el.w) || 10) * sx).toFixed(2)),
        h: Number(((Number(el.h) || 5) * sy).toFixed(2)),
        ...(el.fontSize ? { fontSize: scaleFont(el.fontSize, 7) } : {}),
      }))
    }

    if (tpl.tableConfig) {
      const tc = tpl.tableConfig
      tc.qrWidthMm = Math.max(8, Number(((Number(tc.qrWidthMm) || 18) * sx).toFixed(1)))
      tc.logoWidthMm = Math.max(6, Number(((Number(tc.logoWidthMm) || 10) * sMin).toFixed(1)))
      tc.cellPaddingMm = Math.max(0.3, Number(((Number(tc.cellPaddingMm) || 0.8) * sMin).toFixed(2)))
      tc.headerFontSize = scaleFont(tc.headerFontSize, 7.5)
      tc.footerFontSize = scaleFont(tc.footerFontSize, 6)
      tc.defaultFontSize = scaleFont(tc.defaultFontSize, 6.5)
      tc.rows = (tc.rows || []).map((r) => ({
        ...r,
        ...(r.fontSize ? { fontSize: scaleFont(r.fontSize, tc.defaultFontSize) } : {}),
        ...(r.cells
          ? {
              cells: r.cells.map((c) => ({
                ...c,
                ...(c.fontSize ? { fontSize: scaleFont(c.fontSize, tc.defaultFontSize) } : {}),
              })),
            }
          : {}),
      }))
    }
  }

  tpl.widthMm = spec.labelW
  tpl.heightMm = spec.labelH
  return tpl
}

export function buildCustomPrintDocument(rawTpl: LabelTemplate, items: any[], opts: CustomPrintOptions = {}): string {
  const format = opts.sheetFormat || inferSheetFormat(rawTpl)
  const tpl = fitTemplateToSheetSpec(rawTpl, format)
  const spec = getSheetSpec(format, tpl.widthMm, tpl.heightMm)
  const slotW = spec.labelW
  const slotH = spec.labelH
  const showCutBorder = typeof opts.showCutBorder === 'boolean' ? opts.showCutBorder : getShowCutBorder(tpl)
  const origin = typeof window !== 'undefined' ? window.location.origin : ''
  const slots = expandItemsForPrint(items, opts.copies ?? 1, opts.offset ?? 0, spec.isSheet)
  const totalLabels = items.length * Math.max(1, Math.min(50, Number(opts.copies) || 1))

  let bodyHtml = ''
  let pageCss = ''

  if (spec.isSheet) {
    const perSheet = Math.max(1, spec.perSheet)
    const pages: (any | null)[][] = []
    for (let i = 0; i < slots.length; i += perSheet) {
      const slice = slots.slice(i, i + perSheet)
      // Pad last sheet to full grid like built-in /api/barcode/sheet so backing sheet keeps exact proportions
      while (slice.length < perSheet) slice.push(null)
      pages.push(slice)
    }
    if (pages.length === 0) pages.push(Array(perSheet).fill(null))

    const sheetBg = spec.isYellowBacking ? '#fef08a' : '#ffffff'
    const sheetBorder = spec.isYellowBacking ? '#facc15' : '#e2e8f0'
    const slotBorderCss = showCutBorder ? 'border: 1px dashed #cbd5e1; border-radius: 4px;' : ''
    pageCss = `
      @page { size: 210mm 297mm; margin: 0; }
      .sheet-page {
        width: 210mm;
        box-sizing: border-box;
        margin: 20px auto;
        padding: ${spec.pageMargin};
        background: ${sheetBg};
        border: 1px solid ${sheetBorder};
        border-radius: 6px;
        box-shadow: 0 4px 14px rgba(0,0,0,0.08);
        display: grid;
        grid-template-columns: repeat(${spec.cols}, ${slotW}mm);
        gap: ${spec.rowGapMm}mm ${spec.colGapMm}mm;
        justify-content: center;
        align-content: start;
      }
      .label-slot {
        width: ${slotW}mm;
        height: ${slotH}mm;
        box-sizing: border-box;
        overflow: hidden;
        background: #fff;
        box-shadow: 0 1px 2px rgba(0,0,0,0.08);
        ${slotBorderCss}
      }
      .label-slot.is-blank {
        visibility: hidden;
        background: transparent;
        border: none;
        box-shadow: none;
      }
      @media print {
        body {
          -webkit-print-color-adjust: exact;
          print-color-adjust: exact;
        }
        .sheet-page {
          background: transparent !important;
          border: none !important;
          box-shadow: none !important;
          margin: 0 !important;
        }
        .label-slot {
          box-shadow: none !important;
        }
        .sheet-page:not(:last-child) {
          page-break-after: always;
          break-after: page;
        }
      }
    `

    bodyHtml = pages
      .map((pageSlots) => {
        const cells = pageSlots
          .map((it) =>
            it
              ? `<div class="label-slot">${renderCustomLabelHtml(tpl, it, origin)}</div>`
              : `<div class="label-slot is-blank"></div>`
          )
          .join('')
        return `<div class="sheet-page">${cells}</div>`
      })
      .join('')
  } else {
    pageCss = `
      @page { size: ${tpl.widthMm}mm ${tpl.heightMm}mm; margin: 0; }
      .single-label-page {
        width: ${tpl.widthMm}mm;
        height: ${tpl.heightMm}mm;
        margin: 14px auto;
        box-shadow: 0 2px 10px rgba(0,0,0,0.12);
        background: #fff;
        overflow: hidden;
      }
      @media print {
        .single-label-page {
          margin: 0 !important;
          box-shadow: none !important;
        }
        .single-label-page:not(:last-child) {
          page-break-after: always;
          break-after: page;
        }
      }
    `
    bodyHtml = slots
      .filter(Boolean)
      .map((it) => `<div class="single-label-page">${renderCustomLabelHtml(tpl, it, origin)}</div>`)
      .join('')
  }

  const cleanTitle = escapeHtml((tpl.name || 'Label Kustom').replace(/^[🌟🏷️🖨️📄⭐\s]+/, ''))
  const toolbarHtml = opts.showToolbar
    ? `<div class="screen-toolbar">
        <div style="display:flex;align-items:center;gap:6px;min-width:0;overflow:hidden;">
          <strong style="white-space:nowrap;">Lembar Cetak Label Stiker</strong>
          <span style="opacity:0.78;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;">· ${cleanTitle} (${escapeHtml(spec.shortLabel)} · ${slotW}×${slotH} mm) · ${totalLabels} Stiker Dicetak</span>
        </div>
        <button id="btn-print-sheet" type="button">Cetak Sekarang</button>
      </div>`
    : ''

  const toolbarCss = opts.showToolbar
    ? `
      body { background: #f1f5f9; padding-top: 64px; padding-bottom: 32px; }
      .screen-toolbar {
        position: fixed; top: 0; left: 0; right: 0; height: 52px;
        background: #0f172a; color: #fff; display: flex; align-items: center;
        justify-content: space-between; padding: 0 16px; z-index: 9999;
        box-shadow: 0 2px 8px rgba(0,0,0,0.25); font-size: 13px;
      }
      .screen-toolbar button {
        background: #6366f1; color: #fff; border: none; border-radius: 6px;
        padding: 8px 16px; font-weight: 600; font-size: 13px; cursor: pointer;
      }
      .screen-toolbar button:hover { background: #4f46e5; }
      @media print {
        .screen-toolbar { display: none !important; }
        body { background: #fff !important; padding: 0 !important; }
      }
    `
    : `
      body { background: #fff; margin: 0; padding: 0; }
    `

  return `<!doctype html>
<html lang="id">
<head>
<meta charset="utf-8" />
<title>${cleanTitle}</title>
<style>
  * { box-sizing: border-box; }
  html, body {
    margin: 0;
    padding: 0;
    font-family: 'Inter', 'Segoe UI', Arial, sans-serif;
    color: #000;
    -webkit-print-color-adjust: exact;
    print-color-adjust: exact;
  }
  ${toolbarCss}
  ${pageCss}
</style>
</head>
<body>
  ${toolbarHtml}
  ${bodyHtml}
</body>
</html>`
}

async function waitForDocumentImages(doc: Document, timeoutMs = 10000): Promise<void> {
  const imgs = Array.from(doc.images || [])
  if (imgs.length === 0) return
  await Promise.race([
    Promise.all(
      imgs.map(
        (img) =>
          new Promise<void>((resolve) => {
            const finish = () => {
              if (typeof img.decode === 'function') {
                img.decode().catch(() => {}).finally(resolve)
              } else {
                resolve()
              }
            }
            if (img.complete) {
              finish()
              return
            }
            img.addEventListener('load', finish, { once: true })
            img.addEventListener('error', () => resolve(), { once: true })
          })
      )
    ),
    new Promise<void>((resolve) => setTimeout(resolve, timeoutMs)),
  ])
}

export async function printCustomLabels(
  tpl: LabelTemplate,
  items: any[],
  opts: CustomPrintOptions = {}
): Promise<void> {
  const html = buildCustomPrintDocument(tpl, items, {
    ...opts,
    showToolbar: false,
    autoPrint: false,
  })

  const iframe = document.createElement('iframe')
  iframe.setAttribute('aria-hidden', 'true')
  iframe.style.position = 'fixed'
  iframe.style.left = '-9999px'
  iframe.style.top = '0'
  iframe.style.width = '210mm'
  iframe.style.height = '297mm'
  iframe.style.border = '0'
  iframe.style.opacity = '0'
  iframe.style.pointerEvents = 'none'
  document.body.appendChild(iframe)

  const cleanup = () => {
    if (iframe.parentNode) {
      iframe.parentNode.removeChild(iframe)
    }
  }

  try {
    const doc = iframe.contentDocument || iframe.contentWindow?.document
    if (!doc || !iframe.contentWindow) {
      cleanup()
      throw new Error('Gagal menyiapkan dokumen cetak')
    }

    await new Promise<void>((resolve) => {
      let settled = false
      const done = () => {
        if (!settled) {
          settled = true
          resolve()
        }
      }
      iframe.addEventListener('load', done, { once: true })
      doc.open()
      doc.write(html)
      doc.close()
      if (doc.readyState === 'complete') {
        setTimeout(done, 30)
      }
      setTimeout(done, 1500)
    })

    await waitForDocumentImages(doc)
    await new Promise<void>((resolve) =>
      requestAnimationFrame(() => requestAnimationFrame(() => resolve()))
    )

    const win = iframe.contentWindow
    const fallbackTimer = window.setTimeout(cleanup, 60000)
    win.addEventListener(
      'afterprint',
      () => {
        window.clearTimeout(fallbackTimer)
        setTimeout(cleanup, 200)
      },
      { once: true }
    )
    win.focus()
    win.print()
  } catch (err) {
    cleanup()
    throw err
  }
}

export function openCustomPrintTab(
  win: Window | null,
  tpl: LabelTemplate,
  items: any[],
  opts: CustomPrintOptions = {}
): boolean {
  if (!win) return false
  const html = buildCustomPrintDocument(tpl, items, {
    ...opts,
    showToolbar: true,
  })
  win.document.open()
  win.document.write(html)
  win.document.close()

  // Bind toolbar print handler from parent module so inherited strict CSP (script-src 'self') never blocks it
  const btn = win.document.getElementById('btn-print-sheet') as HTMLButtonElement | null
  const triggerPrint = async () => {
    if (btn) {
      btn.disabled = true
      btn.textContent = 'Menyiapkan gambar...'
    }
    try {
      await waitForDocumentImages(win.document)
      await new Promise<void>((resolve) =>
        win.requestAnimationFrame(() => win.requestAnimationFrame(() => resolve()))
      )
    } finally {
      if (btn) {
        btn.disabled = false
        btn.textContent = 'Cetak Sekarang'
      }
    }
    win.focus()
    win.print()
  }
  if (btn) {
    btn.addEventListener('click', () => {
      void triggerPrint()
    })
  }
  if (opts.autoPrint) {
    void triggerPrint()
  }
  return true
}

export function printUrlInHiddenFrame(url: string): void {
  const iframe = document.createElement('iframe')
  iframe.setAttribute('aria-hidden', 'true')
  iframe.style.cssText = 'position:fixed;right:0;bottom:0;width:0;height:0;border:0;visibility:hidden;'
  let cleaned = false
  const cleanup = () => {
    if (cleaned) return
    cleaned = true
    try {
      iframe.remove()
    } catch {}
  }
  iframe.addEventListener(
    'load',
    () => {
      const win = iframe.contentWindow
      if (!win) {
        cleanup()
        return
      }
      const imgs = Array.from(win.document.images || [])
      const runPrint = () => {
        const fallback = window.setTimeout(cleanup, 60000)
        win.addEventListener(
          'afterprint',
          () => {
            window.clearTimeout(fallback)
            setTimeout(cleanup, 200)
          },
          { once: true }
        )
        win.focus()
        win.print()
      }
      if (!imgs.length) {
        setTimeout(runPrint, 100)
        return
      }
      Promise.race([
        Promise.all(
          imgs.map(
            (img) =>
              new Promise<void>((resolve) => {
                if (img.complete) {
                  resolve()
                  return
                }
                img.addEventListener('load', () => resolve(), { once: true })
                img.addEventListener('error', () => resolve(), { once: true })
              })
          )
        ),
        new Promise<void>((resolve) => setTimeout(resolve, 6000)),
      ]).then(() => setTimeout(runPrint, 100))
    },
    { once: true }
  )
  iframe.src = url
  document.body.appendChild(iframe)
}
