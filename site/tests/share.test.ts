// @vitest-environment jsdom
import { expect, test } from 'vitest'
import { copyOrSelect, shareLinks, shareText } from '../src/lib/share'

test('x link mentions via=enselcuk and encodes text', () => {
  const u = new URL(shareLinks('https://patchdrop.enginsistem.com/tr/', 'tr').x)
  expect(u.origin + u.pathname).toBe('https://x.com/intent/post')
  expect(u.searchParams.get('via')).toBe('enselcuk')
  expect(u.searchParams.get('url')).toBe('https://patchdrop.enginsistem.com/tr/')
  expect(u.searchParams.get('text')).toBe(shareText('tr'))
})
test('linkedin/facebook/whatsapp carry the page url', () => {
  const l = shareLinks('https://patchdrop.enginsistem.com/', 'en')
  expect(l.linkedin).toBe('https://www.linkedin.com/sharing/share-offsite/?url=https%3A%2F%2Fpatchdrop.enginsistem.com%2F')
  expect(l.facebook).toBe('https://www.facebook.com/sharer/sharer.php?u=https%3A%2F%2Fpatchdrop.enginsistem.com%2F')
  expect(l.whatsapp).toContain(encodeURIComponent('https://patchdrop.enginsistem.com/'))
})
test('tr share text is the spec text', () =>
  expect(shareText('tr')).toBe('Kablolu ofis ağında Mac ve Windows arasında dosya paylaşmak için ücretsiz ve açık kaynak bir uygulama: PatchDrop'))
test('clipboard missing selects the visible field', async () => {
  const i = document.createElement('input')
  i.value = 'x'
  expect(await copyOrSelect('x', i, undefined)).toBe('selected')
})
test('clipboard rejects selects', async () => {
  const i = document.createElement('input')
  i.value = 'x'
  expect(await copyOrSelect('x', i, { writeText: () => Promise.reject(new Error('no')) } as unknown as Clipboard)).toBe('selected')
})
test('clipboard works copies', async () => {
  const i = document.createElement('input')
  let got = ''
  expect(await copyOrSelect('x', i, { writeText: async (s: string) => { got = s } } as unknown as Clipboard)).toBe('copied')
  expect(got).toBe('x')
})
