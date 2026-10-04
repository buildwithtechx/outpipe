import { useState } from 'react';
import { Button } from '#/components/ui/button';
import type {
  ReplayRequestInput,
  ReplayResponseOutput,
} from '../services/observability-service';

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

  if (!isOpen) return null;

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
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-xs p-4">
      <div className="w-full max-w-xl rounded-2xl border border-white/10 bg-zinc-900 p-6 shadow-2xl text-white max-h-[90vh] flex flex-col">
        <div className="flex items-center justify-between border-b border-white/10 pb-4 mb-4">
          <h3 className="text-base font-semibold text-white">
            Replay HTTP Request
          </h3>
          <button
            type="button"
            onClick={onClose}
            className="text-white/50 hover:text-white text-lg"
          >
            &times;
          </button>
        </div>

        <form
          onSubmit={handleSubmit}
          className="space-y-4 flex-1 overflow-y-auto pr-1"
        >
          <div className="flex gap-2">
            <div className="w-28 shrink-0">
              <label
                htmlFor="replay-method-select"
                className="block text-xs font-medium text-white/70 mb-1"
              >
                Method
              </label>
              <select
                id="replay-method-select"
                value={method}
                onChange={(e) => setMethod(e.target.value)}
                className="w-full rounded-lg border border-white/10 bg-zinc-800 px-3 py-2 text-xs font-mono text-white focus:border-indigo-500 focus:outline-hidden"
              >
                <option value="GET">GET</option>
                <option value="POST">POST</option>
                <option value="PUT">PUT</option>
                <option value="PATCH">PATCH</option>
                <option value="DELETE">DELETE</option>
              </select>
            </div>
            <div className="flex-1">
              <label
                htmlFor="replay-url-input"
                className="block text-xs font-medium text-white/70 mb-1"
              >
                Target URL
              </label>
              <input
                id="replay-url-input"
                type="url"
                required
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                className="w-full rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-xs font-mono text-white placeholder-white/30 focus:border-indigo-500 focus:outline-hidden"
              />
            </div>
          </div>

          <div>
            <label
              htmlFor="replay-body-input"
              className="block text-xs font-medium text-white/70 mb-1"
            >
              Request Body Payload
            </label>
            <textarea
              id="replay-body-input"
              rows={4}
              value={body}
              onChange={(e) => setBody(e.target.value)}
              className="w-full rounded-lg border border-white/10 bg-zinc-950 px-3 py-2 text-xs font-mono text-white placeholder-white/30 focus:border-indigo-500 focus:outline-hidden"
            />
          </div>

          <div className="flex justify-end gap-2">
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
              disabled={isExecuting}
              className="text-xs bg-indigo-600 hover:bg-indigo-500"
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
      </div>
    </div>
  );
}
