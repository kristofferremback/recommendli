import { redirectingFetch, throwOn404 } from '../spotify/redirecting-fetch.js'

export const types = {
  SET_CURRENT_USER: 'SET_CURRENT_USER',
  SET_CURRENT_USER_FETCH_STATE: 'SET_CURRENT_USER_FETCH_STATE',
  SET_CURRENT_TRACK: 'SET_CURRENT_TRACK',
  SET_CURRENT_TRACK_FETCH_STATE: 'SET_CURRENT_TRACK_FETCH_STATE',
}

/**
 * @template State
 *
 * @param {import('./async-dispatch').actionFunc} setStateAction
 * @param {import('./async-dispatch').asyncThunk<State>} actionFunc
 * @returns
 */
const withState = (setStateAction, actionFunc) => {
  /**
   * @param {import('./async-dispatch').dispatchFunc} dispatch
   * @param {import('./async-dispatch').getStateFunc<any>} getState
   */
  const thunk = async (dispatch, getState) => {
    try {
      dispatch(setStateAction({ state: 'loading' }))
      await actionFunc(dispatch, getState)
      dispatch(setStateAction({ state: 'idle' }))
    } catch (error) {
      dispatch(setStateAction({ state: 'error', error }))
    }
  }
  return thunk
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
  return withState(setCurrentUserFetchState, async (dispatch) => {
    const response = await throwOn404(redirectingFetch('/recommendations/v1/whoami'))
    const user = await response.json()

    dispatch(setCurrentUser(user))
  })
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
  return withState(setCurrentTrackFetchState, async (dispatch) => {
    const response = await throwOn404(redirectingFetch('/recommendations/v1/current-track'))
    const currentTrack = await response.json()

    dispatch(setCurrentTrack(currentTrack))
  })
}
