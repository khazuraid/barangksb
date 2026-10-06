import React from 'react';
import { 
  FileSpreadsheet, 
  ExternalLink, 
  RefreshCw, 
  CheckCircle2, 
  AlertCircle, 
  HardDrive, 
  Layers, 
  ArrowDownLeft, 
  Clock, 
  ShieldCheck, 
  FolderGit2,
  Cloud,
  FileCheck2,
  Lock,
  Sparkles
} from 'lucide-react';
import { User } from 'firebase/auth';

interface GoogleSyncPanelProps {
  user: User | null;
  onLogin: () => void;
  onLogout: () => void;
  isLoggingIn: boolean;
  spreadsheetId: string | null;
  spreadsheetUrl: string | null;
  lastSyncedAt: string | null;
  isSyncing: boolean;
  onTriggerSync: () => void;
  itemsCount: number;
  transactionsCount: number;
}

export const GoogleSyncPanel: React.FC<GoogleSyncPanelProps> = ({
  user,
  onLogin,
  onLogout,
  isLoggingIn,
  spreadsheetId,
  spreadsheetUrl,
  lastSyncedAt,
  isSyncing,
  onTriggerSync,
  itemsCount,
  transactionsCount,
}) => {
  return (
    <div className="space-y-6">
      {/* Header Banner */}
      <div className="bg-white rounded-3xl p-6 sm:p-8 border border-slate-200/90 shadow-xs">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-6">
          <div className="flex items-start gap-4">
            <div className="w-12 h-12 rounded-2xl bg-emerald-500 text-white flex items-center justify-center shrink-0 shadow-md shadow-emerald-500/20">
              <FileSpreadsheet className="w-6 h-6" />
            </div>
            <div>
              <div className="flex items-center gap-2 flex-wrap">
                <h2 className="text-xl sm:text-2xl font-black text-slate-900 tracking-tight">
                  Integrasi Google Sheets & Google Drive Cloud
                </h2>
                <span className="px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-emerald-100 text-emerald-800 border border-emerald-200">
                  Resmi Google Workspace API
                </span>
              </div>
              <p className="text-xs text-slate-500 mt-1 max-w-2xl leading-relaxed">
                Aplikasi ini terhubung langsung dengan Google Drive dan Google Sheets akun Anda. Semua data Master Aset dan log transaksi Penerimaan Barang Baru Masuk tersinkronisasi otomatis dalam format spreadsheet resmi untuk pelaporan instansi yang akurat.
              </p>
            </div>
          </div>

          <div>
            {!user ? (
              <button
                type="button"
                onClick={onLogin}
                disabled={isLoggingIn}
                className="inline-flex items-center gap-2.5 px-5 py-2.5 text-sm font-bold text-slate-700 bg-white hover:bg-slate-50 border border-slate-300 rounded-xl shadow-xs transition-all cursor-pointer ring-1 ring-slate-200"
              >
                <div className="w-4 h-4">
                  <svg viewBox="0 0 48 48" className="w-full h-full">
                    <path fill="#EA4335" d="M24 9.5c3.54 0 6.71 1.22 9.21 3.6l6.85-6.85C35.9 2.38 30.47 0 24 0 14.62 0 6.51 5.38 2.56 13.22l7.98 6.19C12.43 13.72 17.74 9.5 24 9.5z" />
                    <path fill="#4285F4" d="M46.98 24.55c0-1.57-.15-3.09-.38-4.55H24v9.02h12.94c-.58 2.96-2.26 5.48-4.78 7.18l7.73 6c4.51-4.18 7.09-10.36 7.09-17.65z" />
                    <path fill="#FBBC05" d="M10.53 28.59c-.48-1.45-.76-2.99-.76-4.59s.27-3.14.76-4.59l-7.98-6.19C.92 16.46 0 20.12 0 24c0 3.88.92 7.54 2.56 10.78l7.97-6.19z" />
                    <path fill="#34A853" d="M24 48c6.48 0 11.93-2.13 15.89-5.81l-7.73-6c-2.15 1.45-4.92 2.3-8.16 2.3-6.26 0-11.57-4.22-13.47-9.91l-7.98 6.19C6.51 42.62 14.62 48 24 48z" />
                  </svg>
                </div>
                <span>{isLoggingIn ? 'Memproses...' : 'Hubungkan Akun Google'}</span>
              </button>
            ) : (
              <button
                type="button"
                onClick={onTriggerSync}
                disabled={isSyncing}
                className="inline-flex items-center gap-2 px-5 py-2.5 bg-emerald-600 hover:bg-emerald-500 text-white font-bold text-sm rounded-xl shadow-md shadow-emerald-600/30 transition-all cursor-pointer disabled:opacity-50 ring-1 ring-emerald-400"
              >
                <RefreshCw className={`w-4 h-4 ${isSyncing ? 'animate-spin' : ''}`} />
                <span>{isSyncing ? 'Menyinkronkan...' : 'Sinkronisasi Penuh Sekarang'}</span>
              </button>
            )}
          </div>
        </div>
      </div>

      {/* Connection & Spreadsheet Card */}
      {user ? (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Left: User Account Card */}
          <div className="bg-white rounded-3xl p-6 border border-slate-200/90 shadow-xs space-y-4">
            <h3 className="text-xs font-bold uppercase tracking-wider text-slate-400">
              Akun Google Terhubung
            </h3>
            
            <div className="flex items-center gap-3 p-3.5 bg-slate-50 rounded-2xl border border-slate-200">
              {user.photoURL ? (
                <img
                  src={user.photoURL}
                  alt={user.displayName || 'User'}
                  className="w-12 h-12 rounded-xl object-cover ring-1 ring-slate-300"
                />
              ) : (
                <div className="w-12 h-12 rounded-xl bg-indigo-600 text-white flex items-center justify-center font-bold text-lg">
                  {(user.displayName || user.email || 'U')[0].toUpperCase()}
                </div>
              )}
              <div className="min-w-0 flex-1">
                <p className="font-extrabold text-slate-900 text-sm truncate">{user.displayName || 'Pengguna Terhubung'}</p>
                <p className="text-xs text-slate-500 truncate font-mono">{user.email}</p>
                <p className="text-[10px] text-emerald-700 font-bold flex items-center gap-1 mt-0.5">
                  <CheckCircle2 className="w-3 h-3 text-emerald-600" />
                  <span>Izin Google Drive & Sheets Aktif</span>
                </p>
              </div>
            </div>

            <div className="pt-2 flex items-center justify-between text-xs text-slate-500">
              <span>Ingin berganti akun Google?</span>
              <button
                type="button"
                onClick={onLogout}
                className="text-rose-600 hover:underline font-bold cursor-pointer"
              >
                Putuskan Sambungan
              </button>
            </div>
          </div>

          {/* Right: Connected Spreadsheet Info */}
          <div className="bg-white rounded-3xl p-6 border border-slate-200/90 shadow-xs space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="text-xs font-bold uppercase tracking-wider text-slate-400">
                Spreadsheet Dokumen Resmi
              </h3>
              {lastSyncedAt && (
                <span className="text-[10px] text-slate-400 font-mono">
                  Sync: {new Date(lastSyncedAt).toLocaleTimeString('id-ID')}
                </span>
              )}
            </div>

            <div className="p-3.5 bg-emerald-50/50 rounded-2xl border border-emerald-200/80">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-xl bg-emerald-600 text-white flex items-center justify-center shrink-0">
                  <FileSpreadsheet className="w-5 h-5" />
                </div>
                <div className="min-w-0 flex-1">
                  <p className="font-extrabold text-slate-900 text-sm truncate">
                    Inventaris Kantor & Log Barang Masuk
                  </p>
                  <p className="text-[11px] text-slate-500 font-mono truncate">
                    ID: {spreadsheetId || 'Otomatis di Google Drive'}
                  </p>
                </div>
              </div>

              {spreadsheetUrl && (
                <div className="mt-3 pt-2.5 border-t border-emerald-200/60 flex items-center justify-between">
                  <span className="text-[10px] font-bold text-emerald-800">
                    Lembar Kerja Aktif: 2 Tab (Master & Log Masuk)
                  </span>
                  <a
                    href={spreadsheetUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center gap-1 text-xs font-black text-emerald-700 hover:underline"
                  >
                    <span>Buka File di Sheets</span>
                    <ExternalLink className="w-3.5 h-3.5" />
                  </a>
                </div>
              )}
            </div>

            <div className="grid grid-cols-2 gap-3 text-center text-xs">
              <div className="p-2.5 rounded-xl bg-slate-50 border border-slate-200">
                <p className="text-[10px] text-slate-400 font-bold uppercase">Master Barang</p>
                <p className="text-base font-black text-slate-900 mt-0.5">{itemsCount} Baris</p>
              </div>
              <div className="p-2.5 rounded-xl bg-slate-50 border border-slate-200">
                <p className="text-[10px] text-slate-400 font-bold uppercase">Log Barang Masuk</p>
                <p className="text-base font-black text-slate-900 mt-0.5">{transactionsCount} Baris</p>
              </div>
            </div>
          </div>
        </div>
      ) : (
        /* Empty State / Call to Action */
        <div className="bg-slate-50 rounded-3xl p-8 border border-dashed border-slate-300 text-center space-y-4">
          <div className="w-16 h-16 rounded-3xl bg-indigo-50 text-indigo-600 flex items-center justify-center mx-auto">
            <HardDrive className="w-8 h-8" />
          </div>
          <div className="max-w-md mx-auto">
            <h3 className="font-extrabold text-slate-900 text-base">
              Hubungkan dengan Google Drive & Sheets
            </h3>
            <p className="text-xs text-slate-500 mt-1 leading-relaxed">
              Dapatkan kemudahan backup spreadsheet instan ke Google Drive kantor Anda. Setiap barang baru masuk dan foto geotagging akan otomatis terdata secara real-time.
            </p>
          </div>
          <button
            type="button"
            onClick={onLogin}
            disabled={isLoggingIn}
            className="inline-flex items-center gap-2 px-6 py-3 bg-indigo-600 hover:bg-indigo-500 text-white font-bold text-sm rounded-xl shadow-lg shadow-indigo-600/30 transition-all cursor-pointer"
          >
            <span>Hubungkan Akun Google Anda</span>
          </button>
        </div>
      )}

      {/* Cloud & Drive Architecture Feature Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-xs">
        <div className="bg-white p-5 rounded-3xl border border-slate-200/90 shadow-xs space-y-2">
          <div className="w-9 h-9 rounded-xl bg-emerald-50 text-emerald-600 flex items-center justify-center font-bold">
            <CheckCircle2 className="w-5 h-5" />
          </div>
          <h4 className="font-extrabold text-slate-900 text-sm">Sinkronisasi Real-Time</h4>
          <p className="text-slate-500 leading-relaxed">
            Setiap formulir barang masuk disubmit, baris baru langsung ditambahkan ke Google Sheets dan database cloud.
          </p>
        </div>

        <div className="bg-white p-5 rounded-3xl border border-slate-200/90 shadow-xs space-y-2">
          <div className="w-9 h-9 rounded-xl bg-indigo-50 text-indigo-600 flex items-center justify-center font-bold">
            <HardDrive className="w-5 h-5" />
          </div>
          <h4 className="font-extrabold text-slate-900 text-sm">Penyimpanan Foto Drive</h4>
          <p className="text-slate-500 leading-relaxed">
            Foto bukti fisik beserta watermark koordinat GPS geotagging terarsip dan dapat dibuka melalui tautan langsung.
          </p>
        </div>

        <div className="bg-white p-5 rounded-3xl border border-slate-200/90 shadow-xs space-y-2">
          <div className="w-9 h-9 rounded-xl bg-slate-100 text-slate-700 flex items-center justify-center font-bold">
            <Lock className="w-5 h-5" />
          </div>
          <h4 className="font-extrabold text-slate-900 text-sm">Standar Format Pengadaan</h4>
          <p className="text-slate-500 leading-relaxed">
            Struktur tabel spreadsheet mengikuti standar resmi instansi (No Seri, Merk, Type, Tahun, Kondisi Berfungsi, Nilai, AKL/AKD).
          </p>
        </div>
      </div>
    </div>
  );
};
