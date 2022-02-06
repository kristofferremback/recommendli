export const types = {
  SET_WINDOW_VISIBILITY_STATE: 'SET_WINDOW_VISIBILITY_STATE',
}

export const setVisibilityState = (visibilityState) => ({
  type: types.SET_WINDOW_VISIBILITY_STATE,
  payload: visibilityState,
})
