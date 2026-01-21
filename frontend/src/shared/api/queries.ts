import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from './client'

export const queryKeys = {
  user: ['user'] as const,
  currentTrack: ['currentTrack'] as const,
  trackStatus: (id?: string) => ['trackStatus', id] as const,
  indexSummary: ['indexSummary'] as const,
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

export function useGenerateDiscoveryPlaylist() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (dryRun: boolean = false) => api.generateDiscoveryPlaylist(dryRun),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.indexSummary })
    },
  })
}
