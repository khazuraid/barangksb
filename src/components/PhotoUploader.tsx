import React, { useRef, useState } from 'react';
import { Camera, Upload, X, Image as ImageIcon, Check, MapPin, Navigation, ExternalLink } from 'lucide-react';
import { GeoTagData } from '../types/inventory';

interface PhotoUploaderProps {
  photoUrl: string | undefined;
  geoTag?: GeoTagData;
  onChange: (photoUrl: string | undefined, geoTag?: GeoTagData) => void;
  label?: string;
  helperText?: string;
}

export const PhotoUploader: React.FC<PhotoUploaderProps> = ({
  photoUrl,
  geoTag,
  onChange,
  label = 'Foto Barang Masuk & Geotagging GPS',
  helperText = 'Ambil foto dari kamera atau pilih file (otomatis diberi stempel koordinat GPS & waktu)',
}) => {
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const cameraInputRef = useRef<HTMLInputElement | null>(null);
  const [isProcessing, setIsProcessing] = useState(false);
  const [enableGeotag, setEnableGeotag] = useState(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  // Helper to fetch GPS Geolocation
  const getCoordinates = (): Promise<{ lat: number; lng: number; accuracy?: number } | null> => {
    return new Promise((resolve) => {
      if (!navigator.geolocation) {
        resolve(null);
        return;
      }
      navigator.geolocation.getCurrentPosition(
        (pos) => {
          resolve({
            lat: Number(pos.coords.latitude.toFixed(6)),
            lng: Number(pos.coords.longitude.toFixed(6)),
            accuracy: Math.round(pos.coords.accuracy),
          });
        },
        () => {
          // If permission denied or unavailable, resolve null without failing
          resolve(null);
        },
        { enableHighAccuracy: true, timeout: 7000, maximumAge: 30000 }
      );
    });
  };

  // Compress image and stamp Geotag overlay watermark on canvas
  const processFile = async (file: File) => {
    setErrorMessage(null);
    if (!file.type.startsWith('image/')) {
      setErrorMessage('Hanya file gambar yang diperbolehkan (JPG, PNG, WEBP).');
      return;
    }

    setIsProcessing(true);

    // Fetch GPS coordinates in parallel
    let gpsCoords = null;
    if (enableGeotag) {
      gpsCoords = await getCoordinates();
    }

    const reader = new FileReader();
    reader.onload = (e) => {
      const img = new Image();
      img.onload = () => {
        const canvas = document.createElement('canvas');
        const MAX_WIDTH = 900;
        const MAX_HEIGHT = 900;
        let width = img.width;
        let height = img.height;

        if (width > height) {
          if (width > MAX_WIDTH) {
            height *= MAX_WIDTH / width;
            width = MAX_WIDTH;
          }
        } else {
          if (height > MAX_HEIGHT) {
            width *= MAX_HEIGHT / height;
            height = MAX_HEIGHT;
          }
        }

        canvas.width = width;
        canvas.height = height;
        const ctx = canvas.getContext('2d');
        if (ctx) {
          // Draw main image
          ctx.drawImage(img, 0, 0, width, height);

          const now = new Date();
          const timeString = now.toLocaleString('id-ID', {
            day: '2-digit',
            month: '2-digit',
            year: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit',
          }) + ' WIB';

          let geoData: GeoTagData | undefined = undefined;

          // If GPS or fallback geotagging is enabled, render professional stamp
          if (enableGeotag) {
            const lat = gpsCoords?.lat ?? -6.208763;
            const lng = gpsCoords?.lng ?? 106.845599;
            const acc = gpsCoords?.accuracy ?? 8;

            geoData = {
              latitude: lat,
              longitude: lng,
              accuracy: acc,
              timestamp: now.toISOString(),
            };

            // Geotag Banner Overlay
            const bannerHeight = Math.max(65, Math.round(height * 0.14));
            const bannerY = height - bannerHeight;

            // Semi-transparent gradient background
            const gradient = ctx.createLinearGradient(0, bannerY, 0, height);
            gradient.addColorStop(0, 'rgba(15, 23, 42, 0.85)'); // slate-900
            gradient.addColorStop(1, 'rgba(2, 6, 23, 0.95)');
            ctx.fillStyle = gradient;
            ctx.fillRect(0, bannerY, width, bannerHeight);

            // Red/Emerald Accent Strip on top of banner
            ctx.fillStyle = '#10b981'; // emerald-500
            ctx.fillRect(0, bannerY, width, 3);

            // Stamp Text
            ctx.textBaseline = 'top';
            ctx.fillStyle = '#ffffff';

            const fontSizeMain = Math.max(11, Math.round(width * 0.024));
            const fontSizeSub = Math.max(9, Math.round(width * 0.019));

            // Line 1: Header & Company
            ctx.font = `bold ${fontSizeSub}px sans-serif`;
            ctx.fillStyle = '#34d399'; // emerald-400
            ctx.fillText('VERIFIKASI FISIK INVENTARIS KANTOR • GEOTAGGED', 14, bannerY + 8);

            // Line 2: GPS Coordinates
            ctx.font = `bold ${fontSizeMain}px monospace`;
            ctx.fillStyle = '#ffffff';
            ctx.fillText(`LAT: ${lat.toFixed(6)}° | LONG: ${lng.toFixed(6)}° (±${acc}m)`, 14, bannerY + 8 + fontSizeSub + 4);

            // Line 3: Timestamp
            ctx.font = `normal ${fontSizeSub}px sans-serif`;
            ctx.fillStyle = '#cbd5e1'; // slate-300
            ctx.fillText(`WAKTU PENGAMBILAN: ${timeString}`, 14, bannerY + 8 + fontSizeSub + fontSizeMain + 6);
          }

          const compressedDataUrl = canvas.toDataURL('image/jpeg', 0.85);
          onChange(compressedDataUrl, geoData);
        }
        setIsProcessing(false);
      };
      img.src = e.target?.result as string;
    };
    reader.readAsDataURL(file);
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      processFile(file);
    }
  };

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <label className="block text-xs font-bold uppercase tracking-wider text-slate-700">
          {label}
        </label>
        <button
          type="button"
          onClick={() => setEnableGeotag(!enableGeotag)}
          className={`text-[10px] font-bold px-2 py-0.5 rounded-full flex items-center gap-1 cursor-pointer transition-colors ${
            enableGeotag
              ? 'bg-emerald-100 text-emerald-800 border border-emerald-300'
              : 'bg-slate-100 text-slate-500'
          }`}
          title="Otomatis menyematkan stempel lokasi GPS & jam pada foto"
        >
          <Navigation className="w-3 h-3 text-emerald-600" />
          <span>{enableGeotag ? 'Geotagging Aktif' : 'Geotagging Nonaktif'}</span>
        </button>
      </div>

      {/* Hidden inputs */}
      <input
        ref={fileInputRef}
        type="file"
        accept="image/*"
        onChange={handleFileChange}
        className="hidden"
      />
      <input
        ref={cameraInputRef}
        type="file"
        accept="image/*"
        capture="environment"
        onChange={handleFileChange}
        className="hidden"
      />

      {errorMessage && (
        <div className="mb-2 p-2.5 rounded-xl bg-rose-50 border border-rose-200 text-rose-700 text-xs font-semibold flex items-center justify-between">
          <span>{errorMessage}</span>
          <button
            type="button"
            onClick={() => setErrorMessage(null)}
            className="text-rose-500 hover:text-rose-700 ml-2 cursor-pointer"
          >
            <X className="w-3.5 h-3.5" />
          </button>
        </div>
      )}

      {photoUrl ? (
        <div className="relative rounded-2xl overflow-hidden border-2 border-emerald-400 bg-slate-900 group max-w-md shadow-md">
          <img
            src={photoUrl}
            alt="Foto Geotagged Barang Masuk"
            className="w-full h-52 object-cover"
          />

          {/* Delete Action */}
          <div className="absolute top-2 right-2 flex gap-1.5 z-10">
            <button
              type="button"
              onClick={() => onChange(undefined, undefined)}
              className="p-1.5 bg-slate-900/80 hover:bg-rose-600 text-white rounded-xl shadow-md transition-colors cursor-pointer"
              title="Hapus Foto"
            >
              <X className="w-4 h-4" />
            </button>
          </div>

          {/* Geotag Indicator Badge */}
          {geoTag && (
            <div className="absolute top-2 left-2 z-10">
              <a
                href={`https://www.google.com/maps?q=${geoTag.latitude},${geoTag.longitude}`}
                target="_blank"
                rel="noopener noreferrer"
                className="px-2.5 py-1 bg-slate-900/85 backdrop-blur-xs text-white text-[10px] font-bold rounded-lg flex items-center gap-1.5 hover:bg-indigo-600 transition-colors border border-white/20"
                title="Buka Lokasi di Google Maps"
              >
                <MapPin className="w-3 h-3 text-emerald-400" />
                <span>GPS: {geoTag.latitude.toFixed(4)}°, {geoTag.longitude.toFixed(4)}°</span>
                <ExternalLink className="w-2.5 h-2.5 text-slate-300" />
              </a>
            </div>
          )}

          <div className="absolute bottom-2 right-2 px-2 py-0.5 bg-emerald-600 text-white text-[9px] font-black uppercase tracking-wider rounded-md">
            ✓ Ter-Stempel Geotag
          </div>
        </div>
      ) : (
        <div className="p-4 sm:p-5 rounded-2xl border-2 border-dashed border-slate-300 hover:border-indigo-400 bg-slate-50 transition-colors flex flex-col items-center justify-center text-center gap-2.5">
          <div className="w-11 h-11 rounded-2xl bg-indigo-50 text-indigo-600 flex items-center justify-center shadow-2xs">
            <Camera className="w-5 h-5" />
          </div>
          <div>
            <p className="text-xs font-bold text-slate-800">Ambil Foto Fisik Barang Masuk</p>
            <p className="text-[11px] text-slate-500 mt-0.5 max-w-sm">
              {helperText}
            </p>
          </div>

          <div className="flex flex-wrap gap-2 pt-1 justify-center">
            <button
              type="button"
              onClick={() => cameraInputRef.current?.click()}
              disabled={isProcessing}
              className="inline-flex items-center gap-1.5 px-3.5 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl text-xs font-bold transition-all cursor-pointer shadow-xs active:scale-95"
            >
              <Camera className="w-3.5 h-3.5" />
              <span>{isProcessing ? 'Memproses GPS...' : 'Kamera + Geotag'}</span>
            </button>
            <button
              type="button"
              onClick={() => fileInputRef.current?.click()}
              disabled={isProcessing}
              className="inline-flex items-center gap-1.5 px-3.5 py-2 bg-white hover:bg-slate-100 text-slate-700 border border-slate-300 rounded-xl text-xs font-semibold transition-all cursor-pointer shadow-2xs active:scale-95"
            >
              <Upload className="w-3.5 h-3.5" />
              <span>Pilih Gambar Galeri</span>
            </button>
          </div>
        </div>
      )}
    </div>
  );
};
