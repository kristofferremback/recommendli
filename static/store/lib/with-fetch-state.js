export const states = {
  new: 'new',
  idle: 'idle',
  loading: 'loading',
  error: 'error',
}

export const defaultFetchState = () => ({
  state: states.new,
  error: null,
  lastUpdatedAt: null,
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
      dispatch(setStateAction({ state: states.loading, lastUpdatedAt: new Date() }))
      await actionFunc(dispatch, getState)
      dispatch(setStateAction({ state: states.idle, lastUpdatedAt: new Date() }))
    } catch (error) {
      dispatch(setStateAction({ state: states.error, lastUpdatedAt: new Date(), error }))
    }
  }
  return thunk
}

/**
 * @param {string} type
 */
export const createSetFetchState = (type) => {
  return ({ state, lastUpdatedAt, error }) => ({
    type,
    payload: { state, lastUpdatedAt, error },
  })
}
