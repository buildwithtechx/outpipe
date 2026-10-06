import { ConfirmAction } from '#/components/feedback/confirm-action';
import type { SecretItem } from '#/interfaces/secret';

export function SecretVariablesTable({
  secrets,
  reveal,
  environmentSelected,
  deleting,
  onDelete,
}: {
  secrets: SecretItem[];
  reveal: boolean;
  environmentSelected: boolean;
  deleting: boolean;
  onDelete: (id: string) => Promise<unknown>;
}) {
  return (
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
        <table className="w-full text-left text-xs">
          <thead>
            <tr className="border-b border-white/10 bg-white/5 text-white/60">
              <th className="py-2.5 px-4 font-medium">Key</th>
              <th className="py-2.5 px-4 font-medium">Value</th>
              <th className="py-2.5 px-4 font-medium">Comment</th>
              <th className="py-2.5 px-4 text-right font-medium">Actions</th>
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
                  <ConfirmAction
                    title={`Delete ${item.key}?`}
                    description="Remove this variable from the selected environment."
                    label="Delete"
                    pending={deleting}
                    onConfirm={() => onDelete(item.id)}
                  />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
