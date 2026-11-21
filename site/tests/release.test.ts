import { expect, test } from 'vitest'
import { ASSETS, downloadUrl, formatCount, getReleaseInfo, parseLiveStats, shouldShowDownloads } from '../src/lib/release'
import fallback from '../src/lib/release-fallback.json'

const json = (b: unknown) => new Response(JSON.stringify(b), { status: 200 })

test('downloadUrl uses latest/download', () =>
  expect(downloadUrl(ASSETS.winSetup)).toBe('https://github.com/enselcuk/patchdrop/releases/latest/download/PatchDrop-Windows-Setup-x64.exe'))
test('API failure → fallback json, no throw', async () =>
  expect(await getReleaseInfo(async () => { throw new Error('offline') })).toEqual(fallback))
test('403 rate limit → fallback', async () =>
  expect(await getReleaseInfo(async () => new Response('', { status: 403 }))).toEqual(fallback))
test('live success maps version, stars, downloads, sizes', async () => {
  const assets = Object.values(ASSETS).map((name, i) => ({ name, size: 100 + i, download_count: 2 }))
  const f = (async (url: string) =>
    String(url).endsWith('/patchdrop')
      ? json({ stargazers_count: 9 })
      : json([{ draft: true, tag_name: 'v9.9.9', assets: [] }, { tag_name: 'v0.2.0', published_at: '2026-12-01T00:00:00Z', assets }])) as unknown as typeof fetch
  const r = await getReleaseInfo(f)
  expect(r.version).toBe('0.2.0')
  expect(r.publishedAt).toBe('2026-12-01T00:00:00Z')
  expect(r.stars).toBe(9)
  expect(r.downloads).toBe(10)
  expect(r.sizes[ASSETS.macDmg]).toBe(100)
})
test('missing asset → fallback', async () => {
  const f = (async (url: string) =>
    String(url).endsWith('/patchdrop')
      ? json({ stargazers_count: 9 })
      : json([{ tag_name: 'v0.2.0', published_at: 'x', assets: [{ name: ASSETS.macDmg, size: 1 }] }])) as unknown as typeof fetch
  expect(await getReleaseInfo(f)).toEqual(fallback)
})
test('downloads = sum of download_count over all releases and assets', () =>
  expect(parseLiveStats({ stargazers_count: 7 }, [{ assets: [{ download_count: 3 }, { download_count: 4 }] }, { assets: [{ download_count: 5 }] }])).toEqual({ stars: 7, downloads: 12 }))
test('garbage JSON → empty object, never NaN', () => {
  expect(parseLiveStats({ message: 'API rate limit exceeded' }, null)).toEqual({})
  expect(parseLiveStats(null, [null, 1, { assets: 'x' }, { assets: [{ download_count: 'a' }, null] }])).toEqual({ downloads: 0 })
  expect(parseLiveStats({ stargazers_count: NaN }, undefined)).toEqual({})
})
test('threshold', () => { expect(shouldShowDownloads(99)).toBe(false); expect(shouldShowDownloads(100)).toBe(true) })
test('formatCount', () => {
  expect(formatCount(1234, 'en')).toBe('1.2K')
  expect(formatCount(1234, 'tr')).toBe('1,2\u00a0B')
  expect(formatCount(12, 'tr')).toBe('12')
})
