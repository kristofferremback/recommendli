export interface Artist {
  id: string
  name: string
  external_urls: { spotify: string }
}

export interface SpotifyImage {
  url: string
  height: number
  width: number
}

export interface Album {
  id: string
  name: string
  external_urls: { spotify: string }
  images?: SpotifyImage[]
}

export interface Track {
  id: string
  name: string
  album: Album
  artists: Artist[]
  duration_ms?: number
  uri?: string
  external_urls: { spotify: string }
}

export interface User {
  display_name: string
  external_urls: { spotify: string }
}

export interface SimplePlaylist {
  id: string
  name: string
  external_urls: { spotify: string }
  tracks?: { total: number }
}

export interface Playlist extends Omit<SimplePlaylist, 'tracks'> {
  snapshot_id: string
  tracks: Track[]
}

export interface IndexSummary {
  playlist_count: number
  unique_track_count: number
  playlists: SimplePlaylist[]
  last_synced_at?: string
}

export interface CurrentTrackResponse {
  track?: Track
  is_playing: boolean
}

export interface CheckTrackResponse {
  in_library: boolean
  track: Track
  playlists: SimplePlaylist[]
}

export interface LibraryStatus {
  in_library: boolean
  playlists: SimplePlaylist[]
}

export interface PlaybackDevice {
  id?: string
  is_active: boolean
  is_restricted: boolean
  name: string
  type: string
  volume_percent: number
}

export interface PlaybackContext {
  external_urls?: { spotify?: string }
  href?: string
  type?: string
  uri?: string
}

export interface Playback {
  active: boolean
  is_playing: boolean
  progress_ms: number
  /** Spotify's clock at the last state change. Not the time of progress_ms. */
  timestamp: number
  /** Client clock when progress_ms arrived. Stamped in queries.ts, not sent by the server. */
  fetched_at: number
  track?: Track
  device?: PlaybackDevice
  context?: PlaybackContext
  shuffle_state: boolean
  repeat_state?: string
}

export interface PlaybackQueue {
  currently_playing?: Track
  tracks: Track[]
}

export interface RecentlyPlayedItem {
  track: Track
  played_at: string
  context?: PlaybackContext
}

export interface QueueSkipRequest {
  position: number
  expected_track_id: string
  expected_current_track_id?: string
}
