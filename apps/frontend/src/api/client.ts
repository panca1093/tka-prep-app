/**
 * API client wrapper.
 *
 * Once `make generate-ts` is run, this file should consume types from
 * `@tkaprep/shared-types/generated/ts/index.ts`. Until then we hand-write
 * the bare minimum for /health so the skeleton can boot.
 */

const BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';

export interface HealthResponse {
  status: 'ok' | 'degraded';
  timestamp: string;
  version?: string;
}

export async function getHealth(): Promise<HealthResponse> {
  const res = await fetch(`${BASE_URL}/health`);
  if (!res.ok) {
    throw new Error(`Health check failed: HTTP ${res.status}`);
  }
  return res.json();
}
