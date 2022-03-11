import { useMemo } from '../../deps/preact/hooks.js'
import html from '../../lib/html.js'

export const LoadingText = ({ children }) => {
  return html` <span aria-busy="true">${children}</span> `
}

const withConditionalLoading = (Loading) => {
  const ConditionalLoading = ({ isLoading, children }) => {
    return useMemo(() => {
      if (isLoading) {
        return html`<${Loading} />`
      }
      return html`${children}`
    }, [isLoading])
  }

  return ConditionalLoading
}

export default withConditionalLoading
