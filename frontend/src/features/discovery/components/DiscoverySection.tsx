import { useState } from 'react'
import { toast } from 'sonner'
import { useGenerateDiscoveryPlaylist } from '@/shared/api/queries'
import { SpotifyLink } from '@/shared/components/SpotifyLink'
import { ArtistLinks } from '@/shared/components/ArtistLinks'
import type { Playlist } from '@/shared/types/spotify'

export function DiscoverySection() {
  const [playlist, setPlaylist] = useState<Playlist | null>(null)
  const generateMutation = useGenerateDiscoveryPlaylist()

  const handleGenerate = async () => {
    try {
      const result = await generateMutation.mutateAsync(false)
      setPlaylist(result)
      toast.success('Discovery playlist generated!', {
        description: `${result.tracks.length} personalized tracks ready for you`
      })
    } catch (error) {
      toast.error('Failed to generate playlist', {
        description: 'Please try again in a moment'
      })
    }
  }

  return (
    <div className="p-4 md:p-8">
      <div className="container mx-auto">
        {/* Header */}
        <div className="flex flex-col md:flex-row items-start md:items-center justify-between gap-4 md:gap-0 mb-6 md:mb-8">
          <div>
            <h2 className="text-2xl md:text-4xl font-black text-white mb-1 md:mb-2">Discovery Playlist</h2>
            <p className="text-white/60 text-sm md:text-lg">AI-curated tracks based on your taste</p>
          </div>

          <button
            onClick={handleGenerate}
            disabled={generateMutation.isPending}
            className="group relative w-full md:w-auto px-6 md:px-8 py-3 md:py-4 bg-gradient-to-r from-purple-600 via-pink-600 to-blue-600 text-white font-bold text-base md:text-lg rounded-2xl shadow-2xl hover:shadow-purple-500/50 hover:scale-105 active:scale-95 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100 overflow-hidden"
          >
            <span className="relative z-10 flex items-center justify-center gap-2 md:gap-3">
              {generateMutation.isPending ? (
                <>
                  <svg className="animate-spin h-5 w-5 md:h-6 md:w-6" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  Generating...
                </>
              ) : (
                <>
                  <svg className="w-5 h-5 md:w-6 md:h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                  </svg>
                  Generate Playlist
                </>
              )}
            </span>
            <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/20 to-transparent translate-x-[-200%] group-hover:translate-x-[200%] transition-transform duration-1000"></div>
          </button>
        </div>

        {/* Track Grid */}
        {playlist && playlist.tracks.length > 0 && (
          <div className="space-y-4">
            {/* Desktop header row */}
            <div className="hidden md:flex items-center justify-between text-white/50 text-sm uppercase tracking-wider font-semibold px-4">
              <span className="w-12">#</span>
              <span className="flex-1">Track</span>
              <span className="w-48">Artists</span>
              <span className="w-48">Album</span>
            </div>

            <div className="space-y-2">
              {playlist.tracks.map((track, i) => {
                const albumArt = track.album?.images?.[2]?.url || track.album?.images?.[1]?.url
                return (
                  <div
                    key={track.id}
                    className="group bg-white/5 hover:bg-white/10 backdrop-blur-sm rounded-xl p-3 md:p-4 transition-all duration-200 hover:scale-[1.01] border border-white/0 hover:border-white/10"
                  >
                    {/* Desktop layout */}
                    <div className="hidden md:flex items-center gap-4">
                      <span className="w-8 text-right text-white/40 font-mono text-sm group-hover:text-white/80">
                        {i + 1}
                      </span>

                      {albumArt && (
                        <div className="w-12 h-12 rounded-lg overflow-hidden flex-shrink-0 shadow-lg">
                          <img
                            src={albumArt}
                            alt={track.album.name}
                            className="w-full h-full object-cover"
                          />
                        </div>
                      )}

                      <div className="flex-1 min-w-0">
                        <div className="text-white font-semibold truncate group-hover:text-green-400 transition-colors">
                          <SpotifyLink item={track} />
                        </div>
                      </div>

                      <div className="w-48 text-white/70 text-sm truncate">
                        <ArtistLinks artists={track.artists} />
                      </div>

                      <div className="w-48 text-white/50 text-sm truncate">
                        <SpotifyLink item={track.album} />
                      </div>
                    </div>

                    {/* Mobile layout */}
                    <div className="md:hidden flex items-center gap-3">
                      <span className="w-6 text-right text-white/40 font-mono text-xs">
                        {i + 1}
                      </span>

                      {albumArt && (
                        <div className="w-10 h-10 rounded-lg overflow-hidden flex-shrink-0 shadow-lg">
                          <img
                            src={albumArt}
                            alt={track.album.name}
                            className="w-full h-full object-cover"
                          />
                        </div>
                      )}

                      <div className="flex-1 min-w-0">
                        <div className="text-white text-sm font-semibold truncate group-hover:text-green-400 transition-colors">
                          <SpotifyLink item={track} />
                        </div>
                        <div className="text-white/60 text-xs truncate mt-0.5">
                          <ArtistLinks artists={track.artists} />
                        </div>
                      </div>
                    </div>
                  </div>
                )
              })}
            </div>

            <div className="text-center pt-6">
              <p className="text-white/40 text-sm">
                {playlist.tracks.length} track{playlist.tracks.length !== 1 ? 's' : ''} • Powered by your listening history
              </p>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
