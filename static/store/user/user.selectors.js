import createSelector from '../lib/create-selector.js'
import { isReady } from '../lib/with-fetch-state.js'

const selectUser = (state) => state.user

const selectUserFetchState = createSelector([selectUser], (user) => user.fetchState)

const selectUserPreferencesFetchState = createSelector([selectUser], (user) => user.preferencesFetchState)

export const selectCurrentUser = createSelector([selectUser], (user) => user.user)

export const selectUserPreferences = createSelector([selectUser], (user) => user.preferences)

export const selectUserIsReady = createSelector([selectUserFetchState], isReady)

export const selectUserPreferencesAreReady = createSelector([selectUserPreferencesFetchState], isReady)
