import { Eye, EyeOff, History, RotateCcw } from 'lucide-react';
import { useState } from 'react';
import { Badge } from '#/components/ui/badge';
import { Button } from '#/components/ui/button';
import type { SecretItem, SecretVersion } from '#/interfaces/secret';

export interface SecretHistoryContentProps {
  secret: SecretItem;
  versions: SecretVersion[];
  onRollback: (version: number) => Promise<void>;
  isRollingBack?: boolean;
}

export function SecretHistoryContent({
  secret,
  versions,
  onRollback,
  isRollingBack = false,
}: SecretHistoryContentProps) {
  const [revealedVersions, setRevealedVersions] = useState<
    Record<number, boolean>
  >({});
  const [confirmRollbackVersion, setConfirmRollbackVersion] = useState<
    number | null
  >(null);

  const toggleReveal = (v: number) => {
    setRevealedVersions((prev) => ({
      ...prev,
      [v]: !prev[v],
    }));
  };

  const allVersions: SecretVersion[] =
    versions.length > 0
      ? versions
      : [
          {
            version: secret.version,
            value: secret.value,
            comment: secret.comment,
            updatedAt: secret.updatedAt,
          },
        ];

  // Sort descending by version number
  const sorted = [...allVersions].sort((a, b) => b.version - a.version);

  return (
    <div className="space-y-4">
      <div className="rounded-lg border border-white/10 bg-white/3 p-3">
        <div className="flex items-center justify-between">
          <span className="font-mono text-xs font-semibold text-indigo-300">
            {secret.key}
          </span>
          <Badge
            variant="outline"
            className="border-indigo-500/30 text-indigo-300 text-[10px]"
          >
            Current: v{secret.version}
          </Badge>
        </div>
        {secret.comment && (
          <p className="mt-1 text-xs text-white/50">{secret.comment}</p>
        )}
      </div>

      <div className="space-y-3">
        <h4 className="flex items-center gap-1.5 text-xs font-medium text-white/70">
          <History className="size-3.5 text-white/50" />
          <span>Revision History ({sorted.length})</span>
        </h4>

        <div className="divide-y divide-white/5 rounded-lg border border-white/10 bg-white/2">
          {sorted.map((item) => {
            const isCurrent = item.version === secret.version;
            const isRevealed = Boolean(revealedVersions[item.version]);
            const isConfirming = confirmRollbackVersion === item.version;

            return (
              <div
                key={item.version}
                className="flex flex-col gap-2 p-3.5 transition-colors hover:bg-white/2"
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <span className="font-mono text-xs font-bold text-white/90">
                      v{item.version}
                    </span>
                    {isCurrent ? (
                      <Badge className="bg-emerald-500/20 text-emerald-300 border-emerald-500/30 text-[10px] py-0">
                        Active
                      </Badge>
                    ) : (
                      <span className="text-[11px] text-white/40">
                        Archived
                      </span>
                    )}
                  </div>
                  <time className="text-[11px] text-white/40">
                    {new Date(item.updatedAt).toLocaleString()}
                  </time>
                </div>

                <div className="flex items-center justify-between gap-2 rounded border border-white/5 bg-black/30 px-2.5 py-1.5">
                  <span className="font-mono text-xs break-all text-white/80">
                    {isRevealed && item.value !== undefined ? (
                      item.value
                    ) : (
                      <span className="text-white/30 tracking-wider">
                        ••••••••••••••••
                      </span>
                    )}
                  </span>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon-xs"
                    onClick={() => toggleReveal(item.version)}
                    className="text-white/50 hover:text-white shrink-0"
                    title={isRevealed ? 'Hide value' : 'Reveal value'}
                  >
                    {isRevealed ? (
                      <EyeOff className="size-3" />
                    ) : (
                      <Eye className="size-3" />
                    )}
                  </Button>
                </div>

                {!isCurrent && (
                  <div className="flex items-center justify-end pt-1">
                    {isConfirming ? (
                      <div className="flex items-center gap-2">
                        <span className="text-[11px] text-amber-300">
                          Restore v{item.version}?
                        </span>
                        <Button
                          type="button"
                          variant="ghost"
                          size="xs"
                          onClick={() => setConfirmRollbackVersion(null)}
                          className="text-white/50"
                        >
                          Cancel
                        </Button>
                        <Button
                          type="button"
                          size="xs"
                          disabled={isRollingBack}
                          onClick={async () => {
                            await onRollback(item.version);
                            setConfirmRollbackVersion(null);
                          }}
                          className="bg-amber-500/20 text-amber-300 border border-amber-500/30 hover:bg-amber-500/30"
                        >
                          {isRollingBack ? 'Restoring...' : 'Confirm'}
                        </Button>
                      </div>
                    ) : (
                      <Button
                        type="button"
                        variant="outline"
                        size="xs"
                        onClick={() => setConfirmRollbackVersion(item.version)}
                        className="gap-1 border-white/10 text-white/70 hover:text-white text-[11px]"
                      >
                        <RotateCcw className="size-3" />
                        <span>Restore v{item.version}</span>
                      </Button>
                    )}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
