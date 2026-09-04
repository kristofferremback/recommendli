import { Database, Disc3, ListMusic, Plus } from 'lucide-react'
import { useGenerateDiscoveryPlaylist } from '@/shared/api/queries'
import type { Playlist } from '@/shared/types/spotify'
import { ErrorChip, Key, Led, Window } from '../chrome'

export function DiscoveryWindow({ indexCount, onResult, close }: {
  indexCount?: number
  onResult: (playlist: Playlist) => void
  close: () => void
}) {
  const generation = useGenerateDiscoveryPlaylist()
  const build = () => generation.mutateAsync(false).then(onResult, () => undefined)
  const building = generation.isPending

  return (
    <Window code="R // DISCOVERY" plugin="discovery" onClose={close}>
      <div className="well builder">
        <div className="sources">
          <Source name="Discover Weekly" />
          <Source name="Release Radar" orange />
        </div>
        <div className="routing" aria-hidden="true"><i /></div>
        <div className="filter">
          <Database aria-hidden="true" />
          <b>{indexCount?.toLocaleString() ?? '—'}</b>
          <small>known tracks filtered out</small>
        </div>
        <div className="lcd output" aria-live="polite">
          <ListMusic aria-hidden="true" />
          <strong>{generation.data?.name ?? 'Recommendli Discovery'}</strong>
          <span>{generation.data?.tracks.length ?? 0}</span>
        </div>
      </div>
      {generation.error && <ErrorChip message="Build failed" retry={build} />}
      <div className="build">
        <Key variant="primary" big disabled={building} onClick={build}>
          <Plus /> {building ? 'Building' : generation.data ? 'Build again' : 'Build playlist'}
        </Key>
      </div>
      <div className="stage panel-row">
        <Led on={generation.isSuccess} busy={building} />
        <span>{building ? 'Reading sources and filtering' : generation.isSuccess ? 'Playlist saved to Spotify' : 'Ready'}</span>
        {building && <div className="meter playing" aria-hidden="true">{[40, 70, 55, 85, 60].map((height, i) => <i key={i} style={{ height: `${height}%` }} />)}</div>}
      </div>
    </Window>
  )
}

function Source({ name, orange }: { name: string; orange?: boolean }) {
  return (
    <div className="source">
      <span className={`source-icon ${orange ? 'orange' : ''}`}><Disc3 aria-hidden="true" /></span>
      <div><b>{name}</b><small>Source</small></div>
      <Led on />
    </div>
  )
}
