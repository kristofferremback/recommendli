import { withFetchState, createSetFetchState } from '../lib/with-fetch-state.js'
import recommendliClient from '../../recommendli/client.js'

export const types = {
  SET_CURRENT_USER: 'SET_CURRENT_USER',
  SET_CURRENT_USER_FETCH_STATE: 'SET_CURRENT_USER_FETCH_STATE',
  SET_USER_PREFERENCES: 'SET_USER_PREFERENCES',
  SET_USER_PREFERENCES_FETCH_STATE: 'SET_USER_PREFERENCES_FETCH_STATE',
}

const setCurrentUser = (currentUser) => ({
  type: types.SET_CURRENT_USER,
  payload: currentUser,
})

const setUserPreferences = (prefs) => ({
  type: types.SET_USER_PREFERENCES,
  payload: prefs,
})

const setCurrentUserFetchState = createSetFetchState(types.SET_CURRENT_USER_FETCH_STATE)

export const getCurrentUserAsync = () => {
  return withFetchState(setCurrentUserFetchState, async (dispatch) => {
    const user = await recommendliClient.getCurrentUser()

    dispatch(setCurrentUser(user))
  })
}

const setUserPreferencesFetchState = createSetFetchState(types.SET_USER_PREFERENCES_FETCH_STATE)

export const getUserPreferencesAsync = ()=> {
  return withFetchState(setUserPreferencesFetchState, async (dispatch) => {
    const prefs = await recommendliClient.getUserPreferences()

    dispatch(setUserPreferences(prefs))
  })
}
