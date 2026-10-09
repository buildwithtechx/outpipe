import { env } from '#/env';

export function docsUrl(path = '') {
  const base = (env.VITE_OUTPIPE_DOCS_URL ?? 'http://localhost:4323').replace(
    /\/+$/,
    '',
  );
  const suffix = path.replace(/^\/+/, '');
  return suffix ? `${base}/${suffix}` : `${base}/`;
}
