import html from '../../lib/html.js'
import { useMemo, useCallback, useState } from '../../deps/preact/hooks.js'

import { SpotifyLinkable, LinkableArtists } from '../linkable/linkable-component.js'
import { LoadingText } from '../conditional-loading/conditional-loading.component.js'

/**
 * @typedef {import('../../recommendli/client').Artist} Artist
 * @typedef {import('../../recommendli/client').Album} Album
 * @typedef {import('../../recommendli/client').Track} Track
 * @typedef {import('../../recommendli/client').SimplePlaylist} SimplePlaylist
 */

/**
 * @param {Artist[]} artists
 */
const mapArtistNames = (artists) => artists.map((a) => a.name).join(', ')

/**
 * @param {object} opts
 * @param {Track} opts.track
 * @param {boolean} opts.inLibrary
 * @param {boolean} opts.inLibraryLoading
 * @param {SimplePlaylist[]} opts.playlists
 */
const NowPlaying = ({ track, inLibrary, inLibraryLoading, playlists }) => {
  const [isOpen, setIsOpen] = useState(false)
  const toggleIsOpen = useCallback(() => setIsOpen((o) => !o), [setIsOpen])

  const { artistNames, albumName } = useMemo(
    () => ({
      artistNames: html`<${LinkableArtists} artists=${track.artists} />`,
      albumName: html`<${SpotifyLinkable} item=${track.album} />`,
    }),
    [mapArtistNames(track.artists), track.album.name]
  )

  const possiblyPlural = useMemo(() => `playlist${playlists.length !== 1 ? 's' : ''}`, [playlists.length])

  const InLibraryComponent = useCallback(() => {
    if (inLibraryLoading) {
      return html`<${LoadingText}>Checking track status</${LoadingText}>`
    }

    return !inLibrary
      ? html`<div>Track is new! 🎉</div>`
      : html`
          <div>
            <details open=${isOpen} onclick=${toggleIsOpen}>
              <summary>Track already on <strong>${playlists.length}</strong> ${possiblyPlural}</summary>
              <ul>
                ${playlists.map((playlist) => html`<li><${SpotifyLinkable} item=${playlist} /></li>`)}
              </ul>
            </details>
          </div>
        `
  }, [inLibraryLoading, inLibrary, playlists.map((p) => p.id).join(',')])

  return html`
    <div><small>by</small> <strong>${artistNames}</strong></div>
    <div><small>on</small> <strong>${albumName}</strong></div>
    <br />

    <${InLibraryComponent} />
  `
}

const NothingPlaying = () => {
  return html`
    <div>${'\u00a0'}</div>
    <div>${'\u00a0'}</div>
    <br />
    <div>${'\u00a0'}</div>
  `
}

/**
 * @param {{ isPlaying: boolean, track: Track, title?: string }} args
 */
const PlayingHeader = ({ isPlaying, track, title }) => {
  return html`
    <header>
      ${(() => {
        if (title) {
          return html`${title}`
        }
        if (isPlaying) {
          return html`Now playing <${SpotifyLinkable} item=${track} />`
        }
        return html`Nothing is playing`
      })()}
    </header>
  `
}

/**
 * @param {object} opts
 * @param {string} [opts.title]
 * @param {boolean} opts.isPlaying
 * @param {boolean} opts.inLibrary
 * @param {boolean} opts.inLibraryLoading
 * @param {SimplePlaylist[]} opts.playlists
 * @param {Track} opts.track
 */
const Playing = ({ title, isPlaying, track, inLibrary, inLibraryLoading, playlists }) => {
  return html`
    <article>
      <${PlayingHeader} isPlaying=${isPlaying} track=${track} title=${title} />
      <${isPlaying ? NowPlaying : NothingPlaying}
        track=${track}
        inLibrary=${inLibrary}
        playlists=${playlists}
        inLibraryLoading=${inLibraryLoading}
      />
    </article>
  `
}

export default Playing
