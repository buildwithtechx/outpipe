import { Activity, ArrowRightLeft, Eye, Layers } from 'lucide-react';
import { useState } from 'react';
import { Button } from '#/components/ui/button';
import { Card } from '#/components/ui/card';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '#/components/ui/select';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '#/components/ui/tabs';
import { useOrganization } from '#/features/organizations/hooks/use-organization';
import type { RequestCapture } from '#/interfaces';
import { LogStream } from './components/log-stream';
import { ReplayDialog } from './components/replay-dialog';
import { RequestInspectorModal } from './components/request-inspector-modal';
import { TraceWaterfall } from './components/trace-waterfall';
import {
  useCaptures,
  useLogs,
  useObservabilityStats,
  useReplayMutation,
  useTraces,
  useTraceWaterfall,
} from './hooks/use-observability';

export function ObservabilityPage({ orgSlug }: { orgSlug: string }) {
  const { organization } = useOrganization(orgSlug);
  const orgId = organization?.id;

  const [timeRange, setTimeRange] = useState('24h');
  const [selectedTraceId, setSelectedTraceId] = useState<string | null>(null);
  const [inspectedCapture, setInspectedCapture] =
    useState<RequestCapture | null>(null);
  const [showReplay, setShowReplay] = useState(false);
  const [replayInitial, setReplayInitial] =
    useState<Partial<RequestCapture> | null>(null);

  const { data: stats } = useObservabilityStats(orgId, timeRange);
  const { data: traces = [] } = useTraces(orgId);
  const { data: logs = [] } = useLogs(orgId);
  const { data: captures = [] } = useCaptures(orgId);
  const { data: waterfallSpans = [] } = useTraceWaterfall(
    orgId,
    selectedTraceId || undefined,
  );
  const replayMutation = useReplayMutation(orgId || '');

  const handleOpenReplay = (capture: RequestCapture) => {
    setReplayInitial(capture);
    setShowReplay(true);
  };

  return (
    <div className="space-y-8 p-6 lg:p-10 max-w-7xl mx-auto text-white">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-white/10 pb-6">
        <div>
          <div className="flex items-center gap-2.5">
            <div className="flex size-9 items-center justify-center rounded-xl bg-indigo-500/10 border border-indigo-500/20 text-indigo-400">
              <Activity className="size-5" />
            </div>
            <h1 className="text-2xl font-bold tracking-tight text-white">
              Observability & Traces
            </h1>
          </div>
          <p className="mt-1 text-sm text-white/60">
            OpenTelemetry distributed tracing, live request captures &
            centralized logging.
          </p>
        </div>
        <div className="flex items-center gap-3">
          <Select value={timeRange} onValueChange={setTimeRange}>
            <SelectTrigger className="w-36 border-white/10 bg-zinc-800 text-xs text-white">
              <SelectValue placeholder="Time range" />
            </SelectTrigger>
            <SelectContent className="border-white/10 bg-zinc-900 text-white">
              <SelectItem value="1h">Past Hour</SelectItem>
              <SelectItem value="24h">Past 24 Hours</SelectItem>
              <SelectItem value="7d">Past 7 Days</SelectItem>
              <SelectItem value="30d">Past 30 Days</SelectItem>
            </SelectContent>
          </Select>
          <Button
            size="sm"
            onClick={() => {
              setReplayInitial(null);
              setShowReplay(true);
            }}
            className="bg-indigo-600 hover:bg-indigo-500 text-xs text-white"
          >
            <ArrowRightLeft className="mr-1.5 size-3.5" />
            Replay Request
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card className="border-white/10 bg-white/5 p-4 text-white">
          <span className="text-xs text-white/50">Total Requests</span>
          <p className="text-2xl font-semibold mt-1">
            {stats?.totalRequests || 0}
          </p>
        </Card>
        <Card className="border-white/10 bg-white/5 p-4 text-white">
          <span className="text-xs text-emerald-400/80">P50 Latency</span>
          <p className="text-2xl font-semibold text-emerald-400 mt-1">
            {stats?.p50LatencyMs || 0}ms
          </p>
        </Card>
        <Card className="border-white/10 bg-white/5 p-4 text-white">
          <span className="text-xs text-amber-400/80">P95 / P99 Latency</span>
          <p className="text-2xl font-semibold text-amber-400 mt-1">
            {stats?.p95LatencyMs || 0} / {stats?.p99LatencyMs || 0}ms
          </p>
        </Card>
        <Card className="border-white/10 bg-white/5 p-4 text-white">
          <span className="text-xs text-cyan-400/80">Data Transferred</span>
          <p className="text-2xl font-semibold text-cyan-400 mt-1">
            {((stats?.totalBytes || 0) / (1024 * 1024)).toFixed(1)} MB
          </p>
        </Card>
      </div>

      {selectedTraceId && (
        <TraceWaterfall
          traceId={selectedTraceId}
          spans={waterfallSpans}
          onClose={() => setSelectedTraceId(null)}
        />
      )}

      <Tabs defaultValue="traces" className="space-y-4">
        <TabsList className="bg-white/5 border border-white/10 p-1">
          <TabsTrigger value="traces" className="text-xs">
            Traces ({traces.length})
          </TabsTrigger>
          <TabsTrigger value="logs" className="text-xs">
            Logs ({logs.length})
          </TabsTrigger>
          <TabsTrigger value="captures" className="text-xs">
            Tunnel Request Captures ({captures.length})
          </TabsTrigger>
        </TabsList>

        <TabsContent value="traces" className="mt-0">
          <div className="rounded-xl border border-white/10 bg-white/5 overflow-hidden">
            {traces.length === 0 ? (
              <div className="p-12 text-center text-white/40 text-xs">
                No distributed traces received yet. Configure your app with
                OpenTelemetry OTLP endpoint:
                <br />
                <code className="text-indigo-400 mt-2 inline-block font-mono bg-white/5 px-2 py-1 rounded">
                  POST /api/v1/ingest/otlp/v1/traces
                </code>
              </div>
            ) : (
              <table className="w-full text-left text-xs text-white/70">
                <thead className="bg-white/5 text-white/50 border-b border-white/10">
                  <tr>
                    <th className="px-4 py-3 font-medium">Trace Root</th>
                    <th className="px-4 py-3 font-medium">Trace ID</th>
                    <th className="px-4 py-3 font-medium">Duration</th>
                    <th className="px-4 py-3 font-medium">Timestamp</th>
                    <th className="px-4 py-3 font-medium text-right">
                      Actions
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-white/5">
                  {traces.map((t) => (
                    <tr
                      key={t.id}
                      className="hover:bg-white/5 transition-colors"
                    >
                      <td className="px-4 py-3 font-semibold text-white">
                        {t.name}
                      </td>
                      <td className="px-4 py-3 font-mono text-white/50 text-[11px]">
                        {t.traceId}
                      </td>
                      <td className="px-4 py-3 font-mono text-emerald-400 font-semibold">
                        {t.durationMs}ms
                      </td>
                      <td className="px-4 py-3 text-white/40">
                        {new Date(t.startTime).toLocaleTimeString()}
                      </td>
                      <td className="px-4 py-3 text-right">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => setSelectedTraceId(t.traceId)}
                          className="h-7 text-xs border-white/10 bg-white/5 text-white/80"
                        >
                          <Layers className="size-3 mr-1" />
                          Waterfall
                        </Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </TabsContent>

        <TabsContent value="logs" className="mt-0">
          <LogStream logs={logs} onSelectTrace={setSelectedTraceId} />
        </TabsContent>

        <TabsContent value="captures" className="mt-0">
          <div className="rounded-xl border border-white/10 bg-white/5 overflow-hidden">
            {captures.length === 0 ? (
              <div className="p-12 text-center text-white/40 text-xs">
                No tunnel requests captured yet.
              </div>
            ) : (
              <table className="w-full text-left text-xs text-white/70">
                <thead className="bg-white/5 text-white/50 border-b border-white/10">
                  <tr>
                    <th className="px-4 py-3 font-medium">Status</th>
                    <th className="px-4 py-3 font-medium">Method & Path</th>
                    <th className="px-4 py-3 font-medium">Duration</th>
                    <th className="px-4 py-3 font-medium">Payload Size</th>
                    <th className="px-4 py-3 font-medium">Time</th>
                    <th className="px-4 py-3 font-medium text-right">
                      Actions
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-white/5">
                  {captures.map((c) => (
                    <tr
                      key={c.id}
                      className="hover:bg-white/5 transition-colors"
                    >
                      <td className="px-4 py-3">
                        <span
                          className={`px-2 py-0.5 rounded text-[11px] font-bold font-mono ${
                            c.statusCode < 400
                              ? 'bg-emerald-500/15 text-emerald-400'
                              : 'bg-rose-500/15 text-rose-400'
                          }`}
                        >
                          {c.statusCode}
                        </span>
                      </td>
                      <td className="px-4 py-3 font-mono">
                        <strong className="text-white mr-1.5">
                          {c.method}
                        </strong>
                        <span className="text-white/60">{c.path}</span>
                      </td>
                      <td className="px-4 py-3 font-mono">{c.durationMs}ms</td>
                      <td className="px-4 py-3 font-mono text-white/50">
                        {c.requestBodySize} B
                      </td>
                      <td className="px-4 py-3 text-white/40">
                        {new Date(c.timestamp).toLocaleTimeString()}
                      </td>
                      <td className="px-4 py-3 text-right space-x-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => setInspectedCapture(c)}
                          className="h-7 text-xs border-white/10 bg-white/5 text-white/80"
                        >
                          <Eye className="size-3 mr-1" />
                          Inspect
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleOpenReplay(c)}
                          className="h-7 text-xs border-indigo-500/30 text-indigo-300 hover:bg-indigo-500/10"
                        >
                          <ArrowRightLeft className="size-3 mr-1" />
                          Replay
                        </Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </TabsContent>
      </Tabs>

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
                url: '',
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
