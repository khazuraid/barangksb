// SSE realtime — reload halaman saat data berubah
(function () {
  if (!window.EventSource) return;
  const es = new EventSource('/events');
  ['items', 'tx'].forEach((ch) => {
    es.addEventListener(ch.name || ch, () => location.reload());
  });
  es.onmessage = (e) => {
    try {
      const d = JSON.parse(e.data);
      if (d.channel === 'items' || d.channel === 'tx') location.reload();
    } catch (_) {}
  };
})();

// barcode label sheet
function printSheet() {
  const ids = Array.from(document.querySelectorAll('.sheet-check:checked')).map((c) => c.value);
  if (!ids.length) { alert('Pilih minimal satu barang'); return; }
  window.open('/barcode/sheet?ids=' + ids.join(','), '_blank');
}
function selectAll() {
  const boxes = document.querySelectorAll('.sheet-check');
  const all = Array.from(boxes).every((b) => b.checked);
  boxes.forEach((b) => (b.checked = !all));
}
