import html from '../../lib/html.js'
import { useMemo, useCallback, useState } from '../../deps/preact/hooks.js'

import { SpotifyLinkable } from '../linkable/linkable-component.js'
import { LoadingText } from '../conditional-loading/conditional-loading.component.js'

/**
 * @typedef {import('../../recommendli/client').IndexSummary} IdxSummary
 * @typedef {import('../../recommendli/client').SimplePlaylist} SimplePlaylist
 */

// TODO: Finish this component

/**
 *
 * @param {{ playlists: SimplePlaylist[] } } opts
 */
const PlaylistOverview = ({ playlists }) => {
  const [isOpen, setIsOpen] = useState(false)
  const toggleIsOpen = useCallback(() => setIsOpen((o) => !o), [setIsOpen])
  const sortedPlaylists = useMemo(
    () => Array.from(playlists).sort((a, b) => b.name.localeCompare(a.name, 'en-US', { numeric: true })),
    [playlists]
  )

  return html`
    <details open=${isOpen} onclick=${toggleIsOpen}>
      <summary>
        The index contains in total ${' '}
        <strong>${sortedPlaylists.length} ${sortedPlaylists.length === 1 ? 'playlist' : 'playlists'}</strong>
      </summary>
      <ul>
        ${sortedPlaylists.map((playlist) => html`<li><${SpotifyLinkable} item=${playlist} /></li>`)}
      </ul>
    </details>
  `
}

const IndexSummaryHeader = ({ isLoading, title }) => {
  return html`
    <header>
      ${(() => {
        if (title) {
          return html`${title}`
        }
        const defaultTitle = 'Index Summary'
        if (isLoading) {
          return html`<${LoadingText}>${defaultTitle}</${LoadingText}>`
        }
        return html`${defaultTitle}`
      })()}
    </header>
  `
}

/**
 * @param {object} opts
 * @param {number} opts.uniqueTrackCount
 * @param {number} opts.playlistCount
 * @param {SimplePlaylist[]} opts.playlists
 */
const Summary = ({ uniqueTrackCount, playlistCount, playlists }) => {
  return html`
    <div>
      <div><small>Tracks:</small> <strong>${uniqueTrackCount}</strong></div>
      <div><small>Playlists:</small> <strong>${playlistCount}</strong></div>
      <br />
      <div>
        <${PlaylistOverview} playlists=${playlists} />
      </div>
    </div>
  `
}

const NoSummaryYet = () => {
  return html`
    <div>
      <div>${'\u00a0'}</div>
      <div>${'\u00a0'}</div>
    </div>
  `
}

/**
 * @param {object} opts
 * @param {string} opts.title
 * @param {boolean} opts.isLoading
 * @param {IdxSummary} opts.indexSummary
 */
const IndexSummary = ({ title, isLoading, indexSummary }) => {
  return html`
    <article>
      <${IndexSummaryHeader} title=${title} isLoading=${isLoading} />
      <${indexSummary == null ? NoSummaryYet : Summary}
        uniqueTrackCount=${indexSummary?.unique_track_count ?? 0}
        playlistCount=${indexSummary?.playlist_count ?? 0}
        playlists=${indexSummary?.playlists ?? []}
      />
    </article>
  `
}

export default IndexSummary
