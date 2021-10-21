import { html } from 'https://unpkg.com/htm/preact/standalone.module.js'
import { LinkableArtists, SpotifyLinkable } from '../linkable/linkable-component.js'

/**
 * @typedef {import('../../recommendli/client').Artist} Artist
 * @typedef {import('../../recommendli/client').Album} Album
 * @typedef {import('../../recommendli/client').Track} Track
 * @typedef {import('../../recommendli/client').Playlist} Playlist
 */

/**
 *
 * @param {{ tracks: Track[] }} args
 * @returns
 */
const TrackTable = ({ tracks }) => {
  return html`
    <table role="grid">
      <thead>
        <tr>
          <th scope="col">#</th>
          <th scope="col">Name</th>
          <th scope="col">Artists</th>
          <th scope="col">Album</th>
        </tr>
      </thead>
      <tbody>
        ${tracks.map(
          (track, i) => html`
            <tr>
              <th scope="row">${i + 1}</th>
              <td><${SpotifyLinkable} item=${track} /></td>
              <td><${LinkableArtists} artists=${track.artists} /></td>
              <td><${SpotifyLinkable} item=${track.album} /></td>
            </tr>
          `
        )}
      </tbody>
    </table>
  `
}

/**
 * @param {object} args
 * @param {() => void} args.onGeneratePlaylist
 * @param {boolean} args.isLoading
 * @param {Playlist} args.playlist
 * @returns
 */
const DiscoveryPlaylist = ({ onGeneratePlaylist, isLoading, playlist }) => {
  const onClick = () => {
    if (onGeneratePlaylist) {
      onGeneratePlaylist()
    }
  }

  return html`
    <article>
      <header>Discovery</header>
      <button aria-busy=${isLoading} disabled=${isLoading} onClick=${onClick}>
        Generate discover playlist
      </button>
      ${playlist == null
        ? null
        : html`
            <h2><${SpotifyLinkable} item=${playlist} /></h2>
            <${TrackTable} tracks=${playlist.tracks} />
          `}
    </article>
  `
}

export default DiscoveryPlaylist
