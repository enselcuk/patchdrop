import fallback from './release-fallback.json'
import type { Lang } from '../i18n/types'

export const REPO = 'enselcuk/patchdrop'
export const ASSETS = {
  macDmg: 'PatchDrop-macOS.dmg',
  macZip: 'PatchDrop-macOS-universal.zip',
  winSetup: 'PatchDrop-Windows-Setup-x64.exe',
  winZip: 'PatchDrop-Windows-x64.zip',
  sums: 'SHA256SUMS.txt',
} as const

export function downloadUrl(name: string): string {
  return `https://github.com/${REPO}/releases/latest/download/${name}`
}

export interface ReleaseInfo {
  version: string
  publishedAt: string
  stars: number
  downloads: number
  sizes: Record<string, number>
}

const API = `https://api.github.com/repos/${REPO}`
const TIMEOUT_MS = 10_000

const isObj = (v: unknown): v is Record<string, unknown> => typeof v === 'object' && v !== null
const num = (v: unknown): number | undefined => (typeof v === 'number' && Number.isFinite(v) ? v : undefined)

export function parseLiveStats(repoJson: unknown, releasesJson: unknown): { stars?: number; downloads?: number } {
  const out: { stars?: number; downloads?: number } = {}
  const stars = isObj(repoJson) ? num(repoJson.stargazers_count) : undefined
  if (stars !== undefined) out.stars = stars
  if (Array.isArray(releasesJson)) {
    let total = 0
    for (const rel of releasesJson) {
      if (!isObj(rel) || !Array.isArray(rel.assets)) continue
      for (const a of rel.assets) if (isObj(a)) total += num(a.download_count) ?? 0
    }
    out.downloads = total
  }
  return out
}

export async function getReleaseInfo(fetchFn: typeof fetch = fetch): Promise<ReleaseInfo> {
  const ctrl = new AbortController()
  const timer = setTimeout(() => ctrl.abort(), TIMEOUT_MS)
  try {
    const headers: Record<string, string> = { Accept: 'application/vnd.github+json' }
    const token = process.env.GITHUB_TOKEN
    if (token) headers.Authorization = `Bearer ${token}`
    const get = async (path: string): Promise<unknown> => {
      const res = await fetchFn(`${API}${path}`, { headers, signal: ctrl.signal })
      if (!res.ok) throw new Error(`GitHub API ${res.status}`)
      return res.json()
    }
    const [repo, releases] = await Promise.all([get(''), get('/releases?per_page=100')])
    if (!Array.isArray(releases)) return fallback
    const latest = releases.find((r) => isObj(r) && r.draft !== true)
    if (!isObj(latest) || typeof latest.tag_name !== 'string' || !Array.isArray(latest.assets)) return fallback
    const sizes: Record<string, number> = {}
    for (const name of Object.values(ASSETS)) {
      const asset = latest.assets.find((a) => isObj(a) && a.name === name)
      const size = isObj(asset) ? num(asset.size) : undefined
      if (size === undefined) return fallback
      sizes[name] = size
    }
    const live = parseLiveStats(repo, releases)
    return {
      version: latest.tag_name.replace(/^v/, ''),
      publishedAt: typeof latest.published_at === 'string' ? latest.published_at : fallback.publishedAt,
      stars: live.stars ?? 0,
      downloads: live.downloads ?? 0,
      sizes,
    }
  } catch {
    return fallback
  } finally {
    clearTimeout(timer)
  }
}

export function formatCount(n: number, lang: Lang): string {
  return new Intl.NumberFormat(lang, { notation: 'compact', maximumFractionDigits: 1 }).format(n)
}

export function shouldShowDownloads(n: number): boolean {
  return n >= 100
}
