import { AlertTriangle, ArrowRight, Check, Copy, Share2 } from 'lucide-react';
import { useState } from 'react';
import { Button } from '#/components/ui/button';
import { Label } from '#/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '#/components/ui/select';
import type { SecretEnvironment, SecretItem } from '#/interfaces/secret';
import { HoldToDelete } from './hold-to-delete';

export interface BulkActionContentProps {
  action: 'move' | 'delete' | 'share';
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
  onClose: () => void;
}

export function BulkActionContent({
  action,
  selectedSecrets,
  environments,
  currentEnvironmentId,
  onConfirmMove,
  onConfirmDelete,
  onConfirmShare,
  isSubmitting = false,
  onClose,
}: BulkActionContentProps) {
  const otherEnvs = environments.filter((e) => e.id !== currentEnvironmentId);
  const [targetEnvId, setTargetEnvId] = useState(otherEnvs[0]?.id ?? '');
  const [conflictMode, setConflictMode] = useState<'skip' | 'overwrite'>(
    'skip',
  );

  const [expiry, setExpiry] = useState('7d');
  const [maxViews, setMaxViews] = useState('5');
  const [generatedShareUrl, setGeneratedShareUrl] = useState<string | null>(
    null,
  );
  const [copied, setCopied] = useState(false);

  const handleShare = async () => {
    const url = await onConfirmShare(expiry, Number(maxViews) || 5);
    if (url) {
      setGeneratedShareUrl(url);
    }
  };

  const copyUrl = async () => {
    if (!generatedShareUrl) return;
    await navigator.clipboard.writeText(generatedShareUrl);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  if (action === 'move') {
    return (
      <div className="space-y-4 pt-1">
        <p className="text-xs text-white/70">
          Move{' '}
          <strong className="text-indigo-300">{selectedSecrets.length}</strong>{' '}
          selected variable{selectedSecrets.length === 1 ? '' : 's'} to another
          environment.
        </p>

        <div className="max-h-28 overflow-y-auto rounded-lg border border-white/10 bg-white/3 p-2 font-mono text-[11px] text-white/80 space-y-1">
          {selectedSecrets.map((s) => (
            <div key={s.id} className="truncate">
              • {s.key}
            </div>
          ))}
        </div>

        <div className="space-y-1.5">
          <Label className="text-xs text-white/70">Target Environment</Label>
          <Select value={targetEnvId} onValueChange={setTargetEnvId}>
            <SelectTrigger className="w-full border-white/10 bg-white/5 text-xs text-white">
              <SelectValue placeholder="Select target environment" />
            </SelectTrigger>
            <SelectContent className="border-white/10 bg-card text-white">
              {otherEnvs.map((env) => (
                <SelectItem key={env.id} value={env.id}>
                  {env.name} ({env.slug})
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="space-y-1.5">
          <Label className="text-xs text-white/70">Key Conflicts Policy</Label>
          <Select
            value={conflictMode}
            onValueChange={(val) =>
              setConflictMode(val as 'skip' | 'overwrite')
            }
          >
            <SelectTrigger className="w-full border-white/10 bg-white/5 text-xs text-white">
              <SelectValue />
            </SelectTrigger>
            <SelectContent className="border-white/10 bg-card text-white">
              <SelectItem value="skip">
                Skip if key already exists in target
              </SelectItem>
              <SelectItem value="overwrite">
                Overwrite with values from current environment
              </SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div className="flex justify-end gap-2 pt-2">
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={onClose}
            disabled={isSubmitting}
            className="text-white/60"
          >
            Cancel
          </Button>
          <Button
            type="button"
            size="sm"
            disabled={isSubmitting || !targetEnvId}
            onClick={() => onConfirmMove(targetEnvId, conflictMode)}
            className="gap-1.5 bg-indigo-600 hover:bg-indigo-500 text-white"
          >
            <ArrowRight className="size-3.5" />
            <span>{isSubmitting ? 'Moving...' : 'Move Variables'}</span>
          </Button>
        </div>
      </div>
    );
  }

  if (action === 'delete') {
    return (
      <div className="space-y-4 pt-1">
        <div className="flex items-start gap-2.5 rounded-lg border border-rose-500/20 bg-rose-500/10 p-3 text-xs text-rose-300">
          <AlertTriangle className="size-4 shrink-0 mt-0.5" />
          <div>
            <p className="font-semibold">Destructive Action</p>
            <p className="mt-0.5 text-rose-200/80">
              You are about to delete{' '}
              <strong className="text-white">{selectedSecrets.length}</strong>{' '}
              secret variable{selectedSecrets.length === 1 ? '' : 's'}. This
              action cannot be immediately undone.
            </p>
          </div>
        </div>

        <div className="max-h-32 overflow-y-auto rounded-lg border border-white/10 bg-white/3 p-2 font-mono text-[11px] text-white/80 space-y-1">
          {selectedSecrets.map((s) => (
            <div key={s.id} className="truncate text-rose-300/90">
              • {s.key}
            </div>
          ))}
        </div>

        <div className="flex items-center justify-between pt-2">
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={onClose}
            disabled={isSubmitting}
            className="text-white/60"
          >
            Cancel
          </Button>
          <HoldToDelete
            onConfirm={onConfirmDelete}
            loading={isSubmitting}
            label={`Hold to delete ${selectedSecrets.length} secrets`}
          />
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-4 pt-1">
      {generatedShareUrl ? (
        <div className="space-y-3">
          <div className="flex items-center gap-2 text-xs font-semibold text-emerald-300">
            <Check className="size-4 text-emerald-400" />
            <span>Encrypted Secret Link Generated</span>
          </div>

          <div className="flex items-center gap-2 rounded-lg border border-white/10 bg-black/40 p-2">
            <input
              type="text"
              readOnly
              value={generatedShareUrl}
              className="flex-1 bg-transparent font-mono text-xs text-white/90 outline-none"
            />
            <Button
              type="button"
              size="xs"
              variant="outline"
              onClick={copyUrl}
              className="border-white/10 text-white"
            >
              {copied ? (
                <Check className="size-3 text-emerald-400" />
              ) : (
                <Copy className="size-3" />
              )}
              <span className="ml-1">{copied ? 'Copied' : 'Copy'}</span>
            </Button>
          </div>

          <p className="text-[11px] text-white/50">
            The secret bundle is encrypted end-to-end in your browser. The
            decryption key is in the link hash and is never stored on the
            server.
          </p>

          <div className="flex justify-end pt-2">
            <Button
              type="button"
              size="sm"
              onClick={onClose}
              className="bg-primary text-primary-foreground"
            >
              Done
            </Button>
          </div>
        </div>
      ) : (
        <div className="space-y-4">
          <p className="text-xs text-white/70">
            Share{' '}
            <strong className="text-indigo-300">
              {selectedSecrets.length}
            </strong>{' '}
            secret variable{selectedSecrets.length === 1 ? '' : 's'} via a
            single-use end-to-end encrypted link.
          </p>

          <div className="max-h-24 overflow-y-auto rounded-lg border border-white/10 bg-white/3 p-2 font-mono text-[11px] text-white/80 space-y-1">
            {selectedSecrets.map((s) => (
              <div key={s.id} className="truncate">
                • {s.key}
              </div>
            ))}
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <Label className="text-xs text-white/70">Expiration</Label>
              <Select value={expiry} onValueChange={setExpiry}>
                <SelectTrigger className="w-full border-white/10 bg-white/5 text-xs text-white">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent className="border-white/10 bg-card text-white">
                  <SelectItem value="1d">1 Day</SelectItem>
                  <SelectItem value="7d">7 Days</SelectItem>
                  <SelectItem value="30d">30 Days</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-1.5">
              <Label className="text-xs text-white/70">Max Views</Label>
              <Select value={maxViews} onValueChange={setMaxViews}>
                <SelectTrigger className="w-full border-white/10 bg-white/5 text-xs text-white">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent className="border-white/10 bg-card text-white">
                  <SelectItem value="1">1 View (burn on read)</SelectItem>
                  <SelectItem value="5">5 Views</SelectItem>
                  <SelectItem value="10">10 Views</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="flex justify-end gap-2 pt-2">
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={onClose}
              disabled={isSubmitting}
              className="text-white/60"
            >
              Cancel
            </Button>
            <Button
              type="button"
              size="sm"
              disabled={isSubmitting}
              onClick={handleShare}
              className="gap-1.5 bg-emerald-600 hover:bg-emerald-500 text-white"
            >
              <Share2 className="size-3.5" />
              <span>
                {isSubmitting ? 'Encrypting...' : 'Create Secure Share'}
              </span>
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
