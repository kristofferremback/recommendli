import { useSyncExternalStore } from 'react'
import type { Playback } from '@/shared/types/spotify'

/*
 * Position of the playing track between polls. progress_ms is where the track
 * was when the response arrived (fetched_at, stamped by the client), so the
 * position now is that plus the time since, while playing. Spotify's own
 * timestamp is the last state change on Spotify's clock and is not used.
 *
 * Every subscriber reads the same tick, so two clocks on the page never
 * disagree on the second.
 */

export function playbackPosition(playback: Playback | undefined, at: number) {
  if (!playback?.track) return 0
  const elapsed = playback.is_playing ? Math.max(0, at - playback.fetched_at) : 0
  return Math.min(playback.progress_ms + elapsed, playback.track.duration_ms ?? Number.MAX_SAFE_INTEGER)
}

const TICK_MS = 250
const listeners = new Set<() => void>()
let now = Date.now()
let timer: number | undefined

function subscribe(listener: () => void) {
  listeners.add(listener)
  if (timer === undefined) {
    now = Date.now()
    timer = window.setInterval(() => {
      now = Date.now()
      listeners.forEach(notify => notify())
    }, TICK_MS)
  }
  return () => {
    listeners.delete(listener)
    if (listeners.size === 0) {
      window.clearInterval(timer)
      timer = undefined
    }
  }
}

const paused = () => () => {}
const snapshot = () => now

export function usePlaybackProgress(playback?: Playback) {
  const at = useSyncExternalStore(playback?.is_playing ? subscribe : paused, snapshot)
  return playbackPosition(playback, at)
}
