import {
  Copy,
  Eye,
  EyeOff,
  FolderPlus,
  KeyRound,
  Lock,
  Plus,
  Terminal,
  Trash2,
} from 'lucide-react';
import { useState } from 'react';
import { Button } from '#/components/ui/button';
import { useOrganization } from '#/features/organizations/hooks/use-organization';
import { AddSecretModal } from './components/add-secret-modal';
import { EnvironmentModal } from './components/environment-modal';
import { ProjectModal } from './components/project-modal';
import {
  useEnvironments,
  useProjects,
  useSecretMutations,
  useSecretsList,
} from './hooks/use-secrets';

export function SecretsPage({ orgSlug }: { orgSlug: string }) {
  const { organization } = useOrganization(orgSlug);
  const orgId = organization?.id;

  const { data: projects = [], isLoading: loadingProjects } =
    useProjects(orgId);
  const [selectedProjectId, setSelectedProjectId] = useState<string>('');
  const activeProjectId = selectedProjectId || projects[0]?.id || '';

  const { data: environments = [] } = useEnvironments(orgId, activeProjectId);
  const [selectedEnvId, setSelectedEnvId] = useState<string>('');
  const activeEnvId = selectedEnvId || environments[0]?.id || '';

  const [reveal, setReveal] = useState(false);
  const { data: secrets = [] } = useSecretsList(
    orgId,
    activeProjectId,
    activeEnvId,
    reveal,
  );

  const mutations = useSecretMutations(orgId || '');
  const [showAddForm, setShowAddForm] = useState(false);
  const [showProjectModal, setShowProjectModal] = useState(false);
  const [showEnvModal, setShowEnvModal] = useState(false);

  const handleSaveSecret = async (input: {
    key: string;
    value: string;
    comment?: string;
  }) => {
    if (!activeProjectId || !activeEnvId) return;
    await mutations.saveSecret.mutateAsync({
      projectId: activeProjectId,
      environmentId: activeEnvId,
      input,
    });
    setShowAddForm(false);
  };

  const handleCreateProject = async (input: { slug: string; name: string }) => {
    const created = await mutations.addProject.mutateAsync(input);
    setSelectedProjectId(created.id);
    setShowProjectModal(false);
  };

  const handleCreateEnvironment = async (input: {
    slug: string;
    name: string;
  }) => {
    if (!activeProjectId) return;
    const created = await mutations.addEnvironment.mutateAsync({
      projectId: activeProjectId,
      input,
    });
    setSelectedEnvId(created.id);
    setShowEnvModal(false);
  };

  if (loadingProjects) {
    return <p className="p-8 text-sm text-white/55">Loading Secrets Vault…</p>;
  }

  return (
    <div className="space-y-8 p-6 lg:p-10 max-w-7xl mx-auto text-white">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-white/10 pb-6">
        <div>
          <div className="flex items-center gap-2.5">
            <div className="flex size-9 items-center justify-center rounded-xl bg-indigo-500/10 border border-indigo-500/20 text-indigo-400">
              <Lock className="size-5" />
            </div>
            <h1 className="text-2xl font-bold tracking-tight text-white">
              Secrets Vault
            </h1>
          </div>
          <p className="mt-1 text-sm text-white/60">
            Encrypted environment variables, zero-knowledge secret shares, and
            CLI process injection.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setShowProjectModal(true)}
            className="border-white/10 bg-white/5 text-xs text-white hover:bg-white/10"
          >
            <FolderPlus className="size-3.5 mr-1.5" />
            New Project
          </Button>
          {activeProjectId && (
            <Button
              size="sm"
              onClick={() => setShowAddForm(true)}
              className="bg-indigo-600 hover:bg-indigo-500 text-xs text-white font-medium"
            >
              <Plus className="size-3.5 mr-1.5" />
              Add Secret
            </Button>
          )}
        </div>
      </div>

      {projects.length === 0 ? (
        <div className="rounded-2xl border border-dashed border-white/15 bg-white/2.5 p-12 text-center">
          <KeyRound className="mx-auto size-10 text-white/40 mb-3" />
          <h3 className="text-base font-semibold text-white">
            No projects yet
          </h3>
          <p className="mt-1 text-xs text-white/60 max-w-sm mx-auto">
            Create your first secrets project to organize environment variables
            across development, staging, and production.
          </p>
          <Button
            size="sm"
            onClick={() => setShowProjectModal(true)}
            className="mt-4 bg-indigo-600 hover:bg-indigo-500 text-xs"
          >
            Create Project
          </Button>
        </div>
      ) : (
        <div className="space-y-6">
          <div className="flex flex-wrap items-center justify-between gap-4">
            <div className="flex items-center gap-2 overflow-x-auto pb-1">
              {projects.map((proj) => (
                <Button
                  key={proj.id}
                  size="sm"
                  variant={proj.id === activeProjectId ? 'default' : 'outline'}
                  onClick={() => {
                    setSelectedProjectId(proj.id);
                    setSelectedEnvId('');
                  }}
                  className={`text-xs h-8 ${
                    proj.id === activeProjectId
                      ? 'bg-indigo-600 hover:bg-indigo-500 text-white'
                      : 'border-white/10 bg-white/5 text-white/70 hover:bg-white/10'
                  }`}
                >
                  {proj.name}
                </Button>
              ))}
            </div>

            <div className="flex items-center gap-2">
              <span className="text-xs text-white/40">Environment:</span>
              {environments.map((env) => (
                <Button
                  key={env.id}
                  size="sm"
                  variant="outline"
                  onClick={() => setSelectedEnvId(env.id)}
                  className={`text-xs font-mono h-7 px-2.5 ${
                    env.id === activeEnvId
                      ? 'border-indigo-400/40 bg-indigo-500/20 text-indigo-300'
                      : 'border-white/10 bg-white/5 text-white/60 hover:bg-white/10'
                  }`}
                >
                  {env.slug}
                </Button>
              ))}
              <Button
                variant="outline"
                size="sm"
                onClick={() => setShowEnvModal(true)}
                className="size-7 p-0 border-dashed border-white/20 text-white/60 hover:text-white"
                title="Add Environment"
              >
                <Plus className="size-3" />
              </Button>
            </div>
          </div>

          <div className="flex items-center justify-between rounded-xl border border-white/10 bg-black/40 px-4 py-2 font-mono text-xs text-white/70">
            <div className="flex items-center gap-2 overflow-hidden truncate">
              <Terminal className="size-4 shrink-0 text-indigo-400" />
              <span className="truncate">
                outpipe secrets run -p{' '}
                {projects.find((p) => p.id === activeProjectId)?.slug || 'app'}{' '}
                -e{' '}
                {environments.find((e) => e.id === activeEnvId)?.slug ||
                  'production'}{' '}
                -- npm run dev
              </span>
            </div>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => {
                const projSlug =
                  projects.find((p) => p.id === activeProjectId)?.slug || 'app';
                const envSlug =
                  environments.find((e) => e.id === activeEnvId)?.slug ||
                  'production';
                navigator.clipboard.writeText(
                  `outpipe secrets run -p ${projSlug} -e ${envSlug} -- npm run dev`,
                );
              }}
              className="size-7 p-0 ml-3 shrink-0 text-white/40 hover:text-white hover:bg-transparent"
              title="Copy CLI command"
            >
              <Copy className="size-3.5" />
            </Button>
          </div>

          <div className="flex items-center justify-between">
            <h3 className="text-sm font-semibold text-white/80">
              Variables ({secrets.length})
            </h3>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setReveal(!reveal)}
              className="border-white/10 bg-white/5 text-xs text-white/80 hover:text-white"
            >
              {reveal ? (
                <>
                  <EyeOff className="size-3.5 mr-1.5" />
                  Hide Values
                </>
              ) : (
                <>
                  <Eye className="size-3.5 mr-1.5" />
                  Reveal Values
                </>
              )}
            </Button>
          </div>

          <div className="overflow-hidden rounded-xl border border-white/10 bg-white/2.5">
            {secrets.length === 0 ? (
              <div className="p-8 text-center text-xs text-white/50">
                No secrets in this environment yet. Click &quot;Add Secret&quot;
                to define one.
              </div>
            ) : (
              <table className="w-full text-left text-xs">
                <thead>
                  <tr className="border-b border-white/10 bg-white/5 text-white/60">
                    <th className="py-2.5 px-4 font-medium">Key</th>
                    <th className="py-2.5 px-4 font-medium">Value</th>
                    <th className="py-2.5 px-4 font-medium">Comment</th>
                    <th className="py-2.5 px-4 text-right font-medium">
                      Actions
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-white/5 font-mono">
                  {secrets.map((item) => (
                    <tr
                      key={item.id}
                      className="hover:bg-white/2.5 transition-colors"
                    >
                      <td className="py-3 px-4 font-semibold text-indigo-300">
                        {item.key}
                      </td>
                      <td className="py-3 px-4 text-white/90">
                        {reveal && item.value ? (
                          item.value
                        ) : (
                          <span className="text-white/30 tracking-widest">
                            ••••••••••••••••
                          </span>
                        )}
                      </td>
                      <td className="py-3 px-4 font-sans text-white/50">
                        {item.comment || '—'}
                      </td>
                      <td className="py-3 px-4 text-right font-sans">
                        <button
                          type="button"
                          onClick={() => mutations.removeSecret.mutate(item.id)}
                          className="text-white/40 hover:text-rose-400 transition-colors p-1"
                          title="Delete secret"
                        >
                          <Trash2 className="size-3.5" />
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </div>
      )}

      <AddSecretModal
        isOpen={showAddForm}
        onClose={() => setShowAddForm(false)}
        onSave={handleSaveSecret}
        isSaving={mutations.saveSecret.isPending}
      />

      <ProjectModal
        isOpen={showProjectModal}
        onClose={() => setShowProjectModal(false)}
        onCreate={handleCreateProject}
        isCreating={mutations.addProject.isPending}
      />

      <EnvironmentModal
        isOpen={showEnvModal}
        onClose={() => setShowEnvModal(false)}
        onCreate={handleCreateEnvironment}
        isCreating={mutations.addEnvironment.isPending}
      />
    </div>
  );
}
