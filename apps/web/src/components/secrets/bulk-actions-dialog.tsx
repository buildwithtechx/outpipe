import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '#/components/ui/dialog';
import type { SecretEnvironment, SecretItem } from '#/interfaces/secret';
import { BulkActionContent } from './bulk-action-content';

export interface BulkActionsDialogProps {
  action: 'move' | 'delete' | 'share' | null;
  onClose: () => void;
  selectedSecrets: SecretItem[];
  environments: SecretEnvironment[];
  currentEnvironmentId: string;
  onConfirmMove: (
    targetEnvId: string,
    conflictMode: 'skip' | 'overwrite',
  ) => Promise<void>;
  onConfirmDelete: () => Promise<void>;
  onConfirmShare: (expiry: string, maxViews: number) => Promise<string | null>;
  isSubmitting?: boolean;
}

export function BulkActionsDialog({
  action,
  onClose,
  selectedSecrets,
  environments,
  currentEnvironmentId,
  onConfirmMove,
  onConfirmDelete,
  onConfirmShare,
  isSubmitting = false,
}: BulkActionsDialogProps) {
  if (!action) return null;

  const titles = {
    move: 'Move Secrets to Environment',
    delete: 'Delete Selected Secrets',
    share: 'Create Encrypted Secrets Share',
  };

  return (
    <Dialog
      open={Boolean(action)}
      onOpenChange={(open) => {
        if (!open && !isSubmitting) onClose();
      }}
    >
      <DialogContent className="max-w-md border-white/10 bg-card text-white">
        <DialogHeader>
          <DialogTitle className="text-sm font-semibold text-white">
            {titles[action]}
          </DialogTitle>
        </DialogHeader>

        <BulkActionContent
          action={action}
          selectedSecrets={selectedSecrets}
          environments={environments}
          currentEnvironmentId={currentEnvironmentId}
          onConfirmMove={onConfirmMove}
          onConfirmDelete={onConfirmDelete}
          onConfirmShare={onConfirmShare}
          isSubmitting={isSubmitting}
          onClose={onClose}
        />
      </DialogContent>
    </Dialog>
  );
}
