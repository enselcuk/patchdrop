export function bytes(n: number): string {
  if (!n) return '0 B'
  const u = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.min(Math.floor(Math.log(n) / Math.log(1024)), u.length - 1)
  const v = n / Math.pow(1024, i)
  return `${v >= 100 || i === 0 ? v.toFixed(0) : v.toFixed(1)} ${u[i]}`
}

export function osLabel(os: string): string {
  return ({ darwin: 'macOS', windows: 'Windows', linux: 'Linux' } as Record<string, string>)[os] ?? os ?? ''
}

export function baseName(p: string): string {
  const parts = p.split(/[\\/]/).filter(Boolean)
  return parts[parts.length - 1] ?? p
}

export function eta(remaining: number, speed: number): string {
  if (!speed || speed < 1) return ''
  const s = Math.round(remaining / speed)
  if (s < 60) return `${s}s left`
  if (s < 3600) return `${Math.floor(s / 60)}m ${s % 60}s left`
  return `${Math.floor(s / 3600)}h ${Math.floor((s % 3600) / 60)}m left`
}
