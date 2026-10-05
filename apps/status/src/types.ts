export interface MonitorCheckData {
  id: string;
  status: 'up' | 'down' | 'degraded';
  statusCode: number;
  latencyMs: number;
  checkedAt: string;
}

export interface DayUptime {
  date: string;
  uptimePercentage: number;
  status: 'operational' | 'degraded' | 'outage' | 'empty';
  avgLatencyMs: number;
}

export interface StatusMonitor {
  id: string;
  name: string;
  type: string;
  target: string;
  status: 'operational' | 'degraded' | 'down' | 'maintenance' | 'paused';
  uptime90Days: number;
  currentLatencyMs: number;
  history: DayUptime[];
}

export interface IncidentUpdate {
  id: string;
  status: 'investigating' | 'identified' | 'monitoring' | 'resolved';
  message: string;
  createdAt: string;
}

export interface StatusIncident {
  id: string;
  title: string;
  severity: 'minor' | 'major' | 'critical';
  status: 'investigating' | 'identified' | 'monitoring' | 'resolved';
  startedAt: string;
  resolvedAt?: string;
  updates: IncidentUpdate[];
}

export interface StatusPageData {
  title: string;
  description: string;
  systemStatus: 'operational' | 'degraded' | 'outage';
  systemStatusMessage: string;
  monitors: StatusMonitor[];
  activeIncidents: StatusIncident[];
  pastIncidents: StatusIncident[];
  lastUpdated: string;
}
