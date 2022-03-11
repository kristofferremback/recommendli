import html from '../../lib/html.js'
import { useCallback, useEffect } from '../../deps/preact/hooks.js'
import { useContext } from '../../deps/preact/hooks.js'

import { StoreContext } from '../../store/store.js'
import Playing from './playing.component.js'
import withConditionalLoading, { LoadingText } from '../conditional-loading/conditional-loading.component.js'
import useEventListener from '../../hooks/use-event-listener.js'

import {
  selectIsPlaying,
  selectStatusTrackId,
  selectStatusTrackIsLoading,
  selectTrack,
  selectTrackId,
  selectTrackInLibrary,
  selectTrackIsReady,
  selectTrackPlaylists,
} from '../../store/current-track/current-track.selectors.js'
import { selectCurrentUser, selectUserIsReady } from '../../store/user/user.selectors.js'
import {
  checkCurrenTrackStatusAsync,
  getCurrentTrackAsync,
} from '../../store/current-track/current-track.actions.js'
import usePolling from '../../hooks/use-polling.js'
import { setVisibilityState } from '../../store/window/window.actions.js'
import { selectIsVisible } from '../../store/window/window.selectors.js'

const LoadingWrapper = withConditionalLoading(
  () => html`
    <${Playing} isPlaying=${false} title=${html`<${LoadingText}>Player state loading</${LoadingText}>`} />
  `
)

const PlayingContainer = () => {
  /**
   * @typedef {import('../../store/store').initialState} State
   * @typedef {import('../../store/lib/types').Dispatch} DispatchFunc
   *
   * @type {{ state: State, dispatch: DispatchFunc }}
   */
  const { state, dispatch } = useContext(StoreContext)

  const isLoading = [selectTrackIsReady(state), selectUserIsReady(state)].some((ready) => !ready)

  const isVisible = selectIsVisible(state)
  const currentUser = selectCurrentUser(state)

  const track = selectTrack(state)
  const trackId = selectTrackId(state)

  const statusTrackId = selectStatusTrackId(state)
  const trackInLibrary = selectTrackInLibrary(state)
  const trackPlaylists = selectTrackPlaylists(state)
  const satusTrackIsLoading = selectStatusTrackIsLoading(state)

  const isPlaying = selectIsPlaying(state)

  const onVisibilityChange = useCallback(() => dispatch(setVisibilityState(document.visibilityState)), [])
  useEventListener('visibilitychange', onVisibilityChange)

  const pollAction = useCallback(() => dispatch(getCurrentTrackAsync()), [])
  usePolling(pollAction, { isActive: isVisible && currentUser != null })

  useEffect(() => {
    if (isPlaying && trackId !== statusTrackId) {
      dispatch(checkCurrenTrackStatusAsync())
    }
  }, [isPlaying, trackId, statusTrackId])

  return html`
    <${LoadingWrapper} isLoading=${isLoading}>
      <${Playing} track=${track} isPlaying=${isPlaying} inLibraryLoading=${satusTrackIsLoading} inLibrary=${trackInLibrary} playlists=${trackPlaylists} />
    </${LoadingWrapper}>
  `
}

export default PlayingContainer
