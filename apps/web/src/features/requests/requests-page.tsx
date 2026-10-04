import { ArrowRightLeft, Eye } from 'lucide-react';
import { useState } from 'react';
import { Button } from '#/components/ui/button';
import { ReplayDialog, RequestInspectorModal } from '#/features/observability';
import { useReplayMutation } from '#/features/observability/hooks/use-observability';
import type { RequestCapture } from '#/features/observability/services/observability-service';
import { useOrganization } from '#/features/organizations/hooks/use-organization';
import { useRequestEvents } from './hooks/use-request-events';

export function RequestsPage({ orgSlug }: { orgSlug: string }) {
  const organizationQuery = useOrganization(orgSlug);
  const organizationId = organizationQuery.organization?.id;

  const query = useRequestEvents(organizationId);
  const replayMutation = useReplayMutation(organizationId || '');

  const [inspectedCapture, setInspectedCapture] =
    useState<RequestCapture | null>(null);
  const [showReplay, setShowReplay] = useState(false);
  const [replayInitial, setReplayInitial] =
    useState<Partial<RequestCapture> | null>(null);

  if (organizationQuery.isLoading || query.isLoading) {
    return <p className="p-8 text-sm text-white/55">Loading requests…</p>;
  }

  if (
    organizationQuery.isError ||
    query.isError ||
    !organizationQuery.organization
  ) {
    return (
      <p className="p-8 text-sm text-rose-200">
        We could not load request activity.
      </p>
    );
  }

  const organization = organizationQuery.organization;
  const events = query.data?.events ?? [];

  const handleOpenReplay = (capture: RequestCapture) => {
    setReplayInitial(capture);
    setShowReplay(true);
  };

  return (
    <div className="w-full max-w-6xl space-y-6 pb-12 text-white">
      <header className="border-b border-white/10 pb-6 flex items-center justify-between">
        <div>
          <p className="mb-2 text-xs font-semibold uppercase tracking-wider text-indigo-400">
            {organization.name}
          </p>
          <h1 className="text-3xl font-semibold tracking-[-0.04em]">
            Live Requests
          </h1>
          <p className="mt-1 text-sm text-white/55">
            Realtime traffic observed across your active tunnels with payload
            inspector & replay.
          </p>
        </div>
        <Button
          size="sm"
          onClick={() => {
            setReplayInitial(null);
            setShowReplay(true);
          }}
          className="bg-indigo-600 hover:bg-indigo-500 text-xs"
        >
          <ArrowRightLeft className="mr-1.5 size-3.5" />
          Replay Request
        </Button>
      </header>

      <section className="mt-6 overflow-hidden rounded-2xl border border-white/10 bg-white/[0.025]">
        <div className="overflow-x-auto">
          <div className="min-w-[650px]">
            <div className="grid grid-cols-[90px_minmax(0,1fr)_80px_90px_130px] gap-4 border-b border-white/10 px-5 py-3 text-xs uppercase tracking-wider text-white/35">
              <span>Method</span>
              <span>Path / Route</span>
              <span>Status</span>
              <span>Duration</span>
              <span className="text-right">Actions</span>
            </div>
            {events.length ? (
              events.map((event) => {
                const asCapture: RequestCapture = {
                  id: event.id,
                  organizationId: organizationId || '',
                  tunnelId: event.tunnelId || '',
                  timestamp: event.createdAt,
                  method: event.method || 'GET',
                  path: event.path || '/',
                  statusCode: event.statusCode || 200,
                  durationMs: event.durationMillis || 0,
                  requestBodySize: event.bytes,
                  responseBodySize: event.responseBytes || 0,
                  createdAt: event.createdAt,
                };

                return (
                  <div
                    key={event.id}
                    className="grid grid-cols-[90px_minmax(0,1fr)_80px_90px_130px] gap-4 border-b border-white/5 px-5 py-3.5 font-mono text-xs items-center hover:bg-white/5 transition-colors last:border-0"
                  >
                    <span>
                      <span
                        className={`px-2 py-0.5 rounded text-[11px] font-bold ${methodColor(event.method)}`}
                      >
                        {event.method ?? 'GET'}
                      </span>
                    </span>
                    <span className="truncate text-white/90">
                      {event.path ?? event.eventType}
                    </span>
                    <span>
                      <span
                        className={`px-2 py-0.5 rounded text-[11px] font-bold ${statusColor(event.statusCode)}`}
                      >
                        {event.statusCode ?? '—'}
                      </span>
                    </span>
                    <span className="text-white/50">
                      {event.durationMillis ? `${event.durationMillis}ms` : '—'}
                    </span>
                    <div className="flex items-center justify-end gap-1.5">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setInspectedCapture(asCapture)}
                        className="h-7 px-2 text-[11px] border-white/10 bg-white/5 text-white/80"
                        title="Inspect request"
                      >
                        <Eye className="size-3 mr-1" />
                        Inspect
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleOpenReplay(asCapture)}
                        className="h-7 px-2 text-[11px] border-indigo-500/30 text-indigo-300 hover:bg-indigo-500/10"
                        title="Replay request"
                      >
                        <ArrowRightLeft className="size-3" />
                      </Button>
                    </div>
                  </div>
                );
              })
            ) : (
              <p className="px-5 py-12 text-center text-sm text-white/50">
                No requests recorded in the last 24 hours.
              </p>
            )}
          </div>
        </div>
      </section>

      <RequestInspectorModal
        capture={inspectedCapture}
        onClose={() => setInspectedCapture(null)}
        onReplay={handleOpenReplay}
      />

      <ReplayDialog
        isOpen={showReplay}
        onClose={() => setShowReplay(false)}
        initialInput={
          replayInitial
            ? {
                method: replayInitial.method,
                url: `http://localhost:8080${replayInitial.path}`,
                request_body: replayInitial.requestBody,
              }
            : undefined
        }
        onExecute={(input) => replayMutation.mutateAsync(input)}
        isExecuting={replayMutation.isPending}
      />
    </div>
  );
}

function methodColor(method?: string) {
  switch (method?.toUpperCase()) {
    case 'POST':
      return 'bg-emerald-500/20 text-emerald-400';
    case 'PUT':
      return 'bg-amber-500/20 text-amber-400';
    case 'DELETE':
      return 'bg-rose-500/20 text-rose-400';
    case 'PATCH':
      return 'bg-purple-500/20 text-purple-400';
    default:
      return 'bg-blue-500/20 text-blue-400';
  }
}

function statusColor(status?: number) {
  return status && status >= 500
    ? 'bg-rose-500/20 text-rose-300'
    : status && status >= 400
      ? 'bg-amber-500/20 text-amber-300'
      : 'bg-emerald-500/20 text-emerald-300';
}
