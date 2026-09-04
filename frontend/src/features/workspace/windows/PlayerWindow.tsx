import { useEffect, useRef, useState } from 'react'
import type { CSSProperties } from 'react'
import { Database, Disc3, ExternalLink, LayoutGrid, Library, ListEnd, ListMusic, Music2, Pause, Play, Plus, SkipBack, SkipForward } from 'lucide-react'
import { usePlaybackControls, useTrackLibraryStatus } from '@/shared/api/queries'
import { usePlaybackProgress } from '@/shared/hooks/usePlaybackProgress'
import type { Playback } from '@/shared/types/spotify'
import { ErrorChip, Key, Led, Window } from '../chrome'
import { useDesk } from '../desk'
import { artistNames, formatTime } from '../lib/format'
import type { PluginName } from '../plugins'

export function PlayerWindow({ playback, loading, error, retry, open, toggle, hasResult }: {
  playback?: Playback
  loading: boolean
  error: Error | null
  retry: () => void
  open: PluginName[]
  toggle: (plugin: PluginName) => void
  hasResult: boolean
}) {
  const controls = usePlaybackControls()
  const track = playback?.track
  const progress = usePlaybackProgress(playback)
  const duration = track?.duration_ms ?? 0
  const controllable = !!playback?.active && !!playback.device && !playback.device.is_restricted
  const albumArt = track?.album?.images?.[0]?.url
  const commandPending = controls.play.isPending || controls.pause.isPending || controls.next.isPending || controls.previous.isPending
  const commandFailed = [controls.play, controls.pause, controls.next, controls.previous, controls.seek, controls.skipQueue].some(command => command.isError)
  const isOpen = (plugin: PluginName) => open.includes(plugin)
  const desk = useDesk()
  const tidy = desk.desktop && (
    <button type="button" onClick={desk.tidy} aria-label="Tidy windows" title="Tidy windows"><LayoutGrid /></button>
  )

  return (
    <Window code="R // PLAYER" plugin="player" controls={tidy}>
      <div className="brandbar">
        <strong>recommendli</strong>
        <nav className="launcher" aria-label="Plug-ins">
          <Key icon label="Queue and history" lit={isOpen('queue')} aria-pressed={isOpen('queue')} onClick={() => toggle('queue')}><ListEnd /></Key>
          <Key icon label="Discovery builder" lit={isOpen('discovery')} aria-pressed={isOpen('discovery')} onClick={() => toggle('discovery')}><Plus /></Key>
          <Key icon label="Generated tracks" lit={isOpen('tracks')} aria-pressed={isOpen('tracks')} flag={hasResult && !isOpen('tracks')} onClick={() => toggle('tracks')}><ListMusic /></Key>
          <Key icon label="Library" lit={isOpen('library')} aria-pressed={isOpen('library')} onClick={() => toggle('library')}><Database /></Key>
        </nav>
      </div>
      <div className="well">
        <div className="lcd display">
          <div className="cover">
            {albumArt ? <img src={albumArt} alt="" /> : <Disc3 aria-hidden="true" />}
          </div>
          <div className="readout">
            <div className="readout-top">
              <div className={`meter ${playback?.is_playing ? 'playing' : ''}`} aria-hidden="true">
                {[35, 72, 48, 88, 61, 40, 78, 55].map((height, i) => <i key={i} style={{ height: `${height}%` }} />)}
              </div>
              <div className="clock">{formatTime(progress)}</div>
            </div>
            <div className="readout-track" aria-live="polite">
            {track ? (
              <>
                <a className="track-name" href={track.external_urls?.spotify} target="_blank" rel="noreferrer">{track.name}</a>
                <span className="track-artist">{artistNames(track.artists)}</span>
                <Membership trackId={track.id} />
              </>
            ) : (
              <>
                <span className="track-name dim">{loading ? 'Connecting' : error ? 'No signal' : 'Nothing playing'}</span>
                <span className="track-artist dim">{loading || error ? ' ' : 'Open Spotify on a device'}</span>
              </>
            )}
            </div>
          </div>
          {error && <ErrorChip retry={retry} />}
          {!error && commandFailed && <ErrorChip message="Spotify refused the last command" />}
          <SeekBar
            progress={progress}
            duration={duration}
            disabled={!controllable || !duration}
            onCommit={position => controls.seek.mutate(position)}
          />
        </div>
      </div>
      <div className="panel-row">
        <Music2 aria-hidden="true" />
        <span>{playback?.device ? `${playback.device.name} · ${playback.device.volume_percent}%` : 'No active device'}</span>
        <Led on={controllable} />
      </div>
      <div className="transport">
        <Key icon label="Previous" disabled={!controllable || commandPending} onClick={() => controls.previous.mutate()}><SkipBack /></Key>
        <Key variant="primary" label={playback?.is_playing ? 'Pause' : 'Play'} disabled={!controllable || commandPending} onClick={() => playback?.is_playing ? controls.pause.mutate() : controls.play.mutate(undefined)}>
          {playback?.is_playing ? <Pause /> : <Play />}
        </Key>
        <Key icon label="Next" disabled={!controllable || commandPending} onClick={() => controls.next.mutate()}><SkipForward /></Key>
      </div>
    </Window>
  )
}

