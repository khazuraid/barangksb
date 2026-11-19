import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 15000,
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    const path = window.location.pathname
    if (error.response?.status === 401 && path !== '/login' && !path.startsWith('/lapor') && !path.startsWith('/scan/')) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export function safeHref(raw?: string | null): string {
  const s = (raw || '').trim()
  if (!s) return ''
  if (s.startsWith('/api/uploads/') || s.startsWith('/uploads/')) {
    if (!s.startsWith('//') && !s.includes('..')) return s
    return ''
  }
  try {
    const u = new URL(s)
    if (u.protocol === 'https:' || u.protocol === 'http:') return u.toString()
  } catch {
    return ''
  }
  return ''
}

export async function downloadAuthenticatedFile(
  path: string,
  fallbackFilename: string,
  openInNewTab = false
): Promise<void> {
  const res = await api.get(path, { responseType: 'blob' })
  const contentType = String(res.headers['content-type'] || 'application/octet-stream')
  const blob = new Blob([res.data], { type: contentType })
  const blobUrl = URL.createObjectURL(blob)

  if (openInNewTab) {
    window.open(blobUrl, '_blank', 'noopener,noreferrer')
    setTimeout(() => URL.revokeObjectURL(blobUrl), 60_000)
    return
  }

  let filename = fallbackFilename
  const disp = res.headers['content-disposition'] as string | undefined
  if (disp) {
    const m = /filename="?([^";]+)"?/i.exec(disp)
    if (m?.[1]) filename = m[1]
  }
  const a = document.createElement('a')
  a.href = blobUrl
  a.download = filename
  document.body.appendChild(a)
  a.click()
  a.remove()
  setTimeout(() => URL.revokeObjectURL(blobUrl), 10_000)
}

export default api
