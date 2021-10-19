import createSelector from '../lib/create-selector.js'

const selectCurrentTrack = (state) => state.currentTrack

export const selectIsPlaying = createSelector([selectCurrentTrack], (currentTrack) => currentTrack.isPlaying)

export const selectTrack = createSelector([selectCurrentTrack], (currentTrack) => currentTrack.track)
