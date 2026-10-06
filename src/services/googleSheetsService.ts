import { InventoryItem, StockTransaction } from '../types/inventory';

export interface SpreadsheetInfo {
  id: string;
  name: string;
  webViewLink: string;
}

const DEFAULT_SPREADSHEET_TITLE = 'Inventaris Kantor & Log Barang Masuk';

const MASTER_HEADERS = [
  'Ada (Keberadaan)',
  'No Seri',
  'Nama Barang',
  'Kode Barcode / SKU',
  'Merk',
  'Type / Model',
  'Thn Pengadaan',
  'Kondisi (Berfungsi)',
  'Kategori',
  'Lokasi Simpan / Ruangan',
  'Jumlah Stok',
  'Satuan',
  'Nilai Satuan (Rp)',
  'Sumber Pendanaan',
  'Distributor / Vendor',
  'AKL / AKD',
  'Foto Barang',
  'Geotagging GPS',
  'Keterangan',
  'Terakhir Diperbarui'
];

const LOG_IN_HEADERS = [
  'ID Transaksi',
  'Tanggal & Waktu (WIB)',
  'Kode Barcode / SKU',
  'Nama Barang',
  'Jumlah Masuk (+)',
  'Satuan',
  'Merk',
  'Type / Model',
  'No Seri',
  'Thn Pengadaan',
  'Kondisi',
  'Sumber Pendanaan',
  'Distributor / Vendor',
  'Petugas Penerima',
  'No PO / Surat Jalan',
  'Foto Bukti Fisik',
  'Koordinat Geotag GPS',
  'Catatan / Keterangan'
];

/**
 * Find existing spreadsheet in Drive or create a new structured spreadsheet
 */
export async function findOrCreateSpreadsheet(accessToken: string): Promise<SpreadsheetInfo> {
  const query = encodeURIComponent(`name = '${DEFAULT_SPREADSHEET_TITLE}' and mimeType = 'application/vnd.google-apps.spreadsheet' and trashed = false`);
  const searchRes = await fetch(
    `https://www.googleapis.com/drive/v3/files?q=${query}&fields=files(id,name,webViewLink)&pageSize=1`,
    {
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    }
  );

  if (searchRes.ok) {
    const data = await searchRes.json();
    if (data.files && data.files.length > 0) {
      const file = data.files[0];
      return {
        id: file.id,
        name: file.name,
        webViewLink: file.webViewLink || `https://docs.google.com/spreadsheets/d/${file.id}/edit`,
      };
    }
  }

  // Create new spreadsheet
  const createRes = await fetch('https://sheets.googleapis.com/v4/spreadsheets', {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${accessToken}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      properties: {
        title: DEFAULT_SPREADSHEET_TITLE,
      },
      sheets: [
        {
          properties: {
            title: 'Master_Barang',
            gridProperties: { rowCount: 150, columnCount: 22, frozenRowCount: 1 },
          },
        },
        {
          properties: {
            title: 'Log_Barang_Masuk',
            gridProperties: { rowCount: 300, columnCount: 20, frozenRowCount: 1 },
          },
        },
      ],
    }),
  });

  if (!createRes.ok) {
    const errText = await createRes.text();
    throw new Error(`Gagal membuat spreadsheet Google: ${errText}`);
  }

  const createdData = await createRes.json();
  const spreadsheetId = createdData.spreadsheetId;
  const webViewLink = createdData.spreadsheetUrl || `https://docs.google.com/spreadsheets/d/${spreadsheetId}/edit`;

  // Initialize header rows with full column widths
  await updateSheetValues(accessToken, spreadsheetId, 'Master_Barang!A1:T1', [MASTER_HEADERS]);
  await updateSheetValues(accessToken, spreadsheetId, 'Log_Barang_Masuk!A1:R1', [LOG_IN_HEADERS]);

  return {
    id: spreadsheetId,
    name: DEFAULT_SPREADSHEET_TITLE,
    webViewLink,
  };
}

async function updateSheetValues(accessToken: string, spreadsheetId: string, range: string, values: any[][]) {
  const url = `https://sheets.googleapis.com/v4/spreadsheets/${spreadsheetId}/values/${encodeURIComponent(range)}?valueInputOption=USER_ENTERED`;
  const res = await fetch(url, {
    method: 'PUT',
    headers: {
      Authorization: `Bearer ${accessToken}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ values }),
  });

  if (!res.ok) {
    const err = await res.text();
    console.error('Error updating values:', err);
    throw new Error(`Gagal memperbarui sheet data: ${err}`);
  }
  return res.json();
}

async function appendSheetValues(accessToken: string, spreadsheetId: string, range: string, values: any[][]) {
  const url = `https://sheets.googleapis.com/v4/spreadsheets/${spreadsheetId}/values/${encodeURIComponent(range)}:append?valueInputOption=USER_ENTERED&insertDataOption=INSERT_ROWS`;
  const res = await fetch(url, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${accessToken}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ values }),
  });

  if (!res.ok) {
    const err = await res.text();
    console.error('Error appending values:', err);
    throw new Error(`Gagal menambahkan baris ke sheet: ${err}`);
  }
  return res.json();
}

