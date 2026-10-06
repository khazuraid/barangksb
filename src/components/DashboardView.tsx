import React, { useState, useMemo } from 'react';
import { 
  Package, 
  Layers, 
  ArrowDownLeft, 
  AlertTriangle, 
  ScanLine, 
  PlusCircle, 
  Printer, 
  TrendingUp, 
  Building2, 
  Calendar, 
  Clock, 
  ExternalLink, 
  FileText, 
  DollarSign, 
  Image as ImageIcon,
  CheckCircle2,
  MapPin,
  Barcode,
  Sparkles,
  ShieldCheck,
  ChevronRight,
  HardDrive,
  PieChart as PieChartIcon,
  BarChart3,
  SlidersHorizontal,
  Info
} from 'lucide-react';
import { InventoryItem, StockTransaction } from '../types/inventory';

interface DashboardViewProps {
  items: InventoryItem[];
  transactions: StockTransaction[];
  categories: string[];
  onOpenScanner: () => void;
  onOpenNewItem: () => void;
  onOpenMovementIn: (item?: InventoryItem) => void;
  onNavigateTab: (tab: 'items' | 'movement-in' | 'history' | 'barcode' | 'sync') => void;
  onOpenPDFReport: () => void;
}

const CATEGORY_COLORS = [
  '#10b981', // emerald-500
  '#6366f1', // indigo-500
  '#8b5cf6', // violet-500
  '#f59e0b', // amber-500
  '#ec4899', // pink-500
  '#06b6d4', // cyan-500
  '#3b82f6', // blue-500
  '#14b8a6', // teal-500
  '#f97316', // orange-500
  '#64748b', // slate-500
];

