# PatchDrop Landing Page Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A static, bilingual (EN at `/`, TR at `/tr/`) landing page for PatchDrop at `patchdrop.enginsistem.com`, published from `site/` to GitHub Pages.

**Architecture:** Astro static site in `site/`. All copy lives in two typed dictionaries; one layout and section components render both languages. Small pure TypeScript modules (`site/src/lib/`) hold every piece of logic (language choice, platform detection, release data, share links, analytics helpers). They're unit-tested with Vitest, and a build-output test checks the generated HTML. Client scripts only enhance the page: every link works without JavaScript.

**Tech Stack:** Astro 7.x, TypeScript, Vitest, `node-html-parser` (dist tests), `@fontsource-variable/manrope`, `sharp` (image conversion, already an Astro dependency), GitHub Actions (`actions/upload-pages-artifact`, `actions/deploy-pages`), GoatCounter.

**Spec:** `docs/superpowers/specs/2026-11-21-landing-page-design.md`

## Global Constraints

- Site URL `https://patchdrop.enginsistem.com`; `site/public/CNAME` = `patchdrop.enginsistem.com`.
- Repo `enselcuk/patchdrop`; download base `https://github.com/enselcuk/patchdrop/releases/latest/download/`.
- Asset names (exact): `PatchDrop-macOS.dmg`, `PatchDrop-macOS-universal.zip`, `PatchDrop-Windows-Setup-x64.exe`, `PatchDrop-Windows-x64.zip`, `SHA256SUMS.txt`.
- Colors: background `#FFFFFF` → `#EEF4FF`, accent gradient `#2F6BFF` → `#7A5CFF`, text `#0B1B33`, beta amber `#FFF1D6` / `#9A5B00`. Font Manrope (self-hosted). Light mode only. Pill buttons.
- Copyright line: `© 2026 Engin Tasarım Ltd. Şti.`
- No cookies, no `localStorage` other than key `patchdrop-lang`, no third-party scripts except GoatCounter.
- No `nofollow` on outbound links; enginsistem.com links carry `?utm_source=patchdrop`.
- Lighthouse mobile ≥ 95 in all four categories; first-load weight < ~500 KB excluding lazy screenshot.
- Every download link, star link and share link is a real `<a href>` that works with JavaScript off.
- macOS first-launch instructions: System Settings → Privacy & Security → Open Anyway (not right-click → Open).
- Made-by links (exact):
  - Selçuk: LinkedIn `https://www.linkedin.com/in/selcuk-engin/`, X `https://x.com/enselcuk`, YouTube `https://www.youtube.com/@selcukengin`.
  - Engin Sistem: `https://enginsistem.com`, X `https://x.com/enginsistem`, Instagram `https://www.instagram.com/enginsistem/`, YouTube `https://www.youtube.com/@enginsistem`.

## Review Focus

1. **Language preference vs. browser language.** A Turkish-browser visitor who chose EN must stay on `/` (no redirect, no loop). `/tr/` must never redirect. A first visit with `#download` lands on `/tr/#download`. → tests in Task 2.
2. **GitHub API down or rate limited (60/h per IP).** Visitors behind one office IP hit this quickly. The page must keep the build-time numbers, never show `NaN`, `undefined` or an empty badge. → tests in Task 3.
3. **iPad and Android tablets.** iPadOS Safari reports a Mac user agent. It must be treated as mobile (no DMG button), using `maxTouchPoints > 1`. → tests in Task 4.
4. **Arrival with `?utm_source=…` plus an anchor, or a re-share.** The campaign must be counted once and the parameter removed from the address bar without losing the anchor, so shares don't carry `utm_source=linkedin`. → tests in Task 7.
5. **Clipboard or native share unavailable** (desktop Firefox has no `navigator.share`; HTTP or old browsers lack the clipboard API). "Copy link" and "Send link to my computer" must fall back to selecting a visible URL field, not fail silently. → tests in Task 6.

---

## File Structure

