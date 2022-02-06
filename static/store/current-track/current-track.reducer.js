import { defaultFetchState, updateFetchState } from '../lib/with-fetch-state.js'
import { types } from './current-track.actions.js'

export const initialState = {
  track: null,
  isPlaying: false,
  fetchState: defaultFetchState(),
  status: null,
  statusFetchState: defaultFetchState(),
}

export const reducer = (state = initialState, { type, payload }) => {
  switch (type) {
    case types.SET_CURRENT_TRACK:
      return { ...state, track: payload.track, isPlaying: payload.isPlaying }
    case types.SET_CURRENT_TRACK_FETCH_STATE:
      return { ...state, fetchState: updateFetchState(state.fetchState, payload) }
    case types.SET_CURRENT_TRACK_STATUS:
      return { ...state, status: payload }
    case types.SET_CURRENT_TRACK_STATUS_FETCH_STATE:
      return { ...state, statusFetchState: updateFetchState(state.statusFetchState, payload) }
    default:
      return state
  }
}

export default reducer
