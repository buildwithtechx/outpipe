import { expect, it } from 'vitest';
import { redactJsonValue } from '../utils/http-payload-capture';

it('redacts camelCase and separator variants of credential keys', () => {
  const result = redactJsonValue({
    refreshToken: 'one',
    api_key: 'two',
    accessToken: 'three',
    privateKey: 'four',
    clientSecret: 'five',
    safe: 'public',
  });
  expect(result).toEqual({
    refreshToken: '[REDACTED]',
    api_key: '[REDACTED]',
    accessToken: '[REDACTED]',
    privateKey: '[REDACTED]',
    clientSecret: '[REDACTED]',
    safe: 'public',
  });
});

it('redacts credentials and replaces objects beyond the traversal limit', () => {
  let nested: unknown = { password: 'deep-credential', token: 'deep-token' };
  for (let depth = 0; depth < 20; depth++) nested = { nested: [nested] };
  const result = JSON.stringify(
    redactJsonValue({ password: 'credential', nested }),
  );
  expect(result).not.toContain('credential');
  expect(result).not.toContain('deep-token');
  expect(result).toContain('[REDACTED]');
});
