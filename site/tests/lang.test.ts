import { expect, test } from 'vitest'
import { LANG_KEY, REDIRECT_SCRIPT, pickLanguage } from '../src/lib/lang'

test('turkish browser, no saved choice → tr', () => expect(pickLanguage(null, ['tr-TR', 'en'])).toBe('tr'))
test('saved en overrides turkish browser', () => expect(pickLanguage('en', ['tr'])).toBe('en'))
test('saved tr on english browser → tr', () => expect(pickLanguage('tr', ['en-US'])).toBe('tr'))
test('only the FIRST browser language counts', () => expect(pickLanguage(null, ['en-US', 'tr'])).toBe('en'))
test('empty list / garbage saved value → en', () => {
  expect(pickLanguage(null, [])).toBe('en')
  expect(pickLanguage('de', ['de'])).toBe('en')
})

// Evaluate the inline redirect source and compare with pickLanguage.
function runInline(saved: string | null, langs: string[], throws = false): string | null {
  let target: string | null = null
  const localStorage = {
    getItem: (k: string) => {
      if (throws) throw new Error('denied')
      expect(k).toBe(LANG_KEY)
      return saved
    },
  }
  const navigator = { languages: langs, language: langs[0] }
  const location = { search: '?a=1', hash: '#x', replace: (u: string) => { target = u } }
  new Function('localStorage', 'navigator', 'location', REDIRECT_SCRIPT)(localStorage, navigator, location)
  return target
}

test('inline redirect script agrees with pickLanguage', () => {
  const saveds = [null, 'en', 'tr', 'de']
  const lists = [[], ['tr-TR', 'en'], ['en-US', 'tr'], ['tr'], ['de']]
  for (const s of saveds) for (const l of lists) {
    const expected = pickLanguage(s, l) === 'tr' ? '/tr/?a=1#x' : null
    expect(runInline(s, l), `${s} ${l}`).toBe(expected)
  }
})

test('inline redirect survives throwing localStorage', () => {
  expect(runInline(null, ['tr'], true)).toBe('/tr/?a=1#x')
  expect(runInline(null, ['en'], true)).toBeNull()
})
