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
import type {
  UptimeComponent,
  UptimeComponentGroup,
} from '#/components/uptime/incident-component-data';
import { IncidentComponentPicker } from '#/components/uptime/incident-component-picker';
import { useUptimeUnsavedChanges } from '#/components/uptime/use-uptime-unsaved-changes';
import { useSubmitFeedback } from '#/hooks/use-submit-feedback';

interface IncidentModalProps {
  isOpen: boolean;
  onClose: () => void;
  components?: UptimeComponent[];
  groups?: UptimeComponentGroup[];
  onSave: (input: {
    title: string;
    severity: 'minor' | 'major' | 'critical';
    message: string;
    componentIds?: string[];
  }) => Promise<void>;
  isSaving: boolean;
}

export function IncidentModal({
  isOpen,
  onClose,
  components = [],
  groups = [],
  onSave,
  isSaving,
}: IncidentModalProps) {
  const [title, setTitle] = useState('');
  const [severity, setSeverity] = useState<'minor' | 'major' | 'critical'>(
    'minor',
  );
  const [message, setMessage] = useState('');
  const [selectedComponentIds, setSelectedComponentIds] = useState<string[]>(
    [],
  );

  const isDirty =
    isOpen &&
    Boolean(title.trim() || message.trim() || selectedComponentIds.length > 0);
  useUptimeUnsavedChanges(isDirty);

  const { error, submit } = useSubmitFeedback(isOpen);

  const handleToggleComponent = (id: string, checked: boolean) => {
    setSelectedComponentIds((prev) =>
      checked ? [...prev, id] : prev.filter((item) => item !== id),
    );
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim() || !message.trim()) return;
    if (
      !(await submit(() =>
        onSave({
          title: title.trim(),
          severity,
          message: message.trim(),
          componentIds: selectedComponentIds,
        }),
      ))
    )
      return;
    setTitle('');
    setMessage('');
    setSelectedComponentIds([]);
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
            Report New Incident
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
                <SelectTrigger className="w-full border-white/10 bg-background text-xs text-white">
                  <SelectValue placeholder="Severity" />
                </SelectTrigger>
                <SelectContent className="border-white/10 bg-card text-white">
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

            {components.length > 0 && (
              <IncidentComponentPicker
                components={components}
                groups={groups}
                selected={selectedComponentIds}
                onToggle={handleToggleComponent}
              />
            )}

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
                {isSaving ? 'Reporting...' : 'Publish Incident'}
              </Button>
            </div>
          </fieldset>
        </form>
      </DialogContent>
    </Dialog>
  );
}
