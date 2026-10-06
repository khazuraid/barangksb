import { InventoryItem, StockTransaction, DEFAULT_CATEGORIES } from '../types/inventory';

const STORAGE_KEYS = {
  ITEMS: 'inventaris_kantor_items_v3',
  TRANSACTIONS: 'inventaris_kantor_tx_v3',
  CATEGORIES: 'inventaris_kantor_categories_v3',
};

const INITIAL_ITEMS: InventoryItem[] = [
  {
    id: 'item-1',
    sku: 'MED-2024-001',
    name: 'USG Diagnostic Scanner B/W',
    category: 'Peralatan Medis & Alkes (AKL/AKD)',
    location: 'Ruang Poli Medis Lt 1',
    currentStock: 2,
    minStock: 1,
    unit: 'Unit',
    pricePerUnit: 48500000,
    description: 'Unit USG dengan transduser convex, kondisi siap pakai',
    photoUrl: 'https://images.unsplash.com/photo-1516549655169-df83a0774514?w=400&auto=format&fit=crop&q=80',
    geoTag: {
      latitude: -6.208763,
      longitude: 106.845599,
      accuracy: 6,
      timestamp: new Date(Date.now() - 86400000 * 2).toISOString(),
      locationName: 'Kantor Pusat - Gedung Logistik A',
    },
    driveFileLink: 'https://drive.google.com/drive/folders/inventaris-medis',
    merk: 'Mindray',
    typeModel: 'DP-10',
    serialNumber: 'SN-MND-984210',
    procurementYear: 2024,
    conditionStatus: 'Berfungsi',
    fundingSource: 'DAK Kesehatan',
    distributor: 'PT Saba Indomedika',
    aklAkd: 'AKL 21102910482',
    isAvailable: true,
    barcodeFormat: 'CODE128',
    createdAt: new Date(Date.now() - 86400000 * 14).toISOString(),
    updatedAt: new Date(Date.now() - 86400000 * 1).toISOString(),
  },
  {
    id: 'item-2',
    sku: 'IT-2024-002',
    name: 'Printer All-In-One EcoTank',
    category: 'Elektronik & IT Perkantoran',
    location: 'Ruang Tata Usaha (TU)',
    currentStock: 3,
    minStock: 1,
    unit: 'Unit',
    pricePerUnit: 2450000,
    description: 'Printer print, scan, copy dengan tangki tinta original',
    photoUrl: 'https://images.unsplash.com/photo-1612815154858-60aa4c59eaa6?w=400&auto=format&fit=crop&q=80',
    geoTag: {
      latitude: -6.209112,
      longitude: 106.846201,
      accuracy: 8,
      timestamp: new Date(Date.now() - 86400000 * 1).toISOString(),
      locationName: 'Gedung Administrasi - Lt 2',
    },
    driveFileLink: 'https://drive.google.com/drive/folders/inventaris-it',
    merk: 'Epson',
    typeModel: 'L3210 EcoTank',
    serialNumber: 'EPS-X889021',
    procurementYear: 2024,
    conditionStatus: 'Berfungsi',
    fundingSource: 'Kas Operasional Kantor',
    distributor: 'PT Metrodata Electronics',
    aklAkd: 'Non-AKL (Peralatan Umum)',
    isAvailable: true,
    barcodeFormat: 'CODE128',
    createdAt: new Date(Date.now() - 86400000 * 12).toISOString(),
    updatedAt: new Date(Date.now() - 86400000 * 2).toISOString(),
  },
  {
    id: 'item-3',
    sku: 'MED-2024-003',
    name: 'Tensimeter Digital Lengan Standar RS',
    category: 'Peralatan Medis & Alkes (AKL/AKD)',
    location: 'Ruang Pemeriksaan A',
    currentStock: 5,
    minStock: 2,
    unit: 'Unit',
    pricePerUnit: 1150000,
    description: 'Tensimeter otomatis layar LCD dengan cuff ukuran M dan L',
    photoUrl: 'https://images.unsplash.com/photo-1584308666744-24d5c474f2ae?w=400&auto=format&fit=crop&q=80',
    geoTag: {
      latitude: -6.208544,
      longitude: 106.845112,
      accuracy: 5,
      timestamp: new Date(Date.now() - 86400000 * 3).toISOString(),
      locationName: 'Poli Klinik Lt 1',
    },
    driveFileLink: 'https://drive.google.com/drive/folders/inventaris-medis',
    merk: 'Omron',
    typeModel: 'HEM-7156',
    serialNumber: 'OMR-SN-339102',
    procurementYear: 2023,
    conditionStatus: 'Berfungsi',
    fundingSource: 'APBD',
    distributor: 'PT Kimia Farma Trading',
    aklAkd: 'AKL 20501021489',
    isAvailable: true,
    barcodeFormat: 'CODE128',
    createdAt: new Date(Date.now() - 86400000 * 10).toISOString(),
    updatedAt: new Date(Date.now() - 86400000 * 3).toISOString(),
  },
  {
    id: 'item-4',
    sku: 'IT-2024-004',
    name: 'Laptop Staff Administrasi 14 Inch',
    category: 'Elektronik & IT Perkantoran',
    location: 'Ruang Keuangan',
    currentStock: 4,
    minStock: 1,
    unit: 'Unit',
    pricePerUnit: 9800000,
    description: 'Core i5 16GB RAM 512GB SSD Windows 11 Pro',
    photoUrl: 'https://images.unsplash.com/photo-1541807084-5c52b6b3adef?w=400&auto=format&fit=crop&q=80',
    geoTag: {
      latitude: -6.208990,
      longitude: 106.845880,
      accuracy: 7,
      timestamp: new Date(Date.now() - 86400000 * 1).toISOString(),
      locationName: 'Divisi Keuangan & Anggaran',
    },
    driveFileLink: 'https://drive.google.com/drive/folders/inventaris-it',
    merk: 'Lenovo',
    typeModel: 'ThinkPad E14 Gen 4',
    serialNumber: 'PF418892',
    procurementYear: 2024,
    conditionStatus: 'Berfungsi',
    fundingSource: 'APBN',
    distributor: 'PT Synnex Metrodata',
    aklAkd: 'Non-AKL (IT Aset)',
    isAvailable: true,
    barcodeFormat: 'CODE128',
    createdAt: new Date(Date.now() - 86400000 * 8).toISOString(),
    updatedAt: new Date(Date.now() - 86400000 * 1).toISOString(),
  },
  {
    id: 'item-5',
    sku: 'FRN-2024-005',
    name: 'Kursi Kerja Putar Hidrolik Staf',
    category: 'Furnitur & Perlengkapan Ruangan',
    location: 'Ruang Administrasi Utama',
    currentStock: 12,
    minStock: 3,
    unit: 'Unit',
    pricePerUnit: 1450000,
    description: 'Bahan mesh jaring breathable dengan roda nilon',
    photoUrl: 'https://images.unsplash.com/photo-1580481077195-c3a9f0222950?w=400&auto=format&fit=crop&q=80',
    geoTag: {
      latitude: -6.208763,
      longitude: 106.845599,
      accuracy: 10,
      timestamp: new Date(Date.now() - 86400000 * 4).toISOString(),
      locationName: 'Ruang Staf Lt 2',
    },
    driveFileLink: 'https://drive.google.com/drive/folders/inventaris-furnitur',
    merk: 'Indachi',
    typeModel: 'D-300 Mesh',
    serialNumber: 'IND-2024-481',
    procurementYear: 2024,
    conditionStatus: 'Berfungsi',
    fundingSource: 'Kas Operasional Kantor',
    distributor: 'CV Furnitur Prima',
    aklAkd: 'Non-AKL',
    isAvailable: true,
    barcodeFormat: 'CODE128',
    createdAt: new Date(Date.now() - 86400000 * 7).toISOString(),
    updatedAt: new Date(Date.now() - 86400000 * 1).toISOString(),
  },
];

