import React, { useState, useMemo } from 'react';
import { 
  ArrowDownLeft, 
  Search, 
  Download, 
  FileSpreadsheet, 
  Calendar, 
  FileText,
  Image as ImageIcon,
  CheckCircle2,
  ExternalLink,
  Layers,
  MapPin,
  Navigation,
  HardDrive,
  Clock,
  Building2,
  X,
  Sparkles
} from 'lucide-react';
import { StockTransaction } from '../types/inventory';

interface TransactionHistoryViewProps {
  transactions: StockTransaction[];
  onSyncToGoogleSheets: () => void;
  isSyncing: boolean;
  spreadsheetUrl: string | null;
  onOpenPDFReport: () => void;
}

export const TransactionHistoryView: React.FC<TransactionHistoryViewProps> = ({
  transactions,
  onSyncToGoogleSheets,
  isSyncing,
  spreadsheetUrl,
  onOpenPDFReport,
}) => {
  const [searchQuery, setSearchQuery] = useState('');
  const [previewPhoto, setPreviewPhoto] = useState<{ url: string; title: string; geoTag?: any; driveLink?: string } | null>(null);

  const filteredTransactions = useMemo(() => {
    return transactions.filter(t => {
      const q = searchQuery.toLowerCase();
      const matchQuery =
        t.itemName.toLowerCase().includes(q) ||
        t.itemSku.toLowerCase().includes(q) ||
        (t.supplierOrSource && t.supplierOrSource.toLowerCase().includes(q)) ||
        (t.receivedBy && t.receivedBy.toLowerCase().includes(q)) ||
        (t.invoiceOrPoNumber && t.invoiceOrPoNumber.toLowerCase().includes(q)) ||
        (t.serialNumber && t.serialNumber.toLowerCase().includes(q)) ||
        (t.merk && t.merk.toLowerCase().includes(q)) ||
        (t.typeModel && t.typeModel.toLowerCase().includes(q)) ||
        (t.notes && t.notes.toLowerCase().includes(q));

      return matchQuery;
    });
  }, [transactions, searchQuery]);

  const totalInQty = transactions.reduce((acc, c) => acc + c.quantity, 0);

  const handleExportCSV = () => {
    const headers = [
      'No',
      'ID Transaksi',
      'Tanggal & Waktu',
      'No PO / Surat Jalan',
      'Kode Barcode / SKU',
      'Nama Barang',
      'Merk',
      'Type / Model',
      'No Seri',
      'Jumlah Masuk (+)',
      'Satuan',
      'Stok Sebelum',
      'Stok Sesudah',
      'Pemasok / Distributor',
      'Petugas Penerima (GA)',
      'AKL / AKD',
      'Catatan / Keterangan',
    ];

    const rows = filteredTransactions.map((t, idx) => [
      (idx + 1).toString(),
      t.id,
      new Date(t.timestamp).toLocaleString('id-ID'),
      `"${t.invoiceOrPoNumber || '-'}"`,
      t.itemSku,
      `"${t.itemName.replace(/"/g, '""')}"`,
      `"${t.merk || '-'}"`,
      `"${t.typeModel || '-'}"`,
      `"${t.serialNumber || '-'}"`,
      t.quantity,
      t.unit,
      t.previousStock,
      t.newStock,
      `"${(t.distributor || t.supplierOrSource || '-').replace(/"/g, '""')}"`,
      `"${(t.receivedBy || '-').replace(/"/g, '""')}"`,
      `"${t.aklAkd || '-'}"`,
      `"${(t.notes || '-').replace(/"/g, '""')}"`,
    ]);

    const csvContent = 'data:text/csv;charset=utf-8,\uFEFF' + [headers.join(','), ...rows.map(r => r.join(','))].join('\n');
    const encodedUri = encodeURI(csvContent);
    const link = document.createElement('a');
    link.setAttribute('href', encodedUri);
    link.setAttribute('download', `Log_Penerimaan_Barang_Masuk_${new Date().toISOString().slice(0, 10)}.csv`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  return (
    <div className="space-y-4">
      {/* Header and Controls */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-white p-5 sm:p-6 rounded-3xl border border-slate-200/90 shadow-xs">
        <div>
          <div className="flex items-center gap-2.5 flex-wrap">
            <h2 className="text-xl sm:text-2xl font-black text-slate-900 tracking-tight">
              Riwayat Log Penerimaan Barang Masuk (+)
            </h2>
            <span className="px-2.5 py-0.5 rounded-full text-xs font-bold bg-emerald-50 text-emerald-800 border border-emerald-200">
              {transactions.length} Mutasi Tercatat
            </span>
          </div>
          <p className="text-xs text-slate-500 mt-1">
            Audit trail penerimaan barang baru masuk dari rekanan/pengadaan lengkap dengan bukti foto fisik & geotagging GPS
          </p>
        </div>

        <div className="flex items-center gap-2 flex-wrap">
          <button
            onClick={onOpenPDFReport}
            className="inline-flex items-center gap-1.5 px-3.5 py-2 bg-rose-50 hover:bg-rose-100 text-rose-700 border border-rose-200 font-bold text-xs rounded-xl transition-colors cursor-pointer"
            title="Unduh Laporan Audit PDF"
          >
            <FileText className="w-3.5 h-3.5 text-rose-600" />
            <span>Unduh PDF</span>
          </button>

          <button
            onClick={handleExportCSV}
            className="inline-flex items-center gap-1.5 px-3.5 py-2 bg-slate-100 hover:bg-slate-200 text-slate-700 font-bold text-xs rounded-xl transition-colors cursor-pointer border border-slate-200"
            title="Ekspor CSV Excel"
          >
            <Download className="w-3.5 h-3.5" />
            <span>Ekspor CSV</span>
          </button>

          <button
            onClick={onSyncToGoogleSheets}
            disabled={isSyncing}
            className="inline-flex items-center gap-1.5 px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-white font-bold text-xs rounded-xl shadow-md shadow-emerald-600/30 transition-all cursor-pointer disabled:opacity-50 ring-1 ring-emerald-400/40"
          >
            <FileSpreadsheet className="w-3.5 h-3.5" />
            <span>{isSyncing ? 'Sinkronisasi...' : 'Sync ke Google Sheets'}</span>
          </button>

          {spreadsheetUrl && (
            <a
              href={spreadsheetUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-1.5 px-3 py-2 text-xs font-bold text-emerald-800 bg-emerald-50 hover:bg-emerald-100 border border-emerald-200 rounded-xl transition-colors"
            >
              <span>Buka Sheets</span>
              <ExternalLink className="w-3.5 h-3.5" />
            </a>
          )}
        </div>
      </div>

      {/* Metrics Summary Strip */}
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <div className="bg-white p-4 rounded-2xl border border-slate-200/90 shadow-2xs flex items-center justify-between">
          <span className="text-xs font-bold text-slate-500">Total Transaksi Penerimaan</span>
          <span className="text-xl font-black text-slate-900">{transactions.length} Dokumen PO / SJ</span>
        </div>
        <div className="bg-white p-4 rounded-2xl border border-emerald-200 shadow-2xs flex items-center justify-between bg-emerald-50/20">
          <span className="text-xs font-bold text-emerald-800">Total Akumulasi Barang Masuk</span>
          <span className="text-xl font-black text-emerald-600">+{totalInQty.toLocaleString('id-ID')} Unit</span>
        </div>
      </div>

      {/* Search Bar */}
      <div className="bg-white p-4 rounded-2xl border border-slate-200/90 shadow-xs">
        <div className="relative">
          <Search className="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
          <input
            type="text"
            placeholder="Cari berdasarkan nama barang, SKU, vendor/pemasok, nomor PO, no seri, petugas penerima..."
            value={searchQuery}
            onChange={e => setSearchQuery(e.target.value)}
            className="w-full pl-10 pr-9 py-2 text-xs sm:text-sm bg-slate-50 border border-slate-200 rounded-xl focus:outline-hidden focus:ring-2 focus:ring-indigo-500 font-medium"
          />
          {searchQuery && (
            <button
              onClick={() => setSearchQuery('')}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600"
            >
              <X className="w-3.5 h-3.5" />
            </button>
          )}
        </div>
      </div>

      {/* Transactions Table */}
      <div className="bg-white rounded-3xl border border-slate-200/90 shadow-xs overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse text-xs">
            <thead>
              <tr className="bg-slate-900 text-white font-bold uppercase tracking-wider text-[11px] border-b border-slate-800">
                <th className="py-3 px-3 text-center">Foto & GPS</th>
                <th className="py-3 px-4">Waktu & ID Transaksi</th>
                <th className="py-3 px-3">No. Dokumen / PO</th>
                <th className="py-3 px-4">Barang (SKU)</th>
                <th className="py-3 px-3">Merk / Type</th>
                <th className="py-3 px-3">No Seri</th>
                <th className="py-3 px-4 text-center bg-emerald-900 text-white">Jumlah Masuk</th>
                <th className="py-3 px-4">Stok Mutasi</th>
                <th className="py-3 px-4">Pemasok & Petugas</th>
                <th className="py-3 px-3">Status Cloud</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100 text-[11px]">
              {filteredTransactions.length === 0 ? (
                <tr>
                  <td colSpan={10} className="py-14 text-center text-slate-400">
                    <ArrowDownLeft className="w-8 h-8 mx-auto text-slate-300 mb-2" />
                    <p className="font-bold text-slate-700 text-sm">Tidak ada riwayat barang masuk</p>
                    <p className="text-xs text-slate-400 mt-1">Belum ada mutasi yang sesuai dengan kata kunci pencarian</p>
                  </td>
                </tr>
              ) : (
                filteredTransactions.map(tx => {
                  const dateObj = new Date(tx.timestamp);
                  const dateStr = dateObj.toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' });
                  const timeStr = dateObj.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' });

                  return (
                    <tr key={tx.id} className="hover:bg-slate-50/80 transition-colors">
                      {/* Photo Thumbnail */}
                      <td className="py-3 px-3 text-center">
                        {tx.photoUrl ? (
                          <div 
                            onClick={() => setPreviewPhoto({ 
                              url: tx.photoUrl!, 
                              title: `Bukti Penerimaan: ${tx.itemName} (${tx.id})`,
                              geoTag: tx.geoTag,
                              driveLink: tx.driveFileLink
                            })}
                            className="relative w-12 h-12 rounded-xl overflow-hidden border border-emerald-300 cursor-pointer hover:scale-105 transition-transform bg-slate-50 mx-auto shadow-2xs group"
                          >
                            <img src={tx.photoUrl} alt="Foto Masuk" className="w-full h-full object-cover" />
                            {tx.geoTag && (
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

                      {/* Date & ID */}
                      <td className="py-3 px-4 whitespace-nowrap">
                        <div className="font-bold text-slate-900">{dateStr}</div>
                        <div className="text-[10px] text-slate-400 font-mono flex items-center gap-1 mt-0.5">
                          <span>{timeStr}</span>
                          <span>•</span>
                          <span className="text-indigo-600 font-bold">{tx.id}</span>
                        </div>
                      </td>

                      {/* No. Dokumen PO */}
                      <td className="py-3 px-3 whitespace-nowrap">
                        <span className="font-mono text-[10px] font-bold text-emerald-800 bg-emerald-50 px-2 py-0.5 rounded-md border border-emerald-200">
                          {tx.invoiceOrPoNumber || '-'}
                        </span>
                      </td>

                      {/* Item & SKU */}
                      <td className="py-3 px-4 min-w-[170px]">
                        <div className="font-extrabold text-slate-900">{tx.itemName}</div>
                        <div className="font-mono text-[10px] text-indigo-700 font-bold mt-0.5">{tx.itemSku}</div>
                      </td>

                      {/* Merk / Type */}
                      <td className="py-3 px-3 whitespace-nowrap font-medium text-slate-700">
                        <div>{tx.merk || '-'}</div>
                        <div className="text-[10px] text-slate-400">{tx.typeModel || ''}</div>
                      </td>

                      {/* No Seri */}
                      <td className="py-3 px-3 whitespace-nowrap font-mono font-bold text-slate-800 text-[10px]">
                        {tx.serialNumber || '-'}
                      </td>

                      {/* Quantity */}
                      <td className="py-3 px-4 text-center whitespace-nowrap">
                        <span className="text-sm font-black text-emerald-600 bg-emerald-50 px-2 py-0.5 rounded-lg border border-emerald-200">
                          +{tx.quantity} {tx.unit}
                        </span>
                      </td>

                      {/* Stock Change (Previous -> New) */}
                      <td className="py-3 px-4 whitespace-nowrap text-slate-600">
                        <span className="font-mono text-slate-500">{tx.previousStock}</span>
                        <span className="mx-1 text-slate-300">→</span>
                        <span className="font-mono font-bold text-emerald-700">{tx.newStock} {tx.unit}</span>
                      </td>

                      {/* Vendor & Receiver */}
                      <td className="py-3 px-4">
                        <div className="font-bold text-slate-800 truncate max-w-[140px]">
                          {tx.distributor || tx.supplierOrSource || '-'}
                        </div>
                        <div className="text-[10px] text-slate-400">
                          Penerima: <strong className="text-slate-600">{tx.receivedBy || '-'}</strong>
                        </div>
                      </td>

                      {/* Cloud Sync Status */}
                      <td className="py-3 px-3 whitespace-nowrap">
                        <div className="flex items-center gap-1.5 text-[10px] text-emerald-700 font-bold bg-emerald-50/80 px-2 py-1 rounded-lg border border-emerald-200">
                          <CheckCircle2 className="w-3.5 h-3.5 text-emerald-600 shrink-0" />
                          <span>Tersinkron</span>
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

      {/* Photo Preview Lightbox Modal */}
      {previewPhoto && (
        <div 
          onClick={() => setPreviewPhoto(null)}
          className="fixed inset-0 z-50 bg-slate-950/80 backdrop-blur-md flex items-center justify-center p-4 animate-in fade-in duration-200"
        >
          <div className="bg-white rounded-3xl p-5 max-w-lg w-full overflow-hidden shadow-2xl relative" onClick={e => e.stopPropagation()}>
            <div className="flex items-center justify-between pb-3 border-b border-slate-100">
              <h4 className="font-bold text-slate-900 text-sm truncate">{previewPhoto.title}</h4>
              <button 
                onClick={() => setPreviewPhoto(null)}
                className="p-1 text-slate-400 hover:text-slate-700 rounded-lg cursor-pointer"
              >
                ✕
              </button>
            </div>
            
            <div className="mt-3 rounded-2xl overflow-hidden bg-slate-900 max-h-[60vh] flex items-center justify-center">
              <img src={previewPhoto.url} alt={previewPhoto.title} className="max-w-full max-h-[60vh] object-contain" />
            </div>

            {/* Geotagging & Google Drive Links */}
            <div className="mt-4 pt-3 border-t border-slate-100 flex flex-wrap items-center justify-between gap-2">
              {previewPhoto.geoTag ? (
                <div className="flex items-center gap-2">
                  <span className="px-2.5 py-1 rounded-lg bg-emerald-100 text-emerald-800 text-[11px] font-bold flex items-center gap-1.5 border border-emerald-200">
                    <MapPin className="w-3.5 h-3.5 text-emerald-600" />
                    <span>GPS: {previewPhoto.geoTag.latitude.toFixed(6)}°, {previewPhoto.geoTag.longitude.toFixed(6)}°</span>
                  </span>
                  <a
                    href={`https://www.google.com/maps?q=${previewPhoto.geoTag.latitude},${previewPhoto.geoTag.longitude}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center gap-1 text-[11px] font-bold text-indigo-600 hover:text-indigo-800 hover:underline"
                  >
                    <span>Buka Peta</span>
                    <ExternalLink className="w-3 h-3" />
                  </a>
                </div>
              ) : (
                <span className="text-[11px] text-slate-400">Tidak ada metadata geotagging</span>
              )}

              {previewPhoto.driveLink && (
                <a
                  href={previewPhoto.driveLink}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center gap-1.5 px-3 py-1 bg-indigo-50 hover:bg-indigo-100 text-indigo-700 text-xs font-bold rounded-lg border border-indigo-200"
                >
                  <HardDrive className="w-3.5 h-3.5 text-indigo-600" />
                  <span>Google Drive</span>
                </a>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
