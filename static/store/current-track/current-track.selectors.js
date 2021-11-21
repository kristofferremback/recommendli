import createSelector from '../lib/create-selector.js'
import { isReady } from '../lib/with-fetch-state.js'

const selectCurrentTrack = (state) => state.currentTrack

const selectTrackFetchState = createSelector([selectCurrentTrack], (track) => track.fetchState)

export const selectIsPlaying = createSelector([selectCurrentTrack], (currentTrack) => currentTrack.isPlaying)

export const selectTrack = createSelector([selectCurrentTrack], (currentTrack) => currentTrack.track)

export const selectTrackId = createSelector([selectTrack], (track) => (track != null ? track.id : null))

const selectTrackStatus = createSelector([selectCurrentTrack], (currentTrack) => {
  const status =
    currentTrack.status != null
      ? currentTrack.status
      : { inLibrary: false, playlists: undefined, track: undefined }

  return {
    ...status,
    playlists: status.playlists ? status.playlists : [],
    track: status.track ? status.track : { id: null },
  }
})

export const selectTrackInLibrary = createSelector(
  [selectTrackStatus],
  /** @param {{ inLibrary: boolean }} status */
  (status) => status.inLibrary
)

export const selectTrackPlaylists = createSelector(
  [selectTrackStatus],
  /** @param {{ playlists: import('../../recommendli/client.js').Playlist[] }} status */
  (status) => status.playlists
)

export const selectStatusTrackId = createSelector([selectTrackStatus], (status) => status.track.id)

export const selectTrackIsReady = createSelector([selectTrackFetchState], isReady)
