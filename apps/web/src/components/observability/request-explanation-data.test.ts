import { describe, expect, it } from 'vitest';
import type { RequestCapture } from '#/interfaces';
import {
  buildRequestExplanationPreview,
  captureToSummary,
  requestExplanationDemoScenarios,
} from './request-explanation-data';

describe('request-explanation-data', () => {
  it('classifies 5xx as server-error with hypothesis and checks', () => {
    const preview = buildRequestExplanationPreview({
      id: 'req-500',
      timestamp: '2026-10-11T09:00:00Z',
      method: 'POST',
      path: '/api/tunnels',
      statusCode: 502,
      duration: 350,
      traceId: 'tr-abc',
    });

    expect(preview.category).toBe('server-error');
    expect(preview.hypothesis).not.toBeNull();
    expect(preview.hypothesis?.title).toContain('server-side failure');
    expect(preview.nextChecks.length).toBeGreaterThan(0);
    expect(preview.evidence.some((e) => e.value === 'HTTP 502')).toBe(true);
  });

  it('classifies 4xx as client-error', () => {
    const preview = buildRequestExplanationPreview({
      id: 'req-401',
      timestamp: '2026-10-11T09:00:00Z',
      method: 'GET',
      path: '/api/v1/secrets',
      statusCode: 401,
      duration: 45,
    });

    expect(preview.category).toBe('client-error');
    expect(preview.hypothesis?.title).toContain('request rejection');
    expect(
      preview.limitations.some((l) => l.includes('No trace identifier')),
    ).toBe(true);
  });

  it('classifies slow responses exceeding 1000ms', () => {
    const preview = buildRequestExplanationPreview({
      id: 'req-slow',
      timestamp: '2026-10-11T09:00:00Z',
      method: 'GET',
      path: '/api/v1/metrics',
      statusCode: 200,
      duration: 1540,
      traceId: 'tr-slow',
    });

    expect(preview.category).toBe('slow');
    expect(preview.hypothesis?.title).toContain('latency bottleneck');
  });

  it('classifies normal sub-second 200 as healthy with null hypothesis', () => {
    const preview = buildRequestExplanationPreview({
      id: 'req-ok',
      timestamp: '2026-10-11T09:00:00Z',
      method: 'GET',
      path: '/healthz',
      statusCode: 200,
      duration: 18,
    });

    expect(preview.category).toBe('healthy');
    expect(preview.hypothesis).toBeNull();
  });

  it('converts RequestCapture to HttpRequestSummary correctly', () => {
    const capture: RequestCapture = {
      id: 'cap-1',
      organizationId: 'org-1',
      tunnelId: 'tun-1',
      timestamp: '2026-10-11T09:00:00Z',
      method: 'PUT',
      path: '/v1/config',
      statusCode: 204,
      durationMs: 42,
      requestBodySize: 120,
      responseBodySize: 0,
      createdAt: '2026-10-11T09:00:00Z',
    };

    const summary = captureToSummary(capture);
    expect(summary.id).toBe('cap-1');
    expect(summary.duration).toBe(42);
    expect(summary.statusCode).toBe(204);
  });

  it('handles demo scenarios without error', () => {
    for (const scenario of requestExplanationDemoScenarios) {
      const preview = buildRequestExplanationPreview(scenario.request);
      expect(preview.category).toBe(scenario.value);
    }
  });
});
