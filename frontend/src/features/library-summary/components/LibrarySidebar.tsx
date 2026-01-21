import { useState } from 'react'
import { useIndexSummary } from '@/shared/api/queries'
import { useDocumentVisibility } from '@/shared/hooks/useDocumentVisibility'
import { SpotifyLink } from '@/shared/components/SpotifyLink'

export function LibrarySidebar() {
  const isVisible = useDocumentVisibility()
  const [showPlaylists, setShowPlaylists] = useState(true)

  const { data: summary } = useIndexSummary(
    isVisible ? 20000 : false
  )

  const sortedPlaylists = summary?.playlists
    ? [...summary.playlists].sort((a, b) => b.name.localeCompare(a.name, 'en-US', { numeric: true }))
    : []

  return (
    <div className="w-full xl:w-96 bg-black/40 backdrop-blur-xl border-t xl:border-t-0 xl:border-l border-white/10 flex flex-col xl:h-screen xl:sticky xl:top-0">
      {/* Header */}
      <div className="p-6 md:p-8 border-b border-white/10 flex-shrink-0">
        <h2 className="text-xl md:text-2xl font-black text-white mb-2">Your Library</h2>
        <p className="text-white/50 text-xs md:text-sm">Indexed playlists & tracks</p>
      </div>

      {/* Stats */}
      {summary ? (
        <>
          <div className="p-6 md:p-8 space-y-4 md:space-y-6 flex-shrink-0">
            {/* Track Count */}
            <div className="relative group">
              <div className="absolute inset-0 bg-gradient-to-r from-purple-600/20 to-pink-600/20 rounded-2xl blur-xl group-hover:blur-2xl transition-all"></div>
              <div className="relative bg-gradient-to-br from-purple-900/50 to-pink-900/50 backdrop-blur-sm rounded-2xl p-4 md:p-6 border border-purple-500/20">
                <div className="text-purple-300 text-xs md:text-sm font-semibold uppercase tracking-wider mb-1 md:mb-2">
                  Total Tracks
                </div>
                <div className="text-3xl md:text-5xl font-black text-white">
                  {summary.unique_track_count.toLocaleString()}
                </div>
              </div>
            </div>

            {/* Playlist Count */}
            <div className="relative group">
              <div className="absolute inset-0 bg-gradient-to-r from-blue-600/20 to-indigo-600/20 rounded-2xl blur-xl group-hover:blur-2xl transition-all"></div>
              <div className="relative bg-gradient-to-br from-blue-900/50 to-indigo-900/50 backdrop-blur-sm rounded-2xl p-4 md:p-6 border border-blue-500/20">
                <div className="text-blue-300 text-xs md:text-sm font-semibold uppercase tracking-wider mb-1 md:mb-2">
                  Playlists
                </div>
                <div className="text-3xl md:text-5xl font-black text-white">
                  {summary.playlist_count}
                </div>
              </div>
            </div>
          </div>

          {/* Playlist List */}
          <div className="flex-1 overflow-hidden flex flex-col min-h-0">
            <div className="px-6 md:px-8 py-4 border-t border-white/10 flex-shrink-0">
              <button
                onClick={() => setShowPlaylists(!showPlaylists)}
                className="flex items-center gap-2 text-base md:text-lg font-bold text-white hover:text-purple-400 transition-colors w-full text-left"
              >
                <svg
                  className={`w-4 h-4 transition-transform ${showPlaylists ? 'rotate-90' : ''}`}
                  fill="currentColor"
                  viewBox="0 0 20 20"
                >
                  <path fillRule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clipRule="evenodd" />
                </svg>
                All Playlists ({summary.playlists.length})
              </button>
            </div>

            {showPlaylists && (
              <div className="flex-1 overflow-y-auto px-6 md:px-8 pb-6 md:pb-8 min-h-0">
                {sortedPlaylists.map((playlist) => (
                  <div
                    key={playlist.id}
                    className="group flex items-center gap-3 py-2 hover:text-white/100 transition-colors"
                  >
                    <div className="w-1.5 h-1.5 md:w-2 md:h-2 rounded-full bg-purple-500/50 group-hover:bg-pink-500 transition-colors flex-shrink-0"></div>
                    <div className="flex-1 min-w-0">
                      <div className="text-white/70 group-hover:text-white text-xs md:text-sm truncate">
                        <SpotifyLink item={playlist} />
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </>
      ) : (
        <div className="flex-1 flex items-center justify-center p-8">
          <div className="text-center space-y-4">
            <div className="w-12 h-12 md:w-16 md:h-16 mx-auto border-4 border-purple-500/30 rounded-full animate-spin border-t-purple-500"></div>
            <p className="text-white/50 text-sm md:text-base">Loading library...</p>
          </div>
        </div>
      )}
    </div>
  )
}
