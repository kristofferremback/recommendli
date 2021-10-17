/**
 * @typedef {(state: any, action: { type: string, payload: any }) => any} ReducerFunc
 * @param {{ [key: string]: ReducerFunc }} inputReducers
 */
export const combineReducers = (inputReducers) => {
  const reducerEntries = Object.entries(inputReducers)

  /**
   * @param {any} state
   * @param {{ type: string, payload: any }} action
   */
  const combinedState = (state, action) => {
    const nextState = {}

    let hasChanged = false
    for (const [key, reducer] of reducerEntries) {
      const prevState = state[key]
      nextState[key] = reducer(prevState, action)
      hasChanged = hasChanged || prevState !== nextState[key]
    }

    return hasChanged ? nextState : state
  }

  return combinedState
}
