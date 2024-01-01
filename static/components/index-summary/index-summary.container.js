import html from '../../lib/html.js'
import { useCallback, useContext } from '../../deps/preact/hooks.js'

import { StoreContext } from '../../store/store.js'

import usePolling from '../../hooks/use-polling.js'
import {
  selectIndexSummary,
  selectIndexSummaryLoading,
  selectIndexSummaryReady,
} from '../../store/generate/generate.selectors.js'
import { selectIsVisible } from '../../store/window/window.selectors.js'
import { getIndexSummaryAsync } from '../../store/generate/generate.actions.js'
import { selectCurrentUser } from '../../store/user/user.selectors.js'

import IndexSummary from './index-summary.component.js'
import withConditionalLoading, { LoadingText } from '../conditional-loading/conditional-loading.component.js'

const LoadingWrapper = withConditionalLoading(
  () => html` <${IndexSummary} title=${html`<${LoadingText}>Index summary loading</${LoadingText}>`} /> `
)

const IndexSummaryContainer = () => {
  /**
   * @typedef {import('../../store/store').initialState} State
   * @typedef {import('../../store/lib/types').Dispatch} DispatchFunc
   *
   * @type {{ state: State, dispatch: DispatchFunc }}
   */
  const { state, dispatch } = useContext(StoreContext)

  const isReady = selectIndexSummaryReady(state)
  const isLoading = selectIndexSummaryLoading(state)

  const isVisible = selectIsVisible(state)
  const currentUser = selectCurrentUser(state)

  const indexSummary = selectIndexSummary(state)

  const pollAction = useCallback(() => dispatch(getIndexSummaryAsync()), [])
  usePolling(pollAction, { isActive: isVisible && currentUser != null, interval: 20_000 })

  return html`
    <${LoadingWrapper} isLoading=${!isReady}>
      <${IndexSummary} isLoading=${isLoading} indexSummary=${indexSummary} />
    </${LoadingWrapper}>
  `
}

export default IndexSummaryContainer
