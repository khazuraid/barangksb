import React, { useState, useEffect } from 'react';
import { 
  X, 
  ArrowDownLeft, 
  AlertCircle, 
  Check, 
  Package, 
  ShieldCheck, 
  Building2, 
  Calendar, 
  Layers,
  Sparkles,
  Plus,
  Tag,
  DollarSign,
  FileCheck2,
  MapPin,
  Barcode,
  Camera,
  HardDrive
} from 'lucide-react';
import { 
  NewIncomingItemData, 
  UnitType, 
  ConditionStatus, 
  GeoTagData,
  DEFAULT_CATEGORIES
} from '../types/inventory';
import { generateSKU } from '../services/storageService';
import { PhotoUploader } from './PhotoUploader';

interface StockMovementModalProps {
  isOpen: boolean;
  onClose: () => void;
  categories: string[];
  onAddNewCategory?: (name: string) => void;
  itemsCount: number;
  initialSku?: string;
  initialData?: Partial<NewIncomingItemData>;
  onSubmit: (data: NewIncomingItemData) => void;
  currentOfficerName?: string;
}

const UNITS: UnitType[] = ['Unit', 'Pcs', 'Set', 'Box', 'Rim', 'Pack', 'Dus', 'Botol', 'Roll'];
const CONDITIONS: ConditionStatus[] = ['Berfungsi', 'Rusak Ringan', 'Rusak Berat', 'Perlu Kalibrasi'];
const FUNDING_OPTIONS = ['APBD', 'APBN', 'DAK Kesehatan', 'Kas Operasional Kantor', 'Yayasan / Donasi', 'Mandiri'];

