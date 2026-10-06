import React, { useState, useMemo } from 'react';
import { 
  Search, 
  Plus, 
  Edit3, 
  Trash2, 
  Barcode, 
  SlidersHorizontal,
  MapPin,
  Tag,
  AlertCircle,
  FileText,
  Image as ImageIcon,
  ExternalLink,
  Navigation,
  CheckCircle2,
  HardDrive,
  LayoutGrid,
  List,
  Download,
  Building2,
  DollarSign,
  Package,
  Layers,
  Sparkles,
  X
} from 'lucide-react';
import { InventoryItem } from '../types/inventory';

interface InventoryListProps {
  items: InventoryItem[];
  categories: string[];
  onOpenNewItem: () => void;
  onEditItem: (item: InventoryItem) => void;
  onDeleteItem: (item: InventoryItem) => void;
  onOpenBarcodeModal: (item: InventoryItem) => void;
  onOpenPDFReport: () => void;
}

export const InventoryList: React.FC<InventoryListProps> = ({
  items,
  categories,
  onOpenNewItem,
  onEditItem,
  onDeleteItem,
  onOpenBarcodeModal,
  onOpenPDFReport,
}) => {
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<string>('ALL');
  const [selectedCondition, setSelectedCondition] = useState<string>('ALL');
  const [selectedAvailability, setSelectedAvailability] = useState<'ALL' | 'ADA' | 'TIDAK'>('ALL');
  const [viewMode, setViewMode] = useState<'table' | 'grid'>('table');
  const [previewPhoto, setPreviewPhoto] = useState<{ url: string; title: string; geoTag?: any; driveLink?: string } | null>(null);

  const filteredItems = useMemo(() => {
    return items.filter(item => {
      const q = searchTerm.toLowerCase();
      const matchSearch = 
        item.name.toLowerCase().includes(q) ||
        item.sku.toLowerCase().includes(q) ||
        (item.merk && item.merk.toLowerCase().includes(q)) ||
        (item.typeModel && item.typeModel.toLowerCase().includes(q)) ||
        (item.serialNumber && item.serialNumber.toLowerCase().includes(q)) ||
        (item.distributor && item.distributor.toLowerCase().includes(q)) ||
        (item.fundingSource && item.fundingSource.toLowerCase().includes(q)) ||
        (item.aklAkd && item.aklAkd.toLowerCase().includes(q)) ||
        (item.location && item.location.toLowerCase().includes(q)) ||
        (item.description && item.description.toLowerCase().includes(q));

      const matchCategory = selectedCategory === 'ALL' || item.category === selectedCategory;
      const matchCondition = selectedCondition === 'ALL' || item.conditionStatus === selectedCondition;
      
      let matchAvail = true;
      if (selectedAvailability === 'ADA') matchAvail = item.isAvailable !== false;
      if (selectedAvailability === 'TIDAK') matchAvail = item.isAvailable === false;

      return matchSearch && matchCategory && matchCondition && matchAvail;
    });
  }, [items, searchTerm, selectedCategory, selectedCondition, selectedAvailability]);

  const totalFilteredValue = filteredItems.reduce((acc, curr) => acc + ((curr.pricePerUnit || 0) * curr.currentStock), 0);
  const totalFilteredUnits = filteredItems.reduce((acc, curr) => acc + curr.currentStock, 0);

  const formatRupiah = (val?: number) => {
    if (!val) return 'Rp 0';
    return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(val);
  };

  const handleExportCSV = () => {
    const headers = [
      'No',
      'Status Keberadaan (Ada)',
      'No Seri',
      'Kode Barcode / SKU',
      'Nama Barang',
      'Kategori',
      'Merk',
      'Type / Model',
      'Tahun Pengadaan',
      'Kondisi (Berfungsi)',
      'Jumlah Stok',
      'Satuan',
      'Harga Satuan (Rp)',
      'Total Nilai (Rp)',
      'Lokasi Ruangan',
      'Sumber Pendanaan',
      'Distributor / Rekanan',
      'AKL / AKD',
      'Keterangan'
    ];

    const rows = filteredItems.map((item, idx) => [
      (idx + 1).toString(),
      item.isAvailable !== false ? 'ADA' : 'TIDAK',
      `"${item.serialNumber || '-'}"`,
      `"${item.sku}"`,
      `"${item.name.replace(/"/g, '""')}"`,
      `"${item.category}"`,
      `"${item.merk || '-'}"`,
      `"${item.typeModel || '-'}"`,
      item.procurementYear || '-',
      item.conditionStatus || 'Berfungsi',
      item.currentStock,
      item.unit,
      item.pricePerUnit || 0,
      (item.pricePerUnit || 0) * item.currentStock,
      `"${item.location || '-'}"`,
      `"${item.fundingSource || '-'}"`,
      `"${item.distributor || '-'}"`,
      `"${item.aklAkd || '-'}"`,
      `"${(item.description || '-').replace(/"/g, '""')}"`,
    ]);

    const csvContent = 'data:text/csv;charset=utf-8,\uFEFF' + [headers.join(','), ...rows.map(r => r.join(','))].join('\n');
    const encodedUri = encodeURI(csvContent);
    const link = document.createElement('a');
    link.setAttribute('href', encodedUri);
    link.setAttribute('download', `Katalog_Master_Aset_${new Date().toISOString().slice(0, 10)}.csv`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  return (
    <div className="space-y-4">
      {/* Header and Add Button */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-white p-5 sm:p-6 rounded-3xl border border-slate-200/90 shadow-xs">
        <div>
          <div className="flex items-center gap-2.5 flex-wrap">
            <h2 className="text-xl sm:text-2xl font-black text-slate-900 tracking-tight">
              Master Katalog Aset & Inventaris Kantor
            </h2>
            <span className="px-2.5 py-0.5 rounded-full text-xs font-bold bg-emerald-50 text-emerald-800 border border-emerald-200">
              {items.length} Aset Terdaftar
            </span>
          </div>
          <p className="text-xs text-slate-500 mt-1">
            Data spesifikasi inventaris lengkap terpadu (No Seri, Merk, Type, Tahun, Kondisi Berfungsi, Nilai Perolehan, Lokasi, Pendanaan, AKL/AKD, Foto & Geotag GPS)
          </p>
        </div>

        <div className="flex items-center gap-2 flex-wrap">
          {/* View Toggle */}
          <div className="flex items-center p-1 bg-slate-100 rounded-xl border border-slate-200">
            <button
              onClick={() => setViewMode('table')}
              className={`p-1.5 rounded-lg transition-colors cursor-pointer ${
                viewMode === 'table' ? 'bg-white text-indigo-700 shadow-xs' : 'text-slate-400 hover:text-slate-700'
              }`}
              title="Tampilan Tabel Detail"
            >
              <List className="w-4 h-4" />
            </button>
            <button
              onClick={() => setViewMode('grid')}
              className={`p-1.5 rounded-lg transition-colors cursor-pointer ${
                viewMode === 'grid' ? 'bg-white text-indigo-700 shadow-xs' : 'text-slate-400 hover:text-slate-700'
              }`}
              title="Tampilan Grid Kartu Visual"
            >
              <LayoutGrid className="w-4 h-4" />
            </button>
          </div>

          <button
            onClick={handleExportCSV}
            className="inline-flex items-center gap-1.5 px-3 py-2 bg-slate-100 hover:bg-slate-200 text-slate-700 font-bold text-xs rounded-xl transition-colors cursor-pointer border border-slate-200"
            title="Ekspor Data ke CSV Excel"
          >
            <Download className="w-3.5 h-3.5" />
            <span>CSV</span>
          </button>

          <button
            onClick={onOpenPDFReport}
            className="inline-flex items-center gap-1.5 px-3.5 py-2 bg-rose-50 hover:bg-rose-100 text-rose-700 font-bold text-xs rounded-xl transition-colors cursor-pointer border border-rose-200"
            title="Unduh Rekap Laporan PDF"
          >
            <FileText className="w-3.5 h-3.5 text-rose-600" />
            <span>Rekap PDF</span>
          </button>

          <button
            onClick={onOpenNewItem}
            className="inline-flex items-center gap-2 px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-white font-black text-xs sm:text-sm rounded-xl shadow-md shadow-emerald-600/30 active:scale-95 transition-all cursor-pointer ring-1 ring-emerald-400/40"
          >
            <Plus className="w-4 h-4" />
            <span>+ Input Barang Baru Masuk</span>
          </button>
        </div>
      </div>

      {/* Filter and Search Bar */}
      <div className="bg-white p-4 sm:p-5 rounded-3xl border border-slate-200/90 shadow-xs space-y-3">
        <div className="grid grid-cols-1 md:grid-cols-12 gap-3">
          {/* Search box */}
          <div className="md:col-span-5 relative">
            <Search className="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              placeholder="Cari merk, type, no seri, distributor, AKL/AKD, nama barang, lokasi..."
              value={searchTerm}
              onChange={e => setSearchTerm(e.target.value)}
              className="w-full pl-10 pr-9 py-2 text-xs sm:text-sm bg-slate-50 border border-slate-200 rounded-xl focus:outline-hidden focus:ring-2 focus:ring-indigo-500 font-medium"
            />
            {searchTerm && (
              <button
                onClick={() => setSearchTerm('')}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600"
              >
                <X className="w-3.5 h-3.5" />
              </button>
            )}
          </div>

          {/* Category Filter */}
          <div className="md:col-span-3">
            <select
              value={selectedCategory}
              onChange={e => setSelectedCategory(e.target.value)}
              className="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-200 rounded-xl focus:ring-2 focus:ring-indigo-500 font-bold text-slate-700"
            >
              <option value="ALL">Semua Kategori ({categories.length})</option>
              {categories.map(c => (
                <option key={c} value={c}>{c}</option>
              ))}
            </select>
          </div>

          {/* Condition Filter (Berfungsi) */}
          <div className="md:col-span-2">
            <select
              value={selectedCondition}
              onChange={e => setSelectedCondition(e.target.value)}
              className="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-200 rounded-xl focus:ring-2 focus:ring-indigo-500 font-bold text-slate-700"
            >
              <option value="ALL">Semua Kondisi</option>
              <option value="Berfungsi">Berfungsi (Baik)</option>
              <option value="Rusak Ringan">Rusak Ringan</option>
              <option value="Rusak Berat">Rusak Berat</option>
              <option value="Perlu Kalibrasi">Perlu Kalibrasi</option>
            </select>
          </div>

          {/* Availability Filter (Ada) */}
          <div className="md:col-span-2">
            <select
              value={selectedAvailability}
              onChange={e => setSelectedAvailability(e.target.value as any)}
              className="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-200 rounded-xl focus:ring-2 focus:ring-indigo-500 font-bold text-slate-700"
            >
              <option value="ALL">Semua Keberadaan</option>
              <option value="ADA">Ada (Tersedia)</option>
              <option value="TIDAK">Tidak Ada</option>
            </select>
          </div>
        </div>

        {/* Filter Summary & Total Count */}
        <div className="flex items-center justify-between text-xs pt-2 border-t border-slate-100 text-slate-500 flex-wrap gap-2">
          <div className="flex items-center gap-3">
            <span>Ditemukan: <strong className="text-slate-900 font-bold">{filteredItems.length}</strong> jenis aset</span>
            <span>•</span>
            <span>Total Fisik: <strong className="text-emerald-700 font-bold">{totalFilteredUnits} unit</strong></span>
            <span>•</span>
            <span>Nilai Aset: <strong className="text-slate-900 font-bold">{formatRupiah(totalFilteredValue)}</strong></span>
          </div>

          {(searchTerm || selectedCategory !== 'ALL' || selectedCondition !== 'ALL' || selectedAvailability !== 'ALL') && (
            <button
              onClick={() => {
                setSearchTerm('');
                setSelectedCategory('ALL');
                setSelectedCondition('ALL');
                setSelectedAvailability('ALL');
              }}
              className="text-indigo-600 hover:underline font-bold cursor-pointer"
            >
              Reset Filter
            </button>
          )}
        </div>
      </div>

      {/* Main View: Table View OR Visual Grid View */}
      {viewMode === 'table' ? (
        /* TABLE VIEW - Aligned with Indonesian Procurement Standard */
        <div className="bg-white rounded-3xl border border-slate-200/90 shadow-xs overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse text-xs">
              <thead>
                <tr className="bg-slate-900 text-white font-bold uppercase tracking-wider text-[11px] border-b border-slate-800">
                  <th className="py-3 px-3 text-center">Foto & Geotag</th>
                  <th className="py-3 px-3 text-center bg-rose-600 text-white">Ada</th>
                  <th className="py-3 px-3 text-center bg-rose-700 text-white">No Seri</th>
                  <th className="py-3 px-3">Nama Barang & SKU</th>
                  <th className="py-3 px-3">Merk</th>
                  <th className="py-3 px-3">Type</th>
                  <th className="py-3 px-3 text-center">Thn</th>
                  <th className="py-3 px-3 text-center">Kondisi</th>
                  <th className="py-3 px-3 text-center bg-indigo-900 text-white">Stok Fisik</th>
                  <th className="py-3 px-3 text-right">Harga Satuan</th>
                  <th className="py-3 px-3 text-right">Total Nilai</th>
                  <th className="py-3 px-3">Lokasi / Ruangan</th>
                  <th className="py-3 px-3">Pendanaan</th>
                  <th className="py-3 px-3">Distributor</th>
                  <th className="py-3 px-3">AKL / AKD</th>
                  <th className="py-3 px-3 text-center">Aksi</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100 text-[11px]">
                {filteredItems.length === 0 ? (
                  <tr>
                    <td colSpan={16} className="py-14 text-center text-slate-400">
                      <AlertCircle className="w-8 h-8 mx-auto text-slate-300 mb-2" />
                      <p className="font-bold text-slate-700 text-sm">Tidak ada barang yang cocok</p>
                      <p className="text-xs text-slate-400 mt-1">Coba sesuaikan kata kunci pencarian atau reset filter</p>
                    </td>
                  </tr>
                ) : (
                  filteredItems.map(item => {
                    const isAvailable = item.isAvailable !== false;
                    const isFunctional = item.conditionStatus === 'Berfungsi' || !item.conditionStatus;
                    const totalItemVal = (item.pricePerUnit || 0) * item.currentStock;

                    return (
                      <tr key={item.id} className="hover:bg-slate-50/80 transition-colors">
                        {/* Foto & Geotag */}
                        <td className="py-2.5 px-3 text-center">
                          {item.photoUrl ? (
                            <div 
                              onClick={() => setPreviewPhoto({ 
                                url: item.photoUrl!, 
                                title: `${item.name} (${item.sku})`, 
                                geoTag: item.geoTag, 
                                driveLink: item.driveFileLink 
                              })}
                              className="relative w-12 h-12 rounded-xl overflow-hidden border border-slate-300 mx-auto cursor-pointer hover:scale-105 transition-transform bg-slate-100 shadow-2xs group"
                            >
                              <img src={item.photoUrl} alt={item.name} className="w-full h-full object-cover" />
                              {item.geoTag && (
                                <div className="absolute bottom-0 inset-x-0 bg-emerald-600/90 text-white text-[7px] font-bold text-center py-0.5">
                                  GPS
                                </div>
                              )}
                            </div>
                          ) : (
                            <div className="w-12 h-12 rounded-xl border border-dashed border-slate-200 bg-slate-50 flex items-center justify-center text-slate-300 mx-auto">
                              <ImageIcon className="w-4 h-4" />
                            </div>
                          )}
                        </td>

                        {/* Ada (Keberadaan Fisik) */}
                        <td className="py-2.5 px-3 text-center whitespace-nowrap">
                          <span className={`px-2 py-0.5 rounded-md font-black text-[10px] ${
                            isAvailable ? 'bg-emerald-100 text-emerald-800' : 'bg-rose-100 text-rose-800'
                          }`}>
                            {isAvailable ? 'ADA' : 'TIDAK'}
                          </span>
                        </td>

                        {/* No Seri */}
                        <td className="py-2.5 px-3 font-mono font-bold text-slate-900 whitespace-nowrap text-center">
                          <span className="px-2 py-0.5 rounded bg-slate-100 border border-slate-200 text-[10px]">
                            {item.serialNumber || '-'}
                          </span>
                        </td>

                        {/* Nama Barang & SKU Barcode */}
                        <td className="py-2.5 px-3 min-w-[180px]">
                          <p className="font-extrabold text-slate-900 leading-snug">{item.name}</p>
                          <div className="font-mono text-[10px] text-indigo-700 font-bold mt-0.5 flex items-center gap-1.5 flex-wrap">
                            <span className="bg-indigo-50 px-1.5 py-0.2 rounded border border-indigo-200">{item.sku}</span>
                            <span className="text-slate-400 font-sans">• {item.category}</span>
                          </div>
                        </td>

                        {/* Merk */}
                        <td className="py-2.5 px-3 font-bold text-slate-800 whitespace-nowrap">
                          {item.merk || '-'}
                        </td>

                        {/* Type / Model */}
                        <td className="py-2.5 px-3 font-medium text-slate-700 whitespace-nowrap">
                          {item.typeModel || '-'}
                        </td>

                        {/* Thn Pengadaan */}
                        <td className="py-2.5 px-3 text-center font-bold text-slate-700 whitespace-nowrap">
                          {item.procurementYear || '-'}
                        </td>

                        {/* Berfungsi / Kondisi */}
                        <td className="py-2.5 px-3 text-center whitespace-nowrap">
                          <span className={`px-2 py-0.5 rounded-full font-bold text-[10px] ${
                            isFunctional
                              ? 'bg-emerald-100 text-emerald-800'
                              : item.conditionStatus === 'Rusak Ringan'
                              ? 'bg-amber-100 text-amber-800'
                              : 'bg-rose-100 text-rose-800'
                          }`}>
                            {item.conditionStatus || 'Berfungsi'}
                          </span>
                        </td>

                        {/* Stok Fisik Saat Ini & Satuan */}
                        <td className="py-2.5 px-3 text-center whitespace-nowrap">
                          <span className="font-black text-slate-900 text-xs px-2 py-0.5 rounded-lg bg-indigo-50 border border-indigo-200 text-indigo-900">
                            {item.currentStock} {item.unit}
                          </span>
                        </td>

                        {/* Harga Satuan */}
                        <td className="py-2.5 px-3 text-right font-bold text-slate-900 whitespace-nowrap font-mono">
                          {formatRupiah(item.pricePerUnit)}
                        </td>

                        {/* Total Nilai Aset */}
                        <td className="py-2.5 px-3 text-right font-black text-indigo-950 whitespace-nowrap font-mono">
                          {formatRupiah(totalItemVal)}
                        </td>

                        {/* Lokasi / Ruangan */}
                        <td className="py-2.5 px-3 whitespace-nowrap font-medium text-slate-700">
                          <span className="flex items-center gap-1">
                            <MapPin className="w-3 h-3 text-slate-400" />
                            <span>{item.location}</span>
                          </span>
                        </td>

                        {/* Pendanaan */}
                        <td className="py-2.5 px-3 whitespace-nowrap">
                          <span className="px-2 py-0.5 rounded bg-slate-100 text-slate-700 font-semibold text-[10px]">
                            {item.fundingSource || '-'}
                          </span>
                        </td>

                        {/* Distributor */}
                        <td className="py-2.5 px-3 text-slate-800 max-w-[130px] truncate" title={item.distributor}>
                          {item.distributor || '-'}
                        </td>

                        {/* AKL/AKD */}
                        <td className="py-2.5 px-3 font-mono text-[10px] font-bold text-slate-700 whitespace-nowrap">
                          {item.aklAkd || '-'}
                        </td>

                        {/* Aksi */}
                        <td className="py-2.5 px-3 text-center whitespace-nowrap">
                          <div className="flex items-center justify-center gap-1">
                            <button
                              onClick={() => onOpenBarcodeModal(item)}
                              className="p-1.5 rounded-lg bg-indigo-50 hover:bg-indigo-100 text-indigo-700 cursor-pointer transition-colors"
                              title="Cetak Barcode Label"
                            >
                              <Barcode className="w-3.5 h-3.5" />
                            </button>
                            <button
                              onClick={() => onEditItem(item)}
                              className="p-1.5 rounded-lg text-slate-500 hover:text-slate-900 hover:bg-slate-100 cursor-pointer transition-colors"
                              title="Ubah Data Lengkap Aset"
                            >
                              <Edit3 className="w-3.5 h-3.5" />
                            </button>
                            <button
                              onClick={() => onDeleteItem(item)}
                              className="p-1.5 rounded-lg text-slate-400 hover:text-rose-600 hover:bg-rose-50 cursor-pointer transition-colors"
                              title="Hapus Aset"
                            >
                              <Trash2 className="w-3.5 h-3.5" />
                            </button>
                          </div>
                        </td>
                      </tr>
                    );
                  })
                )}
              </tbody>
            </table>
          </div>
        </div>
      ) : (
        /* GRID CARD VIEW - Modern Visual Asset Cards */
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {filteredItems.map(item => {
            const isAvailable = item.isAvailable !== false;
            const isFunctional = item.conditionStatus === 'Berfungsi' || !item.conditionStatus;

            return (
              <div 
                key={item.id}
                className="bg-white rounded-3xl p-4 border border-slate-200/90 shadow-xs hover:shadow-md hover:border-indigo-300 transition-all flex flex-col justify-between group"
              >
                <div>
                  {/* Photo & GPS Badge */}
                  <div className="relative w-full h-40 rounded-2xl overflow-hidden bg-slate-100 border border-slate-200 mb-3">
                    {item.photoUrl ? (
                      <img 
                        src={item.photoUrl} 
                        alt={item.name} 
                        className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
                      />
                    ) : (
                      <div className="w-full h-full flex flex-col items-center justify-center text-slate-300">
                        <ImageIcon className="w-8 h-8 mb-1" />
                        <span className="text-[10px]">Belum Ada Foto</span>
                      </div>
                    )}

                    {/* Overlay Badges */}
                    <div className="absolute top-2 left-2 flex items-center gap-1">
                      <span className={`px-2 py-0.5 rounded-md font-black text-[9px] ${
                        isAvailable ? 'bg-emerald-600 text-white' : 'bg-rose-600 text-white'
                      }`}>
                        {isAvailable ? 'ADA' : 'TIDAK'}
                      </span>
                      {item.geoTag && (
                        <span className="px-1.5 py-0.5 rounded-md bg-slate-900/80 text-emerald-400 font-bold text-[9px] backdrop-blur-xs flex items-center gap-0.5">
                          <MapPin className="w-2.5 h-2.5" />
                          <span>GPS</span>
                        </span>
                      )}
                    </div>

                    <div className="absolute top-2 right-2">
                      <span className={`px-2 py-0.5 rounded-full font-bold text-[9px] ${
                        isFunctional ? 'bg-emerald-100 text-emerald-800' : 'bg-rose-100 text-rose-800'
                      }`}>
                        {item.conditionStatus || 'Berfungsi'}
                      </span>
                    </div>

                    <div className="absolute bottom-2 left-2 right-2">
                      <div className="bg-slate-950/80 backdrop-blur-xs px-2.5 py-1 rounded-xl text-white flex items-center justify-between text-[10px]">
                        <span className="font-mono font-bold text-indigo-300">{item.sku}</span>
                        <span className="font-black text-emerald-400">{item.currentStock} {item.unit}</span>
                      </div>
                    </div>
                  </div>

                  {/* Title & Metadata */}
                  <h3 className="font-extrabold text-slate-900 text-sm leading-snug line-clamp-2">
                    {item.name}
                  </h3>
                  
                  <div className="mt-1.5 space-y-1 text-xs">
                    <div className="flex items-center justify-between text-slate-500">
                      <span>Merk / Type:</span>
                      <strong className="text-slate-800 font-bold truncate max-w-[130px]">
                        {[item.merk, item.typeModel].filter(Boolean).join(' ') || '-'}
                      </strong>
                    </div>

                    <div className="flex items-center justify-between text-slate-500">
                      <span>No. Seri:</span>
                      <strong className="font-mono text-slate-800 truncate max-w-[130px]">
                        {item.serialNumber || '-'}
                      </strong>
                    </div>

                    <div className="flex items-center justify-between text-slate-500">
                      <span>Lokasi:</span>
                      <span className="text-slate-700 truncate max-w-[130px] font-medium">{item.location}</span>
                    </div>

                    <div className="flex items-center justify-between text-slate-500 pt-1 border-t border-slate-100">
                      <span>Harga Satuan:</span>
                      <span className="font-mono font-black text-slate-900">{formatRupiah(item.pricePerUnit)}</span>
                    </div>
                  </div>
                </div>

                {/* Card Footer Actions */}
                <div className="mt-4 pt-3 border-t border-slate-100 flex items-center justify-between gap-1.5">
                  <button
                    onClick={() => onOpenBarcodeModal(item)}
                    className="flex-1 py-1.5 px-2 bg-indigo-50 hover:bg-indigo-100 text-indigo-700 font-bold text-xs rounded-xl flex items-center justify-center gap-1 transition-colors cursor-pointer"
                  >
                    <Barcode className="w-3.5 h-3.5" />
                    <span>Barcode</span>
                  </button>

                  <button
                    onClick={() => onEditItem(item)}
                    className="p-1.5 rounded-xl text-slate-500 hover:text-slate-900 hover:bg-slate-100 cursor-pointer transition-colors"
                    title="Edit Data"
                  >
                    <Edit3 className="w-4 h-4" />
                  </button>

                  <button
                    onClick={() => onDeleteItem(item)}
                    className="p-1.5 rounded-xl text-slate-400 hover:text-rose-600 hover:bg-rose-50 cursor-pointer transition-colors"
                    title="Hapus"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Lightbox Photo Preview Modal with Geotag & Google Drive Link */}
      {previewPhoto && (
        <div 
          onClick={() => setPreviewPhoto(null)}
          className="fixed inset-0 z-50 bg-slate-950/80 backdrop-blur-md flex items-center justify-center p-4 animate-in fade-in duration-200"
        >
          <div className="bg-white rounded-3xl p-5 max-w-lg w-full overflow-hidden shadow-2xl relative" onClick={e => e.stopPropagation()}>
            <div className="flex items-center justify-between pb-3 border-b border-slate-100">
              <div>
                <h4 className="font-bold text-slate-900 text-sm truncate">{previewPhoto.title}</h4>
                <p className="text-[10px] text-slate-500">Dokumentasi Foto Fisik & Verifikasi Geotagging GPS</p>
              </div>
              <button 
                onClick={() => setPreviewPhoto(null)}
                className="p-1.5 text-slate-400 hover:text-slate-700 rounded-lg cursor-pointer"
              >
                ✕
              </button>
            </div>

            <div className="mt-3 rounded-2xl overflow-hidden bg-slate-900 max-h-[60vh] flex items-center justify-center relative">
              <img src={previewPhoto.url} alt={previewPhoto.title} className="max-w-full max-h-[60vh] object-contain" />
            </div>

            {/* Geotag info & Drive Link buttons */}
            <div className="mt-4 pt-3 border-t border-slate-100 flex items-center justify-between gap-2 flex-wrap">
              {previewPhoto.geoTag ? (
                <a
                  href={`https://www.google.com/maps?q=${previewPhoto.geoTag.latitude},${previewPhoto.geoTag.longitude}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-emerald-50 text-emerald-800 text-xs font-bold border border-emerald-200 hover:bg-emerald-100 transition-colors"
                >
                  <MapPin className="w-3.5 h-3.5 text-emerald-600" />
                  <span>Lihat Lokasi GPS di Google Maps</span>
                  <ExternalLink className="w-3 h-3 text-slate-400" />
                </a>
              ) : (
                <span className="text-xs text-slate-400">Tidak ada koordinat GPS</span>
              )}

              {previewPhoto.driveLink ? (
                <a
                  href={previewPhoto.driveLink}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-indigo-50 text-indigo-700 text-xs font-bold border border-indigo-200 hover:bg-indigo-100 transition-colors"
                >
                  <HardDrive className="w-3.5 h-3.5 text-indigo-600" />
                  <span>Buka di Google Drive</span>
                  <ExternalLink className="w-3 h-3 text-indigo-500" />
                </a>
              ) : (
                <span className="text-[11px] text-slate-400">Tersimpan di Cloud Database</span>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
