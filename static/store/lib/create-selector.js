/**
 * @template Out
 * @typedef {(...args) => Out} SelectFunc
 */
const memoizedCreateSelector = () => {
  /** @type {Map<any, { deps: any[], result: any }>} */
  const prevLookup = new Map()

  /**
   * @template Out
   * @param {SelectFunc<any>[]} baseSelectors
   * @param {SelectFunc<Out>} selectFunc
   */
  const createSelector = (baseSelectors, selectFunc) => {
    let prev = prevLookup.get(selectFunc)

    /**
     * @param {any[]} deps
     * @returns {boolean}
     */
    const matchesPrev = (deps) => {
      return prev != null && prev.deps.length === deps.length && prev.deps.every((v, i) => deps[i] === v)
    }

    /**
     * @param {any} state
     * @returns {Out}
     */
    const selector = (state) => {
      const deps = baseSelectors.map((select) => select(state))
      if (prev == null || !matchesPrev(deps)) {
        prev = { deps, result: selectFunc(...deps) }
      }
      return prev.result
    }
    return selector
  }
  return createSelector
}

const createSelector = memoizedCreateSelector()

export default createSelector
