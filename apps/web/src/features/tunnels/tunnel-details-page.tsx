import { Link } from '@tanstack/react-router';
import { ArrowLeft, CircleAlert } from 'lucide-react';
import { Button } from '#/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '#/components/ui/tabs';
import { TunnelCaptureControl } from './components/tunnel-capture-control';
import { TunnelConfigurationForm } from './components/tunnel-configuration-form';
import { TunnelConnectGuide } from './components/tunnel-connect-guide';
import { TunnelDetailActions } from './components/tunnel-detail-actions';
import { TunnelDetailCard } from './components/tunnel-detail-card';
import { TunnelPageState } from './components/tunnel-page-state';
import { TunnelRequestsPanel } from './components/tunnel-requests-panel';
import { TunnelStatusBadge } from './components/tunnel-status-badge';
import { useTunnel } from './hooks/use-tunnel';

type TunnelDetailsPageProps = {
  orgSlug: string;
  tunnelID: string;
};

export function TunnelDetailsPage({
  orgSlug,
  tunnelID,
}: TunnelDetailsPageProps) {
  const tunnelQuery = useTunnel(tunnelID);

  if (tunnelQuery.isLoading) {
    return <TunnelPageState label="Loading tunnel details…" />;
  }

  if (!tunnelQuery.data) {
    return (
      <main className="mx-auto flex min-h-[55vh] max-w-xl flex-col items-center justify-center px-6 text-center text-white">
        <CircleAlert className="size-5 text-rose-200" />
        <h1 className="mt-4 text-xl font-semibold">Tunnel unavailable</h1>
        <p className="mt-2 text-sm leading-6 text-white/55">
          This tunnel could not be found, or your workspace no longer has access
          to it.
        </p>
        <Button
          asChild
          variant="outline"
          className="mt-6 border-white/10 bg-transparent text-white hover:bg-white/10 hover:text-white"
        >
          <Link to="/$orgSlug/tunnels" params={{ orgSlug }}>
            <ArrowLeft />
            Back to tunnels
          </Link>
        </Button>
      </main>
    );
  }

  const { data: tunnel } = tunnelQuery;

  return (
    <div className="w-full max-w-6xl space-y-6 pb-12 text-white">
      <Link
        to="/$orgSlug/tunnels"
        params={{ orgSlug }}
        className="inline-flex items-center gap-2 text-sm text-white/50 transition-colors hover:text-white focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-indigo-300"
      >
        <ArrowLeft className="size-4" />
        All tunnels
      </Link>
      <header className="mt-2 flex flex-col gap-6 border-b border-white/10 pb-6 lg:flex-row lg:items-end lg:justify-between">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-3">
            <h1 className="truncate text-3xl font-semibold tracking-[-0.04em] sm:text-4xl">
              {tunnel.name}
            </h1>
            <TunnelStatusBadge status={tunnel.status} />
          </div>
          <p className="mt-2 font-mono text-sm text-indigo-200">
            {tunnel.publicHostname}
          </p>
        </div>
        <TunnelDetailActions tunnel={tunnel} />
      </header>
      <Tabs defaultValue="overview" className="space-y-6">
        <TabsList aria-label="Tunnel views">
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="connection">Connection</TabsTrigger>
          {(tunnel.protocol === 'http' || tunnel.protocol === 'https') && (
            <TabsTrigger value="requests">Requests</TabsTrigger>
          )}
          <TabsTrigger value="settings">Settings</TabsTrigger>
        </TabsList>
        <TabsContent value="overview" className="space-y-6">
          <TunnelConnectGuide tunnel={tunnel} />
          <TunnelDetailCard tunnel={tunnel} />
        </TabsContent>
        <TabsContent value="connection">
          <TunnelConnectGuide tunnel={tunnel} />
        </TabsContent>
        <TabsContent value="requests">
          <TunnelRequestsPanel tunnel={tunnel} orgSlug={orgSlug} />
        </TabsContent>
        <TabsContent
          value="settings"
          forceMount
          className="space-y-6 data-[state=inactive]:hidden"
        >
          <TunnelConfigurationForm tunnel={tunnel} />
          {(tunnel.protocol === 'http' || tunnel.protocol === 'https') && (
            <TunnelCaptureControl tunnel={tunnel} />
          )}
        </TabsContent>
      </Tabs>
    </div>
  );
}
