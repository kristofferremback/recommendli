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
}

export interface Playlist extends SimplePlaylist {
  snapshot_id: string
  tracks: Track[]
}

export interface IndexSummary {
  playlist_count: number
  unique_track_count: number
  playlists: SimplePlaylist[]
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
