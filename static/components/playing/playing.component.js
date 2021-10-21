import { html, useMemo } from 'https://unpkg.com/htm/preact/standalone.module.js'

import { SpotifyLinkable, LinkableArtists } from '../linkable/linkable-component.js'

/**
 * @typedef {import('../../recommendli/client').Artist} Artist
 * @typedef {import('../../recommendli/client').Album} Album
 * @typedef {import('../../recommendli/client').Track} Track
 */

/**
 * @param {Artist[]} artists
 */
const mapArtistNames = (artists) => artists.map((a) => a.name).join(', ')

/**
 * @param {object} opts
 * @param {Track} opts.track
 */
const NowPlaying = ({ track }) => {
  const { artistNames, albumName } = useMemo(
    () => ({
      artistNames: html`<${LinkableArtists} artists=${track.artists} />`,
      albumName: html`<${SpotifyLinkable} item=${track.album} />`,
    }),
    [mapArtistNames(track.artists), track.album.name]
  )

  return html`
    <div><small>by</small> <strong>${artistNames}</strong></div>
    <div><small>on</small> <strong>${albumName}</strong></div>
  `
}

const NothingPlaying = () => {
  return html`
    <div>
      <div>${'\u00a0'}</div>
      <div>${'\u00a0'}</div>
    </div>
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
 * @param {Track} opts.track
 */
const Playing = ({ title, isPlaying, track }) => {
  return html`
    <article>
      <${PlayingHeader} isPlaying=${isPlaying} track=${track} title=${title} />
      <${isPlaying ? NowPlaying : NothingPlaying} track=${track} />
    </article>
  `
}

export default Playing
