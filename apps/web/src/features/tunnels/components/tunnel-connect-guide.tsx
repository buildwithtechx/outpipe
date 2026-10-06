import { CheckCircle2, ExternalLink, Radio } from 'lucide-react';
import { CopyCommand } from '#/components/ui/copy-command';
import type { Tunnel } from '#/interfaces/tunnel';
import {
  publicTunnelEndpoint,
  tunnelConnectCommand,
} from '../lib/tunnel-command';
import { CliInstall } from './cli-install';

export function TunnelConnectGuide({ tunnel }: { tunnel: Tunnel }) {
  const connected = tunnel.status === 'active';
  const endpoint = publicTunnelEndpoint(tunnel);
  if (tunnel.status === 'revoked' || tunnel.status === 'expired')
    return (
      <section className="rounded-2xl border border-border bg-card p-6">
        <h2 className="font-semibold">This endpoint is no longer available</h2>
        <p className="mt-2 text-sm text-muted-foreground">
          Create a new tunnel to connect your service.
        </p>
      </section>
    );
  return (
    <section
      aria-labelledby="connect-heading"
      className="rounded-2xl border border-border bg-card p-5 sm:p-6"
    >
      <div className="mb-6 flex flex-wrap items-center justify-between gap-3">
        <h2 id="connect-heading" className="text-lg font-semibold">
          {connected
            ? 'Relay reports your tunnel connected'
            : 'Connect your service'}
        </h2>
        <p
          role="status"
          aria-live="polite"
          className="flex items-center gap-2 text-sm text-muted-foreground"
        >
          {connected ? (
            <CheckCircle2 className="size-4 text-emerald-300" />
          ) : (
            <Radio className="size-4 text-amber-300" />
          )}
          {connected ? 'Agent connected' : 'Waiting for an agent'}
        </p>
      </div>
      {connected ? (
        <div className="space-y-4">
          <p className="text-sm text-muted-foreground">
            Keep the CLI running while sharing this endpoint. To stop
            forwarding, press Ctrl+C in the terminal that started it.
          </p>
          {endpoint ? (
            <CopyCommand text={endpoint} label="Copy public endpoint" />
          ) : (
            <p className="text-sm text-muted-foreground">
              Waiting for the relay to assign a public port.
            </p>
          )}
          {(tunnel.protocol === 'http' || tunnel.protocol === 'https') && (
            <a
              href={endpoint ?? undefined}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-2 text-sm text-indigo-200"
            >
              Open your app <ExternalLink className="size-4" />
            </a>
          )}
        </div>
      ) : (
        <ol className="space-y-6">
          <li>
            <h3 className="mb-3 text-sm font-medium">1. Install the CLI</h3>
            <CliInstall />
          </li>
          <li>
            <h3 className="mb-3 text-sm font-medium">
              {tunnel.machineOwned
                ? '2. Use the owning machine credential'
                : '2. Sign in to your workspace account'}
            </h3>
            {tunnel.machineOwned ? (
              <p className="text-sm text-muted-foreground">
                Run this tunnel through the machine that owns it, using its
                scoped agent credential. Account login cannot replace a machine
                credential.
              </p>
            ) : (
              <CopyCommand text="outpipe login" />
            )}
          </li>
          <li>
            <h3 className="mb-3 text-sm font-medium">
              3. Run this command on the machine hosting your service
            </h3>
            <CopyCommand text={tunnelConnectCommand(tunnel)} />
            <p className="text-xs text-muted-foreground">
              Keep the command running. This page refreshes relay status every
              five seconds. Use Ctrl+C in the terminal to stop forwarding.
            </p>
          </li>
        </ol>
      )}
    </section>
  );
}
