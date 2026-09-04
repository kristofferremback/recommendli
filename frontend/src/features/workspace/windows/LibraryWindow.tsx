import { useMemo, useState } from 'react'
import { Clock3, ExternalLink, ListMusic, Music2, RefreshCw, Search, X } from 'lucide-react'
import { useIndexSummary } from '@/shared/api/queries'
import type { useSyncIndex } from '@/shared/api/queries'
import { ErrorChip, Key, Window } from '../chrome'
import { formatClock, padIndex } from '../lib/format'

export function LibraryWindow({ sync, close }: { sync: ReturnType<typeof useSyncIndex>; close: () => void }) {
  const index = useIndexSummary(false)
  const [search, setSearch] = useState('')
  const allPlaylists = index.data?.playlists ?? []
  const playlists = useMemo(() => {
    const needle = search.trim().toLowerCase()
    return needle ? allPlaylists.filter(playlist => playlist.name.toLowerCase().includes(needle)) : allPlaylists
  }, [allPlaylists, search])

  return (
    <Window code="R // LIBRARY" plugin="library" onClose={close}>
      <div className="library">
        <div className="lcd stats">
          <div className="stat"><Music2 aria-hidden="true" /><strong>{index.data?.unique_track_count.toLocaleString() ?? '—'}</strong><small>Tracks</small></div>
          <div className="stat"><ListMusic aria-hidden="true" /><strong>{index.data?.playlist_count.toLocaleString() ?? '—'}</strong><small>Playlists</small></div>
        </div>
        {(index.error || sync.error) && <ErrorChip message={sync.error ? 'Sync failed' : 'Library unavailable'} retry={() => sync.mutate()} />}
        <div className="field">
          <Search aria-hidden="true" />
          <label className="field-label" htmlFor="library-filter">Filter</label>
          <input id="library-filter" value={search} onChange={event => setSearch(event.target.value)} placeholder="Playlist name" />
          {search && <Key variant="dark" icon label="Clear filter" onClick={() => setSearch('')}><X /></Key>}
        </div>
        <div className="list playlists" role="region" aria-label="Playlists">
          <header className="list-head"><ListMusic aria-hidden="true" />{search ? 'Matching' : 'All playlists'}<b>{playlists.length}</b></header>
          {playlists.length > 0 ? (
            <ol>
              {playlists.map((playlist, i) => (
                <li className="row playlist" key={playlist.id}>
                  <span className="row-index">{padIndex(i + 1)}</span>
                  <div className="row-text"><b className="row-title">{playlist.name}</b></div>
                  <span className="row-meta">{playlist.tracks?.total ?? '—'}</span>
                  <div className="row-actions">
                    <a className="key dark icon" href={playlist.external_urls.spotify} target="_blank" rel="noreferrer" aria-label={`Open ${playlist.name} in Spotify`} title="Open in Spotify"><ExternalLink /></a>
                  </div>
                </li>
              ))}
            </ol>
          ) : (
            <div className="empty">{index.isLoading ? 'Loading' : search ? 'No playlist matches' : 'No playlists indexed'}</div>
          )}
        </div>
        <div className="panel-row sync-row">
          <Clock3 aria-hidden="true" />
          <span>{sync.isPending ? 'Syncing' : index.data?.last_synced_at ? `Synced ${formatClock(index.data.last_synced_at)}` : 'Never synced'}</span>
          <Key icon label="Sync library" disabled={sync.isPending} onClick={() => sync.mutate()}><RefreshCw className={sync.isPending ? 'spin' : ''} /></Key>
        </div>
      </div>
    </Window>
  )
}