```
site/
  package.json, astro.config.mjs, tsconfig.json, vitest.config.ts
  public/CNAME, robots.txt, favicon.svg, og-en.png, og-tr.png
  src/
    i18n/en.ts, tr.ts, types.ts          copy; Dict type = shape of en.ts
    lib/lang.ts                          pickLanguage, LANG_KEY
    lib/platform.ts                      detectPlatform
    lib/release.ts                       ReleaseInfo, ASSETS, downloadUrl, getReleaseInfo, formatCount, shouldShowDownloads
    lib/release-fallback.json            last-known release data (committed)
    lib/share.ts                         shareLinks, shareText
    lib/analytics.ts                     stripCampaignParams, trackEvent
    layouts/Page.astro                   <html>, head/SEO, beta bar, header, footer, scripts
    components/Seo.astro, BetaBar.astro, Header.astro, Hero.astro, Why.astro,
               HowItWorks.astro, Features.astro, Download.astro, Share.astro,
               Faq.astro, MadeBy.astro, Footer.astro, LangSwitch.astro
    scripts/enhance.ts                   client: platform, live stats, share, analytics, lang switch
    pages/index.astro, tr/index.astro, 404.astro, tr/404.astro
    assets/screenshot-en.png, screenshot-tr.png, selcuk.jpg, logo.svg
  tests/*.test.ts                        unit tests
  tests/dist.test.ts                     built-HTML checks (runs after `astro build`)
.github/workflows/site.yml
```

Repo files touched outside `site/`: `LICENSE`, `wails.json`, `build/windows/installer/project.nsi`, `README.md`.

---

### Task 1: Astro scaffold, build-output test harness, deploy workflow

**Files:**
- Create: `site/package.json`, `site/astro.config.mjs`, `site/tsconfig.json`, `site/vitest.config.ts`, `site/public/CNAME`, `site/public/robots.txt`, `site/src/layouts/Page.astro` (minimal), `site/src/pages/index.astro`, `site/src/pages/tr/index.astro`, `site/tests/dist.test.ts`, `.github/workflows/site.yml`
- Modify: `.gitignore` (add `site/node_modules/`, `site/dist/`, `site/.astro/`)

**Interfaces:**
- Produces: `Page.astro` props `{ lang: 'en' | 'tr'; path: '/' | '/tr/' }`; npm scripts `build` (`astro build`), `test` (`vitest run --exclude tests/dist.test.ts`), `test:dist` (`vitest run tests/dist.test.ts`); helper in `tests/dist.test.ts`: `load(path: string): HTMLElement` that parses `dist/<path>/index.html`.

- [ ] **Step 1:** Create the project with `npm create astro@latest site -- --template minimal --typescript strict --no-git --install`. Set `site: 'https://patchdrop.enginsistem.com'`, `trailingSlash: 'always'`, `i18n: { defaultLocale: 'en', locales: ['en','tr'], routing: { prefixDefaultLocale: false } }` in `astro.config.mjs`. Add dev deps `vitest`, `node-html-parser`.
- [ ] **Step 2: Write the failing dist test**

```ts
test('both language pages exist with the right lang', () => {
  expect(load('').querySelector('html')!.getAttribute('lang')).toBe('en')
  expect(load('tr').querySelector('html')!.getAttribute('lang')).toBe('tr')
})
test('CNAME is published', () => {
  expect(readFileSync('dist/CNAME', 'utf8').trim()).toBe('patchdrop.enginsistem.com')
})
```

- [ ] **Step 3:** Run `cd site && npm run build && npm run test:dist`. Expected: FAIL (`dist/tr/index.html` missing).
- [ ] **Step 4:** Make `Page.astro` render `<html lang={lang}>`, and the two pages use it. Add `public/CNAME` and `robots.txt` (`Sitemap: https://patchdrop.enginsistem.com/sitemap-index.xml`; add `@astrojs/sitemap` with i18n `{ defaultLocale: 'en', locales: { en: 'en', tr: 'tr' } }`).
- [ ] **Step 5:** Run `npm run build && npm run test:dist`. Expected: PASS.
- [ ] **Step 6:** Write `.github/workflows/site.yml`:
  - **Triggers:** `push` to `main` with paths `site/**` and `.github/workflows/site.yml`, `release: [published]`, and `workflow_dispatch`.
  - **Permissions:** `pages: write`, `id-token: write`, `contents: read`.
  - **Build job:** Node 22, `npm ci` in `site/`, `npm test`, `npm run build` with env `GITHUB_TOKEN` (for the release fetch, Task 3), `npm run test:dist`, then `upload-pages-artifact` with path `site/dist`.
  - **Deploy job:** `deploy-pages`.
  - **Concurrency:** group `pages`.
- [ ] **Step 7: Commit:** `git add site .github .gitignore && git commit -m "site: scaffold Astro landing page and Pages workflow"`

### Task 2: Copy dictionaries, SEO head, language choice and switcher

**Files:**
- Create: `site/src/i18n/types.ts`, `en.ts`, `tr.ts`, `site/src/lib/lang.ts`, `site/src/components/Seo.astro`, `LangSwitch.astro`, `site/tests/i18n.test.ts`, `site/tests/lang.test.ts`
- Modify: `site/src/layouts/Page.astro`, `site/tests/dist.test.ts`

