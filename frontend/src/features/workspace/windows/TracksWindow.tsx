import { ExternalLink, ListMusic, Play } from 'lucide-react'
import { usePlaybackControls } from '@/shared/api/queries'
import type { Playback, Playlist } from '@/shared/types/spotify'
import { Key, TrackRow, Window } from '../chrome'
import { formatTime, padIndex } from '../lib/format'

export function TracksWindow({ playlist, playback, close }: { playlist: Playlist | null; playback?: Playback; close: () => void }) {
  const controls = usePlaybackControls()
  const controllable = !!playback?.device && !playback.device.is_restricted
  return (
    <Window code="R // TRACKS" plugin="tracks" onClose={close}>
      <div className="lcd result-head">
        <ListMusic aria-hidden="true" />
        <b>{playlist?.name ?? 'No playlist yet'}</b>
        <strong>{playlist?.tracks.length ?? 0}</strong>
        {playlist && <a className="key dark icon" href={playlist.external_urls.spotify} target="_blank" rel="noreferrer" aria-label="Open playlist in Spotify" title="Open in Spotify"><ExternalLink /></a>}
      </div>
      <div className="list tracks">
        {playlist ? (
          <ol>
            {playlist.tracks.map((track, i) => (
              <TrackRow
                key={track.id}
                index={padIndex(i + 1)}
                track={track}
                album={track.album.name}
                duration={formatTime(track.duration_ms ?? 0)}
                actions={
                  <>
                    {controllable && <Key variant="dark" icon label={`Play ${track.name}`} title="Play on the active device, replaces the current queue" onClick={() => controls.play.mutate(track.id)}><Play /></Key>}
                    <a className="key dark icon" href={track.external_urls.spotify} target="_blank" rel="noreferrer" aria-label={`Open ${track.name} in Spotify`} title="Open in Spotify"><ExternalLink /></a>
                  </>
                }
              />
            ))}
          </ol>
        ) : (
          <div className="empty"><ListMusic aria-hidden="true" />Build a discovery playlist to list it here</div>
        )}
      </div>
    </Window>
  )
}
