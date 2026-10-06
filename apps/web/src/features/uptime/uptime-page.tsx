import { Link } from '@tanstack/react-router';
import {
  Activity,
  AlertTriangle,
  CheckCircle2,
  LineChart,
  Plus,
  RefreshCw,
  Settings,
} from 'lucide-react';
import { useState } from 'react';
import { ConfirmAction } from '#/components/feedback/confirm-action';
import { Badge } from '#/components/ui/badge';
import { Button } from '#/components/ui/button';
import { Card } from '#/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '#/components/ui/tabs';
import { useOrganization } from '#/features/organizations/hooks/use-organization';
import { AddMonitorModal } from './components/add-monitor-modal';
import { IncidentModal } from './components/incident-modal';
import { IncidentUpdateForm } from './components/incident-update-form';
import { MonitorStatusBadge } from './components/monitor-status-badge';
import {
  useIncidents,
  useMonitors,
  useUptimeMutations,
} from './hooks/use-uptime';

export function UptimePage({ orgSlug }: { orgSlug: string }) {
  const organizationQuery = useOrganization(orgSlug);
  const { organization } = organizationQuery;
  const orgId = organization?.id;

  const monitorQuery = useMonitors(orgId);
  const { data: monitors = [], isLoading: loadingMonitors } = monitorQuery;
  const incidentQuery = useIncidents(orgId);
  const { data: incidents = [] } = incidentQuery;
  const mutations = useUptimeMutations(orgId || '');

  const [showAddMonitor, setShowAddMonitor] = useState(false);
  const [showIncidentModal, setShowIncidentModal] = useState(false);

  const healthyCount = monitors.filter((m) => m.status === 'up').length;
  const downCount = monitors.filter((m) => m.status === 'down').length;
  const activeIncidents = incidents.filter((i) => i.status !== 'resolved');

  const handleCreateMonitor = async (input: {
    name: string;
    type: 'http' | 'https' | 'tcp' | 'icmp';
    target: string;
    interval_seconds: number;
    timeout_seconds: number;
    expected_status_code?: number;
    body_regex?: string;
    max_latency_ms?: number;
  }) => {
    await mutations.addMonitor.mutateAsync(input);
    setShowAddMonitor(false);
  };

  const handleCreateIncident = async (input: {
    title: string;
    severity: 'minor' | 'major' | 'critical';
    message: string;
  }) => {
    await mutations.addIncident.mutateAsync(input);
    setShowIncidentModal(false);
  };

  if (organizationQuery.isLoading || loadingMonitors) {
    return (
      <p className="p-8 text-sm text-white/55">Loading Uptime Monitors…</p>
    );
  }

  if (!orgId || monitorQuery.isError)
    return (
      <div role="alert" className="space-y-3">
        <p>Could not load uptime monitors.</p>
        <Button
          variant="outline"
          onClick={() => {
            void organizationQuery.refetch();
            void monitorQuery.refetch();
          }}
        >
          Try again
        </Button>
      </div>
    );

  return (
    <div className="space-y-8 text-foreground">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-white/10 pb-6">
        <div>
          <div className="flex items-center gap-2.5">
            <div className="flex size-9 items-center justify-center rounded-xl bg-indigo-300/10 border border-indigo-300/20 text-indigo-200">
              <LineChart className="size-5" />
            </div>
            <h1 className="text-2xl font-bold tracking-tight text-white">
              Uptime & Status
            </h1>
          </div>
          <p className="mt-1 text-sm text-white/60">
            Realtime multi-protocol probe checks, active incident management &
            public status page.
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-2.5">
          <Button
            asChild
            variant="outline"
            size="sm"
            className="border-white/10 bg-white/5 text-xs text-white hover:bg-white/10"
          >
            <Link to="/$orgSlug/uptime/status-page" params={{ orgSlug }}>
              <Settings className="mr-1.5 size-3.5" />
              Manage status page
            </Link>
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setShowIncidentModal(true)}
            className="border-rose-500/30 bg-rose-500/10 text-xs text-rose-300 hover:bg-rose-500/20"
          >
            <AlertTriangle className="mr-1.5 size-3.5" />
            Report Incident
          </Button>
          <Button
            size="sm"
            onClick={() => setShowAddMonitor(true)}
            className="bg-primary hover:bg-primary/90 text-xs text-primary-foreground"
          >
            <Plus className="mr-1.5 size-3.5" />
            Add Monitor
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card className="border-white/10 bg-white/5 p-4 text-white">
          <span className="text-xs text-white/50">Total Probes</span>
          <p className="text-2xl font-semibold mt-1">{monitors.length}</p>
        </Card>
        <Card className="border-white/10 bg-white/5 p-4 text-white">
          <span className="text-xs text-emerald-400/80">Healthy Monitors</span>
          <p className="text-2xl font-semibold text-emerald-400 mt-1">
            {healthyCount}
          </p>
        </Card>
        <Card className="border-white/10 bg-white/5 p-4 text-white">
          <span className="text-xs text-rose-400/80">Down / Outage</span>
          <p className="text-2xl font-semibold text-rose-400 mt-1">
            {downCount}
          </p>
        </Card>
        <Card className="border-white/10 bg-white/5 p-4 text-white">
          <span className="text-xs text-amber-400/80">Active Incidents</span>
          <p className="text-2xl font-semibold text-amber-400 mt-1">
            {activeIncidents.length}
          </p>
        </Card>
      </div>

      <Tabs defaultValue="monitors" className="space-y-4">
        <TabsList className="bg-white/5 border border-white/10 p-1">
          <TabsTrigger value="monitors" className="text-xs">
            Probe Monitors ({monitors.length})
          </TabsTrigger>
          <TabsTrigger value="incidents" className="text-xs">
            Incidents ({incidents.length})
          </TabsTrigger>
        </TabsList>

        <TabsContent value="monitors" className="mt-0">
          {monitors.length === 0 ? (
            <div className="rounded-xl border border-white/10 bg-white/5 p-12 text-center">
              <Activity className="size-10 text-white/20 mx-auto mb-3" />
              <h3 className="text-sm font-medium text-white">
                No monitors configured
              </h3>
              <p className="text-xs text-white/50 mt-1 mb-4">
                Start monitoring endpoints, relays, or API routes for automatic
                downtime alerts.
              </p>
              <Button
                size="sm"
                onClick={() => setShowAddMonitor(true)}
                className="bg-emerald-600 hover:bg-emerald-500 text-xs text-white"
              >
                <Plus className="mr-1.5 size-3.5" />
                Create your first monitor
              </Button>
            </div>
          ) : (
            <div className="rounded-xl border border-white/10 bg-white/5 overflow-hidden">
              <table className="w-full text-left text-xs text-white/70">
                <thead className="bg-white/5 text-white/50 border-b border-white/10">
                  <tr>
                    <th className="px-4 py-3 font-medium">Status</th>
                    <th className="px-4 py-3 font-medium">Name & Target</th>
                    <th className="px-4 py-3 font-medium">Protocol</th>
                    <th className="px-4 py-3 font-medium">Interval</th>
                    <th className="px-4 py-3 font-medium text-right">
                      Actions
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-white/5">
                  {monitors.map((m) => {
                    return (
                      <tr
                        key={m.id}
                        className="hover:bg-white/5 transition-colors"
                      >
                        <td className="px-4 py-3">
                          <MonitorStatusBadge status={m.status} />
                        </td>
                        <td className="px-4 py-3">
                          <span className="font-semibold text-white block">
                            {m.name}
                          </span>
                          <span className="font-mono text-white/40">
                            {m.url}
                          </span>
                        </td>
                        <td className="px-4 py-3 uppercase font-mono">
                          {m.protocol}
                        </td>
                        <td className="px-4 py-3">{m.intervalSeconds}s</td>
                        <td className="px-4 py-3 text-right space-x-2">
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => mutations.probeMonitor.mutate(m.id)}
                            className="h-7 text-xs border-white/10 bg-white/5 text-white/80"
                            title="Run probe test now"
                          >
                            <RefreshCw className="size-3 mr-1" />
                            Test
                          </Button>
                          <ConfirmAction
                            title={`Delete ${m.name}?`}
                            description="This removes the monitor and its health checks from your workspace and public status page."
                            label="Delete"
                            pending={mutations.removeMonitor.isPending}
                            onConfirm={() =>
                              mutations.removeMonitor.mutateAsync(m.id)
                            }
                          />
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </TabsContent>

        <TabsContent value="incidents" className="mt-0 space-y-4">
          {incidentQuery.isError ? (
            <p role="alert">
              Could not load incidents.{' '}
              <Button
                variant="outline"
                onClick={() => void incidentQuery.refetch()}
              >
                Try again
              </Button>
            </p>
          ) : incidentQuery.isLoading ? (
            <p role="status">Loading incidents...</p>
          ) : incidents.length === 0 ? (
            <div className="rounded-xl border border-white/10 bg-white/5 p-12 text-center">
              <CheckCircle2 className="size-10 text-emerald-400/30 mx-auto mb-3" />
              <h3 className="text-sm font-medium text-white">
                No incidents logged
              </h3>
              <p className="text-xs text-white/50 mt-1">
                No incidents have been reported. Check monitors for current
                service health.
              </p>
            </div>
          ) : (
            incidents.map((inc) => (
              <Card
                key={inc.id}
                className="border-white/10 bg-white/5 p-4 space-y-3 text-white"
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <span className="font-semibold text-white text-sm">
                      {inc.title}
                    </span>
                    <Badge
                      variant="outline"
                      className="text-[10px] uppercase font-bold border-white/10 text-white/70"
                    >
                      {inc.severity}
                    </Badge>
                  </div>
                  <span className="text-xs text-white/40">
                    {new Date(inc.startedAt).toLocaleString()}
                  </span>
                </div>
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <span className="text-sm capitalize text-muted-foreground">
                    {inc.status}
                  </span>
                  <IncidentUpdateForm incident={inc} />
                </div>
                <div className="space-y-1.5 pl-3 border-l-2 border-white/10 text-xs">
                  {(inc.updates ?? []).map((u) => (
                    <div key={u.id}>
                      <span className="font-semibold capitalize text-emerald-400">
                        {u.status}:{' '}
                      </span>
                      <span className="text-white/70">{u.message}</span>
                    </div>
                  ))}
                </div>
              </Card>
            ))
          )}
        </TabsContent>
      </Tabs>

      <AddMonitorModal
        isOpen={showAddMonitor}
        onClose={() => setShowAddMonitor(false)}
        onSave={handleCreateMonitor}
        isSaving={mutations.addMonitor.isPending}
      />
      <IncidentModal
        isOpen={showIncidentModal}
        onClose={() => setShowIncidentModal(false)}
        onSave={handleCreateIncident}
        isSaving={mutations.addIncident.isPending}
      />
    </div>
  );
}
