import html from '../../lib/html.js'
import { useMemo, useCallback, useState } from '../../deps/preact/hooks.js'
import { InputPair, KeyValueInput, ListInput } from '../input/input.component.js'

/**
 * @typedef {import('../../recommendli/client').UserPrefs} UserPrefs
 */

/**
 * @param {object} a
 * @param {object} b
 */
const shallowCompareObjects = (a, b) => {
  if (Object.keys(a).length !== Object.keys(b).length) {
    return false
  }
  return Object.entries(a).every(([k, v]) => b[k] === v)
}

/**
 * @param {any[]} a
 * @param {any[]} b
 */
const shallowCompareLists = (a, b) => {
  if (a.length !== b.length) {
    return false
  }
  return a.every((v, i) => b[i] === v)
}

/**
 * @param {UserPrefs} prev
 * @param {UserPrefs} prefs
 */
const prefsMatch = (prev, prefs) => {
  return [
    prefs.library_pattern === prev.library_pattern,
    shallowCompareLists(prefs.discovery_playlist_names, prev.discovery_playlist_names),
    shallowCompareObjects(prefs.weighted_words, prev.weighted_words),
    prefs.minimum_album_size === prev.minimum_album_size,
  ].every((v) => v)
}

const parseIntValue = (value) => (!isNaN(value) ? parseInt(value, 10) : 0)

/**
 * @param {object} opts
 * @param {UserPrefs} opts.userPreferences
 */
const UserPreferences = ({ userPreferences }) => {
  const [prefs, setPrefs] = useState(userPreferences)

  /**
   * @typedef {Partial<UserPrefs>} partalUserPrefs
   * @typedef {(prefs: UserPrefs) => partalUserPrefs} setPartialUserPrefs
   *
   * @type {(value: (partalUserPrefs | setPartialUserPrefs )) => void}
   */
  const changePrefs = useCallback(
    (value) => {
      if (typeof value === 'function') {
        return setPrefs({ ...prefs, ...value(prefs) })
      }
      return setPrefs({ ...prefs, ...value })
    },
    [prefs, setPrefs]
  )

  const isChanged = useMemo(() => !prefsMatch(userPreferences, prefs), [prefs, userPreferences])

  return html`
    <article>
      <header>User preferences</header>
      <form>
        <div class="grid">
          <${InputPair}
            name="Library pattern"
            id="prefs-library_pattern"
            value=${prefs.library_pattern}
            onChange=${(value) => changePrefs({ library_pattern: value })}
          />
        </div>
        <div></div>
        <div>
          <${ListInput}
            name="Discovery playlist names"
            baseId="prefs-discovery_playlist_names"
            values=${prefs.discovery_playlist_names.concat('')}
            onChange=${(values) => {
              changePrefs({ discovery_playlist_names: values.filter((v) => !!v) })
            }}
          />
        </div>
        <div>
          <${KeyValueInput}
            name="Weighted words"
            baseId="prefs-weighted_words"
            values=${prefs.weighted_words}
            parseValue=${parseIntValue}
            onChange=${(values) => changePrefs({ weighted_words: values })}
            type="number"
            insertEmpty=${true}
          />
        </div>
        <div class="grid">
          <${InputPair}
            name="Minimum album size"
            id="prefs-minimum_album_size"
            value=${prefs.minimum_album_size}
            type="number"
            parseValue=${parseIntValue}
            onChange=${(value) => changePrefs({ minimum_album_size: value })}
          />
        </div>
      </form>
      <h1>${isChanged ? 'Changed' : 'Not changed'}</h1>
    </article>
  `
}

export default UserPreferences
