import { withFetchState } from '../lib/with-fetch-state.js'
import { redirectingFetch, throwOn404 } from '../../spotify/redirecting-fetch.js'

export const types = {
  SET_CURRENT_USER: 'SET_CURRENT_USER',
  SET_CURRENT_USER_FETCH_STATE: 'SET_CURRENT_USER_FETCH_STATE',
}

const setCurrentUser = (currentUser) => ({
  type: types.SET_CURRENT_USER,
  payload: currentUser,
})

/**
 * @param {object} opts
 * @param {'idle'|'loading'|'error'} opts.state
 * @param {Error} [opts.error]
 */
const setCurrentUserFetchState = ({ state, error }) => ({
  type: types.SET_CURRENT_USER_FETCH_STATE,
  payload: { state, error },
})

export const getCurrentUserAsync = () => {
  return withFetchState(setCurrentUserFetchState, async (dispatch) => {
    const response = await throwOn404(redirectingFetch('/recommendations/v1/whoami'))
    const user = await response.json()

    dispatch(setCurrentUser(user))
  })
}
