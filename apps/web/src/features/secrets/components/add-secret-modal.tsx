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
import { Textarea } from '#/components/ui/textarea';
import { useSubmitFeedback } from '#/hooks/use-submit-feedback';
import type { CreateSecretInput } from '#/interfaces';

interface AddSecretModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: (input: CreateSecretInput) => Promise<void>;
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

  useEffect(() => {
    if (!isOpen) {
      setNewKey('');
      setNewValue('');
      setNewComment('');
    }
  }, [isOpen]);

  const { error, submit } = useSubmitFeedback(isOpen);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newKey.trim() || !newValue.trim()) return;
    if (
      !(await submit(() =>
        onSave({
          key: newKey.trim(),
          value: newValue,
          comment: newComment.trim() || undefined,
        }),
      ))
    )
      return;
    setNewKey('');
    setNewValue('');
    setNewComment('');
  };

  return (
    <Dialog
      open={isOpen}
      onOpenChange={(open) => {
        if (!open && !isSaving) onClose();
      }}
    >
      <DialogContent
        showCloseButton={!isSaving}
        aria-describedby={undefined}
        className="max-w-md border-white/10 bg-card text-white"
      >
        <DialogHeader>
          <DialogTitle className="text-base font-semibold text-white">
            Add Environment Secret
          </DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4 pt-2">
          <fieldset disabled={isSaving} className="contents space-y-4">
            {error && (
              <p role="alert" className="text-sm text-rose-300">
                Could not save. Your input has been kept; please try again.
              </p>
            )}
            <div className="space-y-1.5">
              <Label
                htmlFor="secret-key-input"
                className="text-xs text-white/70"
              >
                Key Name
              </Label>
              <Input
                id="secret-key-input"
                type="text"
                required
                placeholder="e.g. STRIPE_API_KEY"
                value={newKey}
                onChange={(e) => setNewKey(e.target.value)}
                className="border-white/10 bg-white/5 text-xs font-mono text-white placeholder-white/30"
              />
            </div>
            <div className="space-y-1.5">
              <Label
                htmlFor="secret-value-input"
                className="text-xs text-white/70"
              >
                Secret Value
              </Label>
              <Textarea
                id="secret-value-input"
                required
                rows={3}
                placeholder="Enter secret value..."
                value={newValue}
                onChange={(e) => setNewValue(e.target.value)}
                className="border-white/10 bg-white/5 text-xs font-mono text-white placeholder-white/30"
              />
            </div>
            <div className="space-y-1.5">
              <Label
                htmlFor="secret-comment-input"
                className="text-xs text-white/70"
              >
                Comment (optional)
              </Label>
              <Input
                id="secret-comment-input"
                type="text"
                placeholder="e.g. Live billing secret"
                value={newComment}
                onChange={(e) => setNewComment(e.target.value)}
                className="border-white/10 bg-white/5 text-xs text-white placeholder-white/30"
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
                className="text-xs bg-primary hover:bg-primary/90 text-primary-foreground"
              >
                {isSaving ? 'Saving...' : 'Save Secret'}
              </Button>
            </div>
          </fieldset>
        </form>
      </DialogContent>
    </Dialog>
  );
}
