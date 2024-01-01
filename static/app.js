import { useContext, useEffect, useCallback } from './deps/preact/hooks.js'
import html from './lib/html.js'
import { getCurrentUserAsync } from './store/user/user.actions.js'

import { StoreContext } from './store/store.js'

import { states } from './store/lib/with-fetch-state.js'
import { selectCurrentUser } from './store/user/user.selectors.js'

import DiscoveryPlaylist from './components/discovery-playlist/discovery-playlist.component.js'

import { generateDiscoveryPlaylistAsync, getIndexSummaryAsync } from './store/generate/generate.actions.js'
import { selectDiscoveryIsLoading, selectDiscoveryPlaylist } from './store/generate/generate.selectors.js'
import PlayingContainer from './components/playing/playing.container.js'
import IndexSummaryContainer from './components/index-summary/index-summary.container.js'

const App = () => {
  /**
   * @typedef {typeof import('./store/store').initialState} State
   * @typedef {import('./store/lib/types').Dispatch} Dispatch
   *
   * @type {{ state: State, dispatch: Dispatch }}
   */
  const { state, dispatch } = useContext(StoreContext)

  const currentUser = selectCurrentUser(state)

  const discoveryIsLoading = selectDiscoveryIsLoading(state)
  const discoveryPlaylist = selectDiscoveryPlaylist(state)

  useEffect(() => {
    if (currentUser == null && state.user.fetchState.state === states.new) {
      dispatch(getCurrentUserAsync())
    }
  }, [currentUser, state.user.fetchState.state])

  const onGeneratePlaylist = useCallback(() => dispatch(generateDiscoveryPlaylistAsync()), [])

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
        <${PlayingContainer} />
        <${IndexSummaryContainer} />
      </div>
      <div>
        <details>
          <summary>Show entire state</summary>
          <pre>${JSON.stringify(state, null, 4)}</pre>
        </details>
      </div>
    </div>
  `
}

export default App
