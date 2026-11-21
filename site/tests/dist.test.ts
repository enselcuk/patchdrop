import { readFileSync } from 'node:fs'
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
