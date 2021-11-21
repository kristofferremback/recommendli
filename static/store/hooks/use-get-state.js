import { useRef, useCallback } from '../../deps/preact/hooks.js'

export const useGetState = (state) => {
  const lastState = useRef(state)
  return useCallback(() => lastState.current, [])
}
