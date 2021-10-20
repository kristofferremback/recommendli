import { defaultFetchState } from '../lib/with-fetch-state.js'
import { types } from './user.actions.js'

export const initialState = {
  user: null,
  fetchState: defaultFetchState(),
}

export const reducer = (state = initialState, { type, payload }) => {
  switch (type) {
    case types.SET_CURRENT_USER:
      return { ...state, user: payload }
    case types.SET_CURRENT_USER_FETCH_STATE:
      return { ...state, fetchState: payload }
    default:
      return state
  }
}
