import { useEffect, useState } from 'react';
import { Button } from '#/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '#/components/ui/dialog';
import { Input } from '#/components/ui/input';
import { Label } from '#/components/ui/label';
import { Switch } from '#/components/ui/switch';
import { Textarea } from '#/components/ui/textarea';

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
  const [dirty, setDirty] = useState(false);
  useEffect(() => {
    if (!isOpen) {
      setDirty(false);
      return;
    }
    if (dirty) return;
    setSlug(initialSlug);
    setTitle(initialTitle);
    setDescription(initialDescription);
    setIsPublic(initialIsPublic);
  }, [
    isOpen,
    dirty,
    initialSlug,
    initialTitle,
    initialDescription,
    initialIsPublic,
  ]);

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
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-w-md border-white/10 bg-zinc-900 text-white">
        <DialogHeader>
          <DialogTitle className="text-base font-semibold text-white">
            Status Page Settings
          </DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4 pt-2">
          <div className="space-y-1.5">
            <Label
              htmlFor="status-slug-input"
              className="text-xs text-white/70"
            >
              Custom Slug URL
            </Label>
            <div className="flex items-center rounded-md border border-white/10 bg-white/5 px-3 py-1.5 text-xs font-mono text-white/50">
              <span>status.outpipe.dev/</span>
              <Input
                id="status-slug-input"
                type="text"
                required
                value={slug}
                onChange={(e) => {
                  setDirty(true);
                  setSlug(e.target.value);
                }}
                className="h-7 border-0 bg-transparent px-1 text-white font-mono placeholder-white/30 focus-visible:ring-0 shadow-none"
              />
            </div>
          </div>

          <div className="space-y-1.5">
            <Label
              htmlFor="status-title-input"
              className="text-xs text-white/70"
            >
              Page Title
            </Label>
            <Input
              id="status-title-input"
              type="text"
              required
              placeholder="e.g. Acme Network Status"
              value={title}
              onChange={(e) => {
                setDirty(true);
                setTitle(e.target.value);
              }}
              className="border-white/10 bg-white/5 text-xs text-white placeholder-white/30"
            />
          </div>

          <div className="space-y-1.5">
            <Label
              htmlFor="status-desc-input"
              className="text-xs text-white/70"
            >
              Description / Header Message
            </Label>
            <Textarea
              id="status-desc-input"
              rows={2}
              placeholder="Public message shown to your users..."
              value={description}
              onChange={(e) => {
                setDirty(true);
                setDescription(e.target.value);
              }}
              className="border-white/10 bg-white/5 text-xs text-white placeholder-white/30"
            />
          </div>

          <div className="flex items-center justify-between rounded-lg border border-white/10 bg-white/5 p-3">
            <div className="space-y-0.5">
              <Label
                htmlFor="status-public-switch"
                className="text-xs text-white cursor-pointer"
              >
                Public Status Page
              </Label>
              <p className="text-[11px] text-white/50">
                Allow anyone with the link to view real-time incident status
              </p>
            </div>
            <Switch
              id="status-public-switch"
              checked={isPublic}
              onCheckedChange={(value) => {
                setDirty(true);
                setIsPublic(value);
              }}
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
              disabled={isSaving}
              className="text-xs bg-indigo-600 hover:bg-indigo-500 text-white"
            >
              {isSaving ? 'Saving...' : 'Save Settings'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
