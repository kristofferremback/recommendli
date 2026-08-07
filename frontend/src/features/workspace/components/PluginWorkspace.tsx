import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import {
  AlertTriangle, Clock3, Database, Disc3, ExternalLink, Library, ListEnd, ListMusic,
  Music2, Pause, Play, Plus, RefreshCw, Search, SkipBack, SkipForward, X,
} from 'lucide-react'
import { useDocumentVisibility } from '@/shared/hooks/useDocumentVisibility'
import {
  useGenerateDiscoveryPlaylist, useIndexSummary, usePlayback,
  usePlaybackControls, usePlaybackHistory, usePlaybackQueue,
  useSyncIndex, useTrackLibraryStatus,
} from '@/shared/api/queries'
import type { Playback, Playlist, Track } from '@/shared/types/spotify'
import './reference-hybrid.css'

type PluginName = 'queue' | 'discovery' | 'tracks' | 'library'
const plugins: PluginName[] = ['queue', 'discovery', 'tracks', 'library']

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
    const age = index.data.last_synced_at
      ? Date.now() - new Date(index.data.last_synced_at).getTime()
      : Number.POSITIVE_INFINITY
    if (age > 15 * 60 * 1000) {
      attemptedSync.current = true
      sync.mutate()
    }
  }, [index.data, sync])

  const toggle = (plugin: PluginName) => {
    setOpen(current => current.includes(plugin)
      ? current.filter(item => item !== plugin)
      : [...current, plugin])
  }

  const isOpen = (plugin: PluginName) => open.includes(plugin)
  return (
    <main className="rh-page">
      <div className="rh-workspace">
        <PlayerWindow
          playback={playback.data}
          loading={playback.isLoading}
          error={playback.error}
          retry={() => playback.refetch()}
          open={open}
          toggle={toggle}
        />
        {isOpen('queue') && <QueueWindow playback={playback.data} close={() => toggle('queue')} />}
        {(isOpen('discovery') || isOpen('tracks')) && (
          <div className="rh-main-stack">
            {isOpen('discovery') && (
              <DiscoveryWindow
                indexCount={index.data?.unique_track_count}
                onResult={playlist => {
                  setResult(playlist)
                  if (!isOpen('tracks')) setOpen(current => [...current, 'tracks'])
                }}
                close={() => toggle('discovery')}
              />
            )}
            {isOpen('tracks') && (
              <TracksWindow
                playlist={result}
                playback={playback.data}
                close={() => toggle('tracks')}
              />
            )}
          </div>
        )}
        {isOpen('library') && (
          <LibraryWindow close={() => toggle('library')} />
        )}
      </div>
    </main>
  )
}

