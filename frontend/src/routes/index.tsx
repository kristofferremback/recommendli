import { Database, Music2, Radio, ShieldCheck } from 'lucide-react'
import { useCurrentUser } from '@/shared/api/queries'
import { PluginWorkspace } from '@/features/workspace/PluginWorkspace'
import { Led, Window } from '@/features/workspace/chrome'

export function Dashboard() {
  const user = useCurrentUser()
  if (user.isLoading) return <ConnectionPanel loading />
  if (!user.data) return <ConnectionPanel />
  return <PluginWorkspace />
}

function ConnectionPanel({ loading = false }: { loading?: boolean }) {
  return (
    <main className="page connect-page">
      <div className="connect">
        <Window code="R // CONNECTION" plugin="connect">
          <div className="brandbar"><strong>recommendli</strong><Led busy={loading} /></div>
          <div className="well">
            <div className="lcd connect-display" aria-live="polite">
              <Radio aria-hidden="true" />
              <strong>{loading ? 'Connecting' : 'Spotify offline'}</strong>
              <div className={`signal ${loading ? 'active' : ''}`} aria-hidden="true">{[1, 2, 3, 4, 5, 6].map(i => <i key={i} />)}</div>
            </div>
          </div>
          {!loading && (
            <div className="build">
              <a className="key primary big" href="/recommendations/v1/spotify/auth/ui-redirect?url=/"><Music2 /> Connect Spotify</a>
            </div>
          )}
          <div className="permissions">
            <span><Database aria-hidden="true" />Reads your playlists</span>
            <span><ShieldCheck aria-hidden="true" />Creates the discovery playlist</span>
            <span><Radio aria-hidden="true" />Controls the active device</span>
          </div>
        </Window>
      </div>
    </main>
  )
}
