// @ts-check
import { defineConfig } from 'astro/config'
import sitemap from '@astrojs/sitemap'

export default defineConfig({
  site: 'https://patchdrop.enginsistem.com',
  trailingSlash: 'always',
  i18n: {
    defaultLocale: 'en',
    locales: ['en', 'tr'],
    routing: { prefixDefaultLocale: false },
  },
  integrations: [
    sitemap({ i18n: { defaultLocale: 'en', locales: { en: 'en', tr: 'tr' } } }),
  ],
})
