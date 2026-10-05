import { afterEach, expect, it, vi } from 'vitest';
import {
  fetchStatusData,
  fetchStatusResult,
  mapStatusResponse,
  subscribeToStatus,
} from './api';

afterEach(() => vi.unstubAllGlobals());

it('distinguishes unpublished pages from failures that can be retried', async () => {
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: false, status: 404 }));
  expect(await fetchStatusResult('missing')).toEqual({ state: 'not-found' });
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: false, status: 503 }));
  expect(await fetchStatusResult('team')).toEqual({ state: 'unavailable' });
});

it('maps the Go status response without fabricating history or incidents', () => {
  const data = mapStatusResponse({
    overallStatus: 'partial_outage',
    page: { title: 'API Status', description: 'Services' },
    monitors: [
      {
        id: 'one',
        name: 'API',
        protocol: 'http',
        url: 'https://example.com',
        status: 'down',
        uptimeRatio: 72,
        latencyMs: 90,
      },
    ],
    activeIncidents: [
      {
        id: 'incident',
        title: 'Outage',
        status: 'investigating',
        severity: 'critical',
        startedAt: '2026-10-05T00:00:00Z',
        updates: [],
      },
    ],
    pastIncidents: null,
  });
  expect(data.systemStatus).toBe('outage');
  expect(data.monitors[0]).toMatchObject({
    status: 'down',
    type: 'http',
    target: 'https://example.com',
    uptime90Days: 72,
    currentLatencyMs: 90,
    history: [],
  });
  expect(data.activeIncidents[0]?.startedAt).toBe('2026-10-05T00:00:00Z');
  expect(data.pastIncidents).toEqual([]);
});

it('does not display demo operational status for a private or unavailable page', async () => {
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: false, status: 404 }));
  expect(await fetchStatusData('private')).toBeNull();
});

it('subscribes email using the server contract and refuses webhooks', async () => {
  const fetch = vi.fn().mockResolvedValue({ ok: true });
  vi.stubGlobal('fetch', fetch);
  expect(
    await subscribeToStatus('team', 'webhook', 'https://example.com'),
  ).toBe(false);
  expect(fetch).not.toHaveBeenCalled();
  expect(await subscribeToStatus('team', 'email', 'user@example.com')).toBe(
    true,
  );
  expect(fetch).toHaveBeenCalledWith(
    expect.stringContaining('/status/team/subscribe'),
    expect.objectContaining({
      method: 'POST',
      body: JSON.stringify({ email: 'user@example.com' }),
    }),
  );
});
