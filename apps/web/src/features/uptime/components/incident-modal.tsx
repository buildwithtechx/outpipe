import { useState } from 'react';
import { Button } from '#/components/ui/button';

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

  if (!isOpen) return null;

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
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-xs p-4">
      <div className="w-full max-w-md rounded-2xl border border-white/10 bg-zinc-900 p-6 shadow-2xl text-white">
        <h3 className="text-base font-semibold text-white mb-4">
          Report New Incident
        </h3>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label
              htmlFor="incident-title-input"
              className="block text-xs font-medium text-white/70 mb-1"
            >
              Incident Title
            </label>
            <input
              id="incident-title-input"
              type="text"
              required
              placeholder="e.g. Latency spikes on US-East API"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              className="w-full rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-xs text-white placeholder-white/30 focus:border-indigo-500 focus:outline-hidden"
            />
          </div>

          <div>
            <label
              htmlFor="incident-severity-select"
              className="block text-xs font-medium text-white/70 mb-1"
            >
              Severity Level
            </label>
            <select
              id="incident-severity-select"
              value={severity}
              onChange={(e) =>
                setSeverity(e.target.value as 'minor' | 'major' | 'critical')
              }
              className="w-full rounded-lg border border-white/10 bg-zinc-800 px-3 py-2 text-xs text-white focus:border-indigo-500 focus:outline-hidden"
            >
              <option value="minor">Minor Degradation</option>
              <option value="major">Major Outage</option>
              <option value="critical">Critical Emergency</option>
            </select>
          </div>

          <div>
            <label
              htmlFor="incident-message-input"
              className="block text-xs font-medium text-white/70 mb-1"
            >
              Initial Status Message
            </label>
            <textarea
              id="incident-message-input"
              required
              rows={3}
              placeholder="Describe what is being investigated..."
              value={message}
              onChange={(e) => setMessage(e.target.value)}
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
              className="text-xs bg-rose-600 hover:bg-rose-500"
            >
              {isSaving ? 'Reporting...' : 'Publish Incident'}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
