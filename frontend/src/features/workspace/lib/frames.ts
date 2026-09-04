import type { PluginName } from '../plugins'

/* Window geometry for the desktop desk. Pure functions, no DOM. */

export type WindowName = 'player' | PluginName
export const windowNames: WindowName[] = ['player', 'queue', 'discovery', 'tracks', 'library']

/** h is null for windows whose height follows their content (player, discovery). */
export type Frame = { x: number; y: number; w: number; h: number | null; shaded: boolean }
export type Frames = Record<WindowName, Frame>
/** order runs back to front. */
export type DeskLayout = { frames: Frames; order: WindowName[] }
export type Heights = Partial<Record<WindowName, number>>

export const GAP = 14
export const SNAP = 10
export const MIN_W = 320
export const MIN_H = 220
export const TITLE_H = 36

const fixedHeight: Record<WindowName, boolean> = { player: true, queue: false, discovery: true, tracks: false, library: false }
const estimate: Record<WindowName, number> = { player: 456, queue: 560, discovery: 420, tracks: 440, library: 560 }

export function resizesHeight(name: WindowName) {
  return !fixedHeight[name]
}

const columns: WindowName[][] = [['player', 'library'], ['queue', 'discovery', 'tracks']]

/**
 * Two columns under 1100px, otherwise a third for the player column.
 * Open windows stack in each column; a closed window takes the slot it would reopen into.
 */
export function tidyLayout(width: number, heights: Heights, open: WindowName[] = windowNames): DeskLayout {
  const left = width < 1100 ? Math.round((width - GAP) / 2) : Math.round((width - 2 * GAP) / 3)
  const right = width - left - GAP
  const frames = {} as Frames
  columns.forEach((column, i) => {
    let y = 0
    for (const name of column) {
      const h = fixedHeight[name] ? null : estimate[name]
      frames[name] = { x: i === 0 ? 0 : left + GAP, y, w: i === 0 ? left : right, h, shaded: false }
      if (open.includes(name)) y += (h ?? heights[name] ?? estimate[name]) + GAP
    }
  })
  return { frames, order: [...windowNames] }
}

/** Rendered height of a window: title only when shaded, else the set or measured height. */
export function frameHeight(name: WindowName, frame: Frame, heights: Heights) {
  if (frame.shaded) return TITLE_H
  return frame.h ?? heights[name] ?? estimate[name]
}

/** Keep a frame inside the desk width without changing what is stored. */
export function clampFrame(frame: Frame, width: number): Frame {
  const w = Math.max(MIN_W, Math.min(frame.w, width))
  const x = Math.max(0, Math.min(frame.x, width - w))
  const y = Math.max(0, frame.y)
  if (w === frame.w && x === frame.x && y === frame.y) return frame
  return { ...frame, w, x, y }
}

export function snap(value: number, targets: number[], threshold = SNAP) {
  let best = value
  let distance = threshold + 1
  for (const target of targets) {
    const d = Math.abs(target - value)
    if (d < distance) { distance = d; best = target }
  }
  return distance <= threshold ? best : value
}

type Neighbour = { x: number; y: number; w: number; h: number }

/** Snap a moved window to the desk edges and to every neighbour's edge, with or without a gap. */
export function snapMove(frame: Frame, height: number, width: number, neighbours: Neighbour[]): Frame {
  const xs = [0, width - frame.w]
  const ys = [0]
  for (const n of neighbours) {
    xs.push(n.x, n.x + n.w - frame.w, n.x + n.w + GAP, n.x - GAP - frame.w)
    ys.push(n.y, n.y + n.h - height, n.y + n.h + GAP, n.y - GAP - height)
  }
  return clampFrame({ ...frame, x: snap(frame.x, xs), y: snap(frame.y, ys) }, width)
}

/** Snap the right and bottom edges while resizing. */
export function snapResize(frame: Frame, width: number, neighbours: Neighbour[]): Frame {
  const rights = [width]
  const bottoms: number[] = []
  for (const n of neighbours) {
    rights.push(n.x + n.w, n.x - GAP)
    bottoms.push(n.y + n.h, n.y - GAP)
  }
  const w = Math.max(MIN_W, Math.min(snap(frame.x + frame.w, rights) - frame.x, width - frame.x))
  const h = frame.h === null ? null : Math.max(MIN_H, snap(frame.y + frame.h, bottoms) - frame.y)
  return { ...frame, w, h }
}

const STORAGE_KEY = 'recommendli-desk-v1'

export function readLayout(): DeskLayout | null {
  try {
    const value = JSON.parse(localStorage.getItem(STORAGE_KEY) ?? 'null') as Partial<DeskLayout> | null
    if (!value || typeof value !== 'object' || !value.frames) return null
    const frames = {} as Frames
    for (const name of windowNames) {
      const frame = value.frames[name]
      if (!frame || ![frame.x, frame.y, frame.w].every(Number.isFinite)) return null
      const h = fixedHeight[name] || !Number.isFinite(frame.h) ? null : Math.max(MIN_H, frame.h as number)
      frames[name] = { x: frame.x, y: Math.max(0, frame.y), w: frame.w, h, shaded: !!frame.shaded }
    }
    const order = Array.isArray(value.order) ? value.order.filter((name): name is WindowName => windowNames.includes(name)) : []
    return { frames, order: [...order, ...windowNames.filter(name => !order.includes(name))] }
  } catch {
    return null
  }
}

export function writeLayout(layout: DeskLayout) {
  try { localStorage.setItem(STORAGE_KEY, JSON.stringify(layout)) } catch {}
}

/** An untouched desk stores nothing, so it lays itself out again on the next visit. */
export function clearLayout() {
  try { localStorage.removeItem(STORAGE_KEY) } catch {}
}
