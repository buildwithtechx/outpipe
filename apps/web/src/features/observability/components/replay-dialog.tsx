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
import type { ReplayRequestInput, ReplayResponseOutput } from '#/interfaces';

interface ReplayDialogProps {
  isOpen: boolean;
  onClose: () => void;
  initialInput?: Partial<ReplayRequestInput>;
  onExecute: (input: ReplayRequestInput) => Promise<ReplayResponseOutput>;
  isExecuting: boolean;
}

export function ReplayDialog({
  isOpen,
  onClose,
  initialInput,
  onExecute,
  isExecuting,
}: ReplayDialogProps) {
  const [method, setMethod] = useState(initialInput?.method || 'POST');
  const [url, setUrl] = useState(
    initialInput?.url || 'https://httpbin.org/post',
  );
  const [body, setBody] = useState(
    initialInput?.request_body || '{\n  "test": true\n}',
  );
  const [result, setResult] = useState<ReplayResponseOutput | null>(null);
  const [errorMsg, setErrorMsg] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!url.trim()) return;
    setErrorMsg('');
    try {
      const res = await onExecute({
        method,
        url: url.trim(),
        request_body: body,
      });
      setResult(res);
    } catch (err: unknown) {
      setErrorMsg(err instanceof Error ? err.message : 'Replay failed');
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-w-xl border-white/10 bg-zinc-900 text-white max-h-[90vh] overflow-y-auto">
        <DialogHeader className="border-b border-white/10 pb-4">
          <DialogTitle className="text-base font-semibold text-white">
            Replay HTTP Request
          </DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4 pt-2">
          <div className="flex gap-2">
            <div className="w-32 shrink-0 space-y-1.5">
              <Label className="text-xs text-white/70">Method</Label>
              <Select value={method} onValueChange={setMethod}>
                <SelectTrigger className="w-full border-white/10 bg-zinc-800 text-xs font-mono text-white">
                  <SelectValue placeholder="Method" />
                </SelectTrigger>
                <SelectContent className="border-white/10 bg-zinc-900 text-white">
                  <SelectItem value="GET">GET</SelectItem>
                  <SelectItem value="POST">POST</SelectItem>
                  <SelectItem value="PUT">PUT</SelectItem>
                  <SelectItem value="PATCH">PATCH</SelectItem>
                  <SelectItem value="DELETE">DELETE</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="flex-1 space-y-1.5">
              <Label
                htmlFor="replay-url-input"
                className="text-xs text-white/70"
              >
                Target URL
              </Label>
              <Input
                id="replay-url-input"
                type="url"
                required
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                className="border-white/10 bg-white/5 text-xs font-mono text-white placeholder-white/30"
              />
            </div>
          </div>

          <div className="space-y-1.5">
            <Label
              htmlFor="replay-body-input"
              className="text-xs text-white/70"
            >
              Request Body Payload
            </Label>
            <Textarea
              id="replay-body-input"
              rows={4}
              value={body}
              onChange={(e) => setBody(e.target.value)}
              className="border-white/10 bg-zinc-950 text-xs font-mono text-white placeholder-white/30"
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
              disabled={isExecuting}
              className="text-xs bg-indigo-600 hover:bg-indigo-500 text-white"
            >
              {isExecuting ? 'Sending...' : 'Send Request'}
            </Button>
          </div>

          {errorMsg && (
            <div className="p-3 rounded-lg bg-rose-500/10 border border-rose-500/20 text-rose-300 text-xs">
              {errorMsg}
            </div>
          )}

          {result && (
            <div className="mt-4 p-4 rounded-xl border border-white/10 bg-zinc-950 font-mono text-xs space-y-2">
              <div className="flex items-center justify-between pb-2 border-b border-white/10">
                <span
                  className={`px-2 py-0.5 rounded text-xs font-bold ${
                    result.status_code < 400
                      ? 'bg-emerald-500/20 text-emerald-400'
                      : 'bg-rose-500/20 text-rose-400'
                  }`}
                >
                  HTTP {result.status_code} {result.status_text}
                </span>
                <span className="text-white/60">
                  Duration:{' '}
                  <strong className="text-white">{result.duration_ms}ms</strong>
                </span>
              </div>
              <div>
                <span className="text-white/50 block mb-1">Response Body:</span>
                <pre className="p-2.5 rounded bg-black/40 text-emerald-300 overflow-x-auto max-h-48 text-[11px] whitespace-pre-wrap">
                  {result.body || '(empty response)'}
                </pre>
              </div>
            </div>
          )}
        </form>
      </DialogContent>
    </Dialog>
  );
}
