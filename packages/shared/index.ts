export function publicSiteUrl(development: boolean): string {
  return development ? 'http://localhost:3000' : 'https://outpipe.dev';
}
