import { useState } from 'react';
import { Button } from '#/components/ui/button';
import { Input } from '#/components/ui/input';
import { Label } from '#/components/ui/label';
import { useOrganization } from '#/features/organizations/hooks/use-organization';
import { useDomainMutations } from './hooks/use-domain-mutations';
import { useDomains } from './hooks/use-domains';

export function DomainsPage({ orgSlug }: { orgSlug: string }) {
  const organizationQuery = useOrganization(orgSlug);
  const organizationId = organizationQuery.organization?.id;

  const query = useDomains(organizationId);
  const mutations = useDomainMutations(organizationId);

  const [hostname, setHostname] = useState('');
  const [token, setToken] = useState('');

  if (organizationQuery.isLoading || query.isLoading) {
    return <p className="p-8 text-sm text-white/55">Loading domains…</p>;
  }

  if (
    organizationQuery.isError ||
    query.isError ||
    !organizationQuery.organization
  ) {
    return (
      <p className="p-8 text-sm text-rose-200">We could not load domains.</p>
    );
  }

  const domains = query.data ?? [];

  return (
    <main className="mx-auto w-full max-w-6xl px-6 py-12 text-white sm:px-8 lg:py-16">
      <header className="border-b border-white/10 pb-8">
        <p className="mb-3 text-sm text-indigo-200">
          {organizationQuery.organization.name}
        </p>
        <h1 className="text-3xl font-semibold tracking-tight">Domains</h1>
        <p className="mt-3 text-sm text-white/55">
          Stable hostnames for previews and production-like callbacks.
        </p>
      </header>
      <section className="mt-8 rounded-2xl border border-indigo-300/20 bg-indigo-300/[0.04] p-6">
        <h2 className="text-lg font-medium">Connect a domain</h2>
        <form
          className="mt-4 grid gap-4 sm:grid-cols-[1fr_auto]"
          onSubmit={(event) => {
            event.preventDefault();
            mutations.create.mutate({ hostname, verificationMethod: 'dns' });
          }}
        >
          <div className="grid gap-2">
            <Label htmlFor="domain-hostname" className="text-white/70">
              Hostname
            </Label>
            <Input
              id="domain-hostname"
              required
              value={hostname}
              onChange={(event) => setHostname(event.target.value)}
              placeholder="api.example.com"
              className="rounded-xl border-white/10 bg-black/40 text-white"
            />
          </div>
          <Button
            type="submit"
            className="self-end"
            disabled={mutations.create.isPending}
          >
            {mutations.create.isPending ? 'Connecting…' : 'Connect domain'}
          </Button>
        </form>
        {mutations.create.data?.verificationToken && (
          <div className="mt-4 grid gap-2">
            <Label htmlFor="domain-token" className="text-white/70">
              DNS verification token
            </Label>
            <Input
              id="domain-token"
              readOnly
              value={mutations.create.data.verificationToken}
              className="font-mono text-white"
            />
            <p className="text-xs text-white/45">
              Add this token to DNS, then verify the domain below.
            </p>
          </div>
        )}
      </section>
      <section className="mt-8 overflow-hidden rounded-2xl border border-white/10 bg-white/[0.025]">
        {domains.length ? (
          domains.map((domain) => (
            <div
              key={domain.id}
              className="flex flex-wrap items-center justify-between gap-4 border-b border-white/5 px-5 py-4 last:border-0"
            >
              <div>
                <p className="font-mono text-sm text-white/85">
                  {domain.hostname}
                </p>
                <p className="mt-1 text-xs text-white/40">
                  {domain.verificationMethod} verification
                </p>
              </div>
              <div className="flex items-center gap-3">
                <span className="rounded-full border border-white/10 px-3 py-1 text-xs text-white/60">
                  {domain.status}
                </span>
                {domain.status === 'pending' && (
                  <form
                    className="flex gap-2"
                    onSubmit={(event) => {
                      event.preventDefault();
                      if (token)
                        mutations.verify.mutate({ domainId: domain.id, token });
                    }}
                  >
                    <Input
                      aria-label={`Verification token for ${domain.hostname}`}
                      value={token}
                      onChange={(event) => setToken(event.target.value)}
                      placeholder="Token"
                      className="h-8 w-28 text-xs"
                    />
                    <Button
                      type="submit"
                      variant="outline"
                      className="h-8 text-xs"
                      disabled={mutations.verify.isPending}
                    >
                      Verify
                    </Button>
                  </form>
                )}
              </div>
            </div>
          ))
        ) : (
          <p className="px-5 py-12 text-center text-sm text-white/50">
            No custom domains connected yet.
          </p>
        )}
      </section>
    </main>
  );
}