const INITIAL_TRANSACTIONS: StockTransaction[] = [
  {
    id: 'TX-IN-901',
    timestamp: new Date(Date.now() - 86400000 * 2).toISOString(),
    type: 'IN',
    itemId: 'item-1',
    itemSku: 'MED-2024-001',
    itemName: 'USG Diagnostic Scanner B/W',
    quantity: 1,
    unit: 'Unit',
    previousStock: 1,
    newStock: 2,
    supplierOrSource: 'PT Saba Indomedika',
    receivedBy: 'Ahmad Faisal (GA/Logistik)',
    invoiceOrPoNumber: 'PO-MED-2024-082',
    photoUrl: 'https://images.unsplash.com/photo-1516549655169-df83a0774514?w=400&auto=format&fit=crop&q=80',
    geoTag: {
      latitude: -6.208763,
      longitude: 106.845599,
      accuracy: 6,
      timestamp: new Date(Date.now() - 86400000 * 2).toISOString(),
      locationName: 'Kantor Pusat - Gedung Logistik A',
    },
    driveFileLink: 'https://drive.google.com/drive/folders/inventaris-medis',
    merk: 'Mindray',
    typeModel: 'DP-10',
    serialNumber: 'SN-MND-984210',
    procurementYear: 2024,
    conditionStatus: 'Berfungsi',
    fundingSource: 'DAK Kesehatan',
    distributor: 'PT Saba Indomedika',
    aklAkd: 'AKL 21102910482',
    notes: 'Penerimaan alat baru lengkap dengan probe convex & manual book',
  },
  {
    id: 'TX-IN-902',
    timestamp: new Date(Date.now() - 86400000 * 1).toISOString(),
    type: 'IN',
    itemId: 'item-2',
    itemSku: 'IT-2024-002',
    itemName: 'Printer All-In-One EcoTank',
    quantity: 2,
    unit: 'Unit',
    previousStock: 1,
    newStock: 3,
    supplierOrSource: 'PT Metrodata Electronics',
    receivedBy: 'Rian Hidayat (IT)',
    invoiceOrPoNumber: 'SJ-METRO-99124',
    photoUrl: 'https://images.unsplash.com/photo-1612815154858-60aa4c59eaa6?w=400&auto=format&fit=crop&q=80',
    geoTag: {
      latitude: -6.209112,
      longitude: 106.846201,
      accuracy: 8,
      timestamp: new Date(Date.now() - 86400000 * 1).toISOString(),
      locationName: 'Gedung Administrasi - Lt 2',
    },
    driveFileLink: 'https://drive.google.com/drive/folders/inventaris-it',
    merk: 'Epson',
    typeModel: 'L3210 EcoTank',
    serialNumber: 'EPS-X889021',
    procurementYear: 2024,
    conditionStatus: 'Berfungsi',
    fundingSource: 'Kas Operasional Kantor',
    distributor: 'PT Metrodata Electronics',
    aklAkd: 'Non-AKL',
    notes: 'Kondisi kardus bersegel, garansi resmi 2 tahun',
  },
];

