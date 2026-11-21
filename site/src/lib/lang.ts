import en from '../i18n/en'
import tr from '../i18n/tr'
import type { Dict, Lang } from '../i18n/types'

export const LANG_KEY = 'patchdrop-lang'

export function getDict(lang: Lang): Dict {
  return lang === 'tr' ? tr : en
}

/** Saved choice wins; otherwise only the browser's FIRST language counts. */
export function pickLanguage(saved: string | null, browserLangs: readonly string[]): Lang {
  if (saved === 'en' || saved === 'tr') return saved
  const first = (browserLangs[0] ?? '').toLowerCase()
  return first === 'tr' || first.startsWith('tr-') ? 'tr' : 'en'
}

/**
 * Self-contained copy of the logic above, inlined into <head> of `/` so the
 * redirect happens before first paint. Expects `localStorage`, `navigator`
 * and `location` in scope. Kept in sync by tests/lang.test.ts.
 */
export const REDIRECT_SCRIPT = `try{var s=null;try{s=localStorage.getItem(${JSON.stringify(LANG_KEY)})}catch(e){}var t;if(s==='en'||s==='tr'){t=s}else{var l=navigator.languages&&navigator.languages.length?navigator.languages:[navigator.language||''];var f=String(l[0]||'').toLowerCase();t=f==='tr'||f.indexOf('tr-')===0?'tr':'en'}if(t==='tr'){location.replace('/tr/'+location.search+location.hash)}}catch(e){}`
