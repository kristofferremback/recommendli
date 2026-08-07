import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from './client'
import type { Playback, QueueSkipRequest, Track } from '@/shared/types/spotify'

export const queryKeys = {
  user: ['user'] as const,
  currentTrack: ['currentTrack'] as const,
  trackStatus: (id?: string) => ['trackStatus', id] as const,
  libraryStatus: (id?: string) => ['libraryStatus', id] as const,
  indexSummary: ['indexSummary'] as const,
  playback: ['playback'] as const,
  playbackQueue: ['playbackQueue'] as const,
  playbackHistory: ['playbackHistory'] as const,
}

export function useCurrentUser() {
  return useQuery({
    queryKey: queryKeys.user,
    queryFn: api.getCurrentUser,
  })
}

export function useCurrentTrack(enabled: boolean, refetchInterval: number | false) {
  return useQuery({
    queryKey: queryKeys.currentTrack,
    queryFn: api.getCurrentTrack,
    enabled,
    refetchInterval,
    refetchIntervalInBackground: false, // Don't poll when tab is hidden
  })
}

export function useCheckCurrentTrack(trackId: string | undefined, enabled: boolean) {
  return useQuery({
    queryKey: queryKeys.trackStatus(trackId),
    queryFn: api.checkCurrentTrack,
    enabled: enabled && !!trackId,
  })
}

export function useIndexSummary(refetchInterval: number | false) {
  return useQuery({
    queryKey: queryKeys.indexSummary,
    queryFn: api.getIndexSummary,
    refetchInterval,
    refetchIntervalInBackground: false, // Don't poll when tab is hidden
  })
}

export function usePlayback(enabled: boolean, refetchInterval: number | false = 4000) {
  return useQuery({
    queryKey: queryKeys.playback,
    queryFn: api.getPlayback,
    enabled,
    refetchInterval: enabled ? refetchInterval : false,
    refetchIntervalInBackground: false,
    refetchOnWindowFocus: 'always',
    refetchOnReconnect: 'always',
  })
}

export function usePlaybackQueue(enabled: boolean) {
  return useQuery({
    queryKey: queryKeys.playbackQueue,
    queryFn: api.getPlaybackQueue,
    enabled,
    refetchInterval: enabled ? 15000 : false,
    refetchIntervalInBackground: false,
  })
}

export function usePlaybackHistory(enabled: boolean) {
  return useQuery({
    queryKey: queryKeys.playbackHistory,
    queryFn: () => api.getPlaybackHistory(20),
    enabled,
    staleTime: 30000,
  })
}

export function useTrackLibraryStatus(trackId: string | undefined, enabled: boolean) {
  return useQuery({
    queryKey: queryKeys.libraryStatus(trackId),
    queryFn: () => api.getTrackLibraryStatus(trackId!),
    enabled: enabled && !!trackId,
  })
}

export function useSyncIndex() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: api.syncIndex,
    onSuccess: (summary) => {
      queryClient.setQueryData(queryKeys.indexSummary, summary)
      queryClient.invalidateQueries({ queryKey: ['libraryStatus'] })
    },
  })
}

type PlaybackMutationContext = {
  previous?: Playback
  expectedTrackId?: string
}

