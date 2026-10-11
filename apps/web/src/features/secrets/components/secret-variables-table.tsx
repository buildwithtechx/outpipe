import { History, Share2, Trash2 } from 'lucide-react';
import { HoldToDelete } from '#/components/secrets/hold-to-delete';
import { Button } from '#/components/ui/button';
import type { SecretItem } from '#/interfaces/secret';

export interface SecretVariablesTableProps {
  secrets: SecretItem[];
  reveal: boolean;
  environmentSelected: boolean;
  selectedIds?: string[];
  pendingDeleteId?: string;
  onToggleSelect?: (id: string, checked: boolean) => void;
  onToggleSelectAll?: (checked: boolean) => void;
  onOpenHistory?: (secret: SecretItem) => void;
  onOpenBulkAction?: (action: 'move' | 'delete' | 'share') => void;
  onDelete: (id: string) => Promise<unknown>;
}

export function SecretVariablesTable({
  secrets,
  reveal,
  environmentSelected,
  selectedIds = [],
  pendingDeleteId,
  onToggleSelect,
  onToggleSelectAll,
  onOpenHistory,
  onOpenBulkAction,
  onDelete,
}: SecretVariablesTableProps) {
  const isAllSelected =
    secrets.length > 0 && selectedIds.length === secrets.length;
  const isIndeterminate =
    selectedIds.length > 0 && selectedIds.length < secrets.length;

  return (
    <div className="space-y-3">
      {selectedIds.length > 0 && (
        <div className="flex flex-wrap items-center justify-between gap-2 rounded-lg border border-indigo-500/20 bg-indigo-500/10 px-3.5 py-2">
          <span className="text-xs font-medium text-indigo-200">
            {selectedIds.length} of {secrets.length} variable
            {selectedIds.length === 1 ? '' : 's'} selected
          </span>
          <div className="flex items-center gap-1.5">
            <Button
              type="button"
              variant="outline"
              size="xs"
              onClick={() => onOpenBulkAction?.('move')}
              className="border-indigo-400/30 text-indigo-200 hover:bg-indigo-500/20"
            >
              Move
            </Button>
            <Button
              type="button"
              variant="outline"
              size="xs"
              onClick={() => onOpenBulkAction?.('share')}
              className="gap-1 border-emerald-500/30 text-emerald-300 hover:bg-emerald-500/20"
            >
              <Share2 className="size-3" />
              <span>Share</span>
            </Button>
            <Button
              type="button"
              variant="outline"
              size="xs"
              onClick={() => onOpenBulkAction?.('delete')}
              className="gap-1 border-rose-500/30 text-rose-300 hover:bg-rose-500/20"
            >
              <Trash2 className="size-3" />
              <span>Delete</span>
            </Button>
          </div>
        </div>
      )}

      <div className="overflow-x-auto rounded-xl border border-white/10 bg-white/2.5">
        {!environmentSelected ? (
          <p className="p-6 text-sm text-muted-foreground">
            Select or create an environment to manage variables.
          </p>
        ) : secrets.length === 0 ? (
          <div className="p-8 text-center text-xs text-white/50">
            No secrets in this environment yet. Click &quot;Add Secret&quot; to
            define one.
          </div>
        ) : (
          <table
            aria-label="Secret variables"
            className="w-full text-left text-xs"
          >
            <thead>
              <tr className="border-b border-white/10 bg-white/5 text-white/60">
                <th className="w-10 py-2.5 px-3">
                  <input
                    type="checkbox"
                    aria-label="Select all secrets"
                    checked={isAllSelected}
                    ref={(el) => {
                      if (el) el.indeterminate = isIndeterminate;
                    }}
                    onChange={(e) => onToggleSelectAll?.(e.target.checked)}
                    className="size-3.5 rounded border-white/20 bg-white/5 accent-indigo-500 cursor-pointer"
                  />
                </th>
                <th className="py-2.5 px-4 font-medium">Key</th>
                <th className="py-2.5 px-4 font-medium">Value</th>
                <th className="py-2.5 px-4 font-medium">Comment</th>
                <th className="py-2.5 px-4 text-right font-medium">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5 font-mono">
              {secrets.map((item) => {
                const isSelected = selectedIds.includes(item.id);
                return (
                  <tr
                    key={item.id}
                    className={`transition-colors ${
                      isSelected ? 'bg-indigo-500/10' : 'hover:bg-white/2.5'
                    }`}
                  >
                    <td className="w-10 py-3 px-3">
                      <input
                        type="checkbox"
                        aria-label={`Select ${item.key}`}
                        checked={isSelected}
                        onChange={(e) =>
                          onToggleSelect?.(item.id, e.target.checked)
                        }
                        className="size-3.5 rounded border-white/20 bg-white/5 accent-indigo-500 cursor-pointer"
                      />
                    </td>
                    <td className="py-3 px-4 font-semibold text-indigo-300">
                      {item.key}
                    </td>
                    <td className="max-w-xs break-all py-3 px-4 text-white/90">
                      {reveal && item.value !== undefined ? (
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
                      <div className="flex items-center justify-end gap-1.5">
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon-xs"
                          onClick={() => onOpenHistory?.(item)}
                          className="text-white/50 hover:text-white"
                          title="Version History"
                        >
                          <History className="size-3.5" />
                        </Button>
                        <HoldToDelete
                          onConfirm={() => onDelete(item.id)}
                          loading={pendingDeleteId === item.id}
                          label="Delete"
                          className="h-6 px-2 text-[11px]"
                        />
                      </div>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
