import { defaultFetchState, updateFetchState } from '../lib/with-fetch-state.js'
import { types } from './current-track.actions.js'

export const initialState = {
  track: null,
  isPlaying: false,
  fetchState: defaultFetchState(),
}

export const reducer = (state = initialState, { type, payload }) => {
  switch (type) {
    case types.SET_CURRENT_TRACK:
      return { ...state, track: payload.track, isPlaying: payload.isPlaying }
    case types.SET_CURRENT_TRACK_FETCH_STATE:
      return { ...state, fetchState: updateFetchState(state.fetchState, payload) }
    default:
      return state
  }
}

export default reducer
