import { defaultFetchState, updateFetchState } from '../lib/with-fetch-state.js'
import { types } from './generate.actions.js'

export const initialState = {
  discovery: {
    playlist: null,
    fetchState: defaultFetchState(),
  },
}

export const reducer = (state = initialState, { type, payload }) => {
  switch (type) {
    case types.SET_DISCOVERY_PLAYLIST:
      return { ...state, discovery: { ...state.discovery, playlist: payload } }
    case types.SET_DISCOVERY_FETCH_STATE:
      return {
        ...state,
        discovery: { ...state.discovery, fetchState: updateFetchState(state.discovery.fetchState, payload) },
      }
    default:
      return state
  }
}
