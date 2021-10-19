import { useEffect } from 'https://unpkg.com/htm/preact/standalone.module.js'

const usePolling = (action, { isActive, interval = 2000 }) => {
  return useEffect(() => {
    const timerHandles = new Set()

    const startPolling = () => {
      action()
      timerHandles.add(setInterval(() => action(), interval))
    }

    const cleanup = () => {
      for (const handle of timerHandles.values()) {
        console.log('removing handle', handle)
        timerHandles.delete(handle)
        clearInterval(handle)
      }
    }

    if (isActive) {
      cleanup()
      startPolling()
    } else {
      cleanup()
    }

    return cleanup
  }, [isActive])
}

export default usePolling
