import { useState } from 'react';
import { Button } from '#/components/ui/button';

interface AddMonitorModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: (input: {
    name: string;
    type: 'http' | 'https' | 'tcp';
    target: string;
    interval_seconds: number;
    timeout_seconds: number;
    expected_status_code?: number;
  }) => Promise<void>;
  isSaving: boolean;
}

export function AddMonitorModal({
  isOpen,
  onClose,
  onSave,
  isSaving,
}: AddMonitorModalProps) {
  const [name, setName] = useState('');
  const [type, setType] = useState<'http' | 'https' | 'tcp'>('https');
  const [target, setTarget] = useState('');
  const [intervalSeconds, setIntervalSeconds] = useState(60);
  const [timeoutSeconds, setTimeoutSeconds] = useState(10);
  const [expectedCode, setExpectedCode] = useState(200);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim() || !target.trim()) return;
    await onSave({
      name: name.trim(),
      type,
      target: target.trim(),
      interval_seconds: Number(intervalSeconds) || 60,
      timeout_seconds: Number(timeoutSeconds) || 10,
      expected_status_code:
        type === 'tcp' ? undefined : Number(expectedCode) || 200,
    });
    setName('');
    setTarget('');
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-xs p-4">
      <div className="w-full max-w-md rounded-2xl border border-white/10 bg-zinc-900 p-6 shadow-2xl text-white">
        <h3 className="text-base font-semibold text-white mb-4">
          Add Uptime Monitor
        </h3>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label
              htmlFor="monitor-name-input"
              className="block text-xs font-medium text-white/70 mb-1"
            >
              Monitor Name
            </label>
            <input
              id="monitor-name-input"
              type="text"
              required
              placeholder="e.g. Production API Gateway"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-xs text-white placeholder-white/30 focus:border-indigo-500 focus:outline-hidden"
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label
                htmlFor="monitor-type-select"
                className="block text-xs font-medium text-white/70 mb-1"
              >
                Probe Protocol
              </label>
              <select
                id="monitor-type-select"
                value={type}
                onChange={(e) =>
                  setType(e.target.value as 'http' | 'https' | 'tcp')
                }
                className="w-full rounded-lg border border-white/10 bg-zinc-800 px-3 py-2 text-xs text-white focus:border-indigo-500 focus:outline-hidden"
              >
                <option value="https">HTTPS</option>
                <option value="http">HTTP</option>
                <option value="tcp">TCP Port</option>
              </select>
            </div>
            <div>
              <label
                htmlFor="monitor-interval-input"
                className="block text-xs font-medium text-white/70 mb-1"
              >
                Interval (sec)
              </label>
              <input
                id="monitor-interval-input"
                type="number"
                min={10}
                max={3600}
                value={intervalSeconds}
                onChange={(e) => setIntervalSeconds(Number(e.target.value))}
                className="w-full rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-xs text-white focus:border-indigo-500 focus:outline-hidden"
              />
            </div>
          </div>

          <div>
            <label
              htmlFor="monitor-target-input"
              className="block text-xs font-medium text-white/70 mb-1"
            >
              Target Endpoint / Host:Port
            </label>
            <input
              id="monitor-target-input"
              type="text"
              required
              placeholder={
                type === 'tcp'
                  ? '127.0.0.1:8080'
                  : 'https://api.example.com/health'
              }
              value={target}
              onChange={(e) => setTarget(e.target.value)}
              className="w-full rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-xs font-mono text-white placeholder-white/30 focus:border-indigo-500 focus:outline-hidden"
            />
          </div>

          {type !== 'tcp' && (
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label
                  htmlFor="monitor-status-code-input"
                  className="block text-xs font-medium text-white/70 mb-1"
                >
                  Expected Status Code
                </label>
                <input
                  id="monitor-status-code-input"
                  type="number"
                  value={expectedCode}
                  onChange={(e) => setExpectedCode(Number(e.target.value))}
                  className="w-full rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-xs text-white focus:border-indigo-500 focus:outline-hidden"
                />
              </div>
              <div>
                <label
                  htmlFor="monitor-timeout-input"
                  className="block text-xs font-medium text-white/70 mb-1"
                >
                  Timeout (sec)
                </label>
                <input
                  id="monitor-timeout-input"
                  type="number"
                  min={1}
                  max={60}
                  value={timeoutSeconds}
                  onChange={(e) => setTimeoutSeconds(Number(e.target.value))}
                  className="w-full rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-xs text-white focus:border-indigo-500 focus:outline-hidden"
                />
              </div>
            </div>
          )}

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
              {isSaving ? 'Creating...' : 'Create Monitor'}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
