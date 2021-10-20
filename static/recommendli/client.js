import { throwOn404, redirectingFetch } from './redirecting-fetch.js'

const recommendliClient = {
  getCurrentTrack: async () => {
    const response = await throwOn404(redirectingFetch('/recommendations/v1/current-track'))
    const { track, is_playing: isPlaying } = await response.json()
    return { track, isPlaying }
  },
  getCurrentUser: async () => {
    const response = await throwOn404(redirectingFetch('/recommendations/v1/whoami'))
    return await response.json()
  },
}

/**
 *
 * @template T
 * @param {(client: typeof recommendliClient) => T} func
 */
const withRecommendliClient = (func) => func(recommendliClient)

export default withRecommendliClient
