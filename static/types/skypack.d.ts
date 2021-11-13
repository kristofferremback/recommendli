// This file together with jsconfig.json and package.json is what actually enables
// intellisense for the modules being used.
declare module 'https://*'

declare module 'https://cdn.skypack.dev/preact@10.5.15' {
  export { default } from 'preact'
  export * from 'preact'
}

declare module 'https://cdn.skypack.dev/preact@10.5.15/hooks' {
  export { default } from 'preact/hooks'
  export * from 'preact/hooks'
}

declare module 'https://cdn.skypack.dev/htm@3.1.0' {
  export { default } from 'htm'
  export * from 'htm'
}
