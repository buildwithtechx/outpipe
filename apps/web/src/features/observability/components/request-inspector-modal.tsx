import { useState } from 'react';
import { Button } from '#/components/ui/button';
import type { RequestCapture } from '../services/observability-service';

interface RequestInspectorModalProps {
  capture: RequestCapture | null;
  onClose: () => void;
  onReplay: (capture: RequestCapture) => void;
}

export function RequestInspectorModal({
  capture,
  onClose,
  onReplay,
}: RequestInspectorModalProps) {
  const [activeTab, setActiveTab] = useState<'request' | 'response'>('request');
  const [copied, setCopied] = useState(false);

  if (!capture) return null;

  const handleCopy = (text: string) => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const isSuccess = capture.statusCode >= 200 && capture.statusCode < 400;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-xs p-4">
      <div className="w-full max-w-2xl rounded-2xl border border-white/10 bg-zinc-900 p-6 shadow-2xl text-white max-h-[85vh] flex flex-col">
        <div className="flex items-center justify-between border-b border-white/10 pb-4">
          <div className="flex items-center gap-3">
            <span
              className={`px-2 py-0.5 rounded text-xs font-bold font-mono ${
                isSuccess
                  ? 'bg-emerald-500/20 text-emerald-400'
                  : 'bg-rose-500/20 text-rose-400'
              }`}
            >
              {capture.statusCode}
            </span>
            <span className="font-mono font-bold text-white text-sm">
              {capture.method}
            </span>
            <span className="font-mono text-white/70 text-xs truncate max-w-md">
              {capture.path}
            </span>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="text-white/50 hover:text-white text-lg"
          >
            &times;
          </button>
        </div>

        <div className="flex items-center justify-between py-3 border-b border-white/10 text-xs">
          <div className="flex gap-4 text-white/60">
            <span>
              Latency:{' '}
              <strong className="text-white">{capture.durationMs}ms</strong>
            </span>
            <span>
              Req Size:{' '}
              <strong className="text-white">
                {capture.requestBodySize} B
              </strong>
            </span>
            <span>
              Resp Size:{' '}
              <strong className="text-white">
                {capture.responseBodySize} B
              </strong>
            </span>
            <span>
              Time:{' '}
              <strong className="text-white">
                {new Date(capture.timestamp).toLocaleTimeString()}
              </strong>
            </span>
          </div>
          <Button
            size="sm"
            onClick={() => onReplay(capture)}
            className="bg-indigo-600 hover:bg-indigo-500 text-xs h-7"
          >
            Replay Request
          </Button>
        </div>

        <div className="flex items-center gap-4 border-b border-white/10 my-3">
          <button
            type="button"
            onClick={() => setActiveTab('request')}
            className={`pb-2 text-xs font-medium border-b-2 transition-colors ${
              activeTab === 'request'
                ? 'border-indigo-500 text-white'
                : 'border-transparent text-white/50'
            }`}
          >
            Request Headers & Body
          </button>
          <button
            type="button"
            onClick={() => setActiveTab('response')}
            className={`pb-2 text-xs font-medium border-b-2 transition-colors ${
              activeTab === 'response'
                ? 'border-indigo-500 text-white'
                : 'border-transparent text-white/50'
            }`}
          >
            Response Headers & Body
          </button>
        </div>

        <div className="flex-1 overflow-y-auto space-y-4 font-mono text-xs pr-1">
          {activeTab === 'request' ? (
            <>
              <div>
                <span className="text-white/50 block mb-1">
                  Request Headers:
                </span>
                <pre className="p-3 rounded-lg bg-zinc-950 border border-white/10 text-emerald-300 overflow-x-auto">
                  {capture.requestHeaders || 'No custom headers'}
                </pre>
              </div>
              <div>
                <div className="flex items-center justify-between mb-1">
                  <span className="text-white/50">Request Payload Body:</span>
                  <button
                    type="button"
                    onClick={() => handleCopy(capture.requestBody || '')}
                    className="text-[11px] text-indigo-400 hover:underline"
                  >
                    {copied ? 'Copied!' : 'Copy Body'}
                  </button>
                </div>
                <pre className="p-3 rounded-lg bg-zinc-950 border border-white/10 text-white/90 overflow-x-auto whitespace-pre-wrap">
                  {capture.requestBody || '(Empty body)'}
                </pre>
              </div>
            </>
          ) : (
            <>
              <div>
                <span className="text-white/50 block mb-1">
                  Response Headers:
                </span>
                <pre className="p-3 rounded-lg bg-zinc-950 border border-white/10 text-emerald-300 overflow-x-auto">
                  {capture.responseHeaders || 'No headers captured'}
                </pre>
              </div>
              <div>
                <div className="flex items-center justify-between mb-1">
                  <span className="text-white/50">Response Payload Body:</span>
                  <button
                    type="button"
                    onClick={() => handleCopy(capture.responseBody || '')}
                    className="text-[11px] text-indigo-400 hover:underline"
                  >
                    {copied ? 'Copied!' : 'Copy Body'}
                  </button>
                </div>
                <pre className="p-3 rounded-lg bg-zinc-950 border border-white/10 text-white/90 overflow-x-auto whitespace-pre-wrap">
                  {capture.responseBody || '(Empty body)'}
                </pre>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
