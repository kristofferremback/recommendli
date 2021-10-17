export const defaultFetchState = () => ({
  state: 'idle',
  error: null,
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
      dispatch(setStateAction({ state: 'loading' }))
      await actionFunc(dispatch, getState)
      dispatch(setStateAction({ state: 'idle' }))
    } catch (error) {
      dispatch(setStateAction({ state: 'error', error }))
    }
  }
  return thunk
}
