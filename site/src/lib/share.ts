import { getDict } from './lang'
import type { Lang } from '../i18n/types'

export function shareText(lang: Lang): string {
  return getDict(lang).share.text
}

const enc = encodeURIComponent

export function shareLinks(url: string, lang: Lang): { linkedin: string; x: string; facebook: string; whatsapp: string } {
  const text = shareText(lang)
  return {
    linkedin: `https://www.linkedin.com/sharing/share-offsite/?url=${enc(url)}`,
    x: `https://x.com/intent/post?text=${enc(text)}&url=${enc(url)}&via=enselcuk`,
    facebook: `https://www.facebook.com/sharer/sharer.php?u=${enc(url)}`,
    whatsapp: `https://wa.me/?text=${enc(`${text} ${url}`)}`,
  }
}

// Copies the link; when the clipboard is unavailable or refuses, selects the
// visible read-only field so the user can copy by hand. Never throws.
export async function copyOrSelect(url: string, input: HTMLInputElement, clipboard?: Clipboard): Promise<'copied' | 'selected'> {
  try {
    if (clipboard && typeof clipboard.writeText === 'function') {
      await clipboard.writeText(url)
      return 'copied'
    }
  } catch {
    // fall through to selection
  }
  try {
    input.focus()
    input.select()
    input.setSelectionRange(0, input.value.length)
  } catch {
    // selection is best-effort
  }
  return 'selected'
}