export const DashboardView: React.FC<DashboardViewProps> = ({
  items,
  transactions,
  categories,
  onOpenScanner,
  onOpenNewItem,
  onOpenMovementIn,
  onNavigateTab,
  onOpenPDFReport,
}) => {
  // Chart visual configuration state
  const [chartType, setChartType] = useState<'donut' | 'bar'>('donut');
  const [chartMetric, setChartMetric] = useState<'units' | 'value' | 'count'>('units');
  const [hoveredCategory, setHoveredCategory] = useState<string | null>(null);

  // Compute Key Analytics
  const totalItemCount = items.length;
  const totalUnits = items.reduce((acc, curr) => acc + (Number(curr?.currentStock) || 0), 0);
  const totalAssetValue = items.reduce((acc, curr) => acc + ((Number(curr?.pricePerUnit) || 0) * (Number(curr?.currentStock) || 0)), 0);

  // Today's date filter
  const todayStr = new Date().toISOString().slice(0, 10);
  const todayTransactions = transactions.filter(t => t?.timestamp && t.timestamp.slice(0, 10) === todayStr);
  const todayInQty = todayTransactions
    .filter(t => t.type === 'IN')
    .reduce((acc, curr) => acc + (Number(curr?.quantity) || 0), 0);

  // Status conditions
  const availableItems = items.filter(i => i?.isAvailable !== false);
  const functionalItems = items.filter(i => i?.conditionStatus === 'Berfungsi' || !i?.conditionStatus);
  const damagedItems = items.filter(i => i?.conditionStatus && i?.conditionStatus !== 'Berfungsi');
  const lowStockItems = items.filter(i => (Number(i?.currentStock) || 0) <= (Number(i?.minStock) || 0));

  // Category Detailed Analytics for Charts
  const categoryStats = useMemo(() => {
    const map = new Map<string, { itemCount: number; totalUnits: number; totalValue: number }>();
    
    // Initialize with categories
    categories.forEach(cat => {
      if (cat) map.set(cat, { itemCount: 0, totalUnits: 0, totalValue: 0 });
    });

    items.forEach(item => {
      const cat = item?.category || 'Lainnya';
      const existing = map.get(cat) || { itemCount: 0, totalUnits: 0, totalValue: 0 };
      existing.itemCount += 1;
      existing.totalUnits += Number(item?.currentStock) || 0;
      existing.totalValue += (Number(item?.pricePerUnit) || 0) * (Number(item?.currentStock) || 0);
      map.set(cat, existing);
    });

    const list = Array.from(map.entries())
      .filter(([, stat]) => stat.itemCount > 0 || stat.totalUnits > 0)
      .map(([cat, stat], idx) => ({
        category: cat,
        itemCount: stat.itemCount,
        totalUnits: stat.totalUnits,
        totalValue: stat.totalValue,
        color: CATEGORY_COLORS[idx % CATEGORY_COLORS.length],
      }))
      .sort((a, b) => {
        if (chartMetric === 'units') return b.totalUnits - a.totalUnits;
        if (chartMetric === 'value') return b.totalValue - a.totalValue;
        return b.itemCount - a.itemCount;
      });

    return list;
  }, [items, categories, chartMetric]);

  // Total based on active chart metric
  const totalMetricValue = useMemo(() => {
    if (chartMetric === 'units') return totalUnits;
    if (chartMetric === 'value') return totalAssetValue;
    return totalItemCount;
  }, [chartMetric, totalUnits, totalAssetValue, totalItemCount]);

  // Max value for bar chart proportional scaling
  const maxMetricValue = useMemo(() => {
    if (categoryStats.length === 0) return 1;
    return Math.max(
      ...categoryStats.map(s => {
        if (chartMetric === 'units') return s.totalUnits;
        if (chartMetric === 'value') return s.totalValue;
        return s.itemCount;
      })
    );
  }, [categoryStats, chartMetric]);

  // Recent 5 transactions
  const recentTransactions = [...transactions]
    .sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
    .slice(0, 5);

  const formatRupiah = (val?: number | null) => {
    const num = Number(val) || 0;
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      maximumFractionDigits: 0,
    }).format(num);
  };

  const formatMetricDisplay = (val?: number | null) => {
    const num = Number(val) || 0;
    if (chartMetric === 'value') return formatRupiah(num);
    if (chartMetric === 'units') return `${num.toLocaleString('id-ID')} Unit`;
    return `${num} Jenis (SKU)`;
  };

  // Active hovered stat for Donut center display
  const activeHoveredStat = useMemo(() => {
    if (!hoveredCategory) return null;
    return categoryStats.find(s => s.category === hoveredCategory) || null;
  }, [hoveredCategory, categoryStats]);

  // Donut SVG circumference and segment computations
  const radius = 68;
  const circumference = 2 * Math.PI * radius; // ≈ 427.2566

  return (
    <div className="space-y-6">
      {/* Executive Hero Banner */}
      <div className="bg-gradient-to-br from-slate-950 via-slate-900 to-indigo-950 rounded-3xl p-6 sm:p-8 text-white shadow-xl border border-slate-800/80 relative overflow-hidden">
        <div className="absolute right-0 top-0 translate-x-12 -translate-y-12 w-80 h-80 bg-emerald-500/10 rounded-full blur-3xl pointer-events-none" />
        <div className="absolute left-1/3 bottom-0 w-60 h-60 bg-indigo-500/10 rounded-full blur-3xl pointer-events-none" />
        
        <div className="relative z-10 flex flex-col lg:flex-row lg:items-center justify-between gap-6">
          <div className="max-w-2xl">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-emerald-950/80 border border-emerald-700/60 text-emerald-300 text-xs font-bold mb-3">
              <span className="w-2 h-2 rounded-full bg-emerald-400 animate-ping" />
              <span>Sistem Manajemen Aset & Log Barang Masuk Real-Time</span>
            </div>
            <h1 className="text-2xl sm:text-3xl font-black tracking-tight text-white">
              Pusat Kendali Inventaris & Pengadaan Kantor
            </h1>
            <p className="mt-2 text-slate-300 text-xs sm:text-sm leading-relaxed">
              Pencatatan resmi barang masuk dengan bukti foto fisik & geotagging GPS otomatis, penomoran seri, cetak barcode sticker, serta sinkronisasi database cloud Firestore dan Google Sheets.
            </p>
          </div>

          <div className="flex flex-wrap items-center gap-2.5 shrink-0">
            <button
              onClick={onOpenNewItem}
              className="inline-flex items-center gap-2 px-4 py-2.5 bg-emerald-500 hover:bg-emerald-400 text-slate-950 font-black text-xs sm:text-sm rounded-xl shadow-lg shadow-emerald-500/25 active:scale-95 transition-all cursor-pointer ring-1 ring-emerald-300"
            >
              <PlusCircle className="w-4 h-4 text-slate-950" />
              <span>+ Input Barang Baru Masuk</span>
            </button>
            <button
              onClick={onOpenScanner}
              className="inline-flex items-center gap-2 px-4 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white font-bold text-xs sm:text-sm rounded-xl shadow-lg shadow-indigo-600/30 active:scale-95 transition-all cursor-pointer ring-1 ring-indigo-400/30"
            >
              <ScanLine className="w-4 h-4 text-indigo-200" />
              <span>Scan Barcode</span>
            </button>
            <button
              onClick={onOpenPDFReport}
              className="inline-flex items-center gap-2 px-3.5 py-2.5 bg-slate-900 hover:bg-slate-800 text-slate-200 font-bold text-xs sm:text-sm rounded-xl border border-slate-700 active:scale-95 transition-all cursor-pointer"
            >
              <FileText className="w-4 h-4 text-rose-400" />
              <span>Unduh Rekap PDF</span>
            </button>
          </div>
        </div>
      </div>

      {/* 4 Primary Metric Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* Card 1: Total SKU Master */}
        <div className="bg-white rounded-3xl p-5 border border-slate-200/90 shadow-xs hover:border-indigo-300 transition-all group">
          <div className="flex items-center justify-between">
            <span className="text-[11px] font-bold uppercase tracking-wider text-slate-400">Total Jenis Aset</span>
            <div className="w-10 h-10 rounded-2xl bg-indigo-50 text-indigo-600 flex items-center justify-center group-hover:scale-105 transition-transform">
              <Package className="w-5 h-5" />
            </div>
          </div>
          <div className="mt-3 flex items-baseline gap-2">
            <span className="text-3xl font-black text-slate-900 tracking-tight">{totalItemCount}</span>
            <span className="text-xs font-bold text-slate-500">Item Master</span>
          </div>
          <div className="mt-3 pt-3 border-t border-slate-100 flex items-center justify-between text-xs">
            <span className="text-slate-500">{categories.length} Kategori</span>
            <button 
              onClick={() => onNavigateTab('items')} 
              className="text-indigo-600 hover:text-indigo-700 font-bold flex items-center gap-0.5 cursor-pointer"
            >
              <span>Katalog</span>
              <ChevronRight className="w-3.5 h-3.5" />
            </button>
          </div>
        </div>

        {/* Card 2: Total Fisik Barang */}
        <div className="bg-white rounded-3xl p-5 border border-slate-200/90 shadow-xs hover:border-emerald-300 transition-all group">
          <div className="flex items-center justify-between">
            <span className="text-[11px] font-bold uppercase tracking-wider text-slate-400">Total Fisik Barang</span>
            <div className="w-10 h-10 rounded-2xl bg-emerald-50 text-emerald-600 flex items-center justify-center group-hover:scale-105 transition-transform">
              <Layers className="w-5 h-5" />
            </div>
          </div>
          <div className="mt-3 flex items-baseline gap-2">
            <span className="text-3xl font-black text-slate-900 tracking-tight">{totalUnits.toLocaleString('id-ID')}</span>
            <span className="text-xs font-bold text-slate-500">Unit / Pcs</span>
          </div>
          <div className="mt-3 pt-3 border-t border-slate-100 flex items-center justify-between text-xs">
            <span className="text-emerald-700 font-semibold">{availableItems.length} Ada Fisik</span>
            <span className="text-slate-400">{lowStockItems.length} Min. Stok</span>
          </div>
        </div>

        {/* Card 3: Total Nilai Aset Investasi */}
        <div className="bg-white rounded-3xl p-5 border border-slate-200/90 shadow-xs hover:border-violet-300 transition-all group">
          <div className="flex items-center justify-between">
            <span className="text-[11px] font-bold uppercase tracking-wider text-slate-400">Estimasi Nilai Aset</span>
            <div className="w-10 h-10 rounded-2xl bg-violet-50 text-violet-600 flex items-center justify-center group-hover:scale-105 transition-transform">
              <DollarSign className="w-5 h-5" />
            </div>
          </div>
          <div className="mt-3">
            <span className="text-xl sm:text-2xl font-black text-slate-900 tracking-tight block truncate">
              {formatRupiah(totalAssetValue)}
            </span>
            <span className="text-[11px] text-slate-400 font-medium">Berdasarkan Harga Perolehan</span>
          </div>
          <div className="mt-3 pt-3 border-t border-slate-100 flex items-center justify-between text-xs">
            <span className="text-slate-500">Standar Pengadaan</span>
            <span className="text-violet-600 font-bold">100% Tercatat</span>
          </div>
        </div>

        {/* Card 4: Penerimaan Hari Ini */}
        <div className="bg-white rounded-3xl p-5 border border-emerald-200 shadow-xs hover:border-emerald-400 transition-all group bg-gradient-to-b from-white to-emerald-50/30">
          <div className="flex items-center justify-between">
            <span className="text-[11px] font-bold uppercase tracking-wider text-emerald-800">Barang Masuk Hari Ini</span>
            <div className="w-10 h-10 rounded-2xl bg-emerald-500 text-white flex items-center justify-center group-hover:scale-105 transition-transform shadow-sm">
              <ArrowDownLeft className="w-5 h-5" />
            </div>
          </div>
          <div className="mt-3 flex items-baseline gap-2">
            <span className="text-3xl font-black text-emerald-700 tracking-tight">+{todayInQty}</span>
            <span className="text-xs font-bold text-emerald-600">Unit Masuk</span>
          </div>
          <div className="mt-3 pt-3 border-t border-emerald-100 flex items-center justify-between text-xs">
            <span className="text-emerald-700 font-semibold">{todayTransactions.length} Dokumen PO</span>
            <button 
              onClick={() => onNavigateTab('history')} 
              className="text-emerald-700 hover:underline font-bold cursor-pointer"
            >
              Lihat Log →
            </button>
          </div>
        </div>
      </div>

      {/* FEATURED: Visual Stock Analytics by Category (Donut Chart & Bar Chart) */}
      <div className="bg-white rounded-3xl p-6 sm:p-7 border border-slate-200/90 shadow-xs space-y-6">
        {/* Header with Chart Type & Metric Controls */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b border-slate-100">
          <div>
            <div className="flex items-center gap-2">
              <span className="p-2 rounded-xl bg-indigo-50 text-indigo-600">
                <PieChartIcon className="w-5 h-5" />
              </span>
              <div>
                <h3 className="text-lg font-black text-slate-900 tracking-tight">
                  Ringkasan Statistik Stok Berdasarkan Kategori
                </h3>
                <p className="text-xs text-slate-500">
                  Visualisasi proporsi sebaran inventaris kantor dalam bagan donat & diagram batang komparatif
                </p>
              </div>
            </div>
          </div>

          <div className="flex items-center gap-2.5 flex-wrap">
            {/* Metric Switcher */}
            <div className="flex items-center p-1 bg-slate-100 rounded-xl border border-slate-200 text-xs font-bold">
              <button
                type="button"
                onClick={() => setChartMetric('units')}
                className={`px-3 py-1.5 rounded-lg transition-all cursor-pointer ${
                  chartMetric === 'units'
                    ? 'bg-white text-indigo-700 shadow-xs'
                    : 'text-slate-500 hover:text-slate-900'
                }`}
                title="Berdasarkan Total Jumlah Unit Fisik"
              >
                Fisik (Unit)
              </button>
              <button
                type="button"
                onClick={() => setChartMetric('value')}
                className={`px-3 py-1.5 rounded-lg transition-all cursor-pointer ${
                  chartMetric === 'value'
                    ? 'bg-white text-indigo-700 shadow-xs'
                    : 'text-slate-500 hover:text-slate-900'
                }`}
                title="Berdasarkan Akumulasi Nilai Aset (Rp)"
              >
                Nilai (Rp)
              </button>
              <button
                type="button"
                onClick={() => setChartMetric('count')}
                className={`px-3 py-1.5 rounded-lg transition-all cursor-pointer ${
                  chartMetric === 'count'
                    ? 'bg-white text-indigo-700 shadow-xs'
                    : 'text-slate-500 hover:text-slate-900'
                }`}
                title="Berdasarkan Jumlah Jenis Barang (SKU)"
              >
                Jenis (SKU)
              </button>
            </div>

            {/* Chart Type Toggle (Donut vs Bar) */}
            <div className="flex items-center p-1 bg-slate-100 rounded-xl border border-slate-200">
              <button
                type="button"
                onClick={() => setChartType('donut')}
                className={`p-1.5 rounded-lg transition-colors cursor-pointer flex items-center gap-1.5 text-xs font-bold ${
                  chartType === 'donut'
                    ? 'bg-white text-indigo-700 shadow-xs'
                    : 'text-slate-500 hover:text-slate-900'
                }`}
                title="Tampilkan Bagan Donut"
              >
                <PieChartIcon className="w-4 h-4" />
                <span className="hidden md:inline">Donat</span>
              </button>
              <button
                type="button"
                onClick={() => setChartType('bar')}
                className={`p-1.5 rounded-lg transition-colors cursor-pointer flex items-center gap-1.5 text-xs font-bold ${
                  chartType === 'bar'
                    ? 'bg-white text-indigo-700 shadow-xs'
                    : 'text-slate-500 hover:text-slate-900'
                }`}
                title="Tampilkan Bagan Batang"
              >
                <BarChart3 className="w-4 h-4" />
                <span className="hidden md:inline">Batang</span>
              </button>
            </div>
          </div>
        </div>

        {/* Content: Dual Chart & Breakdown Panels */}
        {categoryStats.length === 0 ? (
          <div className="py-12 text-center text-slate-400">
            <Package className="w-10 h-10 mx-auto text-slate-300 mb-2" />
            <p className="font-bold text-slate-700 text-sm">Belum ada data barang untuk ditampilkan dalam bagan</p>
            <p className="text-xs text-slate-400 mt-1">Daftarkan barang baru masuk untuk melihat ringkasan visual statistik.</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-center">
            {/* Chart Visual Section (5 cols) */}
            <div className="lg:col-span-5 flex flex-col items-center justify-center p-4 bg-slate-50/60 rounded-3xl border border-slate-100 relative min-h-[300px]">
              {chartType === 'donut' ? (
                /* DONUT CHART SVG */
                <div className="relative flex items-center justify-center">
                  <svg className="w-64 h-64 sm:w-72 sm:h-72 -rotate-90 transform" viewBox="0 0 200 200">
                    {/* Background Track Circle */}
                    <circle
                      cx="100"
                      cy="100"
                      r={radius}
                      fill="transparent"
                      stroke="#f1f5f9"
                      strokeWidth="24"
                    />

                    {/* Donut Slices */}
                    {(() => {
                      let accumulatedLength = 0;
                      return categoryStats.map((stat) => {
                        const val =
                          chartMetric === 'units'
                            ? stat.totalUnits
                            : chartMetric === 'value'
                            ? stat.totalValue
                            : stat.itemCount;
                        const pct = totalMetricValue > 0 ? val / totalMetricValue : 0;
                        const sliceLength = pct * circumference;
                        const offset = accumulatedLength;
                        accumulatedLength += sliceLength;

                        const isHovered = hoveredCategory === stat.category;

                        return (
                          <circle
                            key={stat.category}
                            cx="100"
                            cy="100"
                            r={radius}
                            fill="transparent"
                            stroke={stat.color}
                            strokeWidth={isHovered ? '28' : '22'}
                            strokeDasharray={`${sliceLength} ${circumference - sliceLength}`}
                            strokeDashoffset={-offset}
                            strokeLinecap="butt"
                            className="transition-all duration-300 cursor-pointer hover:opacity-90"
                            onMouseEnter={() => setHoveredCategory(stat.category)}
                            onMouseLeave={() => setHoveredCategory(null)}
                          />
                        );
                      });
                    })()}
                  </svg>

                  {/* Center Stat Readout */}
                  <div className="absolute inset-0 flex flex-col items-center justify-center text-center p-6 pointer-events-none select-none">
                    {activeHoveredStat ? (
                      <div className="animate-in fade-in zoom-in-95 duration-150 max-w-[170px]">
                        <span className="w-2.5 h-2.5 rounded-full inline-block mx-auto mb-1" style={{ backgroundColor: activeHoveredStat.color }} />
                        <p className="text-[11px] font-black text-slate-800 line-clamp-1">
                          {activeHoveredStat.category}
                        </p>
                        <p className="text-lg font-black text-slate-900 mt-0.5 tracking-tight">
                          {chartMetric === 'units'
                            ? `${activeHoveredStat.totalUnits.toLocaleString('id-ID')} Unit`
                            : chartMetric === 'value'
                            ? formatRupiah(activeHoveredStat.totalValue)
                            : `${activeHoveredStat.itemCount} SKU`}
                        </p>
                        <span className="px-2 py-0.5 rounded-full text-[10px] font-bold bg-slate-900 text-white mt-1 inline-block">
                          {totalMetricValue > 0
                            ? (
                                ((chartMetric === 'units'
                                  ? activeHoveredStat.totalUnits
                                  : chartMetric === 'value'
                                  ? activeHoveredStat.totalValue
                                  : activeHoveredStat.itemCount) /
                                  totalMetricValue) *
                                100
                              ).toFixed(1)
                            : 0}
                          % Porsi
                        </span>
                      </div>
                    ) : (
                      <div>
                        <span className="text-[10px] font-bold uppercase tracking-wider text-slate-400">
                          {chartMetric === 'units'
                            ? 'Total Stok Fisik'
                            : chartMetric === 'value'
                            ? 'Total Nilai Aset'
                            : 'Total Jenis Barang'}
                        </span>
                        <p className="text-xl sm:text-2xl font-black text-slate-900 tracking-tight mt-0.5">
                          {chartMetric === 'units'
                            ? `${totalUnits.toLocaleString('id-ID')} Unit`
                            : chartMetric === 'value'
                            ? formatRupiah(totalAssetValue)
                            : `${totalItemCount} SKU`}
                        </p>
                        <p className="text-[10px] text-slate-400 font-medium mt-0.5">
                          {categoryStats.length} Kategori Aktif
                        </p>
                      </div>
                    )}
                  </div>
                </div>
              ) : (
                /* BAR CHART COMPARISON VIEW */
                <div className="w-full space-y-3 p-2">
                  <div className="flex items-center justify-between text-[11px] text-slate-400 font-bold uppercase pb-1 border-b border-slate-200">
                    <span>Kategori</span>
                    <span>
                      {chartMetric === 'units' ? 'Jumlah Unit' : chartMetric === 'value' ? 'Nilai Aset' : 'SKU'}
                    </span>
                  </div>
                  {categoryStats.map(stat => {
                    const val =
                      chartMetric === 'units'
                        ? stat.totalUnits
                        : chartMetric === 'value'
                        ? stat.totalValue
                        : stat.itemCount;
                    const pct = maxMetricValue > 0 ? (val / maxMetricValue) * 100 : 0;
                    const isHovered = hoveredCategory === stat.category;

                    return (
                      <div
                        key={stat.category}
                        className={`space-y-1 p-1.5 rounded-xl transition-all cursor-pointer ${
                          isHovered ? 'bg-white shadow-xs' : ''
                        }`}
                        onMouseEnter={() => setHoveredCategory(stat.category)}
                        onMouseLeave={() => setHoveredCategory(null)}
                      >
                        <div className="flex items-center justify-between text-xs">
                          <span className="font-bold text-slate-800 truncate max-w-[200px] flex items-center gap-1.5">
                            <span className="w-2.5 h-2.5 rounded-full shrink-0" style={{ backgroundColor: stat.color }} />
                            <span>{stat.category}</span>
                          </span>
                          <span className="font-mono font-black text-slate-900 text-xs">
                            {chartMetric === 'value' ? formatRupiah(val) : `${val.toLocaleString('id-ID')}`}
                          </span>
                        </div>
                        <div className="w-full h-2.5 rounded-full bg-slate-200/80 overflow-hidden">
                          <div
                            className="h-full rounded-full transition-all duration-500"
                            style={{
                              width: `${Math.max(pct, 5)}%`,
                              backgroundColor: stat.color,
                            }}
                          />
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
            </div>

            {/* Data Breakdown Table & Interactive Legend (7 cols) */}
            <div className="lg:col-span-7 space-y-3">
              <div className="flex items-center justify-between text-xs font-bold text-slate-500 pb-2 border-b border-slate-100">
                <span>Rincian Komparasi per Kategori</span>
                <span className="text-[11px] font-normal text-slate-400">Arahkan kursor untuk menyorot</span>
              </div>

              <div className="divide-y divide-slate-100 max-h-[300px] overflow-y-auto pr-1">
                {categoryStats.map((stat) => {
                  const val =
                    chartMetric === 'units'
                      ? stat.totalUnits
                      : chartMetric === 'value'
                      ? stat.totalValue
                      : stat.itemCount;
                  const sharePct = totalMetricValue > 0 ? ((val / totalMetricValue) * 100).toFixed(1) : '0';
                  const isHovered = hoveredCategory === stat.category;

                  return (
                    <div
                      key={stat.category}
                      onMouseEnter={() => setHoveredCategory(stat.category)}
                      onMouseLeave={() => setHoveredCategory(null)}
                      className={`py-2.5 px-3 rounded-2xl transition-all cursor-pointer flex items-center justify-between gap-3 ${
                        isHovered ? 'bg-indigo-50/70 ring-1 ring-indigo-200' : 'hover:bg-slate-50/80'
                      }`}
                    >
                      <div className="flex items-center gap-2.5 min-w-0">
                        <span
                          className="w-3.5 h-3.5 rounded-lg shrink-0 shadow-2xs"
                          style={{ backgroundColor: stat.color }}
                        />
                        <div className="min-w-0">
                          <p className="text-xs font-extrabold text-slate-900 truncate">
                            {stat.category}
                          </p>
                          <div className="flex items-center gap-2 text-[10px] text-slate-500 mt-0.5">
                            <span>{stat.itemCount} SKU Master</span>
                            <span>•</span>
                            <span className="font-semibold text-slate-700">{stat.totalUnits.toLocaleString('id-ID')} Unit</span>
                            <span>•</span>
                            <span>{formatRupiah(stat.totalValue)}</span>
                          </div>
                        </div>
                      </div>

                      <div className="text-right shrink-0">
                        <span className="font-mono font-black text-slate-900 text-xs block">
                          {formatMetricDisplay(val)}
                        </span>
                        <span className="text-[10px] font-bold text-indigo-600 bg-indigo-50 px-1.5 py-0.2 rounded-md">
                          {sharePct}% Porsi
                        </span>
                      </div>
                    </div>
                  );
                })}
              </div>

              <div className="pt-3 border-t border-slate-100 flex items-center justify-between text-xs text-slate-500">
                <span className="flex items-center gap-1.5 text-slate-500">
                  <Info className="w-3.5 h-3.5 text-slate-400" />
                  <span>Kalkulasi mencakup seluruh inventaris aktif di Master Barang</span>
                </span>
                <button
                  type="button"
                  onClick={() => onNavigateTab('items')}
                  className="font-bold text-indigo-600 hover:underline cursor-pointer flex items-center gap-1"
                >
                  <span>Buka Master Barang</span>
                  <ChevronRight className="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Analytics & Quick Actions Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        {/* Left Column: Quick Actions & Status Strip (5 cols) */}
        <div className="lg:col-span-5 space-y-6">
          {/* Status Keberadaan & Kondisi Card */}
          <div className="bg-white rounded-3xl p-6 border border-slate-200/90 shadow-xs space-y-4">
            <h3 className="font-extrabold text-slate-900 text-sm tracking-tight flex items-center gap-2">
              <ShieldCheck className="w-4 h-4 text-emerald-600" />
              <span>Status Kondisi Fisik & Keberadaan Aset</span>
            </h3>

            <div className="grid grid-cols-3 gap-2.5 text-center">
              <div className="p-3 rounded-2xl bg-emerald-50 border border-emerald-100">
                <p className="text-[10px] font-bold text-emerald-700 uppercase">Berfungsi</p>
                <p className="text-lg font-black text-emerald-900 mt-0.5">{functionalItems.length}</p>
                <span className="text-[9px] text-emerald-600 font-semibold">Siap Pakai</span>
              </div>
              <div className="p-3 rounded-2xl bg-amber-50 border border-amber-100">
                <p className="text-[10px] font-bold text-amber-700 uppercase">Perbaikan</p>
                <p className="text-lg font-black text-amber-900 mt-0.5">{damagedItems.length}</p>
                <span className="text-[9px] text-amber-600 font-semibold">Perlu Servis</span>
              </div>
              <div className="p-3 rounded-2xl bg-slate-50 border border-slate-200">
                <p className="text-[10px] font-bold text-slate-500 uppercase">Tersedia</p>
                <p className="text-lg font-black text-slate-800 mt-0.5">{availableItems.length}</p>
                <span className="text-[9px] text-slate-500 font-semibold">Ada Fisik</span>
              </div>
            </div>

            <div className="p-3 rounded-2xl bg-slate-50 border border-slate-100 text-xs text-slate-600 space-y-1.5">
              <div className="flex items-center justify-between">
                <span>Stok di Bawah Batas Minimum:</span>
                <span className="font-bold text-amber-600">{lowStockItems.length} Item</span>
              </div>
              <div className="flex items-center justify-between">
                <span>Akumulasi Nilai Aset Baik:</span>
                <span className="font-bold text-slate-900 font-mono">
                  {formatRupiah(functionalItems.reduce((acc, i) => acc + ((i.pricePerUnit || 0) * i.currentStock), 0))}
                </span>
              </div>
            </div>
          </div>

          {/* Quick Actions Shortcuts */}
          <div className="bg-slate-900 text-white rounded-3xl p-6 border border-slate-800 shadow-md">
            <h3 className="font-bold text-white text-sm flex items-center gap-2 mb-3">
              <Sparkles className="w-4 h-4 text-emerald-400" />
              <span>Aksi Cepat Staf Pengadaan</span>
            </h3>
            <p className="text-xs text-slate-400 mb-4 leading-relaxed">
              Pintasan alur operasional barang masuk kantor, pembuatan label, dan sinkronisasi laporan:
            </p>

            <div className="grid grid-cols-2 gap-2.5 text-xs font-bold">
              <button
                onClick={onOpenNewItem}
                className="p-3 bg-emerald-600/20 hover:bg-emerald-600/30 border border-emerald-500/30 rounded-2xl text-emerald-300 flex flex-col items-start gap-1.5 transition-colors cursor-pointer text-left"
              >
                <PlusCircle className="w-4 h-4 text-emerald-400" />
                <span>+ Barang Baru Masuk</span>
              </button>

              <button
                onClick={() => onNavigateTab('barcode')}
                className="p-3 bg-indigo-600/20 hover:bg-indigo-600/30 border border-indigo-500/30 rounded-2xl text-indigo-300 flex flex-col items-start gap-1.5 transition-colors cursor-pointer text-left"
              >
                <Printer className="w-4 h-4 text-indigo-400" />
                <span>Cetak Barcode Label</span>
              </button>

              <button
                onClick={onOpenScanner}
                className="p-3 bg-slate-800 hover:bg-slate-700/80 border border-slate-700 rounded-2xl text-slate-300 flex flex-col items-start gap-1.5 transition-colors cursor-pointer text-left"
              >
                <ScanLine className="w-4 h-4 text-slate-400" />
                <span>Scan Barcode Kamera</span>
              </button>

              <button
                onClick={() => onNavigateTab('sync')}
                className="p-3 bg-slate-800 hover:bg-slate-700/80 border border-slate-700 rounded-2xl text-slate-300 flex flex-col items-start gap-1.5 transition-colors cursor-pointer text-left"
              >
                <HardDrive className="w-4 h-4 text-slate-400" />
                <span>Google Drive & Sheets</span>
              </button>
            </div>
          </div>
        </div>

        {/* Right Column: Recent Incoming Goods Audit Feed (7 cols) */}
        <div className="lg:col-span-7">
          <div className="bg-white rounded-3xl p-6 border border-slate-200/90 shadow-xs h-full flex flex-col justify-between">
            <div>
              <div className="flex items-center justify-between mb-4">
                <div>
                  <h3 className="font-extrabold text-slate-900 text-sm tracking-tight flex items-center gap-2">
                    <ArrowDownLeft className="w-4 h-4 text-emerald-600" />
                    <span>Log Penerimaan Barang Masuk Terbaru</span>
                  </h3>
                  <p className="text-[11px] text-slate-500 mt-0.5">
                    Transaksi mutasi masuk lengkap dengan verifikasi foto fisik & geotag GPS
                  </p>
                </div>
                <button
                  onClick={() => onNavigateTab('history')}
                  className="text-xs font-bold text-indigo-600 hover:underline cursor-pointer flex items-center gap-1"
                >
                  <span>Lihat Semua ({transactions.length})</span>
                  <ChevronRight className="w-3.5 h-3.5" />
                </button>
              </div>

              {recentTransactions.length === 0 ? (
                <div className="py-12 text-center text-slate-400 border border-dashed border-slate-200 rounded-2xl">
                  <ArrowDownLeft className="w-8 h-8 mx-auto text-slate-300 mb-2" />
                  <p className="font-bold text-slate-600 text-xs">Belum ada catatan barang masuk</p>
                  <p className="text-[11px] text-slate-400 mt-1">Gunakan tombol "Input Barang Baru Masuk" untuk mencatat.</p>
                </div>
              ) : (
                <div className="divide-y divide-slate-100">
                  {recentTransactions.map(tx => {
                    const dateObj = new Date(tx.timestamp);
                    const formattedDate = dateObj.toLocaleDateString('id-ID', { day: 'numeric', month: 'short' });
                    const formattedTime = dateObj.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' });

                    return (
                      <div key={tx.id} className="py-3.5 flex items-center justify-between gap-3 hover:bg-slate-50/60 rounded-xl px-2 transition-colors">
                        <div className="flex items-center gap-3 min-w-0">
                          {/* Thumbnail with GPS badge */}
                          {tx.photoUrl ? (
                            <div className="relative w-12 h-12 rounded-xl overflow-hidden border border-slate-200 shrink-0 bg-slate-100 shadow-2xs">
                              <img src={tx.photoUrl} alt={tx.itemName} className="w-full h-full object-cover" />
                              {tx.geoTag && (
                                <div className="absolute bottom-0 inset-x-0 bg-emerald-600/90 text-white text-[7px] font-bold text-center py-0.2">
                                  GPS
                                </div>
                              )}
                            </div>
                          ) : (
                            <div className="w-12 h-12 rounded-xl bg-slate-100 border border-slate-200 flex items-center justify-center text-slate-400 shrink-0">
                              <ImageIcon className="w-5 h-5 text-slate-300" />
                            </div>
                          )}

                          <div className="min-w-0">
                            <p className="font-extrabold text-slate-900 text-xs truncate">{tx.itemName}</p>
                            <div className="flex items-center gap-2 text-[10px] text-slate-500 font-medium mt-0.5">
                              <span className="font-mono text-indigo-600 font-bold">{tx.itemSku}</span>
                              <span>•</span>
                              <span>{tx.supplierOrSource || 'Vendor Kantor'}</span>
                              {tx.serialNumber && (
                                <>
                                  <span>•</span>
                                  <span className="font-mono">{tx.serialNumber}</span>
                                </>
                              )}
                            </div>
                            <p className="text-[10px] text-slate-400 mt-0.5">
                              Penerima: <strong className="text-slate-600">{tx.receivedBy || 'Staff GA'}</strong>
                            </p>
                          </div>
                        </div>

                        <div className="text-right shrink-0">
                          <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg bg-emerald-50 text-emerald-800 font-black text-xs border border-emerald-200/80">
                            +{tx.quantity} {tx.unit}
                          </span>
                          <p className="text-[10px] text-slate-400 mt-1 font-mono">
                            {formattedDate}, {formattedTime}
                          </p>
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
            </div>

            <div className="mt-4 pt-4 border-t border-slate-100 flex items-center justify-between text-xs text-slate-500">
              <span className="flex items-center gap-1.5 text-emerald-700 font-semibold">
                <CheckCircle2 className="w-4 h-4 text-emerald-600" />
                <span>Semua mutasi barang masuk terotentikasi & tersimpan aman</span>
              </span>
              <button
                onClick={() => onNavigateTab('history')}
                className="font-bold text-indigo-600 hover:underline cursor-pointer"
              >
                Lihat Audit Trail →
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
