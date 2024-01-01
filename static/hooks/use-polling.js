import { useEffect } from '../deps/preact/hooks.js'

const controllablePromise = () => {
  /** @type {(value: any) => void} */
  let resolve
  /** @type {(reason?: any) => void} */
  let reject
  const promise = new Promise((res, rej) => {
    resolve = res
    reject = rej
  })
  return { promise, resolve, reject }
}

const errDone = new Error('done')

const usePolling = (action, { isActive, interval = 2000 }) => {
  return useEffect(() => {
    const timeoutHandles = new Set()
    const controllers = []

    const safeAction = async () => {
      try {
        await action()
      } catch (e) {
        console.error('Polling action failed', e)
      }
    }

    const startPolling = async () => {
      const ctrl = controllablePromise()
      controllers.push(ctrl)

      let prev = Date.now()
      await safeAction()
      if (!isActive) return

      while (isActive) {
        const sleepDur = Math.max(0, interval - (Date.now() - prev))
        console.debug(`Sleeping for ${sleepDur}ms`)

        await Promise.race([
          ctrl.promise,
          new Promise(async (resolve) => timeoutHandles.add(setTimeout(resolve, sleepDur))),
        ])

        prev = Date.now()
        await safeAction()
      }
    }

    const cleanup = () => {
      while (controllers.length) {
        controllers.pop().reject(errDone)
      }

      for (const handle of timeoutHandles.values()) {
        timeoutHandles.delete(handle)
        clearTimeout(handle)
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
