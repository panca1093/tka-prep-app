const apiBase = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080/api/v1'
const uploadsBase = apiBase.replace(/\/api\/v1\/?$/, '')

export function resolveUrl(path: string | null | undefined): string {
  if (!path) return ''
  if (path.startsWith('http://') || path.startsWith('https://')) return path
  return uploadsBase + path
}
