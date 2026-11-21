import { detectPlatform } from '../lib/platform'
import { formatCount, parseLiveStats, shouldShowDownloads } from '../lib/release'
import { copyOrSelect } from '../lib/share'

const API = 'https://api.github.com/repos/enselcuk/patchdrop'
const TIMEOUT_MS = 4000

type Nav = Navigator & { userAgentData?: { platform?: string; mobile?: boolean } }

function showHero(doc: Document, env: Window): void {
  const nav = env.navigator as Nav
  const p = detectPlatform({
    ua: nav.userAgent ?? '',
    platform: nav.platform,
    maxTouchPoints: nav.maxTouchPoints,
    uaDataPlatform: nav.userAgentData?.platform,
    uaDataMobile: nav.userAgentData?.mobile,
  })
  const key = p === 'mac' || p === 'windows' || p === 'mobile' ? p : 'other'
  doc.querySelectorAll<HTMLElement>('[data-hero-platform]').forEach((el) => { el.hidden = el.dataset.for !== key })
}

async function liveStats(doc: Document, env: Window): Promise<void> {
  const lang = doc.documentElement.lang === 'tr' ? 'tr' : 'en'
  const ctrl = new AbortController()
  const timer = env.setTimeout(() => ctrl.abort(), TIMEOUT_MS)
  try {
    const get = async (path: string): Promise<unknown> => {
      const res = await env.fetch(`${API}${path}`, { signal: ctrl.signal, headers: { Accept: 'application/vnd.github+json' } })
      if (!res.ok) throw new Error(String(res.status))
      return res.json()
    }
    const [repo, releases] = await Promise.all([get(''), get('/releases?per_page=100')])
    const { stars, downloads } = parseLiveStats(repo, releases)
    const starsEl = doc.querySelector('[data-stars]')
    if (starsEl && typeof stars === 'number' && Number.isFinite(stars)) starsEl.textContent = formatCount(stars, lang)
    const dlEl = doc.querySelector<HTMLElement>('[data-downloads]')
    const dlCount = doc.querySelector('[data-downloads-count]')
    if (dlEl && dlCount && typeof downloads === 'number' && shouldShowDownloads(downloads)) {
      dlCount.textContent = formatCount(downloads, lang)
      dlEl.hidden = false
    }
  } catch {
    // keep the numbers baked in at build time
  } finally {
    env.clearTimeout(timer)
  }
}

function wireShare(doc: Document, env: Window): void {
  const section = doc.querySelector<HTMLElement>('#share')
  const input = doc.querySelector<HTMLInputElement>('[data-share-url]')
  if (!section || !input) return
  const nav = env.navigator as Navigator
  const status = doc.querySelector<HTMLElement>('[data-share-status]')
  const copyBtn = doc.querySelector<HTMLElement>('[data-track="share-copy"]')
  const nativeBtn = doc.querySelector<HTMLElement>('[data-track="share-native"]')
  const sendBtn = doc.querySelector<HTMLElement>('[data-track="send-to-computer"]')
  const canShare = typeof nav.share === 'function'
  const data = () => ({ title: section.dataset.shareTitle ?? 'PatchDrop', text: section.dataset.shareText ?? '', url: input.value })

  const announce = (result: 'copied' | 'selected') => {
    if (status) status.textContent = result === 'copied' ? (copyBtn?.dataset.copied ?? '') : ''
  }
  const copy = async () => announce(await copyOrSelect(input.value, input, nav.clipboard))

  copyBtn?.addEventListener('click', () => { void copy() })
  if (nativeBtn && canShare) {
    nativeBtn.hidden = false
    nativeBtn.addEventListener('click', () => { nav.share(data()).catch(() => {}) })
  }
  sendBtn?.addEventListener('click', (e) => {
    if (canShare) {
      e.preventDefault()
      nav.share(data()).catch(() => {})
    } else {
      // href="#share" still scrolls to the link field; copy or select it as we go.
      void copy()
    }
  })
}

export async function enhance(doc: Document, env: Window): Promise<void> {
  showHero(doc, env)
  wireShare(doc, env)
  await liveStats(doc, env)
}
