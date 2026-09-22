import { Copy, KeyRound } from 'lucide-react';
import { useState } from 'react';
import { Button } from '#/components/ui/button';
import { useAuthSession } from '#/features/auth/hooks/use-auth-session';
import { useMembers } from '#/features/organizations/hooks/use-members';
import { useOrganization } from '#/features/organizations/hooks/use-organization';
import { useApiKeyMutations } from './hooks/use-api-key-mutations';
import { useApiKeys } from './hooks/use-api-keys';

export function ApiKeysPage({ orgSlug }: { orgSlug: string }) {
  const organizationQuery = useOrganization(orgSlug);
  const organizationId = organizationQuery.organization?.id;

  const query = useApiKeys(organizationId);
  const mutations = useApiKeyMutations(organizationId);
  const membersQuery = useMembers(organizationId);
  const { user } = useAuthSession();

  const [newlyCreatedKey, setNewlyCreatedKey] = useState<{
    name: string;
    token: string;
  } | null>(null);
  const [copied, setCopied] = useState(false);

  if (
    organizationQuery.isLoading ||
    query.isLoading ||
    membersQuery.isLoading
  ) {
    return <p className="p-8 text-sm text-white/55">Loading API keys…</p>;
  }

  if (
    organizationQuery.isError ||
    query.isError ||
    !organizationQuery.organization
  ) {
    return (
      <p className="p-8 text-sm text-rose-200">We could not load API keys.</p>
    );
  }

  const organization = organizationQuery.organization;
  const currentMember = membersQuery.data?.find((m) => m.userId === user?.id);
  const isOwner =
    organization.ownerId === user?.id || currentMember?.role === 'owner';
  const isAdmin = isOwner || currentMember?.role === 'admin';

  const createKey = () => {
    const name = window.prompt('Name this API key');
    if (!name?.trim()) return;

    mutations.create.mutate(
      {
        name: name.trim(),
        scopes: ['tunnels:read'],
      },
      {
        onSuccess: (data) => {
          setNewlyCreatedKey({
            name: data.key.name,
            token: data.token,
          });
        },
      },
    );
  };

  const copyToken = () => {
    if (!newlyCreatedKey?.token) return;
    void navigator.clipboard.writeText(newlyCreatedKey.token);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="w-full max-w-6xl space-y-6 pb-12 text-white">
      <header className="border-b border-white/10 pb-6">
        <p className="mb-2 text-xs font-semibold uppercase tracking-wider text-indigo-400">
          {organization.name}
        </p>
        <h1 className="text-3xl font-semibold tracking-[-0.04em]">API keys</h1>
        <div className="mt-2 flex flex-wrap items-center justify-between gap-4">
          <p className="text-sm text-white/55">
            Credentials used by tools and automation to access this workspace.
          </p>
          {isAdmin ? (
            <Button
              type="button"
              onClick={createKey}
              disabled={mutations.create.isPending}
            >
              {mutations.create.isPending ? 'Creating…' : 'Create API key'}
            </Button>
          ) : (
            <span className="text-xs text-white/40">
              Only workspace administrators can generate API keys.
            </span>
          )}
        </div>
      </header>

      {/* Creation Error Feedback */}
      {mutations.create.isError && (
        <div className="rounded-xl border border-rose-500/20 bg-rose-500/10 p-4 text-xs text-rose-300">
          Failed to create API key:{' '}
          {mutations.create.error instanceof Error
            ? mutations.create.error.message
            : 'An error occurred'}
        </div>
      )}

      {/* Newly Created Key Alert Banner */}
      {newlyCreatedKey && (
        <div className="rounded-2xl border border-emerald-500/30 bg-emerald-950/20 p-5 shadow-lg">
          <div className="flex items-start justify-between gap-4">
            <div className="space-y-1">
              <div className="flex items-center gap-2 text-sm font-semibold text-emerald-300">
                <KeyRound className="size-4" />
                <span>API Key Generated: {newlyCreatedKey.name}</span>
              </div>
              <p className="text-xs text-white/70">
                Copy this key now. For security purposes,{' '}
                <strong className="text-white">
                  it will never be displayed again
                </strong>
                .
              </p>
            </div>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={copyToken}
              className="shrink-0 border-emerald-400/40 bg-emerald-500/20 text-emerald-200 hover:bg-emerald-500/30"
            >
              <Copy className="mr-1.5 size-3.5" />
              {copied ? 'Copied!' : 'Copy Key'}
            </Button>
          </div>
          <div className="mt-3 overflow-x-auto rounded-xl border border-white/10 bg-black/60 p-3">
            <code className="font-mono text-xs text-emerald-200 select-all">
              {newlyCreatedKey.token}
            </code>
          </div>
        </div>
      )}

      <section className="mt-8 overflow-hidden rounded-2xl border border-white/10 bg-white/[0.025]">
        {query.data?.length ? (
          query.data.map((key) => (
            <div
              key={key.id}
              className="flex flex-wrap items-center justify-between gap-4 border-b border-white/5 px-5 py-4 last:border-0"
            >
              <div>
                <p className="font-medium">{key.name}</p>
                <p className="mt-1 font-mono text-sm text-indigo-200">
                  {key.prefix}••••••••
                </p>
              </div>
              <div className="text-right text-xs text-white/45">
                <p>{key.revokedAt ? 'Revoked' : 'Active'}</p>
                <p className="mt-1">{key.scopes}</p>
                {!key.revokedAt && isAdmin && (
                  <button
                    type="button"
                    className="mt-2 text-rose-200 hover:text-rose-100"
                    onClick={() => mutations.revoke.mutate(key.id)}
                    disabled={mutations.revoke.isPending}
                  >
                    Revoke
                  </button>
                )}
              </div>
            </div>
          ))
        ) : (
          <p className="px-5 py-12 text-center text-sm text-white/50">
            No organization API keys found.
          </p>
        )}
      </section>
    </div>
  );
}
