import React, { useState } from 'react';
import { 
  X, 
  ShieldCheck, 
  User, 
  Building2, 
  CheckCircle2, 
  Lock, 
  Sparkles, 
  Boxes,
  ArrowRight,
  LogIn,
  KeyRound,
  BadgeCheck,
  UserCheck
} from 'lucide-react';
import { OfficeUserProfile } from '../types/inventory';

interface LoginModalProps {
  isOpen: boolean;
  currentUser: OfficeUserProfile | null;
  onClose: () => void;
  onGoogleLogin: () => void;
  isLoggingIn: boolean;
  onStaffLogin: (profile: OfficeUserProfile) => void;
}

export const PRESET_OFFICE_ROLES: OfficeUserProfile[] = [
  {
    id: 'staff-1',
    name: 'Ahmad Faisal, S.T.',
    role: 'Kepala Bagian Umum & GA',
    email: 'ahmad.faisal@kantor.id',
    avatar: 'https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?w=150&auto=format&fit=crop&q=80',
    isGoogleConnected: false,
  },
  {
    id: 'staff-2',
    name: 'Rian Hidayat',
    role: 'Petugas Gudang & Barcode',
    email: 'rian.hidayat@kantor.id',
    avatar: 'https://images.unsplash.com/photo-1519085360753-af0119f7cbe7?w=150&auto=format&fit=crop&q=80',
    isGoogleConnected: false,
  },
  {
    id: 'staff-3',
    name: 'Siti Rahmawati, S.E.',
    role: 'Auditor Aset & Pejabat Pengadaan',
    email: 'siti.rahmawati@kantor.id',
    avatar: 'https://images.unsplash.com/photo-1573496359142-b8d87734a5a2?w=150&auto=format&fit=crop&q=80',
    isGoogleConnected: false,
  },
  {
    id: 'staff-4',
    name: 'Dr. Hendra Wijaya',
    role: 'Penanggung Jawab Aset Teknis',
    email: 'hendra.wijaya@kantor.id',
    avatar: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=150&auto=format&fit=crop&q=80',
    isGoogleConnected: false,
  },
];

