import React, { useState } from 'react';
import { 
  X, 
  FileText, 
  Download, 
  Building2, 
  Calendar, 
  UserCheck, 
  Check, 
  Layers,
  Sparkles
} from 'lucide-react';
import { InventoryItem, StockTransaction } from '../types/inventory';
import { generateInventoryPDF } from '../utils/pdfReport';

interface PDFReportModalProps {
  isOpen: boolean;
  items: InventoryItem[];
  transactions: StockTransaction[];
  onClose: () => void;
  onSuccess: (filename: string) => void;
}

export const PDFReportModal: React.FC<PDFReportModalProps> = ({
  isOpen,
  items,
  transactions,
  onClose,
  onSuccess,
}) => {
  const currentMonthYear = new Date().toLocaleDateString('id-ID', { month: 'long', year: 'numeric' });
  const [companyName, setCompanyName] = useState<string>('PT INVENTARIS NUSANTARA');
  const [monthYear, setMonthYear] = useState<string>(currentMonthYear);
  const [preparedBy, setPreparedBy] = useState<string>('Ahmad Faisal (Staff GA & Logistik)');
  const [approvedBy, setApprovedBy] = useState<string>('Budi Santoso, S.E. (Head of Finance & Operations)');
  const [isGenerating, setIsGenerating] = useState<boolean>(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  if (!isOpen) return null;

  const handleDownload = () => {
    setIsGenerating(true);
    setErrorMessage(null);
    try {
      const filename = generateInventoryPDF(items, transactions, {
        companyName: companyName.trim() || 'INVENTARIS KANTOR',
        monthYear: monthYear.trim() || currentMonthYear,
        preparedBy: preparedBy.trim() || 'Staff Inventaris',
        approvedBy: approvedBy.trim() || 'Pimpinan Kantor',
      });
      onSuccess(filename);
      onClose();
    } catch (err: any) {
      console.error('PDF Generation error:', err);
      setErrorMessage('Gagal membuat file PDF: ' + (err?.message || err));
    } finally {
      setIsGenerating(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto bg-slate-900/60 backdrop-blur-xs flex items-center justify-center p-3 sm:p-5">
      <div className="bg-white rounded-3xl max-w-lg w-full p-6 sm:p-8 shadow-2xl border border-slate-100 relative animate-in fade-in zoom-in-95 duration-200">
        <button
          onClick={onClose}
          className="absolute top-5 right-5 p-2 text-slate-400 hover:text-slate-600 hover:bg-slate-100 rounded-xl transition-colors cursor-pointer"
        >
          <X className="w-5 h-5" />
        </button>

        {/* Modal Header */}
        <div className="flex items-center gap-3.5 mb-5">
          <div className="w-12 h-12 rounded-2xl bg-indigo-50 text-indigo-600 flex items-center justify-center shadow-xs">
            <FileText className="w-6 h-6" />
          </div>
          <div>
            <h3 className="text-xl font-bold text-slate-900 leading-snug">
              Cetak Rekap Laporan Bulanan (PDF)
            </h3>
            <p className="text-xs text-slate-500">
              Dokumen resmi siap cetak & tanda tangan pimpinan kantor tanpa perlu buka Google Sheets
            </p>
          </div>
        </div>

        {errorMessage && (
          <div className="mb-4 p-3 rounded-xl bg-rose-50 border border-rose-200 text-rose-700 text-xs font-semibold flex items-center gap-2">
            <span className="w-2 h-2 rounded-full bg-rose-500 shrink-0" />
            <span>{errorMessage}</span>
          </div>
        )}

        {/* Form Fields */}
        <div className="space-y-4">
          {/* Company Name */}
          <div>
            <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1">
              Nama Instansi / Perusahaan di Kop Surat
            </label>
            <div className="relative">
              <input
                type="text"
                value={companyName}
                onChange={e => setCompanyName(e.target.value)}
                placeholder="Contoh: PT INVENTARIS NUSANTARA"
                className="w-full pl-9 pr-3.5 py-2.5 text-xs font-semibold bg-slate-50 border border-slate-300 rounded-xl focus:ring-2 focus:ring-indigo-500"
              />
              <Building2 className="w-4 h-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
            </div>
          </div>

          {/* Month / Period */}
          <div>
            <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1">
              Periode Bulan Rekap
            </label>
            <div className="relative">
              <input
                type="text"
                value={monthYear}
                onChange={e => setMonthYear(e.target.value)}
                placeholder="Contoh: Oktober 2026"
                className="w-full pl-9 pr-3.5 py-2.5 text-xs font-semibold bg-slate-50 border border-slate-300 rounded-xl focus:ring-2 focus:ring-indigo-500"
              />
              <Calendar className="w-4 h-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
            </div>
          </div>

          {/* Prepared By & Approved By */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div>
              <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1">
                Petugas Pembuat (GA/Gudang)
              </label>
              <input
                type="text"
                value={preparedBy}
                onChange={e => setPreparedBy(e.target.value)}
                placeholder="Nama staf pembuat"
                className="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-300 rounded-xl focus:ring-2 focus:ring-indigo-500 font-medium"
              />
            </div>

            <div>
              <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1">
                Pimpinan yang Menyetujui
              </label>
              <input
                type="text"
                value={approvedBy}
                onChange={e => setApprovedBy(e.target.value)}
                placeholder="Nama kepala bagian"
                className="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-300 rounded-xl focus:ring-2 focus:ring-indigo-500 font-medium"
              />
            </div>
          </div>

          {/* Included Content Summary Box */}
          <div className="p-3.5 rounded-xl bg-slate-50 border border-slate-200 text-xs space-y-1.5">
            <span className="font-bold text-slate-800 flex items-center gap-1.5">
              <Layers className="w-3.5 h-3.5 text-indigo-600" />
              <span>Isi Laporan PDF yang Dihasilkan:</span>
            </span>
            <ul className="text-slate-600 text-[11px] space-y-1 pl-4 list-disc">
              <li>Kop surat formal & ringkasan eksekutif (total unit & nilai aset inventaris kantor)</li>
              <li>Tabel master stok <strong>{items.length} barang</strong> dengan lokasi rak & status stok</li>
              <li>Tabel riwayat mutasi barang masuk & keluar terkini</li>
              <li>Kolom tanda tangan resmi pejabat berwenang</li>
            </ul>
          </div>
        </div>

        {/* Modal Actions */}
        <div className="flex items-center justify-end gap-3 pt-5 mt-4 border-t border-slate-100">
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2 text-xs font-semibold text-slate-700 hover:bg-slate-100 rounded-xl transition-colors cursor-pointer"
          >
            Batal
          </button>
          <button
            type="button"
            onClick={handleDownload}
            disabled={isGenerating}
            className="inline-flex items-center gap-2 px-5 py-2.5 bg-indigo-600 hover:bg-indigo-700 text-white font-semibold text-xs rounded-xl shadow-xs transition-colors cursor-pointer disabled:opacity-50"
          >
            <Download className="w-4 h-4" />
            <span>{isGenerating ? 'Menyiapkan Dokumen...' : 'Unduh Laporan PDF'}</span>
          </button>
        </div>
      </div>
    </div>
  );
};
