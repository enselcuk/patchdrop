export interface GoatCounter { count?: (vars: { path: string; title: string; event: boolean }) => void }

const STRIP_TIMEOUT_MS = 3000

/** Remove utm_* and ref from a URL. Returns the cleaned URL, or null when nothing was removed. */
export function stripCampaignParams(href: string): string | null {
  const url = new URL(href)
  const drop = [...url.searchParams.keys()].filter((k) => k.startsWith('utm_') || k === 'ref')
  if (drop.length === 0) return null
  for (const k of new Set(drop)) url.searchParams.delete(k)
  return url.toString()
}

/** Count a named event. No-op when GoatCounter is missing (blocked); never throws. */
export function trackEvent(name: string, gc?: GoatCounter): void {
  try {
    if (gc && typeof gc.count === 'function') gc.count({ path: name, title: name, event: true })
  } catch {
    // analytics must never break the page
  }
}

/**
 * Strip campaign params from the address bar only after GoatCounter has had the
 * chance to read them: script load (and window load, when it sends the first view)
 * plus one tick; or after 3 s if the script is blocked.
 */
export function wireCampaignCleanup(doc: Document, env: Window): void {
  const href = env.location?.href
  if (!href || !env.history?.replaceState) return
  let cleaned: string | null
  try { cleaned = stripCampaignParams(href) } catch { return }
  if (!cleaned) return
  const target = cleaned
  let done = false
  const strip = () => {
    if (done) return
    done = true
    env.clearTimeout(fallback)
    try { env.history.replaceState(null, '', target) } catch { /* ignore */ }
  }
  const fallback = env.setTimeout(strip, STRIP_TIMEOUT_MS)
  const afterTick = () => { env.setTimeout(strip, 0) }
  doc.querySelector('script[data-goatcounter]')?.addEventListener('load', () => {
    // count.js sends the first view on window load; wait for that too
    if (doc.readyState === 'complete' || typeof env.addEventListener !== 'function') afterTick()
    else env.addEventListener('load', afterTick, { once: true })
  }, { once: true })
}
