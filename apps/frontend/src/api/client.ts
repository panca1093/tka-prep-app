import createClient, { type Middleware } from 'openapi-fetch'
import type { paths } from '@tkaprep/shared-types'

// Token is stored here so the auth store can update it without a circular dep.
let _accessToken = ''

export function setAccessToken(token: string) {
  _accessToken = token
}

export function clearAccessToken() {
  _accessToken = ''
}

export function getAccessToken(): string {
  return _accessToken
}

const authMiddleware: Middleware = {
  async onRequest({ request }) {
    if (_accessToken) {
      request.headers.set('Authorization', `Bearer ${_accessToken}`)
    }
    return request
  },
}

const client = createClient<paths>({
  baseUrl: (import.meta.env.VITE_API_BASE_URL as string) || 'http://localhost:8080/api/v1',
})

client.use(authMiddleware)

export default client
