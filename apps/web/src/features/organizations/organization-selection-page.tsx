import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Link, useNavigate } from '@tanstack/react-router';
import { ArrowRight, Check, LoaderCircle } from 'lucide-react';
import { useEffect, useState } from 'react';
import { AuthPageShell } from '#/features/auth/components/auth-page-shell';
import {
  checkOrganizationSlug,
  createOrganization,
  getOrganizations,
} from '#/features/auth/services/auth-service';

const slugPattern = /^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/;

function rememberOrganization(slug: string) {
  window.localStorage.setItem('outpipe_last_organization', slug);
}

function slugify(value: string) {
  return value
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .slice(0, 63);
}

function organizationInitial(name: string) {
  return name.trim().charAt(0).toUpperCase() || '?';
}

export function OrganizationSelectionPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [showCreate, setShowCreate] = useState(false);
  const [name, setName] = useState('');
  const [slug, setSlug] = useState('');
  const [generatedSlug, setGeneratedSlug] = useState('');
  const organizationsQuery = useQuery({
    queryKey: ['organizations'],
    queryFn: getOrganizations,
  });
  const slugQuery = useQuery({
    queryKey: ['organization-slug', slug],
    queryFn: () => checkOrganizationSlug(slug),
    enabled: false,
    retry: false,
  });
  const createMutation = useMutation({
    mutationFn: () => createOrganization(name.trim(), slug),
    onSuccess: async (organization) => {
      rememberOrganization(organization.slug);
      await queryClient.invalidateQueries({ queryKey: ['organizations'] });
      await navigate({
        to: '/$orgSlug',
        params: { orgSlug: organization.slug },
      });
    },
  });

  useEffect(() => {
    if (!showCreate || !slugPattern.test(slug)) return;
    const timeout = window.setTimeout(() => {
      void slugQuery.refetch();
    }, 350);
    return () => window.clearTimeout(timeout);
  }, [showCreate, slug, slugQuery.refetch]);

  const organizations = organizationsQuery.data ?? [];
  const slugError =
    slug && !slugPattern.test(slug)
      ? 'Use lowercase letters, numbers, and single hyphens.'
      : null;
  const canCreate = Boolean(
    name.trim() && slugPattern.test(slug) && slugQuery.data?.available,
  );

  return (
    <AuthPageShell
      title={showCreate ? 'Create a workspace' : 'Choose a workspace'}
      description={
        showCreate
          ? 'Give your team a clear home for tunnels and environments.'
          : 'Select the organization you want to open.'
      }
      footer={null}
    >
      {organizationsQuery.isLoading && (
        <p className="text-sm text-white/45">Loading workspaces…</p>
      )}
      {organizationsQuery.isError && (
        <p className="text-sm leading-6 text-rose-200">
          We could not load your workspaces. Please refresh and try again.
        </p>
      )}
      {!organizationsQuery.isLoading && !showCreate && (
        <div className="space-y-3 text-left">
          {organizations.length === 0 && (
            <p className="mb-5 text-sm leading-6 text-white/55">
              You do not belong to a workspace yet. Create one to start opening
              tunnels.
            </p>
          )}
          {organizations.map((organization) => (
            <Link
              key={organization.id}
              to="/$orgSlug"
              params={{ orgSlug: organization.slug }}
              onClick={() => rememberOrganization(organization.slug)}
              className="group flex items-center justify-between rounded-xl border border-white/10 bg-[#111] px-4 py-4 text-left transition-colors hover:border-indigo-300/45 hover:bg-indigo-300/10 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-300"
            >
              <span className="flex min-w-0 items-center gap-3">
                <span className="flex size-10 shrink-0 items-center justify-center rounded-lg border border-white/10 bg-white/5 text-sm font-semibold text-indigo-200 transition-colors group-hover:border-indigo-300/40">
                  {organizationInitial(organization.name)}
                </span>
                <span className="min-w-0">
                  <span className="block truncate font-medium text-white">
                    {organization.name}
                  </span>
                  <span className="mt-1 block truncate font-mono text-xs text-white/40">
                    {organization.slug}
                  </span>
                </span>
              </span>
              <ArrowRight className="size-4 shrink-0 text-white/35 transition-transform group-hover:translate-x-1 group-hover:text-indigo-200" />
            </Link>
          ))}
        </div>
      )}
      {!organizationsQuery.isLoading && showCreate && (
        <form
          className="space-y-4 text-left"
          onSubmit={(event) => {
            event.preventDefault();
            if (canCreate) createMutation.mutate();
          }}
        >
          <label className="block space-y-2 text-sm text-white/65">
            Workspace name
            <input
              value={name}
              onChange={(event) => {
                const value = event.target.value;
                setName(value);
                if (!slug || slug === generatedSlug) {
                  const nextSlug = slugify(value);
                  setSlug(nextSlug);
                  setGeneratedSlug(nextSlug);
                }
              }}
              placeholder="Acme Labs"
              maxLength={120}
              required
              className="h-12 w-full rounded-xl border border-white/10 bg-[#111] px-4 text-sm text-white outline-none placeholder:text-white/30 focus:border-indigo-300/60 focus:ring-2 focus:ring-indigo-300/20"
            />
          </label>
          <label className="block space-y-2 text-sm text-white/65">
            Workspace slug
            <span className="relative block">
              <input
                value={slug}
                onChange={(event) => {
                  setSlug(event.target.value.toLowerCase());
                  setGeneratedSlug('');
                }}
                placeholder="acme-labs"
                maxLength={63}
                required
                className="h-12 w-full rounded-xl border border-white/10 bg-[#111] px-4 font-mono text-sm text-white outline-none placeholder:text-white/30 focus:border-indigo-300/60 focus:ring-2 focus:ring-indigo-300/20"
              />
              {slugQuery.data?.available && (
                <Check className="absolute right-4 top-1/2 size-4 -translate-y-1/2 text-emerald-300" />
              )}
              {slugQuery.isFetching && (
                <LoaderCircle className="absolute right-4 top-1/2 size-4 -translate-y-1/2 animate-spin text-white/40" />
              )}
            </span>
          </label>
          {(slugError || slugQuery.data?.available === false) && (
            <p className="text-sm text-rose-200">
              {slugError ?? 'That slug is already in use.'}
            </p>
          )}
          {createMutation.isError && (
            <p className="text-sm text-rose-200">
              We could not create that workspace. Try another slug.
            </p>
          )}
          <button
            type="submit"
            disabled={createMutation.isPending || !canCreate}
            className="h-12 w-full rounded-xl bg-indigo-300 px-4 text-sm font-semibold text-black transition-colors hover:bg-indigo-200 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-300 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {createMutation.isPending
              ? 'Creating workspace…'
              : 'Create workspace'}
          </button>
        </form>
      )}
      {!organizationsQuery.isLoading && (
        <button
          type="button"
          onClick={() => setShowCreate((value) => !value)}
          className="mt-6 text-sm text-white/45 transition-colors hover:text-indigo-200 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-indigo-300"
        >
          {showCreate ? 'Back to workspaces' : 'Create a new workspace'}
        </button>
      )}
    </AuthPageShell>
  );
}
