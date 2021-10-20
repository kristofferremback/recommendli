import { withFetchState, createSetFetchState } from '../lib/with-fetch-state.js'
import withRecommendliClient from '../../recommendli/client.js'

export const types = {
  SET_CURRENT_TRACK: 'SET_CURRENT_TRACK',
  SET_CURRENT_TRACK_FETCH_STATE: 'SET_CURRENT_TRACK_FETCH_STATE',
}

const setCurrentTrack = ({ track, isPlaying }) => ({
  type: types.SET_CURRENT_TRACK,
  payload: { track, isPlaying },
})

const setCurrentTrackFetchState = createSetFetchState(types.SET_CURRENT_TRACK_FETCH_STATE)

export const getCurrentTrackAsync = () => {
  return withRecommendliClient((client) => {
    return withFetchState(setCurrentTrackFetchState, async (dispatch) => {
      const { track, isPlaying } = await client.getCurrentTrack()

      dispatch(setCurrentTrack({ isPlaying, track }))
    })
  })
}
