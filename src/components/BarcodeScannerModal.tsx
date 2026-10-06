import React, { useState, useEffect, useRef } from 'react';
import { 
  X, 
  ScanLine, 
  Camera, 
  Keyboard, 
  ArrowDownLeft, 
  ArrowUpRight, 
  Plus, 
  Package, 
  CheckCircle2, 
  AlertCircle,
  RefreshCw
} from 'lucide-react';
import { InventoryItem } from '../types/inventory';
import { playScanBeep } from '../utils/audio';

interface BarcodeScannerModalProps {
  isOpen: boolean;
  items: InventoryItem[];
  onClose: () => void;
  onSelectAction: (type: 'VIEW' | 'PRINT_BARCODE' | 'NEW_WITH_SKU', item?: InventoryItem, scannedCode?: string) => void;
}

export const BarcodeScannerModal: React.FC<BarcodeScannerModalProps> = ({
  isOpen,
  items,
  onClose,
  onSelectAction,
}) => {
  const [activeMode, setActiveMode] = useState<'camera' | 'manual'>('camera');
  const [scannedResult, setScannedResult] = useState<string | null>(null);
  const [matchedItem, setMatchedItem] = useState<InventoryItem | null>(null);
  const [manualCode, setManualCode] = useState<string>('');
  const [cameraError, setCameraError] = useState<string | null>(null);
  const [isScanning, setIsScanning] = useState<boolean>(false);

  const scannerRef = useRef<any>(null);
  const html5QrCodeId = 'reader-barcode-scanner';

  // Handle scanned string
  const handleBarcodeDecoded = (decodedText: string) => {
    let cleanCode = decodedText.trim();
    // In case QR contains JSON object
    try {
      if (cleanCode.startsWith('{') && cleanCode.endsWith('}')) {
        const parsed = JSON.parse(cleanCode);
        if (parsed.sku) cleanCode = parsed.sku;
      }
    } catch {
      // not json, keep string
    }

    playScanBeep();
    setScannedResult(cleanCode);

    // Look for matching item
    const found = items.find(
      i => i.sku.toLowerCase() === cleanCode.toLowerCase() ||
           i.name.toLowerCase() === cleanCode.toLowerCase()
    );
    setMatchedItem(found || null);
  };

  // Start Camera Scanner
  useEffect(() => {
    if (!isOpen || activeMode !== 'camera') {
      stopCameraScanner();
      return;
    }

    let isMounted = true;
    setCameraError(null);

    const initScanner = async () => {
      try {
        // Small delay to ensure DOM element is ready
        await new Promise(r => setTimeout(r, 200));
        if (!isMounted) return;

        const { Html5Qrcode } = await import('html5-qrcode');
        if (!isMounted) return;

        const scanner = new Html5Qrcode(html5QrCodeId);
        scannerRef.current = scanner;

        await scanner.start(
          { facingMode: 'environment' },
          {
            fps: 10,
            qrbox: { width: 250, height: 250 },
            aspectRatio: 1.0,
          },
          (decodedText) => {
            handleBarcodeDecoded(decodedText);
          },
          () => {
            // frame by frame scan callback
          }
        );
        if (isMounted) setIsScanning(true);
      } catch (err: any) {
        console.warn('Camera scanner start error:', err);
        if (isMounted) {
          setCameraError(
            'Kamera tidak dapat diakses atau izin belum diberikan. Anda tetap bisa menggunakan pemindai barcode manual/gun.'
          );
          setActiveMode('manual');
        }
      }
    };

    initScanner();

    return () => {
      isMounted = false;
      stopCameraScanner();
    };
  }, [isOpen, activeMode]);

  const stopCameraScanner = async () => {
    if (scannerRef.current) {
      try {
        if (scannerRef.current.isScanning) {
          await scannerRef.current.stop();
        }
        scannerRef.current.clear();
      } catch (e) {
        // ignore
      }
      scannerRef.current = null;
      setIsScanning(false);
    }
  };

  const handleClose = () => {
    stopCameraScanner();
    setScannedResult(null);
    setMatchedItem(null);
    setManualCode('');
    onClose();
  };

  const handleManualSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!manualCode.trim()) return;
    handleBarcodeDecoded(manualCode);
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto bg-slate-900/60 backdrop-blur-xs flex items-center justify-center p-3 sm:p-5">
      <div className="bg-white rounded-3xl max-w-lg w-full p-6 sm:p-8 shadow-2xl border border-slate-100 relative animate-in fade-in zoom-in-95 duration-200">
        <button
          onClick={handleClose}
          className="absolute top-5 right-5 p-2 text-slate-400 hover:text-slate-600 hover:bg-slate-100 rounded-xl transition-colors cursor-pointer"
        >
          <X className="w-5 h-5" />
        </button>

        {/* Modal Header */}
        <div className="flex items-center gap-3.5 mb-5">
          <div className="w-12 h-12 rounded-2xl bg-indigo-50 text-indigo-600 flex items-center justify-center shadow-xs">
            <ScanLine className="w-6 h-6 animate-pulse" />
          </div>
          <div>
            <h3 className="text-xl font-bold text-slate-900 leading-snug">
              Pemindai Barcode & QR Real-Time
            </h3>
            <p className="text-xs text-slate-500">
              Scan barcode langsung dari kamera HP/laptop atau input cepat barcode gun
            </p>
          </div>
        </div>

        {/* Mode Switcher Tabs */}
        <div className="flex gap-2 p-1 bg-slate-100 rounded-xl mb-4 text-xs font-bold">
          <button
            type="button"
            onClick={() => setActiveMode('camera')}
            className={`flex-1 py-2 px-3 rounded-lg flex items-center justify-center gap-1.5 transition-colors cursor-pointer ${
              activeMode === 'camera'
                ? 'bg-white text-indigo-700 shadow-xs'
                : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            <Camera className="w-4 h-4" />
            <span>Kamera Live</span>
          </button>
          <button
            type="button"
            onClick={() => setActiveMode('manual')}
            className={`flex-1 py-2 px-3 rounded-lg flex items-center justify-center gap-1.5 transition-colors cursor-pointer ${
              activeMode === 'manual'
                ? 'bg-white text-indigo-700 shadow-xs'
                : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            <Keyboard className="w-4 h-4" />
            <span>Barcode Gun / Input Cepat</span>
          </button>
        </div>

        {/* Scanner View Area */}
        {activeMode === 'camera' ? (
          <div className="space-y-3">
            <div className="relative rounded-2xl overflow-hidden bg-slate-950 border border-slate-800 aspect-square max-h-64 flex items-center justify-center">
              <div id={html5QrCodeId} className="w-full h-full" />
              {/* Scan target visual overlay */}
              <div className="absolute inset-0 pointer-events-none border-2 border-indigo-400/40 rounded-2xl flex items-center justify-center">
                <div className="w-48 h-48 border-2 border-dashed border-indigo-400 rounded-xl animate-pulse" />
              </div>
            </div>
            <p className="text-center text-[11px] text-slate-500">
              Arahkan kamera ke barcode 1D atau QR Code pada stiker barang kantor
            </p>
          </div>
        ) : (
          <form onSubmit={handleManualSubmit} className="space-y-3">
            <div>
              <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1">
                Scan dengan Barcode Gun USB atau Ketik SKU
              </label>
              <div className="flex gap-2">
                <input
                  type="text"
                  autoFocus
                  value={manualCode}
                  onChange={e => setManualCode(e.target.value.toUpperCase())}
                  placeholder="Contoh: ATK-2024-001"
                  className="flex-1 px-3.5 py-2.5 text-sm font-mono font-bold bg-slate-50 border border-slate-300 rounded-xl focus:ring-2 focus:ring-indigo-500"
                />
                <button
                  type="submit"
                  className="px-4 py-2.5 bg-indigo-600 hover:bg-indigo-700 text-white font-semibold text-xs rounded-xl shadow-xs transition-colors cursor-pointer"
                >
                  Cari
                </button>
              </div>
              <p className="text-[11px] text-slate-400 mt-1">
                Tip: Alat scanner barcode USB otomatis menekan Enter setelah scan
              </p>
            </div>
          </form>
        )}

        {cameraError && (
          <div className="mt-3 p-3 rounded-xl bg-amber-50 border border-amber-200 text-amber-900 text-xs flex items-center gap-2">
            <AlertCircle className="w-4 h-4 shrink-0" />
            <span>{cameraError}</span>
          </div>
        )}

        {/* Scan Result Card */}
        {scannedResult && (
          <div className="mt-5 p-4 rounded-2xl border bg-slate-50 border-slate-200 animate-in fade-in slide-in-from-bottom-2 duration-200">
            <div className="flex items-center justify-between mb-2">
              <span className="text-[11px] font-bold uppercase tracking-wider text-slate-500">
                Hasil Pemindaian Barcode
              </span>
              <span className="font-mono font-bold text-xs bg-indigo-100 text-indigo-800 px-2 py-0.5 rounded-md">
                {scannedResult}
              </span>
            </div>

            {matchedItem ? (
              <div className="space-y-3">
                <div className="bg-white p-3.5 rounded-xl border border-slate-200 flex items-start gap-3">
                  <div className="w-10 h-10 rounded-xl bg-emerald-50 text-emerald-600 flex items-center justify-center shrink-0">
                    <Package className="w-5 h-5" />
                  </div>
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-2">
                      <span className="text-[10px] px-2 py-0.5 rounded bg-slate-100 text-slate-700 font-semibold">
                        {matchedItem.category}
                      </span>
                      <span className="text-xs text-slate-500 truncate">{matchedItem.location}</span>
                    </div>
                    <p className="text-sm font-bold text-slate-900 mt-1 truncate">{matchedItem.name}</p>
                    <p className="text-xs font-semibold text-slate-700 mt-0.5">
                      Stok Saat Ini:{' '}
                      <span className="font-extrabold text-indigo-600 text-sm">
                        {matchedItem.currentStock} {matchedItem.unit}
                      </span>{' '}
                      (Batas Min: {matchedItem.minStock})
                    </p>
                  </div>
                </div>

                {/* Instant Actions for matched item */}
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 pt-1">
                  <button
                    type="button"
                    onClick={() => {
                      handleClose();
                      onSelectAction('VIEW', matchedItem);
                    }}
                    className="p-2.5 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl font-bold text-xs flex items-center justify-center gap-1.5 transition-colors cursor-pointer shadow-xs"
                  >
                    <Package className="w-3.5 h-3.5" />
                    <span>Lihat & Edit Data Barang</span>
                  </button>
                  <button
                    type="button"
                    onClick={() => {
                      handleClose();
                      onSelectAction('PRINT_BARCODE', matchedItem);
                    }}
                    className="p-2.5 bg-slate-800 hover:bg-slate-900 text-white rounded-xl font-bold text-xs flex items-center justify-center gap-1.5 transition-colors cursor-pointer shadow-xs"
                  >
                    <ScanLine className="w-3.5 h-3.5" />
                    <span>Cetak Stiker Barcode</span>
                  </button>
                </div>
              </div>
            ) : (
              <div className="space-y-3">
                <div className="bg-amber-50 p-3.5 rounded-xl border border-amber-200 text-xs text-amber-900">
                  <p className="font-bold">Barang Belum Terdaftar di Sistem</p>
                  <p className="mt-0.5 text-amber-800">
                    Kode barcode "{scannedResult}" belum ada dalam database inventaris kantor.
                  </p>
                </div>

                <button
                  type="button"
                  onClick={() => {
                    handleClose();
                    onSelectAction('NEW_WITH_SKU', undefined, scannedResult);
                  }}
                  className="w-full p-2.5 bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl font-bold text-xs flex items-center justify-center gap-1.5 transition-colors cursor-pointer shadow-xs"
                >
                  <Plus className="w-4 h-4" />
                  <span>Daftarkan Sebagai Barang Baru Masuk (+)</span>
                </button>
              </div>
            )}
          </div>
        )}

        {/* Modal Footer */}
        <div className="mt-6 pt-3 border-t border-slate-100 flex items-center justify-between">
          <span className="text-[11px] text-slate-400 flex items-center gap-1">
            <CheckCircle2 className="w-3.5 h-3.5 text-emerald-500" />
            <span>Deteksi Otomatis & Bip Audio Aktif</span>
          </span>
          <button
            type="button"
            onClick={handleClose}
            className="px-4 py-1.5 text-xs font-semibold text-slate-700 hover:bg-slate-100 rounded-xl transition-colors cursor-pointer"
          >
            Tutup
          </button>
        </div>
      </div>
    </div>
  );
};
