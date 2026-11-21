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
 * Page.astro sets window.goatcounter = { no_onload: true } before count.js, so count.js
 * does not send the first view itself. We send it (exactly once) when count.js has run,
 * then strip the campaign params one tick later. If count.js errors (blocked) nothing will read the
 * URL, so strip at once. If neither happens, strip 3 s after window load (never earlier).
 */
export function wireCampaignCleanup(doc: Document, env: Window): void {
  const gcOf = () => (env as Window & { goatcounter?: GoatCounter & { count?: (v?: unknown) => void } }).goatcounter
  const script = doc.querySelector('script[data-goatcounter]')
  let counted = false
  let stripped = false
  let timer: number | undefined

  const strip = () => {
    if (stripped) return
    stripped = true
    if (timer !== undefined) env.clearTimeout(timer)
    const href = env.location?.href
    if (!href || !env.history?.replaceState) return
    try {
      const cleaned = stripCampaignParams(href)
      if (cleaned) env.history.replaceState(null, '', cleaned)
    } catch {
      // ignore
    }
  }
  const pageview = () => {
    if (counted) return
    counted = true
    try { (gcOf()?.count as ((v?: unknown) => void) | undefined)?.() } catch { /* never break the page */ }
    env.setTimeout(strip, 0)
  }

  if (typeof gcOf()?.count === 'function') {
    pageview()
    return
  }
  script?.addEventListener('load', pageview, { once: true })
  script?.addEventListener('error', () => { counted = true; strip() }, { once: true })
  const arm = () => { timer = env.setTimeout(strip, STRIP_TIMEOUT_MS) }
  if (doc.readyState === 'complete' || typeof env.addEventListener !== 'function') arm()
  else env.addEventListener('load', arm, { once: true })
}
