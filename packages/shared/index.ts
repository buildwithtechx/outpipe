export function publicSiteUrl(
  development: boolean,
  configuredUrl?: string,
): string {
  if (configuredUrl?.trim()) {
    const value = configuredUrl.trim();
    const message =
      'Public site URL must be an HTTP(S) URL without credentials, query or fragment';
    let url: URL;
    try {
      url = new URL(value);
    } catch {
      throw new Error(message);
    }
    if (
      !['http:', 'https:'].includes(url.protocol) ||
      url.username ||
      url.password ||
      /[?#]/.test(value)
    ) {
      throw new Error(message);
    }
    return url.href.replace(/\/$/, '');
  }
  return development ? 'http://localhost:3000' : 'https://outpipe.dev';
}
