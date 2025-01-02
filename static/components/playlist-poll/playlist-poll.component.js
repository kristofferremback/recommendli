import html from '../../lib/html.js'
import { useState, useEffect } from '../../deps/preact/hooks.js'
import { LinkableName } from '../linkable/linkable-component.js'

/**
 * @param {object} opts
 * @param {import('../../recommendli/client.js').Playlist} opts.playlist
 * @param {string} opts.playlistId
 * @param {function(string): void} opts.setPlaylistId
 * @returns
 */
const PlaylistPoll = ({ playlist, playlistId, setPlaylistId }) => {
  const [snapshotIds, setSnapshotIds] = useState([])

  useEffect(() => {
    if (playlist?.snapshot_id) {
      setSnapshotIds((ids) => [...ids, { id: playlist.snapshot_id, at: new Date() }])
    }
  }, [playlist?.snapshot_id])

  return html`
    <article>
      <header>${playlist?.name || 'Input a playlist ID'} | ${playlistId}</header>
      <div>
        <input
          type="text"
          value=${playlistId}
          onInput=${(e) => {
            setPlaylistId(e.target.value)
          }}
        />
      </div>
      <div>
        <h2>Playlist</h2>
        <p><${LinkableName} item=${playlist} /></p>
        <p>Snapshot ID: ${playlist?.snapshot_id || ''}</p>
      </div>
      <div>
        <h2>Snapshot IDs</h2>
        <ul>
          ${snapshotIds.map(
            ({ id, at }) => html`
              <li>
                <div>${id}</div>
                <div>${at.toISOString()}</div>
              </li>
            `
          )}
        </ul>
    </article>
  `
}

export default PlaylistPoll
