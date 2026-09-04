import { useCallback, useRef } from 'react'
import type { ButtonHTMLAttributes, KeyboardEvent, ReactNode } from 'react'
import { AlertTriangle, ChevronUp, Music2, RefreshCw, X } from 'lucide-react'
import type { Track } from '@/shared/types/spotify'
import { useDeskIfAny } from './desk'
import { artistNames } from './lib/format'
import type { WindowName } from './lib/frames'

/* Shared hardware chrome: one window frame, one button, one row. */

export function Window({ code, plugin, onClose, controls, children }: {
  code: string
  plugin: WindowName | 'connect'
  onClose?: () => void
  /** Extra title bar buttons, rendered before shade and close. */
  controls?: ReactNode
  children: ReactNode
}) {
  const name = windowName(code)
  const desk = useDeskIfAny()
  // Only plug-in windows on a desktop desk get a frame; the connection panel and phones render statically.
  const managed = desk && plugin !== 'connect' ? { desk, id: plugin } : null
  const frame = managed?.desk.frame(managed.id)
  // The desk api changes on every move, so the ref callback reads it through a ref to stay stable.
  const latest = useRef(managed)
  latest.current = managed
  const register = useCallback((element: HTMLElement | null) => latest.current?.desk.register(latest.current.id, element), [])
  const shaded = !!frame?.shaded
  const style = frame && managed
    ? { left: frame.x, top: frame.y, width: frame.w, height: shaded || frame.h === null ? undefined : frame.h, zIndex: managed.desk.zIndex(managed.id) }
    : undefined
  const classes = ['win', `win-${plugin}`, shaded && 'shaded', managed?.desk.dragging === plugin && 'dragging'].filter(Boolean).join(' ')
  const keys = (event: KeyboardEvent<HTMLElement>) => {
    if (!managed || event.target !== event.currentTarget) return
    const step = event.shiftKey ? 40 : 10
    const moves: Record<string, [number, number]> = { ArrowLeft: [-step, 0], ArrowRight: [step, 0], ArrowUp: [0, -step], ArrowDown: [0, step] }
    if (moves[event.key]) {
      event.preventDefault()
      managed.desk.nudge(managed.id, ...moves[event.key])
    } else if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault()
      managed.desk.toggleShade(managed.id)
    }
  }
  return (
    <section
      ref={managed ? register : undefined}
      className={classes}
      style={style}
      aria-label={name}
      onPointerDownCapture={frame && managed ? () => managed.desk.raise(managed.id) : undefined}
    >
      <header
        className="win-title"
        tabIndex={frame ? 0 : undefined}
        aria-label={frame ? `${name} title bar` : undefined}
        title={frame ? 'Drag to move. Double-click to shade. Arrow keys move when focused.' : undefined}
        onPointerDown={frame && managed ? event => managed.desk.startDrag(managed.id, event) : undefined}
        onDoubleClick={frame && managed ? () => managed.desk.toggleShade(managed.id) : undefined}
        onKeyDown={frame ? keys : undefined}
      >
        <h2>{code}</h2>
        <div className="win-controls">
          {controls}
          {frame && managed && (
            <button type="button" onClick={() => managed.desk.toggleShade(managed.id)} aria-pressed={shaded} aria-label={`Shade ${name}`} title={shaded ? 'Unshade' : 'Shade'}><ChevronUp /></button>
          )}
          {onClose && <button type="button" onClick={onClose} aria-label={`Close ${name}`} title="Close"><X /></button>}
        </div>
      </header>
      <div className="win-body">
        <i className="screw tl" /><i className="screw tr" /><i className="screw bl" /><i className="screw br" />
        {children}
      </div>
      {frame && managed && !shaded && <i className={frame.h === null ? 'win-resize wide' : 'win-resize'} aria-hidden="true" onPointerDown={event => managed.desk.startResize(managed.id, event)} />}
    </section>
  )
}

/** "R // PLAYER" reads as "R slash slash player" to a screen reader; name the region "Player" instead. */
function windowName(code: string) {
  const last = code.split('//').pop()?.trim() ?? code
  return last.charAt(0) + last.slice(1).toLowerCase()
}

type KeyProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  label?: string
  variant?: 'plain' | 'primary' | 'dark'
  icon?: boolean
  lit?: boolean
  big?: boolean
  flag?: boolean
}

export function Key({ label, variant = 'plain', icon, lit, big, flag, className, children, ...props }: KeyProps) {
  const classes = ['key', variant !== 'plain' && variant, icon && 'icon', lit && 'lit', big && 'big', flag && 'flag', className]
  return (
    <button type="button" className={classes.filter(Boolean).join(' ')} aria-label={label} title={label} {...props}>
      {children}
    </button>
  )
}

export function Led({ on, busy }: { on?: boolean; busy?: boolean }) {
  return <i className={`led ${busy ? 'busy' : on ? 'on' : ''}`} aria-hidden="true" />
}

export function ErrorChip({ message = 'Spotify did not answer', retry }: { message?: string; retry?: () => void }) {
  return (
    <div className="error-chip" role="alert">
      <AlertTriangle aria-hidden="true" />
      <span>{message}</span>
      {retry && <Key variant="dark" icon label="Retry" onClick={retry}><RefreshCw /></Key>}
    </div>
  )
}

export function TrackList({ icon, title, count, active, loading, empty, children }: {
  icon: ReactNode
  title: string
  count: number
  active?: boolean
  loading?: boolean
  empty: string
  children: ReactNode
}) {
  return (
    <div className={`list ${active ? 'active' : ''}`} role="region" aria-label={title}>
      <header className="list-head">{icon}{title}<b>{count}</b></header>
      {count > 0 ? <ol>{children}</ol> : <div className="empty">{loading ? 'Loading' : empty}</div>}
    </div>
  )
}

export function TrackRow({ index, track, album, duration, actions }: {
  index: string
  track: Track
  album?: string
  duration?: string
  actions: ReactNode
}) {
  return (
    <li className={`row ${album !== undefined ? 'has-album' : ''}`}>
      <span className="row-index">{index}</span>
      <TrackArt track={track} />
      <div className="row-text">
        <b className="row-title">{track.name}</b>
        <span className="row-sub">{artistNames(track.artists)}</span>
      </div>
      {album !== undefined && <span className="row-meta album">{album}</span>}
      {duration !== undefined && <span className="row-meta">{duration}</span>}
      <div className="row-actions">{actions}</div>
    </li>
  )
}

export function TrackArt({ track }: { track: Track }) {
  const src = track.album?.images?.[2]?.url ?? track.album?.images?.[0]?.url
  return src
    ? <img className="art" src={src} alt="" loading="lazy" />
    : <span className="art" aria-hidden="true"><Music2 /></span>
}
