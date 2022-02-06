import createSelector from '../lib/create-selector.js'
import { isReady } from '../lib/with-fetch-state.js'

const selectUser = (state) => state.user

const selectUserFetchState = createSelector([selectUser], (user) => user.fetchState)

export const selectCurrentUser = createSelector([selectUser], (user) => user.user)

export const selectUserIsReady = createSelector([selectUserFetchState], isReady)
