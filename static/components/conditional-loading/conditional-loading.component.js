import { html } from 'https://unpkg.com/htm/preact/standalone.module.js'

export const LoadingText = ({ children }) => {
  return html` <span aria-busy="true">${children}</span> `
}

const withConditionalLoading = (Loading) => {
  const ConditionalLoading = ({ isLoading, children }) => {
    if (isLoading) {
      return html`<${Loading} />`
    }
    return html`${children}`
  }

  return ConditionalLoading
}

export default withConditionalLoading
