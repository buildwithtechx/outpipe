import { useState } from 'react';
import type { TelemetrySpan } from '../services/observability-service';

interface TraceWaterfallProps {
  traceId: string;
  spans: TelemetrySpan[];
  onClose: () => void;
}

export function TraceWaterfall({
  traceId,
  spans,
  onClose,
}: TraceWaterfallProps) {
  const [selectedSpanId, setSelectedSpanId] = useState<string | null>(null);

  if (spans.length === 0) {
    return (
      <div className="rounded-xl border border-white/10 bg-zinc-900 p-6 text-center text-white/50 text-xs">
        No spans found for trace {traceId}.
      </div>
    );
  }

  const sortedSpans = [...spans].sort(
    (a, b) => new Date(a.startTime).getTime() - new Date(b.startTime).getTime(),
  );

  const traceStart = new Date(sortedSpans[0].startTime).getTime();
  const traceEnd = Math.max(
    ...sortedSpans.map((s) => new Date(s.endTime || s.startTime).getTime()),
  );
  const totalDuration = Math.max(traceEnd - traceStart, 1);

  const selectedSpan = sortedSpans.find((s) => s.id === selectedSpanId);

  return (
    <div className="rounded-xl border border-white/10 bg-zinc-900/90 p-5 space-y-4 text-white">
      <div className="flex items-center justify-between border-b border-white/10 pb-3">
        <div>
          <div className="flex items-center gap-2">
            <span className="text-xs font-mono uppercase bg-indigo-500/20 text-indigo-300 px-2 py-0.5 rounded">
              Trace Waterfall
            </span>
            <span className="text-xs font-mono text-white/60">{traceId}</span>
          </div>
          <p className="text-xs text-white/40 mt-1">
            Total latency: {totalDuration}ms &bull; {spans.length} spans
          </p>
        </div>
        <button
          type="button"
          onClick={onClose}
          className="text-xs text-white/60 hover:text-white px-2.5 py-1 rounded border border-white/10"
        >
          Close Waterfall
        </button>
      </div>

      <div className="space-y-2">
        {sortedSpans.map((span) => {
          const spanStart = new Date(span.startTime).getTime();
          const spanDuration = span.durationMs || 1;
          const leftPercent = Math.max(
            0,
            Math.min(100, ((spanStart - traceStart) / totalDuration) * 100),
          );
          const widthPercent = Math.max(
            1,
            Math.min(100 - leftPercent, (spanDuration / totalDuration) * 100),
          );

          const isError =
            span.statusCode === 'ERROR' || span.statusCode === 'error';
          const isSelected = span.id === selectedSpanId;

          return (
            <button
              type="button"
              key={span.id}
              onClick={() => setSelectedSpanId(isSelected ? null : span.id)}
              className={`w-full text-left p-2.5 rounded-lg border transition-colors cursor-pointer text-xs ${
                isSelected
                  ? 'border-indigo-500 bg-white/10'
                  : 'border-white/5 bg-white/5 hover:bg-white/10'
              }`}
            >
              <div className="flex items-center justify-between mb-1.5 font-mono">
                <div className="flex items-center gap-2 truncate pr-2">
                  <span className="font-semibold text-white truncate">
                    {span.name}
                  </span>
                  <span className="text-[10px] text-white/40 uppercase">
                    {span.kind || 'INTERNAL'}
                  </span>
                </div>
                <span
                  className={`font-semibold shrink-0 ${isError ? 'text-rose-400' : 'text-emerald-400'}`}
                >
                  {span.durationMs}ms
                </span>
              </div>

              <div className="w-full bg-white/5 h-2 rounded-full overflow-hidden relative">
                <div
                  className={`h-full rounded-full absolute ${
                    isError
                      ? 'bg-rose-500'
                      : 'bg-gradient-to-r from-indigo-500 to-cyan-400'
                  }`}
                  style={{
                    left: `${leftPercent}%`,
                    width: `${widthPercent}%`,
                  }}
                />
              </div>
            </button>
          );
        })}
      </div>

      {selectedSpan && (
        <div className="mt-4 p-3 rounded-lg border border-white/10 bg-black/40 text-xs font-mono space-y-2">
          <div className="text-white font-semibold">
            Span Details: {selectedSpan.name}
          </div>
          <div className="grid grid-cols-2 gap-2 text-white/60">
            <div>
              Span ID: <span className="text-white">{selectedSpan.spanId}</span>
            </div>
            <div>
              Parent ID:{' '}
              <span className="text-white">
                {selectedSpan.parentSpanId || '(root)'}
              </span>
            </div>
            <div>
              Status:{' '}
              <span className="text-white">
                {selectedSpan.statusCode || 'OK'}
              </span>
            </div>
            <div>
              Duration:{' '}
              <span className="text-white">{selectedSpan.durationMs}ms</span>
            </div>
          </div>
          {selectedSpan.attributes && (
            <div className="mt-2">
              <div className="text-white/60 mb-1">Attributes:</div>
              <pre className="p-2 bg-white/5 rounded text-[11px] overflow-x-auto text-emerald-300">
                {selectedSpan.attributes}
              </pre>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
