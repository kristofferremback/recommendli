// @ts-ignore
import { html, render, useMemo, useReducer } from 'https://unpkg.com/htm/preact/standalone.module.js'

import App from './app.js'

import { StoreContext, globalReducer, initialState } from './store/store.js'
import useAsyncDispatch from './store/lib/async-dispatch.js'

const AppContainer = () => {
  const [state, dispatch] = useReducer(globalReducer, initialState)
  const asyncDispatch = useAsyncDispatch(dispatch, () => state)

  const contextValue = useMemo(() => ({ state, dispatch: asyncDispatch }), [state, asyncDispatch])

  return html`
  <${StoreContext.Provider} value=${contextValue}>
    <${App} />
  </${StoreContext.Provider}>
  `
}

render(html`<${AppContainer} />`, document.body)
