export function publicSiteUrl(
  development: boolean,
  configuredUrl?: string,
): string {
  if (configuredUrl?.trim()) {
    const url = new URL(configuredUrl.trim());
    if (
      !['http:', 'https:'].includes(url.protocol) ||
      url.username ||
      url.password ||
      url.search ||
      url.hash
    ) {
      throw new Error(
        'Public site URL must be an HTTP(S) URL without credentials, query or fragment',
      );
    }
    return url.href.replace(/\/$/, '');
  }
  return development ? 'http://localhost:3000' : 'https://outpipe.dev';
}
