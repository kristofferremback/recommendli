import html from '../../lib/html.js'

const replacer = (_, value) => {
  if (value instanceof Error || value?.constructor?.name?.includes('Error')) {
    return Object.getOwnPropertyNames(value).reduce((acc, key) => {
      acc[key] = value[key]
      return acc
    }, {})
  }
  return value
}

/**
 * @param {object} params
 * @param {object} params.state
 */
const StateDump = ({ state }) => {
  return html`
    <details>
      <summary>Show entire state</summary>
      <pre>${JSON.stringify(state, replacer, 4)}</pre>
    </details>
  `
}

export default StateDump
