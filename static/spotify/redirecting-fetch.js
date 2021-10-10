/**
 * @param {RequestInfo} input
 * @param {RequestInit} [init]
 * @returns Promise<Response>
 */
export const redirectingFetch = async (input, init = {}) => {
  const response = await fetch(input, { ...init, redirect: 'manual' })
  if (response.type === 'opaqueredirect') {
    location.replace(
      `/recommendations/v1/spotify/auth/ui-redirect?url=${encodeURIComponent(window.location.href)}`
    )
  }
  return response
}

/**
 * @param {Promise<Response>} promise
 * @returns Promise<Response>
 */
export const throwOn404 = async (promise) => {
  const response = await promise
  if (response.status < 400) {
    return response
  }

  const error = new Error(`Unexpected status ${response.status} - ${response.statusText}`)
  if (response.headers.get('content-type') === 'application/json') {
    const body = await response.json()
    // @ts-ignore
    error.body = body
  }
  throw error
}
