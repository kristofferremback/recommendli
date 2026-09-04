import type { ButtonHTMLAttributes, ReactNode } from 'react'
import { AlertTriangle, Music2, RefreshCw, X } from 'lucide-react'
import type { Track } from '@/shared/types/spotify'
import { artistNames } from './lib/format'

/* Shared hardware chrome: one window frame, one button, one row. */

export function Window({ code, plugin, onClose, children }: {
  code: string
  plugin: string
  onClose?: () => void
  children: ReactNode
}) {
  const name = windowName(code)
  return (
    <section className={`win win-${plugin}`} aria-label={name}>
      <header className="win-title">
        <h2>{code}</h2>
        {onClose && (
          <div className="win-controls">
            <button type="button" onClick={onClose} aria-label={`Close ${name}`} title="Close"><X /></button>
          </div>
        )}
      </header>
      <div className="win-body">
        <i className="screw tl" /><i className="screw tr" /><i className="screw bl" /><i className="screw br" />
        {children}
      </div>
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
