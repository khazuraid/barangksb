import jsPDF from 'jspdf';
import autoTable from 'jspdf-autotable';
import { InventoryItem, StockTransaction } from '../types/inventory';

interface PDFReportOptions {
  companyName?: string;
  monthYear?: string;
  preparedBy?: string;
  approvedBy?: string;
  includeTransactions?: boolean;
}

export const generateInventoryPDF = (
  items: InventoryItem[],
  transactions: StockTransaction[],
  options: PDFReportOptions = {}
) => {
  const doc = new jsPDF({
    orientation: 'portrait',
    unit: 'mm',
    format: 'a4',
  });

  const companyName = options.companyName || 'PT INVENTARIS NUSANTARA';
  const monthYear = options.monthYear || new Date().toLocaleDateString('id-ID', { month: 'long', year: 'numeric' });
  const preparedBy = options.preparedBy || 'Ahmad Faisal (Staff GA & Inventaris)';
  const approvedBy = options.approvedBy || 'Budi Santoso, S.E. (Head of Finance & Operations)';
  const printDate = new Date().toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' });

  // Compute metrics
  const totalItems = items.length;
  const totalStockUnits = items.reduce((acc, i) => acc + i.currentStock, 0);
  const totalAssetValue = items.reduce((acc, i) => acc + ((i.pricePerUnit || 0) * i.currentStock), 0);
  const lowStockCount = items.filter(i => i.currentStock <= i.minStock).length;

  const totalIn = transactions.reduce((acc, t) => acc + t.quantity, 0);

  // 1. Header / Kop Surat Formal
  doc.setFillColor(30, 41, 59); // slate-800
  doc.rect(0, 0, 210, 8, 'F');

  doc.setFont('helvetica', 'bold');
  doc.setFontSize(16);
  doc.setTextColor(15, 23, 42); // slate-900
  doc.text(companyName.toUpperCase(), 14, 20);

  doc.setFont('helvetica', 'normal');
  doc.setFontSize(9);
  doc.setTextColor(100, 116, 139); // slate-500
  doc.text('Divisi General Affair, Pengadaan & Manajemen Fasilitas Kantor', 14, 25);
  doc.text('Laporan Resmi Pengawasan Stok & Mutasi Logistik Perkantoran', 14, 29);

  // Divider line
  doc.setDrawColor(203, 213, 225); // slate-300
  doc.setLineWidth(0.6);
  doc.line(14, 32, 196, 32);

  // Title Box
  doc.setFont('helvetica', 'bold');
  doc.setFontSize(13);
  doc.setTextColor(30, 41, 59);
  doc.text(`LAPORAN BULANAN INVENTARIS & MUTASI STOK BARANG`, 14, 40);

  doc.setFont('helvetica', 'normal');
  doc.setFontSize(9);
  doc.setTextColor(71, 85, 105);
  doc.text(`Periode: ${monthYear}  |  Tanggal Cetak Dokumen: ${printDate}`, 14, 45);

  // 2. Summary KPI Box
  doc.setFillColor(248, 250, 252); // slate-50
  doc.setDrawColor(226, 232, 240); // slate-200
  doc.roundedRect(14, 49, 182, 22, 3, 3, 'FD');

  doc.setFont('helvetica', 'bold');
  doc.setFontSize(8);
  doc.setTextColor(100, 116, 139);
  doc.text('TOTAL ITEM KATALOG', 20, 56);
  doc.text('TOTAL UNIT FISIK', 62, 56);
  doc.text('TOTAL BARANG MASUK', 106, 56);
  doc.text('ESTIMASI NILAI ASET', 152, 56);

  doc.setFont('helvetica', 'bold');
  doc.setFontSize(11);
  doc.setTextColor(15, 23, 42);
  doc.text(`${totalItems} Jenis`, 20, 64);
  doc.text(`${totalStockUnits.toLocaleString('id-ID')} Unit`, 62, 64);
  doc.text(`+${totalIn.toLocaleString('id-ID')} Unit`, 106, 64);

  const formattedValue = new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(totalAssetValue);
  doc.text(formattedValue, 152, 64);

  // 3. Table 1: Master Stok Barang
  doc.setFont('helvetica', 'bold');
  doc.setFontSize(10);
  doc.setTextColor(30, 41, 59);
  doc.text('1. DAFTAR MASTER INVENTARIS ASET & BARANG KANTOR', 14, 78);

  const itemRows = items.map((item, idx) => {
    const isAvail = item.isAvailable !== false ? 'ADA' : 'TIDAK';
    const merkModel = [item.merk, item.typeModel].filter(Boolean).join(' / ') || '-';

    return [
      (idx + 1).toString(),
      isAvail,
      item.serialNumber || '-',
      item.name,
      item.sku,
      merkModel,
      (item.procurementYear || '-').toString(),
      item.conditionStatus || 'Berfungsi',
      item.category,
      `${item.currentStock} ${item.unit}`,
      item.fundingSource || '-',
    ];
  });

  autoTable(doc, {
    startY: 81,
    head: [['No', 'Ada', 'No Seri', 'Nama Barang', 'Kode SKU', 'Merk / Type', 'Thn', 'Kondisi', 'Kategori', 'Stok', 'Pendanaan']],
    body: itemRows,
    theme: 'grid',
    headStyles: {
      fillColor: [30, 41, 59],
      textColor: [255, 255, 255],
      fontSize: 7.5,
      fontStyle: 'bold',
      halign: 'center',
    },
    styles: {
      fontSize: 7,
      cellPadding: 1.8,
      textColor: [30, 41, 59],
    },
    columnStyles: {
      0: { halign: 'center', cellWidth: 7 },
      1: { halign: 'center', fontStyle: 'bold', cellWidth: 10 },
      2: { fontStyle: 'bold', cellWidth: 20 },
      3: { cellWidth: 32 },
      4: { fontStyle: 'bold', cellWidth: 20 },
      5: { cellWidth: 22 },
      6: { halign: 'center', cellWidth: 11 },
      7: { halign: 'center', cellWidth: 16 },
      8: { cellWidth: 24 },
      9: { halign: 'center', fontStyle: 'bold', cellWidth: 14 },
      10: { cellWidth: 18 },
    },
    margin: { left: 14, right: 14 },
  });

  // 4. Table 2: Log Penerimaan Barang Baru Masuk
  const lastTableY = (doc as any).lastAutoTable?.finalY || 160;

  let currentY = lastTableY + 10;
  if (currentY > 230) {
    doc.addPage();
    currentY = 20;
  }

  doc.setFont('helvetica', 'bold');
  doc.setFontSize(10);
  doc.setTextColor(30, 41, 59);
  doc.text('2. LOG RIWAYAT PENERIMAAN BARANG BARU MASUK', 14, currentY);

  const recentTx = [...transactions]
    .sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
    .slice(0, 15);

  const txRows = recentTx.map((tx, idx) => {
    const dateStr = new Date(tx.timestamp).toLocaleDateString('id-ID', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
    });
    const merkSn = [tx.merk, tx.serialNumber].filter(Boolean).join(' • ') || '-';
    const party = `Vendor: ${tx.supplierOrSource || '-'} (Penerima: ${tx.receivedBy || '-'})`;

    return [
      (idx + 1).toString(),
      dateStr,
      tx.id,
      `${tx.itemName} [${tx.itemSku}]`,
      `+${tx.quantity} ${tx.unit}`,
      merkSn,
      tx.conditionStatus || 'Berfungsi',
      party,
      tx.invoiceOrPoNumber || tx.notes || '-',
    ];
  });

  autoTable(doc, {
    startY: currentY + 3,
    head: [['No', 'Tanggal', 'ID Masuk', 'Nama Barang & SKU', 'Jumlah Masuk', 'Merk & No Seri', 'Kondisi', 'Vendor & Petugas Penerima', 'No. PO / BAP']],
    body: txRows,
    theme: 'grid',
    headStyles: {
      fillColor: [16, 149, 100], // emerald-600
      textColor: [255, 255, 255],
      fontSize: 7.5,
      fontStyle: 'bold',
      halign: 'center',
    },
    styles: {
      fontSize: 7,
      cellPadding: 1.8,
      textColor: [30, 41, 59],
    },
    columnStyles: {
      0: { halign: 'center', cellWidth: 7 },
      1: { halign: 'center', cellWidth: 16 },
      2: { fontStyle: 'bold', cellWidth: 18 },
      3: { cellWidth: 36 },
      4: { halign: 'center', fontStyle: 'bold', textColor: [16, 149, 100], cellWidth: 18 },
      5: { cellWidth: 24 },
      6: { halign: 'center', cellWidth: 16 },
      7: { cellWidth: 38 },
      8: { cellWidth: 21 },
    },
    margin: { left: 14, right: 14 },
  });

  // 5. Formal Signature Area (Pengesahan Dokumen)
  const finalY = (doc as any).lastAutoTable?.finalY || 200;
  let signY = finalY + 12;

  if (signY > 240) {
    doc.addPage();
    signY = 30;
  }

  doc.setFont('helvetica', 'normal');
  doc.setFontSize(8.5);
  doc.setTextColor(51, 65, 85);

  // Left signature (Dibuat oleh)
  doc.text('Dibuat & Diverifikasi oleh:', 20, signY);
  doc.text('Petugas Pengelola Inventaris Kantor,', 20, signY + 4);
  doc.line(20, signY + 24, 75, signY + 24);
  doc.setFont('helvetica', 'bold');
  doc.text(preparedBy, 20, signY + 28);
  doc.setFont('helvetica', 'normal');
  doc.text('Bagian Umum & Logistik (GA)', 20, signY + 32);

  // Right signature (Disetujui oleh)
  doc.text('Mengetahui & Menyetujui:', 130, signY);
  doc.text('Kepala Bagian Operasional & Keuangan,', 130, signY + 4);
  doc.line(130, signY + 24, 185, signY + 24);
  doc.setFont('helvetica', 'bold');
  doc.text(approvedBy, 130, signY + 28);
  doc.setFont('helvetica', 'normal');
  doc.text('Manajemen Fasilitas Kantor', 130, signY + 32);

  // Save PDF
  const filename = `Laporan_Inventaris_Kantor_${monthYear.replace(/\s+/g, '_')}.pdf`;
  doc.save(filename);
  return filename;
};
