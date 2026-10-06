const PHOTO_FOLDER_NAME = 'Foto Inventaris & Geotag';

export interface DriveUploadResult {
  fileId: string;
  name: string;
  webViewLink: string;
  thumbnailLink?: string;
}

export interface DriveFolderInfo {
  id: string;
  name: string;
  webViewLink: string;
}

/**
 * Find or create dedicated folder in Google Drive for inventory photos
 */
export async function findOrCreatePhotoFolder(accessToken: string): Promise<DriveFolderInfo> {
  const query = encodeURIComponent(`name = '${PHOTO_FOLDER_NAME}' and mimeType = 'application/vnd.google-apps.folder' and trashed = false`);
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
      const folder = data.files[0];
      return {
        id: folder.id,
        name: folder.name,
        webViewLink: folder.webViewLink || `https://drive.google.com/drive/folders/${folder.id}`,
      };
    }
  }

  // Create folder
  const createRes = await fetch('https://www.googleapis.com/drive/v3/files', {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${accessToken}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      name: PHOTO_FOLDER_NAME,
      mimeType: 'application/vnd.google-apps.folder',
      description: 'Folder penyimpanan otomatis foto barang masuk dan bukti geotagging',
    }),
  });

  if (!createRes.ok) {
    const err = await createRes.text();
    throw new Error(`Gagal membuat folder di Google Drive: ${err}`);
  }

  const folderData = await createRes.json();
  return {
    id: folderData.id,
    name: PHOTO_FOLDER_NAME,
    webViewLink: `https://drive.google.com/drive/folders/${folderData.id}`,
  };
}

/**
 * Upload geotagged photo directly to Google Drive folder
 */
export async function uploadGeotaggedPhotoToDrive(
  accessToken: string,
  folderId: string,
  filename: string,
  dataUrl: string,
  description?: string
): Promise<DriveUploadResult> {
  // Convert base64 dataUrl to blob
  const res = await fetch(dataUrl);
  const blob = await res.blob();

  const metadata = {
    name: filename,
    parents: [folderId],
    mimeType: 'image/jpeg',
    description: description || 'Foto Geotagged Barang Masuk Inventaris Kantor',
  };

  const form = new FormData();
  form.append('metadata', new Blob([JSON.stringify(metadata)], { type: 'application/json' }));
  form.append('file', blob);

  const uploadRes = await fetch(
    'https://www.googleapis.com/upload/drive/v3/files?uploadType=multipart&fields=id,name,webViewLink,thumbnailLink',
    {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
      body: form,
    }
  );

  if (!uploadRes.ok) {
    const err = await uploadRes.text();
    throw new Error(`Gagal mengunggah foto ke Google Drive: ${err}`);
  }

  const fileData = await uploadRes.json();
  return {
    fileId: fileData.id,
    name: fileData.name,
    webViewLink: fileData.webViewLink || `https://drive.google.com/file/d/${fileData.id}/view`,
    thumbnailLink: fileData.thumbnailLink,
  };
}
