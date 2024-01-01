import { defaultFetchState, updateFetchState } from '../lib/with-fetch-state.js'
import { types } from './generate.actions.js'

export const initialState = {
  discovery: {
    playlist: null,
    fetchState: defaultFetchState(),
  },
  indexSummary: {
    unique_track_count: 0,
    playlist_count: 0,
    /** @type {import('../../recommendli/client.js').SimplePlaylist[]} */
    playlists: [],
    fetchState: defaultFetchState(),
  },
}

export const reducer = (state = initialState, { type, payload }) => {
  switch (type) {
    case types.SET_DISCOVERY_PLAYLIST:
      return {
        ...state,
        discovery: { ...state.discovery, playlist: payload },
      }
    case types.SET_DISCOVERY_FETCH_STATE:
      return {
        ...state,
        discovery: {
          ...state.discovery,
          fetchState: updateFetchState(state.discovery.fetchState, payload),
        },
      }
    case types.SET_INDEX_FETCH_STATE:
      return {
        ...state,
        indexSummary: {
          ...state.indexSummary,
          fetchState: updateFetchState(state.indexSummary.fetchState, payload),
        },
      }
    case types.SET_SUMMARY_INDEX:
      return {
        ...state,
        indexSummary: {
          ...state.indexSummary,
          unique_track_count: payload.unique_track_count,
          playlist_count: payload.playlist_count,
          playlists: payload.playlists,
        },
      }
    default:
      return state
  }
}
