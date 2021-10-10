import { html, useContext, useEffect } from 'https://unpkg.com/htm/preact/standalone.module.js'
import { getCurrentTrackAsync, getCurrentUserAsync } from './store/user.js'

import { StoreContext } from './store/store.js'

const App = () => {
  const { state, dispatch } = useContext(StoreContext)

  useEffect(() => {
    if (state.user == null && state.userFetch.state === 'idle') {
      dispatch(getCurrentUserAsync())
    }
  }, [state.user, state.userFetch.state])

  useEffect(() => {
    let timerHandle = null
    if (state.user != null) {
      dispatch(getCurrentTrackAsync())
      // Perhaps do something better than setInterval here?
      timerHandle = setInterval(() => dispatch(getCurrentTrackAsync()), 2000)
    }

    return () => {
      if (timerHandle != null) {
        clearInterval(timerHandle)
      }
    }
  }, [state.user])

  return html`
    <div class="app">
      <pre>${JSON.stringify(state, null, 4)}</pre>
    </div>
  `
}

export default App
