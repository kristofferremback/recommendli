// @ts-ignore
import { useMemo, useReducer, useCallback, useRef } from 'https://unpkg.com/htm/preact/standalone.module.js'
import withAsyncDispatch from './async-dispatch.js'

const useThunkReducer = (reducer, initialState, initializer) => {
  const lastState = useRef(initialState)
  const getState = useCallback(() => lastState.current, [])

  const [state, dispatch] = useReducer(
    (state, action) => (lastState.current = reducer(state, action)),
    initialState,
    initializer
  )
  const asyncDispatch = useMemo(() => withAsyncDispatch(dispatch, getState), [dispatch, getState])

  return [state, asyncDispatch]
}

export default useThunkReducer
