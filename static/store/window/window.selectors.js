import createSelector from '../lib/create-selector.js'

const selectWindow = (state) => state.window

export const selectIsVisible = createSelector(
  [selectWindow],
  (window) => window.visibilityState === 'visible'
)
