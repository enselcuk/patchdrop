export type Platform = 'mac' | 'windows' | 'linux' | 'mobile' | 'unknown'
/** Which hero block (`[data-hero-platform][data-for=…]`) a platform gets. */
export type HeroBlock = 'mac' | 'windows' | 'mobile' | 'other'

export interface PlatformEnv {
  ua: string
  maxTouchPoints?: number
  uaDataMobile?: boolean
}

const MOBILE_UA = /iPhone|iPad|Android|Mobile/
const IPAD_DESKTOP_UA = /Macintosh/
const MAC_UA = /Mac/
const WINDOWS_UA = /Windows/
const LINUX_UA = /Linux|CrOS|X11/

// Pure: classifies by user-agent and touch support only. Checks run in order and
// the first match wins.
export function detectPlatform(env: PlatformEnv): Platform {
  if (env.uaDataMobile) return 'mobile'
  if (MOBILE_UA.test(env.ua)) return 'mobile'
  // iPadOS Safari reports a Macintosh UA; a touch-capable Mac with more than one
  // touch point is an iPad, not a Mac.
  if (IPAD_DESKTOP_UA.test(env.ua) && (env.maxTouchPoints ?? 0) > 1) return 'mobile'
  if (MAC_UA.test(env.ua)) return 'mac'
  if (WINDOWS_UA.test(env.ua)) return 'windows'
  if (LINUX_UA.test(env.ua)) return 'linux'
  return 'unknown'
}

export function heroBlock(p: Platform): HeroBlock {
  return p === 'mac' || p === 'windows' || p === 'mobile' ? p : 'other'
}

/**
 * Self-contained copy of detectPlatform + heroBlock, inlined into <head> so the
 * hero shows the right download block before first paint (CSS keys off
 * `html[data-platform]`). Without JS the attribute is never set and the
 * "other" block stays visible. Expects `navigator` and `document` in scope.
 * Kept in sync by tests/platform.test.ts.
 */
export const PLATFORM_SCRIPT = `try{var n=navigator,u=n.userAgent||'',d=n.userAgentData,p;if(d&&d.mobile||/${MOBILE_UA.source}/.test(u)||/${IPAD_DESKTOP_UA.source}/.test(u)&&(n.maxTouchPoints||0)>1){p='mobile'}else if(/${MAC_UA.source}/.test(u)){p='mac'}else if(/${WINDOWS_UA.source}/.test(u)){p='windows'}else{p='other'}document.documentElement.dataset.platform=p}catch(e){}`
