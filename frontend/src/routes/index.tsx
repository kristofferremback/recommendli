import { Database, Music2, Radio, ShieldCheck } from 'lucide-react'
import { useCurrentUser } from '@/shared/api/queries'
import { PluginWorkspace } from '@/features/workspace/components/PluginWorkspace'

export function Dashboard() {
  const user = useCurrentUser()
  if (user.isLoading) return <ConnectionPanel loading />
  if (!user.data) return <ConnectionPanel />
  return <PluginWorkspace />
}

function ConnectionPanel({ loading = false }: { loading?: boolean }) {
  return (
    <main className="rh-page rh-connection-page">
      <section className="rh-window rh-connection">
        <i className="rh-screw tl" /><i className="rh-screw tr" /><i className="rh-screw bl" /><i className="rh-screw br" />
        <header className="rh-title"><span>R // CONNECTION</span><div className="rh-window-dots"><i /><i /><i /></div></header>
        <div className="rh-brandbar"><strong>recommendli</strong><i className={`rh-connection-led ${loading ? 'busy' : ''}`} /></div>
        <div className="rh-well">
          <div className="rh-lcd rh-connection-display">
            <Radio aria-hidden="true" />
            <strong>{loading ? 'CONNECTING' : 'SPOTIFY OFFLINE'}</strong>
            <div className={`rh-signal ${loading ? 'active' : ''}`} aria-hidden="true">{[1, 2, 3, 4, 5, 6].map(i => <i key={i} />)}</div>
          </div>
        </div>
        {!loading && <a className="rh-connect-button" href="/recommendations/v1/spotify/auth/ui-redirect?url=/"><Music2 /> CONNECT SPOTIFY</a>}
        <div className="rh-permissions">
          <span><Database />Reads playlists</span>
          <span><ShieldCheck />Creates playlists</span>
          <span><Radio />Controls active device</span>
        </div>
      </section>
    </main>
  )
}
