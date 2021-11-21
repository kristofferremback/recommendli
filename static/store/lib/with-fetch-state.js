export const states = {
  new: 'new',
  idle: 'idle',
  loading: 'loading',
  error: 'error',
}

export const defaultFetchState = () => ({
  state: states.new,
  error: null,
  lastResponseAt: null,
})

/**
 * @param {import('./types').ActionFunc} setStateAction
 * @param {import('./types').AsyncThunk} actionFunc
 * @returns
 */
export const withFetchState = (setStateAction, actionFunc) => {
  /**
   * @param {import('./types').Dispatch} dispatch
   * @param {import('./types').GetState} getState
   */
  const thunk = async (dispatch, getState) => {
    try {
      dispatch(setStateAction({ state: states.loading }))
      await actionFunc(dispatch, getState)
      dispatch(setStateAction({ state: states.idle, lastResponseAt: new Date() }))
    } catch (error) {
      console.log(error)
      dispatch(setStateAction({ state: states.error, lastResponseAt: new Date(), error }))
    }
  }
  return thunk
}

/**
 * @param {string} type
 */
export const createSetFetchState = (type) => {
  return ({ state, lastResponseAt, error }) => ({
    type,
    payload: { state, lastResponseAt, error },
  })
}

export const updateFetchState = (prev, next) => {
  return { ...next, lastResponseAt: next.lastResponseAt || prev.lastResponseAt }
}

export const isReady = (s) => s.state !== states.new && s.lastResponseAt != null
