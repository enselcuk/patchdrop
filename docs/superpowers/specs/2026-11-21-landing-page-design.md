# PatchDrop landing page: design

Date: 2026-11-21
Status: approved 2026-11-21

## Intent

PatchDrop is a free, open-source (MIT) desktop app for sharing files between
Mac and Windows computers on a **wired** office network. It was built because
LocalSend-style apps often fail to find each other on Ethernet.

Selçuk Engin will announce it to his followers on LinkedIn, X and Facebook. The
landing page is where those posts land. It exists to get reach and traction,
and through that to get Selçuk and his brand **Engin Sistem** (enginsistem.com)
noticed. Long term, he hopes the project grows into a well-known reference with
many GitHub stars and downloads, and maybe a community.

A visitor's journey, in priority order (from Selçuk):

1. **Try the app**: download it, use it and appreciate it.
2. **Star it** on GitHub.
3. **Share it** with friends and colleagues.
4. **Get curious who made it** and discover Engin Sistem.

Side benefit: a backlink from the page and the GitHub repo to enginsistem.com.

Positioning (from Selçuk): PatchDrop is *not* a LocalSend competitor. It is a
handy tool that covers a case LocalSend doesn't handle well: wired networks.

## Goals

- A visitor arriving from a social post understands within seconds what
  PatchDrop does and can download the right file for their computer.
- Turkish browsers get Turkish content; everyone else gets English.
- Social posts show a proper preview card in the post's language.
- Starring and sharing are always one click away.
- A clear, friendly **beta** notice: tested only by a few colleagues in
  Selçuk's office on Mac and Windows.
- Better than <https://localsend.org/tr> *for PatchDrop's situation*: sharper
  message, faster page, a clear download path. It can't compete on LocalSend's
  numbers or press, and shouldn't try.
- Hosted free on GitHub Pages at **patchdrop.enginsistem.com**.

## Non-goals

- No Discord, forum, Reddit or other community. **GitHub is the only
  community** (Issues, stars, the repo).
- No blog, docs site, changelog page, newsletter or sign-up.
- No mobile app. The page explains that PatchDrop is for desktop computers.
- No App Store presence for now.
- No languages beyond Turkish and English for now. The structure should make a
  third language easy to add later.

## Anti-goals (failure even if it "works")

- Reads like an advert for Engin Sistem. The brand section comes after the
  visitor has gotten value, and reads as a story.
- Overclaims or attacks LocalSend, e.g. "LocalSend doesn't work on cable". The
  accurate claim is that Wi-Fi-style discovery *often fails to find devices* on
  wired networks. Credit LocalSend respectfully.
- Broken or stale download links.
- A slow or heavy page, or a cookie banner, on first contact from a phone.
- Hiding the beta status.

## Constraints

- **The repo must be public** before launch: free GitHub Pages needs a public
  repo, and visitors can't download release files from a private repo. Selçuk
  has agreed to make it public. Commit history shows the author email
  `info@engintasarim.com`, which he has confirmed is fine.
- Company and brand naming: the legal company is **Engin Tasarım Ltd. Şti.**
  It is the **copyright holder** (decided by Selçuk): the footer, `LICENSE`
  ("Copyright (c) 2026 Engin Tasarım Ltd. Şti.") and the app metadata in
  `wails.json` all name the company, changed before the repo goes public.
  Note: the Windows installer tool (NSIS) failed on a non-ASCII copyright
  string before, so the installer's copyright text must either be stored in a
  form NSIS accepts or use the ASCII spelling "Engin Tasarim Ltd. Sti.". The
  rebuilt installer must be checked. The brand being promoted is **Engin Sistem**
  (enginsistem.com).
- Custom domain `patchdrop.enginsistem.com`. Selçuk will add the DNS record:
  `CNAME patchdrop → enselcuk.github.io`.
- The apps are not signed or notarized. On macOS 15 and later, right-click →
  Open no longer works for such apps: the first launch is blocked, and the
  user must go to **System Settings → Privacy & Security → Open Anyway**. The
  page explains these steps (short, with one small illustration). Windows
  SmartScreen shows "Windows protected your PC" → More info → Run anyway.
  The README's install notes are corrected the same way.
