import { useState } from 'react';
import { Button } from '#/components/ui/button';

interface ProjectModalProps {
  isOpen: boolean;
  onClose: () => void;
  onCreate: (input: { slug: string; name: string }) => Promise<void>;
  isCreating: boolean;
}

export function ProjectModal({
  isOpen,
  onClose,
  onCreate,
  isCreating,
}: ProjectModalProps) {
  const [name, setName] = useState('');
  const [slug, setSlug] = useState('');

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim() || !slug.trim()) return;
    await onCreate({ slug: slug.trim().toLowerCase(), name: name.trim() });
    setName('');
    setSlug('');
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-xs p-4">
      <div className="w-full max-w-sm rounded-2xl border border-white/10 bg-zinc-900 p-6 shadow-2xl text-white">
        <h3 className="text-base font-semibold text-white mb-4">
          New Secrets Project
        </h3>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label
              htmlFor="project-name-input"
              className="block text-xs font-medium text-white/70 mb-1"
            >
              Project Name
            </label>
            <input
              id="project-name-input"
              type="text"
              required
              placeholder="e.g. Core API"
              value={name}
              onChange={(e) => {
                setName(e.target.value);
                if (!slug) {
                  setSlug(
                    e.target.value
                      .toLowerCase()
                      .replace(/[^a-z0-9]/g, '-')
                      .replace(/-+/g, '-'),
                  );
                }
              }}
              className="w-full rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-xs text-white placeholder-white/30 focus:border-indigo-500 focus:outline-hidden"
            />
          </div>
          <div>
            <label
              htmlFor="project-slug-input"
              className="block text-xs font-medium text-white/70 mb-1"
            >
              Project Slug
            </label>
            <input
              id="project-slug-input"
              type="text"
              required
              placeholder="e.g. core-api"
              value={slug}
              onChange={(e) => setSlug(e.target.value)}
              className="w-full rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-xs font-mono text-white placeholder-white/30 focus:border-indigo-500 focus:outline-hidden"
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
              disabled={isCreating}
              className="text-xs bg-indigo-600 hover:bg-indigo-500"
            >
              Create
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
