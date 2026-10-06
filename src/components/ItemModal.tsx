import React, { useState, useEffect } from 'react';
import { 
  X, 
  Sparkles, 
  Package, 
  MapPin, 
  Layers, 
  AlertCircle, 
  Save, 
  Plus, 
  Tag, 
  ShieldCheck, 
  Barcode, 
  DollarSign, 
  Calendar,
  Building2,
  HardDrive
} from 'lucide-react';
import { InventoryItem, UnitType, ConditionStatus, GeoTagData } from '../types/inventory';
import { generateSKU } from '../services/storageService';
import { PhotoUploader } from './PhotoUploader';

interface ItemModalProps {
  isOpen: boolean;
  itemToEdit: InventoryItem | null;
  itemsCount: number;
  categories: string[];
  onAddNewCategory?: (name: string) => void;
  onClose: () => void;
  onSave: (item: Partial<InventoryItem>) => void;
}

const UNITS: UnitType[] = ['Unit', 'Pcs', 'Set', 'Box', 'Rim', 'Pack', 'Dus', 'Botol', 'Roll'];
const CONDITIONS: ConditionStatus[] = ['Berfungsi', 'Rusak Ringan', 'Rusak Berat', 'Perlu Kalibrasi'];
const FUNDING_OPTIONS = ['APBD', 'APBN', 'DAK Kesehatan', 'Kas Operasional Kantor', 'Yayasan / Donasi', 'Mandiri'];

