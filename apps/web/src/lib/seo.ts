import { env } from '#/env';

export const siteUrl = env.VITE_OUTPIPE_SITE_URL ?? 'https://outpipe.dev';
export const siteName = 'Outpipe';
export const siteDescription =
  'Secure public access for local and private services with one CLI, desktop app, and developer protocol.';
export const socialImage = `${siteUrl}/og-image.png`;

export interface SeoOptions {
  title: string;
  description?: string;
  path?: string;
  image?: string;
  imageAlt?: string;
  type?: 'website' | 'article';
}

export function createSeo({
  title,
  description = siteDescription,
  path = '/',
  image,
  imageAlt,
  type = 'website',
}: SeoOptions) {
  const canonical = new URL(path, siteUrl).toString();
  const formattedTitle = title.includes(siteName)
    ? title
    : `${title} — ${siteName}`;
  const resolvedImage = image
    ? image.startsWith('http')
      ? image
      : new URL(image, siteUrl).toString()
    : socialImage;
  const resolvedImageAlt = imageAlt ?? `${formattedTitle} social preview`;

  return {
    meta: [
      { title: formattedTitle },
      { name: 'description', content: description },
      { property: 'og:type', content: type },
      { property: 'og:locale', content: 'en_US' },
      { property: 'og:site_name', content: siteName },
      { property: 'og:title', content: formattedTitle },
      { property: 'og:description', content: description },
      { property: 'og:url', content: canonical },
      { property: 'og:image', content: resolvedImage },
      { property: 'og:image:secure_url', content: resolvedImage },
      { property: 'og:image:alt', content: resolvedImageAlt },
      { property: 'og:image:width', content: '1200' },
      { property: 'og:image:height', content: '630' },
      { property: 'og:image:type', content: 'image/png' },
      { name: 'twitter:card', content: 'summary_large_image' },
      { name: 'twitter:url', content: canonical },
      { name: 'twitter:title', content: formattedTitle },
      { name: 'twitter:description', content: description },
      { name: 'twitter:image', content: resolvedImage },
      { name: 'twitter:image:alt', content: resolvedImageAlt },
    ],
    links: [{ rel: 'canonical', href: canonical }],
  };
}
