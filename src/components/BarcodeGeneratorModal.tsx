import React, { useState, useEffect, useRef } from 'react';
import { 
  X, 
  Printer, 
  Download, 
  FileSpreadsheet, 
  Sliders, 
  Building2, 
  Calendar, 
  Check, 
  Layers, 
  QrCode, 
  Barcode as BarcodeIcon,
  RefreshCw
} from 'lucide-react';
import JsBarcode from 'jsbarcode';
import QRCode from 'qrcode';
import { InventoryItem } from '../types/inventory';

interface BarcodeGeneratorModalProps {
  isOpen: boolean;
  item: InventoryItem | null;
  allItems: InventoryItem[];
  onClose: () => void;
  onSaveToDrive?: (itemName: string, dataUrl: string) => Promise<void>;
  isDriveConnected?: boolean;
}

export const BarcodeGeneratorModal: React.FC<BarcodeGeneratorModalProps> = ({
  isOpen,
  item,
  allItems,
  onClose,
  onSaveToDrive,
  isDriveConnected = false,
}) => {
  const [selectedItemId, setSelectedItemId] = useState<string>('');
  const [barcodeType, setBarcodeType] = useState<'CODE128' | 'QR'>('CODE128');
  const [companyName, setCompanyName] = useState<string>('PT INVENTARIS NUSANTARA');
  const [labelSize, setLabelSize] = useState<'rack' | 'asset' | 'compact'>('rack');
  const [batchMode, setBatchMode] = useState<boolean>(false);
  const [isSavingToDrive, setIsSavingToDrive] = useState<boolean>(false);
  const [driveSavedMsg, setDriveSavedMsg] = useState<string | null>(null);
  const [driveErrorMsg, setDriveErrorMsg] = useState<string | null>(null);

  const barcodeSvgRef = useRef<SVGSVGElement | null>(null);
  const qrCanvasRef = useRef<HTMLCanvasElement | null>(null);

  useEffect(() => {
    if (item) {
      setSelectedItemId(item.id);
      setBatchMode(false);
    } else if (allItems.length > 0) {
      setSelectedItemId(allItems[0].id);
    }
  }, [item, allItems, isOpen]);

  const activeItem = allItems.find(i => i.id === selectedItemId) || item || allItems[0];

  // Render single barcode / QR
  useEffect(() => {
    if (!isOpen || !activeItem || batchMode) return;

    if (barcodeType === 'CODE128') {
      if (barcodeSvgRef.current) {
        try {
          JsBarcode(barcodeSvgRef.current, activeItem.sku, {
            format: 'CODE128',
            width: labelSize === 'asset' ? 1.5 : 2,
            height: labelSize === 'asset' ? 40 : 55,
            displayValue: true,
            font: 'monospace',
            fontOptions: 'bold',
            fontSize: 13,
            textMargin: 3,
            margin: 4,
          });
        } catch (err) {
          console.error('JsBarcode rendering error:', err);
        }
      }
    } else if (barcodeType === 'QR') {
      if (qrCanvasRef.current) {
        QRCode.toCanvas(
          qrCanvasRef.current,
          JSON.stringify({
            sku: activeItem.sku,
            name: activeItem.name,
            loc: activeItem.location,
            cat: activeItem.category,
          }),
          {
            width: labelSize === 'asset' ? 110 : 140,
            margin: 1,
            color: {
              dark: '#0f172a',
              light: '#ffffff',
            },
          },
          err => {
            if (err) console.error('QR error:', err);
          }
        );
      }
    }
  }, [isOpen, activeItem, barcodeType, labelSize, batchMode]);

  if (!isOpen) return null;

  // Print helper with iframe safety
  const handlePrint = () => {
    try {
      if (typeof window !== 'undefined' && typeof window.print === 'function') {
        window.print();
      }
    } catch (e) {
      console.warn('Window print restricted:', e);
    }
  };

  // Download barcode image
  const handleDownloadPNG = () => {
    if (!activeItem) return;
    if (barcodeType === 'QR' && qrCanvasRef.current) {
      const url = qrCanvasRef.current.toDataURL('image/png');
      const a = document.createElement('a');
      a.href = url;
      a.download = `QR-${activeItem.sku}.png`;
      a.click();
    } else if (barcodeSvgRef.current) {
      const svg = barcodeSvgRef.current;
      const xml = new XMLSerializer().serializeToString(svg);
      const svg64 = btoa(unescape(encodeURIComponent(xml)));
      const image64 = 'data:image/svg+xml;base64,' + svg64;

      const img = new Image();
      img.src = image64;
      img.onload = () => {
        const canvas = document.createElement('canvas');
        canvas.width = img.width || 300;
        canvas.height = img.height || 120;
        const ctx = canvas.getContext('2d');
        if (ctx) {
          ctx.fillStyle = '#ffffff';
          ctx.fillRect(0, 0, canvas.width, canvas.height);
          ctx.drawImage(img, 0, 0);
          const a = document.createElement('a');
          a.href = canvas.toDataURL('image/png');
          a.download = `Barcode-${activeItem.sku}.png`;
          a.click();
        }
      };
    }
  };

  const handleSaveToGoogleDrive = async () => {
    if (!onSaveToDrive || !activeItem) return;
    setIsSavingToDrive(true);
    setDriveSavedMsg(null);
    setDriveErrorMsg(null);
    try {
      let dataUrl = '';
      if (barcodeType === 'QR' && qrCanvasRef.current) {
        dataUrl = qrCanvasRef.current.toDataURL('image/png');
      } else if (barcodeSvgRef.current) {
        const svg = barcodeSvgRef.current;
        const xml = new XMLSerializer().serializeToString(svg);
        const svg64 = btoa(unescape(encodeURIComponent(xml)));
        dataUrl = 'data:image/svg+xml;base64,' + svg64;
      }
      await onSaveToDrive(`Barcode_${activeItem.sku}.png`, dataUrl);
      setDriveSavedMsg('Berhasil disimpan ke Google Drive Anda!');
      setTimeout(() => setDriveSavedMsg(null), 4000);
    } catch (err: any) {
      setDriveErrorMsg(`Gagal menyimpan ke Google Drive: ${err?.message || err}`);
      setTimeout(() => setDriveErrorMsg(null), 5000);
    } finally {
      setIsSavingToDrive(false);
    }
  };

  const printDate = new Date().toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  });

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto bg-slate-900/60 backdrop-blur-xs flex items-center justify-center p-3 sm:p-5">
      <div className="bg-white rounded-3xl max-w-3xl w-full p-6 sm:p-8 shadow-2xl border border-slate-100 relative animate-in fade-in zoom-in-95 duration-200">
        <button
          onClick={onClose}
          className="absolute top-5 right-5 p-2 text-slate-400 hover:text-slate-600 hover:bg-slate-100 rounded-xl transition-colors cursor-pointer"
        >
          <X className="w-5 h-5" />
        </button>

        {/* Modal Header */}
        <div className="flex items-center gap-3.5 mb-6">
          <div className="w-12 h-12 rounded-2xl bg-indigo-50 text-indigo-600 flex items-center justify-center shadow-xs">
            <Printer className="w-6 h-6" />
          </div>
          <div>
            <h3 className="text-xl font-bold text-slate-900 leading-snug">
              Generator & Cetak Label Barcode Kantor
            </h3>
            <p className="text-xs text-slate-500">
              Buat barcode 1D atau QR Code 2D untuk stiker rak gudang, aset kantor, atau inventaris
            </p>
          </div>
        </div>

        {/* Controls Grid */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
          {/* Select Mode / Item */}
          <div>
            <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1.5">
              Pilih Barang
            </label>
            <select
              value={selectedItemId}
              onChange={e => setSelectedItemId(e.target.value)}
              disabled={batchMode}
              className="w-full px-3 py-2 text-xs font-medium bg-slate-50 border border-slate-300 rounded-xl focus:ring-2 focus:ring-indigo-500"
            >
              {allItems.map(i => (
                <option key={i.id} value={i.id}>
                  [{i.sku}] {i.name}
                </option>
              ))}
            </select>
          </div>

          {/* Barcode Type */}
          <div>
            <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1.5">
              Format Kode
            </label>
            <div className="flex gap-2">
              <button
                type="button"
                onClick={() => setBarcodeType('CODE128')}
                className={`flex-1 py-2 px-3 text-xs font-bold rounded-xl border flex items-center justify-center gap-1.5 cursor-pointer ${
                  barcodeType === 'CODE128'
                    ? 'bg-indigo-600 text-white border-indigo-600'
                    : 'bg-slate-50 text-slate-700 border-slate-300 hover:bg-slate-100'
                }`}
              >
                <BarcodeIcon className="w-4 h-4" />
                <span>Code 128</span>
              </button>
              <button
                type="button"
                onClick={() => setBarcodeType('QR')}
                className={`flex-1 py-2 px-3 text-xs font-bold rounded-xl border flex items-center justify-center gap-1.5 cursor-pointer ${
                  barcodeType === 'QR'
                    ? 'bg-indigo-600 text-white border-indigo-600'
                    : 'bg-slate-50 text-slate-700 border-slate-300 hover:bg-slate-100'
                }`}
              >
                <QrCode className="w-4 h-4" />
                <span>QR Code</span>
              </button>
            </div>
          </div>

          {/* Label Template Size */}
          <div>
            <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1.5">
              Ukuran Stiker Label
            </label>
            <select
              value={labelSize}
              onChange={e => setLabelSize(e.target.value as any)}
              className="w-full px-3 py-2 text-xs font-medium bg-slate-50 border border-slate-300 rounded-xl focus:ring-2 focus:ring-indigo-500"
            >
              <option value="rack">Stiker Rak Standar (70 x 40 mm)</option>
              <option value="asset">Label Aset Kompak (50 x 25 mm)</option>
              <option value="compact">Label Kotak Mini</option>
            </select>
          </div>
        </div>

        {/* Company Name Customization */}
        <div className="mb-6 flex flex-col sm:flex-row items-center gap-3">
          <div className="flex-1 w-full">
            <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1">
              Nama Kantor / Perusahaan di Label
            </label>
            <div className="relative">
              <input
                type="text"
                value={companyName}
                onChange={e => setCompanyName(e.target.value)}
                placeholder="Contoh: PT INVENTARIS NUSANTARA"
                className="w-full pl-9 pr-3 py-2 text-xs font-semibold bg-slate-50 border border-slate-300 rounded-xl focus:ring-2 focus:ring-indigo-500"
              />
              <Building2 className="w-4 h-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
            </div>
          </div>

          {/* Toggle Batch Print */}
          <div className="w-full sm:w-auto pt-2 sm:pt-4">
            <button
              type="button"
              onClick={() => setBatchMode(!batchMode)}
              className={`w-full sm:w-auto px-4 py-2 text-xs font-bold rounded-xl border flex items-center justify-center gap-2 cursor-pointer transition-colors ${
                batchMode
                  ? 'bg-slate-900 text-white border-slate-900'
                  : 'bg-white text-slate-700 border-slate-300 hover:bg-slate-50'
              }`}
            >
              <Layers className="w-4 h-4" />
              <span>{batchMode ? 'Mode Batch: Semua Barang' : 'Cetak Lembar Batch A4'}</span>
            </button>
          </div>
        </div>

        {/* Live Preview Area (Ready for window.print()) */}
        <div className="bg-slate-100 p-6 rounded-2xl border border-slate-200 mb-6 flex flex-col items-center justify-center">
          <span className="text-[11px] font-bold uppercase tracking-wider text-slate-500 mb-3">
            Pratinjau Stiker Label Fisik
          </span>

          {!batchMode ? (
            /* Single Label Preview */
            <div
              id="printable-barcode-single"
              className={`bg-white border-2 border-slate-800 rounded-xl shadow-md p-4 text-center text-slate-900 flex flex-col items-center justify-between ${
                labelSize === 'asset' ? 'w-[260px] min-h-[140px]' : 'w-[320px] min-h-[170px]'
              }`}
            >
              {/* Header */}
              <div className="w-full border-b border-slate-800 pb-1 mb-2">
                <p className="text-[10px] font-black uppercase tracking-wider text-slate-800 truncate">
                  {companyName || 'INVENTARIS KANTOR'}
                </p>
                <p className="text-xs font-black truncate text-slate-900 mt-0.5">
                  {activeItem?.name}
                </p>
                {(activeItem?.merk || activeItem?.serialNumber) && (
                  <p className="text-[9px] font-mono text-slate-600 truncate mt-0.5">
                    {[activeItem?.merk, activeItem?.typeModel].filter(Boolean).join(' ')} {activeItem?.serialNumber ? `• SN: ${activeItem.serialNumber}` : ''}
                  </p>
                )}
              </div>

              {/* Barcode Graphic */}
              <div className="my-1 flex items-center justify-center">
                {barcodeType === 'CODE128' ? (
                  <svg ref={barcodeSvgRef} className="max-w-full" />
                ) : (
                  <canvas ref={qrCanvasRef} />
                )}
              </div>

              {/* Footer Details */}
              <div className="w-full border-t border-slate-800 pt-1 mt-1 flex items-center justify-between text-[9px] text-slate-600 font-bold">
                <span className="truncate max-w-[150px]">{activeItem?.location}</span>
                <span>Tgl Cetak: {printDate}</span>
              </div>
            </div>
          ) : (
            /* Batch Grid Preview (Lembar Cetak Banyak) */
            <div className="w-full max-h-64 overflow-y-auto bg-white p-4 rounded-xl border border-slate-300">
              <p className="text-xs font-bold text-slate-700 mb-3">
                Format Lembar A4 ({allItems.length} label stiker siap cetak):
              </p>
              <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
                {allItems.map(it => (
                  <div
                    key={it.id}
                    className="border border-slate-800 rounded-lg p-2 text-center text-slate-900 bg-white"
                  >
                    <p className="text-[9px] font-black uppercase tracking-wide truncate">{companyName}</p>
                    <p className="text-[10px] font-bold truncate mt-0.5">{it.name}</p>
                    <div className="my-1 py-1 font-mono font-bold text-xs bg-slate-50 border border-dashed border-slate-300 rounded">
                      *{it.sku}*
                    </div>
                    <p className="text-[8px] text-slate-500 truncate">{it.location}</p>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        {driveSavedMsg && (
          <div className="mb-4 p-3 rounded-xl bg-emerald-50 border border-emerald-200 text-emerald-800 text-xs flex items-center gap-2">
            <Check className="w-4 h-4 text-emerald-600 shrink-0" />
            <span>{driveSavedMsg}</span>
          </div>
        )}

        {driveErrorMsg && (
          <div className="mb-4 p-3 rounded-xl bg-rose-50 border border-rose-200 text-rose-800 text-xs flex items-center gap-2">
            <span className="w-2 h-2 rounded-full bg-rose-500 shrink-0" />
            <span>{driveErrorMsg}</span>
          </div>
        )}

        {/* Footer Actions */}
        <div className="flex flex-wrap items-center justify-between gap-3 pt-3 border-t border-slate-100">
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={handleDownloadPNG}
              disabled={batchMode}
              className="inline-flex items-center gap-1.5 px-3.5 py-2 text-xs font-semibold text-slate-700 bg-slate-100 hover:bg-slate-200 rounded-xl transition-colors cursor-pointer disabled:opacity-50"
              title="Unduh PNG"
            >
              <Download className="w-4 h-4" />
              <span>Unduh PNG</span>
            </button>

            {isDriveConnected && onSaveToDrive && (
              <button
                type="button"
                onClick={handleSaveToGoogleDrive}
                disabled={isSavingToDrive || batchMode}
                className="inline-flex items-center gap-1.5 px-3.5 py-2 text-xs font-semibold text-emerald-700 bg-emerald-50 hover:bg-emerald-100 border border-emerald-200 rounded-xl transition-colors cursor-pointer disabled:opacity-50"
                title="Simpan Barcode ke Google Drive"
              >
                {isSavingToDrive ? <RefreshCw className="w-4 h-4 animate-spin" /> : <FileSpreadsheet className="w-4 h-4" />}
                <span>{isSavingToDrive ? 'Menyimpan...' : 'Simpan ke Drive'}</span>
              </button>
            )}
          </div>

          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-xs font-medium text-slate-700 hover:bg-slate-100 rounded-xl transition-colors cursor-pointer"
            >
              Tutup
            </button>
            <button
              type="button"
              onClick={handlePrint}
              className="inline-flex items-center gap-2 px-5 py-2.5 bg-indigo-600 hover:bg-indigo-700 text-white font-semibold text-xs rounded-xl shadow-xs transition-colors cursor-pointer"
            >
              <Printer className="w-4 h-4" />
              <span>{batchMode ? `Cetak Lembar Batch (${allItems.length} Label)` : 'Cetak Label Ini'}</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
