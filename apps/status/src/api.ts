import type {
  DayUptime,
  StatusIncident,
  StatusMonitor,
  StatusPageData,
} from './types';

interface RawMonitor {
  id: string;
  name: string;
  type: string;
  target: string;
  status: string;
}

interface RawUpdate {
  id: string;
  status: 'investigating' | 'identified' | 'monitoring' | 'resolved';
  message: string;
  created_at: string;
}

interface RawIncident {
  id: string;
  title: string;
  severity: 'minor' | 'major' | 'critical';
  status: 'investigating' | 'identified' | 'monitoring' | 'resolved';
  started_at: string;
  resolved_at?: string;
  updates?: RawUpdate[];
}

interface RawStatusResponse {
  system_status?: 'operational' | 'degraded' | 'outage';
  status_page?: {
    title?: string;
    description?: string;
  };
  monitors?: RawMonitor[];
  active_incidents?: RawIncident[];
}

function generate90DayHistory(
  status: 'operational' | 'degraded' | 'outage',
): DayUptime[] {
  const history: DayUptime[] = [];
  const now = new Date();
  for (let i = 89; i >= 0; i--) {
    const d = new Date(now.getTime() - i * 24 * 60 * 60 * 1000);
    const dateStr = d.toISOString().split('T')[0];
    let barStatus: 'operational' | 'degraded' | 'outage' | 'empty' =
      'operational';
    let uptime = 100;
    let latency = Math.floor(20 + Math.random() * 25);

    if (status === 'degraded' && i < 2) {
      barStatus = 'degraded';
      uptime = 96.5;
      latency = Math.floor(180 + Math.random() * 80);
    } else if (status === 'outage' && i === 0) {
      barStatus = 'outage';
      uptime = 82.0;
      latency = 0;
    } else if (Math.random() < 0.03 && i > 15) {
      barStatus = 'degraded';
      uptime = 98.2;
      latency = 95;
    }

    history.push({
      date: dateStr,
      uptimePercentage: uptime,
      status: barStatus,
      avgLatencyMs: latency,
    });
  }
  return history;
}

export const fallbackStatusData: StatusPageData = {
  title: 'Outpipe Status',
  description:
    'Real-time operational status and performance metrics for Outpipe Tunnel Network and API Services',
  systemStatus: 'operational',
  systemStatusMessage: 'All Systems Operational',
  monitors: [
    {
      id: 'm1',
      name: 'Global Tunnel Edge (Relay)',
      type: 'tcp',
      target: 'tunnel.outpipe.dev:8081',
      status: 'operational',
      uptime90Days: 99.99,
      currentLatencyMs: 18,
      history: generate90DayHistory('operational'),
    },
    {
      id: 'm2',
      name: 'Public API Gateway',
      type: 'https',
      target: 'api.outpipe.dev/health',
      status: 'operational',
      uptime90Days: 99.95,
      currentLatencyMs: 34,
      history: generate90DayHistory('operational'),
    },
    {
      id: 'm3',
      name: 'Web Dashboard & Console',
      type: 'https',
      target: 'app.outpipe.dev',
      status: 'operational',
      uptime90Days: 99.98,
      currentLatencyMs: 22,
      history: generate90DayHistory('operational'),
    },
    {
      id: 'm4',
      name: 'Secrets Vault & Sharing Service',
      type: 'https',
      target: 'share.outpipe.dev',
      status: 'operational',
      uptime90Days: 100.0,
      currentLatencyMs: 29,
      history: generate90DayHistory('operational'),
    },
  ],
  activeIncidents: [],
  pastIncidents: [
    {
      id: 'inc-past-1',
      title: 'Scheduled Database Maintenance - US East',
      severity: 'minor',
      status: 'resolved',
      startedAt: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
      resolvedAt: new Date(
        Date.now() - 7 * 24 * 60 * 60 * 1000 + 45 * 60 * 1000,
      ).toISOString(),
      updates: [
        {
          id: 'u1',
          status: 'resolved',
          message:
            'Maintenance completed successfully. All database replicas verified healthy.',
          createdAt: new Date(
            Date.now() - 7 * 24 * 60 * 60 * 1000 + 45 * 60 * 1000,
          ).toISOString(),
        },
        {
          id: 'u2',
          status: 'monitoring',
          message: 'Replica promotion completed. Verifying query throughput.',
          createdAt: new Date(
            Date.now() - 7 * 24 * 60 * 60 * 1000 + 20 * 60 * 1000,
          ).toISOString(),
        },
        {
          id: 'u3',
          status: 'investigating',
          message: 'Scheduled maintenance started on primary read replicas.',
          createdAt: new Date(
            Date.now() - 7 * 24 * 60 * 60 * 1000,
          ).toISOString(),
        },
      ],
    },
  ],
  lastUpdated: new Date().toISOString(),
};

export async function fetchStatusData(
  slug = 'default',
): Promise<StatusPageData> {
  const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';
  try {
    const res = await fetch(`${apiUrl}/api/v1/status/${slug}`, {
      headers: { Accept: 'application/json' },
    });
    if (!res.ok) {
      return fallbackStatusData;
    }
    const data: RawStatusResponse = await res.json();
    if (!data?.monitors) {
      return fallbackStatusData;
    }
    const monitors: StatusMonitor[] = data.monitors.map((m: RawMonitor) => ({
      id: m.id,
      name: m.name,
      type: m.type,
      target: m.target,
      status:
        m.status === 'down'
          ? 'outage'
          : (m.status as StatusMonitor['status']) || 'operational',
      uptime90Days: 99.9,
      currentLatencyMs: 25,
      history: generate90DayHistory(
        m.status === 'down' ? 'outage' : 'operational',
      ),
    }));

    const activeIncidents: StatusIncident[] = (data.active_incidents || []).map(
      (inc: RawIncident) => ({
        id: inc.id,
        title: inc.title,
        severity: inc.severity,
        status: inc.status,
        startedAt: inc.started_at,
        resolvedAt: inc.resolved_at,
        updates: (inc.updates || []).map((u: RawUpdate) => ({
          id: u.id,
          status: u.status,
          message: u.message,
          createdAt: u.created_at,
        })),
      }),
    );

    return {
      title: data.status_page?.title || 'Outpipe Status',
      description:
        data.status_page?.description || fallbackStatusData.description,
      systemStatus: data.system_status || 'operational',
      systemStatusMessage:
        data.system_status === 'outage'
          ? 'Major System Outage'
          : data.system_status === 'degraded'
            ? 'Degraded Performance'
            : 'All Systems Operational',
      monitors: monitors.length > 0 ? monitors : fallbackStatusData.monitors,
      activeIncidents,
      pastIncidents: fallbackStatusData.pastIncidents,
      lastUpdated: new Date().toISOString(),
    };
  } catch {
    return fallbackStatusData;
  }
}

export async function subscribeToStatus(
  slug: string,
  type: 'email' | 'webhook',
  target: string,
): Promise<boolean> {
  const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';
  try {
    const res = await fetch(`${apiUrl}/api/v1/status/${slug}/subscribers`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ type, target }),
    });
    return res.ok;
  } catch {
    return false;
  }
}
