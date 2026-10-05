import type {
  ObservabilityStats,
  ReplayRequestInput,
  ReplayResponseOutput,
  RequestCapture,
  TelemetryLog,
  TelemetrySpan,
} from '#/interfaces';
import { apiClient } from '#/lib/api-client';
import {
  normalizeObservability,
  normalizeObservabilityList,
  type ObservabilityResponse,
} from './observability-response';

export type {
  ObservabilityStats,
  ReplayRequestInput,
  ReplayResponseOutput,
  RequestCapture,
  TelemetryLog,
  TelemetrySpan,
};

export function getObservabilityStats(organizationId: string, range = '24h') {
  return apiClient
    .get<ObservabilityResponse<ObservabilityStats>>(
      `/api/v1/organizations/${organizationId}/observability/stats?range=${range}`,
    )
    .then(normalizeObservability<ObservabilityStats>);
}

export function getTraces(organizationId: string, limit = 50) {
  return apiClient
    .get<ObservabilityResponse<TelemetrySpan>[]>(
      `/api/v1/organizations/${organizationId}/observability/traces?limit=${limit}`,
    )
    .then(normalizeObservabilityList<TelemetrySpan>);
}

export function getTraceWaterfall(organizationId: string, traceId: string) {
  return apiClient
    .get<ObservabilityResponse<TelemetrySpan>[]>(
      `/api/v1/organizations/${organizationId}/observability/traces/${traceId}`,
    )
    .then(normalizeObservabilityList<TelemetrySpan>);
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
  return apiClient
    .get<ObservabilityResponse<TelemetryLog>[]>(
      `/api/v1/organizations/${organizationId}/observability/logs?${params.toString()}`,
    )
    .then(normalizeObservabilityList<TelemetryLog>);
}

export function getCaptures(
  organizationId: string,
  tunnelId?: string,
  limit = 50,
) {
  const params = new URLSearchParams();
  if (tunnelId) params.set('tunnel_id', tunnelId);
  params.set('limit', String(limit));
  return apiClient
    .get<ObservabilityResponse<RequestCapture>[]>(
      `/api/v1/organizations/${organizationId}/observability/captures?${params.toString()}`,
    )
    .then(normalizeObservabilityList<RequestCapture>);
}

export function getCapture(organizationId: string, captureId: string) {
  return apiClient
    .get<ObservabilityResponse<RequestCapture>>(
      `/api/v1/organizations/${organizationId}/observability/captures/${captureId}`,
    )
    .then(normalizeObservability<RequestCapture>);
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