**Interfaces:**
- Produces:
  - `type Dict` (= `typeof en`); `getDict(lang: 'en'|'tr'): Dict`.
  - `LANG_KEY = 'patchdrop-lang'`.
  - `pickLanguage(saved: string | null, browserLangs: readonly string[]): 'en' | 'tr'`.
  - `Seo.astro` props `{ lang, path, title, description, ogImage }`.
  - Dictionary top-level keys (each later task fills its own section): `meta`, `beta`, `nav`, `hero`, `why`, `how`, `features`, `download`, `share`, `faq`, `madeBy`, `footer`, `notFound`.

- [ ] **Step 1: Write failing tests**

```ts
// lang.test.ts
test('turkish browser, no saved choice → tr', () => expect(pickLanguage(null, ['tr-TR','en'])).toBe('tr'))
test('saved en overrides turkish browser', () => expect(pickLanguage('en', ['tr'])).toBe('en'))
test('saved tr on english browser → tr', () => expect(pickLanguage('tr', ['en-US'])).toBe('tr'))
test('only the FIRST browser language counts', () => expect(pickLanguage(null, ['en-US','tr'])).toBe('en'))
test('empty list / garbage saved value → en', () => { expect(pickLanguage(null, [])).toBe('en'); expect(pickLanguage('de', ['de'])).toBe('en') })
// i18n.test.ts
test('tr has exactly the same keys as en, none empty', () => expect(deepKeys(tr)).toEqual(deepKeys(en)))
// dist.test.ts additions
test('hreflang + canonical on both pages', () => { for (const p of ['', 'tr']) {
  const h = load(p)
  expect(h.querySelector('link[hreflang="en"]')!.getAttribute('href')).toBe('https://patchdrop.enginsistem.com/')
  expect(h.querySelector('link[hreflang="tr"]')!.getAttribute('href')).toBe('https://patchdrop.enginsistem.com/tr/')
  expect(h.querySelector('link[hreflang="x-default"]')!.getAttribute('href')).toBe('https://patchdrop.enginsistem.com/') } })
test('per-language OG tags', () => {
  expect(load('tr').querySelector('meta[property="og:image"]')!.getAttribute('content')).toBe('https://patchdrop.enginsistem.com/og-tr.png')
  expect(load('').querySelector('meta[name="twitter:card"]')!.getAttribute('content')).toBe('summary_large_image') })
test('only / carries the redirect script', () => {
  expect(load('').querySelector('script[data-lang-redirect]')).toBeTruthy()
  expect(load('tr').querySelector('script[data-lang-redirect]')).toBeNull() })
```

- [ ] **Step 2:** Run `npm test && npm run build && npm run test:dist`. Expected: FAIL (modules missing).
- [ ] **Step 3:** Implement `pickLanguage` in `lib/lang.ts`. Write `en.ts`/`tr.ts` with the `meta`, `nav` and `notFound` sections filled and every other section present. `Dict = typeof en`, and `tr.ts` is typed `: Dict` so a missing key fails `astro check`. Write `Seo.astro` (title, description, canonical, three hreflang links, `og:*`, `twitter:*`, `og:locale` `en_US`/`tr_TR`). `Page.astro` on `/` only emits an `is:inline` `<script data-lang-redirect>` in `<head>`. It reads `localStorage[LANG_KEY]` inside try/catch, and if `pickLanguage(...) === 'tr'`, runs `location.replace('/tr/' + location.search + location.hash)`. Inline the logic: no import at that point. A unit test checks the inline copy agrees with `pickLanguage`, by exporting the inline source as a string from `lang.ts` and evaluating it in the test. `LangSwitch.astro` renders two `<a>` links (`/` and `/tr/`). Clicking saves `LANG_KEY` before navigating.
- [ ] **Step 4:** Run the same commands. Expected: PASS.
- [ ] **Step 5: Commit:** `git commit -m "site: i18n dictionaries, SEO head, language detection"`

### Task 3: Release data and live numbers

**Files:**
- Create: `site/src/lib/release.ts`, `site/src/lib/release-fallback.json`, `site/tests/release.test.ts`

**Interfaces:**
- Produces:
  - `ASSETS = { macDmg, macZip, winSetup, winZip, sums }`, holding the exact names from Global Constraints.
  - `downloadUrl(name: string): string`.
  - `interface ReleaseInfo { version: string; publishedAt: string; stars: number; downloads: number; sizes: Record<string, number> }`.
  - `getReleaseInfo(fetchFn?: typeof fetch): Promise<ReleaseInfo>`, used at build time. It uses `process.env.GITHUB_TOKEN` when present.
  - `formatCount(n: number, lang): string`.
  - `shouldShowDownloads(n: number): boolean`.
  - `parseLiveStats(repoJson: unknown, releasesJson: unknown): { stars?: number; downloads?: number }`, used client-side.

