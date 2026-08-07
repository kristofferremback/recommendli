import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from './client'

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

export function usePlaybackControls() {
  const queryClient = useQueryClient()
  const refresh = () => {
    queryClient.invalidateQueries({ queryKey: queryKeys.playback })
    queryClient.invalidateQueries({ queryKey: queryKeys.playbackQueue })
    queryClient.invalidateQueries({ queryKey: queryKeys.playbackHistory })
  }

  return {
    play: useMutation({ mutationFn: api.play, onSuccess: refresh }),
    pause: useMutation({ mutationFn: api.pause, onSuccess: refresh }),
    next: useMutation({ mutationFn: api.next, onSuccess: refresh }),
    previous: useMutation({ mutationFn: api.previous, onSuccess: refresh }),
    seek: useMutation({ mutationFn: api.seek, onSuccess: refresh }),
    skipQueue: useMutation({ mutationFn: api.skipPlaybackQueue, onSuccess: refresh }),
  }
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
