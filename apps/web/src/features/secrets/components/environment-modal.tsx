import { useState } from 'react';
import { Button } from '#/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '#/components/ui/dialog';
import { Input } from '#/components/ui/input';
import { Label } from '#/components/ui/label';

interface EnvironmentModalProps {
  isOpen: boolean;
  onClose: () => void;
  onCreate: (input: { slug: string; name: string }) => Promise<void>;
  isCreating: boolean;
}

export function EnvironmentModal({
  isOpen,
  onClose,
  onCreate,
  isCreating,
}: EnvironmentModalProps) {
  const [name, setName] = useState('');
  const [slug, setSlug] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim() || !slug.trim()) return;
    await onCreate({ slug: slug.trim().toLowerCase(), name: name.trim() });
    setName('');
    setSlug('');
  };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-w-sm border-white/10 bg-zinc-900 text-white">
        <DialogHeader>
          <DialogTitle className="text-base font-semibold text-white">
            New Environment
          </DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4 pt-2">
          <div className="space-y-1.5">
            <Label htmlFor="env-name-input" className="text-xs text-white/70">
              Environment Name
            </Label>
            <Input
              id="env-name-input"
              type="text"
              required
              placeholder="e.g. Staging"
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
              className="border-white/10 bg-white/5 text-xs text-white placeholder-white/30"
            />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="env-slug-input" className="text-xs text-white/70">
              Slug
            </Label>
            <Input
              id="env-slug-input"
              type="text"
              required
              placeholder="e.g. staging"
              value={slug}
              onChange={(e) => setSlug(e.target.value)}
              className="border-white/10 bg-white/5 text-xs font-mono text-white placeholder-white/30"
            />
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={onClose}
              className="text-xs border-white/10 bg-white/5 text-white/70 hover:text-white"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              size="sm"
              disabled={isCreating}
              className="text-xs bg-indigo-600 hover:bg-indigo-500 text-white"
            >
              {isCreating ? 'Creating...' : 'Create'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
