import { useState } from 'react'
import { toast } from 'sonner'
import { useGenerateDiscoveryPlaylist } from '@/shared/api/queries'
import { GenerateButton } from './GenerateButton'
import { TrackTable } from './TrackTable'
import type { Playlist } from '@/shared/types/spotify'

/**
 * Discovery Panel - orchestrates playlist generation
 * Follows composition pattern: delegates rendering to smaller components
 */
export function DiscoveryPanel() {
  const [playlist, setPlaylist] = useState<Playlist | null>(null)
  const generateMutation = useGenerateDiscoveryPlaylist()

  const handleGenerate = async () => {
    try {
      const result = await generateMutation.mutateAsync(false)
      setPlaylist(result)
      toast.success('Playlist generated!', {
        description: `${result.tracks.length} tracks added to your recommendations`
      })
    } catch (error) {
      toast.error('Failed to generate playlist', {
        description: 'Please try again in a moment'
      })
    }
  }

  return (
    <article className="bg-white/80 backdrop-blur-sm rounded-2xl shadow-xl border border-slate-200/60 p-6 hover:shadow-2xl transition-all duration-300">
      <header className="mb-6">
        <h2 className="text-xl font-bold text-slate-800 flex items-center gap-2">
          <span className="w-1 h-6 bg-gradient-to-b from-blue-500 to-indigo-500 rounded-full"></span>
          Discovery Playlist
        </h2>
      </header>
      <GenerateButton
        onGenerate={handleGenerate}
        isLoading={generateMutation.isPending}
      />
      {playlist && <TrackTable tracks={playlist.tracks} />}
    </article>
  )
}
