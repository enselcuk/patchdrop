// @vitest-environment jsdom
import { afterEach, beforeEach, expect, test, vi } from 'vitest'
import { stripCampaignParams, trackEvent, wireCampaignCleanup } from '../src/lib/analytics'
import { enhance } from '../src/scripts/enhance'

test('removes utm_* and keeps other params and the hash', () => expect(stripCampaignParams('https://patchdrop.enginsistem.com/tr/?utm_source=linkedin&utm_medium=social&x=1#download')).toBe('https://patchdrop.enginsistem.com/tr/?x=1#download'))
test('removes ref too', () => expect(stripCampaignParams('https://patchdrop.enginsistem.com/?ref=x')).toBe('https://patchdrop.enginsistem.com/'))
test('nothing to strip → null', () => expect(stripCampaignParams('https://patchdrop.enginsistem.com/#faq')).toBeNull())
test('only unrelated params → null', () => expect(stripCampaignParams('https://patchdrop.enginsistem.com/?x=1')).toBeNull())
test('trackEvent without goatcounter does not throw', () => expect(() => trackEvent('download-mac-dmg', undefined)).not.toThrow())
test('trackEvent with no count function does not throw', () => expect(() => trackEvent('x', {} as never)).not.toThrow())
test('trackEvent swallows errors from count', () => expect(() => trackEvent('x', { count: () => { throw new Error('boom') } })).not.toThrow())
test('trackEvent calls count with event:true', () => { const count = vi.fn(); trackEvent('star-hero', { count }); expect(count).toHaveBeenCalledWith({ path: 'star-hero', title: 'star-hero', event: true }) })

// --- click delegation through enhance ---
const failing = () => Promise.reject(new Error('offline'))
const mkEnv = (gc: unknown) =>
  ({ navigator: { userAgent: 'x', platform: '', maxTouchPoints: 0 }, fetch: failing, AbortController, setTimeout, clearTimeout, location: { hash: '' }, goatcounter: gc }) as unknown as Window

test('click on [data-track] (and on a child) counts the event', async () => {
  document.body.innerHTML = `<a href="#x" data-track="download-mac-dmg"><span id="inner">Get</span></a><button data-track="star-hero">Star</button><a href="#y" id="plain">no</a>`
  const count = vi.fn()
  await enhance(document, mkEnv({ count }))
  document.querySelector<HTMLElement>('#inner')!.click()
  expect(count).toHaveBeenLastCalledWith({ path: 'download-mac-dmg', title: 'download-mac-dmg', event: true })
  document.querySelector<HTMLElement>('button')!.click()
  expect(count).toHaveBeenLastCalledWith({ path: 'star-hero', title: 'star-hero', event: true })
  const n = count.mock.calls.length
  document.querySelector<HTMLElement>('#plain')!.click()
  expect(count).toHaveBeenCalledTimes(n)
})

test('click is safe when goatcounter is blocked', async () => {
  document.body.innerHTML = `<button data-track="share-x">x</button>`
  await enhance(document, mkEnv(undefined))
  expect(() => document.querySelector<HTMLElement>('button')!.click()).not.toThrow()
})

// --- page view + strip timing ---
const URL_UTM = 'https://patchdrop.enginsistem.com/?utm_source=x&a=1#h'
const mk = (href: string, gc?: unknown, readyState: DocumentReadyState = 'complete') => {
  const replaceState = vi.fn()
  const listeners: Record<string, () => void> = {}
  const env = { location: { href }, history: { replaceState }, setTimeout, clearTimeout, goatcounter: gc,
    addEventListener: (n: string, f: () => void) => { listeners[n] = f } } as unknown as Window
  Object.defineProperty(document, 'readyState', { value: readyState, configurable: true })
  return { env, replaceState, listeners }
}
const script = () => document.querySelector('script')!
beforeEach(() => { vi.useFakeTimers(); document.head.innerHTML = '<script data-goatcounter="x" src="//gc.zgo.at/count.js"></script>' })
afterEach(() => { vi.useRealTimers(); document.head.innerHTML = ''; delete (document as { readyState?: unknown }).readyState })

test('script still loading at 3 s from enhance (window not loaded yet): params survive', () => {
  const { env, replaceState } = mk(URL_UTM, { no_onload: true }, 'loading')
  wireCampaignCleanup(document, env)
  vi.advanceTimersByTime(3000)
  expect(replaceState).not.toHaveBeenCalled()
})

test('load: count() once with no args, then strip after a tick', () => {
  const gc: { no_onload: boolean; count?: ReturnType<typeof vi.fn> } = { no_onload: true }
  const { env, replaceState } = mk(URL_UTM, gc)
  wireCampaignCleanup(document, env)
  gc.count = vi.fn()
  script().dispatchEvent(new Event('load'))
  expect(gc.count).toHaveBeenCalledTimes(1)
  expect(gc.count).toHaveBeenCalledWith()
  expect(replaceState).not.toHaveBeenCalled()
  vi.advanceTimersByTime(0)
  expect(replaceState).toHaveBeenCalledWith(null, '', 'https://patchdrop.enginsistem.com/?a=1#h')
  vi.advanceTimersByTime(10000)
  expect(gc.count).toHaveBeenCalledTimes(1)
  expect(replaceState).toHaveBeenCalledTimes(1)
})

test('already loaded before enhance: count() once, then strip', () => {
  const count = vi.fn()
  const { env, replaceState } = mk(URL_UTM, { count })
  wireCampaignCleanup(document, env)
  expect(count).toHaveBeenCalledTimes(1)
  expect(replaceState).not.toHaveBeenCalled()
  vi.advanceTimersByTime(0)
  expect(replaceState).toHaveBeenCalledTimes(1)
  script().dispatchEvent(new Event('load'))
  vi.advanceTimersByTime(5000)
  expect(count).toHaveBeenCalledTimes(1)
})

test('error (blocked): strip immediately, nothing counted', () => {
  const { env, replaceState } = mk(URL_UTM, { no_onload: true })
  wireCampaignCleanup(document, env)
  script().dispatchEvent(new Event('error'))
  expect(replaceState).toHaveBeenCalledTimes(1)
  vi.advanceTimersByTime(5000)
  expect(replaceState).toHaveBeenCalledTimes(1)
})

test('never loads: strip 3 s after window load, not before', () => {
  const { env, replaceState, listeners } = mk(URL_UTM, { no_onload: true }, 'loading')
  wireCampaignCleanup(document, env)
  vi.advanceTimersByTime(10000)
  expect(replaceState).not.toHaveBeenCalled()
  listeners.load()
  vi.advanceTimersByTime(2999)
  expect(replaceState).not.toHaveBeenCalled()
  vi.advanceTimersByTime(1)
  expect(replaceState).toHaveBeenCalledTimes(1)
})

test('late load after the 3 s strip still counts the view once', () => {
  const gc: { count?: ReturnType<typeof vi.fn> } = {}
  const { env } = mk(URL_UTM, gc)
  wireCampaignCleanup(document, env)
  vi.advanceTimersByTime(3000)
  gc.count = vi.fn()
  script().dispatchEvent(new Event('load'))
  expect(gc.count).toHaveBeenCalledTimes(1)
})

test('nothing to strip: view still counted, no replaceState', () => {
  const count = vi.fn()
  const { env, replaceState } = mk('https://patchdrop.enginsistem.com/#faq', { count })
  wireCampaignCleanup(document, env)
  vi.advanceTimersByTime(5000)
  expect(count).toHaveBeenCalledTimes(1)
  expect(replaceState).not.toHaveBeenCalled()
})
