import { SpotifyLink } from './SpotifyLink'
import type { Artist } from '@/shared/types/spotify'

interface ArtistLinksProps {
  artists: Artist[]
}

export function ArtistLinks({ artists }: ArtistLinksProps) {
  return (
    <>
      {artists.map((artist, i) => (
        <span key={artist.id}>
          <SpotifyLink item={artist} />
          {i < artists.length - 1 && ', '}
        </span>
      ))}
    </>
  )
}
