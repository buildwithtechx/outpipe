import { expect, it } from 'vitest';
import {
  normalizeObservability,
  normalizeObservabilityList,
} from './observability-response';

it('maps stats, waterfall timestamps, and capture payload fields', () => {
  expect(
    normalizeObservability<{
      totalRequests: number;
      p95LatencyMs: number;
      chartData: unknown[];
    }>({ total_requests: 18, p95_latency_ms: 43, chart_data: [] }),
  ).toEqual({ totalRequests: 18, p95LatencyMs: 43, chartData: [] });
  expect(
    normalizeObservabilityList<{
      startTime: string;
      endTime: string;
      durationMs: number;
    }>([{ start_time: 'start', end_time: 'end', duration_ms: 55 }]),
  ).toEqual([{ startTime: 'start', endTime: 'end', durationMs: 55 }]);
  expect(
    normalizeObservability<{ requestBody: string; requestHeaders: string }>({
      request_body: '{"field_name":1}',
      request_headers: '{"x-custom-header":"value"}',
    }),
  ).toEqual({
    requestBody: '{"field_name":1}',
    requestHeaders: '{"x-custom-header":"value"}',
  });
  expect(normalizeObservabilityList(null)).toEqual([]);
});
