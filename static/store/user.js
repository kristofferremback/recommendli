import redirectingFetch from '../spotify/redirecting-fetch.js'

export const types = {
  SET_CURRENT_USER: 'SET_CURRENT_USER',
  SET_CURRENT_USER_FETCH_STATE: 'SET_CURRENT_USER_FETCH_STATE',
  SET_CURRENT_TRACK: 'SET_CURRENT_TRACK',
  SET_CURRENT_TRACK_FETCH_STATE: 'SET_CURRENT_TRACK_FETCH_STATE',
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
  return async (dispatch) => {
    dispatch(setCurrentUserFetchState({ state: 'loading' }))
    try {
      const response = await redirectingFetch('/recommendations/v1/whoami')
      await checkError(response)
      const user = await response.json()

      dispatch(setCurrentUser(user))
      dispatch(setCurrentUserFetchState({ state: 'idle' }))
    } catch (error) {
      dispatch(setCurrentUserFetchState({ state: 'error', error }))
    }
  }
}

const setCurrentTrack = (currentTrack) => ({
  type: types.SET_CURRENT_TRACK,
  payload: currentTrack,
})

/**
 * @param {object} opts
 * @param {'idle'|'loading'|'error'} opts.state
 * @param {Error} [opts.error]
 */
const setCurrentTrackFetchState = ({ state, error }) => ({
  type: types.SET_CURRENT_TRACK_FETCH_STATE,
  payload: { state, error },
})

export const getCurrentTrackAsync = () => {
  return async (dispatch) => {
    dispatch(setCurrentTrackFetchState({ state: 'loading' }))
    try {
      const response = await redirectingFetch('/recommendations/v1/current-track')
      await checkError(response)
      const currentTrack = await response.json()
      console.log(currentTrack)

      dispatch(setCurrentTrack(currentTrack))
      dispatch(setCurrentTrackFetchState({ state: 'idle' }))
    } catch (error) {
      dispatch(setCurrentTrackFetchState({ state: 'error', error }))
    }
  }
}

/**
 * @param {Response} response
 */
async function checkError(response) {
  if (response.status >= 400) {
    const error = new Error(`Unexpected status ${response.status} - ${response.statusText}`)
    if (response.headers.get('content-type') === 'application/json') {
      const body = await response.json()
      // @ts-ignore
      error.body = body
    }
    throw error
  }
}
