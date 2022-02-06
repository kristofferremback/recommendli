import { useEffect } from '../deps/preact/hooks.js'

const useEventListener = (eventName, onEvent) => {
  return useEffect(() => {
    const eventListener = (e) => onEvent(e)
    document.addEventListener(eventName, eventListener)
    return () => document.removeEventListener(eventName, eventListener)
  }, [])
}

export default useEventListener
