// Kamera scanner untuk halaman Barang Masuk (html5-qrcode)
(function () {
  const btn = document.getElementById('btn-scan');
  const box = document.getElementById('scanner-box');
  const skuInput = document.getElementById('movement-sku');
  if (!btn || !box) return;
  let scanner = null;

  btn.addEventListener('click', function () {
    if (scanner && scanner.isScanning) {
      scanner.clear().then(() => { box.classList.add('hidden'); });
      scanner = null;
      btn.textContent = '📷 Scan Kamera';
      return;
    }
    box.classList.remove('hidden');
    btn.textContent = '✕ Stop Scan';
    scanner = new Html5Qrcode('scanner-box');
    scanner.start(
      { facingMode: 'environment' },
      { fps: 10, qrbox: { width: 240, height: 140 } },
      (text) => {
        skuInput.value = text;
        skuInput.dispatchEvent(new Event('input'));
        skuInput.focus();
      },
      () => {} // per-frame errors: abaikan
    ).catch((err) => {
      box.innerHTML = '<p class="alert alert-error">Kamera tidak tersedia: ' + err + '</p>';
    });
  });

  // Ketik SKU → auto pilih item di dropdown
  const select = document.getElementById('item-select');
  const lookup = document.getElementById('item-lookup');
  if (skuInput && select) {
    const doLookup = () => {
      const q = skuInput.value.trim().toLowerCase();
      if (!q) { lookup.textContent = 'Ketik SKU untuk mencari barang…'; return; }
      let found = null;
      Array.from(select.options).forEach((opt) => {
        const sku = (opt.dataset.sku || '').toLowerCase();
        if (sku === q) found = opt;
      });
      if (found) {
        select.value = found.value;
        lookup.innerHTML = '✓ Ditemukan: <strong>' + found.textContent + '</strong>';
      } else {
        select.value = '';
        lookup.innerHTML = 'SKU <code>' + q + '</code> belum terdaftar — <a href="/items/new">daftarkan barang baru</a>';
      }
    };
    skuInput.addEventListener('input', doLookup);
    skuInput.addEventListener('change', doLookup);
  }
})();
