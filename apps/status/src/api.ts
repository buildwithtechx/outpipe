import type { StatusIncident, StatusMonitor, StatusPageData } from './types';

interface RawMonitor {
  id: string;
  name: string;
  protocol: string;
  url: string;
  status: 'up' | 'down' | 'degraded' | 'paused';
  uptimeRatio: number;
  latencyMs: number;
}

export interface RawStatusResponse {
  overallStatus:
    | 'operational'
    | 'degraded_performance'
    | 'partial_outage'
    | 'major_outage';
  page: { title: string; description?: string };
  monitors: RawMonitor[] | null;
  activeIncidents: StatusIncident[] | null;
  pastIncidents: StatusIncident[] | null;
}

export function mapStatusResponse(data: RawStatusResponse): StatusPageData {
  const systemStatus =
    data.overallStatus === 'operational'
      ? 'operational'
      : data.overallStatus === 'degraded_performance'
        ? 'degraded'
        : 'outage';
  const monitors: StatusMonitor[] = (data.monitors ?? []).map((monitor) => ({
    id: monitor.id,
    name: monitor.name,
    type: monitor.protocol,
    target: monitor.url,
    status: monitor.status === 'up' ? 'operational' : monitor.status,
    uptime90Days: monitor.uptimeRatio,
    currentLatencyMs: monitor.latencyMs,
    history: [],
  }));
  const incidents = (items: StatusIncident[] | null) =>
    (items ?? []).map((incident) => ({
      ...incident,
      updates: incident.updates ?? [],
    }));
  return {
    title: data.page.title,
    description: data.page.description ?? '',
    systemStatus,
    systemStatusMessage:
      data.overallStatus === 'partial_outage'
        ? 'Partial System Outage'
        : systemStatus === 'outage'
          ? 'Major System Outage'
          : systemStatus === 'degraded'
            ? 'Degraded Performance'
            : 'All Systems Operational',
    monitors,
    activeIncidents: incidents(data.activeIncidents),
    pastIncidents: incidents(data.pastIncidents),
    lastUpdated: new Date().toISOString(),
  };
}

export async function fetchStatusData(
  slug = 'default',
): Promise<StatusPageData | null> {
  const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';
  try {
    const response = await fetch(
      `${apiUrl}/api/v1/status/${encodeURIComponent(slug)}`,
      {
        headers: { Accept: 'application/json' },
      },
    );
    if (!response.ok) return null;
    return mapStatusResponse(await response.json());
  } catch {
    return null;
  }
}

export async function subscribeToStatus(
  slug: string,
  type: 'email' | 'webhook',
  target: string,
): Promise<boolean> {
  if (type !== 'email') return false;
  const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';
  try {
    const response = await fetch(
      `${apiUrl}/api/v1/status/${encodeURIComponent(slug)}/subscribe`,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email: target }),
      },
    );
    return response.ok;
  } catch {
    return false;
  }
}
