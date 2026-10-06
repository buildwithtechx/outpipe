import { Link } from '@tanstack/react-router';
import { useCaptures } from '#/features/observability/hooks/use-observability';
import type { Tunnel } from '#/interfaces/tunnel';

export function TunnelRequestsPanel({
  tunnel,
  orgSlug,
}: {
  tunnel: Tunnel;
  orgSlug: string;
}) {
  const query = useCaptures(tunnel.organizationId, tunnel.id);
  return (
    <section className="space-y-4">
      <div>
        <h2 className="font-semibold">Recent captured requests</h2>
        <p className="mt-2 text-sm text-muted-foreground">
          Capture must be enabled to inspect HTTP traffic. Capture can include
          sensitive data; only enable it when needed.
        </p>
      </div>
      {query.isLoading ? (
        <p role="status">Loading requests…</p>
      ) : query.isError ? (
        <div role="alert">
          <p>Could not load requests.</p>
          <button
            type="button"
            onClick={() => void query.refetch()}
            className="mt-2 underline"
          >
            Try again
          </button>
        </div>
      ) : query.data?.length ? (
        <div className="overflow-x-auto rounded-xl border border-border">
          <table className="w-full text-left text-sm">
            <caption className="sr-only">
              Recent captured requests for this tunnel
            </caption>
            <thead>
              <tr>
                {['Method', 'Path', 'Status', 'Latency'].map((label) => (
                  <th
                    key={label}
                    scope="col"
                    className="p-3 font-medium text-muted-foreground"
                  >
                    {label}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {query.data.map((capture) => (
                <tr key={capture.id} className="border-t border-border">
                  <td className="p-3">{capture.method}</td>
                  <td className="max-w-xs truncate p-3 font-mono">
                    {capture.path}
                  </td>
                  <td className="p-3">{capture.statusCode}</td>
                  <td className="p-3">{capture.durationMs} ms</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <p className="rounded-xl border border-dashed border-border p-6 text-sm text-muted-foreground">
          No captured requests yet. Enable capture in Settings and send a
          request to the public endpoint.
        </p>
      )}
      <Link
        to="/$orgSlug/observability"
        params={{ orgSlug }}
        className="inline-block text-sm text-indigo-200 underline"
      >
        Open request inspector and replay
      </Link>
    </section>
  );
}
