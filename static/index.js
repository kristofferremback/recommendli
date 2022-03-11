import html from './lib/html.js'
import { render } from './deps/preact.js'
import { useMemo } from './deps/preact/hooks.js'
import App from './app.js'

import { globalReducer, initialState, useStore } from './store/store.js'
import loggerMiddleware from './store/lib/middleware/logger.middleware.js'
import thunkMiddleware from './store/lib/middleware/thunk.middleware.js'

const AppContainer = () => {
  const { StoreContext, contextValue } = useStore(globalReducer, initialState, [
    thunkMiddleware,
    // loggerMiddleware,
  ])

  return html`
    <${StoreContext.Provider} value=${contextValue}>
      <${App} />
    </${StoreContext.Provider}>
  `
}

render(html`<${AppContainer} />`, document.body)
