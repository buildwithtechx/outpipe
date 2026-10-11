import { describe, expect, it } from 'vitest';
import { cachedDashboardRedisRead } from '../src/lib/dashboard-redis-cache';

describe('dashboard-redis-cache', () => {
  it('coalesces concurrent reads for the same cache key', async () => {
    let calls = 0;
    const load = async () => {
      calls += 1;
      await new Promise((resolve) => setTimeout(resolve, 5));
      return { count: 42 };
    };

    const [first, second] = await Promise.all([
      cachedDashboardRedisRead('test-key-a', load),
      cachedDashboardRedisRead('test-key-a', load),
    ]);

    expect(calls).toBe(1);
    expect(first).toEqual({ count: 42 });
    expect(second).toEqual({ count: 42 });
  });

  it('keeps distinct entries for different keys', async () => {
    let calls = 0;
    const load = async (val: string) => {
      calls += 1;
      return { val };
    };

    const first = await cachedDashboardRedisRead('key-1', () => load('one'));
    const second = await cachedDashboardRedisRead('key-2', () => load('two'));

    expect(calls).toBe(2);
    expect(first.val).toBe('one');
    expect(second.val).toBe('two');
  });
});