export const ItemModal: React.FC<ItemModalProps> = ({
  isOpen,
  itemToEdit,
  itemsCount,
  categories,
  onAddNewCategory,
  onClose,
  onSave,
}) => {
  // 1. Identitas & Barcode
  const [sku, setSku] = useState('');
  const [name, setName] = useState('');
  const [category, setCategory] = useState<string>(categories[0] || 'Peralatan Medis & Alkes (AKL/AKD)');
  const [location, setLocation] = useState('');

  // 2. Kuantitas & Nilai
  const [currentStock, setCurrentStock] = useState<number>(1);
  const [minStock, setMinStock] = useState<number>(1);
  const [unit, setUnit] = useState<UnitType>('Unit');
  const [pricePerUnit, setPricePerUnit] = useState<number>(0);
  const [description, setDescription] = useState('');

  // 3. Spesifikasi
  const [merk, setMerk] = useState('');
  const [typeModel, setTypeModel] = useState('');
  const [serialNumber, setSerialNumber] = useState('');
  const [procurementYear, setProcurementYear] = useState<string>(new Date().getFullYear().toString());
  const [conditionStatus, setConditionStatus] = useState<ConditionStatus>('Berfungsi');
  const [isAvailable, setIsAvailable] = useState<boolean>(true);

  // 4. Legalitas & Pengadaan
  const [fundingSource, setFundingSource] = useState('APBD');
  const [distributor, setDistributor] = useState('');
  const [aklAkd, setAklAkd] = useState('');

  // 5. Foto & Geotag
  const [photoUrl, setPhotoUrl] = useState<string | undefined>(undefined);
  const [geoTag, setGeoTag] = useState<GeoTagData | undefined>(undefined);

  // Helper
  const [showNewCatInput, setShowNewCatInput] = useState(false);
  const [newCatName, setNewCatName] = useState('');
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (itemToEdit) {
      setSku(itemToEdit.sku);
      setName(itemToEdit.name);
      setCategory(itemToEdit.category);
      setLocation(itemToEdit.location);
      setCurrentStock(itemToEdit.currentStock);
      setMinStock(itemToEdit.minStock);
      setUnit(itemToEdit.unit);
      setPricePerUnit(itemToEdit.pricePerUnit || 0);
      setDescription(itemToEdit.description || '');
      setPhotoUrl(itemToEdit.photoUrl);
      setGeoTag(itemToEdit.geoTag);
      setMerk(itemToEdit.merk || '');
      setTypeModel(itemToEdit.typeModel || '');
      setSerialNumber(itemToEdit.serialNumber || '');
      setProcurementYear(itemToEdit.procurementYear ? itemToEdit.procurementYear.toString() : new Date().getFullYear().toString());
      setConditionStatus(itemToEdit.conditionStatus || 'Berfungsi');
      setFundingSource(itemToEdit.fundingSource || 'APBD');
      setDistributor(itemToEdit.distributor || '');
      setAklAkd(itemToEdit.aklAkd || '');
      setIsAvailable(itemToEdit.isAvailable !== false);
    } else {
      const defaultCat = categories[0] || 'Peralatan Medis & Alkes (AKL/AKD)';
      const newSku = generateSKU(defaultCat, itemsCount);
      setSku(newSku);
      setName('');
      setCategory(defaultCat);
      setLocation('Gudang Utama - Rak A1');
      setCurrentStock(1);
      setMinStock(1);
      setUnit('Unit');
      setPricePerUnit(0);
      setDescription('');
      setPhotoUrl(undefined);
      setGeoTag(undefined);
      setMerk('');
      setTypeModel('');
      setSerialNumber('');
      setProcurementYear(new Date().getFullYear().toString());
      setConditionStatus('Berfungsi');
      setFundingSource('APBD');
      setDistributor('');
      setAklAkd('');
      setIsAvailable(true);
    }
    setShowNewCatInput(false);
    setNewCatName('');
    setError(null);
  }, [itemToEdit, isOpen, itemsCount, categories]);

  if (!isOpen) return null;

  const handleCategoryChange = (newCat: string) => {
    setCategory(newCat);
    if (!itemToEdit) {
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
    setShowNewCatInput(false);
    setNewCatName('');
  };

  const handleAutoGenerateSKU = () => {
    setSku(generateSKU(category, itemsCount));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) {
      setError('Nama barang wajib diisi!');
      return;
    }
    if (!sku.trim()) {
      setError('Kode Barcode / SKU wajib diisi!');
      return;
    }

    onSave({
      name: name.trim(),
      sku: sku.trim(),
      category,
      location: location.trim() || 'Gudang Utama',
      currentStock: Number(currentStock) || 0,
      minStock: Number(minStock) || 0,
      unit,
      pricePerUnit: Number(pricePerUnit) || 0,
      description: description.trim(),
      photoUrl,
      geoTag,
      merk: merk.trim(),
      typeModel: typeModel.trim(),
      serialNumber: serialNumber.trim(),
      procurementYear: procurementYear ? Number(procurementYear) : undefined,
      conditionStatus,
      fundingSource,
      distributor: distributor.trim(),
      aklAkd: aklAkd.trim(),
      isAvailable,
    });

    onClose();
  };

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
        {/* Modal Header */}
        <div className="p-5 sm:p-6 bg-slate-950 text-white flex items-center justify-between border-b border-slate-800 shrink-0">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-2xl bg-indigo-600 text-white flex items-center justify-center shadow-lg shadow-indigo-600/20 font-black">
              <Package className="w-5 h-5 text-white" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h3 className="text-lg sm:text-xl font-black tracking-tight text-white">
                  {itemToEdit ? 'Ubah Data Inventaris Aset' : 'Tambah Aset ke Master'}
                </h3>
                <span className="px-2 py-0.5 rounded-full text-[10px] font-bold bg-indigo-950 text-indigo-300 border border-indigo-700">
                  {itemToEdit ? itemToEdit.sku : 'Master Baru'}
                </span>
              </div>
              <p className="text-xs text-slate-400 mt-0.5">
                Perbarui atribut inventaris sesuai format resmi pengadaan dan kelengkapan fisik
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

        {/* Form Body */}
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
              <div className="md:col-span-2">
                <label className="block text-xs font-bold text-slate-700 mb-1">
                  Nama Barang / Aset <span className="text-rose-500">*</span>
                </label>
                <input
                  type="text"
                  required
                  placeholder="Nama barang lengkap..."
                  value={name}
                  onChange={e => setName(e.target.value)}
                  className="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl font-semibold text-slate-900 focus:ring-2 focus:ring-indigo-500 focus:bg-white text-xs sm:text-sm"
                />
              </div>

              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">
                  Kategori Inventaris <span className="text-rose-500">*</span>
                </label>
                {!showNewCatInput ? (
                  <div className="flex gap-1.5">
                    <select
                      value={category}
                      onChange={e => handleCategoryChange(e.target.value)}
                      className="flex-1 px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl font-bold text-slate-800 focus:ring-2 focus:ring-indigo-500 text-xs"
                    >
                      {categories.map(c => (
                        <option key={c} value={c}>{c}</option>
                      ))}
                    </select>
                    <button
                      type="button"
                      onClick={() => setShowNewCatInput(true)}
                      className="px-2.5 py-2 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-xl font-bold text-xs"
                    >
                      <Plus className="w-4 h-4" />
                    </button>
                  </div>
                ) : (
                  <div className="flex gap-1.5">
                    <input
                      type="text"
                      placeholder="Kategori baru..."
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

              <div>
                <div className="flex items-center justify-between mb-1">
                  <label className="text-xs font-bold text-slate-700">
                    Kode Barcode / SKU <span className="text-rose-500">*</span>
                  </label>
                  {!itemToEdit && (
                    <button
                      type="button"
                      onClick={handleAutoGenerateSKU}
                      className="text-[10px] text-indigo-600 font-bold hover:underline flex items-center gap-1 cursor-pointer"
                    >
                      <Sparkles className="w-3 h-3" />
                      <span>Auto-Generate</span>
                    </button>
                  )}
                </div>
                <div className="relative">
                  <Barcode className="w-4 h-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
                  <input
                    type="text"
                    required
                    value={sku}
                    onChange={e => setSku(e.target.value.toUpperCase())}
                    className="w-full pl-9 pr-3 py-2 bg-slate-50 border border-slate-200 rounded-xl font-mono font-bold text-indigo-700 text-xs focus:ring-2 focus:ring-indigo-500 focus:bg-white"
                  />
                </div>
              </div>

              <div className="md:col-span-2">
                <label className="block text-xs font-bold text-slate-700 mb-1">
                  Lokasi Simpan / Ruangan
                </label>
                <div className="relative">
                  <MapPin className="w-4 h-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
                  <input
                    type="text"
                    value={location}
                    onChange={e => setLocation(e.target.value)}
                    className="w-full pl-9 pr-3 py-2 bg-slate-50 border border-slate-200 rounded-xl text-slate-900 text-xs focus:ring-2 focus:ring-indigo-500 focus:bg-white"
                  />
                </div>
              </div>
            </div>
          </div>

          {/* Section 2: Spesifikasi Teknis */}
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
              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">Merk</label>
                <input
                  type="text"
                  value={merk}
                  onChange={e => setMerk(e.target.value)}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl font-semibold text-slate-900 text-xs"
                />
              </div>

              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">Type / Model</label>
                <input
                  type="text"
                  value={typeModel}
                  onChange={e => setTypeModel(e.target.value)}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl text-slate-900 text-xs"
                />
              </div>

              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">No. Seri (SN)</label>
                <input
                  type="text"
                  value={serialNumber}
                  onChange={e => setSerialNumber(e.target.value)}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl font-mono font-bold text-slate-900 text-xs"
                />
              </div>

              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">Tahun Pengadaan</label>
                <input
                  type="number"
                  value={procurementYear}
                  onChange={e => setProcurementYear(e.target.value)}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl font-bold text-slate-900 text-xs"
                />
              </div>

              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">Kondisi (Berfungsi)</label>
                <select
                  value={conditionStatus}
                  onChange={e => setConditionStatus(e.target.value as ConditionStatus)}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl font-bold text-slate-800 text-xs"
                >
                  {CONDITIONS.map(cond => (
                    <option key={cond} value={cond}>{cond}</option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">Keberadaan Fisik</label>
                <select
                  value={isAvailable ? 'ADA' : 'TIDAK'}
                  onChange={e => setIsAvailable(e.target.value === 'ADA')}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl font-black text-slate-800 text-xs"
                >
                  <option value="ADA">ADA (Tersedia)</option>
                  <option value="TIDAK">TIDAK (Belum Tersedia)</option>
                </select>
              </div>
            </div>
          </div>

          {/* Section 3: Kuantitas & Nilai */}
          <div className="space-y-4">
            <div className="flex items-center gap-2 pb-2 border-b border-slate-100">
              <span className="w-6 h-6 rounded-lg bg-indigo-100 text-indigo-700 font-black flex items-center justify-center text-xs">
                3
              </span>
              <h4 className="font-extrabold text-slate-900 text-sm tracking-tight">
                Stok Fisik & Nilai Perolehan
              </h4>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-4">
              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">Stok Saat Ini</label>
                <input
                  type="number"
                  min="0"
                  value={currentStock}
                  onChange={e => setCurrentStock(parseInt(e.target.value) || 0)}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl font-black text-slate-900 text-base"
                />
              </div>

              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">Satuan</label>
                <select
                  value={unit}
                  onChange={e => setUnit(e.target.value as UnitType)}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl font-bold text-slate-800 text-xs"
                >
                  {UNITS.map(u => (
                    <option key={u} value={u}>{u}</option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">Batas Minimum</label>
                <input
                  type="number"
                  min="0"
                  value={minStock}
                  onChange={e => setMinStock(parseInt(e.target.value) || 0)}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl font-bold text-slate-800 text-xs"
                />
              </div>

              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">Harga Satuan (Rp)</label>
                <input
                  type="number"
                  min="0"
                  value={pricePerUnit}
                  onChange={e => setPricePerUnit(parseInt(e.target.value) || 0)}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl font-mono font-bold text-slate-900 text-xs"
                />
                <span className="text-[10px] text-slate-400 font-mono mt-1 block">
                  Total: {formatRupiah((pricePerUnit || 0) * (currentStock || 0))}
                </span>
              </div>
            </div>
          </div>

          {/* Section 4: Legalitas & Pendanaan */}
          <div className="space-y-4">
            <div className="flex items-center gap-2 pb-2 border-b border-slate-100">
              <span className="w-6 h-6 rounded-lg bg-indigo-100 text-indigo-700 font-black flex items-center justify-center text-xs">
                4
              </span>
              <h4 className="font-extrabold text-slate-900 text-sm tracking-tight">
                Sumber Pendanaan & Regulasi
              </h4>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">Sumber Pendanaan</label>
                <select
                  value={fundingSource}
                  onChange={e => setFundingSource(e.target.value)}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl font-bold text-slate-800 text-xs"
                >
                  {FUNDING_OPTIONS.map(opt => (
                    <option key={opt} value={opt}>{opt}</option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">Distributor / Rekanan</label>
                <input
                  type="text"
                  value={distributor}
                  onChange={e => setDistributor(e.target.value)}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl text-slate-900 text-xs"
                />
              </div>

              <div>
                <label className="block text-xs font-bold text-slate-700 mb-1">AKL / AKD</label>
                <input
                  type="text"
                  value={aklAkd}
                  onChange={e => setAklAkd(e.target.value)}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl font-mono text-slate-900 text-xs"
                />
              </div>

              <div className="sm:col-span-3">
                <label className="block text-xs font-bold text-slate-700 mb-1">Keterangan Tambahan</label>
                <textarea
                  rows={2}
                  value={description}
                  onChange={e => setDescription(e.target.value)}
                  placeholder="Keterangan kondisi atau spesifikasi..."
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl text-slate-900 text-xs"
                />
              </div>
            </div>
          </div>

          {/* Section 5: Foto & Geotag */}
          <div className="space-y-4">
            <div className="flex items-center gap-2 pb-2 border-b border-slate-100">
              <span className="w-6 h-6 rounded-lg bg-indigo-100 text-indigo-700 font-black flex items-center justify-center text-xs">
                5
              </span>
              <h4 className="font-extrabold text-slate-900 text-sm tracking-tight">
                Dokumentasi Foto Fisik & Geotagging GPS
              </h4>
            </div>

            <PhotoUploader
              photoUrl={photoUrl}
              geoTag={geoTag}
              onChange={(url, gTag) => {
                setPhotoUrl(url);
                setGeoTag(gTag);
              }}
              label="Foto Fisik Aset Inventaris"
              helperText="Foto otomatis diberi stempel koordinat GPS & waktu resmi"
            />
          </div>
        </form>

        {/* Modal Footer */}
        <div className="p-4 sm:p-5 bg-slate-50 border-t border-slate-200 flex items-center justify-between gap-3 shrink-0">
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2 text-xs sm:text-sm font-bold text-slate-600 hover:text-slate-900 rounded-xl transition-colors cursor-pointer"
          >
            Batal
          </button>

          <button
            type="button"
            onClick={handleSubmit}
            className="inline-flex items-center gap-2 px-5 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white font-bold text-xs sm:text-sm rounded-xl shadow-md shadow-indigo-600/30 transition-all cursor-pointer ring-1 ring-indigo-400"
          >
            <Save className="w-4 h-4" />
            <span>Simpan Perubahan Data</span>
          </button>
        </div>
      </div>
    </div>
  );
};
