import { useMemo } from '../../deps/preact/hooks.js'

const compose = (...funcs) => {
  if (funcs.length === 0) {
    return (arg) => arg
  }

  if (funcs.length === 1) {
    return funcs[0]
  }

  return funcs.reduce((a, b) => {
    return (...args) => a(b(...args))
  })
}

const useMiddleware = (getState, baseDispatch, middlewares) => {
  return useMemo(() => {
    /** @type {import('../lib/types').Dispatch} */
    let dispatch = () => {
      throw new Error("This dispatch should't be used")
    }

    const middlewareAPI = {
      getState,
      /** @type {import('../lib/types').Dispatch} */
      dispatch: (action, ...args) => dispatch(action, ...args),
    }

    const chain = middlewares.map((middleware) => middleware(middlewareAPI))
    dispatch = compose(...chain)(baseDispatch)

    return dispatch
  }, [baseDispatch])
}

export default useMiddleware
