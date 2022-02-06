import createSelector from '../lib/create-selector.js'
import { states } from '../lib/with-fetch-state.js'

const selectGenerate = (state) => state.generate

const selectDiscovery = createSelector([selectGenerate], (generate) => generate.discovery)

export const selectDiscoveryIsLoading = createSelector(
  [selectDiscovery],
  (discovery) => discovery.fetchState.state === states.loading
)

export const selectDiscoveryPlaylist = createSelector([selectDiscovery], (discovery) => discovery.playlist)
