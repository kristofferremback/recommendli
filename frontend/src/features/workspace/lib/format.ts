export function formatTime(milliseconds: number) {
  const seconds = Math.max(0, Math.floor(milliseconds / 1000))
  return `${Math.floor(seconds / 60).toString().padStart(2, '0')}:${(seconds % 60).toString().padStart(2, '0')}`
}

export function padIndex(index: number) {
  return String(index).padStart(2, '0')
}

export function artistNames(artists: Array<{ name: string }>) {
  return artists.map(artist => artist.name).join(', ')
}

export function formatClock(iso: string) {
  return new Date(iso).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}