- [ ] **Step 1: Write failing tests**

```ts
test('downloadUrl uses latest/download', () => expect(downloadUrl(ASSETS.winSetup)).toBe('https://github.com/enselcuk/patchdrop/releases/latest/download/PatchDrop-Windows-Setup-x64.exe'))
test('API failure → fallback json, no throw', async () => expect(await getReleaseInfo(async () => { throw new Error('offline') })).toEqual(fallback))
test('403 rate limit → fallback', async () => expect(await getReleaseInfo(async () => new Response('', { status: 403 }))).toEqual(fallback))
test('downloads = sum of download_count over all releases and assets', () => expect(parseLiveStats({ stargazers_count: 7 }, [{ assets: [{ download_count: 3 }, { download_count: 4 }] }, { assets: [{ download_count: 5 }] }])).toEqual({ stars: 7, downloads: 12 }))
test('garbage JSON → empty object, never NaN', () => expect(parseLiveStats({ message: 'API rate limit exceeded' }, null)).toEqual({}))
test('threshold', () => { expect(shouldShowDownloads(99)).toBe(false); expect(shouldShowDownloads(100)).toBe(true) })
test('formatCount', () => { expect(formatCount(1234, 'en')).toBe('1.2k'); expect(formatCount(1234, 'tr')).toBe('1,2B'); expect(formatCount(12, 'tr')).toBe('12') })
```

- [ ] **Step 2:** Run `npm test -- release`. Expected: FAIL.
- [ ] **Step 3:** Implement the functions. `getReleaseInfo` calls `GET /repos/enselcuk/patchdrop` and `GET /repos/enselcuk/patchdrop/releases?per_page=100`. The version comes from the newest non-draft release. Sizes are read by asset name. On any failure or missing asset it returns the fallback. `formatCount` uses `Intl.NumberFormat(lang, { notation: 'compact', maximumFractionDigits: 1 })`. Turkish compact output for thousands is "B" (bin); the test pins the real output. If the runtime prints `1,2 B`, update the assertion to match, without hand-rolling. Fallback JSON holds v0.1.0: version `0.1.0`, stars `0`, downloads `0`, sizes for the five assets taken from the current release.
- [ ] **Step 4:** Run `npm test -- release`. Expected: PASS.
- [ ] **Step 5: Commit:** `git commit -m "site: release data with fallback and live-count parsing"`

### Task 4: Platform detection

**Files:**
- Create: `site/src/lib/platform.ts`, `site/tests/platform.test.ts`

**Interfaces:**
- Produces: `type Platform = 'mac' | 'windows' | 'linux' | 'mobile' | 'unknown'`; `detectPlatform(env: { ua: string; platform?: string; maxTouchPoints?: number; uaDataPlatform?: string; uaDataMobile?: boolean }): Platform`.

- [ ] **Step 1: Write failing tests**, with real UA strings:

| Device | Inputs | Expected |
|---|---|---|
| Mac Safari/Chrome | `maxTouchPoints` 0 | `'mac'` |
| iPadOS Safari | Mac UA, `maxTouchPoints` 5 | `'mobile'` |
| iPhone | | `'mobile'` |
| Android phone | | `'mobile'` |
| Android tablet | | `'mobile'` |
| Windows 11 Edge | | `'windows'` |
| Windows on ARM | | `'windows'` |
| Ubuntu Firefox | | `'linux'` |
| ChromeOS | | `'linux'` |
| Any UA | `uaDataMobile: true` | `'mobile'` |
| Empty UA | | `'unknown'` |

- [ ] **Step 2:** Run `npm test -- platform`. Expected: FAIL.
- [ ] **Step 3:** Implement the checks in this order, returning at the first match:
  1. `uaDataMobile` set → `'mobile'`
  2. Mobile token (`iPhone|iPad|Android|Mobile`) → `'mobile'`
  3. `Macintosh` with `maxTouchPoints > 1` → `'mobile'`
  4. `Mac` → `'mac'`
  5. `Windows` → `'windows'`
  6. `Linux|CrOS|X11` → `'linux'`
  7. Otherwise → `'unknown'`
- [ ] **Step 4:** Run. Expected: PASS.
- [ ] **Step 5: Commit:** `git commit -m "site: platform detection"`

### Task 5: Page sections and assets

