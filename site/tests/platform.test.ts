import { expect, test } from 'vitest'
import { detectPlatform } from '../src/lib/platform'

const UA = {
  macSafari: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Safari/605.1.15',
  macChrome: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36',
  windowsChrome: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36',
  windowsEdge: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36 Edg/125.0.0.0',
  windowsArmEdge: 'Mozilla/5.0 (Windows NT 10.0; Win64; ARM64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36 Edg/125.0.0.0',
  iPadSafari: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Safari/605.1.15',
  iPhoneSafari: 'Mozilla/5.0 (iPhone; CPU iPhone OS 17_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Mobile/15E148 Safari/604.1',
  androidPhone: 'Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Mobile Safari/537.36',
  androidTablet: 'Mozilla/5.0 (Linux; Android 14; SM-X810) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36',
  ubuntuFirefox: 'Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:126.0) Gecko/20100101 Firefox/126.0',
  chromeOs: 'Mozilla/5.0 (X11; CrOS x86_64 15633.69.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36',
}

test('Mac Safari → mac', () => expect(detectPlatform({ ua: UA.macSafari, maxTouchPoints: 0 })).toBe('mac'))
test('Mac Chrome → mac', () => expect(detectPlatform({ ua: UA.macChrome, maxTouchPoints: 0 })).toBe('mac'))
test('Windows 11 Chrome → windows', () => expect(detectPlatform({ ua: UA.windowsChrome, maxTouchPoints: 0 })).toBe('windows'))
test('Edge on Windows → windows', () => expect(detectPlatform({ ua: UA.windowsEdge, maxTouchPoints: 0 })).toBe('windows'))
test('Edge on Windows on ARM → windows', () => expect(detectPlatform({ ua: UA.windowsArmEdge, maxTouchPoints: 0 })).toBe('windows'))
test('iPadOS Safari (Macintosh UA, touch) → mobile', () => expect(detectPlatform({ ua: UA.iPadSafari, maxTouchPoints: 5 })).toBe('mobile'))
test('iPhone Safari → mobile', () => expect(detectPlatform({ ua: UA.iPhoneSafari, maxTouchPoints: 5 })).toBe('mobile'))
test('Android phone Chrome → mobile', () => expect(detectPlatform({ ua: UA.androidPhone, maxTouchPoints: 5 })).toBe('mobile'))
test('Android tablet Chrome → mobile', () => expect(detectPlatform({ ua: UA.androidTablet, maxTouchPoints: 5 })).toBe('mobile'))
test('Ubuntu Firefox → linux', () => expect(detectPlatform({ ua: UA.ubuntuFirefox, maxTouchPoints: 0 })).toBe('linux'))
test('ChromeOS Chrome → linux', () => expect(detectPlatform({ ua: UA.chromeOs, maxTouchPoints: 0 })).toBe('linux'))
test('uaDataMobile true overrides a desktop UA → mobile', () =>
  expect(detectPlatform({ ua: UA.macChrome, maxTouchPoints: 0, uaDataMobile: true })).toBe('mobile'))
test('empty UA → unknown', () => expect(detectPlatform({ ua: '' })).toBe('unknown'))

// Extra cases beyond the table.
test('Mac UA with maxTouchPoints 1 (touch-capable display) → mac', () =>
  expect(detectPlatform({ ua: UA.macSafari, maxTouchPoints: 1 })).toBe('mac'))
test('Mac UA with maxTouchPoints undefined → mac', () =>
  expect(detectPlatform({ ua: UA.macSafari })).toBe('mac'))
