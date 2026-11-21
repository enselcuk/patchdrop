// Builds the 1200×630 social preview images (public/og-en.png, public/og-tr.png)
// from an SVG template plus the app screenshot. Run: node scripts/make-og.mjs
// The output PNGs are committed; rerun after changing the headline or screenshot.
import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import sharp from 'sharp'

const root = join(dirname(fileURLToPath(import.meta.url)), '..')

const W = 1200
const H = 630
// librsvg (inside sharp) can't load the self-hosted Manrope woff2 files, so the
// images use the system sans-serif (Helvetica Neue on macOS, where they're made).
const FONT = `Manrope, 'Helvetica Neue', Arial, sans-serif`

// Mirrors the hero headline in src/i18n/{en,tr}.ts, broken into lines that fit.
const copy = {
  en: {
    lines: [['Share files over', false], ['the office cable.', false], ['Just drag and drop.', true]],
    sub: 'Free & open source · Mac + Windows · Beta',
  },
  tr: {
    lines: [['Kablo üzerinden,', false], ['bir sürükle-bırak', true], ['uzaklıkta.', true]],
    sub: 'Ücretsiz ve açık kaynak · Mac + Windows · Beta',
  },
}

const esc = (s) => s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
const logoInner = readFileSync(join(root, 'src/assets/logo.svg'), 'utf8')
  .replace(/^[\s\S]*?<svg[^>]*>/, '')
  .replace(/<\/svg>\s*$/, '')
  .replace(/pd-g/g, 'og-logo-g')

// Screenshot panel: bleeds off the right and bottom edges.
const SHOT = { x: 640, y: 128, w: W - 640, h: H - 128, scale: 760 }

function svg(lang) {
  const c = copy[lang]
  const lineH = 70
  const top = 252
  const text = c.lines
    .map(([s, accent], i) =>
      `<text x="72" y="${top + i * lineH}" font-size="60" font-weight="800" letter-spacing="-1.8" fill="${accent ? 'url(#og-accent)' : '#0B1B33'}">${esc(s)}</text>`)
    .join('')
  return `<svg xmlns="http://www.w3.org/2000/svg" width="${W}" height="${H}" viewBox="0 0 ${W} ${H}">
  <defs>
    <linearGradient id="og-bg" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#FFFFFF"/><stop offset="1" stop-color="#EEF4FF"/></linearGradient>
    <linearGradient id="og-accent" x1="0" y1="0" x2="1" y2="0"><stop offset="0" stop-color="#2F6BFF"/><stop offset="1" stop-color="#7A5CFF"/></linearGradient>
    <radialGradient id="og-glow" cx="0.85" cy="0.1" r="0.7"><stop offset="0" stop-color="#7A5CFF" stop-opacity="0.16"/><stop offset="1" stop-color="#7A5CFF" stop-opacity="0"/></radialGradient>
    <filter id="og-shadow" x="-20%" y="-20%" width="140%" height="140%"><feGaussianBlur stdDeviation="18"/></filter>
  </defs>
  <rect width="${W}" height="${H}" fill="url(#og-bg)"/>
  <rect width="${W}" height="${H}" fill="url(#og-glow)"/>
  <g transform="translate(72 64) scale(0.875)">${logoInner}</g>
  <text x="140" y="104" font-family="${FONT}" font-size="34" font-weight="800" letter-spacing="-0.8" fill="#0B1B33">PatchDrop</text>
  <g font-family="${FONT}">${text}</g>
  <text x="72" y="${top + c.lines.length * lineH + 6}" font-family="${FONT}" font-size="24" font-weight="600" fill="#44516B">${esc(c.sub)}</text>
  <text x="72" y="574" font-family="${FONT}" font-size="22" font-weight="700" fill="#2459E6">patchdrop.enginsistem.com</text>
  <rect x="${SHOT.x}" y="${SHOT.y + 14}" width="${SHOT.w + 40}" height="${SHOT.h + 40}" rx="22" fill="#2F6BFF" opacity="0.28" filter="url(#og-shadow)"/>
  <rect x="${SHOT.x - 1}" y="${SHOT.y - 1}" width="${SHOT.w + 40}" height="${SHOT.h + 40}" rx="19" fill="#0B1B33" opacity="0.12"/>
</svg>`
}

async function shotPanel(lang) {
  const resized = await sharp(join(root, `src/assets/screenshot-${lang}.png`)).resize({ width: SHOT.scale }).toBuffer()
  const crop = await sharp(resized).extract({ left: 0, top: 0, width: SHOT.w, height: SHOT.h }).toBuffer()
  const mask = Buffer.from(`<svg xmlns="http://www.w3.org/2000/svg" width="${SHOT.w}" height="${SHOT.h}"><rect width="${SHOT.w + 40}" height="${SHOT.h + 40}" rx="18" fill="#fff"/></svg>`)
  return sharp(crop).composite([{ input: mask, blend: 'dest-in' }]).png().toBuffer()
}

for (const lang of ['en', 'tr']) {
  const out = join(root, `public/og-${lang}.png`)
  await sharp(Buffer.from(svg(lang)))
    .composite([{ input: await shotPanel(lang), left: SHOT.x, top: SHOT.y }])
    .png({ compressionLevel: 9, palette: false })
    .toFile(out)
  const meta = await sharp(out).metadata()
  console.log(`${out}: ${meta.width}×${meta.height}`)
}
