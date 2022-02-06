const loggerMiddleware = ({ getState }) => {
  return (next) => {
    const prevState = getState()
    return (action) => {
      next(action)

      const nextState = getState()
      console.group(`Action: ${action.type}`)
      console.debug('Payload', action.payload)
      console.debug('Prev state:', prevState)
      console.debug('Next state:', nextState)
      console.groupEnd()
    }
  }
}

export default loggerMiddleware
