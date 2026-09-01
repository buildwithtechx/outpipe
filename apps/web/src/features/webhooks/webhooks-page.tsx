import { useState } from 'react';
import { Button } from '#/components/ui/button';
import { Input } from '#/components/ui/input';
import { Label } from '#/components/ui/label';
import { useOrganization } from '#/features/organizations/hooks/use-organization';
import { useWebhookMutations } from './hooks/use-webhook-mutations';
import { useWebhooks } from './hooks/use-webhooks';

export function WebhooksPage({ orgSlug }: { orgSlug: string }) {
  const organizationQuery = useOrganization(orgSlug);
  const query = useWebhooks(organizationQuery.organization?.id);
  const mutations = useWebhookMutations(organizationQuery.organization?.id);
  const [name, setName] = useState('');
  const [url, setUrl] = useState('');
  if (organizationQuery.isLoading || query.isLoading)
    return <p className="p-8 text-sm text-white/55">Loading webhooks…</p>;
  if (
    organizationQuery.isError ||
    query.isError ||
    !organizationQuery.organization
  )
    return (
      <p className="p-8 text-sm text-rose-200">We could not load webhooks.</p>
    );
  const webhooks = query.data ?? [];
  return (
    <main className="mx-auto w-full max-w-6xl px-6 py-12 text-white sm:px-8 lg:py-16">
      <header className="border-b border-white/10 pb-8">
        <p className="mb-3 text-sm text-indigo-200">
          {organizationQuery.organization.name}
        </p>
        <h1 className="text-3xl font-semibold tracking-tight">Webhooks</h1>
        <p className="mt-3 text-sm text-white/55">
          Send workspace lifecycle events to the tools your team already uses.
        </p>
      </header>
      <section className="mt-8 rounded-2xl border border-indigo-300/20 bg-indigo-300/[0.04] p-6">
        <h2 className="text-lg font-medium">Add an endpoint</h2>
        <form
          className="mt-4 grid gap-4 sm:grid-cols-2"
          onSubmit={(event) => {
            event.preventDefault();
            mutations.create.mutate({
              name,
              url,
              events: [
                'tunnel.connected',
                'tunnel.disconnected',
                'tunnel.revoked',
              ],
            });
          }}
        >
          <div className="grid gap-2">
            <Label htmlFor="webhook-name">Name</Label>
            <Input
              id="webhook-name"
              required
              value={name}
              onChange={(event) => setName(event.target.value)}
              placeholder="Deploy notifications"
              className="rounded-xl border-white/10 bg-black/40 text-white"
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="webhook-url">Endpoint URL</Label>
            <Input
              id="webhook-url"
              required
              type="url"
              value={url}
              onChange={(event) => setUrl(event.target.value)}
              placeholder="https://example.com/hooks"
              className="rounded-xl border-white/10 bg-black/40 text-white"
            />
          </div>
          <Button
            type="submit"
            className="sm:col-span-2 sm:justify-self-start"
            disabled={mutations.create.isPending}
          >
            {mutations.create.isPending ? 'Adding…' : 'Add webhook'}
          </Button>
        </form>
        {mutations.create.data?.secret && (
          <p className="mt-4 rounded-xl bg-black/40 p-3 font-mono text-xs text-emerald-200">
            Signing secret: {mutations.create.data.secret}
          </p>
        )}
      </section>
      <section className="mt-8 overflow-hidden rounded-2xl border border-white/10 bg-white/[0.025]">
        {webhooks.length ? (
          webhooks.map((webhook) => (
            <div
              key={webhook.id}
              className="flex flex-wrap items-center justify-between gap-4 border-b border-white/5 px-5 py-4 last:border-0"
            >
              <div>
                <p className="font-medium">{webhook.name}</p>
                <p className="mt-1 font-mono text-xs text-white/45">
                  {webhook.url}
                </p>
              </div>
              <button
                type="button"
                className="text-xs text-rose-200 hover:text-rose-100"
                onClick={() => mutations.remove.mutate(webhook.id)}
                disabled={mutations.remove.isPending}
              >
                Remove
              </button>
            </div>
          ))
        ) : (
          <p className="px-5 py-12 text-center text-sm text-white/50">
            No webhook endpoints configured.
          </p>
        )}
      </section>
    </main>
  );
}
