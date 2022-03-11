import { useMemo, useEffect, useState } from '../../../deps/preact/hooks.js'
import { usePrevious } from '../../../hooks/use-previous.js'

/**
 * @template T
 * @param {[string, T][]} pairs
 * @returns {{ [key: string]: T }}
 */
const toObject = (pairs) => {
  return pairs.reduce((obj, [key, value]) => ({ ...obj, [key]: value }), {})
}

/**
 * @param {string[]} prevKeys
 * @param {string[]} newKeys
 * @returns {string[]}
 */
const orderByPrev = (prevKeys, newKeys) => {
  const ordered = Array.from({ length: newKeys.length })
  let unusedIndexes = ordered.map((_, i) => i)
  const unusedKeys = []

  for (const key of newKeys) {
    const i = prevKeys.indexOf(key)
    if (i === -1) {
      unusedKeys.push(key)
    } else {
      unusedIndexes = unusedIndexes.filter((ki) => ki !== i)
      ordered[i] = key
    }
  }

  for (let i = 0; i < unusedKeys.length; i++) {
    const key = unusedKeys[i]
    const index = unusedIndexes[i]
    ordered[index] = key
  }

  console.log('from', JSON.stringify(prevKeys))
  console.log('to  ', JSON.stringify(ordered))
  return ordered
}

/**
 * @template T
 * @param {{ [key: string]: T }} values
 * @returns {[[string, T][], typeof toObject]}
 */
export const useOrderedPairList = (values) => {
  const sortedKeys = Object.keys(values).sort()
  const sortedVals = sortedKeys.map((k) => `${values[k]}`)
  const keys = useMemo(() => Object.keys(values), [sortedKeys.join(',')])
  const vals = useMemo(() => keys.map((k) => values[k]), [keys, sortedVals.join(',')])

  const entries = useMemo(() => keys.map((k, i) => [k, vals[i]]), [keys, vals])

  const [pairs, setPairs] = useState(entries)

  useEffect(() => {
    const orderedKeys = orderByPrev(
      // @ts-ignore
      pairs.map(([k]) => k),
      entries.map(([k]) => k)
    )

    setPairs(orderedKeys.map((key) => [key, values[key]]))
  }, [entries])

  // @ts-ignore
  return [pairs, toObject]
}