export const LoginModal: React.FC<LoginModalProps> = ({
  isOpen,
  currentUser,
  onClose,
  onGoogleLogin,
  isLoggingIn,
  onStaffLogin,
}) => {
  const [activeTab, setActiveTab] = useState<'staff' | 'custom' | 'google'>('staff');
  const [customName, setCustomName] = useState('');
  const [customRole, setCustomRole] = useState('Staff Logistik & Pengadaan');
  const [customEmail, setCustomEmail] = useState('');
  const [customPin, setCustomPin] = useState('');
  const [error, setError] = useState<string | null>(null);

  if (!isOpen) return null;

  const handleCustomSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!customName.trim()) {
      setError('Silakan masukkan nama petugas!');
      return;
    }

    const email = customEmail.trim() || `${customName.toLowerCase().replace(/\s+/g, '.')}@kantor.id`;
    onStaffLogin({
      id: `custom-${Date.now()}`,
      name: customName.trim(),
      role: customRole,
      email: email,
      avatar: `https://api.dicebear.com/7.x/initials/svg?seed=${encodeURIComponent(customName)}`,
      isGoogleConnected: false,
    });
    onClose();
  };

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto bg-slate-950/80 backdrop-blur-md flex items-center justify-center p-3 sm:p-5">
      <div className="bg-white rounded-3xl max-w-lg w-full p-6 sm:p-8 shadow-2xl border border-slate-100 relative animate-in fade-in zoom-in-95 duration-200">
        <button
          onClick={onClose}
          className="absolute top-5 right-5 p-2 text-slate-400 hover:text-slate-700 hover:bg-slate-100 rounded-xl transition-colors cursor-pointer"
        >
          <X className="w-5 h-5" />
        </button>

        {/* Brand Icon Header */}
        <div className="text-center mb-6">
          <div className="w-14 h-14 rounded-2xl bg-gradient-to-tr from-slate-950 via-indigo-900 to-indigo-600 text-white flex items-center justify-center mx-auto shadow-xl shadow-indigo-950/20 mb-3 border border-indigo-500/20">
            <Boxes className="w-7 h-7" />
          </div>
          <h3 className="text-2xl font-black text-slate-900 tracking-tight">
            Autentikasi Petugas Inventaris
          </h3>
          <p className="text-xs text-slate-500 mt-1 max-w-xs mx-auto">
            Masuk sebagai staf resmi untuk otorisasi pencatatan barang baru masuk, cetak label barcode & sync cloud
          </p>
        </div>

        {/* Tab Switcher */}
        <div className="flex p-1 bg-slate-100 rounded-2xl mb-5 text-xs font-bold">
          <button
            type="button"
            onClick={() => setActiveTab('staff')}
            className={`flex-1 py-2 px-3 rounded-xl transition-all cursor-pointer flex items-center justify-center gap-1.5 ${
              activeTab === 'staff'
                ? 'bg-white text-indigo-700 shadow-xs'
                : 'text-slate-500 hover:text-slate-800'
            }`}
          >
            <UserCheck className="w-3.5 h-3.5" />
            <span>Pilih Staf</span>
          </button>
          <button
            type="button"
            onClick={() => setActiveTab('custom')}
            className={`flex-1 py-2 px-3 rounded-xl transition-all cursor-pointer flex items-center justify-center gap-1.5 ${
              activeTab === 'custom'
                ? 'bg-white text-indigo-700 shadow-xs'
                : 'text-slate-500 hover:text-slate-800'
            }`}
          >
            <KeyRound className="w-3.5 h-3.5" />
            <span>Input Akun Baru</span>
          </button>
          <button
            type="button"
            onClick={() => setActiveTab('google')}
            className={`flex-1 py-2 px-3 rounded-xl transition-all cursor-pointer flex items-center justify-center gap-1.5 ${
              activeTab === 'google'
                ? 'bg-white text-indigo-700 shadow-xs'
                : 'text-slate-500 hover:text-slate-800'
            }`}
          >
            <ShieldCheck className="w-3.5 h-3.5" />
            <span>Google Login</span>
          </button>
        </div>

        {error && (
          <div className="mb-4 p-3 bg-rose-50 border border-rose-200 text-rose-800 rounded-xl text-xs font-semibold">
            {error}
          </div>
        )}

        {/* Tab 1: Quick Preset Office Roles */}
        {activeTab === 'staff' && (
          <div className="space-y-2.5">
            <p className="text-xs text-slate-400 font-semibold mb-2">
              Pilih profil staf untuk masuk secara instan:
            </p>
            {PRESET_OFFICE_ROLES.map(role => {
              const isSelected = currentUser?.id === role.id;
              return (
                <div
                  key={role.id}
                  onClick={() => {
                    onStaffLogin(role);
                    onClose();
                  }}
                  className={`p-3.5 rounded-2xl border transition-all cursor-pointer flex items-center justify-between group ${
                    isSelected
                      ? 'bg-indigo-50/70 border-indigo-400 ring-1 ring-indigo-300'
                      : 'bg-white border-slate-200/90 hover:border-indigo-300 hover:bg-slate-50/80 shadow-2xs'
                  }`}
                >
                  <div className="flex items-center gap-3">
                    <img
                      src={role.avatar}
                      alt={role.name}
                      className="w-10 h-10 rounded-xl object-cover ring-1 ring-slate-200"
                    />
                    <div>
                      <div className="flex items-center gap-1.5">
                        <p className="text-xs sm:text-sm font-extrabold text-slate-900 group-hover:text-indigo-600 transition-colors">
                          {role.name}
                        </p>
                        {isSelected && (
                          <span className="text-[10px] font-bold text-indigo-700 bg-indigo-100 px-1.5 py-0.2 rounded-md">
                            Aktif
                          </span>
                        )}
                      </div>
                      <p className="text-xs font-semibold text-slate-500">{role.role}</p>
                      <p className="text-[10px] text-slate-400 font-mono">{role.email}</p>
                    </div>
                  </div>

                  <div className="text-slate-300 group-hover:text-indigo-600 group-hover:translate-x-1 transition-all">
                    <ArrowRight className="w-5 h-5" />
                  </div>
                </div>
              );
            })}
          </div>
        )}

        {/* Tab 2: Custom Staff Input Form */}
        {activeTab === 'custom' && (
          <form onSubmit={handleCustomSubmit} className="space-y-3.5 text-xs">
            <div>
              <label className="block text-xs font-bold text-slate-700 mb-1">
                Nama Petugas Lengkap <span className="text-rose-500">*</span>
              </label>
              <input
                type="text"
                required
                placeholder="Contoh: Budi Santoso, S.Kom."
                value={customName}
                onChange={e => setCustomName(e.target.value)}
                className="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl font-bold text-slate-900 focus:ring-2 focus:ring-indigo-500"
              />
            </div>

            <div>
              <label className="block text-xs font-bold text-slate-700 mb-1">
                Jabatan / Unit Kerja
              </label>
              <select
                value={customRole}
                onChange={e => setCustomRole(e.target.value)}
                className="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl font-bold text-slate-800 focus:ring-2 focus:ring-indigo-500"
              >
                <option value="Staff Logistik & Pengadaan">Staff Logistik & Pengadaan</option>
                <option value="Petugas Penerimaan Barang">Petugas Penerimaan Barang</option>
                <option value="Admin Gudang & Inventaris">Admin Gudang & Inventaris</option>
                <option value="Kepala Bagian Umum & GA">Kepala Bagian Umum & GA</option>
                <option value="Auditor Internal Aset">Auditor Internal Aset</option>
                <option value="Teknisi & Pemeliharaan Aset">Teknisi & Pemeliharaan Aset</option>
              </select>
            </div>

            <div>
              <label className="block text-xs font-bold text-slate-700 mb-1">
                Email Kantor (Opsional)
              </label>
              <input
                type="email"
                placeholder="budi.santoso@kantor.id"
                value={customEmail}
                onChange={e => setCustomEmail(e.target.value)}
                className="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-slate-900 focus:ring-2 focus:ring-indigo-500"
              />
            </div>

            <div>
              <label className="block text-xs font-bold text-slate-700 mb-1">
                PIN Akses Cepat (4-6 Digit)
              </label>
              <input
                type="password"
                maxLength={6}
                placeholder="••••"
                value={customPin}
                onChange={e => setCustomPin(e.target.value)}
                className="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl font-mono text-center tracking-widest text-base focus:ring-2 focus:ring-indigo-500"
              />
            </div>

            <button
              type="submit"
              className="w-full py-3 bg-indigo-600 hover:bg-indigo-500 text-white font-black rounded-xl shadow-lg shadow-indigo-600/30 transition-all cursor-pointer mt-2"
            >
              Masuk & Simpan Profil Petugas
            </button>
          </form>
        )}

        {/* Tab 3: Google Workspace Sign-In */}
        {activeTab === 'google' && (
          <div className="space-y-4 text-center py-2">
            <div className="p-4 bg-slate-50 rounded-2xl border border-slate-200 text-xs text-slate-600 leading-relaxed">
              Login dengan Akun Google resmi kantor Anda untuk mengaktifkan otomatisasi penyimpanan spreadsheet di <strong>Google Sheets</strong> dan backup foto inventaris di <strong>Google Drive</strong>.
            </div>

            <button
              type="button"
              onClick={() => {
                onGoogleLogin();
                onClose();
              }}
              disabled={isLoggingIn}
              className="w-full py-3 px-4 bg-white hover:bg-slate-50 text-slate-800 font-bold text-sm rounded-xl border border-slate-300 shadow-sm flex items-center justify-center gap-3 transition-all cursor-pointer disabled:opacity-50"
            >
              <div className="w-5 h-5">
                <svg viewBox="0 0 48 48" className="w-full h-full">
                  <path fill="#EA4335" d="M24 9.5c3.54 0 6.71 1.22 9.21 3.6l6.85-6.85C35.9 2.38 30.47 0 24 0 14.62 0 6.51 5.38 2.56 13.22l7.98 6.19C12.43 13.72 17.74 9.5 24 9.5z" />
                  <path fill="#4285F4" d="M46.98 24.55c0-1.57-.15-3.09-.38-4.55H24v9.02h12.94c-.58 2.96-2.26 5.48-4.78 7.18l7.73 6c4.51-4.18 7.09-10.36 7.09-17.65z" />
                  <path fill="#FBBC05" d="M10.53 28.59c-.48-1.45-.76-2.99-.76-4.59s.27-3.14.76-4.59l-7.98-6.19C.92 16.46 0 20.12 0 24c0 3.88.92 7.54 2.56 10.78l7.97-6.19z" />
                  <path fill="#34A853" d="M24 48c6.48 0 11.93-2.13 15.89-5.81l-7.73-6c-2.15 1.45-4.92 2.3-8.16 2.3-6.26 0-11.57-4.22-13.47-9.91l-7.98 6.19C6.51 42.62 14.62 48 24 48z" />
                </svg>
              </div>
              <span>{isLoggingIn ? 'Menghubungkan...' : 'Lanjutkan dengan Google'}</span>
            </button>
          </div>
        )}
      </div>
    </div>
  );
};
