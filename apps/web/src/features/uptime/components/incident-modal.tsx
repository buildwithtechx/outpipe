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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '#/components/ui/select';
import { Textarea } from '#/components/ui/textarea';

interface IncidentModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: (input: {
    title: string;
    severity: 'minor' | 'major' | 'critical';
    message: string;
  }) => Promise<void>;
  isSaving: boolean;
}

export function IncidentModal({
  isOpen,
  onClose,
  onSave,
  isSaving,
}: IncidentModalProps) {
  const [title, setTitle] = useState('');
  const [severity, setSeverity] = useState<'minor' | 'major' | 'critical'>(
    'minor',
  );
  const [message, setMessage] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim() || !message.trim()) return;
    await onSave({
      title: title.trim(),
      severity,
      message: message.trim(),
    });
    setTitle('');
    setMessage('');
  };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-w-md border-white/10 bg-zinc-900 text-white">
        <DialogHeader>
          <DialogTitle className="text-base font-semibold text-white">
            Report New Incident
          </DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4 pt-2">
          <div className="space-y-1.5">
            <Label
              htmlFor="incident-title-input"
              className="text-xs text-white/70"
            >
              Incident Title
            </Label>
            <Input
              id="incident-title-input"
              type="text"
              required
              placeholder="e.g. Latency spikes on US-East API"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              className="border-white/10 bg-white/5 text-xs text-white placeholder-white/30"
            />
          </div>

          <div className="space-y-1.5">
            <Label className="text-xs text-white/70">Severity Level</Label>
            <Select
              value={severity}
              onValueChange={(val) =>
                setSeverity(val as 'minor' | 'major' | 'critical')
              }
            >
              <SelectTrigger className="w-full border-white/10 bg-zinc-800 text-xs text-white">
                <SelectValue placeholder="Severity" />
              </SelectTrigger>
              <SelectContent className="border-white/10 bg-zinc-900 text-white">
                <SelectItem value="minor">Minor Degradation</SelectItem>
                <SelectItem value="major">Major Outage</SelectItem>
                <SelectItem value="critical">Critical Emergency</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-1.5">
            <Label
              htmlFor="incident-message-input"
              className="text-xs text-white/70"
            >
              Initial Status Message
            </Label>
            <Textarea
              id="incident-message-input"
              required
              rows={3}
              placeholder="Describe what is being investigated..."
              value={message}
              onChange={(e) => setMessage(e.target.value)}
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
              className="text-xs bg-rose-600 hover:bg-rose-500 text-white"
            >
              {isSaving ? 'Reporting...' : 'Publish Incident'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
