import { withFetchState, createSetFetchState } from '../lib/with-fetch-state.js'
import recommendliClient from '../../recommendli/client.js'

export const types = {
  SET_DISCOVERY_PLAYLIST: 'SET_DISCOVERY_PLAYLIST',
  SET_DISCOVERY_FETCH_STATE: 'SET_DISCOVERY_PLAYLIST_FETCH_STATE',
}

const setDiscoveryPlaylist = (playlist) => ({
  type: types.SET_DISCOVERY_PLAYLIST,
  payload: playlist,
})

const setDiscoveryFetchState = createSetFetchState(types.SET_DISCOVERY_FETCH_STATE)

export const generateDiscoveryPlaylistAsync = ({ dryRun = false } = {}) => {
  return withFetchState(setDiscoveryFetchState, async (dispatch) => {
    const playlist = await recommendliClient.generateDiscoveryPlaylist({ dryRun })
    dispatch(setDiscoveryPlaylist(playlist))
  })
}
