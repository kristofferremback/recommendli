import html from '../../lib/html.js'
import { useCallback, useMemo, useEffect } from '../../deps/preact/hooks.js'
import { useOrderedPairList } from './hooks/use-ordered-pairs.js'

/**
 * @template T
 * @typedef {(value: string) => T} ParseFunc
 */

/**
 * @typedef {"button" | "checkbox" | "color" | "date" | "datetime-local" | "email" | "file" | "hidden" | "image" | "month" | "number" | "password" | "radio" | "range" | "reset" | "search" | "submit" | "tel" | "text" | "time" | "url" | "week"} InputType
 * @typedef {Event & { target: { value: string } }} EventWithTarget
 */

/**
 * @template T
 * @typedef {object} InputProps
 * @property {string} id
 * @property {T} value
 * @property {InputType} type
 * @property {ParseFunc<T>} [parseValue]
 * @property {(value: T, e: EventWithTarget) => void} [onChange]
 */

/**
 * @template T
 * @param {T} v
 * @returns {T}
 */
const passThrough = (v) => v

const noop = () => void 0

/**
 * @template T
 * @param {T} func
 * @returns {T | typeof passThrough}
 */
const useParseFunc = (func) => {
  return useMemo(() => {
    if (typeof func === 'function') return func
    return passThrough
  }, [func])
}

/**
 * @template T
 * @param {T} func
 * @returns {T | typeof noop}
 */
const useOnChange = (func) => {
  return useMemo(() => {
    if (typeof func === 'function') return func
    return noop
  }, [func])
}

/**
 * @template T
 * @param {InputProps<T>} opts
 */
export const Input = ({ id, value, parseValue, onChange, type = 'text', ...props }) => {
  const parseFunc = useParseFunc(parseValue)
  const onChangeFunc = useOnChange(onChange)
  const onInput = useCallback(
    (e) => {
      onChangeFunc(parseFunc(e.target.value), e)
    },
    [parseFunc, onChangeFunc]
  )

  // console.log(new Date().toISOString(), `rerendering input`, JSON.stringify({ id, value, type }))

  return html`<input id=${id} value=${value} type=${type} onInput=${onInput} ...${props} /> `
}

/**
 * @template T
 * @param {{ name: string } & InputProps<T>} opts
 */
export const InputPair = ({ name, id, value, parseValue, onChange, type = 'text' }) => html`
  <label for=${id}>${name}</label>
  <${Input} id=${id} value=${value} type=${type} parseValue=${parseValue} onChange=${onChange} />
`

/**
 * @template T
 *
 * @param {object} opts
 * @param {string} opts.name
 * @param {string} opts.baseId
 * @param {T[]} opts.values
 * @param {ParseFunc<T>} [opts.parseValue]
 * @param {(values: T[], index: number, event: EventWithTarget) => void} [opts.onChange]
 * @param {InputType} [opts.type]
 */
export const ListInput = ({ name, baseId, values, parseValue, onChange, type = 'text' }) => {
  const onChangeFunc = useOnChange(onChange)
  const onValueChange = useCallback(
    (values, value, event, index) => {
      const changedValues = values.map((v, i) => (index === i ? value : v))
      onChangeFunc(changedValues, index, event)
    },
    [onChangeFunc]
  )

  return html`
    <fieldset>
      <legend>${name}</legend>
      ${values.map(
        (value, i) => html`
          <${Input}
            id=${`${baseId}-${i}`}
            value=${value}
            type=${type}
            parseValue=${parseValue}
            onChange=${(v, e) => onValueChange(values, v, e, i)}
          />
        `
      )}
    </fieldset>
  `
}

/**
 * @template T
 *
 * @param {object} opts
 * @param {string} opts.name
 * @param {string} opts.baseId
 * @param {{ [key: string]: T }} opts.values
 * @param {ParseFunc<T>} [opts.parseValue]
 * @param {(values: { [key: string]: T }, key: string, value: T, event: EventWithTarget) => void} [opts.onChange]
 * @param {InputType} [opts.type]
 * @param {boolean} [opts.insertEmpty]
 */
export const KeyValueInput = ({
  name,
  baseId,
  values,
  parseValue,
  onChange,
  type = 'text',
  insertEmpty = false,
}) => {
  const onChangeFunc = useOnChange(onChange)
  // TODO: Figure out why we keep re-rendering the inputs in this component
  // (which loses cmd+z history...). The ListInput component doesn't, so something's amiss...
  const [pairs, toObject] = useOrderedPairList(values)

  const renderPairs = useMemo(() => {
    if (insertEmpty) return pairs.concat([['', undefined]])
    return pairs
  }, [pairs, insertEmpty])

  const onValueChange = useCallback(
    (pairs, value, index, event) => {
      const key = pairs[index][0]
      const updated = [key, value]
      const changed = toObject(pairs.map((pair, i) => (index === i ? updated : pair)))

      onChangeFunc(changed, key, value, event)
    },
    [onChangeFunc]
  )

  const onKeyChanged = useCallback(
    (pairs, key, index, event) => {
      const value = index < pairs.length ? pairs[index][1] : undefined
      const updated = [key, value]
      const changed = toObject(
        pairs
          .map((pair, i) => (index === i ? updated : pair))
          .filter(([k, v]) => {
            if (!insertEmpty) return true
            const noKey = k == null || k === ''
            const noValue = v == null || v === ''
            return !(noKey && noValue)
          })
      )

      onChangeFunc(changed, key, value, event)
    },
    [onChangeFunc]
  )

  return html`
    <fieldset>
      <legend>${name}</legend>
      ${renderPairs.map(([key, value], i) => {
        return html`<div class="grid">
          <${Input}
            id=${`${baseId}-key-${i}`}
            value=${key}
            onChange=${(changedKey, e) => onKeyChanged(pairs, changedKey, i, e)}
          />
          <${Input}
            id=${`${baseId}-value-${i}`}
            value=${value}
            parseValue=${parseValue}
            onChange=${(changedValue, e) => onValueChange(pairs, changedValue, i, e)}
            type=${type}
          />
        </div>`
      })}
    </fieldset>
  `
}
