import html from '../../lib/html.js'
import { useContext, useEffect } from '../../deps/preact/hooks.js'
import { StoreContext } from '../../store/store.js'
import withConditionalLoading, { LoadingText } from '../conditional-loading/conditional-loading.component.js'
import UserPreferences from './user-preferences.component.js'

import {
  selectUserIsReady,
  selectUserPreferences,
  selectUserPreferencesAreReady,
} from '../../store/user/user.selectors.js'
import { states } from '../../store/lib/with-fetch-state.js'
import { getUserPreferencesAsync } from '../../store/user/user.actions.js'

const LoadingWrapper = withConditionalLoading(
  () => html`<${LoadingText}>User preferences loading</${LoadingText}>`
)

const UserPreferencesContainer = () => {
  /**
   * @typedef {import('../../store/store').initialState} State
   * @typedef {import('../../store/lib/types').Dispatch} DispatchFunc
   *
   * @type {{ state: State, dispatch: DispatchFunc }}
   */
  const { state, dispatch } = useContext(StoreContext)

  const isLoading = [selectUserIsReady(state), selectUserPreferencesAreReady(state)].some((ready) => !ready)
  const userPreferences = selectUserPreferences(state)

  useEffect(() => {
    if (userPreferences == null && state.user.preferencesFetchState.state === states.new) {
      dispatch(getUserPreferencesAsync())
    }
  }, [userPreferences, state.user.preferencesFetchState.state])

  return html`
    <${LoadingWrapper} isLoading=${isLoading}>
      <${UserPreferences} userPreferences=${userPreferences} />
    </${LoadingWrapper}>
  `
}

export default UserPreferencesContainer
