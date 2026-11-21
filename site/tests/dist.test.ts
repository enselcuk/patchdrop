import { existsSync, readFileSync } from 'node:fs'
import { parse, type HTMLElement } from 'node-html-parser'
import { expect, test } from 'vitest'

export function load(path: string): HTMLElement {
  return parse(readFileSync(`dist/${path}/index.html`.replace('//', '/'), 'utf8'))
}

test('both language pages exist with the right lang', () => {
  expect(load('').querySelector('html')!.getAttribute('lang')).toBe('en')
  expect(load('tr').querySelector('html')!.getAttribute('lang')).toBe('tr')
})

test('CNAME is published', () => {
  expect(readFileSync('dist/CNAME', 'utf8').trim()).toBe('patchdrop.enginsistem.com')
})

test('hreflang + canonical on both pages', () => {
  for (const p of ['', 'tr']) {
    const h = load(p)
    expect(h.querySelector('link[hreflang="en"]')!.getAttribute('href')).toBe('https://patchdrop.enginsistem.com/')
    expect(h.querySelector('link[hreflang="tr"]')!.getAttribute('href')).toBe('https://patchdrop.enginsistem.com/tr/')
    expect(h.querySelector('link[hreflang="x-default"]')!.getAttribute('href')).toBe('https://patchdrop.enginsistem.com/')
  }
  expect(load('').querySelector('link[rel="canonical"]')!.getAttribute('href')).toBe('https://patchdrop.enginsistem.com/')
  expect(load('tr').querySelector('link[rel="canonical"]')!.getAttribute('href')).toBe('https://patchdrop.enginsistem.com/tr/')
})

test('per-language OG tags', () => {
  expect(load('tr').querySelector('meta[property="og:image"]')!.getAttribute('content')).toBe('https://patchdrop.enginsistem.com/og-tr.png')
  expect(load('').querySelector('meta[property="og:image"]')!.getAttribute('content')).toBe('https://patchdrop.enginsistem.com/og-en.png')
  expect(load('').querySelector('meta[property="og:locale"]')!.getAttribute('content')).toBe('en_US')
  expect(load('tr').querySelector('meta[property="og:locale"]')!.getAttribute('content')).toBe('tr_TR')
  expect(load('').querySelector('meta[name="twitter:card"]')!.getAttribute('content')).toBe('summary_large_image')
})

test('only / carries the redirect script', () => {
  expect(load('').querySelector('script[data-lang-redirect]')).toBeTruthy()
  expect(load('tr').querySelector('script[data-lang-redirect]')).toBeNull()
})

test('language switch links exist', () => {
  for (const p of ['', 'tr']) {
    expect(load(p).querySelector('a[data-lang-link="en"]')!.getAttribute('href')).toBe('/')
    expect(load(p).querySelector('a[data-lang-link="tr"]')!.getAttribute('href')).toBe('/tr/')
  }
})


test('server-rendered download links point at stable asset names', () => {
  const hrefs = load('').querySelectorAll('a[data-track^="download-"]').map(a => a.getAttribute('href'))
  expect(hrefs).toContain('https://github.com/enselcuk/patchdrop/releases/latest/download/PatchDrop-macOS.dmg')
  expect(hrefs).toContain('https://github.com/enselcuk/patchdrop/releases/latest/download/PatchDrop-Windows-Setup-x64.exe') })
test('made-by links exact and followed', () => {
  const a = load('tr').querySelectorAll('#made-by a')
  expect(a.map(x => x.getAttribute('href'))).toEqual(expect.arrayContaining(['https://enginsistem.com/?utm_source=patchdrop','https://www.linkedin.com/in/selcuk-engin/','https://x.com/enselcuk','https://www.youtube.com/@selcukengin','https://x.com/enginsistem','https://www.instagram.com/enginsistem/','https://www.youtube.com/@enginsistem']))
  expect(a.every(x => !(x.getAttribute('rel') ?? '').includes('nofollow'))).toBe(true) })
test('mac tips say Open Anyway, never right-click → Open', () => {
  for (const p of ['', 'tr']) { const t = load(p).querySelector('#download')!.text
    expect(t).toMatch(/Open Anyway|Yine de Aç/); expect(t).not.toMatch(/right-click|sağ tık/i) } })
test('footer copyright', () => expect(load('').querySelector('footer')!.text).toContain('© 2026 Engin Tasarım Ltd. Şti.'))
test('404 pages exist in both languages', () => { expect(existsSync('dist/404.html')).toBe(true); expect(existsSync('dist/tr/404/index.html')).toBe(true) })
test('beta bar links to GitHub issues', () => expect(load('').querySelector('[data-beta] a')!.getAttribute('href')).toBe('https://github.com/enselcuk/patchdrop/issues'))

test('hero renders four platform blocks; only "other" is visible without JS', () => {
  for (const p of ['', 'tr']) {
    const blocks = load(p).querySelectorAll('[data-hero-platform]')
    expect(blocks.map(b => b.getAttribute('data-for')).sort()).toEqual(['mac', 'mobile', 'other', 'windows'])
    expect(blocks.filter(b => !b.hasAttribute('hidden')).map(b => b.getAttribute('data-for'))).toEqual(['other'])
    const other = blocks.find(b => b.getAttribute('data-for') === 'other')!
    expect(other.querySelectorAll('a[data-track^="download-"]').length).toBe(2)
  }
})
test('live-number hooks and anchors', () => {
  for (const p of ['', 'tr']) {
    const h = load(p)
    expect(h.querySelector('[data-stars]')!.text.trim()).not.toBe('')
    expect(h.querySelector('[data-downloads]')!.hasAttribute('hidden')).toBe(true)
    expect(h.querySelector('#download')).toBeTruthy()
    expect(h.querySelector('#share')).toBeTruthy()
    const tracks = h.querySelectorAll('a[data-track]').map(a => a.getAttribute('data-track'))
    for (const t of ['download-mac-dmg', 'download-mac-zip', 'download-win-setup', 'download-win-zip', 'star-hero', 'star-share', 'send-to-computer'])
      expect(tracks).toContain(t)
  }
})
test('one h1, landmarks, images have alt and size', () => {
  for (const p of ['', 'tr']) {
    const h = load(p)
    expect(h.querySelectorAll('h1').length).toBe(1)
    for (const tag of ['header', 'main', 'footer']) expect(h.querySelector(tag)).toBeTruthy()
    for (const img of h.querySelectorAll('img')) {
      expect(img.getAttribute('alt')).toBeTruthy()
      expect(img.getAttribute('width')).toBeTruthy()
      expect(img.getAttribute('height')).toBeTruthy()
    }
  }
})
test('TR headline is exact', () => expect(load('tr').querySelector('h1')!.text.replace(/\s+/g, ' ').trim()).toBe('Kablo üzerinden, bir sürükle-bırak uzaklıkta.'))
test('FAQ uses details and covers troubleshooting, LocalSend and GoatCounter', () => {
  for (const p of ['', 'tr']) {
    const faq = load(p).querySelector('#faq')!
    expect(faq.querySelectorAll('details').length).toBeGreaterThanOrEqual(6)
    expect(faq.text).toContain('47820'); expect(faq.text).toContain('47821')
    expect(faq.text).toContain('LocalSend'); expect(faq.text).toContain('GoatCounter')
  }
})
