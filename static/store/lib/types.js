/**
 *
 * @typedef {object} Action
 * @property {string} type
 * @property {any} payload
 *
 *
 * @typedef {(state: any, action: Action) => any} ReducerFunc
 *
 * @typedef {(action: Action) => void} dispatchFunc
 *
 * @typedef {(state: any) => any} GetState
 *
 * @typedef {(dispatch: Dispatch, getState: GetState) => void} thunk
 * @typedef {(dispatch: Dispatch, getState: GetState) => Promise<any>} asyncThunk
 *
 * @typedef {(dispatch: Dispatch) => void} simpleThunk
 * @typedef {(dispatch: Dispatch) => Promise<any>} simpleAsyncThunk
 *
 * @typedef {thunk | simpleThunk} Thunk
 * @typedef {asyncThunk | simpleAsyncThunk} AsyncThunk

 * @typedef {(input: Action | Thunk | AsyncThunk) => void} Dispatch

 * @typedef {(...args: any) => Action} ActionFunc
 */

export default {}