- PatchDrop ports: UDP 47820 (discovery), TCP 47821 (transfers).
- Current release: **v0.1.0**, pre-release, at
  `github.com/enselcuk/patchdrop/releases`.

## Approaches considered

- **A · Astro, one page per language (chosen).** English at `/`, Turkish at
  `/tr/`, both built from one layout and two text files. Each URL has its own
  title, description and share-preview image, and the two are linked as
  translations for search engines. A GitHub Actions workflow builds and
  publishes on push. Trade-off: one more tool and a build step.
- **B · One hand-written page that swaps text in the browser.** No build, but
  social previews and Google see only one language, so Turkish posts would
  show an English card. Rejected: works against the launch, which runs on
  shares.
- **C · Two hand-written HTML pages.** Same result as A without a tool, but
  every change is made twice and the pages drift apart. Rejected for upkeep.

Selçuk picked **A**.

## Decided with Selçuk

- **Address:** `patchdrop.enginsistem.com`.
- **Visual direction: C, "bright consumer product".** Light, friendly,
  centered hero. White to very light blue background (`#FFFFFF` → `#EEF4FF`),
  blue-to-violet accent gradient (`#2F6BFF` → `#7A5CFF`), dark navy text
  (`#0B1B33`), pill-shaped buttons, an amber "Beta" pill. Font: Manrope.
  The brand link to Engin Sistem comes through the "Made by" section, not the
  colors. Mockup: `.superpowers/brainstorm/…/content/visual-direction.html`.
- **Page sections, in order** (mockup: `…/content/page-layout.html`):
  0. **Beta bar** (top, dismissible): "Beta: tested in our office on Mac and
     Windows. Something not working? Tell us on GitHub" → GitHub Issues.
  1. **Hero:** headline, one-sentence description, a big download button for
     the visitor's detected OS, "Star on GitHub" with the live count, a small
     line "Also for Windows/macOS · Free & open source · v0.1.0 beta", and a
     real screenshot of the app.
  2. **Why PatchDrop:** the wired-network story. Wi-Fi-style apps often can't
     find each other on cabled office networks; PatchDrop announces itself on
     every wired adapter. Credits LocalSend respectfully.
  3. **How it works:** 3 steps: install on both computers → drop files on the
     window → pick the other computer and Send.
  4. **Features:** 6 one-line tiles: full cable speed; encrypted, verified
     devices; no internet or cloud; whole folders; Mac + Windows; free, MIT,
     no ads.
  5. **Download:** macOS and Windows side by side, with file type and size,
     first-launch tips (macOS: Privacy & Security → Open Anyway), and links to release notes and checksums. Windows
     offers the setup installer first and the portable zip second.
  6. **Liked it? Share it:** LinkedIn, X, Facebook, WhatsApp, copy link, plus
     another Star button. A ready-made message in the page's language.
  7. **FAQ:** is it free; does it need internet; is it safe; devices don't
     see each other, what now (firewall, Ethernet set to Private, Add by IP,
     Scan network); why not just use LocalSend; where do files go.
  8. **Made by:** "Built by Selçuk Engin · Engin Sistem", Selçuk's photo, a
     short bio, and two groups of links (see "Made by links" below).
  9. **Footer:** GitHub, report a bug, releases, MIT License, TR/EN switch,
     "© 2026 Engin Tasarım Ltd. Şti."
- **Languages:** Turkish for Turkish browsers, English for everyone else.

## Design details (Claude's calls, open to correction)

### Site structure and hosting

- Lives in a new `site/` folder in the PatchDrop repo, built with **Astro**
  (current 7.x) as a fully static site. No server code.
- Deployed by a GitHub Actions workflow (`.github/workflows/site.yml`) using
  the official Pages actions. It runs on pushes to `main` that touch `site/`,
  on manual dispatch, and when a release is published (so the page's fallback
  version and links stay current).
- `site/public/CNAME` contains `patchdrop.enginsistem.com`. Pages set to
  "GitHub Actions" as source, HTTPS enforced.
- The repo's homepage field and README link to the site. The site links to the
  repo. That gives the backlink in both directions.

