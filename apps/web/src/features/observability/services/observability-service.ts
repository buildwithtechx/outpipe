import { apiClient } from '#/lib/api-client';

export interface TelemetrySpan {
  id: string;
  organizationId: string;
  traceId: string;
  spanId: string;
  parentSpanId?: string;
  name: string;
  kind: string;
  startTime: string;
  endTime: string;
  durationMs: number;
  statusCode: string;
  statusMessage?: string;
  attributes?: string;
  events?: string;
  createdAt: string;
}

export interface TelemetryLog {
  id: string;
  organizationId: string;
  traceId?: string;
  spanId?: string;
  timestamp: string;
  severity: string;
  body: string;
  attributes?: string;
  createdAt: string;
}

export interface RequestCapture {
  id: string;
  organizationId: string;
  tunnelId: string;
  timestamp: string;
  method: string;
  path: string;
  statusCode: number;
  durationMs: number;
  requestHeaders?: string;
  requestBody?: string;
  requestBodySize: number;
  responseHeaders?: string;
  responseBody?: string;
  responseBodySize: number;
  createdAt: string;
}

export interface ObservabilityStats {
  totalRequests: number;
  requestsChange: number;
  totalBytes: number;
  dataTransferChange: number;
  p50LatencyMs: number;
  p95LatencyMs: number;
  p99LatencyMs: number;
  chartData: Array<{
    time: string;
    requests: number;
    errors: number;
    bytes: number;
  }>;
}

export interface ReplayRequestInput {
  url: string;
  method: string;
  headers?: Record<string, string>;
  request_body?: string;
}

export interface ReplayResponseOutput {
  status_code: number;
  status_text: string;
  headers: Record<string, string>;
  body: string;
  duration_ms: number;
}

export function getObservabilityStats(organizationId: string, range = '24h') {
  return apiClient.get<ObservabilityStats>(
    `/api/v1/organizations/${organizationId}/observability/stats?range=${range}`,
  );
}

export function getTraces(organizationId: string, limit = 50) {
  return apiClient.get<TelemetrySpan[]>(
    `/api/v1/organizations/${organizationId}/observability/traces?limit=${limit}`,
  );
}

export function getTraceWaterfall(organizationId: string, traceId: string) {
  return apiClient.get<TelemetrySpan[]>(
    `/api/v1/organizations/${organizationId}/observability/traces/${traceId}`,
  );
}

export function getLogs(
  organizationId: string,
  search?: string,
  severity?: string,
  limit = 100,
) {
  const params = new URLSearchParams();
  if (search) params.set('search', search);
  if (severity) params.set('severity', severity);
  params.set('limit', String(limit));
  return apiClient.get<TelemetryLog[]>(
    `/api/v1/organizations/${organizationId}/observability/logs?${params.toString()}`,
  );
}

export function getCaptures(
  organizationId: string,
  tunnelId?: string,
  limit = 50,
) {
  const params = new URLSearchParams();
  if (tunnelId) params.set('tunnel_id', tunnelId);
  params.set('limit', String(limit));
  return apiClient.get<RequestCapture[]>(
    `/api/v1/organizations/${organizationId}/observability/captures?${params.toString()}`,
  );
}

export function getCapture(organizationId: string, captureId: string) {
  return apiClient.get<RequestCapture>(
    `/api/v1/organizations/${organizationId}/observability/captures/${captureId}`,
  );
}

export function executeReplay(
  organizationId: string,
  input: ReplayRequestInput,
) {
  return apiClient.post<ReplayResponseOutput>(
    `/api/v1/organizations/${organizationId}/observability/replay`,
    input,
  );
}
