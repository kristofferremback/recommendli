import { useState } from 'react'
import { Clock3, ListEnd, Play, SkipForward } from 'lucide-react'
import { usePlaybackControls, usePlaybackHistory, usePlaybackQueue } from '@/shared/api/queries'
import { useDocumentVisibility } from '@/shared/hooks/useDocumentVisibility'
import { usePlaybackProgress } from '@/shared/hooks/usePlaybackProgress'
import type { Playback } from '@/shared/types/spotify'
import { ErrorChip, Key, TrackList, TrackRow, Window } from '../chrome'
import { formatTime, padIndex } from '../lib/format'

export function QueueWindow({ playback, close }: { playback?: Playback; close: () => void }) {
  const [tab, setTab] = useState<'history' | 'queue'>('queue')
  const visible = useDocumentVisibility()
  const queue = usePlaybackQueue(visible)
  const history = usePlaybackHistory(visible)
  const controls = usePlaybackControls()
  const progress = usePlaybackProgress(playback)
  const historyItems = history.data ?? []
  const queueItems = queue.data?.tracks ?? []

  return (
    <Window code="R // QUEUE" plugin="queue" onClose={close}>
      <div className="now-line">
        <Play aria-hidden="true" />
        <b>{playback?.track?.name ?? 'Nothing playing'}</b>
        <span>{formatTime(progress)} / {formatTime(playback?.track?.duration_ms ?? 0)}</span>
      </div>
      <div className="tabs" role="group" aria-label="Timeline">
        <Key aria-pressed={tab === 'history'} lit={tab === 'history'} onClick={() => setTab('history')}><Clock3 /> History</Key>
        <Key aria-pressed={tab === 'queue'} lit={tab === 'queue'} onClick={() => setTab('queue')}><ListEnd /> Up next</Key>
      </div>
      {(queue.error || history.error) && <ErrorChip retry={() => { void queue.refetch(); void history.refetch() }} />}
      <div className="timeline">
        <TrackList icon={<Clock3 aria-hidden="true" />} title="History" count={historyItems.length} active={tab === 'history'} loading={history.isLoading} empty="Nothing played yet">
          {historyItems.map((item, i) => (
            <TrackRow
              key={`${item.played_at}-${item.track.id}`}
              index={`-${padIndex(i + 1)}`}
              track={item.track}
              actions={<Key variant="dark" icon label={`Play again: ${item.track.name}`} onClick={() => controls.play.mutate(item.track.id)}><Play /></Key>}
            />
          ))}
        </TrackList>
        <TrackList icon={<ListEnd aria-hidden="true" />} title="Up next" count={queueItems.length} active={tab === 'queue'} loading={queue.isLoading} empty="Queue is empty">
          {queueItems.map((track, i) => (
            <TrackRow
              key={`${track.id}-${i}`}
              index={`+${padIndex(i + 1)}`}
              track={track}
              actions={<Key variant="dark" icon label={`Skip to: ${track.name}`} onClick={() => controls.skipQueue.mutate({ position: i, expected_track_id: track.id, expected_current_track_id: queue.data?.currently_playing?.id })}><SkipForward /></Key>}
            />
          ))}
        </TrackList>
      </div>
    </Window>
  )
}