function Membership({ trackId }: { trackId: string }) {
  const status = useTrackLibraryStatus(trackId, true)
  const [open, setOpen] = useState(false)
  const root = useRef<HTMLDivElement>(null)
  useEffect(() => setOpen(false), [trackId])
  useEffect(() => {
    if (!open) return
    const away = (event: PointerEvent) => { if (!root.current?.contains(event.target as Node)) setOpen(false) }
    const escape = (event: KeyboardEvent) => { if (event.key === 'Escape') setOpen(false) }
    document.addEventListener('pointerdown', away)
    document.addEventListener('keydown', escape)
    return () => { document.removeEventListener('pointerdown', away); document.removeEventListener('keydown', escape) }
  }, [open])

  if (!status.data) {
    return <div className="membership"><span className="chip"><Library aria-hidden="true" /><span>{status.isError ? 'Library unknown' : 'Checking library'}</span></span></div>
  }
  if (!status.data.in_library) {
    return <div className="membership"><span className="chip"><Library aria-hidden="true" /><span>New to library</span></span></div>
  }
  const playlists = status.data.playlists
  return (
    <div className="membership" ref={root}>
      <button type="button" className="chip member" aria-expanded={open} onClick={() => setOpen(value => !value)}>
        <Library aria-hidden="true" /><span>In {playlists.length} {playlists.length === 1 ? 'playlist' : 'playlists'}</span>
      </button>
      {open && (
        <div className="popover">
          {playlists.map(playlist => (
            <a key={playlist.id} href={playlist.external_urls.spotify} target="_blank" rel="noreferrer">
              <span>{playlist.name}</span><ExternalLink aria-hidden="true" />
            </a>
          ))}
        </div>
      )}
    </div>
  )
}

/** Keys that move a range input. Tab must not arm a drag, its keyup lands on the next element. */
const SEEK_KEYS = new Set(['ArrowLeft', 'ArrowRight', 'ArrowUp', 'ArrowDown', 'Home', 'End', 'PageUp', 'PageDown'])

function SeekBar({ progress, duration, disabled, onCommit }: { progress: number; duration: number; disabled: boolean; onCommit: (position: number) => void }) {
  const [value, setValue] = useState(progress)
  const dragging = useRef(false)
  useEffect(() => { if (!dragging.current) setValue(progress) }, [progress])
  const commit = () => {
    if (!dragging.current) return
    dragging.current = false
    onCommit(value)
  }
  const percent = duration ? Math.min(100, value / duration * 100) : 0
  return (
    <input
      className="seek"
      style={{ '--progress': `${percent}%` } as CSSProperties}
      type="range"
      min={0}
      max={Math.max(duration, 1)}
      value={Math.min(value, duration || 1)}
      disabled={disabled}
      aria-label="Playback position"
      aria-valuetext={`${formatTime(value)} of ${formatTime(duration)}`}
      onPointerDown={() => { dragging.current = true }}
      onChange={event => setValue(Number(event.target.value))}
      onPointerUp={commit}
      onPointerCancel={commit}
      onKeyDown={event => { if (SEEK_KEYS.has(event.key)) dragging.current = true }}
      onKeyUp={commit}
      onBlur={commit}
    />
  )
}