function PlayerWindow({ playback, loading, error, retry, open, toggle }: {
  playback?: Playback
  loading: boolean
  error: Error | null
  retry: () => void
  open: PluginName[]
  toggle: (plugin: PluginName) => void
}) {
  const controls = usePlaybackControls()
  const track = playback?.track
  const status = useTrackLibraryStatus(track?.id, !!track)
  const [membershipOpen, setMembershipOpen] = useState(false)
  useEffect(() => setMembershipOpen(false), [track?.id])
  const progress = usePlaybackProgress(playback)
  const duration = track?.duration_ms ?? 0
  const controllable = !!playback?.active && !!playback.device && !playback.device.is_restricted
  const albumArt = track?.album?.images?.[0]?.url
  const commandPending = controls.play.isPending || controls.pause.isPending || controls.next.isPending || controls.previous.isPending

  return (
    <HardwareWindow className="rh-player" code="R // PLAYER" fixed>
      <div className="rh-brandbar">
        <strong>recommendli</strong>
        <nav className="rh-launcher" aria-label="Plug-ins">
          <Launcher icon={<ListEnd />} label="Queue and history" active={open.includes('queue')} onClick={() => toggle('queue')} />
          <Launcher icon={<Plus />} label="Discovery builder" active={open.includes('discovery')} onClick={() => toggle('discovery')} />
          <Launcher icon={<ListMusic />} label="Generated tracks" active={open.includes('tracks')} available onClick={() => toggle('tracks')} />
          <Launcher icon={<Database />} label="Library" active={open.includes('library')} onClick={() => toggle('library')} />
        </nav>
      </div>
      <div className="rh-well">
        <div className="rh-lcd rh-player-display">
          <div className="rh-cover">
            {albumArt ? <img src={albumArt} alt="" /> : <Disc3 aria-hidden="true" />}
          </div>
          <div className="rh-readout">
            <div className="rh-clock">{formatTime(progress)}</div>
            {error && <LocalError retry={retry} />}
            <div className={`rh-meter ${playback?.is_playing ? 'is-playing' : ''}`} aria-hidden="true">
              {[35, 72, 48, 88, 61, 40, 78].map((height, i) => <i key={i} style={{ height: `${height}%` }} />)}
            </div>
            {loading ? (
              <div className="rh-muted rh-loading-bars">···</div>
            ) : error ? null : track ? (
              <>
                <a className="rh-track-name" href={track.external_urls?.spotify} target="_blank" rel="noreferrer">{track.name}</a>
                <span className="rh-artist">{track.artists.map(artist => artist.name).join(', ')}</span>
                <div className={`rh-membership ${status.data?.in_library ? 'is-member' : 'is-new'}`}>
                  {status.data?.in_library ? (
                    <button className="rh-membership-button" aria-expanded={membershipOpen} onClick={() => setMembershipOpen(open => !open)}>
                      <Library aria-hidden="true" /> IN LIBRARY · {status.data.playlists.length}
                    </button>
                  ) : (
                    <span className="rh-membership-state"><Library aria-hidden="true" /> {status.isLoading ? 'CHECKING' : 'NEW TO LIBRARY'}</span>
                  )}
                  {membershipOpen && <div className="rh-membership-popover">{status.data?.playlists.map(playlist => (
                    <a key={playlist.id} href={playlist.external_urls.spotify} target="_blank" rel="noreferrer">
                      <span>{playlist.name}</span><ExternalLink aria-hidden="true" />
                    </a>
                  ))}</div>}
                </div>
              </>
            ) : (
              <><span className="rh-track-name">NOTHING PLAYING</span><span className="rh-artist">OPEN SPOTIFY ON A DEVICE</span></>
            )}
            <SeekBar
              progress={progress}
              duration={duration}
              disabled={!controllable || !duration}
              onCommit={position => controls.seek.mutate(position)}
            />
          </div>
        </div>
      </div>
      <div className="rh-device">
        <Music2 aria-hidden="true" />
        <span>{playback?.device ? `${playback.device.name} · ${playback.device.volume_percent}%` : '—'}</span>
        <i className={controllable ? 'online' : ''} />
      </div>
      <div className="rh-transport">
        <HardwareButton label="Previous" disabled={!controllable || commandPending} onClick={() => controls.previous.mutate()}><SkipBack /></HardwareButton>
        <HardwareButton primary label={playback?.is_playing ? 'Pause' : 'Play'} disabled={!controllable || commandPending} onClick={() => playback?.is_playing ? controls.pause.mutate() : controls.play.mutate(undefined)}>
          {playback?.is_playing ? <Pause /> : <Play />}
        </HardwareButton>
        <HardwareButton label="Next" disabled={!controllable || commandPending} onClick={() => controls.next.mutate()}><SkipForward /></HardwareButton>
      </div>
    </HardwareWindow>
  )
}

function QueueWindow({ playback, close }: { playback?: Playback; close: () => void }) {
  const [tab, setTab] = useState<'history' | 'queue'>('queue')
  const visible = useDocumentVisibility()
  const queue = usePlaybackQueue(visible)
  const history = usePlaybackHistory(visible)
  const controls = usePlaybackControls()
  return (
    <HardwareWindow className="rh-queue" code="R // QUEUE" onClose={close}>
      <div className="rh-mobile-tabs">
        <button className={tab === 'history' ? 'active' : ''} onClick={() => setTab('history')} aria-label="History"><Clock3 /> {history.data?.length ?? 0}</button>
        <button className={tab === 'queue' ? 'active' : ''} onClick={() => setTab('queue')} aria-label="Queue"><ListEnd /> {queue.data?.tracks.length ?? 0}</button>
      </div>
      <div className="rh-current"><Play /> <b>{playback?.track?.name ?? '—'}</b><span>{formatTime(playback?.progress_ms ?? 0)} / {formatTime(playback?.track?.duration_ms ?? 0)}</span></div>
      {(queue.error || history.error) && <LocalError retry={() => { queue.refetch(); history.refetch() }} />}
      <div className="rh-timeline-grid">
        <TrackList title={<Clock3 />} count={history.data?.length ?? 0} className={tab === 'history' ? 'mobile-active' : ''}>
          {history.data?.map((item, i) => (
            <CompactTrack key={`${item.played_at}-${item.track.id}`} track={item.track} index={`-${String(i + 1).padStart(2, '0')}`} actionLabel="Play again" action={<Play />} onAction={() => controls.play.mutate(item.track.id)} />
          ))}
        </TrackList>
        <TrackList title={<ListEnd />} count={queue.data?.tracks.length ?? 0} className={tab === 'queue' ? 'mobile-active' : ''}>
          {queue.data?.tracks.map((track, i) => (
            <CompactTrack key={`${track.id}-${i}`} track={track} index={`+${String(i + 1).padStart(2, '0')}`} actionLabel="Skip here" action={<SkipForward />} onAction={() => controls.skipQueue.mutate({ position: i, expected_track_id: track.id, expected_current_track_id: queue.data.currently_playing?.id })} />
          ))}
        </TrackList>
      </div>
    </HardwareWindow>
  )
}

