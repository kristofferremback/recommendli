import { SpotifyLink } from '@/shared/components/SpotifyLink'
import { ArtistLinks } from '@/shared/components/ArtistLinks'
import type { Track } from '@/shared/types/spotify'

interface TrackTableProps {
  tracks: Track[]
}

export function TrackTable({ tracks }: TrackTableProps) {
  return (
    <div className="mt-6 -mx-2">
      <div className="max-h-96 overflow-y-auto rounded-lg">
        <table className="w-full text-sm" role="grid">
          <thead className="sticky top-0 bg-slate-50/95 backdrop-blur-sm">
            <tr>
              <th scope="col" className="text-left w-12">#</th>
              <th scope="col" className="text-left">Track</th>
              <th scope="col" className="text-left">Artists</th>
              <th scope="col" className="text-left">Album</th>
            </tr>
          </thead>
          <tbody>
            {tracks.map((track, i) => (
              <tr key={track.id} className="group">
                <td className="text-slate-500 font-mono text-xs">{i + 1}</td>
                <td className="font-medium">
                  <SpotifyLink item={track} />
                </td>
                <td className="text-slate-600">
                  <ArtistLinks artists={track.artists} />
                </td>
                <td className="text-slate-600">
                  <SpotifyLink item={track.album} />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <div className="mt-4 text-center">
        <span className="text-xs font-medium text-slate-500 bg-slate-100 px-3 py-1 rounded-full">
          {tracks.length} track{tracks.length !== 1 ? 's' : ''} generated
        </span>
      </div>
    </div>
  )
}
