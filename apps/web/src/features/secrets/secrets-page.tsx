import { Eye, EyeOff, FolderPlus, KeyRound, Lock, Plus } from 'lucide-react';
import { useState } from 'react';
import { QueryFeedback } from '#/components/feedback/query-feedback';
import { Button } from '#/components/ui/button';
import { CopyCommand } from '#/components/ui/copy-command';
import { useOrganization } from '#/features/organizations/hooks/use-organization';
import { AddSecretModal } from './components/add-secret-modal';
import { EnvironmentModal } from './components/environment-modal';
import { ProjectModal } from './components/project-modal';
import { SecretVariablesTable } from './components/secret-variables-table';
import {
  useEnvironments,
  useProjects,
  useSecretMutations,
  useSecretsList,
} from './hooks/use-secrets';

export function SecretsPage({ orgSlug }: { orgSlug: string }) {
  const organizationQuery = useOrganization(orgSlug);
  const { organization } = organizationQuery;
  const orgId = organization?.id;

  const projectQuery = useProjects(orgId);
  const { data: projects = [], isLoading: loadingProjects } = projectQuery;
  const [selectedProjectId, setSelectedProjectId] = useState<string>('');
  const activeProjectId = projects.some(
    (project) => project.id === selectedProjectId,
  )
    ? selectedProjectId
    : projects[0]?.id || '';

  const environmentQuery = useEnvironments(orgId, activeProjectId);
  const { data: environments = [] } = environmentQuery;
  const [selectedEnvId, setSelectedEnvId] = useState<string>('');
  const activeEnvId = environments.some(
    (environment) => environment.id === selectedEnvId,
  )
    ? selectedEnvId
    : environments[0]?.id || '';

  const [revealContext, setRevealContext] = useState<string | null>(null);
  const context = `${orgId}:${activeProjectId}:${activeEnvId}`;
  const reveal = revealContext === context;
  const secretQuery = useSecretsList(
    orgId,
    activeProjectId,
    activeEnvId,
    reveal,
  );
  const { data: secrets = [] } = secretQuery;

  const mutations = useSecretMutations(orgId || '');
  const [showAddForm, setShowAddForm] = useState(false);
  const [showProjectModal, setShowProjectModal] = useState(false);
  const [showEnvModal, setShowEnvModal] = useState(false);

  const handleSaveSecret = async (input: {
    key: string;
    value: string;
    comment?: string;
  }) => {
    if (!activeProjectId || !activeEnvId)
      throw new Error('Select an environment first');
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
    setSelectedEnvId('');
    setRevealContext(null);
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
    setRevealContext(null);
    setShowEnvModal(false);
  };

  if (organizationQuery.isLoading || loadingProjects) {
    return <p className="p-8 text-sm text-white/55">Loading Secrets Vault…</p>;
  }

  if (!orgId && !organizationQuery.isError)
    return (
      <p role="alert">Workspace not found. Choose an available workspace.</p>
    );

  if (organizationQuery.isError || projectQuery.isError)
    return (
      <div role="alert" className="space-y-3">
        <p>
          {organizationQuery.isError
            ? 'Could not load the workspace.'
            : 'Could not load secrets projects.'}
        </p>
        <Button
          variant="outline"
          onClick={() => {
            if (organizationQuery.isError) void organizationQuery.refetch();
            else if (orgId) void projectQuery.refetch();
          }}
        >
          Try again
        </Button>
      </div>
    );

  return (
    <div className="space-y-8 text-foreground">
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
          {activeProjectId && activeEnvId && (
            <Button
              size="sm"
              onClick={() => setShowAddForm(true)}
              className="bg-primary hover:bg-primary/90 text-xs text-primary-foreground font-medium"
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
            className="mt-4 bg-primary hover:bg-primary/90 text-xs text-primary-foreground"
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
                    setRevealContext(null);
                  }}
                  className={`text-xs h-8 ${
                    proj.id === activeProjectId
                      ? 'bg-primary hover:bg-primary/90 text-primary-foreground'
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
                  onClick={() => {
                    setSelectedEnvId(env.id);
                    setRevealContext(null);
                  }}
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
                aria-label="Add environment"
                onClick={() => setShowEnvModal(true)}
                className="size-7 p-0 border-dashed border-white/20 text-white/60 hover:text-white"
                title="Add Environment"
              >
                <Plus className="size-3" />
              </Button>
            </div>
          </div>

          <QueryFeedback query={environmentQuery} label="environments">
            {!environments.length && (
              <p className="rounded-xl border border-dashed border-border p-4 text-sm text-muted-foreground">
                Create an environment to add variables.
              </p>
            )}
          </QueryFeedback>
          {activeProjectId && activeEnvId && (
            <div className="space-y-2">
              <CopyCommand
                text={`outpipe secrets run --project ${activeProjectId} --environment ${activeEnvId} -- npm run dev`}
                label="Copy secrets CLI command"
              />
              <p className="text-xs text-muted-foreground">
                Set OUTPIPE_TOKEN to a machine token scoped to this environment
                with secrets:read permission. Secret values stay out of the
                command.
              </p>
            </div>
          )}

          <div className="flex items-center justify-between">
            <h3 className="text-sm font-semibold text-white/80">
              Variables ({secrets.length})
            </h3>
            <Button
              variant="outline"
              size="sm"
              aria-pressed={reveal}
              disabled={!activeEnvId || (!reveal && secretQuery.isFetching)}
              onClick={() => setRevealContext(reveal ? null : context)}
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

          <QueryFeedback query={secretQuery} label="variables">
            <SecretVariablesTable
              secrets={secrets}
              reveal={reveal}
              environmentSelected={Boolean(activeEnvId)}
              pendingDeleteId={
                mutations.removeSecret.isPending
                  ? mutations.removeSecret.variables
                  : undefined
              }
              onDelete={(id) => mutations.removeSecret.mutateAsync(id)}
            />
          </QueryFeedback>
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
