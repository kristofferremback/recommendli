// @ts-ignore
import { html, useContext, useEffect, useCallback } from 'https://unpkg.com/htm/preact/standalone.module.js'
import { getCurrentUserAsync } from './store/user/user.actions.js'
import { getCurrentTrackAsync } from './store/current-track/current-track.actions.js'
import { setVisibilityState } from './store/window/window.actions.js'

import { StoreContext } from './store/store.js'

import { states } from './store/lib/with-fetch-state.js'
import { selectIsVisible } from './store/window/window.selectors.js'
import { selectCurrentUser } from './store/user/user.selectors.js'
import useEventListener from './hooks/use-event-listener.js'
import { selectIsPlaying, selectTrack } from './store/current-track/current-track.selectors.js'

import usePolling from './hooks/use-polling.js'

import Playing from './components/playing/playing.component.js'
import DiscoveryPlaylist from './components/discovery-playlist/discovery-playlist.component.js'

import withConditionalLoading, {
  LoadingText,
} from './components/conditional-loading/conditional-loading.component.js'
import { generateDiscoveryPlaylistAsync } from './store/generate/generate.actions.js'
import { selectDiscoveryIsLoading, selectDiscoveryPlaylist } from './store/generate/generate.selectors.js'

const LoadingPlayer = withConditionalLoading(
  () => html`
    <${Playing} isPlaying=${false} title=${html`<${LoadingText}>Player state loading</${LoadingText}>`} />
  `
)

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
  const discoveryIsLoading = selectDiscoveryIsLoading(state)
  const discoveryPlaylist = selectDiscoveryPlaylist(state)

  useEffect(() => {
    if (currentUser == null && state.user.fetchState.state === states.new) {
      dispatch(getCurrentUserAsync())
    }
  }, [currentUser, state.user.fetchState.state])

  const onVisibilityChange = useCallback(() => dispatch(setVisibilityState(document.visibilityState)), [])
  useEventListener('visibilitychange', onVisibilityChange)

  const pollAction = useCallback(() => dispatch(getCurrentTrackAsync()), [])
  usePolling(pollAction, { isActive: isVisible && currentUser != null })

  const onGeneratePlaylist = useCallback(() => dispatch(generateDiscoveryPlaylistAsync({ dryRun: true })), [])

  const isReady = [state.currentTrack.fetchState, state.user.fetchState].every(
    (s) => s.state !== states.new && s.lastResponseAt != null
  )

  return html`
    <div class="container-fluid">
      <nav>
        <ul>
          <li><strong>Recommendli</strong></li>
        </ul>
        </nav>
        <div class="grid">
          <${DiscoveryPlaylist}
            onGeneratePlaylist=${onGeneratePlaylist}
            isLoading=${discoveryIsLoading}
            playlist=${discoveryPlaylist}
          />
          <${LoadingPlayer} isLoading=${!isReady}>
            <${Playing} track=${track} isPlaying=${isPlaying} />
          </${LoadingPlayer}>
        </div>
      <pre>${JSON.stringify(state, null, 4)}</pre>
    </div>
  `
}

export default App
