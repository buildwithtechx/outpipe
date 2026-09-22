import { useEffect, useRef, useState } from 'react';
import { Button } from '#/components/ui/button';
import { useMemberMutations } from './hooks/use-member-mutations';
import { useMembers } from './hooks/use-members';
import { useOrganization } from './hooks/use-organization';

export function MembersPage({ orgSlug }: { orgSlug: string }) {
  const organizationQuery = useOrganization(orgSlug);
  const organizationId = organizationQuery.organization?.id;

  const query = useMembers(organizationId);
  const mutations = useMemberMutations(organizationId);
  const [inviteSuccess, setInviteSuccess] = useState<string | null>(null);
  const timeoutRef = useRef<NodeJS.Timeout | null>(null);

  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, []);

  if (organizationQuery.isLoading || query.isLoading) {
    return <p className="p-8 text-sm text-white/55">Loading members…</p>;
  }

  if (
    organizationQuery.isError ||
    query.isError ||
    !organizationQuery.organization
  ) {
    return (
      <p className="p-8 text-sm text-rose-200">
        We could not load workspace members.
      </p>
    );
  }

  const organization = organizationQuery.organization;

  const invite = () => {
    const emailPrompt = window.prompt('Email address to invite:');
    if (emailPrompt === null) return;
    const email = emailPrompt.trim();
    if (!email) return;

    const rolePrompt = window.prompt('Role (member, admin, viewer):', 'member');
    if (rolePrompt === null) return;
    const roleInput = rolePrompt.trim().toLowerCase();
    const role: 'admin' | 'member' | 'viewer' =
      roleInput === 'admin' || roleInput === 'viewer' ? roleInput : 'member';

    mutations.invite.mutate(
      { email, role },
      {
        onSuccess: () => {
          if (timeoutRef.current) {
            clearTimeout(timeoutRef.current);
          }
          setInviteSuccess(`Invitation sent to ${email} as ${role}.`);
          timeoutRef.current = setTimeout(() => {
            setInviteSuccess(null);
            timeoutRef.current = null;
          }, 4000);
        },
      },
    );
  };

  return (
    <div className="w-full max-w-6xl space-y-6 pb-12 text-white">
      <header className="border-b border-white/10 pb-6">
        <p className="mb-2 text-xs font-semibold uppercase tracking-wider text-indigo-400">
          {organization.name}
        </p>
        <h1 className="text-3xl font-semibold tracking-[-0.04em]">Members</h1>
        <div className="mt-2 flex flex-wrap items-center justify-between gap-4">
          <p className="text-sm text-white/55">
            People with access to this workspace.
          </p>
          <Button
            type="button"
            onClick={invite}
            disabled={mutations.invite.isPending}
          >
            {mutations.invite.isPending ? 'Sending…' : 'Invite member'}
          </Button>
        </div>
      </header>

      {/* Action feedback */}
      {inviteSuccess && (
        <div className="rounded-xl border border-emerald-500/30 bg-emerald-500/10 p-4 text-xs text-emerald-300">
          {inviteSuccess}
        </div>
      )}
      {mutations.invite.isError && (
        <div className="rounded-xl border border-rose-500/30 bg-rose-500/10 p-4 text-xs text-rose-300">
          Failed to send invitation:{' '}
          {mutations.invite.error instanceof Error
            ? mutations.invite.error.message
            : 'An error occurred'}
        </div>
      )}
      {mutations.remove.isError && (
        <div className="rounded-xl border border-rose-500/30 bg-rose-500/10 p-4 text-xs text-rose-300">
          Failed to remove member:{' '}
          {mutations.remove.error instanceof Error
            ? mutations.remove.error.message
            : 'An error occurred'}
        </div>
      )}

      <section className="mt-8 overflow-hidden rounded-2xl border border-white/10 bg-white/2.5">
        {query.data?.length ? (
          query.data.map((member) => (
            <div
              key={member.id}
              className="flex items-center justify-between gap-4 border-b border-white/5 px-5 py-4 last:border-0"
            >
              <div>
                <p className="font-mono text-sm text-white/80">
                  {member.userId}
                </p>
                <p className="mt-1 text-xs text-white/40">
                  Joined {formatDate(member.createdAt)}
                </p>
              </div>
              <div className="flex items-center gap-3">
                <span className="rounded-full border border-white/10 px-3 py-1 text-xs text-white/55">
                  {member.role}
                </span>
                {member.role !== 'owner' && (
                  <button
                    type="button"
                    className="text-xs text-rose-200 hover:text-rose-100 disabled:opacity-50"
                    onClick={() => {
                      if (
                        window.confirm('Remove this member from the workspace?')
                      ) {
                        mutations.remove.mutate(member.userId);
                      }
                    }}
                    disabled={mutations.remove.isPending}
                  >
                    {mutations.remove.isPending ? 'Removing…' : 'Remove'}
                  </button>
                )}
              </div>
            </div>
          ))
        ) : (
          <p className="px-5 py-12 text-center text-sm text-white/50">
            No members found.
          </p>
        )}
      </section>
    </div>
  );
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' }).format(
    new Date(value),
  );
}
