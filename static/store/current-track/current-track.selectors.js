import createSelector from '../lib/create-selector.js'

const selectCurrentTrack = (state) => state.currentTrack

export const selectIsPlaying = createSelector([selectCurrentTrack], (currentTrack) => currentTrack.isPlaying)

export const selectTrack = createSelector([selectCurrentTrack], (currentTrack) => currentTrack.track)

export const selectTrackId = createSelector([selectTrack], (track) => (track != null ? track.id : null))

const selectTrackStatus = createSelector([selectCurrentTrack], (currentTrack) =>
  currentTrack.status != null ? currentTrack.status : { inLibrary: false }
)

export const selectTrackInLibrary = createSelector([selectTrackStatus], (status) => status.inLibrary)