function DiscoveryWindow({ indexCount, onResult, close }: { indexCount?: number; onResult: (playlist: Playlist) => void; close: () => void }) {
  const generation = useGenerateDiscoveryPlaylist()
  const build = async () => onResult(await generation.mutateAsync(false))
  return (
    <HardwareWindow code="R // DISCOVERY" onClose={close}>
      <div className="rh-well rh-generator">
        <div className="rh-sources"><Source name="Discover Weekly" count={50} /><Source name="Release Radar" count={30} orange /></div>
        <div className="rh-routing" aria-hidden="true"><i /></div>
        <div className="rh-filter"><Database /><b>{indexCount?.toLocaleString() ?? '—'}</b></div>
        <div className="rh-lcd rh-output"><ListMusic /><strong>RECOMMENDLI DISCOVERY</strong><span>{generation.data?.tracks.length ?? 0}</span></div>
      </div>
      {generation.error && <LocalError retry={build} />}
      <div className="rh-build"><HardwareButton primary label="Build playlist" disabled={generation.isPending} onClick={build}><Plus /> {generation.isPending ? 'BUILDING' : 'BUILD PLAYLIST'}</HardwareButton></div>
      <div className="rh-stages">{[ListMusic, Database, Disc3, Plus].map((Icon, i) => <i key={i} className={generation.isPending && i === 2 ? 'active' : generation.isSuccess ? 'done' : ''}><Icon /></i>)}</div>
    </HardwareWindow>
  )
}

function TracksWindow({ playlist, playback, close }: { playlist: Playlist | null; playback?: Playback; close: () => void }) {
  const controls = usePlaybackControls()
  const controllable = !!playback?.device && !playback.device.is_restricted
  return (
    <HardwareWindow code="R // TRACKS" onClose={close}>
      <div className="rh-lcd rh-result-head"><ListMusic /><b>{playlist?.name ?? 'NO OUTPUT'}</b><strong>{playlist?.tracks.length ?? 0}</strong>{playlist && <a href={playlist.external_urls.spotify} target="_blank" rel="noreferrer" aria-label="Open playlist"><ExternalLink /></a>}</div>
      <div className="rh-track-table">
        {playlist?.tracks.map((track, i) => (
          <div className="rh-result-row" key={track.id}>
            <span>{String(i + 1).padStart(2, '0')}</span><TrackArt track={track} /><div><b>{track.name}</b><small>{track.artists.map(a => a.name).join(', ')}</small></div><span className="rh-desktop-meta">{track.album.name}</span><span>{formatTime(track.duration_ms ?? 0)}</span>
            <div>{controllable && <button onClick={() => controls.play.mutate(track.id)} aria-label={`Play ${track.name}`} title="Play on the active Spotify device; replaces the current queue"><Play /></button>}<a href={track.external_urls.spotify} target="_blank" rel="noreferrer" aria-label={`Open ${track.name} in Spotify`}><ExternalLink /></a></div>
          </div>
        )) ?? <div className="rh-empty"><ListMusic /></div>}
      </div>
    </HardwareWindow>
  )
}

