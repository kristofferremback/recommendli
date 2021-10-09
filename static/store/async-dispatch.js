/**
 * @typedef action
 * @property {string} type
 * @property {any} payload
 */

/**
 * @typedef {(action) => void} dispatchFunc
 */

/**
 * @template T
 * @typedef {() => T} getStateFunc
 */

/**
 * @template T
 * @typedef {(dispatch: dispatchFunc, getState: getStateFunc<T>) => void} thunk
 */

/**
 * @template T
 * @typedef {(dispatch: dispatchFunc, getState: getStateFunc<T>) => Promise} asyncThunk
 */

const isPromise = (v) => v && typeof v.then === 'function' && typeof v.catch === 'function'

/**
 * @template State
 *
 * @param {dispatchFunc} dispatch
 * @param {getStateFunc<State>} getState
 */
const useAsyncDispatch = (dispatch, getState) => {
  /**
   * @param {action|thunk<State>|asyncThunk<State>} input
   */
  const asyncDispatch = (input) => {
    if (typeof input !== 'function') {
      // "normal" dispatch
      return dispatch(input)
    }

    input(dispatch, getState)
  }
  return asyncDispatch
}

export default useAsyncDispatch
