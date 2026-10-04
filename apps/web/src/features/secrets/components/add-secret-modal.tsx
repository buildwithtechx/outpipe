import { useState } from 'react';
import { Button } from '#/components/ui/button';

interface AddSecretModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: (input: {
    key: string;
    value: string;
    comment?: string;
  }) => Promise<void>;
  isSaving: boolean;
}

export function AddSecretModal({
  isOpen,
  onClose,
  onSave,
  isSaving,
}: AddSecretModalProps) {
  const [newKey, setNewKey] = useState('');
  const [newValue, setNewValue] = useState('');
  const [newComment, setNewComment] = useState('');

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newKey.trim() || !newValue.trim()) return;
    await onSave({
      key: newKey.trim(),
      value: newValue.trim(),
      comment: newComment.trim() || undefined,
    });
    setNewKey('');
    setNewValue('');
    setNewComment('');
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-xs p-4">
      <div className="w-full max-w-md rounded-2xl border border-white/10 bg-zinc-900 p-6 shadow-2xl text-white">
        <h3 className="text-base font-semibold text-white mb-4">
          Add Environment Secret
        </h3>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label
              htmlFor="secret-key-input"
              className="block text-xs font-medium text-white/70 mb-1"
            >
              Key Name
            </label>
            <input
              id="secret-key-input"
              type="text"
              required
              placeholder="e.g. STRIPE_API_KEY"
              value={newKey}
              onChange={(e) => setNewKey(e.target.value)}
              className="w-full rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-xs font-mono text-white placeholder-white/30 focus:border-indigo-500 focus:outline-hidden"
            />
          </div>
          <div>
            <label
              htmlFor="secret-value-input"
              className="block text-xs font-medium text-white/70 mb-1"
            >
              Secret Value
            </label>
            <textarea
              id="secret-value-input"
              required
              rows={3}
              placeholder="Enter secret value..."
              value={newValue}
              onChange={(e) => setNewValue(e.target.value)}
              className="w-full rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-xs font-mono text-white placeholder-white/30 focus:border-indigo-500 focus:outline-hidden"
            />
          </div>
          <div>
            <label
              htmlFor="secret-comment-input"
              className="block text-xs font-medium text-white/70 mb-1"
            >
              Comment (optional)
            </label>
            <input
              id="secret-comment-input"
              type="text"
              placeholder="e.g. Live billing secret"
              value={newComment}
              onChange={(e) => setNewComment(e.target.value)}
              className="w-full rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-xs text-white placeholder-white/30 focus:border-indigo-500 focus:outline-hidden"
            />
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={onClose}
              className="text-xs border-white/10 bg-white/5 text-white/70"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              size="sm"
              disabled={isSaving}
              className="text-xs bg-indigo-600 hover:bg-indigo-500"
            >
              Save Secret
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
