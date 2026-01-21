import { useCurrentTrack, useCheckCurrentTrack } from '@/shared/api/queries'
import { useDocumentVisibility } from '@/shared/hooks/useDocumentVisibility'
import { SpotifyLink } from '@/shared/components/SpotifyLink'
import { ArtistLinks } from '@/shared/components/ArtistLinks'

export function NowPlayingPanel() {
  const isVisible = useDocumentVisibility()

  const { data: currentTrack } = useCurrentTrack(
    isVisible,
    isVisible ? 2000 : false
  )

  const track = currentTrack?.track
  const isPlaying = currentTrack?.is_playing ?? false

  const { data: trackStatus, isLoading: statusLoading } = useCheckCurrentTrack(
    track?.id,
    isPlaying
  )

  if (!track) {
    return (
      <article className="bg-white/80 backdrop-blur-sm rounded-2xl shadow-xl border border-slate-200/60 p-6 hover:shadow-2xl transition-all duration-300">
        <header className="mb-6">
          <h2 className="text-xl font-bold text-slate-800 flex items-center gap-2">
            <span className="w-1 h-6 bg-gradient-to-b from-green-500 to-emerald-500 rounded-full"></span>
            Now Playing
          </h2>
        </header>
        <div className="flex items-center gap-3 text-slate-500">
          <div className="w-12 h-12 rounded-lg bg-slate-100 flex items-center justify-center">
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19V6l12-3v13M9 19c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zm12-3c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zM9 10l12-3" />
            </svg>
          </div>
          <p className="text-sm font-medium">Nothing playing</p>
        </div>
      </article>
    )
  }

  return (
    <article className="bg-white/80 backdrop-blur-sm rounded-2xl shadow-xl border border-slate-200/60 p-6 hover:shadow-2xl transition-all duration-300">
      <header className="mb-6">
        <h2 className="text-xl font-bold text-slate-800 flex items-center gap-2">
          <span className="w-1 h-6 bg-gradient-to-b from-green-500 to-emerald-500 rounded-full"></span>
          Now Playing
        </h2>
      </header>
      <div className="space-y-3">
        <div className="font-semibold text-lg text-slate-900">
          <SpotifyLink item={track} />
        </div>
        <div className="text-sm text-slate-600">
          by <ArtistLinks artists={track.artists} />
        </div>
        <div className="text-sm text-slate-600">
          on <SpotifyLink item={track.album} />
        </div>

        {isPlaying && (
          <div className="mt-6 pt-4 border-t border-slate-200">
            {statusLoading ? (
              <p className="text-sm text-slate-500 flex items-center gap-2" aria-busy="true">
                <svg className="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Checking track status...
              </p>
            ) : trackStatus?.in_library ? (
              <details className="text-sm">
                <summary className="cursor-pointer flex items-center gap-2">
                  <span className="px-2 py-1 bg-amber-100 text-amber-800 rounded-md font-medium">
                    In {trackStatus.playlists.length} playlist{trackStatus.playlists.length !== 1 ? 's' : ''}
                  </span>
                </summary>
                <ul className="mt-3 ml-4 space-y-2">
                  {trackStatus.playlists.map((playlist) => (
                    <li key={playlist.id} className="flex items-center gap-2">
                      <span className="w-1.5 h-1.5 rounded-full bg-blue-500"></span>
                      <SpotifyLink item={playlist} />
                    </li>
                  ))}
                </ul>
              </details>
            ) : (
              <div className="px-3 py-2 bg-gradient-to-r from-green-50 to-emerald-50 border border-green-200 rounded-lg">
                <p className="text-sm text-green-700 font-semibold flex items-center gap-2">
                  <span className="text-lg">🎉</span>
                  Track is new!
                </p>
              </div>
            )}
          </div>
        )}
      </div>
    </article>
  )
}
