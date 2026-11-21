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
