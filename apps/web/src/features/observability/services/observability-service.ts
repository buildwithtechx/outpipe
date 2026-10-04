import type {
  ObservabilityStats,
  ReplayRequestInput,
  ReplayResponseOutput,
  RequestCapture,
  TelemetryLog,
  TelemetrySpan,
} from '#/interfaces';
import { apiClient } from '#/lib/api-client';

export type {
  ObservabilityStats,
  ReplayRequestInput,
  ReplayResponseOutput,
  RequestCapture,
  TelemetryLog,
  TelemetrySpan,
};

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
