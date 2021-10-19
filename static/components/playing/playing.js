import { html, useMemo } from 'https://unpkg.com/htm/preact/standalone.module.js'

/**
 * @param {{ name: string }[]} artists
 */
const formatArtists = (artists) => artists.map((a) => a.name).join(', ')

/**
 * @param {object} opts
 * @param {object} opts.track
 * @param {string} opts.track.name
 * @param {{ name: string }} opts.track.album
 * @param {{ name: string }[]} opts.track.artists
 */
const NowPlaying = ({ track }) => {
  const { name, artistNames, albumName } = useMemo(
    () => ({
      name: track.name,
      artistNames: formatArtists(track.artists),
      albumName: track.album.name,
    }),
    [track.name, formatArtists(track.artists), track.album.name]
  )

  return html`
    <div>
      <div><strong>${name}</strong></div>
      <div><small>by</small> <strong>${artistNames}</strong></div>
      <div><small>on</small> <strong>${albumName}</strong></div>
    </div>
  `
}

const NothingPlaying = () => {
  return html`
    <div>
      <div>${'\u00a0'}</div>
      <div>${'\u00a0'}</div>
      <div>${'\u00a0'}</div>
    </div>
  `
}

/**
 * @param {object} opts
 * @param {boolean} opts.isPlaying
 * @param {object} opts.track
 * @param {string} opts.track.name
 * @param {{ name: string }} opts.track.album
 * @param {{ name: string }[]} opts.track.artists
 */
const Playing = ({ isPlaying, track }) => {
  return html`
    <article>
      <header>${isPlaying ? 'Now playing' : 'Nothing is playing'}</header>
      <${isPlaying ? NowPlaying : NothingPlaying} track=${track} />
    </article>
  `
}

export default Playing
