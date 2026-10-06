import React, { useState, useRef, useEffect } from 'react';
import { 
  Boxes, 
  ScanLine, 
  ArrowDownLeft, 
  History, 
  Printer, 
  FileSpreadsheet, 
  LogOut, 
  CheckCircle2, 
  FileText,
  LogIn,
  ChevronDown,
  HardDrive,
  UserCheck,
  Building2,
  LayoutDashboard,
  Package,
  Clock,
  Sparkles,
  ShieldCheck,
  Search
} from 'lucide-react';
import { OfficeUserProfile } from '../types/inventory';

interface NavbarProps {
  activeTab: 'dashboard' | 'items' | 'movement-in' | 'history' | 'barcode' | 'sync';
  setActiveTab: (tab: 'dashboard' | 'items' | 'movement-in' | 'history' | 'barcode' | 'sync') => void;
  onOpenScanner: () => void;
  onOpenPDFReport: () => void;
  currentUser: OfficeUserProfile | null;
  onOpenLoginModal: () => void;
  onLogout: () => void;
  isLoggingIn: boolean;
  spreadsheetUrl: string | null;
  isSyncing: boolean;
  lowStockCount: number;
}

export const Navbar: React.FC<NavbarProps> = ({
  activeTab,
  setActiveTab,
  onOpenScanner,
  onOpenPDFReport,
  currentUser,
  onOpenLoginModal,
  onLogout,
  spreadsheetUrl,
  lowStockCount,
}) => {
  const [isProfileMenuOpen, setIsProfileMenuOpen] = useState(false);
  const [currentTime, setCurrentTime] = useState<string>('');
  const [currentDate, setCurrentDate] = useState<string>('');
  const profileMenuRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    const updateClock = () => {
      const now = new Date();
      setCurrentTime(
        now.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' }) + ' WIB'
      );
      setCurrentDate(
        now.toLocaleDateString('id-ID', { weekday: 'short', day: 'numeric', month: 'short', year: 'numeric' })
      );
    };
    updateClock();
    const interval = setInterval(updateClock, 1000);
    return () => clearInterval(interval);
  }, []);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (profileMenuRef.current && !profileMenuRef.current.contains(event.target as Node)) {
        setIsProfileMenuOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  return (
    <header className="sticky top-0 z-40 bg-slate-950/95 border-b border-slate-800/80 text-white backdrop-blur-2xl shadow-lg shadow-black/20">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Top Header Row */}
        <div className="flex items-center justify-between h-16 gap-3 sm:gap-4">
          {/* Brand Identity */}
          <div 
            className="flex items-center gap-3 cursor-pointer select-none group" 
            onClick={() => setActiveTab('dashboard')}
          >
            <div className="w-10 h-10 rounded-xl bg-gradient-to-tr from-emerald-500 via-teal-400 to-indigo-500 flex items-center justify-center text-slate-950 font-black shadow-md shadow-emerald-500/20 group-hover:scale-105 transition-all ring-1 ring-white/20">
              <Boxes className="w-5 h-5 text-slate-950" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="font-black text-white tracking-tight text-lg">
                  Inventaris<span className="text-emerald-400">Kantor</span>
                </span>
                <span className="hidden sm:inline-flex items-center gap-1 text-[10px] font-bold text-emerald-300 bg-emerald-950/80 border border-emerald-800/80 px-2 py-0.5 rounded-md">
                  <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />
                  Live Sync
                </span>
              </div>
              <p className="text-[11px] text-slate-400 font-medium hidden md:block">
                Sistem Manajemen Aset & Penerimaan Barang Masuk Resmi
              </p>
            </div>
          </div>

          {/* Center: Live Time / Status Badge (Desktop) */}
          <div className="hidden xl:flex items-center gap-2 text-xs font-medium text-slate-400 bg-slate-900/90 border border-slate-800 px-3.5 py-1.5 rounded-xl">
            <Clock className="w-3.5 h-3.5 text-indigo-400" />
            <span>{currentDate}</span>
            <span className="text-slate-600">•</span>
            <span className="font-mono font-bold text-slate-200">{currentTime}</span>
          </div>

          {/* Right Tools & User Profile */}
          <div className="flex items-center gap-2 sm:gap-2.5">
            {/* Laporan PDF Button */}
            <button
              onClick={onOpenPDFReport}
              className="inline-flex items-center gap-1.5 px-3 py-2 text-xs font-bold text-slate-200 bg-slate-900 hover:bg-slate-800 border border-slate-800 hover:border-slate-700 rounded-xl transition-all cursor-pointer shadow-xs active:scale-95"
              title="Unduh Rekap Inventaris PDF"
            >
              <FileText className="w-3.5 h-3.5 text-rose-400" />
              <span className="hidden sm:inline">Laporan PDF</span>
            </button>

            {/* Quick Scanner Action */}
            <button
              onClick={onOpenScanner}
              className="inline-flex items-center gap-2 px-3.5 py-2 text-xs sm:text-sm font-bold text-white bg-indigo-600 hover:bg-indigo-500 active:scale-95 rounded-xl shadow-md shadow-indigo-600/30 transition-all cursor-pointer ring-1 ring-indigo-400/30"
              title="Pindai Barcode / QR dengan Kamera atau Barcode Scanner"
            >
              <ScanLine className="w-4 h-4 text-indigo-200" />
              <span className="hidden xs:inline">Scan Barcode</span>
            </button>

            {/* Google Sheets Link */}
            {spreadsheetUrl && (
              <a
                href={spreadsheetUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="hidden lg:inline-flex items-center gap-1.5 px-3 py-2 text-xs font-bold text-emerald-300 bg-emerald-950/60 hover:bg-emerald-900/60 border border-emerald-800/60 rounded-xl transition-colors"
                title="Buka Spreadsheet di Google Sheets"
              >
                <FileSpreadsheet className="w-3.5 h-3.5 text-emerald-400" />
                <span>Sheets</span>
              </a>
            )}

            {/* User Profile / Login */}
            {currentUser ? (
              <div className="relative" ref={profileMenuRef}>
                <button
                  type="button"
                  onClick={() => setIsProfileMenuOpen(!isProfileMenuOpen)}
                  className="flex items-center gap-2.5 p-1.5 pr-2.5 bg-slate-900 hover:bg-slate-800 border border-slate-800 hover:border-slate-700 rounded-xl transition-all cursor-pointer shadow-xs group"
                >
                  <div className="relative">
                    {currentUser.avatar ? (
                      <img
                        src={currentUser.avatar}
                        alt={currentUser.name}
                        className="w-7 h-7 rounded-lg object-cover ring-1 ring-slate-700"
                      />
                    ) : (
                      <div className="w-7 h-7 rounded-lg bg-emerald-600 text-white flex items-center justify-center font-bold text-xs ring-1 ring-slate-700">
                        {currentUser.name.charAt(0).toUpperCase()}
                      </div>
                    )}
                    <span className="absolute -bottom-0.5 -right-0.5 w-2.5 h-2.5 bg-emerald-400 border-2 border-slate-950 rounded-full" />
                  </div>

                  <div className="hidden sm:block text-left">
                    <p className="text-xs font-bold text-white leading-tight max-w-[130px] truncate">
                      {currentUser.name}
                    </p>
                    <p className="text-[10px] text-slate-400 truncate max-w-[130px]">
                      {currentUser.role}
                    </p>
                  </div>

                  <ChevronDown className="w-3.5 h-3.5 text-slate-400 group-hover:text-white transition-transform" />
                </button>

                {/* Profile Dropdown */}
                {isProfileMenuOpen && (
                  <div className="absolute right-0 mt-2 w-72 bg-slate-900 rounded-2xl shadow-2xl border border-slate-800 p-3 z-50 animate-in fade-in zoom-in-95 duration-150">
                    <div className="p-3 bg-slate-950 rounded-xl border border-slate-800/80 mb-2">
                      <div className="flex items-center gap-3">
                        <img
                          src={currentUser.avatar || 'https://api.dicebear.com/7.x/initials/svg?seed=Staff'}
                          alt={currentUser.name}
                          className="w-10 h-10 rounded-xl object-cover ring-1 ring-slate-700"
                        />
                        <div className="min-w-0">
                          <p className="text-xs font-bold text-white truncate">{currentUser.name}</p>
                          <p className="text-[11px] text-emerald-400 font-medium truncate">{currentUser.role}</p>
                          <p className="text-[10px] text-slate-400 font-mono truncate">{currentUser.email}</p>
                        </div>
                      </div>

                      <div className="mt-3 pt-2 border-t border-slate-800/80 flex items-center justify-between text-[10px]">
                        <span className="text-slate-400">Petugas Aktif:</span>
                        <span className="text-emerald-400 font-bold flex items-center gap-1">
                          <CheckCircle2 className="w-3 h-3" />
                          <span>Terverifikasi</span>
                        </span>
                      </div>
                    </div>

                    <div className="space-y-1">
                      <button
                        onClick={() => {
                          setIsProfileMenuOpen(false);
                          setActiveTab('sync');
                        }}
                        className="w-full px-3 py-2 text-xs font-medium text-slate-300 hover:text-white hover:bg-slate-800 rounded-lg transition-colors flex items-center gap-2 cursor-pointer"
                      >
                        <HardDrive className="w-3.5 h-3.5 text-indigo-400" />
                        <span>Google Sheets & Drive</span>
                      </button>

                      <button
                        onClick={() => {
                          setIsProfileMenuOpen(false);
                          onOpenLoginModal();
                        }}
                        className="w-full px-3 py-2 text-xs font-medium text-slate-300 hover:text-white hover:bg-slate-800 rounded-lg transition-colors flex items-center gap-2 cursor-pointer"
                      >
                        <UserCheck className="w-3.5 h-3.5 text-emerald-400" />
                        <span>Ganti Profil / Akun Petugas</span>
                      </button>

                      <div className="border-t border-slate-800 my-1" />

                      <button
                        onClick={() => {
                          setIsProfileMenuOpen(false);
                          onLogout();
                        }}
                        className="w-full px-3 py-2 text-xs font-bold text-rose-400 hover:bg-rose-950/40 rounded-lg transition-colors flex items-center gap-2 cursor-pointer"
                      >
                        <LogOut className="w-3.5 h-3.5" />
                        <span>Keluar (Logout)</span>
                      </button>
                    </div>
                  </div>
                )}
              </div>
            ) : (
              <button
                type="button"
                onClick={onOpenLoginModal}
                className="inline-flex items-center gap-2 px-3.5 py-2 text-xs font-bold text-white bg-indigo-600 hover:bg-indigo-500 rounded-xl transition-all cursor-pointer shadow-sm ring-1 ring-indigo-400/20"
              >
                <LogIn className="w-3.5 h-3.5" />
                <span>Masuk / Login</span>
              </button>
            )}
          </div>
        </div>

        {/* Tab Navigation Menu - Sleek Segmented Style */}
        <nav className="flex items-center gap-1 sm:gap-1.5 overflow-x-auto py-2 no-scrollbar border-t border-slate-800/80 text-xs sm:text-sm">
          <button
            onClick={() => setActiveTab('dashboard')}
            className={`px-3 py-1.5 rounded-lg font-semibold transition-all whitespace-nowrap cursor-pointer flex items-center gap-2 ${
              activeTab === 'dashboard'
                ? 'bg-slate-800 text-white shadow-xs ring-1 ring-white/10'
                : 'text-slate-400 hover:text-white hover:bg-slate-900/80'
            }`}
          >
            <LayoutDashboard className="w-4 h-4 text-slate-400" />
            <span>Dashboard</span>
          </button>

          <button
            onClick={() => setActiveTab('items')}
            className={`px-3 py-1.5 rounded-lg font-semibold transition-all whitespace-nowrap cursor-pointer flex items-center gap-2 ${
              activeTab === 'items'
                ? 'bg-slate-800 text-white shadow-xs ring-1 ring-white/10'
                : 'text-slate-400 hover:text-white hover:bg-slate-900/80'
            }`}
          >
            <Package className="w-4 h-4 text-slate-400" />
            <span>Master Barang</span>
            {lowStockCount > 0 && (
              <span className="text-[10px] font-bold text-amber-400 bg-amber-950/80 border border-amber-800 px-1.5 py-0.2 rounded-full">
                {lowStockCount}
              </span>
            )}
          </button>

          {/* Highlighted Barang Masuk (+) Button */}
          <button
            onClick={() => setActiveTab('movement-in')}
            className={`px-3.5 py-1.5 rounded-lg font-bold transition-all whitespace-nowrap cursor-pointer flex items-center gap-1.5 ${
              activeTab === 'movement-in'
                ? 'bg-emerald-500 text-slate-950 shadow-md shadow-emerald-500/20 ring-1 ring-emerald-300'
                : 'text-emerald-400 bg-emerald-950/60 hover:bg-emerald-900/60 border border-emerald-800/60'
            }`}
          >
            <ArrowDownLeft className="w-4 h-4" />
            <span>Barang Masuk (+)</span>
          </button>

          <button
            onClick={() => setActiveTab('history')}
            className={`px-3 py-1.5 rounded-lg font-semibold transition-all whitespace-nowrap cursor-pointer flex items-center gap-2 ${
              activeTab === 'history'
                ? 'bg-slate-800 text-white shadow-xs ring-1 ring-white/10'
                : 'text-slate-400 hover:text-white hover:bg-slate-900/80'
            }`}
          >
            <History className="w-4 h-4 text-slate-400" />
            <span>Riwayat Barang Masuk</span>
          </button>

          <button
            onClick={() => setActiveTab('barcode')}
            className={`px-3 py-1.5 rounded-lg font-semibold transition-all whitespace-nowrap cursor-pointer flex items-center gap-2 ${
              activeTab === 'barcode'
                ? 'bg-slate-800 text-white shadow-xs ring-1 ring-white/10'
                : 'text-slate-400 hover:text-white hover:bg-slate-900/80'
            }`}
          >
            <Printer className="w-4 h-4 text-slate-400" />
            <span>Cetak Barcode & Label</span>
          </button>

          <button
            onClick={() => setActiveTab('sync')}
            className={`px-3 py-1.5 rounded-lg font-semibold transition-all whitespace-nowrap cursor-pointer flex items-center gap-2 ${
              activeTab === 'sync'
                ? 'bg-slate-800 text-white shadow-xs ring-1 ring-white/10'
                : 'text-slate-400 hover:text-white hover:bg-slate-900/80'
            }`}
          >
            <FileSpreadsheet className="w-4 h-4 text-slate-400" />
            <span>Google Sheets & Drive</span>
          </button>
        </nav>
      </div>
    </header>
  );
};