function formatDate(iso: string) {
  try {
    const d = new Date(iso);
    return d.toLocaleString('id-ID', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  } catch {
    return iso;
  }
}

export async function syncAllToGoogleSheets(
  accessToken: string,
  spreadsheetId: string,
  items: InventoryItem[],
  transactions: StockTransaction[]
) {
  // 1. Prepare Master_Barang rows
  const masterRows: any[][] = [
    MASTER_HEADERS,
    ...items.map(item => [
      item.isAvailable !== false ? 'ADA' : 'TIDAK',
      item.serialNumber || '-',
      item.name,
      item.sku,
      item.merk || '-',
      item.typeModel || '-',
      item.procurementYear || '-',
      item.conditionStatus || 'Berfungsi',
      item.category,
      item.location,
      item.currentStock,
      item.unit,
      item.pricePerUnit || 0,
      item.fundingSource || '-',
      item.distributor || '-',
      item.aklAkd || '-',
      item.photoUrl ? (item.photoUrl.startsWith('data:') ? '[Foto Terlampir]' : item.photoUrl) : '-',
      item.geoTag ? `${item.geoTag.latitude.toFixed(6)}, ${item.geoTag.longitude.toFixed(6)}` : '-',
      item.description || '-',
      formatDate(item.updatedAt || item.createdAt),
    ]),
  ];
  await updateSheetValues(accessToken, spreadsheetId, 'Master_Barang!A1:T' + Math.max(items.length + 1, 50), masterRows);

  // 2. Prepare Log_Barang_Masuk rows
  const inTx = transactions.filter(t => t.type === 'IN');
  const inRows: any[][] = [
    LOG_IN_HEADERS,
    ...inTx.map(t => [
      t.id,
      formatDate(t.timestamp),
      t.itemSku,
      t.itemName,
      t.quantity,
      t.unit,
      t.merk || '-',
      t.typeModel || '-',
      t.serialNumber || '-',
      t.procurementYear || '-',
      t.conditionStatus || 'Berfungsi',
      t.fundingSource || '-',
      t.distributor || t.supplierOrSource || '-',
      t.receivedBy || '-',
      t.invoiceOrPoNumber || '-',
      t.photoUrl ? (t.photoUrl.startsWith('data:') ? '[Foto Bukti Fisik Terlampir]' : t.photoUrl) : '-',
      t.geoTag ? `${t.geoTag.latitude.toFixed(6)}, ${t.geoTag.longitude.toFixed(6)}` : '-',
      t.notes || '-',
    ]),
  ];
  await updateSheetValues(accessToken, spreadsheetId, 'Log_Barang_Masuk!A1:R' + Math.max(inTx.length + 1, 50), inRows);

  return true;
}

export async function appendTransactionRealtime(
  accessToken: string,
  spreadsheetId: string,
  transaction: StockTransaction,
  allItems: InventoryItem[]
) {
  try {
    const row = [
      transaction.id,
      formatDate(transaction.timestamp),
      transaction.itemSku,
      transaction.itemName,
      transaction.quantity,
      transaction.unit,
      transaction.merk || '-',
      transaction.typeModel || '-',
      transaction.serialNumber || '-',
      transaction.procurementYear || '-',
      transaction.conditionStatus || 'Berfungsi',
      transaction.fundingSource || '-',
      transaction.distributor || transaction.supplierOrSource || '-',
      transaction.receivedBy || '-',
      transaction.invoiceOrPoNumber || '-',
      transaction.photoUrl ? (transaction.photoUrl.startsWith('data:') ? '[Foto Bukti Fisik Terlampir]' : transaction.photoUrl) : '-',
      transaction.geoTag ? `${transaction.geoTag.latitude.toFixed(6)}, ${transaction.geoTag.longitude.toFixed(6)}` : '-',
      transaction.notes || '-',
    ];
    await appendSheetValues(accessToken, spreadsheetId, 'Log_Barang_Masuk!A:R', [row]);

    // Refresh Master Barang sheet so stock numbers stay 100% updated in real-time
    const masterRows: any[][] = [
      MASTER_HEADERS,
      ...allItems.map(item => [
        item.isAvailable !== false ? 'ADA' : 'TIDAK',
        item.serialNumber || '-',
        item.name,
        item.sku,
        item.merk || '-',
        item.typeModel || '-',
        item.procurementYear || '-',
        item.conditionStatus || 'Berfungsi',
        item.category,
        item.location,
        item.currentStock,
        item.unit,
        item.pricePerUnit || 0,
        item.fundingSource || '-',
        item.distributor || '-',
        item.aklAkd || '-',
        item.photoUrl ? (item.photoUrl.startsWith('data:') ? '[Foto Terlampir]' : item.photoUrl) : '-',
        item.geoTag ? `${item.geoTag.latitude.toFixed(6)}, ${item.geoTag.longitude.toFixed(6)}` : '-',
        item.description || '-',
        formatDate(item.updatedAt || item.createdAt),
      ]),
    ];
    await updateSheetValues(accessToken, spreadsheetId, 'Master_Barang!A1:T' + Math.max(allItems.length + 1, 20), masterRows);
    return true;
  } catch (err) {
    console.error('Realtime sync append failed:', err);
    return false;
  }
}
