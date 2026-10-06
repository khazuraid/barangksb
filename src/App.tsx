/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useEffect } from 'react';
import { User } from 'firebase/auth';
import { 
  initAuth, 
  googleSignIn, 
  getAccessToken, 
  logout 
} from './services/firebaseAuth';
import { 
  findOrCreateSpreadsheet, 
  syncAllToGoogleSheets, 
  appendTransactionRealtime 
} from './services/googleSheetsService';
import { 
  getStoredItems, 
  saveStoredItems, 
  getStoredTransactions, 
  saveStoredTransactions,
  getStoredCategories,
  saveStoredCategories
} from './services/storageService';
import { 
  subscribeItems, 
  saveItemToFirestore, 
  deleteItemFromFirestore,
  subscribeTransactions, 
  saveTransactionToFirestore,
  subscribeCategories,
  saveCategoryToFirestore,
  seedInitialFirestoreData,
  testConnection
} from './services/firestoreService';
import { InventoryItem, StockTransaction, NewIncomingItemData, OfficeUserProfile } from './types/inventory';
import { Navbar } from './components/Navbar';
import { DashboardView } from './components/DashboardView';
import { InventoryList } from './components/InventoryList';
import { ItemModal } from './components/ItemModal';
import { StockMovementModal } from './components/StockMovementModal';
import { BarcodeGeneratorModal } from './components/BarcodeGeneratorModal';
import { BarcodeScannerModal } from './components/BarcodeScannerModal';
import { TransactionHistoryView } from './components/TransactionHistoryView';
import { GoogleSyncPanel } from './components/GoogleSyncPanel';
import { PDFReportModal } from './components/PDFReportModal';
import { LoginModal, PRESET_OFFICE_ROLES } from './components/LoginModal';
import { ConfirmDialog } from './components/ConfirmDialog';
import { playSuccessSound } from './utils/audio';
import { CheckCircle2, AlertCircle, Info, X, Cloud, CloudCheck } from 'lucide-react';