**Files:**
- Create:
  - Section components: `site/src/components/{BetaBar,Header,Hero,Why,HowItWorks,Features,Download,Faq,MadeBy,Footer}.astro`
  - Assets: `site/src/assets/{logo.svg,screenshot-en.png,screenshot-tr.png,selcuk.jpg}`, `site/public/favicon.svg`, `site/public/og-en.png`, `site/public/og-tr.png`
  - 404 pages: `site/src/pages/404.astro`, `site/src/pages/tr/404.astro`
- Modify: `en.ts`, `tr.ts` (all section copy), `Page.astro`, `site/tests/dist.test.ts`

**Interfaces:**
- Consumes: `Dict`, `ASSETS`, `downloadUrl`, `getReleaseInfo` (called once in each page's frontmatter), `Page.astro` props.
- Produces: element hooks used by Task 6–7 scripts. These are exactly the names below:
  - `[data-hero-platform]` blocks with `data-for="mac|windows|other|mobile"`; the `other` block is the server default.
  - `[data-stars]` and `[data-downloads]` (the latter `hidden` by default).
  - `a[data-track]` with values such as `download-mac-dmg`, `download-win-setup`, `download-win-zip`, `download-mac-zip`, `star-hero`, `star-share`, `send-to-computer`.
  - `#download` and `#share` anchors.

- [ ] **Step 1: Write failing dist tests**

```ts
test('server-rendered download links point at stable asset names', () => {
  const hrefs = load('').querySelectorAll('a[data-track^="download-"]').map(a => a.getAttribute('href'))
  expect(hrefs).toContain('https://github.com/enselcuk/patchdrop/releases/latest/download/PatchDrop-macOS.dmg')
  expect(hrefs).toContain('https://github.com/enselcuk/patchdrop/releases/latest/download/PatchDrop-Windows-Setup-x64.exe') })
test('made-by links exact and followed', () => {
  const a = load('tr').querySelectorAll('#made-by a')
  expect(a.map(x => x.getAttribute('href'))).toEqual(expect.arrayContaining(['https://enginsistem.com/?utm_source=patchdrop','https://www.linkedin.com/in/selcuk-engin/','https://x.com/enselcuk','https://www.youtube.com/@selcukengin','https://x.com/enginsistem','https://www.instagram.com/enginsistem/','https://www.youtube.com/@enginsistem']))
  expect(a.every(x => !(x.getAttribute('rel') ?? '').includes('nofollow'))).toBe(true) })
test('mac tips say Open Anyway, never right-click → Open', () => {
  for (const p of ['', 'tr']) { const t = load(p).querySelector('#download')!.text
    expect(t).toMatch(/Open Anyway|Yine de Aç/); expect(t).not.toMatch(/right-click|sağ tık/i) } })
test('footer copyright', () => expect(load('').querySelector('footer')!.text).toContain('© 2026 Engin Tasarım Ltd. Şti.'))
test('404 pages exist in both languages', () => { expect(existsSync('dist/404.html')).toBe(true); expect(existsSync('dist/tr/404/index.html')).toBe(true) })
test('beta bar links to GitHub issues', () => expect(load('').querySelector('[data-beta] a')!.getAttribute('href')).toBe('https://github.com/enselcuk/patchdrop/issues'))
```

- [ ] **Step 2:** Run `npm run build && npm run test:dist`. Expected: FAIL.
- [ ] **Step 3: Assets.**
  - **Photo:** download `https://pbs.twimg.com/profile_images/2055608326021382144/ZDvFBKxv_400x400.jpg` into its own new directory, check it's a JPEG with `file`, then move it to `src/assets/selcuk.jpg`. Render it with Astro `<Image>` at 200×200 WebP.
  - **App screenshots:** run the app with `PATCHDROP_PROFILE=shot wails dev` in light mode. The peers shown in the list are two headless instances, e.g. `PATCHDROP_PROFILE=p1 PATCHDROP_PORT=47861 build/bin/PatchDrop.app/Contents/MacOS/PatchDrop receive`. Each one's name comes from the `deviceName` field in `~/Library/Application Support/PatchDrop-<profile>/config.json`, which you edit before starting it. Use `Tasarım-PC` and `Muhasebe-Mac` for TR, and `Design-PC` and `Accounts-Mac` for EN. Capture at 2× and crop the window. Delete the temporary profiles and `~/Downloads/PatchDrop` afterwards.
  - **Logo and favicon:** `logo.svg` and `favicon.svg` are a rounded square in the accent gradient with a white plug-and-arrow mark.
  - **Share images:** build `og-en.png` and `og-tr.png` (1200×630) from an SVG template (headline + screenshot + logo) with `sharp` in `site/scripts/make-og.mjs`. Commit the PNGs.
- [ ] **Step 4: Sections**, following spec "Page sections" 0–9 and the mockup `page-layout.html`:
  - **Hero:** renders all four `data-hero-platform` blocks. Only `other`, with both buttons, is visible without JS.
  - **Download section:** shows sizes from `ReleaseInfo.sizes` in MB, macOS first-launch steps (Open Anyway) and Windows SmartScreen steps (More info → Run anyway), plus links to the release notes (`https://github.com/enselcuk/patchdrop/releases/latest`) and `SHA256SUMS.txt`.
  - **FAQ:** uses `<details>` items. It includes the troubleshooting answer (firewall, Ethernet set to Private, Add by IP, Scan network, ports UDP 47820 / TCP 47821), "Why not LocalSend?" crediting LocalSend respectfully, and the GoatCounter privacy line.
  - **Layout:** responsive from 360 px; any animation is disabled under `prefers-reduced-motion: reduce`; the screenshot uses `loading="lazy"` except in the hero; Manrope via `@fontsource-variable/manrope` (latin + latin-ext subsets, for Turkish characters).
  - **Copy rules:** write both languages natively, with no literal translation. The TR headline is exactly "Kablo üzerinden, bir sürükle-bırak uzaklıkta." and the bio is the spec's proposal.
- [ ] **Step 5:** Run `npm run build && npm run test:dist && npx astro check`. Expected: PASS, 0 errors.
- [ ] **Step 6: Commit:** `git commit -m "site: page sections, copy, assets, 404"`

### Task 6: Client enhancement: platform, live stats, sharing

**Files:**
- Create: `site/src/lib/share.ts`, `site/src/scripts/enhance.ts`, `site/src/components/Share.astro`, `site/tests/share.test.ts`, `site/tests/enhance.test.ts`
- Modify: `Page.astro` (load `enhance.ts` as a module script), `en.ts`/`tr.ts` (`share` section)

**Interfaces:**
- Consumes: `detectPlatform`, `parseLiveStats`, `formatCount`, `shouldShowDownloads`, Task 5 element hooks.
- Produces:
  - `shareText(lang): string`.
  - `shareLinks(url: string, lang): { linkedin: string; x: string; facebook: string; whatsapp: string }`.
  - `copyOrSelect(url: string, input: HTMLInputElement, clipboard?: Clipboard): Promise<'copied' | 'selected'>`.
  - `enhance(doc: Document, env: Window): void`.

- [ ] **Step 1: Write failing tests**

```ts
test('x link mentions via=enselcuk and encodes text', () => { const u = new URL(shareLinks('https://patchdrop.enginsistem.com/tr/', 'tr').x)
  expect(u.origin + u.pathname).toBe('https://x.com/intent/post'); expect(u.searchParams.get('via')).toBe('enselcuk')
  expect(u.searchParams.get('url')).toBe('https://patchdrop.enginsistem.com/tr/'); expect(u.searchParams.get('text')).toBe(shareText('tr')) })
test('linkedin/facebook/whatsapp carry the page url', () => { const l = shareLinks('https://patchdrop.enginsistem.com/', 'en')
  expect(l.linkedin).toBe('https://www.linkedin.com/sharing/share-offsite/?url=https%3A%2F%2Fpatchdrop.enginsistem.com%2F')
  expect(l.facebook).toBe('https://www.facebook.com/sharer/sharer.php?u=https%3A%2F%2Fpatchdrop.enginsistem.com%2F')
  expect(l.whatsapp).toContain(encodeURIComponent('https://patchdrop.enginsistem.com/')) })
test('tr share text is the spec text', () => expect(shareText('tr')).toBe('Kablolu ofis ağında Mac ve Windows arasında dosya paylaşmak için ücretsiz ve açık kaynak bir uygulama: PatchDrop'))
test('clipboard missing → selects the visible field', async () => { const i = document.createElement('input'); i.value = 'x'
  expect(await copyOrSelect('x', i, undefined)).toBe('selected') })
test('clipboard rejects → selects', async () => expect(await copyOrSelect('x', input, { writeText: () => Promise.reject() } as any)).toBe('selected'))
test('enhance shows only the mac hero block on mac', () => { /* jsdom page with the 4 blocks; env.navigator mac */
  enhance(doc, macEnv); expect(visible(doc)).toEqual(['mac']) })
test('enhance on iPad shows mobile block', () => { enhance(doc, iPadEnv); expect(visible(doc)).toEqual(['mobile']) })
test('enhance keeps baked numbers when fetch fails', async () => { await enhance(doc, envWithFailingFetch); expect(doc.querySelector('[data-stars]')!.textContent).toBe('0') })
```

  Use the `jsdom` Vitest environment for `enhance.test.ts`, via a `// @vitest-environment jsdom` comment, and add the `jsdom` dev dependency.
- [ ] **Step 2:** Run `npm test -- share enhance`. Expected: FAIL.
- [ ] **Step 3:** Implement. In `enhance`:
  1. **Hero:** pick the platform and show the matching hero block.
  2. **Live stats:** fetch repo and releases JSON with a 4 s `AbortController` timeout, then update `[data-stars]`. Unhide `[data-downloads]` only when `shouldShowDownloads(downloads)` is true.
  3. **Native share:** on mobile, a `Share` button uses `navigator.share({ title, text, url })` when it exists, and otherwise scrolls to `#share`.
  4. **Copy link:** the button and "Send link to my computer" use `copyOrSelect`.

  `Share.astro` renders the four network links as plain `<a target="_blank" rel="noopener">`, plus a visible read-only URL input and the Star button.
- [ ] **Step 4:** Run `npm test && npm run build && npm run test:dist`. Expected: PASS.
- [ ] **Step 5: Commit:** `git commit -m "site: platform-aware hero, live GitHub numbers, sharing"`

### Task 7: GoatCounter analytics

**Files:**
- Create: `site/src/lib/analytics.ts`, `site/tests/analytics.test.ts`
- Modify: `Page.astro`, `enhance.ts`, `site/src/config.ts` (create: `export const GOATCOUNTER_CODE = 'patchdrop'`)

**Interfaces:**
- Produces: `stripCampaignParams(href: string): string | null`, which returns the cleaned URL, or `null` if there was nothing to strip. Also `trackEvent(name: string, gc?: { count: Function }): void`, a no-op when GoatCounter isn't loaded.

- [ ] **Step 1: Write failing tests**

```ts
test('removes utm_* and keeps other params and the hash', () => expect(stripCampaignParams('https://patchdrop.enginsistem.com/tr/?utm_source=linkedin&utm_medium=social&x=1#download')).toBe('https://patchdrop.enginsistem.com/tr/?x=1#download'))
test('removes ref too', () => expect(stripCampaignParams('https://patchdrop.enginsistem.com/?ref=x')).toBe('https://patchdrop.enginsistem.com/'))
test('nothing to strip → null', () => expect(stripCampaignParams('https://patchdrop.enginsistem.com/#faq')).toBeNull())
test('trackEvent without goatcounter does not throw', () => expect(() => trackEvent('download-mac-dmg', undefined)).not.toThrow())
test('trackEvent calls count with event:true', () => { const count = vi.fn(); trackEvent('star-hero', { count }); expect(count).toHaveBeenCalledWith({ path: 'star-hero', title: 'star-hero', event: true }) })
```

- [ ] **Step 2:** Run `npm test -- analytics`. Expected: FAIL.
- [ ] **Step 3:** Implement. `Page.astro` adds `<script data-goatcounter="https://${GOATCOUNTER_CODE}.goatcounter.com/count" async src="//gc.zgo.at/count.js">`. GoatCounter reads `utm_source` and `ref` from the URL on its first page view. Strip the parameters only after that page view has been sent: listen for the script's `load` event, then wait one tick, then `history.replaceState(null, '', cleaned)`. If the script never loads (blocked), strip after 3 s anyway. In `enhance.ts`, one delegated click listener calls `trackEvent(el.dataset.track)` for every `a[data-track]`, including share links (`share-linkedin`, `share-x`, `share-facebook`, `share-whatsapp`, `share-copy`, `share-native`) and lang switches (`lang-en`, `lang-tr`).
- [ ] **Step 4:** Run `npm test && npm run build && npm run test:dist`. Expected: PASS. Add a dist assertion that the GoatCounter script tag is present on both pages.
- [ ] **Step 5: Commit:** `git commit -m "site: GoatCounter page views, events and campaign cleanup"`

### Task 8: Repo changes: copyright, installer name, README

**Files:**
- Modify: `LICENSE` (line 3 → `Copyright (c) 2026 Engin Tasarım Ltd. Şti.`), `wails.json` (`info.copyright`, `author.name` → `Engin Tasarım Ltd. Şti.`), `build/windows/installer/project.nsi` (`OutFile "..\..\bin\${INFO_PRODUCTNAME}-Windows-Setup-x64.exe"`), `README.md`

- [ ] **Step 1:** Edit `wails.json`, then rebuild the installer as the README's Docker section describes: `wails build -platform windows/amd64 -nsis -o PatchDrop.exe`, then the Docker `makensis` command. Expected: `build/bin/PatchDrop-Windows-Setup-x64.exe` exists. If NSIS reports "Bad text encoding" for the copyright, set an ASCII override in `project.nsi`: `!define INFO_COPYRIGHT "Copyright (c) 2026 Engin Tasarim Ltd. Sti."`. Put it before the `!include`, so `wails_tools.nsh` keeps it. Then rebuild.
- [ ] **Step 2:** Check the installer: unpack it with 7-Zip in Docker (as done before), and check that `PatchDrop.exe` matches `build/bin/PatchDrop.exe` by sha256.
- [ ] **Step 3:** Update `README.md`:
  - **Top:** a link to `https://patchdrop.enginsistem.com`.
  - **Install table:** the new installer name, and macOS steps "System Settings → Privacy & Security → Open Anyway" in place of right-click → Open.
  - **Beta wording:** "Beta: tested in our office on Mac and Windows."
  - **License:** "MIT, © 2026 Engin Tasarım Ltd. Şti.".
- [ ] **Step 4:** Run `go test ./internal/...`. Expected: all `ok`.
- [ ] **Step 5: Commit:** `git commit -m "Copyright to Engin Tasarım Ltd. Şti., stable installer name, README for launch"`

### Task 9: Quality pass and local review with Selçuk

- [ ] **Step 1:** Run `cd site && npm run build && npx astro preview --port 4321`. Open it in the browser pane.
- [ ] **Step 2:** Check by hand at 375 px and 1280 px:
  - TR redirect (set `navigator.language` via a fresh profile, or test with `?` + saved key)
  - the EN↔TR switch persists
  - hero per platform (DevTools device emulation: iPhone, iPad, Windows, Linux)
  - JS disabled: all links are present
- [ ] **Step 3:** Run Lighthouse (mobile) on `/` and `/tr/`. Expected: all four categories ≥ 95. Check the transfer size in the Network panel: under ~500 KB excluding the lazy screenshot. Fix and re-run until met.
- [ ] **Step 4:** Show Selçuk both languages, and ask him to review all the Turkish copy and the bio. Apply his edits and rebuild. **Stop here until he approves the content.**
- [ ] **Step 5: Commit:** `git commit -m "site: copy review fixes"`

### Task 10: Launch (each GitHub step needs Selçuk's go-ahead right before it)

- [ ] **Step 1:** Re-upload the installer under its new name to v0.1.0: `gh release upload v0.1.0 build/bin/PatchDrop-Windows-Setup-x64.exe`. Then delete the old `PatchDrop-Setup-0.1.0-amd64.exe` and upload the regenerated `SHA256SUMS.txt` with `--clobber`. If Task 8 changed `PatchDrop.exe`, replace `PatchDrop-Windows-x64.zip` too.
- [ ] **Step 2:** Mark it a normal release: `gh release edit v0.1.0 --prerelease=false --latest`. Verify: `curl -sIL https://github.com/enselcuk/patchdrop/releases/latest/download/PatchDrop-macOS.dmg | grep -i '^location'` resolves to the v0.1.0 asset. That only works once the repo is public, so it's re-checked after Step 3.
- [ ] **Step 3:** Make the repo public: `gh repo edit enselcuk/patchdrop --visibility public --accept-visibility-change-consequences`.
- [ ] **Step 4:** Turn on Pages with Actions as the source: `gh api -X POST repos/enselcuk/patchdrop/pages -f build_type=workflow`. Push `main`, and confirm the `site` workflow succeeds.
- [ ] **Step 5:** Selçuk adds DNS `CNAME patchdrop → enselcuk.github.io`. Then run `gh api -X PUT repos/enselcuk/patchdrop/pages -f cname=patchdrop.enginsistem.com`. Once the certificate is issued, run `-F https_enforced=true`. Set the repo homepage: `gh repo edit --homepage https://patchdrop.enginsistem.com`.
- [ ] **Step 6:** Verify on the live site:
  - both languages and the TR redirect
  - every download link returns 200 after redirects
  - the live star count updates
  - preview cards in LinkedIn Post Inspector, the X post composer and the Facebook Sharing Debugger, for `/` and `/tr/`
  - a test visit with `?utm_source=test` shows up in GoatCounter, and the parameter disappears from the address bar
  - Lighthouse on the live URL
- [ ] **Step 7:** Give Selçuk the per-network post links (`/tr/?utm_source=linkedin`, `…=x`, `…=facebook`, `…=instagram` and the English equivalents).
