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
 * @typedef Playlist
 * @property {string} id
 * @property {string} name
 * @property {string} snapshot_id
 * @property {ExternalUrls} external_urls
 * @property {Track[]} tracks
 *
 * @typedef SimplePlaylist
 * @property {string} id
 * @property {string} name
 * @property {ExternalUrls} external_urls
 *
 * @typedef IndexSummary
 * @property {number} playlist_count
 * @property {number} unique_track_count
 * @property {SimplePlaylist[]} playlists
 */

const recommendliClient = {
  /**
   * @returns {Promise<{ isPlaying: boolean, track?: Track }>}
   */
  getCurrentTrack: async () => {
    const response = await throwOn404(redirectingFetch('/recommendations/v1/current-track'))
    const { track, is_playing: isPlaying } = await response.json()
    return { track, isPlaying }
  },
  /**
   * @returns {Promise<User>}
   */
  getCurrentUser: async () => {
    const response = await throwOn404(redirectingFetch('/recommendations/v1/whoami'))
    return await response.json()
  },
  /**
   * @returns {Promise<Playlist>}
   */
  generateDiscoveryPlaylist: async ({ dryRun = false } = {}) => {
    const response = await throwOn404(
      redirectingFetch(`/recommendations/v1/generate-discovery-playlist?dryrun=${dryRun || false}`)
    )
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
    const response = await throwOn404(redirectingFetch('/recommendations/v1/check-current-track-in-library'))
    const { in_library: inLibrary, track, playlists } = await response.json()
    return {
      inLibrary,
      track,
      playlists: !playlists ? [] : playlists.map((p) => ({ ...p, tracks: undefined })),
    }
  },

  /**
   * @returns {Promise<IndexSummary>}
   */
  getIndexSummary: async () => {
    const response = await throwOn404(redirectingFetch('/recommendations/v1/index/summary'))
    return await response.json()
  },

  /**
   *
   * @param {string} id
   * @returns {Promise<Playlist>}
   */
  getPlaylistById: async (id) => {
    const response = await throwOn404(redirectingFetch(`/recommendations/v1/playlists/${id}`))
    return await response.json()
  },
}

export default recommendliClient
