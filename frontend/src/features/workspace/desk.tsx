import { createContext, useCallback, useContext, useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react'
import type { PointerEvent as ReactPointerEvent, ReactNode } from 'react'
import { clampFrame, clearLayout, frameHeight, readLayout, snapMove, snapResize, tidyLayout, writeLayout } from './lib/frames'
import type { DeskLayout, Frame, Heights, WindowName } from './lib/frames'
import { windowNames } from './lib/frames'

/*
 * The desk owns where every window sits on desktop: position, size, shade and
 * stacking order. Windows read their frame through useDesk(). On phones the
 * desk is inert and windows stack in document order.
 */

export const DESKTOP_QUERY = '(min-width: 761px)'

export function useDesktop() {
  const [desktop, setDesktop] = useState(() => window.matchMedia(DESKTOP_QUERY).matches)
  useEffect(() => {
    const media = window.matchMedia(DESKTOP_QUERY)
    const change = () => setDesktop(media.matches)
    media.addEventListener('change', change)
    return () => media.removeEventListener('change', change)
  }, [])
  return desktop
}

type DeskApi = {
  desktop: boolean
  dragging: WindowName | null
  frame: (name: WindowName) => Frame | undefined
  zIndex: (name: WindowName) => number
  register: (name: WindowName, element: HTMLElement | null) => void
  raise: (name: WindowName) => void
  startDrag: (name: WindowName, event: ReactPointerEvent<HTMLElement>) => void
  startResize: (name: WindowName, event: ReactPointerEvent<HTMLElement>) => void
  nudge: (name: WindowName, dx: number, dy: number) => void
  toggleShade: (name: WindowName) => void
  tidy: () => void
}

const DeskContext = createContext<DeskApi | null>(null)

export function useDesk() {
  const desk = useContext(DeskContext)
  if (!desk) throw new Error('useDesk needs a Desk ancestor')
  return desk
}

/** For chrome that also renders outside the desk, like the connection panel. */
export function useDeskIfAny() {
  return useContext(DeskContext)
}

const PAGE_PADDING = 12
const DESK_MAX = 1400

function initialWidth() {
  return Math.min(DESK_MAX, window.innerWidth - 2 * PAGE_PADDING)
}

export function Desk({ desktop, children }: { desktop: boolean; children: ReactNode }) {
  const deskRef = useRef<HTMLDivElement>(null)
  const [width, setWidth] = useState(initialWidth)
  const [stored] = useState(readLayout)
  const [layout, setLayout] = useState<DeskLayout>(() => stored ?? tidyLayout(initialWidth(), {}))
  const [heights, setHeights] = useState<Heights>({})
  const [dragging, setDragging] = useState<WindowName | null>(null)
  const [mounted, setMounted] = useState<WindowName[]>([])
  const elements = useRef(new Map<WindowName, HTMLElement>())
  const observer = useRef<ResizeObserver | null>(null)
  const live = useRef({ layout, heights, width, dragging })
  live.current = { layout, heights, width, dragging }

  useLayoutEffect(() => {
    const desk = deskRef.current
    if (!desk) return
    const measure = () => setWidth(desk.clientWidth)
    measure()
    const ro = new ResizeObserver(measure)
    ro.observe(desk)
    return () => ro.disconnect()
  }, [desktop])

  // An untouched desk keeps the tidy layout and follows its width and content heights.
  // The first drag, resize, shade or nudge hands control to the user; tidy hands it back.
  const [auto, setAuto] = useState(stored === null)
  useLayoutEffect(() => {
    if (auto && desktop && mounted.length > 0) setLayout(current => ({ ...tidyLayout(width, heights, mounted), order: current.order }))
  }, [auto, desktop, width, heights, mounted])

  useEffect(() => {
    if (dragging !== null) return
    if (auto) clearLayout()
    else writeLayout(layout)
  }, [layout, dragging, auto])

  // Heights remember the last unshaded measurement, so tidy can size a shaded window's slot.
  const remeasure = useCallback(() => {
    const measured = measureAll(elements.current)
    for (const name of elements.current.keys()) if (live.current.layout.frames[name].shaded) delete measured[name]
    setHeights(current => {
      const names = Object.keys(measured) as WindowName[]
      return names.every(name => current[name] === measured[name]) ? current : { ...current, ...measured }
    })
  }, [])

  const raise = useCallback((name: WindowName) => {
    setLayout(current => current.order[current.order.length - 1] === name
      ? current
      : { ...current, order: [...current.order.filter(item => item !== name), name] })
  }, [])

  // Windows present at the first commit keep their stored stacking; a window opened later comes to the front.
  // StrictMode re-attaches the same element, which is not a reopen.
  const ready = useRef(false)
  useEffect(() => { ready.current = true }, [])
  const seen = useRef(new Map<WindowName, HTMLElement>())

  const register = useCallback((name: WindowName, element: HTMLElement | null) => {
    observer.current ??= new ResizeObserver(remeasure)
    const previous = elements.current.get(name)
    if (previous && previous !== element) observer.current.unobserve(previous)
    if (element) {
      elements.current.set(name, element)
      observer.current.observe(element)
      if (ready.current && seen.current.get(name) !== element) raise(name)
      seen.current.set(name, element)
    } else {
      elements.current.delete(name)
      if (live.current.dragging === name) setDragging(null)
    }
    const names = windowNames.filter(item => elements.current.has(item))
    setMounted(current => current.length === names.length && current.every((item, i) => item === names[i]) ? current : names)
  }, [remeasure, raise])

  const effective = useCallback((name: WindowName) => clampFrame(live.current.layout.frames[name], live.current.width), [])

  const neighbours = useCallback((except: WindowName) => {
    const { layout, heights } = live.current
    return [...elements.current.keys()].filter(name => name !== except).map(name => {
      const frame = effective(name)
      return { x: frame.x, y: frame.y, w: frame.w, h: frameHeight(name, frame, heights) }
    })
  }, [effective])

  const update = useCallback((name: WindowName, change: (frame: Frame) => Frame) => {
    setLayout(current => ({ ...current, frames: { ...current.frames, [name]: change(current.frames[name]) } }))
  }, [])

  const track = useCallback((name: WindowName, event: ReactPointerEvent<HTMLElement>, onMove: (dx: number, dy: number, start: Frame) => Frame) => {
    if (!desktop || event.button !== 0) return
    const target = event.currentTarget
    const origin = { x: event.clientX, y: event.clientY, frame: effective(name) }
    event.preventDefault()
    target.setPointerCapture(event.pointerId)
    target.focus({ preventScroll: true })
    raise(name)
    setDragging(name)
    // A click that never moves keeps the desk in auto layout.
    const move = (ev: PointerEvent) => {
      const dx = ev.clientX - origin.x
      const dy = ev.clientY - origin.y
      if (dx === 0 && dy === 0) return
      setAuto(false)
      update(name, () => onMove(dx, dy, origin.frame))
    }
    const stop = () => {
      target.removeEventListener('pointermove', move)
      target.removeEventListener('pointerup', stop)
      target.removeEventListener('pointercancel', stop)
      setDragging(null)
    }
    target.addEventListener('pointermove', move)
    target.addEventListener('pointerup', stop)
    target.addEventListener('pointercancel', stop)
  }, [desktop, effective, raise, update])

  const startDrag = useCallback((name: WindowName, event: ReactPointerEvent<HTMLElement>) => {
    if ((event.target as Element).closest('button')) return
    track(name, event, (dx, dy, start) => {
      const moved = { ...start, x: start.x + dx, y: start.y + dy }
      return snapMove(moved, frameHeight(name, start, live.current.heights), live.current.width, neighbours(name))
    })
  }, [track, neighbours])

  const startResize = useCallback((name: WindowName, event: ReactPointerEvent<HTMLElement>) => {
    track(name, event, (dx, dy, start) => {
      const grown = { ...start, w: start.w + dx, h: start.h === null ? null : start.h + dy }
      return snapResize(grown, live.current.width, neighbours(name))
    })
  }, [track, neighbours])

  const nudge = useCallback((name: WindowName, dx: number, dy: number) => {
    setAuto(false)
    update(name, frame => clampFrame({ ...frame, x: frame.x + dx, y: frame.y + dy }, live.current.width))
  }, [update])

  const toggleShade = useCallback((name: WindowName) => {
    setAuto(false)
    update(name, frame => ({ ...frame, shaded: !frame.shaded }))
  }, [update])

  const tidy = useCallback(() => {
    setAuto(true)
    remeasure()
  }, [remeasure])

  const api = useMemo<DeskApi>(() => ({
    desktop,
    dragging,
    frame: name => desktop ? clampFrame(layout.frames[name], width) : undefined,
    zIndex: name => layout.order.indexOf(name) + 1,
    register, raise, startDrag, startResize, nudge, toggleShade, tidy,
  }), [desktop, dragging, layout, width, register, raise, startDrag, startResize, nudge, toggleShade, tidy])

  const extent = desktop
    ? Math.max(0, ...mounted.map(name => {
      const frame = clampFrame(layout.frames[name], width)
      return frame.y + frameHeight(name, frame, heights)
    }))
    : undefined

  return (
    <DeskContext.Provider value={api}>
      <div className="desk" ref={deskRef} style={extent ? { minHeight: extent } : undefined}>
        {children}
      </div>
    </DeskContext.Provider>
  )
}

function measureAll(elements: Map<WindowName, HTMLElement>): Heights {
  const heights: Heights = {}
  for (const [name, element] of elements) heights[name] = element.offsetHeight
  return heights
}
