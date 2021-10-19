import createSelector from '../lib/create-selector.js'

const selectUser = (state) => state.user

export const selectCurrentUser = createSelector([selectUser], (user) => user.user)
