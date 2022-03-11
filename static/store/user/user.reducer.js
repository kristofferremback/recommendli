import { defaultFetchState, updateFetchState } from '../lib/with-fetch-state.js'
import { types } from './user.actions.js'

export const initialState = {
  user: null,
  fetchState: defaultFetchState(),
  preferences: null,
  preferencesFetchState: defaultFetchState(),
}

export const reducer = (state = initialState, { type, payload }) => {
  switch (type) {
    case types.SET_CURRENT_USER:
      return { ...state, user: payload }
    case types.SET_CURRENT_USER_FETCH_STATE:
      return { ...state, fetchState: updateFetchState(state.fetchState, payload) }
    case types.SET_USER_PREFERENCES:
      return { ...state, preferences: payload }
    case types.SET_USER_PREFERENCES_FETCH_STATE:
      return { ...state, preferencesFetchState: updateFetchState(state.preferencesFetchState, payload) }
    default:
      return state
  }
}
