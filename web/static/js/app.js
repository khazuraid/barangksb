// ============================================================
// Inventaris Kantor — app.js v2
// Dark mode · toast · HTMX helpers · QR print
// ============================================================

// ---- dark mode toggle ----
(function () {
  const saved = localStorage.getItem('ik-theme');
  if (saved === 'dark') document.documentElement.setAttribute('data-theme', 'dark');
  document.addEventListener('click', function (e) {
    const btn = e.target.closest('#theme-toggle');
    if (!btn) return;
    const dark = document.documentElement.getAttribute('data-theme') === 'dark';
    document.documentElement.toggleAttribute('data-theme', !dark);
    if (!dark) document.documentElement.setAttribute('data-theme', 'dark');
    else document.documentElement.removeAttribute('data-theme');
    localStorage.setItem('ik-theme', dark ? 'light' : 'dark');
  });
})();

// ---- toast helper (dipakai HTMX response header atau manual) ----
function toast(msg, type) {
  const zone = document.getElementById('toast-zone');
  if (!zone) return;
  const el = document.createElement('div');
  el.className = 'toast-item';
  el.textContent = msg;
  zone.appendChild(el);
  setTimeout(() => el.remove(), 4000);
}

// ---- HTMX: toast dari header X-Toast setelah swap ----
document.body.addEventListener('htmx:afterSwap', function (e) {
  const t = e.detail.xhr && e.detail.xhr.getResponseHeader && e.detail.xhr.getResponseHeader('X-Toast');
  if (t) toast(t);
});

// ---- HTMX: push URL agar back/forward jalan ----
document.body.addEventListener('htmx:pushedIntoHistory', function () {
  window.dispatchEvent(new Event('resize'));
});

// ---- barcode label sheet ----
function printSheet() {
  const ids = Array.from(document.querySelectorAll('.sheet-check:checked')).map((c) => c.value);
  if (!ids.length) { toast('Pilih minimal satu barang'); return; }
  window.open('/barcode/sheet?ids=' + ids.join(','), '_blank');
}
function selectAll() {
  const boxes = document.querySelectorAll('.sheet-check');
  const all = Array.from(boxes).every((b) => b.checked);
  boxes.forEach((b) => (b.checked = !all));
}

// ---- HTMX error feedback ----
document.body.addEventListener('htmx:responseError', function (e) {
  toast('Terjadi kesalahan (' + (e.detail.xhr ? e.detail.xhr.status : '?') + ')');
});
