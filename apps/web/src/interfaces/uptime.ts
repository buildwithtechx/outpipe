export interface UptimeMonitor {
  id: string;
  organizationId: string;
  name: string;
  type: 'http' | 'https' | 'tcp';
  target: string;
  intervalSeconds: number;
  timeoutSeconds: number;
  expectedStatusCode?: number;
  status: 'active' | 'degraded' | 'down' | 'paused';
  createdAt: string;
  updatedAt: string;
}

export interface UptimeCheck {
  id: string;
  monitorId: string;
  status: 'up' | 'down' | 'degraded';
  statusCode: number;
  latencyMs: number;
  errorMessage?: string;
  checkedAt: string;
}

export interface UptimeIncidentUpdate {
  id: string;
  incidentId: string;
  status: 'investigating' | 'identified' | 'monitoring' | 'resolved';
  message: string;
  createdAt: string;
}

export interface UptimeIncident {
  id: string;
  organizationId: string;
  title: string;
  severity: 'minor' | 'major' | 'critical';
  status: 'investigating' | 'identified' | 'monitoring' | 'resolved';
  startedAt: string;
  resolvedAt?: string;
  updates: UptimeIncidentUpdate[];
  createdAt: string;
  updatedAt: string;
}

export interface UptimeStatusPage {
  id: string;
  organizationId: string;
  slug: string;
  title: string;
  description: string;
  isPublic: boolean;
  createdAt: string;
  updatedAt: string;
}
