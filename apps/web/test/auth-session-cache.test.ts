import { describe, expect, it } from 'vitest';
import { getCachedAuthSession } from '../src/lib/auth-session-cache';

function request(cookie: string): Request {
  return new Request('https://outpipe.test/api/team/stats/overview', {
    headers: { cookie },
  });
}

describe('auth-session-cache', () => {
  it('coalesces concurrent session reads for the same cookie', async () => {
    const previousTtl = process.env.AUTH_SESSION_CACHE_TTL_MS;
    process.env.AUTH_SESSION_CACHE_TTL_MS = '1000';
    let calls = 0;
    try {
      const load = async () => {
        calls += 1;
        await new Promise((resolve) => setTimeout(resolve, 5));
        return { user: { id: 'user-a' } };
      };
      const [first, second] = await Promise.all([
        getCachedAuthSession(request('session-a'), load),
        getCachedAuthSession(request('session-a'), load),
      ]);
      expect(calls).toBe(1);
      expect(first).toEqual(second);
    } finally {
      if (previousTtl === undefined)
        delete process.env.AUTH_SESSION_CACHE_TTL_MS;
      else process.env.AUTH_SESSION_CACHE_TTL_MS = previousTtl;
    }
  });

  it('does not share sessions between different cookies', async () => {
    const previousTtl = process.env.AUTH_SESSION_CACHE_TTL_MS;
    process.env.AUTH_SESSION_CACHE_TTL_MS = '1000';
    let calls = 0;
    try {
      const load = async (req: Request) => {
        calls += 1;
        return { user: { id: req.headers.get('cookie') || 'unknown' } };
      };
      const first = await getCachedAuthSession(request('session-b'), load);
      const second = await getCachedAuthSession(request('session-c'), load);
      expect(calls).toBe(2);
      expect(first?.user.id).toBe('session-b');
      expect(second?.user.id).toBe('session-c');
    } finally {
      if (previousTtl === undefined)
        delete process.env.AUTH_SESSION_CACHE_TTL_MS;
      else process.env.AUTH_SESSION_CACHE_TTL_MS = previousTtl;
    }
  });
});
