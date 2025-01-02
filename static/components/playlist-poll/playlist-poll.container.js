import html from '../../lib/html.js'
import { useState, useCallback, useContext } from '../../deps/preact/hooks.js'

import { StoreContext } from '../../store/store.js'

import usePolling from '../../hooks/use-polling.js'
import { selectIsVisible } from '../../store/window/window.selectors.js'
import recommendliClient from '../../recommendli/client.js'
import { defaultFetchState, states } from '../../store/lib/with-fetch-state.js'
import PlaylistPoll from './playlist-poll.component.js'

const PlaylistPollContainer = ({}) => {
  /**
   * @typedef {import('../../store/store').initialState} State
   * @typedef {import('../../store/lib/types').Dispatch} DispatchFunc
   *
   * @type {{ state: State, dispatch: DispatchFunc }}
   */
  const { state, dispatch } = useContext(StoreContext)

  const [playlistId, setPlaylistId] = useState(null)
  const [playlist, setPlaylist] = useState(null)

  const [fetchState, setFetchState] = useState(states.new)
  const [fetchError, setFetchError] = useState(null)
  const [lastResponseAt, setLastResponseAt] = useState(null)

  const isVisible = selectIsVisible(state)
  const pollAction = useCallback(async () => {
    setFetchState(states.loading)
    try {
      const pl = await recommendliClient.getPlaylistById(playlistId)
      setPlaylist(pl)
      setFetchState(states.idle)
      setFetchError(null)
    } catch (error) {
      setFetchError(error)
      setFetchState(states.error)
    }
  }, [playlistId])

  usePolling(pollAction, { isActive: isVisible && !!playlistId, interval: 1000 })

  return html`
    <${PlaylistPoll} playlist=${playlist} playlistId=${playlistId} setPlaylistId=${setPlaylistId} />
  `
}

export default PlaylistPollContainer
