// @ts-ignore
import { html, useContext, useEffect, useState } from 'https://unpkg.com/htm/preact/standalone.module.js'
import { getCurrentUserAsync } from './store/user/user.actions.js'
import { getCurrentTrackAsync } from './store/current-track/current-track.actions.js'
import { setVisibilityState } from './store/window/window.actions.js'

import { StoreContext } from './store/store.js'

import Playing from './components/playing/playing.js'
import { selectIsVisible } from './store/window/window.selectors.js'

const App = () => {
  /**
   * @typedef {typeof import('./store/store').initialState} State
   * @typedef {import('./store/lib/async-dispatch').asyncDispatchFunc<any>} asyncDispatchFunc
   *
   * @type {{ state: State, dispatch: asyncDispatchFunc }}
   */
  const { state, dispatch } = useContext(StoreContext)
  const isVisible = selectIsVisible(state)

  useEffect(() => {
    if (state.user == null && state.user.fetchState.state === 'idle') {
      dispatch(getCurrentUserAsync())
    }
  }, [state.user, state.user.fetchState.state])

  useEffect(() => {
    const eventListeners = { visibilitychange: () => dispatch(setVisibilityState(document.visibilityState)) }
    for (const [event, listener] of Object.entries(eventListeners)) {
      document.addEventListener(event, listener)
    }

    return () => {
      for (const [event, listener] of Object.entries(eventListeners)) {
        document.removeEventListener(event, listener)
      }
    }
  }, [])

  useEffect(() => {
    const timers = new Set()

    const startPolling = () => {
      dispatch(getCurrentTrackAsync())
      timers.add(setInterval(() => dispatch(getCurrentTrackAsync()), 2000))
    }
    const cleanup = () => {
      for (const timerHandle of timers.values()) {
        timers.delete(timerHandle)
        clearInterval(timerHandle)
      }
    }

    if (isVisible && state.user != null) {
      cleanup()
      startPolling()
    } else {
      cleanup()
    }

    return cleanup
  }, [isVisible, state.user])

  return html`
    <div class="app">
      <${Playing} track=${state.currentTrack.track} isPlaying=${state.currentTrack.isPlaying} />
    </div>
  `
}

export default App
