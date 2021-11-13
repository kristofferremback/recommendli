import { useEffect, useRef } from '../deps/preact/hooks.js'

export const usePrevious = (value) => {
  const ref = useRef(value)
  useEffect(() => {
    ref.current = value
  }, [value])
  return ref.current
}

export const useLastNonNullish = (value) => {
  const ref = useRef(value)
  useEffect(() => {
    if (value != null) {
      ref.current = value
    }
  }, [value])
  return ref.current
}
