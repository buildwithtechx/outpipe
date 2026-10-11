import { LockKeyhole } from 'lucide-react';
import { SideSheet } from '#/components/ui/side-sheet';
import type { SecretItem, SecretVersion } from '#/interfaces/secret';
import { SecretHistoryContent } from './secret-history-content';

export interface SecretHistorySheetProps {
  open: boolean;
  onClose: () => void;
  secret: SecretItem | null;
  versions?: SecretVersion[];
  onRollback: (version: number) => Promise<void>;
  isRollingBack?: boolean;
}

export function SecretHistorySheet({
  open,
  onClose,
  secret,
  versions = [],
  onRollback,
  isRollingBack = false,
}: SecretHistorySheetProps) {
  return (
    <SideSheet
      open={open && Boolean(secret)}
      onClose={onClose}
      title="Secret Version History"
      description={
        secret ? `Revisions and history for ${secret.key}` : undefined
      }
      closeDisabled={isRollingBack}
      footer={
        <p className="flex items-center gap-1.5 text-[11px] text-white/50">
          <LockKeyhole className="size-3 text-white/40 shrink-0" />
          <span>
            Values are masked by default. Restoring an earlier version records
            an audit entry.
          </span>
        </p>
      }
    >
      {secret && (
        <SecretHistoryContent
          secret={secret}
          versions={versions}
          onRollback={onRollback}
          isRollingBack={isRollingBack}
        />
      )}
    </SideSheet>
  );
}
