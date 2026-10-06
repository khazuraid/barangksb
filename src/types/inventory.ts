export const DEFAULT_CATEGORIES: string[] = [
  'Peralatan Medis & Alkes (AKL/AKD)',
  'Elektronik & IT Perkantoran',
  'Furnitur & Perlengkapan Ruangan',
  'ATK (Alat Tulis Kantor)',
  'Alat Laboratorium & Diagnostik',
  'Pantri & Fasilitas Umum',
  'Kebersihan & Sanitasi',
  'Keamanan & K3',
  'Lainnya',
];

export type ItemCategory = string;

export type ConditionStatus = 'Berfungsi' | 'Rusak Ringan' | 'Rusak Berat' | 'Perlu Kalibrasi';

export type UnitType = 'Unit' | 'Pcs' | 'Set' | 'Box' | 'Rim' | 'Pack' | 'Dus' | 'Botol' | 'Roll';

export interface OfficeUserProfile {
  id: string;
  name: string;
  email: string;
  role: string;
  avatar?: string;
  isGoogleConnected?: boolean;
}

export interface GeoTagData {
  latitude: number;
  longitude: number;
  accuracy?: number;
  timestamp: string;
  locationName?: string;
}

export interface InventoryItem {
  id: string;
  sku: string; // Barcode / SKU code
  name: string; // Nama Barang
  category: ItemCategory; // Kategori
  location: string; // Lokasi Simpan / Ruangan
  currentStock: number; // Stok Saat Ini
  minStock: number; // Batas Minimum Alert
  unit: UnitType; // Satuan
  pricePerUnit?: number; // Harga (Nilai Perolehan Rupiah)
  description?: string; // Keterangan
  photoUrl?: string; // Foto Fisik Barang Masuk
  geoTag?: GeoTagData; // Koordinat Geotagging & Timestamp Foto
  driveFileLink?: string; // Tautan File di Google Drive

  // Atribut Standar Sesuai Format Kolom Pengadaan:
  merk?: string; // Merk
  typeModel?: string; // Type / Model
  serialNumber?: string; // No Seri
  procurementYear?: number | string; // Thn Pengadaan
  conditionStatus?: ConditionStatus; // Berfungsi (Berfungsi / Rusak / dll)
  fundingSource?: string; // Pendanaan (e.g. APBN, APBD, DAK, Kas Kantor, Operasional)
  distributor?: string; // Distributor
  aklAkd?: string; // AKL/AKD (Nomor Izin Edar / Kategori Regulasi)
  isAvailable?: boolean; // Ada (Status Keberadaan Fisik)

  barcodeFormat?: 'CODE128' | 'QR' | 'EAN13';
  createdAt: string;
  updatedAt: string;
}

export type MovementType = 'IN';

export interface NewIncomingItemData {
  sku: string;
  name: string;
  category: ItemCategory;
  location: string;
  quantity: number; // Jumlah fisik barang baru yang masuk
  minStock: number;
  unit: UnitType;
  pricePerUnit: number;
  description?: string;
  photoUrl?: string;
  geoTag?: GeoTagData;
  driveFileLink?: string;

  // Atribut Pengadaan Resmi
  merk?: string;
  typeModel?: string;
  serialNumber?: string;
  procurementYear?: number | string;
  conditionStatus?: ConditionStatus;
  fundingSource?: string;
  distributor?: string;
  aklAkd?: string;
  isAvailable?: boolean;

  // Dokumen Penerimaan Barang Baru
  receivedBy?: string;
  invoiceOrPoNumber?: string;
  notes?: string;
}

export interface StockTransaction {
  id: string;
  timestamp: string; // ISO string
  type: 'IN'; // Hanya barang masuk
  itemId: string;
  itemSku: string;
  itemName: string;
  quantity: number;
  unit: UnitType;
  previousStock: number;
  newStock: number;
  
  // Specific to Barang Masuk (IN) & Kolom Standar
  supplierOrSource?: string; // Distributor / Vendor
  receivedBy?: string; // Petugas penerima barang
  invoiceOrPoNumber?: string; // No PO / Surat Jalan / Faktur
  photoUrl?: string; // Foto bukti fisik barang masuk / nota
  geoTag?: GeoTagData; // Data Geotagging Foto Bukti Fisik
  driveFileLink?: string; // Tautan File di Google Drive

  // Atribut barang yang masuk
  merk?: string; // Merk
  typeModel?: string; // Type
  serialNumber?: string; // No seri
  procurementYear?: number | string; // Thn Pengadaan
  conditionStatus?: ConditionStatus; // Berfungsi
  fundingSource?: string; // Pendanaan
  distributor?: string; // Distributor
  aklAkd?: string; // AKL/AKD

  notes?: string; // Keterangan
  syncedToGoogleSheets?: boolean;
}

export interface GoogleSheetsConfig {
  spreadsheetId: string | null;
  spreadsheetUrl: string | null;
  spreadsheetName: string;
  lastSyncedAt: string | null;
  autoSync: boolean;
}