### Languages

- `/` is English, `/tr/` is Turkish. All text lives in two dictionaries
  (`site/src/i18n/en.ts`, `tr.ts`) with the same keys, used by one shared
  layout. A missing key fails the build.
- Each page sets `<html lang>`, its own title, meta description, canonical URL,
  and `hreflang` links to both versions plus `x-default` → `/`.
- **Automatic choice:** a tiny inline script at the top of `/`. If the visitor
  hasn't picked a language before, and their browser's first preferred
  language starts with `tr`, it redirects to `/tr/` (keeping any `#anchor`).
  `/tr/` never redirects away. Crawlers aren't redirected, because the script
  only acts on a real browser language.
- **TR/EN switch** in the top-right and footer. Choosing saves the preference
  in `localStorage`, which overrides detection on later visits. No cookies.

### Downloads

- **Stable asset names** so links never go stale:
  `https://github.com/enselcuk/patchdrop/releases/latest/download/<name>`.
  Names:
  - `PatchDrop-macOS.dmg`
  - `PatchDrop-macOS-universal.zip`
  - `PatchDrop-Windows-Setup-x64.exe` (renamed from
    `PatchDrop-Setup-0.1.0-amd64.exe`)
  - `PatchDrop-Windows-x64.zip`
  - `SHA256SUMS.txt`

  The Windows installer's `OutFile` in `build/windows/installer/project.nsi`
  changes to match, and the v0.1.0 release asset is re-uploaded under the new
  name.
- **"Latest" and pre-releases:** GitHub's `/releases/latest` skips
  pre-releases, and v0.1.0 is marked pre-release. So at launch, v0.1.0 is
  marked as a normal release. The beta status is carried by the version label
  and the page's beta bar instead.
- **OS detection** for the hero button: macOS → DMG; Windows → setup .exe;
  anything else on desktop (e.g. Linux) → both buttons, plus a note that only
  Mac and Windows are supported so far.
- **Phones and tablets** (most social clicks): no download button, because the
  app can't run there. The hero shows "PatchDrop runs on Mac and Windows. Open
  this page on your computer" with **Send link to my computer** (the phone's
  share sheet, or copy link as a fallback) and the Star button. The rest of the
  page is unchanged, so they can still read, star and share.

### Live numbers

- At build time, the workflow fetches the latest release (version, asset
  sizes) and the star count from the GitHub API, and bakes them into the page
  as the default values.
- In the browser, the page refreshes the star count and the total download
  count from the public GitHub API (no token). If the API fails or is rate
  limited, the built-in values stay.
- The download count is shown only once it reaches 100, so the page doesn't
  advertise a tiny number at launch. Stars are always shown.

### Sharing

- Share links use each network's plain share URL (no SDKs, no tracking
  scripts): LinkedIn, X, Facebook, WhatsApp, plus copy link. On phones, a
  single **Share** button uses the phone's native share sheet when available.
- The shared URL is the page in its current language. The pre-filled text is
  in that language too, e.g. TR: "Kablolu ofis ağında Mac ve Windows arasında
  dosya paylaşmak için ücretsiz ve açık kaynak bir uygulama: PatchDrop".
- **Preview cards:** one 1200×630 Open Graph image per language, in the page
  style (headline + app screenshot + PatchDrop logo). Also `twitter:card =
  summary_large_image`.

### Analytics (GoatCounter)

- Selçuk creates the GoatCounter account himself (site code, e.g.
  `patchdrop` → `patchdrop.goatcounter.com`) and checks that its plan fits
  this use. The site code is a single config value in the site.
- The count script loads async, after the page renders, from GoatCounter. It
  sets no cookies and stores nothing in the browser, so no banner is needed.
- Page views are counted per path, so `/` and `/tr/` show EN vs TR.
- Events: download clicks (per OS and file), Star clicks, share clicks (per
  network), language switches, and "Send link to my computer" on phones.
- Social posts use per-network links, e.g. `?utm_source=linkedin`, which
  GoatCounter shows as the campaign/referrer. The page removes the parameter
  from the address bar after counting, so it isn't copied into re-shares.
