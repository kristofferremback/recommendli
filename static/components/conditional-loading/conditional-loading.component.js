import html from '../../lib/html.js'

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
