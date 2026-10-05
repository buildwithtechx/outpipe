import { describe, expect, it } from 'vitest';
import { publicSiteUrl } from './index';

describe('publicSiteUrl', () => {
  it('defaults to the hosted dashboard or local development dashboard', () => {
    expect(publicSiteUrl(false)).toBe('https://outpipe.dev');
    expect(publicSiteUrl(true)).toBe('http://localhost:3000');
    expect(publicSiteUrl(false, ' ')).toBe('https://outpipe.dev');
  });

  it('supports alternate ports, remote development and deployment paths', () => {
    expect(publicSiteUrl(true, 'http://localhost:3001/')).toBe(
      'http://localhost:3001',
    );
    expect(publicSiteUrl(true, ' https://preview.outpipe.dev/ ')).toBe(
      'https://preview.outpipe.dev',
    );
    expect(publicSiteUrl(false, 'https://example.com/dashboard/')).toBe(
      'https://example.com/dashboard',
    );
  });

  it.each([
    'javascript:alert(1)',
    'https://user:password@example.com',
    'https://example.com?token=secret',
    'https://example.com#fragment',
    'https://example.com?',
    'https://example.com#',
    'outpipe.dev',
    '//example.com',
    'invalid',
  ])('rejects unsafe or ambiguous configured links: %s', (url) => {
    expect(() => publicSiteUrl(false, url)).toThrow(
      'Public site URL must be an HTTP(S) URL without credentials, query or fragment',
    );
  });
});
