import createSelector from '../lib/create-selector.js'
import { isReady, states } from '../lib/with-fetch-state.js'

const selectGenerate = (state) => state.generate

const selectDiscovery = createSelector([selectGenerate], (generate) => generate.discovery)

export const selectIndexSummary = createSelector([selectGenerate], (generate) => generate.indexSummary)

export const selectIndexSummaryReady = createSelector([selectIndexSummary], (s) => isReady(s.fetchState))

export const selectIndexSummaryLoading = createSelector(
  [selectIndexSummary],
  (indexSummary) => indexSummary.fetchState.state === states.loading
)

export const selectDiscoveryIsLoading = createSelector(
  [selectDiscovery],
  (discovery) => discovery.fetchState.state === states.loading
)

export const selectDiscoveryPlaylist = createSelector([selectDiscovery], (discovery) => discovery.playlist)
