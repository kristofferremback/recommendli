// @ts-ignore
import { createContext } from '../deps/preact.js'
import { combineReducers } from './lib/combine-reducers.js'
import * as currentTrackReducer from './current-track/current-track.reducer.js'
import * as userReducer from './user/user.reducer.js'
import * as windowReducer from './window/window.reducer.js'
import * as generateReducer from './generate/generate.reducer.js'

export const StoreContext = createContext(undefined)

const reducerMappings = {
  user: {
    initialState: userReducer.initialState,
    reducer: userReducer.reducer,
  },
  currentTrack: {
    initialState: currentTrackReducer.initialState,
    reducer: currentTrackReducer.reducer,
  },
  window: {
    initialState: windowReducer.initialState,
    reducer: windowReducer.reducer,
  },
  generate: {
    initialState: generateReducer.initialState,
    reducer: generateReducer.reducer,
  },
}

/**
 * "Clever" way of creating both combined reducers and combine initial state
 *
 * @typedef {typeof reducerMappings} mapping
 * @typedef {keyof mapping} keys
 *
 * @typedef {{ [key in keys]: mapping[key]['reducer'] }} reducers
 * @typedef {{ [key in keys]: mapping[key]['initialState'] }} initialState
 *
 * @param {typeof reducerMappings} mappings
 * @returns {{ reducers: reducers, initialState: initialState } }}
 */
const combine = (mappings) => {
  // @ts-ignore
  return Object.entries(mappings).reduce(
    ({ reducers, initialState }, [key, { reducer, initialState: initState }]) => ({
      reducers: { ...reducers, [key]: reducer },
      initialState: { ...initialState, [key]: initState },
    }),
    {
      reducers: {},
      initialState: {},
    }
  )
}

const combined = combine(reducerMappings)

export const initialState = combined.initialState

export const globalReducer = combineReducers(combined.reducers)
