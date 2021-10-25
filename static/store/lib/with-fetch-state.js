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
 * @template State
 *
 * @param {import('./async-dispatch').actionFunc} setStateAction
 * @param {import('./async-dispatch').asyncThunk<State>} actionFunc
 * @returns
 */
export const withFetchState = (setStateAction, actionFunc) => {
  /**
   * @param {import('./async-dispatch').dispatchFunc} dispatch
   * @param {import('./async-dispatch').getStateFunc<any>} getState
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
