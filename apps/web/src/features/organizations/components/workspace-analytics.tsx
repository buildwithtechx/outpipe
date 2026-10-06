import { useState } from 'react';
import { Button } from '#/components/ui/button';
import { useObservabilityStats } from '#/features/observability/hooks/use-observability';

export function WorkspaceAnalytics({
  organizationId,
}: {
  organizationId: string;
}) {
  const [range, setRange] = useState('24h');
  const [metric, setMetric] = useState<'requests' | 'errors' | 'bytes'>(
    'requests',
  );
  const query = useObservabilityStats(organizationId, range);
  const data = query.data;
  const occurrences = new Map<string, number>();
  const points = (data?.chartData ?? []).map((point) => {
    const occurrence = occurrences.get(point.time) ?? 0;
    occurrences.set(point.time, occurrence + 1);
    return { ...point, id: `${point.time}:${occurrence}` };
  });
  const max = Math.max(1, ...points.map((point) => point[metric]));
  const format = (value: number) =>
    metric === 'bytes'
      ? `${(value / 1024).toFixed(1)} KiB`
      : new Intl.NumberFormat().format(value);
  return (
    <section
      aria-labelledby="traffic-heading"
      className="space-y-5 rounded-2xl border border-border bg-card p-5 sm:p-6"
    >
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h2 id="traffic-heading" className="text-lg font-semibold">
            Traffic & performance
          </h2>
          <p className="mt-1 text-sm text-muted-foreground">
            Totals include captures and telemetry spans. The chart shows
            captured HTTP requests; transfer measures captured body bytes.
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <label htmlFor="traffic-range" className="sr-only">
            Traffic time range
          </label>
          <select
            id="traffic-range"
            value={range}
            onChange={(event) => setRange(event.target.value)}
            className="min-h-11 rounded-lg border border-border bg-background px-3 text-sm"
          >
            <option value="1h">Last hour</option>
            <option value="24h">Last 24 hours</option>
            <option value="7d">Last 7 days</option>
            <option value="30d">Last 30 days</option>
          </select>
          <Button
            variant="outline"
            disabled={query.isFetching}
            onClick={() => void query.refetch()}
          >
            {query.isFetching ? 'Refreshing…' : 'Refresh'}
          </Button>
        </div>
      </div>
      {query.isLoading ? (
        <p role="status" className="py-8 text-sm text-muted-foreground">
          Loading traffic…
        </p>
      ) : query.isError ? (
        <p role="alert" className="py-8 text-sm text-rose-300">
          Traffic could not be loaded. Use Refresh to try again.
        </p>
      ) : (
        data && (
          <>
            <div className="grid gap-4 sm:grid-cols-3">
              {[
                [
                  'Requests & spans',
                  new Intl.NumberFormat().format(data.totalRequests),
                ],
                [
                  'Captured transfer',
                  `${(data.totalBytes / 1024).toFixed(1)} KiB`,
                ],
                ['P95 latency', `${data.p95LatencyMs} ms`],
              ].map(([label, value]) => (
                <div
                  key={label}
                  className="rounded-xl border border-border p-4"
                >
                  <p className="text-xs text-muted-foreground">{label}</p>
                  <p className="mt-3 text-2xl font-semibold">{value}</p>
                </div>
              ))}
            </div>
            <fieldset
              aria-label="Chart metric"
              className="flex flex-wrap gap-2"
            >
              {(['requests', 'errors', 'bytes'] as const).map((value) => (
                <Button
                  key={value}
                  variant={metric === value ? 'secondary' : 'ghost'}
                  aria-pressed={metric === value}
                  onClick={() => setMetric(value)}
                >
                  {value === 'bytes'
                    ? 'Transfer'
                    : value === 'errors'
                      ? 'Errors'
                      : 'Requests'}
                </Button>
              ))}
            </fieldset>
            {points.some((point) => point[metric] > 0) ? (
              <figure>
                <figcaption className="sr-only">
                  {metric} over {range}. Exact values are available in the table
                  below.
                </figcaption>
                <div
                  className="flex h-40 items-end gap-1 overflow-hidden border-b border-border"
                  aria-hidden="true"
                >
                  {points.map((point) => (
                    <div
                      key={point.id}
                      title={`${point.time}: ${format(point[metric])}`}
                      className="min-w-0 flex-1 rounded-t bg-indigo-300/70"
                      style={{
                        height: `${Math.max(1, (point[metric] / max) * 100)}%`,
                      }}
                    />
                  ))}
                </div>
                <div className="mt-2 flex justify-between gap-3 text-xs text-muted-foreground">
                  <span>{points[0]?.time}</span>
                  <span>{points.at(-1)?.time}</span>
                </div>
              </figure>
            ) : (
              <p className="rounded-xl border border-dashed border-border p-6 text-sm text-muted-foreground">
                No {metric === 'bytes' ? 'transfer' : metric} recorded in this
                time range. Enable capture on a tunnel to inspect its HTTP
                traffic.
              </p>
            )}
            <details>
              <summary className="cursor-pointer text-sm text-muted-foreground">
                View exact values
              </summary>
              <div className="mt-3 max-h-64 overflow-auto">
                <table className="w-full text-left text-sm">
                  <caption className="sr-only">
                    Traffic samples for the selected time range
                  </caption>
                  <thead>
                    <tr>
                      {['Time', 'Requests', 'Errors', 'Bytes'].map((label) => (
                        <th
                          key={label}
                          scope="col"
                          className="p-2 text-muted-foreground"
                        >
                          {label}
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {points.map((point) => (
                      <tr key={point.id} className="border-t border-border">
                        <td className="p-2">{point.time}</td>
                        <td className="p-2">{point.requests}</td>
                        <td className="p-2">{point.errors}</td>
                        <td className="p-2">{point.bytes}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </details>
            <p className="text-xs text-muted-foreground" role="status">
              {query.isFetching
                ? 'Updating…'
                : `Updated ${new Date(query.dataUpdatedAt).toLocaleTimeString()}`}
            </p>
          </>
        )
      )}
    </section>
  );
}
