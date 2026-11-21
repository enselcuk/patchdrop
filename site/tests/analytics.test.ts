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

// --- strip timing ---
const mkTimingEnv = (href: string) => {
  const replaceState = vi.fn()
  return { env: { location: { href }, history: { replaceState }, setTimeout, clearTimeout } as unknown as Window, replaceState }
}
beforeEach(() => { vi.useFakeTimers(); document.head.innerHTML = '<script data-goatcounter="x" src="//gc.zgo.at/count.js"></script>' })
afterEach(() => { vi.useRealTimers(); document.head.innerHTML = '' })

test('strips after script load plus one tick, not before', () => {
  const { env, replaceState } = mkTimingEnv('https://patchdrop.enginsistem.com/?utm_source=x&a=1#h')
  wireCampaignCleanup(document, env)
  vi.advanceTimersByTime(1000)
  expect(replaceState).not.toHaveBeenCalled()
  document.querySelector('script')!.dispatchEvent(new Event('load'))
  expect(replaceState).not.toHaveBeenCalled()
  vi.advanceTimersByTime(0)
  expect(replaceState).toHaveBeenCalledWith(null, '', 'https://patchdrop.enginsistem.com/?a=1#h')
})

test('strips at 3 s if the script never loads', () => {
  const { env, replaceState } = mkTimingEnv('https://patchdrop.enginsistem.com/?utm_source=x')
  wireCampaignCleanup(document, env)
  vi.advanceTimersByTime(2999)
  expect(replaceState).not.toHaveBeenCalled()
  vi.advanceTimersByTime(1)
  expect(replaceState).toHaveBeenCalledTimes(1)
})

test('strips only once when load comes before the timeout', () => {
  const { env, replaceState } = mkTimingEnv('https://patchdrop.enginsistem.com/?ref=x')
  wireCampaignCleanup(document, env)
  document.querySelector('script')!.dispatchEvent(new Event('load'))
  vi.advanceTimersByTime(5000)
  expect(replaceState).toHaveBeenCalledTimes(1)
})

test('does nothing when there is nothing to strip', () => {
  const { env, replaceState } = mkTimingEnv('https://patchdrop.enginsistem.com/#faq')
  wireCampaignCleanup(document, env)
  document.querySelector('script')!.dispatchEvent(new Event('load'))
  vi.advanceTimersByTime(5000)
  expect(replaceState).not.toHaveBeenCalled()
})
