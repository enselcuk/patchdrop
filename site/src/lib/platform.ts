export type Platform = 'mac' | 'windows' | 'linux' | 'mobile' | 'unknown'

export interface PlatformEnv {
  ua: string
  platform?: string
  maxTouchPoints?: number
  uaDataPlatform?: string
  uaDataMobile?: boolean
}

const MOBILE_UA = /iPhone|iPad|Android|Mobile/
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
  if (/Macintosh/.test(env.ua) && (env.maxTouchPoints ?? 0) > 1) return 'mobile'
  if (MAC_UA.test(env.ua)) return 'mac'
  if (WINDOWS_UA.test(env.ua)) return 'windows'
  if (LINUX_UA.test(env.ua)) return 'linux'
  return 'unknown'
}
