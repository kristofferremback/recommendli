import { throwOn404, redirectingFetch } from './redirecting-fetch.js'

/**
 * @typedef ExternalUrls
 * @property {string} spotify
 *
 * @typedef Artist
 * @property {string} name
 * @property {ExternalUrls} external_urls
 *
 * @typedef Album
 * @property {string} id
 * @property {string} name
 * @property {ExternalUrls} external_urls
 *
 * @typedef Track
 * @property {string} name
 * @property {string} id
 * @property {Album} album
 * @property {Artist[]} artists
 * @property {ExternalUrls} external_urls
 *
 * @typedef User
 * @property {string} display_name
 * @property {ExternalUrls} external_urls
 *
 * @typedef UserPrefs
 * @property {string} library_pattern
 * @property {string[]} discovery_playlist_names
 * @property {{ [key: string]: number }} weighted_words
 * @property {number} minimum_album_size
 *
 * @typedef Playlist
 * @property {string} id
 * @property {string} name
 * @property {ExternalUrls} external_urls
 * @property {Track[]} tracks
 *
 * @typedef SimplePlaylist
 * @property {string} id
 * @property {string} name
 * @property {ExternalUrls} external_urls
 *
 */

/**
 * @param {RequestInfo} input
 * @param {RequestInit} [init]
 * @returns Promise<Response>
 */
const req = async (input, init = {}) => {
  return await throwOn404(redirectingFetch(input, init))
}

/**
 * @template T
 * @param {RequestInfo} input
 * @param {T} body
 * @param {RequestInit} [init]
 * @returns Promise<Response>
 */
const reqPostJSON = async (input, body, init = {}) => {
  const headers = new Headers(init.headers || {})
  headers.set('content-type', 'application/json')
  return req(input, { ...init, method: 'POST', body: JSON.stringify(body), headers: headers })
}

const recommendliClient = {
  /**
   * @returns {Promise<{ isPlaying: boolean, track?: Track }>}
   */
  getCurrentTrack: async () => {
    const response = await req('/recommendations/v1/current-track')
    const { track, is_playing: isPlaying } = await response.json()
    return { track, isPlaying }
  },
  /**
   * @returns {Promise<User>}
   */
  getCurrentUser: async () => {
    const response = await req('/recommendations/v1/whoami')
    return await response.json()
  },
  /**
   * @returns {Promise<Playlist>}
   */
  generateDiscoveryPlaylist: async ({ dryRun = false } = {}) => {
    const response = await req(`/recommendations/v1/generate-discovery-playlist?dryrun=${dryRun || false}`)
    const fullPlaylist = await response.json()
    return {
      ...fullPlaylist,
      tracks: fullPlaylist.tracks.items.map((item) => item.track),
    }
  },
  /**
   * @returns {Promise<{ inLibrary: boolean, track: Track, playlists: SimplePlaylist[] }>}
   */
  checkCurrentTrack: async () => {
    const response = await req('/recommendations/v1/check-current-track-in-library')
    const { in_library: inLibrary, track, playlists } = await response.json()
    return {
      inLibrary,
      track,
      playlists: !playlists ? [] : playlists.map((p) => ({ ...p, tracks: undefined })),
    }
  },
  /**
   * @returns {Promise<UserPrefs>}
   */
  getUserPreferences: async () => {
    const response = await req('/recommendations/v1/user-preferences')
    return await response.json()
  },

  /**
   * @param {UserPrefs} prefs
   * @returns {Promise<UserPrefs>}
   */
  setUserPreferences: async (prefs) => {
    const response = await reqPostJSON('/recommendations/v1/user-preferences', prefs)
    return await response.json()
  },
}

export default recommendliClient
