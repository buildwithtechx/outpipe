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

interface AddMonitorModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: (input: {
    name: string;
    type: 'http' | 'https' | 'tcp' | 'icmp';
    target: string;
    interval_seconds: number;
    timeout_seconds: number;
    expected_status_code?: number;
    body_regex?: string;
    max_latency_ms?: number;
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
  const [type, setType] = useState<'http' | 'https' | 'tcp' | 'icmp'>('https');
  const [target, setTarget] = useState('');
  const [intervalSeconds, setIntervalSeconds] = useState(60);
  const [timeoutSeconds, setTimeoutSeconds] = useState(10);
  const [expectedCode, setExpectedCode] = useState(200);
  const [bodyRegex, setBodyRegex] = useState('');
  const [maxLatencyMs, setMaxLatencyMs] = useState(0);

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
        type === 'tcp' || type === 'icmp'
          ? undefined
          : Number(expectedCode) || 200,
      body_regex: type === 'http' || type === 'https' ? bodyRegex : undefined,
      max_latency_ms: maxLatencyMs || undefined,
    });
    setName('');
    setTarget('');
  };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-w-md border-white/10 bg-zinc-900 text-white">
        <DialogHeader>
          <DialogTitle className="text-base font-semibold text-white">
            Add Uptime Monitor
          </DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4 pt-2">
          <div className="space-y-1.5">
            <Label
              htmlFor="monitor-name-input"
              className="text-xs text-white/70"
            >
              Monitor Name
            </Label>
            <Input
              id="monitor-name-input"
              type="text"
              required
              placeholder="e.g. Production API Gateway"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="border-white/10 bg-white/5 text-xs text-white placeholder-white/30"
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <Label className="text-xs text-white/70">Probe Protocol</Label>
              <Select
                value={type}
                onValueChange={(val) =>
                  setType(val as 'http' | 'https' | 'tcp' | 'icmp')
                }
              >
                <SelectTrigger className="w-full border-white/10 bg-zinc-800 text-xs text-white">
                  <SelectValue placeholder="Protocol" />
                </SelectTrigger>
                <SelectContent className="border-white/10 bg-zinc-900 text-white">
                  <SelectItem value="https">HTTPS</SelectItem>
                  <SelectItem value="http">HTTP</SelectItem>
                  <SelectItem value="icmp">ICMP</SelectItem>
                  <SelectItem value="tcp">TCP Port</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-1.5">
              <Label
                htmlFor="monitor-interval-input"
                className="text-xs text-white/70"
              >
                Interval (sec)
              </Label>
              <Input
                id="monitor-interval-input"
                type="number"
                min={10}
                max={86400}
                value={intervalSeconds}
                onChange={(e) => setIntervalSeconds(Number(e.target.value))}
                className="border-white/10 bg-white/5 text-xs text-white"
              />
            </div>
          </div>

          <div className="space-y-1.5">
            <Label
              htmlFor="monitor-target-input"
              className="text-xs text-white/70"
            >
              Target Endpoint / Host:Port
            </Label>
            <Input
              id="monitor-target-input"
              type="text"
              required
              placeholder={
                type === 'tcp'
                  ? 'example.com:443'
                  : type === 'icmp'
                    ? 'example.com'
                    : `${type}://api.example.com/health`
              }
              value={target}
              onChange={(e) => setTarget(e.target.value)}
              className="border-white/10 bg-white/5 text-xs font-mono text-white placeholder-white/30"
            />
          </div>

          {(type === 'http' || type === 'https') && (
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label
                  htmlFor="monitor-status-code-input"
                  className="text-xs text-white/70"
                >
                  Expected Status Code
                </Label>
                <Input
                  id="monitor-status-code-input"
                  type="number"
                  value={expectedCode}
                  onChange={(e) => setExpectedCode(Number(e.target.value))}
                  className="border-white/10 bg-white/5 text-xs text-white"
                />
              </div>
              <div className="space-y-1.5">
                <Label
                  htmlFor="monitor-timeout-input"
                  className="text-xs text-white/70"
                >
                  Timeout (sec)
                </Label>
                <Input
                  id="monitor-timeout-input"
                  type="number"
                  min={1}
                  max={60}
                  value={timeoutSeconds}
                  onChange={(e) => setTimeoutSeconds(Number(e.target.value))}
                  className="border-white/10 bg-white/5 text-xs text-white"
                />
              </div>
            </div>
          )}

          {(type === 'http' || type === 'https') && (
            <div className="space-y-1.5">
              <Label htmlFor="monitor-body-regex">Body matches regex</Label>
              <Input
                id="monitor-body-regex"
                value={bodyRegex}
                maxLength={1024}
                onChange={(event) => setBodyRegex(event.target.value)}
                placeholder="Optional response assertion"
              />
            </div>
          )}
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <Label htmlFor="monitor-latency">Maximum latency (ms)</Label>
              <Input
                id="monitor-latency"
                type="number"
                min={0}
                max={60000}
                value={maxLatencyMs}
                onChange={(event) =>
                  setMaxLatencyMs(Number(event.target.value))
                }
              />
            </div>
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
              {isSaving ? 'Creating...' : 'Create Monitor'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
