import type { User, CurrentTrackResponse, Playlist, CheckTrackResponse, IndexSummary } from '@/shared/types/spotify'

const BASE_URL = '/recommendations/v1'

/**
 * Fetch wrapper with error handling
 * Throws on 4xx/5xx responses for TanStack Query error handling
 */
async function fetchAPI(url: string, init?: RequestInit): Promise<Response> {
  const response = await fetch(url, init)

  if (!response.ok) {
    const error = new Error(`HTTP ${response.status}: ${response.statusText}`)
    try {
      const body = await response.json()
      Object.assign(error, { status: response.status, body })
    } catch {}
    throw error
  }

  return response
}

export const api = {
  getCurrentUser: async (): Promise<User> => {
    const res = await fetchAPI(`${BASE_URL}/whoami`)
    return res.json()
  },

  getCurrentTrack: async (): Promise<CurrentTrackResponse> => {
    const res = await fetchAPI(`${BASE_URL}/current-track`)
    return res.json()
  },

  checkCurrentTrack: async (): Promise<CheckTrackResponse> => {
    const res = await fetchAPI(`${BASE_URL}/check-current-track-in-library`)
    return res.json()
  },

  generateDiscoveryPlaylist: async (dryRun = false): Promise<Playlist> => {
    const res = await fetchAPI(`${BASE_URL}/generate-discovery-playlist?dryrun=${dryRun}`)
    const data = await res.json()
    return {
      ...data,
      tracks: data.tracks.items?.map((item: any) => item.track) || data.tracks,
    }
  },

  getIndexSummary: async (): Promise<IndexSummary> => {
    const res = await fetchAPI(`${BASE_URL}/index/summary`)
    return res.json()
  },
}
