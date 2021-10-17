import { withFetchState } from '../lib/with-fetch-state.js'
import { redirectingFetch, throwOn404 } from '../../spotify/redirecting-fetch.js'

export const types = {
  SET_CURRENT_TRACK: 'SET_CURRENT_TRACK',
  SET_CURRENT_TRACK_FETCH_STATE: 'SET_CURRENT_TRACK_FETCH_STATE',
}

const setCurrentTrack = ({ track, isPlaying }) => ({
  type: types.SET_CURRENT_TRACK,
  payload: { track, isPlaying },
})

/**
 * @param {object} opts
 * @param {'idle'|'loading'|'error'} opts.state
 * @param {Error} [opts.error]
 */
const setCurrentTrackFetchState = ({ state, error }) => ({
  type: types.SET_CURRENT_TRACK_FETCH_STATE,
  payload: { state, error },
})

export const getCurrentTrackAsync = () => {
  return withFetchState(setCurrentTrackFetchState, async (dispatch) => {
    const response = await throwOn404(redirectingFetch('/recommendations/v1/current-track'))
    const { track, is_playing: isPlaying } = await response.json()

    dispatch(setCurrentTrack({ isPlaying, track }))
  })
}
