import { env } from '#/env';

export const siteUrl = env.VITE_OUTPIPE_SITE_URL ?? 'https://outpipe.dev';
export const siteName = 'Outpipe';
export const siteDescription =
  'Secure public access for local and private services with one CLI, desktop app, and developer protocol.';
export const socialImage = `${siteUrl}/og-image.png`;

interface SeoOptions {
  title: string;
  description?: string;
  path?: string;
}

export function createSeo({
  title,
  description = siteDescription,
  path = '/',
}: SeoOptions) {
  const canonical = new URL(path, siteUrl).toString();

  return {
    meta: [
      { title },
      { name: 'description', content: description },
      { property: 'og:type', content: 'website' },
      { property: 'og:site_name', content: siteName },
      { property: 'og:title', content: title },
      { property: 'og:description', content: description },
      { property: 'og:url', content: canonical },
      { property: 'og:image', content: socialImage },
      { property: 'og:image:alt', content: `${siteName} social preview` },
      { property: 'og:image:width', content: '1200' },
      { property: 'og:image:height', content: '630' },
      { property: 'og:image:type', content: 'image/png' },
      { name: 'twitter:card', content: 'summary_large_image' },
      { name: 'twitter:url', content: canonical },
      { name: 'twitter:title', content: title },
      { name: 'twitter:description', content: description },
      { name: 'twitter:image', content: socialImage },
    ],
    links: [{ rel: 'canonical', href: canonical }],
  };
}
