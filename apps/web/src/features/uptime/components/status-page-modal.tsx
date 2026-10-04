import { useState } from 'react';
import { Button } from '#/components/ui/button';

interface StatusPageModalProps {
  isOpen: boolean;
  onClose: () => void;
  initialSlug?: string;
  initialTitle?: string;
  initialDescription?: string;
  initialIsPublic?: boolean;
  onSave: (input: {
    slug: string;
    title: string;
    description: string;
    is_public: boolean;
  }) => Promise<void>;
  isSaving: boolean;
}

export function StatusPageModal({
  isOpen,
  onClose,
  initialSlug = 'default',
  initialTitle = 'System Status',
  initialDescription = '',
  initialIsPublic = true,
  onSave,
  isSaving,
}: StatusPageModalProps) {
  const [slug, setSlug] = useState(initialSlug);
  const [title, setTitle] = useState(initialTitle);
  const [description, setDescription] = useState(initialDescription);
  const [isPublic, setIsPublic] = useState(initialIsPublic);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!slug.trim() || !title.trim()) return;
    await onSave({
      slug: slug.trim().toLowerCase(),
      title: title.trim(),
      description: description.trim(),
      is_public: isPublic,
    });
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-xs p-4">
      <div className="w-full max-w-md rounded-2xl border border-white/10 bg-zinc-900 p-6 shadow-2xl text-white">
        <h3 className="text-base font-semibold text-white mb-4">
          Status Page Settings
        </h3>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label
              htmlFor="status-slug-input"
              className="block text-xs font-medium text-white/70 mb-1"
            >
              Custom Slug URL
            </label>
            <div className="flex items-center rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-xs font-mono text-white/50">
              <span>status.outpipe.dev/</span>
              <input
                id="status-slug-input"
                type="text"
                required
                value={slug}
                onChange={(e) => setSlug(e.target.value)}
                className="w-full bg-transparent text-white font-mono placeholder-white/30 focus:outline-hidden ml-1"
              />
            </div>
          </div>

          <div>
            <label
              htmlFor="status-title-input"
              className="block text-xs font-medium text-white/70 mb-1"
            >
              Page Title
            </label>
            <input
              id="status-title-input"
              type="text"
              required
              placeholder="e.g. Acme Network Status"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              className="w-full rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-xs text-white placeholder-white/30 focus:border-indigo-500 focus:outline-hidden"
            />
          </div>

          <div>
            <label
              htmlFor="status-desc-input"
              className="block text-xs font-medium text-white/70 mb-1"
            >
              Description / Header Message
            </label>
            <textarea
              id="status-desc-input"
              rows={2}
              placeholder="Public message shown to your users..."
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="w-full rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-xs text-white placeholder-white/30 focus:border-indigo-500 focus:outline-hidden"
            />
          </div>

          <div className="flex items-center gap-2 pt-1">
            <input
              id="status-public-checkbox"
              type="checkbox"
              checked={isPublic}
              onChange={(e) => setIsPublic(e.target.checked)}
              className="rounded border-white/20 bg-white/10 text-indigo-600 focus:ring-0"
            />
            <label
              htmlFor="status-public-checkbox"
              className="text-xs text-white/90 cursor-pointer"
            >
              Enable public status page access
            </label>
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
              {isSaving ? 'Saving...' : 'Save Settings'}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
