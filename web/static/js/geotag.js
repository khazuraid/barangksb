// Geotag — ambil GPS browser, isi field hidden sebelum submit
(function () {
  const btn = document.getElementById('btn-geotag');
  const status = document.getElementById('geotag-status');
  if (!btn) return;

  function set(lat, lng, acc, name) {
    document.getElementById('geo_lat').value = lat || '';
    document.getElementById('geo_lng').value = lng || '';
    document.getElementById('geo_acc').value = acc || '';
    document.getElementById('geo_name').value = name || '';
    status.textContent = lat
      ? '✓ GPS: ' + lat.toFixed(6) + ', ' + lng.toFixed(6) + ' (±' + Math.round(acc) + 'm)'
      : 'GPS tidak tersedia';
  }

  btn.addEventListener('click', function () {
    if (!navigator.geolocation) { set(null); return; }
    status.textContent = 'Mengambil GPS…';
    navigator.geolocation.getCurrentPosition(
      (pos) => set(pos.coords.latitude, pos.coords.longitude, pos.coords.accuracy || 0, ''),
      (err) => { status.textContent = 'GPS gagal: ' + err.message; set(null); },
      { enableHighAccuracy: true, timeout: 10000 }
    );
  });

  // Reverse-geocode ringan: nama petugas saja (tanpa API eksternal — YAGNI)
  const form = document.getElementById('movement-form');
  if (form) {
    form.addEventListener('submit', () => {
      if (!document.getElementById('geo_name').value) {
        document.getElementById('geo_name').value = 'Lokasi manual';
      }
    });
  }
})();
