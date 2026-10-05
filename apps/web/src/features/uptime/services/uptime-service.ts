import type {
  UptimeCheck,
  UptimeIncident,
  UptimeIncidentUpdate,
  UptimeMonitor,
  UptimeStatusPage,
} from '#/interfaces';
import { apiClient } from '#/lib/api-client';

export type {
  UptimeCheck,
  UptimeIncident,
  UptimeIncidentUpdate,
  UptimeMonitor,
  UptimeStatusPage,
};

export function getMonitors(organizationId: string) {
  return apiClient.get<UptimeMonitor[]>(
    `/api/v1/organizations/${organizationId}/uptime/monitors`,
  );
}

export function createMonitor(
  organizationId: string,
  input: {
    name: string;
    type: 'http' | 'https' | 'tcp' | 'icmp';
    target: string;
    interval_seconds?: number;
    timeout_seconds?: number;
    expected_status_code?: number;
    body_regex?: string;
    max_latency_ms?: number;
  },
) {
  return apiClient.post<UptimeMonitor>(
    `/api/v1/organizations/${organizationId}/uptime/monitors`,
    {
      name: input.name,
      protocol: input.type,
      url: input.target,
      intervalSeconds: input.interval_seconds,
      timeoutSeconds: input.timeout_seconds,
      expectedStatusCode: input.expected_status_code,
      bodyRegex: input.body_regex,
      maxLatencyMs: input.max_latency_ms,
    },
  );
}

export function updateMonitor(
  organizationId: string,
  monitorId: string,
  input: Partial<{
    name: string;
    type: 'http' | 'https' | 'tcp' | 'icmp';
    target: string;
    interval_seconds: number;
    timeout_seconds: number;
    expected_status_code: number;
    status: 'active' | 'degraded' | 'down' | 'paused';
  }>,
) {
  return apiClient.put<UptimeMonitor>(
    `/api/v1/organizations/${organizationId}/uptime/monitors/${monitorId}`,
    input,
  );
}

export function deleteMonitor(organizationId: string, monitorId: string) {
  return apiClient.delete<void>(
    `/api/v1/organizations/${organizationId}/uptime/monitors/${monitorId}`,
  );
}

export function testMonitor(organizationId: string, monitorId: string) {
  return apiClient.post<UptimeCheck>(
    `/api/v1/organizations/${organizationId}/uptime/monitors/${monitorId}/probe`,
    {},
  );
}

export function getMonitorChecks(
  organizationId: string,
  monitorId: string,
  limit = 50,
) {
  return apiClient.get<UptimeCheck[]>(
    `/api/v1/organizations/${organizationId}/uptime/monitors/${monitorId}/checks?limit=${limit}`,
  );
}

export function getIncidents(organizationId: string) {
  return apiClient.get<UptimeIncident[]>(
    `/api/v1/organizations/${organizationId}/uptime/incidents`,
  );
}

export function createIncident(
  organizationId: string,
  input: {
    title: string;
    severity: 'minor' | 'major' | 'critical';
    message: string;
  },
) {
  return apiClient.post<UptimeIncident>(
    `/api/v1/organizations/${organizationId}/uptime/incidents`,
    input,
  );
}

export function updateIncident(
  organizationId: string,
  incidentId: string,
  input: {
    status: 'investigating' | 'identified' | 'monitoring' | 'resolved';
    message: string;
  },
) {
  return apiClient.post<UptimeIncidentUpdate>(
    `/api/v1/organizations/${organizationId}/uptime/incidents/${incidentId}/updates`,
    input,
  );
}

export function getStatusPageConfig(organizationId: string) {
  return apiClient.get<UptimeStatusPage>(
    `/api/v1/organizations/${organizationId}/uptime/status-page`,
  );
}

export function updateStatusPageConfig(
  organizationId: string,
  input: {
    slug: string;
    title: string;
    description: string;
    is_public: boolean;
  },
) {
  return apiClient.put<UptimeStatusPage>(
    `/api/v1/organizations/${organizationId}/uptime/status-page`,
    input,
  );
}
