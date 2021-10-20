// @ts-ignore
import { html, useContext, useEffect, useCallback } from 'https://unpkg.com/htm/preact/standalone.module.js'
import { getCurrentUserAsync } from './store/user/user.actions.js'
import { getCurrentTrackAsync } from './store/current-track/current-track.actions.js'
import { setVisibilityState } from './store/window/window.actions.js'

import { StoreContext } from './store/store.js'

import Playing from './components/playing/playing.js'
import { selectIsVisible } from './store/window/window.selectors.js'
import usePolling from './hooks/use-polling.js'
import { selectCurrentUser } from './store/user/user.selectors.js'
import useEventListener from './hooks/use-event-listener.js'
import { selectIsPlaying, selectTrack } from './store/current-track/current-track.selectors.js'
import { states } from './store/lib/with-fetch-state.js'

const App = () => {
  /**
   * @typedef {typeof import('./store/store').initialState} State
   * @typedef {import('./store/lib/async-dispatch').asyncDispatchFunc<State>} asyncDispatchFunc
   *
   * @type {{ state: State, dispatch: asyncDispatchFunc }}
   */
  const { state, dispatch } = useContext(StoreContext)

  const isVisible = selectIsVisible(state)
  const currentUser = selectCurrentUser(state)
  const isPlaying = selectIsPlaying(state)
  const track = selectTrack(state)

  useEffect(() => {
    if (currentUser == null && state.user.fetchState.state === states.new) {
      dispatch(getCurrentUserAsync())
    }
  }, [currentUser, state.user.fetchState.state])

  const onVisibilityChange = useCallback(() => dispatch(setVisibilityState(document.visibilityState)), [])
  useEventListener('visibilitychange', onVisibilityChange)

  const action = useCallback(() => dispatch(getCurrentTrackAsync()), [])
  usePolling(action, { isActive: isVisible && currentUser != null })

  return html`
    <div class="app">
      <${Playing} track=${track} isPlaying=${isPlaying} />
      <pre>${JSON.stringify(state, null, 4)}</pre>
    </div>
  `
}

export default App
