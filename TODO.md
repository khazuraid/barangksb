# Status Perubahan & Daftar Pekerjaan Tersisa

Terakhir diperbarui: 2026-11-20

## Sudah Selesai & Terverifikasi

### 1. Perbaikan Audit Backend & Keamanan
- **Single source of truth sanitasi spreadsheet**: `SanitizeSpreadsheetCell` dipindah ke `backend/internal/service/spreadsheet.go` dan dipakai bersama oleh `handler` dan `telegram`.
- **Worker & Telegram**: `InitSender(ctx)` diaktifkan pada mode worker (`cmd/server/main.go`), mutasi stok Telegram memakai transaksi `FOR UPDATE`, dan error query tidak lagi diabaikan (`backend/internal/telegram/bot.go`).
- **Konsistensi upload & hapus foto**: `SavePhoto` dan `DeleteUploadedPhoto` disatukan di `backend/internal/handler/upload.go` (tanpa fallback diam-diam ke disk saat MinIO gagal), `PublicReportHandler` memakai `UploadHandler` yang sama (`backend/internal/handler/public_report.go`), dan `getDBReferences` digabung menjadi satu query `UNION ALL`.
- **Parsing angka ketat & DB pool**: `parseStrictInt64` / `parseStrictInt32` di `backend/internal/handler/helpers.go`, `cmd/server/main.go` memakai `db.Open(cfg.DBURL)`.
- **Unit test Go**: race condition di `internal/worker/worker_test.go` diperbaiki; test baru ditambahkan di `internal/auth/auth_test.go`, `internal/middleware/hardening_test.go`, `internal/telegram/bot_test.go`, dan `internal/handler/handler_test.go`.

### 2. Kalibrasi Cetak Barcode (Tom & Jerry / A4 / Thermal)
- `@page { size: 210mm 297mm; margin: 0; }`, `.sheet-page` `width: 210mm; box-sizing: border-box;` dipertahankan di `@media print`, dan page break `:not(:last-child)` diperbaiki di `backend/internal/handler/barcode.go` dan `frontend/src/utils/customLabelPrint.ts`.
- Preset label dan helper cetak diekstrak ke `frontend/src/utils/labelPresets.ts` dan `printUrlInHiddenFrame` di `frontend/src/utils/customLabelPrint.ts`.

### 3. Unit Fisik Terdaftar di Barcode, Detail Barang, & Scan QR
- **Backend**: `GET /api/barcode/units` (`ListUnits`) ditambahkan dan diproteksi JWT + RBAC di `backend/internal/middleware/middleware.go`; `PNG` dan `Sheet` fallback ke `i.sku` jika `u.barcode` kosong; `ScanItem` (`GET /api/scan/:id`) selalu mengembalikan seluruh `item_units` terdaftar beserta foto & detail unit (dengan redaksi field sensitif saat unauthenticated).
- **Frontend**:
  - `frontend/src/views/BarcodeView.vue`: switcher **Master Barang** vs **Daftar Unit Fisik Terdaftar**, filter `item_id`, auto-select dari query URL, dan badge info unit pada kartu.
  - `frontend/src/views/ItemDetailView.vue`: tombol **Label QR** & **Generate Barcode Semua Unit** mengarah ke `/barcode?mode=units&item_id=...&auto_select=1`.
  - `frontend/src/views/ScanView.vue`: tab **Unit Fisik (N)** dan ringkasan unit di tab **Spesifikasi** menampilkan seluruh unit terdaftar, foto yang bisa diperbesar, dan detail lengkap unit.

### 4. Redesign Halaman Pengaduan Publik (`/lapor`)
- `frontend/src/views/PublicReportView.vue` didesain ulang mengikuti sistem desain Clinical Precision: header sticky, hero band, layout 2 kolom (form + sidebar alur penanganan), pemilih kategori visual, upload foto drag-and-drop, riwayat tiket lokal (`localStorage`), timeline pelacakan 3 tahap, dan lightbox foto.

### 5. Pembersihan File Deploy Usang
- Menghapus `Caddyfile`, `frontend/Dockerfile`, dan `frontend/nginx.conf`, serta memperbarui `README.md`.

### Hasil Verifikasi Lokal
- `cd backend && go test -count=1 -race ./...` → **LULUS** (`auth`, `handler`, `middleware`, `telegram`, `worker`).
- `cd frontend && npx vue-tsc --noEmit && npm run build` → **LULUS**.

---

## Yang Belum Dikerjakan / Perlu Verifikasi

1. **Uji cetak fisik ke printer & kertas stiker asli**
   - Kalibrasi CSS `@page` dan margin Tom & Jerry (103, 107, 108, 121), A4 (3×7, 4×10), serta thermal roll (58/80 mm) belum diuji cetak langsung ke hardware printer fisik.
2. **Uji integrasi end-to-end dengan PostgreSQL, Redis, & MinIO live**
   - Endpoint baru/diubah (`GET /api/barcode/units`, `GET /api/scan/:id`, `POST /api/public/reports`, `POST /api/public/upload`, dan worker Telegram) baru terverifikasi lewat unit test Go; belum dijalankan terhadap container DB/Redis/MinIO live.
3. **Uji visual & interaksi browser live**
   - Halaman `/lapor`, `/scan/:id`, `/barcode`, dan `/items/:id` belum dibuka dan diuji langsung di browser (desktop & mobile, light & dark mode).
4. **Pemecahan sub-komponen UI berukuran besar (opsional / maintainability)** — *sebagian selesai (2026-11-20)*
   - Tab **Printer & Lembar Cetak** diekstrak dari `LabelDesignerModal.vue` ke `frontend/src/components/LabelPrinterPanel.vue` (memakai `defineModel`; `vue-tsc` & `npm run build` lulus). `LabelDesignerModal.vue` kini 3.038 baris.
   - Belum: canvas editor, inspector baris/sel, tab Kop/Barcode/Footer, dan `BarcodeView.vue` (1.757 baris). Perlu uji visual di browser setelah dipecah.
5. ~~**Audit sisa file deploy lama**~~ — **Selesai (2026-11-20)**
   - `backend/Dockerfile` dihapus (tidak dipakai: `docker-compose.yml` memakai `Dockerfile` root, dan versi backend tidak membangun frontend untuk `go:embed`). `backend/dist/.gitkeep` dihapus (embed memakai `backend/cmd/server/dist`, bukan `backend/dist`).
6. **Build ulang & deploy ke server produksi**
   - Perubahan masih berada di repo lokal dan belum di-deploy ke `https://barang.kesling.biz.id`.
