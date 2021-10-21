import { html, useMemo } from 'https://unpkg.com/htm/preact/standalone.module.js'

/**
 * @typedef Artist
 * @property {string} name
 * @property {{ spotify: string }} external_urls
 *
 * @typedef Album
 * @property {string} name
 * @property {{ spotify: string }} external_urls
 *
 * @typedef Track
 * @property {string} name
 * @property {Album} album
 * @property {Artist[]} artists
 * @property {{ spotify: string }} external_urls
 */

/**
 * @param {Artist[]} artists
 */
const mapArtistNames = (artists) => artists.map((a) => a.name).join(', ')

const LinkableName = ({ url, name }) => {
  return html` <a href="${url}">${name}</a> `
}

/**
 * @param {{ artists: Artist[] }} opts
 * @returns
 */
const LinkableArtists = ({ artists }) => {
  return html`
    ${artists.map((a, i, arr) => {
      return html`<${LinkableName} name=${a.name} url=${a.external_urls.spotify} /> ${i < arr.length - 1
          ? ','
          : ''}`
    })}
  `
}

/**
 * @param {object} opts
 * @param {Track} opts.track
 */
const NowPlaying = ({ track }) => {
  const { artistNames, albumName } = useMemo(
    () => ({
      artistNames: html`<${LinkableArtists} artists=${track.artists} />`,
      albumName: html`<${LinkableName} name=${track.album.name} url=${track.album.external_urls.spotify} />`,
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
          return html`Now playing <${LinkableName} name=${track.name} url=${track.external_urls.spotify} />`
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
    <article class="playing">
      <${PlayingHeader} isPlaying=${isPlaying} track=${track} title=${title} />
      <${isPlaying ? NowPlaying : NothingPlaying} track=${track} />
    </article>
  `
}

export default Playing
