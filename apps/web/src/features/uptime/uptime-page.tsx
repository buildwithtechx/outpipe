import {
  Activity,
  AlertTriangle,
  CheckCircle2,
  LineChart,
  Plus,
  RefreshCw,
  Settings,
  Trash2,
} from 'lucide-react';
import { useState } from 'react';
import { Button } from '#/components/ui/button';
import { useOrganization } from '#/features/organizations/hooks/use-organization';
import { AddMonitorModal } from './components/add-monitor-modal';
import { IncidentModal } from './components/incident-modal';
import { StatusPageModal } from './components/status-page-modal';
import {
  useIncidents,
  useMonitors,
  useStatusPageConfig,
  useUptimeMutations,
} from './hooks/use-uptime';

export function UptimePage({ orgSlug }: { orgSlug: string }) {
  const { organization } = useOrganization(orgSlug);
  const orgId = organization?.id;

  const { data: monitors = [], isLoading: loadingMonitors } =
    useMonitors(orgId);
  const { data: incidents = [] } = useIncidents(orgId);
  const { data: statusPage } = useStatusPageConfig(orgId);
  const mutations = useUptimeMutations(orgId || '');

  const [activeTab, setActiveTab] = useState<'monitors' | 'incidents'>(
    'monitors',
  );
  const [showAddMonitor, setShowAddMonitor] = useState(false);
  const [showIncidentModal, setShowIncidentModal] = useState(false);
  const [showStatusModal, setShowStatusModal] = useState(false);

  const healthyCount = monitors.filter((m) => m.status === 'active').length;
  const downCount = monitors.filter((m) => m.status === 'down').length;
  const activeIncidents = incidents.filter((i) => i.status !== 'resolved');

  const handleCreateMonitor = async (input: {
    name: string;
    type: 'http' | 'https' | 'tcp';
    target: string;
    interval_seconds: number;
    timeout_seconds: number;
    expected_status_code?: number;
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

  const handleSaveStatusPage = async (input: {
    slug: string;
    title: string;
    description: string;
    is_public: boolean;
  }) => {
    await mutations.saveStatusPage.mutateAsync(input);
    setShowStatusModal(false);
  };

  if (loadingMonitors) {
    return (
      <p className="p-8 text-sm text-white/55">Loading Uptime Monitors…</p>
    );
  }

  return (
    <div className="space-y-8 p-6 lg:p-10 max-w-7xl mx-auto text-white">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-white/10 pb-6">
        <div>
          <div className="flex items-center gap-2.5">
            <div className="flex size-9 items-center justify-center rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-400">
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
            variant="outline"
            size="sm"
            onClick={() => setShowStatusModal(true)}
            className="border-white/10 bg-white/5 text-xs text-white hover:bg-white/10"
          >
            <Settings className="mr-1.5 size-3.5" />
            Status Page Settings
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
            className="bg-emerald-600 hover:bg-emerald-500 text-xs text-white"
          >
            <Plus className="mr-1.5 size-3.5" />
            Add Monitor
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="rounded-xl border border-white/10 bg-white/5 p-4">
          <span className="text-xs text-white/50">Total Probes</span>
          <p className="text-2xl font-semibold mt-1">{monitors.length}</p>
        </div>
        <div className="rounded-xl border border-white/10 bg-white/5 p-4">
          <span className="text-xs text-emerald-400/80">Healthy Monitors</span>
          <p className="text-2xl font-semibold text-emerald-400 mt-1">
            {healthyCount}
          </p>
        </div>
        <div className="rounded-xl border border-white/10 bg-white/5 p-4">
          <span className="text-xs text-rose-400/80">Down / Outage</span>
          <p className="text-2xl font-semibold text-rose-400 mt-1">
            {downCount}
          </p>
        </div>
        <div className="rounded-xl border border-white/10 bg-white/5 p-4">
          <span className="text-xs text-amber-400/80">Active Incidents</span>
          <p className="text-2xl font-semibold text-amber-400 mt-1">
            {activeIncidents.length}
          </p>
        </div>
      </div>

      <div className="flex items-center gap-4 border-b border-white/10">
        <button
          type="button"
          onClick={() => setActiveTab('monitors')}
          className={`pb-3 text-sm font-medium border-b-2 transition-colors ${
            activeTab === 'monitors'
              ? 'border-emerald-500 text-white'
              : 'border-transparent text-white/50 hover:text-white/80'
          }`}
        >
          Probe Monitors ({monitors.length})
        </button>
        <button
          type="button"
          onClick={() => setActiveTab('incidents')}
          className={`pb-3 text-sm font-medium border-b-2 transition-colors ${
            activeTab === 'incidents'
              ? 'border-emerald-500 text-white'
              : 'border-transparent text-white/50 hover:text-white/80'
          }`}
        >
          Incidents ({incidents.length})
        </button>
      </div>

      {activeTab === 'monitors' ? (
        <div className="space-y-4">
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
                className="bg-emerald-600 text-xs"
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
                  {monitors.map((m) => (
                    <tr
                      key={m.id}
                      className="hover:bg-white/5 transition-colors"
                    >
                      <td className="px-4 py-3">
                        <span
                          className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full font-medium ${
                            m.status === 'active'
                              ? 'bg-emerald-500/15 text-emerald-400'
                              : 'bg-rose-500/15 text-rose-400'
                          }`}
                        >
                          <span
                            className={`size-1.5 rounded-full ${m.status === 'active' ? 'bg-emerald-400' : 'bg-rose-400'}`}
                          />
                          {m.status}
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <span className="font-semibold text-white block">
                          {m.name}
                        </span>
                        <span className="font-mono text-white/40">
                          {m.target}
                        </span>
                      </td>
                      <td className="px-4 py-3 uppercase font-mono">
                        {m.type}
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
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => mutations.removeMonitor.mutate(m.id)}
                          className="h-7 text-xs border-rose-500/20 text-rose-400 hover:bg-rose-500/10"
                        >
                          <Trash2 className="size-3" />
                        </Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      ) : (
        <div className="space-y-4">
          {incidents.length === 0 ? (
            <div className="rounded-xl border border-white/10 bg-white/5 p-12 text-center">
              <CheckCircle2 className="size-10 text-emerald-400/30 mx-auto mb-3" />
              <h3 className="text-sm font-medium text-white">
                No incidents logged
              </h3>
              <p className="text-xs text-white/50 mt-1">
                All monitored systems are operational.
              </p>
            </div>
          ) : (
            incidents.map((inc) => (
              <div
                key={inc.id}
                className="rounded-xl border border-white/10 bg-white/5 p-4 space-y-3"
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <span className="font-semibold text-white text-sm">
                      {inc.title}
                    </span>
                    <span className="px-2 py-0.5 rounded text-[10px] uppercase font-bold bg-white/10 text-white/70">
                      {inc.severity}
                    </span>
                  </div>
                  <span className="text-xs text-white/40">
                    {new Date(inc.startedAt).toLocaleString()}
                  </span>
                </div>
                <div className="space-y-1.5 pl-3 border-l-2 border-white/10 text-xs">
                  {inc.updates.map((u) => (
                    <div key={u.id}>
                      <span className="font-semibold capitalize text-emerald-400">
                        {u.status}:{' '}
                      </span>
                      <span className="text-white/70">{u.message}</span>
                    </div>
                  ))}
                </div>
              </div>
            ))
          )}
        </div>
      )}

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
      <StatusPageModal
        isOpen={showStatusModal}
        onClose={() => setShowStatusModal(false)}
        initialSlug={statusPage?.slug}
        initialTitle={statusPage?.title}
        initialDescription={statusPage?.description}
        initialIsPublic={statusPage?.isPublic}
        onSave={handleSaveStatusPage}
        isSaving={mutations.saveStatusPage.isPending}
      />
    </div>
  );
}
