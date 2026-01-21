import { useIndexSummary } from '@/shared/api/queries'
import { useDocumentVisibility } from '@/shared/hooks/useDocumentVisibility'
import { SpotifyLink } from '@/shared/components/SpotifyLink'

export function LibrarySummaryPanel() {
  const isVisible = useDocumentVisibility()

  const { data: summary } = useIndexSummary(
    isVisible ? 20000 : false
  )

  return (
    <article className="bg-white/80 backdrop-blur-sm rounded-2xl shadow-xl border border-slate-200/60 p-6 hover:shadow-2xl transition-all duration-300">
      <header className="mb-6">
        <h2 className="text-xl font-bold text-slate-800 flex items-center gap-2">
          <span className="w-1 h-6 bg-gradient-to-b from-purple-500 to-pink-500 rounded-full"></span>
          Library Index
        </h2>
      </header>
      {summary ? (
        <div className="space-y-4">
          <div className="grid grid-cols-2 gap-3">
            <div className="bg-gradient-to-br from-blue-50 to-indigo-50 rounded-xl p-4 border border-blue-100">
              <div className="text-xs font-semibold text-blue-600 uppercase tracking-wider mb-1">Tracks</div>
              <div className="text-2xl font-bold text-blue-900">{summary.unique_track_count.toLocaleString()}</div>
            </div>
            <div className="bg-gradient-to-br from-purple-50 to-pink-50 rounded-xl p-4 border border-purple-100">
              <div className="text-xs font-semibold text-purple-600 uppercase tracking-wider mb-1">Playlists</div>
              <div className="text-2xl font-bold text-purple-900">{summary.playlist_count}</div>
            </div>
          </div>
          <div>
            <details>
              <summary className="cursor-pointer">
                <span className="font-medium">
                  {summary.playlists.length} {summary.playlists.length === 1 ? 'playlist' : 'playlists'} indexed
                </span>
              </summary>
              <ul className="mt-3 space-y-2 max-h-64 overflow-y-auto">
                {summary.playlists
                  .sort((a, b) => b.name.localeCompare(a.name, 'en-US', { numeric: true }))
                  .map((playlist) => (
                    <li key={playlist.id} className="flex items-center gap-2 pl-4">
                      <span className="w-1.5 h-1.5 rounded-full bg-purple-500"></span>
                      <SpotifyLink item={playlist} />
                    </li>
                  ))}
              </ul>
            </details>
          </div>
        </div>
      ) : (
        <div className="flex items-center justify-center py-8">
          <div className="text-center text-slate-400">
            <svg className="w-12 h-12 mx-auto mb-2 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
            <p className="text-sm font-medium">Loading summary...</p>
          </div>
        </div>
      )}
    </article>
  )
}
