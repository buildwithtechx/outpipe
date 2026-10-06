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
import { useSubmitFeedback } from '#/hooks/use-submit-feedback';

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
  const [slugEdited, setSlugEdited] = useState(false);

  const { error, submit } = useSubmitFeedback(isOpen);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim() || !slug.trim()) return;
    if (
      !(await submit(() =>
        onCreate({ slug: slug.trim().toLowerCase(), name: name.trim() }),
      ))
    )
      return;
    setName('');
    setSlug('');
    setSlugEdited(false);
  };

  return (
    <Dialog
      open={isOpen}
      onOpenChange={(open) => {
        if (!open && !isCreating) onClose();
      }}
    >
      <DialogContent
        showCloseButton={!isCreating}
        aria-describedby={undefined}
        className="max-w-sm border-white/10 bg-card text-white"
      >
        <DialogHeader>
          <DialogTitle className="text-base font-semibold text-white">
            New Secrets Project
          </DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4 pt-2">
          <fieldset disabled={isCreating} className="contents space-y-4">
            {error && (
              <p role="alert" className="text-sm text-rose-300">
                Could not save. Your input has been kept; please try again.
              </p>
            )}
            <div className="space-y-1.5">
              <Label
                htmlFor="project-name-input"
                className="text-xs text-white/70"
              >
                Project Name
              </Label>
              <Input
                id="project-name-input"
                type="text"
                required
                placeholder="e.g. Core API"
                value={name}
                onChange={(e) => {
                  setName(e.target.value);
                  if (!slugEdited) {
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
              <Label
                htmlFor="project-slug-input"
                className="text-xs text-white/70"
              >
                Project Slug
              </Label>
              <Input
                id="project-slug-input"
                type="text"
                required
                placeholder="e.g. core-api"
                value={slug}
                onChange={(e) => {
                  setSlugEdited(true);
                  setSlug(e.target.value);
                }}
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
                className="text-xs bg-primary hover:bg-primary/90 text-primary-foreground"
              >
                {isCreating ? 'Creating...' : 'Create'}
              </Button>
            </div>
          </fieldset>
        </form>
      </DialogContent>
    </Dialog>
  );
}
