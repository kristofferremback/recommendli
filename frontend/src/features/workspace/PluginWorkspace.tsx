import { useEffect, useRef, useState } from 'react'
import type { Dispatch, SetStateAction } from 'react'
import { useDocumentVisibility } from '@/shared/hooks/useDocumentVisibility'
import { useIndexSummary, usePlayback, useSyncIndex } from '@/shared/api/queries'
import type { Playlist } from '@/shared/types/spotify'
import { plugins } from './plugins'
import type { PluginName } from './plugins'
import { PlayerWindow } from './windows/PlayerWindow'
import { QueueWindow } from './windows/QueueWindow'
import { DiscoveryWindow } from './windows/DiscoveryWindow'
import { TracksWindow } from './windows/TracksWindow'
import { LibraryWindow } from './windows/LibraryWindow'

const SYNC_AFTER_MS = 15 * 60 * 1000

export function PluginWorkspace() {
  const visible = useDocumentVisibility()
  const [open, setOpen] = usePluginLayout()
  const [result, setResult] = useState<Playlist | null>(null)
  const playback = usePlayback(visible, visible ? 4000 : false)
  const wasVisible = useRef(visible)
  useEffect(() => {
    if (visible && !wasVisible.current) void playback.refetch()
    wasVisible.current = visible
  }, [visible, playback.refetch])
  useEffect(() => {
    const refresh = () => { if (!document.hidden) void playback.refetch() }
    window.addEventListener('focus', refresh)
    return () => window.removeEventListener('focus', refresh)
  }, [playback.refetch])

  const index = useIndexSummary(false)
  const sync = useSyncIndex()
  const attemptedSync = useRef(false)
  useEffect(() => {
    if (attemptedSync.current || !index.data || sync.isPending) return
    const age = index.data.last_synced_at ? Date.now() - new Date(index.data.last_synced_at).getTime() : Number.POSITIVE_INFINITY
    if (age > SYNC_AFTER_MS) {
      attemptedSync.current = true
      sync.mutate()
    }
  }, [index.data, sync])

  const toggle = (plugin: PluginName) => {
    setOpen(current => current.includes(plugin) ? current.filter(item => item !== plugin) : [...current, plugin])
  }
  const isOpen = (plugin: PluginName) => open.includes(plugin)

  return (
    <main className="page">
      <div className="desk">
        <PlayerWindow
          playback={playback.data}
          loading={playback.isLoading}
          error={playback.error}
          retry={() => playback.refetch()}
          open={open}
          toggle={toggle}
          hasResult={!!result}
        />
        {isOpen('queue') && <QueueWindow playback={playback.data} close={() => toggle('queue')} />}
        {(isOpen('discovery') || isOpen('tracks')) && (
          <div className="main-stack">
            {isOpen('discovery') && (
              <DiscoveryWindow
                indexCount={index.data?.unique_track_count}
                onResult={playlist => {
                  setResult(playlist)
                  setOpen(current => current.includes('tracks') ? current : [...current, 'tracks'])
                }}
                close={() => toggle('discovery')}
              />
            )}
            {isOpen('tracks') && <TracksWindow playlist={result} playback={playback.data} close={() => toggle('tracks')} />}
          </div>
        )}
        {isOpen('library') && <LibraryWindow sync={sync} close={() => toggle('library')} />}
      </div>
    </main>
  )
}

const DESKTOP_QUERY = '(min-width: 761px)'

function usePluginLayout(): [PluginName[], Dispatch<SetStateAction<PluginName[]>>] {
  const [desktop, setDesktop] = useState(() => window.matchMedia(DESKTOP_QUERY).matches)
  const [open, setOpen] = useState<PluginName[]>(() => readLayout(window.matchMedia(DESKTOP_QUERY).matches))
  useEffect(() => {
    const media = window.matchMedia(DESKTOP_QUERY)
    const change = () => { setDesktop(media.matches); setOpen(readLayout(media.matches)) }
    media.addEventListener('change', change)
    return () => media.removeEventListener('change', change)
  }, [])
  useEffect(() => {
    try { localStorage.setItem(layoutKey(desktop), JSON.stringify(open)) } catch {}
  }, [desktop, open])
  return [open, setOpen]
}

function layoutKey(desktop: boolean) {
  return `recommendli-layout-${desktop ? 'desktop' : 'mobile'}-v1`
}

function readLayout(desktop: boolean): PluginName[] {
  try {
    const value = JSON.parse(localStorage.getItem(layoutKey(desktop)) ?? 'null')
    if (Array.isArray(value)) return value.filter((item): item is PluginName => plugins.includes(item))
  } catch {}
  return desktop ? [...plugins] : ['queue', 'library']
}
