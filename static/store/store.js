import { createContext } from 'https://unpkg.com/htm/preact/standalone.module.js'

import { types } from './user.js'

export const StoreContext = createContext()

export const initialState = {
  user: null,
  userFetch: {
    state: 'idle',
    error: null,
  },
  currentTrack: null,
  currentTrackFetch: {
    state: 'idle',
    error: null,
  },
}

export const globalReducer = (state = initialState, { type, payload }) => {
  switch (type) {
    case types.SET_CURRENT_USER:
      return { ...state, user: payload }
    case types.SET_CURRENT_USER_FETCH_STATE:
      return { ...state, userFetch: payload }
    case types.SET_CURRENT_TRACK:
      return { ...state, currentTrack: payload }
    case types.SET_CURRENT_TRACK_FETCH_STATE:
      return { ...state, currentTrackFetch: payload }
    default:
      return state
  }
}