function LibraryWindow({ close }: { close: () => void }) {
  const index = useIndexSummary(false)
  const sync = useSyncIndex()
  const [search, setSearch] = useState('')
  const [drawer, setDrawer] = useState(false)
  const closeDrawer = useCallback(() => setDrawer(false), [])
  const allPlaylists = index.data?.playlists ?? []
  const playlists = useMemo(() => allPlaylists.filter(p => p.name.toLowerCase().includes(search.toLowerCase())), [allPlaylists, search])
  const browser = <><label className="rh-search"><Search /><input value={search} onChange={event => setSearch(event.target.value)} placeholder="Search playlists" /></label><PlaylistRows playlists={playlists} /></>
  return (
    <>
      <HardwareWindow className="rh-library" code="R // LIBRARY" onClose={close}>
        <div className="rh-lcd rh-stats"><Stat icon={<Music2 />} value={index.data?.unique_track_count} /><Stat icon={<ListMusic />} value={index.data?.playlist_count} /></div>
        {(index.error || sync.error) && <LocalError retry={() => sync.mutate()} />}
        <div className="rh-desktop-browser">{browser}</div>
        <div className="rh-mobile-preview"><PlaylistRows playlists={allPlaylists.slice(0, 5)} compact /></div>
        <button className="rh-open-browser" onClick={() => setDrawer(true)}><Search /> <ListMusic /> <span>{index.data?.playlist_count ?? 0}</span></button>
        <div className="rh-sync"><span><Clock3 /> {index.data?.last_synced_at ? new Date(index.data.last_synced_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : '—'}</span><button onClick={() => sync.mutate()} disabled={sync.isPending} aria-label="Synchronize library"><RefreshCw className={sync.isPending ? 'spin' : ''} /></button></div>
      </HardwareWindow>
      {drawer && createPortal(<PlaylistDrawer close={closeDrawer}>{browser}</PlaylistDrawer>, document.body)}
    </>
  )
}

function PlaylistRows({ playlists, compact = false }: { playlists: Array<{ id: string; name: string; tracks?: { total: number }; external_urls: { spotify: string } }>; compact?: boolean }) {
  return <div className={`rh-playlists ${compact ? 'is-compact' : ''}`}>{playlists.map((playlist, i) => <div key={playlist.id}><span>{String(i + 1).padStart(2, '0')}</span><b>{playlist.name}</b><small>{playlist.tracks?.total ?? '—'}</small><a href={playlist.external_urls.spotify} target="_blank" rel="noreferrer" aria-label={`Open ${playlist.name}`}><ExternalLink /></a></div>)}</div>
}

function PlaylistDrawer({ close, children }: { close: () => void; children: React.ReactNode }) {
  useEffect(() => {
    const scrollY = window.scrollY
    const previous = {
      htmlOverflow: document.documentElement.style.overflow,
      bodyOverflow: document.body.style.overflow,
      bodyPosition: document.body.style.position,
      bodyTop: document.body.style.top,
      bodyWidth: document.body.style.width,
    }
    document.documentElement.style.overflow = 'hidden'
    document.body.style.overflow = 'hidden'
    document.body.style.position = 'fixed'
    document.body.style.top = `-${scrollY}px`
    document.body.style.width = '100%'
    const keydown = (event: KeyboardEvent) => { if (event.key === 'Escape') close() }
    window.addEventListener('keydown', keydown)
    return () => {
      document.documentElement.style.overflow = previous.htmlOverflow
      document.body.style.overflow = previous.bodyOverflow
      document.body.style.position = previous.bodyPosition
      document.body.style.top = previous.bodyTop
      document.body.style.width = previous.bodyWidth
      window.removeEventListener('keydown', keydown)
      window.scrollTo(0, scrollY)
    }
  }, [close])
  return <aside className="rh-drawer" role="dialog" aria-modal="true" aria-label="Playlists"><header className="rh-title"><span>R // PLAYLISTS</span><button onClick={close} aria-label="Close playlists"><X /></button></header>{children}</aside>
}

function LocalError({ retry }: { retry: () => void }) {
  return <div className="rh-local-error" role="alert"><AlertTriangle /><span>ERR</span><button onClick={retry} aria-label="Retry"><RefreshCw /></button></div>
}

function HardwareWindow({ code, onClose, children, className = '', fixed = false }: { code: string; onClose?: () => void; children: React.ReactNode; className?: string; fixed?: boolean }) {
  return <section className={`rh-window ${className}`}><i className="rh-screw tl" /><i className="rh-screw tr" /><i className="rh-screw bl" /><i className="rh-screw br" /><header className="rh-title"><span>{code}</span>{onClose ? <button onClick={onClose} aria-label={`Close ${code}`}><X /></button> : <div className="rh-window-dots"><i /><i /><i /></div>}</header>{children}{fixed && null}</section>
}
function Launcher({ icon, label, active, available, onClick }: { icon: React.ReactNode; label: string; active: boolean; available?: boolean; onClick: () => void }) { return <button className={`rh-launch ${active ? 'active' : ''} ${available ? 'available' : ''}`} aria-label={label} aria-pressed={active} title={label} onClick={onClick}>{icon}</button> }
function HardwareButton({ children, label, primary, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement> & { label: string; primary?: boolean }) { return <button className={`rh-key ${primary ? 'primary' : ''}`} aria-label={label} title={label} {...props}>{children}</button> }
function TrackList({ title, count, children, className = '' }: { title: React.ReactNode; count: number; children: React.ReactNode; className?: string }) { return <div className={`rh-list ${className}`}><header>{title}<b>{count}</b></header>{children}</div> }
function CompactTrack({ track, index, action, actionLabel, onAction }: { track: Track; index: string; action: React.ReactNode; actionLabel: string; onAction: () => void }) { return <div className="rh-compact-track"><span>{index}</span><TrackArt track={track} /><div><b>{track.name}</b><small>{track.artists.map(a => a.name).join(', ')}</small></div><button aria-label={`${actionLabel}: ${track.name}`} title={actionLabel} onClick={onAction}>{action}</button></div> }
function TrackArt({ track }: { track: Track }) { const src = track.album?.images?.[2]?.url ?? track.album?.images?.[0]?.url; return src ? <img className="rh-art" src={src} alt="" /> : <span className="rh-art"><Music2 /></span> }
function Source({ name, count, orange }: { name: string; count: number; orange?: boolean }) { return <div className="rh-source"><span className={orange ? 'orange' : ''}><Disc3 /></span><div><b>{name}</b><small>{count}</small></div><i /></div> }
function Stat({ icon, value }: { icon: React.ReactNode; value?: number }) { return <div>{icon}<strong>{value?.toLocaleString() ?? '—'}</strong></div> }

function usePluginLayout(): [PluginName[], React.Dispatch<React.SetStateAction<PluginName[]>>] {
  const [desktop, setDesktop] = useState(() => window.matchMedia('(min-width: 761px)').matches)
  const [open, setOpen] = useState<PluginName[]>(() => readLayout(window.matchMedia('(min-width: 761px)').matches))
  useEffect(() => {
    const media = window.matchMedia('(min-width: 761px)')
    const change = () => { setDesktop(media.matches); setOpen(readLayout(media.matches)) }
    media.addEventListener('change', change)
    return () => media.removeEventListener('change', change)
  }, [])
  useEffect(() => localStorage.setItem(`recommendli-layout-${desktop ? 'desktop' : 'mobile'}-v1`, JSON.stringify(open)), [desktop, open])
  return [open, setOpen]
}
function readLayout(desktop: boolean): PluginName[] { try { const value = JSON.parse(localStorage.getItem(`recommendli-layout-${desktop ? 'desktop' : 'mobile'}-v1`) ?? 'null'); if (Array.isArray(value)) return value.filter(item => plugins.includes(item)) } catch {} return desktop ? [...plugins] : [] }
function SeekBar({ progress, duration, disabled, onCommit }: { progress: number; duration: number; disabled: boolean; onCommit: (position: number) => void }) {
  const [value, setValue] = useState(progress)
  const dragging = useRef(false)
  useEffect(() => { if (!dragging.current) setValue(progress) }, [progress])
  const commit = () => {
    if (!dragging.current) return
    dragging.current = false
    onCommit(value)
  }
  return <input className="rh-seek" style={{ '--rh-progress': `${duration ? Math.min(100, value / duration * 100) : 0}%` } as React.CSSProperties} type="range" min={0} max={Math.max(duration, 1)} value={Math.min(value, duration || 1)} disabled={disabled} aria-label="Playback position" onPointerDown={() => { dragging.current = true }} onChange={event => setValue(Number(event.target.value))} onPointerUp={commit} onKeyDown={() => { dragging.current = true }} onKeyUp={commit} />
}
function usePlaybackProgress(playback?: Playback) { const [progress, setProgress] = useState(playback?.progress_ms ?? 0); useEffect(() => { const update = () => setProgress(Math.min((playback?.progress_ms ?? 0) + (playback?.is_playing ? Math.max(0, Date.now() - (playback.timestamp || Date.now())) : 0), playback?.track?.duration_ms ?? Number.MAX_SAFE_INTEGER)); update(); const timer = window.setInterval(update, 500); return () => window.clearInterval(timer) }, [playback]); return progress }
function formatTime(milliseconds: number) { const seconds = Math.max(0, Math.floor(milliseconds / 1000)); return `${Math.floor(seconds / 60).toString().padStart(2, '0')}:${(seconds % 60).toString().padStart(2, '0')}` }
