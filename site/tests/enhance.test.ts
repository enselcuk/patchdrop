// @vitest-environment jsdom
import { beforeEach, expect, test } from 'vitest'
import { enhance } from '../src/scripts/enhance'

const HTML = `
<div data-hero-platform data-for="mac" hidden></div>
<div data-hero-platform data-for="windows" hidden></div>
<div data-hero-platform data-for="other"></div>
<div data-hero-platform data-for="mobile" hidden><a href="#share" data-track="send-to-computer">Send</a></div>
<span data-stars>0</span>
<p data-downloads hidden><span data-downloads-count>0</span> downloads</p>
<section id="share" data-share-title="PatchDrop" data-share-text="txt">
  <input data-share-url readonly value="https://patchdrop.enginsistem.com/">
  <button data-track="share-copy" data-copied="Link copied">Copy</button>
  <button data-track="share-native" hidden>Share</button>
  <span data-share-status aria-live="polite"></span>
</section>`
let doc: Document
beforeEach(() => { document.body.innerHTML = HTML; doc = document })
const visible = (d: Document) => [...d.querySelectorAll<HTMLElement>('[data-hero-platform]')].filter((e) => !e.hidden).map((e) => e.dataset.for)

const env = (ua: string, fetchFn: unknown, extra: Record<string, unknown> = {}, touch = 0) =>
  ({ navigator: { userAgent: ua, platform: '', maxTouchPoints: touch, ...extra }, fetch: fetchFn, AbortController, setTimeout, clearTimeout, location: { hash: '' } }) as unknown as Window
const failing = () => Promise.reject(new Error('offline'))
const MAC = 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15'
const WIN = 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)'
const LINUX = 'Mozilla/5.0 (X11; Linux x86_64)'

test('shows only the mac hero block on mac', async () => { await enhance(doc, env(MAC, failing)); expect(visible(doc)).toEqual(['mac']) })
test('windows', async () => { await enhance(doc, env(WIN, failing)); expect(visible(doc)).toEqual(['windows']) })
test('linux keeps other', async () => { await enhance(doc, env(LINUX, failing)); expect(visible(doc)).toEqual(['other']) })
test('iPad shows mobile block', async () => { await enhance(doc, env(MAC, failing, {}, 5)); expect(visible(doc)).toEqual(['mobile']) })
test('keeps baked numbers when fetch fails', async () => {
  await enhance(doc, env(MAC, failing))
  expect(doc.querySelector('[data-stars]')!.textContent).toBe('0')
  expect(doc.querySelector<HTMLElement>('[data-downloads]')!.hidden).toBe(true)
})
const okFetch = (stars: unknown, dl: number) => (url: string) =>
  Promise.resolve({ ok: true, json: async () => (String(url).includes('releases') ? [{ assets: [{ download_count: dl }] }] : { stargazers_count: stars }) })
test('live numbers update stars and unhide downloads', async () => {
  await enhance(doc, env(MAC, okFetch(1234, 500)))
  expect(doc.querySelector('[data-stars]')!.textContent).toBe('1.2K')
  expect(doc.querySelector('[data-downloads-count]')!.textContent).toBe('500')
  expect(doc.querySelector<HTMLElement>('[data-downloads]')!.hidden).toBe(false)
})
test('low downloads stay hidden', async () => {
  await enhance(doc, env(MAC, okFetch(7, 5)))
  expect(doc.querySelector('[data-stars]')!.textContent).toBe('7')
  expect(doc.querySelector<HTMLElement>('[data-downloads]')!.hidden).toBe(true)
})
test('non-finite stars are ignored', async () => {
  await enhance(doc, env(MAC, okFetch('x', 500)))
  expect(doc.querySelector('[data-stars]')!.textContent).toBe('0')
})
test('copy button announces via live region (clipboard unavailable selects)', async () => {
  await enhance(doc, env(MAC, failing))
  doc.querySelector<HTMLElement>('[data-track="share-copy"]')!.click()
  await new Promise((r) => setTimeout(r, 0))
  expect(doc.querySelector('[data-share-status]')!.textContent).toBe('')
})
test('copy success announces', async () => {
  await enhance(doc, env(MAC, failing, { clipboard: { writeText: async () => {} } }))
  doc.querySelector<HTMLElement>('[data-track="share-copy"]')!.click()
  await new Promise((r) => setTimeout(r, 0))
  expect(doc.querySelector('[data-share-status]')!.textContent).toBe('Link copied')
})
test('native share button shown only when navigator.share exists', async () => {
  await enhance(doc, env(MAC, failing))
  expect(doc.querySelector<HTMLElement>('[data-track="share-native"]')!.hidden).toBe(true)
  let arg: unknown
  await enhance(doc, env(MAC, failing, { share: async (a: unknown) => { arg = a } }))
  const b = doc.querySelector<HTMLElement>('[data-track="share-native"]')!
  expect(b.hidden).toBe(false)
  b.click()
  expect(arg).toEqual({ title: 'PatchDrop', text: 'txt', url: 'https://patchdrop.enginsistem.com/' })
})
test('send-to-computer uses navigator.share and prevents default', async () => {
  let arg: unknown
  await enhance(doc, env(MAC, failing, { share: async (a: unknown) => { arg = a } }, 5))
  const ev = new MouseEvent('click', { bubbles: true, cancelable: true })
  doc.querySelector('[data-track="send-to-computer"]')!.dispatchEvent(ev)
  expect(ev.defaultPrevented).toBe(true)
  expect(arg).toBeTruthy()
})
