import { FolderPlus, Lock, Plus } from 'lucide-react';
import { Button } from '#/components/ui/button';

export interface SecretsHeaderProps {
  canAddSecret: boolean;
  onNewProject: () => void;
  onAddSecret: () => void;
}

export function SecretsHeader({
  canAddSecret,
  onNewProject,
  onAddSecret,
}: SecretsHeaderProps) {
  return (
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
          Encrypted environment variables, zero-knowledge secret shares, and CLI
          process injection.
        </p>
      </div>

      <div className="flex items-center gap-3">
        <Button
          variant="outline"
          size="sm"
          onClick={onNewProject}
          className="border-white/10 bg-white/5 text-xs text-white hover:bg-white/10"
        >
          <FolderPlus className="size-3.5 mr-1.5" />
          New Project
        </Button>
        {canAddSecret && (
          <Button
            size="sm"
            onClick={onAddSecret}
            className="bg-primary hover:bg-primary/90 text-xs text-primary-foreground font-medium"
          >
            <Plus className="size-3.5 mr-1.5" />
            Add Secret
          </Button>
        )}
      </div>
    </div>
  );
}
