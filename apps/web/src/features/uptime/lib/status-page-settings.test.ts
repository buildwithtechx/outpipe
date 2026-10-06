import { expect, it } from 'vitest';
import {
  publicStatusUrl,
  validStatusDomain,
  validStatusSlug,
} from './status-page-settings';

it('validates slugs without accepting traversal or broken links', () => {
  for (const slug of ['service', 'service-2', 'a', 'a'.repeat(128)])
    expect(validStatusSlug(slug)).toBe(true);
  for (const slug of [
    '',
    '-service',
    'service-',
    '../private',
    'a/b',
    'a b',
    'UPPER',
    'a'.repeat(129),
  ])
    expect(validStatusSlug(slug)).toBe(false);
});

it('accepts optional hostnames while rejecting URLs and credentials', () => {
  for (const domain of ['', 'status.example.com', 'STATUS.EXAMPLE.COM'])
    expect(validStatusDomain(domain)).toBe(true);
  for (const domain of [
    'https://status.example.com',
    'status.example.com/path',
    'user@status.example.com',
    '-status.example.com',
    'status.example.com:443',
  ])
    expect(validStatusDomain(domain)).toBe(false);
});

it('builds local and hosted URLs while preserving an existing base path', () => {
  expect(publicStatusUrl('my-service', 'http://localhost:4322/')).toBe(
    'http://localhost:4322/my-service',
  );
  expect(publicStatusUrl('service', 'https://status.example.com/pages/')).toBe(
    'https://status.example.com/pages/service',
  );
  expect(publicStatusUrl('a/b', 'https://status.example.com')).toBe(
    'https://status.example.com/a%2Fb',
  );
});

it.each([
  'javascript:alert(1)',
  'https://user:secret@example.com',
  'https://example.com?next=evil',
  'https://example.com#private',
])('rejects unsafe status site configuration %s', (base) => {
  expect(() => publicStatusUrl('service', base)).toThrow();
});