export const StockMovementModal: React.FC<StockMovementModalProps> = ({
  isOpen,
  onClose,
  categories,
  onAddNewCategory,
  itemsCount,
  initialSku,
  initialData,
  onSubmit,
  currentOfficerName = 'Ahmad Faisal (GA/Logistik)',
}) => {
  // 1. Identitas Barang
  const [sku, setSku] = useState<string>('');
  const [name, setName] = useState<string>('');
  const [category, setCategory] = useState<string>(categories[0] || DEFAULT_CATEGORIES[0]);
  const [location, setLocation] = useState<string>('Gudang Utama - Rak A1');

  // 2. Kuantitas & Nilai Masuk
  const [quantity, setQuantity] = useState<number>(1);
  const [unit, setUnit] = useState<UnitType>('Unit');
  const [minStock, setMinStock] = useState<number>(1);
  const [pricePerUnit, setPricePerUnit] = useState<number>(0);

  // 3. Spesifikasi Teknis
  const [merk, setMerk] = useState<string>('');
  const [typeModel, setTypeModel] = useState<string>('');
  const [serialNumber, setSerialNumber] = useState<string>('');
  const [procurementYear, setProcurementYear] = useState<string>(new Date().getFullYear().toString());
  const [conditionStatus, setConditionStatus] = useState<ConditionStatus>('Berfungsi');
  const [isAvailable, setIsAvailable] = useState<boolean>(true);

  // 4. Legalitas & Pengadaan
  const [fundingSource, setFundingSource] = useState<string>('APBD');
  const [distributor, setDistributor] = useState<string>('PT Saba Indomedika');
  const [aklAkd, setAklAkd] = useState<string>('');

  // 5. Dokumen & Bukti Fisik
  const [receivedBy, setReceivedBy] = useState<string>(currentOfficerName);
  const [invoiceOrPoNumber, setInvoiceOrPoNumber] = useState<string>('');
  const [notes, setNotes] = useState<string>('');
  const [photoUrl, setPhotoUrl] = useState<string | undefined>(undefined);
  const [geoTag, setGeoTag] = useState<GeoTagData | undefined>(undefined);

  // Custom Category Input
  const [showNewCatInput, setShowNewCatInput] = useState<boolean>(false);
  const [newCatName, setNewCatName] = useState<string>('');
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (isOpen) {
      const activeCat = categories[0] || DEFAULT_CATEGORIES[0];
      setCategory(activeCat);

      if (initialSku) {
        setSku(initialSku);
      } else if (initialData?.sku) {
        setSku(initialData.sku);
      } else {
        setSku(generateSKU(activeCat, itemsCount));
      }

      setName(initialData?.name || '');
      setLocation(initialData?.location || 'Gudang Utama - Rak A1');
      setQuantity(initialData?.quantity || 1);
      setUnit(initialData?.unit || 'Unit');
      setMinStock(initialData?.minStock || 1);
      setPricePerUnit(initialData?.pricePerUnit || 0);

      setMerk(initialData?.merk || '');
      setTypeModel(initialData?.typeModel || '');
      setSerialNumber(initialData?.serialNumber || `SN-${new Date().getFullYear().toString().slice(-2)}${Math.floor(1000 + Math.random() * 9000)}`);
      setProcurementYear(initialData?.procurementYear ? initialData.procurementYear.toString() : new Date().getFullYear().toString());
      setConditionStatus(initialData?.conditionStatus || 'Berfungsi');
      setIsAvailable(initialData?.isAvailable !== false);
      setFundingSource(initialData?.fundingSource || 'APBD');
      setDistributor(initialData?.distributor || 'PT Saba Indomedika');
      setAklAkd(initialData?.aklAkd || '');

      setReceivedBy(initialData?.receivedBy || currentOfficerName);
      setInvoiceOrPoNumber(initialData?.invoiceOrPoNumber || `PO-${new Date().getFullYear()}-${Math.floor(100 + Math.random() * 900)}`);
      setNotes(initialData?.notes || '');
      setPhotoUrl(initialData?.photoUrl || undefined);
      setGeoTag(initialData?.geoTag || undefined);

      setShowNewCatInput(false);
      setNewCatName('');
      setError(null);
    }
  }, [isOpen, initialSku, initialData, itemsCount, categories, currentOfficerName]);

  if (!isOpen) return null;

  const handleCategoryChange = (newCat: string) => {
    setCategory(newCat);
    if (!initialSku && !initialData?.sku) {
      setSku(generateSKU(newCat, itemsCount));
    }
  };

  const handleCreateCategory = () => {
    const trimmed = newCatName.trim();
    if (!trimmed) return;
    if (onAddNewCategory) {
      onAddNewCategory(trimmed);
    }
    setCategory(trimmed);
    setSku(generateSKU(trimmed, itemsCount));
    setShowNewCatInput(false);
    setNewCatName('');
  };

  const handleAutoGenerateSKU = () => {
    setSku(generateSKU(category, itemsCount));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) {
      setError('Nama barang baru wajib diisi!');
      return;
    }
    if (!sku.trim()) {
      setError('Kode Barcode / SKU wajib diisi!');
      return;
    }
    if (quantity <= 0) {
      setError('Jumlah barang baru masuk harus minimal 1 unit!');
      return;
    }
    // Mandatory photo verification for incoming stock with geotagging
    if (!photoUrl) {
      setError('Wajib melampirkan foto fisik barang masuk beserta verifikasi geotagging GPS!');
      return;
    }

    const payload: NewIncomingItemData = {
      sku: sku.trim(),
      name: name.trim(),
      category,
      location: location.trim() || 'Gudang Utama',
      quantity: Number(quantity) || 1,
      minStock: Number(minStock) || 1,
      unit,
      pricePerUnit: Number(pricePerUnit) || 0,
      description: notes.trim(),
      photoUrl,
      geoTag,
      driveFileLink: 'https://drive.google.com/drive/folders/inventaris-kantor',
      merk: merk.trim(),
      typeModel: typeModel.trim(),
      serialNumber: serialNumber.trim(),
      procurementYear: procurementYear ? Number(procurementYear) : new Date().getFullYear(),
      conditionStatus,
      isAvailable,
      fundingSource,
      distributor: distributor.trim(),
      aklAkd: aklAkd.trim(),
      receivedBy: receivedBy.trim(),
      invoiceOrPoNumber: invoiceOrPoNumber.trim(),
      notes: notes.trim(),
    };

    onSubmit(payload);
  };

  const totalCalculatedValue = (Number(pricePerUnit) || 0) * (Number(quantity) || 1);

  const formatRupiah = (val?: number | null) => {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      maximumFractionDigits: 0,
    }).format(Number(val) || 0);
  };

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto bg-slate-950/75 backdrop-blur-md flex items-center justify-center p-3 sm:p-5">
      <div className="bg-white rounded-3xl max-w-4xl w-full max-h-[92vh] flex flex-col shadow-2xl border border-slate-100 relative animate-in fade-in zoom-in-95 duration-200 overflow-hidden">
        {/* Header Modal */}
        <div className="p-5 sm:p-6 bg-slate-950 text-white flex items-center justify-between border-b border-slate-800 shrink-0">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-2xl bg-emerald-500 text-slate-950 flex items-center justify-center shadow-lg shadow-emerald-500/20 font-black">
              <ArrowDownLeft className="w-6 h-6 text-slate-950" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h3 className="text-lg sm:text-xl font-black tracking-tight text-white">
                  Formulir Penerimaan Barang Baru Masuk (+)
                </h3>
                <span className="px-2 py-0.5 rounded-full text-[10px] font-bold bg-emerald-950 text-emerald-300 border border-emerald-700">
                  Registrasi Aset Baru
                </span>
              </div>
              <p className="text-xs text-slate-400 mt-0.5">
                Pencatatan lengkap pengadaan barang baru beserta verifikasi foto fisik & koordinat GPS geotagging
              </p>
            </div>
          </div>

          <button
            onClick={onClose}
            className="p-2 text-slate-400 hover:text-white hover:bg-slate-800 rounded-xl transition-colors cursor-pointer"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Scrollable Form Body */}
        <form onSubmit={handleSubmit} className="overflow-y-auto p-5 sm:p-7 space-y-6 flex-1 text-xs sm:text-sm">
          {error && (
            <div className="p-3.5 bg-rose-50 border border-rose-200 text-rose-800 rounded-2xl flex items-center gap-2.5 text-xs font-semibold">
              <AlertCircle className="w-4 h-4 text-rose-600 shrink-0" />
              <span>{error}</span>
            </div>
          )}

          {/* Section 1: Identitas & Barcode */}
          <div className="space-y-4">
            <div className="flex items-center gap-2 pb-2 border-b border-slate-100">
              <span className="w-6 h-6 rounded-lg bg-indigo-100 text-indigo-700 font-black flex items-center justify-center text-xs">
                1
              </span>
              <h4 className="font-extrabold text-slate-900 text-sm tracking-tight">
                Identitas Barang & Barcode SKU
              </h4>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {/* Nama Barang */}
              <div className="md:col-span-2">
                <label className="block text-xs font-bold text-slate-700 mb-1">
                  Nama Barang / Aset Baru <span className="text-rose-500">*</span>
                </label>
                <input
                  type="text"
                  required
                  placeholder="Contoh: USG Diagnostic Scanner B/W Mindray DP-10"
                  value={name}
                  onChange={e => setName(e.target.value)}
                  className="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-slate-900 font-semibold focus:ring-2 focus:ring-emerald-500 focus:bg-white text-xs sm:text-sm"
                />
              </div>

              {/* Kategori */}
              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">
                  Kategori Inventaris <span className="text-rose-500">*</span>
                </label>
                {!showNewCatInput ? (
                  <div className="flex gap-1.5">
                    <select
                      value={category}
                      onChange={e => handleCategoryChange(e.target.value)}
                      className="flex-1 px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl font-bold text-slate-800 focus:ring-2 focus:ring-emerald-500 text-xs"
                    >
                      {categories.map(c => (
                        <option key={c} value={c}>{c}</option>
                      ))}
                    </select>
                    <button
                      type="button"
                      onClick={() => setShowNewCatInput(true)}
                      className="px-2.5 py-2 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-xl font-bold text-xs"
                      title="Tambah Kategori Baru"
                    >
                      <Plus className="w-4 h-4" />
                    </button>
                  </div>
                ) : (
                  <div className="flex gap-1.5">
                    <input
                      type="text"
                      placeholder="Nama kategori baru..."
                      value={newCatName}
                      onChange={e => setNewCatName(e.target.value)}
                      className="flex-1 px-3 py-2 bg-white border border-indigo-300 rounded-xl text-xs font-bold text-slate-900"
                    />
                    <button
                      type="button"
                      onClick={handleCreateCategory}
                      className="px-3 py-2 bg-indigo-600 text-white rounded-xl font-bold text-xs"
                    >
                      Simpan
                    </button>
                    <button
                      type="button"
                      onClick={() => setShowNewCatInput(false)}
                      className="px-2.5 py-2 bg-slate-100 text-slate-500 rounded-xl text-xs font-bold"
                    >
                      ✕
                    </button>
                  </div>
                )}
              </div>

              {/* Barcode SKU */}
              <div>
                <div className="flex items-center justify-between mb-1">
                  <label className="text-xs font-bold text-slate-700">
                    Kode Barcode / SKU <span className="text-rose-500">*</span>
                  </label>
                  <button
                    type="button"
                    onClick={handleAutoGenerateSKU}
                    className="text-[10px] text-indigo-600 font-bold hover:underline flex items-center gap-1 cursor-pointer"
                  >
                    <Sparkles className="w-3 h-3" />
                    <span>Auto-Generate</span>
                  </button>
                </div>
                <div className="relative">
                  <Barcode className="w-4 h-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
                  <input
                    type="text"
                    required
                    value={sku}
                    onChange={e => setSku(e.target.value.toUpperCase())}
                    className="w-full pl-9 pr-3 py-2 bg-slate-50 border border-slate-200 rounded-xl font-mono font-bold text-indigo-700 text-xs focus:ring-2 focus:ring-emerald-500 focus:bg-white"
                  />
                </div>
              </div>

              {/* Lokasi Ruangan */}
              <div className="md:col-span-2">
                <label className="block text-xs font-bold text-slate-700 mb-1">
                  Lokasi Simpan / Ruangan <span className="text-rose-500">*</span>
                </label>
                <div className="relative">
                  <MapPin className="w-4 h-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
                  <input
                    type="text"
                    placeholder="Contoh: Ruang Poli Medis Lt 1 / Gudang Utama Rak A1"
                    value={location}
                    onChange={e => setLocation(e.target.value)}
                    className="w-full pl-9 pr-3 py-2 bg-slate-50 border border-slate-200 rounded-xl font-medium text-slate-900 text-xs focus:ring-2 focus:ring-emerald-500 focus:bg-white"
                  />
                </div>
              </div>
            </div>
          </div>

          {/* Section 2: Spesifikasi & Identitas Teknis */}
          <div className="space-y-4">
            <div className="flex items-center gap-2 pb-2 border-b border-slate-100">
              <span className="w-6 h-6 rounded-lg bg-indigo-100 text-indigo-700 font-black flex items-center justify-center text-xs">
                2
              </span>
              <h4 className="font-extrabold text-slate-900 text-sm tracking-tight">
                Spesifikasi Teknis & Nomor Seri
              </h4>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
              {/* Merk */}
              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">Merk / Brand</label>
                <input
                  type="text"
                  placeholder="Contoh: Mindray / Epson / Lenovo"
                  value={merk}
                  onChange={e => setMerk(e.target.value)}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl font-semibold text-slate-900 text-xs focus:ring-2 focus:ring-emerald-500 focus:bg-white"
                />
              </div>

              {/* Type / Model */}
              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">Type / Model</label>
                <input
                  type="text"
                  placeholder="Contoh: DP-10 / L3210"
                  value={typeModel}
                  onChange={e => setTypeModel(e.target.value)}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl font-medium text-slate-900 text-xs focus:ring-2 focus:ring-emerald-500 focus:bg-white"
                />
              </div>

              {/* No Seri */}
              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">No. Seri Fisik (SN)</label>
                <input
                  type="text"
                  placeholder="Contoh: SN-MND-984210"
                  value={serialNumber}
                  onChange={e => setSerialNumber(e.target.value)}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl font-mono font-bold text-slate-900 text-xs focus:ring-2 focus:ring-emerald-500 focus:bg-white"
                />
              </div>

              {/* Thn Pengadaan */}
              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">Tahun Pengadaan</label>
                <input
                  type="number"
                  value={procurementYear}
                  onChange={e => setProcurementYear(e.target.value)}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl font-bold text-slate-900 text-xs focus:ring-2 focus:ring-emerald-500 focus:bg-white"
                />
              </div>

              {/* Kondisi (Berfungsi) */}
              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">Kondisi (Berfungsi)</label>
                <select
                  value={conditionStatus}
                  onChange={e => setConditionStatus(e.target.value as ConditionStatus)}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl font-bold text-slate-800 text-xs focus:ring-2 focus:ring-emerald-500"
                >
                  {CONDITIONS.map(cond => (
                    <option key={cond} value={cond}>{cond}</option>
                  ))}
                </select>
              </div>

              {/* Status Keberadaan Ada */}
              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">Keberadaan Fisik</label>
                <select
                  value={isAvailable ? 'ADA' : 'TIDAK'}
                  onChange={e => setIsAvailable(e.target.value === 'ADA')}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl font-black text-slate-800 text-xs focus:ring-2 focus:ring-emerald-500"
                >
                  <option value="ADA">ADA (Tersedia Fisik)</option>
                  <option value="TIDAK">TIDAK (Belum Tersedia)</option>
                </select>
              </div>
            </div>
          </div>

          {/* Section 3: Kuantitas, Satuan & Nilai Perolehan */}
          <div className="space-y-4">
            <div className="flex items-center gap-2 pb-2 border-b border-slate-100">
              <span className="w-6 h-6 rounded-lg bg-indigo-100 text-indigo-700 font-black flex items-center justify-center text-xs">
                3
              </span>
              <h4 className="font-extrabold text-slate-900 text-sm tracking-tight">
                Kuantitas Masuk & Nilai Perolehan
              </h4>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-4">
              {/* Jumlah Masuk */}
              <div>
                <label className="block text-xs font-bold text-emerald-800 mb-1">
                  Jumlah Barang Masuk (+) <span className="text-rose-500">*</span>
                </label>
                <input
                  type="number"
                  min="1"
                  required
                  value={quantity}
                  onChange={e => setQuantity(Math.max(1, parseInt(e.target.value) || 1))}
                  className="w-full px-3.5 py-2.5 bg-emerald-50 border border-emerald-300 rounded-xl font-black text-emerald-900 text-base focus:ring-2 focus:ring-emerald-500"
                />
              </div>

              {/* Satuan */}
              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">Satuan</label>
                <select
                  value={unit}
                  onChange={e => setUnit(e.target.value as UnitType)}
                  className="w-full px-3 py-2.5 bg-slate-50 border border-slate-200 rounded-xl font-bold text-slate-800 text-xs focus:ring-2 focus:ring-emerald-500"
                >
                  {UNITS.map(u => (
                    <option key={u} value={u}>{u}</option>
                  ))}
                </select>
              </div>

              {/* Batas Minimum */}
              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">Batas Min. Alert</label>
                <input
                  type="number"
                  min="0"
                  value={minStock}
                  onChange={e => setMinStock(Math.max(0, parseInt(e.target.value) || 0))}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl font-bold text-slate-800 text-xs focus:ring-2 focus:ring-emerald-500"
                />
              </div>

              {/* Harga Satuan */}
              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">Harga Satuan (Rp)</label>
                <input
                  type="number"
                  min="0"
                  placeholder="0"
                  value={pricePerUnit}
                  onChange={e => setPricePerUnit(Math.max(0, parseInt(e.target.value) || 0))}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl font-mono font-bold text-slate-900 text-xs focus:ring-2 focus:ring-emerald-500"
                />
                <span className="text-[10px] text-slate-400 font-mono mt-1 block">
                  Total: {formatRupiah(totalCalculatedValue)}
                </span>
              </div>
            </div>
          </div>

          {/* Section 4: Legalitas & Sumber Pengadaan */}
          <div className="space-y-4">
            <div className="flex items-center gap-2 pb-2 border-b border-slate-100">
              <span className="w-6 h-6 rounded-lg bg-indigo-100 text-indigo-700 font-black flex items-center justify-center text-xs">
                4
              </span>
              <h4 className="font-extrabold text-slate-900 text-sm tracking-tight">
                Sumber Pendanaan & Dokumen Vendor
              </h4>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
              {/* Pendanaan */}
              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">Sumber Pendanaan</label>
                <select
                  value={fundingSource}
                  onChange={e => setFundingSource(e.target.value)}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl font-bold text-slate-800 text-xs focus:ring-2 focus:ring-emerald-500"
                >
                  {FUNDING_OPTIONS.map(opt => (
                    <option key={opt} value={opt}>{opt}</option>
                  ))}
                </select>
              </div>

              {/* Distributor / Rekanan */}
              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">Distributor / Vendor</label>
                <input
                  type="text"
                  placeholder="Contoh: PT Saba Indomedika"
                  value={distributor}
                  onChange={e => setDistributor(e.target.value)}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl font-medium text-slate-900 text-xs focus:ring-2 focus:ring-emerald-500"
                />
              </div>

              {/* AKL/AKD */}
              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">AKL / AKD / Izin Edar</label>
                <input
                  type="text"
                  placeholder="Contoh: AKL 21102910482"
                  value={aklAkd}
                  onChange={e => setAklAkd(e.target.value)}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl font-mono text-slate-900 text-xs focus:ring-2 focus:ring-emerald-500"
                />
              </div>

              {/* Petugas Penerima */}
              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">Petugas Penerima (GA)</label>
                <input
                  type="text"
                  value={receivedBy}
                  onChange={e => setReceivedBy(e.target.value)}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl font-bold text-slate-900 text-xs focus:ring-2 focus:ring-emerald-500"
                />
              </div>

              {/* No PO / Surat Jalan */}
              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">No. PO / Surat Jalan</label>
                <input
                  type="text"
                  placeholder="Contoh: PO-2024-881"
                  value={invoiceOrPoNumber}
                  onChange={e => setInvoiceOrPoNumber(e.target.value)}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl font-mono font-bold text-emerald-800 text-xs focus:ring-2 focus:ring-emerald-500"
                />
              </div>

              {/* Keterangan */}
              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">Catatan Tambahan</label>
                <input
                  type="text"
                  placeholder="Keterangan kelengkapan unit..."
                  value={notes}
                  onChange={e => setNotes(e.target.value)}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl text-slate-900 text-xs focus:ring-2 focus:ring-emerald-500"
                />
              </div>
            </div>
          </div>

          {/* Section 5: Bukti Fisik Foto dengan Geotagging GPS */}
          <div className="space-y-4">
            <div className="flex items-center gap-2 pb-2 border-b border-slate-100">
              <span className="w-6 h-6 rounded-lg bg-emerald-100 text-emerald-700 font-black flex items-center justify-center text-xs">
                5
              </span>
              <div className="flex items-center gap-2">
                <h4 className="font-extrabold text-slate-900 text-sm tracking-tight">
                  Dokumentasi Foto Fisik & Geotagging GPS
                </h4>
                <span className="px-2 py-0.5 rounded-full text-[10px] font-bold bg-rose-100 text-rose-800">
                  Wajib
                </span>
              </div>
            </div>

            <PhotoUploader
              photoUrl={photoUrl}
              geoTag={geoTag}
              onChange={(url, gTag) => {
                setPhotoUrl(url);
                setGeoTag(gTag);
                if (url && error?.includes('foto')) {
                  setError(null);
                }
              }}
              label="Ambil Foto Fisik Barang Baru Masuk"
              helperText="Foto otomatis diberi stempel watermark koordinat GPS, alamat, dan waktu resmi penerimaan"
            />
          </div>

          {/* Live Preview Card */}
          <div className="p-4 bg-slate-50 rounded-2xl border border-slate-200/90 text-xs space-y-2">
            <div className="flex items-center justify-between">
              <span className="font-bold text-slate-500 uppercase tracking-wider text-[10px]">
                Ringkasan Aset yang Akan Terdaftar
              </span>
              <span className="font-mono font-bold text-indigo-700">{sku || '-'}</span>
            </div>
            <div className="flex items-center justify-between font-bold text-slate-800">
              <span>{name || '(Nama Barang Baru)'}</span>
              <span className="text-emerald-700">+{quantity} {unit} ({formatRupiah(totalCalculatedValue)})</span>
            </div>
            <div className="flex items-center justify-between text-[11px] text-slate-500">
              <span>Merk: {merk || '-'} • No Seri: {serialNumber || '-'}</span>
              <span>Lokasi: {location}</span>
            </div>
          </div>
        </form>

        {/* Modal Footer Actions */}
        <div className="p-4 sm:p-5 bg-slate-50 border-t border-slate-200 flex items-center justify-between gap-3 shrink-0">
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2 text-xs sm:text-sm font-bold text-slate-600 hover:text-slate-900 hover:bg-slate-200 rounded-xl transition-colors cursor-pointer"
          >
            Batal
          </button>

          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={handleSubmit}
              className="inline-flex items-center gap-2 px-5 py-2.5 bg-emerald-600 hover:bg-emerald-500 text-white font-black text-xs sm:text-sm rounded-xl shadow-lg shadow-emerald-900/30 active:scale-95 transition-all cursor-pointer ring-1 ring-emerald-400"
            >
              <Check className="w-4 h-4" />
              <span>Simpan Penerimaan Barang Baru (+)</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
