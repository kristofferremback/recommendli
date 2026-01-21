import { useCurrentTrack, useCheckCurrentTrack } from '@/shared/api/queries'
import { useDocumentVisibility } from '@/shared/hooks/useDocumentVisibility'
import { SpotifyLink } from '@/shared/components/SpotifyLink'
import { ArtistLinks } from '@/shared/components/ArtistLinks'

export function NowPlayingHero() {
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
      <div className="relative min-h-[60vh] flex items-center justify-center p-8">
        <div className="text-center space-y-4">
          <div className="w-32 h-32 mx-auto rounded-full bg-white/5 backdrop-blur-sm border border-white/10 flex items-center justify-center">
            <svg className="w-16 h-16 text-white/30" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19V6l12-3v13M9 19c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zm12-3c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zM9 10l12-3" />
            </svg>
          </div>
          <p className="text-xl text-white/50 font-medium">Nothing playing right now</p>
          <p className="text-sm text-white/30">Start playing music on Spotify to see it here</p>
        </div>
      </div>
    )
  }

  const albumArt = track.album?.images?.[0]?.url || track.album?.images?.[1]?.url

  return (
    <div className="relative bg-black/40 backdrop-blur-xl border-b border-white/10">
      {/* Background with album art blur */}
      {albumArt && (
        <div className="absolute inset-0 bg-cover bg-center pointer-events-none opacity-60">
          <div
            className="absolute inset-0 backdrop-blur-3xl bg-black/60"
            style={{
              backgroundImage: `url(${albumArt})`,
              backgroundSize: 'cover',
              backgroundPosition: 'center',
            }}
          ></div>
        </div>
      )}

      {/* Content */}
      <div className="relative z-10 max-w-7xl mx-auto p-8 md:p-12">
        <div className="flex flex-col md:flex-row items-center gap-8">
          {/* Album Art */}
          <div className="flex-shrink-0 rounded-xl overflow-hidden shadow-2xl border-2 border-white/10 w-48 h-48">
            {albumArt ? (
              <img
                src={albumArt}
                alt={track.album.name}
                className="w-full h-full object-cover"
              />
            ) : (
              <div className="w-full h-full bg-gradient-to-br from-purple-600 to-blue-600"></div>
            )}
          </div>

          {/* Track Info */}
          <div className="flex-1 text-center md:text-left space-y-3">
            <div className="flex items-center gap-3 justify-center md:justify-start">
              {isPlaying && (
                <span className="w-3 h-3 bg-green-500 rounded-full animate-pulse flex-shrink-0"></span>
              )}
              <h1 className="text-4xl md:text-5xl font-black text-white leading-tight">
                {track.name}
              </h1>
            </div>

            <p className="text-xl md:text-2xl text-white/70">
              <ArtistLinks artists={track.artists} />
            </p>

            <p className="text-lg text-white/50">
              <SpotifyLink item={track.album} />
            </p>

            {/* Library status */}
            {!statusLoading && trackStatus && (
              <div className="pt-4">
                {trackStatus.in_library ? (
                  <details className="inline-block">
                    <summary className="cursor-pointer text-amber-300 bg-amber-500/30 hover:bg-amber-500/40 px-4 py-2 rounded-full text-sm font-semibold border border-amber-500/50 hover:border-amber-500/70 transition-all">
                      In {trackStatus.playlists.length} playlist{trackStatus.playlists.length !== 1 ? 's' : ''}
                    </summary>
                    <div className="mt-3 p-4 bg-black/80 backdrop-blur-sm rounded-xl space-y-1 max-w-md border border-white/10">
                      {trackStatus.playlists.slice(0, 10).map((playlist) => (
                        <div key={playlist.id} className="text-white/70 text-sm">
                          • <SpotifyLink item={playlist} />
                        </div>
                      ))}
                      {trackStatus.playlists.length > 10 && (
                        <p className="text-white/40 text-sm italic">
                          ...and {trackStatus.playlists.length - 10} more
                        </p>
                      )}
                    </div>
                  </details>
                ) : (
                  <span className="text-green-300 bg-green-500/30 px-4 py-2 rounded-full text-sm font-semibold border border-green-500/50">
                    New track!
                  </span>
                )}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
