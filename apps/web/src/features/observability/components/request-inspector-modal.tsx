import { useState } from 'react';
import { Badge } from '#/components/ui/badge';
import { Button } from '#/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '#/components/ui/dialog';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '#/components/ui/tabs';
import type { RequestCapture } from '#/interfaces';

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
  const [copied, setCopied] = useState(false);

  if (!capture) return null;

  const handleCopy = (text: string) => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const isSuccess = capture.statusCode >= 200 && capture.statusCode < 400;

  return (
    <Dialog open={Boolean(capture)} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-w-2xl border-white/10 bg-zinc-900 text-white max-h-[85vh] flex flex-col">
        <DialogHeader className="border-b border-white/10 pb-4">
          <DialogTitle className="flex items-center gap-3">
            <Badge
              variant={isSuccess ? 'default' : 'destructive'}
              className={`font-mono text-xs ${
                isSuccess
                  ? 'bg-emerald-500/20 text-emerald-400 hover:bg-emerald-500/30'
                  : 'bg-rose-500/20 text-rose-400 hover:bg-rose-500/30'
              }`}
            >
              {capture.statusCode}
            </Badge>
            <span className="font-mono font-bold text-white text-sm">
              {capture.method}
            </span>
            <span className="font-mono text-white/70 text-xs truncate max-w-md">
              {capture.path}
            </span>
          </DialogTitle>
        </DialogHeader>

        <div className="flex items-center justify-between py-2 border-b border-white/10 text-xs">
          <div className="flex gap-4 text-white/60">
            <span>
              Latency:{' '}
              <strong className="text-white">{capture.durationMs}ms</strong>
            </span>
            <span>
              Req:{' '}
              <strong className="text-white">
                {capture.requestBodySize} B
              </strong>
            </span>
            <span>
              Resp:{' '}
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
            className="bg-indigo-600 hover:bg-indigo-500 text-xs h-7 text-white"
          >
            Replay Request
          </Button>
        </div>

        <Tabs
          defaultValue="request"
          className="flex-1 overflow-hidden flex flex-col mt-2"
        >
          <TabsList className="w-full justify-start border-b border-white/10 rounded-none bg-transparent p-0 h-auto">
            <TabsTrigger
              value="request"
              className="rounded-none border-b-2 border-transparent data-[state=active]:border-indigo-500 data-[state=active]:bg-transparent data-[state=active]:text-white text-white/50 text-xs pb-2"
            >
              Request Headers & Body
            </TabsTrigger>
            <TabsTrigger
              value="response"
              className="rounded-none border-b-2 border-transparent data-[state=active]:border-indigo-500 data-[state=active]:bg-transparent data-[state=active]:text-white text-white/50 text-xs pb-2"
            >
              Response Headers & Body
            </TabsTrigger>
          </TabsList>

          <TabsContent
            value="request"
            className="flex-1 overflow-y-auto space-y-4 font-mono text-xs pr-1 pt-3"
          >
            <div>
              <span className="text-white/50 block mb-1">Request Headers:</span>
              <pre className="p-3 rounded-lg bg-zinc-950 border border-white/10 text-emerald-300 overflow-x-auto">
                {capture.requestHeaders || 'No custom headers'}
              </pre>
            </div>
            <div>
              <div className="flex items-center justify-between mb-1">
                <span className="text-white/50">Request Payload Body:</span>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={() => handleCopy(capture.requestBody || '')}
                  className="h-6 text-[11px] text-indigo-400 hover:text-indigo-300 hover:bg-white/5 px-2"
                >
                  {copied ? 'Copied!' : 'Copy Body'}
                </Button>
              </div>
              <pre className="p-3 rounded-lg bg-zinc-950 border border-white/10 text-white/90 overflow-x-auto whitespace-pre-wrap">
                {capture.requestBody || '(Empty body)'}
              </pre>
            </div>
          </TabsContent>

          <TabsContent
            value="response"
            className="flex-1 overflow-y-auto space-y-4 font-mono text-xs pr-1 pt-3"
          >
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
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={() => handleCopy(capture.responseBody || '')}
                  className="h-6 text-[11px] text-indigo-400 hover:text-indigo-300 hover:bg-white/5 px-2"
                >
                  {copied ? 'Copied!' : 'Copy Body'}
                </Button>
              </div>
              <pre className="p-3 rounded-lg bg-zinc-950 border border-white/10 text-white/90 overflow-x-auto whitespace-pre-wrap">
                {capture.responseBody || '(Empty body)'}
              </pre>
            </div>
          </TabsContent>
        </Tabs>
      </DialogContent>
    </Dialog>
  );
}