export const getStoredItems = (): InventoryItem[] => {
  try {
    const raw = localStorage.getItem(STORAGE_KEYS.ITEMS);
    if (!raw) {
      localStorage.setItem(STORAGE_KEYS.ITEMS, JSON.stringify(INITIAL_ITEMS));
      return INITIAL_ITEMS;
    }
    return JSON.parse(raw);
  } catch {
    return INITIAL_ITEMS;
  }
};

export const saveStoredItems = (items: InventoryItem[]): void => {
  try {
    localStorage.setItem(STORAGE_KEYS.ITEMS, JSON.stringify(items));
  } catch (err) {
    console.warn('LocalStorage saveItems fallback:', err);
  }
};

export const getStoredTransactions = (): StockTransaction[] => {
  try {
    const raw = localStorage.getItem(STORAGE_KEYS.TRANSACTIONS);
    if (!raw) {
      localStorage.setItem(STORAGE_KEYS.TRANSACTIONS, JSON.stringify(INITIAL_TRANSACTIONS));
      return INITIAL_TRANSACTIONS;
    }
    return JSON.parse(raw);
  } catch {
    return INITIAL_TRANSACTIONS;
  }
};

export const saveStoredTransactions = (transactions: StockTransaction[]): void => {
  try {
    localStorage.setItem(STORAGE_KEYS.TRANSACTIONS, JSON.stringify(transactions));
  } catch (err) {
    console.warn('LocalStorage saveTransactions fallback:', err);
  }
};

export const getStoredCategories = (): string[] => {
  try {
    const raw = localStorage.getItem(STORAGE_KEYS.CATEGORIES);
    if (!raw) {
      localStorage.setItem(STORAGE_KEYS.CATEGORIES, JSON.stringify(DEFAULT_CATEGORIES));
      return DEFAULT_CATEGORIES;
    }
    return JSON.parse(raw);
  } catch {
    return DEFAULT_CATEGORIES;
  }
};

export const saveStoredCategories = (categories: string[]): void => {
  try {
    localStorage.setItem(STORAGE_KEYS.CATEGORIES, JSON.stringify(categories));
  } catch (err) {
    console.warn('LocalStorage saveCategories fallback:', err);
  }
};

export const generateSKU = (category: string, currentCount: number): string => {
  let prefix = 'BRG';
  const c = category.toLowerCase();
  if (c.includes('medis') || c.includes('alkes') || c.includes('akl')) prefix = 'MED';
  else if (c.includes('elektronik') || c.includes('it')) prefix = 'IT';
  else if (c.includes('furnitur') || c.includes('ruang')) prefix = 'FRN';
  else if (c.includes('atk') || c.includes('tulis')) prefix = 'ATK';
  else if (c.includes('lab') || c.includes('diagnostik')) prefix = 'LAB';
  else if (c.includes('pantri') || c.includes('umum')) prefix = 'PAN';
  else if (c.includes('kebersihan')) prefix = 'KBR';
  else if (c.includes('k3') || c.includes('aman')) prefix = 'K3';

  const year = new Date().getFullYear();
  const randomSuffix = Math.floor(100 + Math.random() * 900);
  const seq = String(currentCount + 1).padStart(3, '0');
  return `${prefix}-${year}-${seq}${randomSuffix % 10}`;
};
