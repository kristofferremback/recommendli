import { html } from './lib/html.js'
import { render } from './deps/preact.js'
import { useMemo } from './deps/preact/hooks.js'
import App from './app.js'

import { StoreContext, globalReducer, initialState } from './store/store.js'
import useThunkReducer from './store/lib/use-thunk-reducer.js'

const AppContainer = () => {
  const [state, dispatch] = useThunkReducer(globalReducer, initialState)
  const contextValue = useMemo(() => ({ state, dispatch }), [state, dispatch])

  return html`
  <${StoreContext.Provider} value=${contextValue}>
    <${App} />
  </${StoreContext.Provider}>
  `
}

render(html`<${AppContainer} />`, document.body)
