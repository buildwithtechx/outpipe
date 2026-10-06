export function validStatusSlug(slug: string) {
  return /^[a-z0-9](?:[a-z0-9-]{0,126}[a-z0-9])?$/.test(slug);
}

export function validStatusDomain(domain: string) {
  return (
    !domain ||
    /^(?=.{4,253}$)(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,63}$/i.test(
      domain,
    )
  );
}

export function publicStatusUrl(
  slug: string,
  base: string,
  domain = '',
): string | null {
  if (!validStatusSlug(slug) || !validStatusDomain(domain)) return null;
  try {
    const url = new URL(domain ? `https://${domain}` : base);
    if (
      !['http:', 'https:'].includes(url.protocol) ||
      url.username ||
      url.password ||
      (!domain && /[?#]/.test(base))
    )
      return null;
    return `${url.href.replace(/\/$/, '')}/${encodeURIComponent(slug)}`;
  } catch {
    return null;
  }
}
