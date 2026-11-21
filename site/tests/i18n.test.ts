import { expect, test } from 'vitest'
import en from '../src/i18n/en'
import tr from '../src/i18n/tr'
import { getDict } from '../src/lib/lang'

function deepKeys(o: unknown, prefix = ''): string[] {
  if (Array.isArray(o)) return o.flatMap((v, i) => deepKeys(v, `${prefix}[${i}]`))
  if (o && typeof o === 'object') {
    return Object.entries(o).flatMap(([k, v]) => deepKeys(v, prefix ? `${prefix}.${k}` : k)).sort()
  }
  return [prefix]
}
function leaves(o: unknown): unknown[] {
  if (Array.isArray(o)) return o.flatMap(leaves)
  if (o && typeof o === 'object') return Object.values(o).flatMap(leaves)
  return [o]
}

test('tr has exactly the same keys as en', () => expect(deepKeys(tr)).toEqual(deepKeys(en)))
test('no empty values', () => {
  for (const d of [en, tr]) for (const v of leaves(d)) expect(typeof v === 'string' && v.trim() !== '').toBe(true)
})
test('top-level sections', () => {
  expect(Object.keys(en).sort()).toEqual(
    ['meta', 'beta', 'nav', 'hero', 'why', 'how', 'features', 'download', 'share', 'faq', 'madeBy', 'footer', 'notFound'].sort())
})
test('getDict', () => { expect(getDict('en')).toBe(en); expect(getDict('tr')).toBe(tr) })