- The FAQ or footer states plainly: "Anonymous visit counts with GoatCounter.
  No cookies, no personal data."

### Look and assets

- Manrope, self-hosted (no Google Fonts request).
- Real app screenshots, taken from the running app in light mode with
  realistic device names (e.g. "Tasarım-PC", "Muhasebe-Mac" on the Turkish
  page; "Design-PC", "Accounts-Mac" on the English one).
- A simple PatchDrop logo mark and favicon, matching the accent gradient. It
  can be replaced later if Engin Sistem designs a proper one.
- Responsive from 360 px phones up. Supports `prefers-reduced-motion`. Uses
  light mode only, matching direction C.

### Quality bar

- Lighthouse on mobile: Performance, Accessibility, Best Practices and SEO all
  ≥ 95.
- Total page weight under ~500 KB on first load, excluding the screenshot,
  which is lazy-loaded below the fold where possible.
- Works with JavaScript off: all text and download links are present; only
  OS detection, live counts and the native share button need JS.
- Valid `hreflang`, `sitemap.xml`, `robots.txt`, and a 404 page in both
  languages.
- Analytics with **GoatCounter** (decided by Selçuk): cookieless, no consent
  banner, one small async script. See "Analytics" above.

## Made by links

- **Photo:** Selçuk's profile photo, from
  `https://pbs.twimg.com/profile_images/2055608326021382144/ZDvFBKxv_400x400.jpg`.
  Downloaded once and committed to the site (as an optimized WebP/JPEG, about
  200×200 shown), not hot-linked from X.
- **Selçuk:** LinkedIn `https://www.linkedin.com/in/selcuk-engin/`, X
  `https://x.com/enselcuk`, YouTube `https://www.youtube.com/@selcukengin`.
- **Engin Sistem:** website `https://enginsistem.com` (main button), X
  `https://x.com/enginsistem`, Instagram `https://www.instagram.com/enginsistem/`,
  YouTube `https://www.youtube.com/@enginsistem`.
- No Facebook profile link, since none was given. Facebook stays a share
  target.
- Outbound links open in the same tab and carry no `nofollow`, so the
  enginsistem.com link counts as a real backlink. Links to enginsistem.com add
  `?utm_source=patchdrop` so Selçuk can see that traffic in his own analytics.
- X share links add `via=enselcuk`, so shares mention Selçuk.

## Content notes

- Headline (TR, from the mockup): "Kablo üzerinden, bir sürükle-bırak
  uzaklıkta." The English headline is written natively, not translated word
  for word, e.g. "Share files over the office cable. Just drag and drop."
- Beta copy reflects what Selçuk said: tested by a few colleagues in his
  office, on Mac and Windows, not on every device or OS.
- Bio (proposal, to be confirmed): Selçuk Engin, founder of Engin Sistem, an
  Istanbul software company building B2B and custom business systems since
  2003. PatchDrop started as a fix for his own office's wired network.
- The README's "not yet tested between two physical machines" lines are updated
  to match the beta wording.

## Who does what

- Claude writes both the English and the Turkish copy. Selçuk reviews all text,
  especially the Turkish, during the local review (launch step 1).
- Claude runs the GitHub changes with `gh`, asking Selçuk right before each one
  that changes public state (repo visibility, release flags, asset rename,
  Pages, custom domain, repo homepage). Selçuk adds the DNS record himself.

## Launch steps (order)

1. Build the site and review it locally with Selçuk.
2. Rename the Windows installer asset; mark v0.1.0 as a normal release.
3. Make the repo public; enable Pages (GitHub Actions source).
4. Selçuk adds the DNS record; set the custom domain; wait for HTTPS.
5. Check both languages, downloads, share previews (LinkedIn Post Inspector,
   X and Facebook debuggers), and Lighthouse on the live URL.
6. Selçuk posts.

## Left to the builder

- Exact component structure inside Astro, CSS approach (plain CSS or scoped
  Astro styles; no CSS framework needed), and icon set (inline SVG).
- Exact wording of English copy and FAQ answers, following the content notes.
- How the OG images are produced (hand-made SVG→PNG at build or committed
  PNGs).