export function usePlaybackControls() {
  const queryClient = useQueryClient()
  const current = () => queryClient.getQueryData<Playback>(queryKeys.playback)
  const setPlayback = (playback?: Playback) => {
    if (playback) queryClient.setQueryData(queryKeys.playback, playback)
  }
  const optimistic = (update: (playback: Playback) => Playback): PlaybackMutationContext => {
    const previous = current()
    if (previous) setPlayback(update(previous))
    return { previous }
  }
  const optimisticTrack = (track?: Track): PlaybackMutationContext => {
    const context = optimistic(playback => track ? {
      ...playback,
      active: true,
      is_playing: true,
      progress_ms: 0,
      timestamp: Date.now(),
      track,
    } : playback)
    return { ...context, expectedTrackId: track?.id }
  }
  const rollback = (_error: Error, _variables: unknown, context?: PlaybackMutationContext) => {
    setPlayback(context?.previous)
  }
  const refreshTimeline = () => {
    queryClient.invalidateQueries({ queryKey: queryKeys.playbackQueue })
    queryClient.invalidateQueries({ queryKey: queryKeys.playbackHistory })
  }
  const reconcile = (matches: (playback: Playback) => boolean) => {
    void reconcilePlayback(matches).then(playback => {
      if (playback) setPlayback(playback)
      refreshTimeline()
    })
  }

  return {
    play: useMutation<void, Error, string | undefined, PlaybackMutationContext>({
      mutationFn: api.play,
      onMutate: trackId => {
        if (!trackId) return optimistic(playback => ({ ...playback, is_playing: true, timestamp: Date.now() }))
        const queue = queryClient.getQueryData<{ tracks: Track[] }>(queryKeys.playbackQueue)
        const history = queryClient.getQueryData<Array<{ track: Track }>>(queryKeys.playbackHistory)
        return optimisticTrack(queue?.tracks.find(track => track.id === trackId) ?? history?.find(item => item.track.id === trackId)?.track)
      },
      onError: rollback,
      onSuccess: (_data, trackId) => reconcile(playback => playback.is_playing && (!trackId || playback.track?.id === trackId)),
    }),
    pause: useMutation<void, Error, void, PlaybackMutationContext>({
      mutationFn: api.pause,
      onMutate: () => optimistic(playback => ({ ...playback, is_playing: false })),
      onError: rollback,
      onSuccess: () => reconcile(playback => !playback.is_playing),
    }),
    next: useMutation<void, Error, void, PlaybackMutationContext>({
      mutationFn: api.next,
      onMutate: () => {
        const queue = queryClient.getQueryData<{ tracks: Track[] }>(queryKeys.playbackQueue)
        return optimisticTrack(queue?.tracks[0])
      },
      onError: rollback,
      onSuccess: (_data, _variables, context) => reconcile(playback => context?.expectedTrackId
        ? playback.track?.id === context.expectedTrackId
        : playback.track?.id !== context?.previous?.track?.id),
    }),
    previous: useMutation<void, Error, void, PlaybackMutationContext>({
      mutationFn: api.previous,
      onMutate: () => {
        const history = queryClient.getQueryData<Array<{ track: Track }>>(queryKeys.playbackHistory)
        return optimisticTrack(history?.[0]?.track)
      },
      onError: rollback,
      onSuccess: (_data, _variables, context) => reconcile(playback => context?.expectedTrackId
        ? playback.track?.id === context.expectedTrackId
        : playback.track?.id !== context?.previous?.track?.id),
    }),
    seek: useMutation<void, Error, number, PlaybackMutationContext>({
      mutationFn: api.seek,
      onMutate: positionMs => optimistic(playback => ({ ...playback, progress_ms: positionMs, timestamp: Date.now() })),
      onError: rollback,
      onSuccess: (_data, positionMs) => reconcile(playback => Math.abs(playback.progress_ms - positionMs) < 3000),
    }),
    skipQueue: useMutation<void, Error, QueueSkipRequest, PlaybackMutationContext>({
      mutationFn: api.skipPlaybackQueue,
      onMutate: request => {
        const queue = queryClient.getQueryData<{ tracks: Track[] }>(queryKeys.playbackQueue)
        return optimisticTrack(queue?.tracks[request.position])
      },
      onError: rollback,
      onSuccess: (_data, request) => reconcile(playback => playback.track?.id === request.expected_track_id),
    }),
  }
}

async function reconcilePlayback(matches: (playback: Playback) => boolean): Promise<Playback | undefined> {
  const delays = [120, 250, 450, 800, 1300, 2000]
  let latest: Playback | undefined
  for (const delay of delays) {
    await new Promise(resolve => window.setTimeout(resolve, delay))
    try {
      latest = await api.getPlayback()
      if (matches(latest)) return latest
    } catch {
      // The normal query error state handles persistent Spotify/API failures.
    }
  }
  return latest
}

export function useGenerateDiscoveryPlaylist() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (dryRun: boolean = false) => api.generateDiscoveryPlaylist(dryRun),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.indexSummary })
    },
  })
}
