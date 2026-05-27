import client from '@/api/client'
import type { components } from '@tkaprep/shared-types'

type Category = components['schemas']['Category']

export async function updateOwnedCategory(
  id: string,
  name: string,
  description?: string,
): Promise<{ ok: boolean; status?: number; message?: string; data?: Category }> {
  const { data, error, response } = await client.PUT('/contributor/categories/{id}', {
    params: { path: { id } },
    body: { name, description: description || undefined },
  })
  if (error) {
    const msg = (error as any)?.error?.message ?? 'Gagal menyimpan kategori.'
    return { ok: false, status: response?.status, message: msg }
  }
  return { ok: true, data }
}
