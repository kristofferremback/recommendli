import { useMemo, useReducer } from '../../deps/preact/hooks.js'
import useMiddleware from './use-middleware.js'
import { useGetState } from './use-get-state.js'

const createUseStore = (StoreContext) => {
  /**
   * @param {*} globalReducer
   * @param {*} initialState
   * @param {Function[]} middlewares
   * @returns
   */
  return (globalReducer, initialState, middlewares) => {
    const [state, baseDispatch] = useReducer(globalReducer, initialState)

    const getState = useGetState(state)
    const dispatch = useMiddleware(getState, baseDispatch, middlewares)

    const contextValue = useMemo(() => ({ state, dispatch }), [state, dispatch])

    return { state, dispatch, StoreContext, contextValue }
  }
}

export default createUseStore
