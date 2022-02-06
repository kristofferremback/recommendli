import { types } from './window.actions.js'

export const initialState = {
  visibilityState: 'visible',
}

export const reducer = (state = initialState, { type, payload }) => {
  switch (type) {
    case types.SET_WINDOW_VISIBILITY_STATE:
      return { ...state, visibilityState: payload }
    default:
      return state
  }
}