export default function App() {
  // Navigation: Barang Masuk only
  const [activeTab, setActiveTab] = useState<
    'dashboard' | 'items' | 'movement-in' | 'history' | 'barcode' | 'sync'
  >('dashboard');

  // Authentication State
  const [user, setUser] = useState<User | null>(null);
  const [currentUser, setCurrentUser] = useState<OfficeUserProfile | null>(() => {
    const saved = localStorage.getItem('inventaris_current_user');
    if (saved) {
      try {
        return JSON.parse(saved);
      } catch {
        // ignore
      }
    }
    return PRESET_OFFICE_ROLES[0];
  });
  const [isLoginModalOpen, setIsLoginModalOpen] = useState<boolean>(false);
  const [isLoggingIn, setIsLoggingIn] = useState<boolean>(false);

  // Firestore Database Connection State
  const [isDatabaseConnected, setIsDatabaseConnected] = useState<boolean>(false);

  // Google Sheets state
  const [spreadsheetId, setSpreadsheetId] = useState<string | null>(() => {
    return localStorage.getItem('inventaris_spreadsheet_id') || null;
  });
  const [spreadsheetUrl, setSpreadsheetUrl] = useState<string | null>(() => {
    return localStorage.getItem('inventaris_spreadsheet_url') || null;
  });
  const [lastSyncedAt, setLastSyncedAt] = useState<string | null>(() => {
    return localStorage.getItem('inventaris_last_sync') || null;
  });
  const [isSyncing, setIsSyncing] = useState<boolean>(false);

  // Inventory Data & Categories State
  const [items, setItems] = useState<InventoryItem[]>(() => getStoredItems());
  const [transactions, setTransactions] = useState<StockTransaction[]>(() => getStoredTransactions());
  const [categories, setCategories] = useState<string[]>(() => getStoredCategories());

  // Toast notification state
  const [toast, setToast] = useState<{ message: string; type: 'success' | 'error' | 'info' } | null>(null);

  const showToast = (message: string, type: 'success' | 'error' | 'info' = 'success') => {
    setToast({ message, type });
    setTimeout(() => {
      setToast(prev => (prev?.message === message ? null : prev));
    }, 4500);
  };

  // Modals state
  const [isItemModalOpen, setIsItemModalOpen] = useState(false);
  const [editingItem, setEditingItem] = useState<InventoryItem | null>(null);

  const [isMovementModalOpen, setIsMovementModalOpen] = useState(false);
  const [movementInitialSku, setMovementInitialSku] = useState<string | undefined>(undefined);
  const [movementInitialData, setMovementInitialData] = useState<Partial<NewIncomingItemData> | undefined>(undefined);

  const [isBarcodeModalOpen, setIsBarcodeModalOpen] = useState(false);
  const [barcodeActiveItem, setBarcodeActiveItem] = useState<InventoryItem | null>(null);

  const [isScannerOpen, setIsScannerOpen] = useState(false);
  const [isPDFModalOpen, setIsPDFModalOpen] = useState(false);

  // Confirm dialog state
  const [confirmDialog, setConfirmDialog] = useState<{
    isOpen: boolean;
    title: string;
    message: string;
    type?: 'danger' | 'warning' | 'info';
    confirmText?: string;
    onConfirm: () => void;
  }>({
    isOpen: false,
    title: '',
    message: '',
    onConfirm: () => {},
  });

  // 1. Initialize Firestore real-time subscriptions & testing
  useEffect(() => {
    let unsubItems = () => {};
    let unsubTx = () => {};
    let unsubCats = () => {};

    const setupDatabase = async () => {
      try {
        const isOnline = await testConnection();
        setIsDatabaseConnected(isOnline);

        // Seed initial data to Firestore if empty
        await seedInitialFirestoreData(items, transactions);

        // Subscribe to items from Firestore in real-time
        unsubItems = subscribeItems((cloudItems) => {
          if (cloudItems && cloudItems.length > 0) {
            setItems(cloudItems);
            saveStoredItems(cloudItems);
            setIsDatabaseConnected(true);
          }
        });

        // Subscribe to stock-in transactions
        unsubTx = subscribeTransactions((cloudTxs) => {
          if (cloudTxs && cloudTxs.length > 0) {
            setTransactions(cloudTxs);
            saveStoredTransactions(cloudTxs);
          }
        });

        // Subscribe to categories
        unsubCats = subscribeCategories((cloudCats) => {
          if (cloudCats && cloudCats.length > 0) {
            setCategories(cloudCats);
            saveStoredCategories(cloudCats);
          }
        });
      } catch (e) {
        console.warn('Database initialization note:', e);
      }
    };

    setupDatabase();

    return () => {
      unsubItems();
      unsubTx();
      unsubCats();
    };
  }, []);

  // 2. Initialize Google Workspace Auth
  useEffect(() => {
    const unsubscribe = initAuth(
      (currentUser, token) => {
        setUser(currentUser);
        setupGoogleSheet(token);
      },
      () => {
        setUser(null);
      }
    );
    return () => unsubscribe();
  }, []);

  const setupGoogleSheet = async (token: string) => {
    try {
      const sheetInfo = await findOrCreateSpreadsheet(token);
      setSpreadsheetId(sheetInfo.id);
      setSpreadsheetUrl(sheetInfo.webViewLink);
      localStorage.setItem('inventaris_spreadsheet_id', sheetInfo.id);
      localStorage.setItem('inventaris_spreadsheet_url', sheetInfo.webViewLink);
    } catch (err) {
      console.warn('Google Sheet setup note:', err);
    }
  };

  const handleGoogleLogin = async () => {
    setIsLoggingIn(true);
    try {
      const result = await googleSignIn();
      if (result) {
        setUser(result.user);
        const gUser: OfficeUserProfile = {
          id: result.user.uid,
          name: result.user.displayName || 'Akun Google Kantor',
          email: result.user.email || 'user@kantor.id',
          role: 'Staff Inventaris & Admin Google',
          avatar: result.user.photoURL || undefined,
          isGoogleConnected: true,
        };
        setCurrentUser(gUser);
        localStorage.setItem('inventaris_current_user', JSON.stringify(gUser));
        showToast(`Selamat datang, ${gUser.name}! Terhubung ke Google Workspace.`, 'success');
        await setupGoogleSheet(result.accessToken);
        setIsLoginModalOpen(false);
      }
    } catch (err: any) {
      console.error('Login error:', err);
      showToast(err?.message || 'Gagal masuk dengan Google', 'error');
    } finally {
      setIsLoggingIn(false);
    }
  };

  const handleLogout = async () => {
    await logout();
    setUser(null);
    setCurrentUser(null);
    localStorage.removeItem('inventaris_current_user');
    showToast('Berhasil keluar dari akun.', 'info');
  };

  const lowStockCount = items.filter(i => i.currentStock <= i.minStock).length;

  // Add custom category
  const handleAddNewCategory = (newCat: string) => {
    if (!categories.includes(newCat)) {
      const updated = [...categories, newCat];
      setCategories(updated);
      saveStoredCategories(updated);
      saveCategoryToFirestore(newCat).catch(console.warn);
      showToast(`Kategori "${newCat}" berhasil ditambahkan & disimpan ke database!`, 'success');
    }
  };

  // Full Sync with Google Sheets
  const handlePerformFullSync = async () => {
    const token = await getAccessToken();
    if (!token) {
      handleGoogleLogin();
      return;
    }

    setIsSyncing(true);
    try {
      let activeSheetId = spreadsheetId;
      if (!activeSheetId) {
        const sheetInfo = await findOrCreateSpreadsheet(token);
        activeSheetId = sheetInfo.id;
        setSpreadsheetId(sheetInfo.id);
        setSpreadsheetUrl(sheetInfo.webViewLink);
        localStorage.setItem('inventaris_spreadsheet_id', sheetInfo.id);
        localStorage.setItem('inventaris_spreadsheet_url', sheetInfo.webViewLink);
      }

      await syncAllToGoogleSheets(token, activeSheetId, items, transactions);
      const now = new Date().toISOString();
      setLastSyncedAt(now);
      localStorage.setItem('inventaris_last_sync', now);
      playSuccessSound();
      showToast('Seluruh data berhasil disinkronkan ke Google Sheets!', 'success');
    } catch (err: any) {
      console.error('Sync failed:', err);
      showToast(`Gagal sinkronisasi: ${err?.message || err}`, 'error');
    } finally {
      setIsSyncing(false);
    }
  };

  const confirmAndTriggerSync = () => {
    setConfirmDialog({
      isOpen: true,
      title: 'Sinkronisasi ke Google Sheets',
      message: `Aplikasi akan memperbarui spreadsheet "Inventaris Kantor & Log Barang Masuk" di Google Drive Anda dengan ${items.length} master barang dan ${transactions.length} riwayat barang masuk.\n\nLanjutkan sinkronisasi?`,
      type: 'info',
      confirmText: 'Sinkronkan Sekarang',
      onConfirm: () => {
        setConfirmDialog(prev => ({ ...prev, isOpen: false }));
        handlePerformFullSync();
      },
    });
  };

  // Save or Update Item (Syncs to Firestore & Google Sheets)
  const handleSaveItem = async (itemData: Partial<InventoryItem>) => {
    let savedItem: InventoryItem;
    let updatedList: InventoryItem[];

    if (editingItem) {
      savedItem = {
        ...editingItem,
        ...itemData,
        updatedAt: new Date().toISOString(),
      } as InventoryItem;

      updatedList = items.map(i => i.id === editingItem.id ? savedItem : i);
      showToast(`Data "${savedItem.name}" berhasil diperbarui & disimpan ke database!`, 'success');
    } else {
      savedItem = {
        id: `item-${Date.now()}`,
        sku: itemData.sku!,
        name: itemData.name!,
        category: itemData.category!,
        location: itemData.location || 'Gudang Kantor',
        currentStock: itemData.currentStock || 0,
        minStock: itemData.minStock || 0,
        unit: itemData.unit || 'Pcs',
        pricePerUnit: itemData.pricePerUnit || 0,
        description: itemData.description,
        photoUrl: itemData.photoUrl,
        barcodeFormat: itemData.barcodeFormat || 'CODE128',
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };
      updatedList = [savedItem, ...items];
      showToast(`Barang baru "${savedItem.name}" [${savedItem.sku}] tersimpan ke database!`, 'success');
    }

    setItems(updatedList);
    saveStoredItems(updatedList);
    playSuccessSound();
    setIsItemModalOpen(false);
    setEditingItem(null);

    // Save to Firestore Database in real-time
    saveItemToFirestore(savedItem).catch(err => console.warn('Firestore item save error:', err));

    // Auto-sync to Google Sheets in background if logged in
    getAccessToken().then(token => {
      if (token && spreadsheetId) {
        syncAllToGoogleSheets(token, spreadsheetId, updatedList, transactions).catch(console.warn);
      }
    });
  };

  // Delete Item (Firestore & Google Sheets)
  const handleDeleteItem = (itemToDelete: InventoryItem) => {
    setConfirmDialog({
      isOpen: true,
      title: 'Hapus Data Barang?',
      message: `Apakah Anda yakin ingin menghapus "${itemToDelete.name}" (${itemToDelete.sku}) dari sistem database inventaris kantor?\n\nTindakan ini tidak dapat dibatalkan.`,
      type: 'danger',
      confirmText: 'Hapus Barang',
      onConfirm: () => {
        const updatedList = items.filter(i => i.id !== itemToDelete.id);
        setItems(updatedList);
        saveStoredItems(updatedList);
        setConfirmDialog(prev => ({ ...prev, isOpen: false }));
        showToast(`Barang "${itemToDelete.name}" telah dihapus dari database.`, 'info');

        // Delete from Firestore
        deleteItemFromFirestore(itemToDelete.id).catch(console.warn);

        // Update Google Sheets
        getAccessToken().then(token => {
          if (token && spreadsheetId) {
            syncAllToGoogleSheets(token, spreadsheetId, updatedList, transactions).catch(console.warn);
          }
        });
      },
    });
  };

  // Handle Real-Time Penerimaan Barang Baru Masuk (+) with Photo & Geotagging
  const handleStockMovementSubmit = async (data: NewIncomingItemData) => {
    const newItemId = `item-${Date.now()}`;
    const qty = Number(data.quantity) || 1;

    // 1. Create the NEW InventoryItem
    const newItem: InventoryItem = {
      id: newItemId,
      sku: data.sku,
      name: data.name,
      category: data.category,
      location: data.location || 'Gudang Kantor',
      currentStock: qty, // Stok fisik barang baru masuk
      minStock: Number(data.minStock) || 0,
      unit: data.unit,
      pricePerUnit: Number(data.pricePerUnit) || 0,
      description: data.notes || data.description || '',
      photoUrl: data.photoUrl,
      geoTag: data.geoTag,
      driveFileLink: data.driveFileLink,
      merk: data.merk,
      typeModel: data.typeModel,
      serialNumber: data.serialNumber,
      procurementYear: data.procurementYear,
      conditionStatus: data.conditionStatus,
      fundingSource: data.fundingSource,
      distributor: data.distributor,
      aklAkd: data.aklAkd,
      isAvailable: data.isAvailable !== false,
      barcodeFormat: 'CODE128',
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };

    // 2. Create the Transaction Record in Riwayat Barang Masuk
    const newTx: StockTransaction = {
      id: `TX-IN-${Date.now().toString().slice(-6)}`,
      timestamp: new Date().toISOString(),
      type: 'IN',
      itemId: newItemId,
      itemSku: data.sku,
      itemName: data.name,
      quantity: qty,
      unit: data.unit,
      previousStock: 0, // 0 karena barang baru pertama kali masuk!
      newStock: qty,
      supplierOrSource: data.distributor || 'Vendor / Distributor',
      receivedBy: data.receivedBy || 'Staff Logistik',
      invoiceOrPoNumber: data.invoiceOrPoNumber,
      photoUrl: data.photoUrl,
      geoTag: data.geoTag,
      driveFileLink: data.driveFileLink,
      merk: data.merk,
      typeModel: data.typeModel,
      serialNumber: data.serialNumber,
      procurementYear: data.procurementYear,
      conditionStatus: data.conditionStatus,
      fundingSource: data.fundingSource,
      distributor: data.distributor,
      aklAkd: data.aklAkd,
      notes: data.notes,
      syncedToGoogleSheets: false,
    };

    const updatedItems = [newItem, ...items];
    const updatedTransactions = [newTx, ...transactions];

    setItems(updatedItems);
    saveStoredItems(updatedItems);

    setTransactions(updatedTransactions);
    saveStoredTransactions(updatedTransactions);

    playSuccessSound();
    setIsMovementModalOpen(false);
    setMovementInitialSku(undefined);
    setMovementInitialData(undefined);

    showToast(
      `Barang Baru Masuk: +${qty} ${data.unit} "${data.name}" [${data.sku}] berhasil didaftarkan & disinkronkan ke database!`,
      'success'
    );

    // Save to Firestore Database in real-time
    saveItemToFirestore(newItem).catch(console.warn);
    saveTransactionToFirestore(newTx).catch(console.warn);

    // Real-Time Google Sheets Sync
    getAccessToken().then(token => {
      if (token && spreadsheetId) {
        appendTransactionRealtime(token, spreadsheetId, newTx, updatedItems)
          .then(success => {
            if (success) {
              setLastSyncedAt(new Date().toISOString());
            }
          })
          .catch(console.warn);
      }
    });

    // Offer to immediately print barcode label for the newly arrived item
    setConfirmDialog({
      isOpen: true,
      title: 'Cetak Stiker Barcode Barang Baru?',
      message: `Barang baru "${newItem.name}" (${newItem.sku}) telah berhasil didaftarkan ke inventaris kantor.\n\nApakah Anda ingin langsung mencetak stiker barcode / QR code untuk ditempel pada fisik barang sekarang?`,
      type: 'info',
      confirmText: 'Cetak Barcode Sekarang',
      onConfirm: () => {
        setConfirmDialog(prev => ({ ...prev, isOpen: false }));
        setBarcodeActiveItem(newItem);
        setIsBarcodeModalOpen(true);
      },
    });
  };

  // Barcode Scanner Action Handler
  const handleScannerSelectAction = (
    type: 'VIEW' | 'PRINT_BARCODE' | 'NEW_WITH_SKU',
    item?: InventoryItem,
    scannedCode?: string
  ) => {
    setIsScannerOpen(false);
    if (type === 'VIEW' && item) {
      setEditingItem(item);
      setIsItemModalOpen(true);
    } else if (type === 'PRINT_BARCODE' && item) {
      setBarcodeActiveItem(item);
      setIsBarcodeModalOpen(true);
    } else if (type === 'NEW_WITH_SKU' && scannedCode) {
      setMovementInitialSku(scannedCode);
      setMovementInitialData(undefined);
      setIsMovementModalOpen(true);
    }
  };

  // Upload Barcode to Google Drive
  const handleSaveBarcodeToDrive = async (filename: string, dataUrl: string) => {
    const token = await getAccessToken();
    if (!token) {
      throw new Error('Silakan hubungkan akun Google terlebih dahulu.');
    }

    const res = await fetch(dataUrl);
    const blob = await res.blob();

    const metadata = {
      name: filename,
      mimeType: 'image/png',
      description: 'Stiker Barcode Inventaris Kantor',
    };

    const form = new FormData();
    form.append('metadata', new Blob([JSON.stringify(metadata)], { type: 'application/json' }));
    form.append('file', blob);

    const uploadRes = await fetch(
      'https://www.googleapis.com/upload/drive/v3/files?uploadType=multipart',
      {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${token}`,
        },
        body: form,
      }
    );

    if (!uploadRes.ok) {
      const err = await uploadRes.text();
      throw new Error(`Gagal upload ke Drive: ${err}`);
    }
  };

  return (
    <div className="min-h-screen bg-slate-100/70 text-slate-800 flex flex-col font-sans selection:bg-indigo-500 selection:text-white">
      {/* Toast Notification Banner */}
      {toast && (
        <div className="fixed bottom-6 right-6 z-50 animate-in fade-in slide-in-from-bottom-5 duration-300 max-w-md">
          <div
            className={`p-4 rounded-2xl shadow-xl border flex items-center gap-3 ${
              toast.type === 'success'
                ? 'bg-emerald-900 text-white border-emerald-700'
                : toast.type === 'error'
                ? 'bg-rose-900 text-white border-rose-700'
                : 'bg-slate-900 text-white border-slate-700'
            }`}
          >
            {toast.type === 'success' ? (
              <CheckCircle2 className="w-5 h-5 text-emerald-400 shrink-0" />
            ) : toast.type === 'error' ? (
              <AlertCircle className="w-5 h-5 text-rose-400 shrink-0" />
            ) : (
              <Info className="w-5 h-5 text-blue-400 shrink-0" />
            )}
            <p className="text-xs sm:text-sm font-semibold flex-1 leading-snug">{toast.message}</p>
            <button
              onClick={() => setToast(null)}
              className="p-1 rounded-lg hover:bg-white/10 text-white/70 hover:text-white cursor-pointer"
            >
              <X className="w-4 h-4" />
            </button>
          </div>
        </div>
      )}

      {/* Main Top Navigation */}
      <Navbar
        activeTab={activeTab}
        setActiveTab={tab => {
          if (tab === 'movement-in') {
            setMovementInitialSku(undefined);
            setMovementInitialData(undefined);
            setIsMovementModalOpen(true);
          } else {
            setActiveTab(tab);
          }
        }}
        onOpenScanner={() => setIsScannerOpen(true)}
        onOpenPDFReport={() => setIsPDFModalOpen(true)}
        currentUser={currentUser}
        onOpenLoginModal={() => setIsLoginModalOpen(true)}
        onLogout={handleLogout}
        isLoggingIn={isLoggingIn}
        spreadsheetUrl={spreadsheetUrl}
        isSyncing={isSyncing}
        lowStockCount={lowStockCount}
      />

      {/* Database Status Strip */}
      <div className="bg-slate-900/95 border-b border-slate-800 py-1.5 px-4 sm:px-6 text-[11px] text-slate-400 flex items-center justify-between max-w-7xl mx-auto w-full shadow-xs">
        <div className="flex items-center gap-2">
          <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
          <span className="font-bold text-slate-200">Database Firestore:</span>
          <span>Tersinkronisasi ({items.length} Master Aset • {transactions.length} Riwayat Masuk)</span>
        </div>
        <div className="hidden sm:flex items-center gap-3">
          <span>Petugas: <strong className="text-emerald-400 font-bold">{currentUser?.name || 'Staf Inventaris'}</strong></span>
          <span>•</span>
          <span>Foto Geotag GPS: <strong className="text-emerald-400 font-bold">Aktif</strong></span>
        </div>
      </div>

      {/* Main View Area */}
      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-6">
        {activeTab === 'dashboard' && (
          <DashboardView
            items={items}
            transactions={transactions}
            categories={categories}
            onOpenScanner={() => setIsScannerOpen(true)}
            onOpenNewItem={() => {
              setMovementInitialSku(undefined);
              setMovementInitialData(undefined);
              setIsMovementModalOpen(true);
            }}
            onOpenMovementIn={(item) => {
              if (item) {
                setMovementInitialData({
                  name: item.name,
                  category: item.category,
                  location: item.location,
                  unit: item.unit,
                  merk: item.merk,
                  typeModel: item.typeModel,
                  distributor: item.distributor,
                  fundingSource: item.fundingSource,
                  pricePerUnit: item.pricePerUnit,
                  aklAkd: item.aklAkd,
                });
              } else {
                setMovementInitialData(undefined);
              }
              setMovementInitialSku(undefined);
              setIsMovementModalOpen(true);
            }}
            onNavigateTab={tab => {
              if (tab === 'movement-in') {
                setMovementInitialSku(undefined);
                setMovementInitialData(undefined);
                setIsMovementModalOpen(true);
              } else {
                setActiveTab(tab);
              }
            }}
            onOpenPDFReport={() => setIsPDFModalOpen(true)}
          />
        )}

        {activeTab === 'items' && (
          <InventoryList
            items={items}
            categories={categories}
            onOpenNewItem={() => {
              setMovementInitialSku(undefined);
              setMovementInitialData(undefined);
              setIsMovementModalOpen(true);
            }}
            onEditItem={item => {
              setEditingItem(item);
              setIsItemModalOpen(true);
            }}
            onDeleteItem={handleDeleteItem}
            onOpenBarcodeModal={item => {
              setBarcodeActiveItem(item);
              setIsBarcodeModalOpen(true);
            }}
            onOpenPDFReport={() => setIsPDFModalOpen(true)}
          />
        )}

        {activeTab === 'history' && (
          <TransactionHistoryView
            transactions={transactions}
            onSyncToGoogleSheets={confirmAndTriggerSync}
            isSyncing={isSyncing}
            spreadsheetUrl={spreadsheetUrl}
            onOpenPDFReport={() => setIsPDFModalOpen(true)}
          />
        )}

        {activeTab === 'barcode' && (
          <div className="space-y-4">
            <div className="bg-white p-6 sm:p-7 rounded-3xl border border-slate-200/90 shadow-xs flex flex-col sm:flex-row sm:items-center justify-between gap-4">
              <div>
                <div className="flex items-center gap-2">
                  <h2 className="text-xl sm:text-2xl font-black text-slate-900 tracking-tight">
                    Pusat Cetak & Generator Barcode Label
                  </h2>
                  <span className="px-2.5 py-0.5 rounded-full text-xs font-bold bg-indigo-50 text-indigo-700 border border-indigo-200">
                    Stiker 1D & QR Code
                  </span>
                </div>
                <p className="text-xs text-slate-500 mt-1">
                  Buat stiker barcode resmi untuk rak gudang, label aset kantor, atau cetak lembar batch format A4
                </p>
              </div>
              <button
                onClick={() => {
                  setBarcodeActiveItem(items[0] || null);
                  setIsBarcodeModalOpen(true);
                }}
                className="inline-flex items-center gap-2 px-5 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white font-black text-xs sm:text-sm rounded-xl shadow-md shadow-indigo-600/30 active:scale-95 transition-all cursor-pointer ring-1 ring-indigo-400/40"
              >
                <span>Buka Generator Label Stiker</span>
              </button>
            </div>

            {/* Grid of Items for Barcode Label Printing */}
            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
              {items.map(it => (
                <div
                  key={it.id}
                  onClick={() => {
                    setBarcodeActiveItem(it);
                    setIsBarcodeModalOpen(true);
                  }}
                  className="bg-white p-4 rounded-3xl border border-slate-200/90 hover:border-indigo-400 hover:shadow-md transition-all cursor-pointer flex flex-col justify-between group"
                >
                  <div>
                    {it.photoUrl ? (
                      <div className="w-full h-32 rounded-2xl overflow-hidden mb-3 bg-slate-100 border border-slate-200 relative">
                        <img src={it.photoUrl} alt={it.name} className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300" />
                        <span className="absolute top-2 left-2 font-mono text-[9px] font-bold bg-slate-900/80 text-white px-2 py-0.5 rounded-md backdrop-blur-xs">
                          {it.sku}
                        </span>
                      </div>
                    ) : (
                      <div className="w-full h-24 rounded-2xl bg-slate-50 border border-slate-200 flex items-center justify-center text-slate-400 mb-3">
                        <span className="font-mono text-xs font-bold bg-slate-100 text-slate-700 px-2 py-0.5 rounded">
                          {it.sku}
                        </span>
                      </div>
                    )}
                    <h4 className="font-extrabold text-slate-900 text-sm line-clamp-1 group-hover:text-indigo-600 transition-colors">
                      {it.name}
                    </h4>
                    <p className="text-xs text-slate-500 mt-0.5 truncate">{it.location}</p>
                    {it.serialNumber && (
                      <p className="text-[10px] font-mono text-slate-400 mt-0.5">SN: {it.serialNumber}</p>
                    )}
                  </div>
                  <div className="mt-3 pt-2.5 border-t border-slate-100 flex items-center justify-between text-xs">
                    <span className="text-[10px] text-slate-400 truncate max-w-[120px]">{it.category}</span>
                    <span className="font-bold text-indigo-600 group-hover:underline flex items-center gap-1">
                      <span>Cetak Label</span>
                      <span>→</span>
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {activeTab === 'sync' && (
          <GoogleSyncPanel
            user={user}
            onLogin={handleGoogleLogin}
            onLogout={handleLogout}
            isLoggingIn={isLoggingIn}
            spreadsheetId={spreadsheetId}
            spreadsheetUrl={spreadsheetUrl}
            lastSyncedAt={lastSyncedAt}
            isSyncing={isSyncing}
            onTriggerSync={confirmAndTriggerSync}
            itemsCount={items.length}
            transactionsCount={transactions.length}
          />
        )}
      </main>

      {/* Footer */}
      <footer className="border-t border-slate-200 bg-white py-4 mt-8">
        <div className="max-w-7xl mx-auto px-4 text-center text-xs text-slate-500 flex flex-col sm:flex-row items-center justify-between gap-2">
          <span>InventarisKantor • Sistem Barcode & Barang Masuk Terintegrasi Firestore Database & Google Workspace</span>
          <span>Status: Terhubung & Sinkron Real-Time</span>
        </div>
      </footer>

      {/* Add / Edit Item Modal */}
      <ItemModal
        isOpen={isItemModalOpen}
        itemToEdit={editingItem}
        itemsCount={items.length}
        categories={categories}
        onAddNewCategory={handleAddNewCategory}
        onClose={() => {
          setIsItemModalOpen(false);
          setEditingItem(null);
        }}
        onSave={handleSaveItem}
      />

      {/* Stock-In / Penerimaan Barang Baru Masuk Modal with Mandatory Photo & GPS Geotag */}
      <StockMovementModal
        isOpen={isMovementModalOpen}
        categories={categories}
        onAddNewCategory={handleAddNewCategory}
        itemsCount={items.length}
        initialSku={movementInitialSku}
        initialData={{
          ...movementInitialData,
          receivedBy: movementInitialData?.receivedBy || currentUser?.name || 'Ahmad Faisal (GA)',
        }}
        onClose={() => {
          setIsMovementModalOpen(false);
          setMovementInitialSku(undefined);
          setMovementInitialData(undefined);
        }}
        onSubmit={handleStockMovementSubmit}
      />

      {/* Barcode & QR Generator Modal */}
      <BarcodeGeneratorModal
        isOpen={isBarcodeModalOpen}
        item={barcodeActiveItem}
        allItems={items}
        onClose={() => {
          setIsBarcodeModalOpen(false);
          setBarcodeActiveItem(null);
        }}
        onSaveToDrive={handleSaveBarcodeToDrive}
        isDriveConnected={!!user}
      />

      {/* Live Barcode / QR Scanner Modal */}
      <BarcodeScannerModal
        isOpen={isScannerOpen}
        items={items}
        onClose={() => setIsScannerOpen(false)}
        onSelectAction={handleScannerSelectAction}
      />

      {/* PDF Monthly Report Modal */}
      <PDFReportModal
        isOpen={isPDFModalOpen}
        items={items}
        transactions={transactions}
        onClose={() => setIsPDFModalOpen(false)}
        onSuccess={(filename) => {
          playSuccessSound();
          showToast(`Laporan PDF "${filename}" berhasil diunduh!`, 'success');
        }}
      />

      {/* Login Modal for Office Staff & Google */}
      <LoginModal
        isOpen={isLoginModalOpen}
        currentUser={currentUser}
        onClose={() => setIsLoginModalOpen(false)}
        onGoogleLogin={handleGoogleLogin}
        isLoggingIn={isLoggingIn}
        onStaffLogin={(profile) => {
          setCurrentUser(profile);
          localStorage.setItem('inventaris_current_user', JSON.stringify(profile));
          showToast(`Berhasil masuk sebagai ${profile.name} (${profile.role})`, 'success');
          setIsLoginModalOpen(false);
        }}
      />

      {/* Confirmation Dialog */}
      <ConfirmDialog
        isOpen={confirmDialog.isOpen}
        title={confirmDialog.title}
        message={confirmDialog.message}
        type={confirmDialog.type}
        confirmText={confirmDialog.confirmText}
        onConfirm={confirmDialog.onConfirm}
        onCancel={() => setConfirmDialog(prev => ({ ...prev, isOpen: false }))}
      />
    </div>
  );
}
