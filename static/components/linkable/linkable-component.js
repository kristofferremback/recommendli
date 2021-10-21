import { html } from 'https://unpkg.com/htm/preact/standalone.module.js'

/**
 * @typedef {import('../../recommendli/client').Artist} Artist
 * @typedef {import('../../recommendli/client').Album} Album
 * @typedef {import('../../recommendli/client').Track} Track
 */

export const LinkableName = ({ url, name }) => {
  return html` <a href="${url}">${name}</a> `
}

/**
 * @param {{ artists: Artist[] }} opts
 * @returns
 */
export const LinkableArtists = ({ artists }) => {
  return html`
    ${artists.map((artist, i, arr) => {
      return html`<${SpotifyLinkable} item=${artist} /> ${i < arr.length - 1 ? ',' : ''}`
    })}
  `
}

/**
 *
 * @param {{ item: { name: string, external_urls: { spotify: string } } }} args
 */
export const SpotifyLinkable = ({ item }) => {
  const url = item.external_urls ? item.external_urls.spotify : '#'
  return html`<${LinkableName} name=${item.name} url=${url} />`
}
