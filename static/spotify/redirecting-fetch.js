/**
 * @param {RequestInfo} input
 * @param {RequestInit} [init]
 * @returns Promise<Response>
 */
const redirectingFetch = async (input, init = {}) => {
  const response = await fetch(input, { ...init, redirect: 'manual' })
  if (response.type === 'opaqueredirect') {
    location.replace(
      `/recommendations/v1/spotify/auth/ui-redirect?url=${encodeURIComponent(window.location.href)}`
    )
  }
  return response
}

export default redirectingFetch
