import {
  Activity,
  Clock,
  ExternalLink,
  GitCommit,
  Terminal,
} from 'lucide-react';
import type { AgentEvidenceReference } from '#/lib/agent/protocol';

export interface AgentEvidenceCardsProps {
  evidence: AgentEvidenceReference[];
}

export function AgentEvidenceCards({ evidence }: AgentEvidenceCardsProps) {
  if (evidence.length === 0) return null;

  return (
    <div className="space-y-2 pt-1">
      <span className="text-[11px] font-semibold text-white/50 uppercase tracking-wider block">
        Retrieved Evidence ({evidence.length})
      </span>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
        {evidence.map((item) => {
          const p = item.presentation;
          return (
            <div
              key={item.id}
              className="p-2.5 rounded-lg border border-white/10 bg-white/3 hover:bg-white/5 transition flex flex-col justify-between text-xs space-y-1.5"
            >
              <div className="flex items-center justify-between text-white/60">
                <span className="font-medium text-white/90 truncate flex items-center gap-1.5">
                  {p?.kind === 'request' && (
                    <Activity className="size-3.5 text-indigo-400 shrink-0" />
                  )}
                  {p?.kind === 'trace' && (
                    <Clock className="size-3.5 text-cyan-400 shrink-0" />
                  )}
                  {p?.kind === 'log' && (
                    <Terminal className="size-3.5 text-amber-400 shrink-0" />
                  )}
                  {p?.kind === 'comparison' && (
                    <GitCommit className="size-3.5 text-purple-400 shrink-0" />
                  )}
                  <span className="truncate">{item.label}</span>
                </span>
                {item.href && (
                  <a
                    href={item.href}
                    target="_blank"
                    rel="noreferrer"
                    className="text-white/40 hover:text-white shrink-0 ml-1"
                    title="View Evidence Details"
                  >
                    <ExternalLink className="size-3" />
                  </a>
                )}
              </div>

              {p?.kind === 'request' && (
                <div className="flex items-center justify-between font-mono text-[11px]">
                  <span className="text-white/70">
                    {p.method} {p.path}
                  </span>
                  <span
                    className={
                      (p.statusCode ?? 200) >= 500
                        ? 'text-rose-400'
                        : (p.statusCode ?? 200) >= 400
                          ? 'text-amber-400'
                          : 'text-emerald-400'
                    }
                  >
                    HTTP {p.statusCode} · {p.durationMs}ms
                  </span>
                </div>
              )}

              {p?.kind === 'trace' && (
                <div className="space-y-1 font-mono text-[11px]">
                  <div className="text-white/60">
                    {p.spanCount} spans recorded
                  </div>
                  {p.spans.slice(0, 2).map((s) => (
                    <div
                      key={s.spanId}
                      className="flex items-center justify-between text-[10px] text-white/50 truncate"
                    >
                      <span className="truncate">{s.operationName}</span>
                      <span className="shrink-0 ml-1">{s.durationMs}ms</span>
                    </div>
                  ))}
                </div>
              )}

              {p?.kind === 'comparison' && (
                <div className="flex items-center justify-between text-[11px] text-white/70">
                  <span>24h Baseline:</span>
                  <span className="font-mono">
                    p95: {p.p95DurationMs}ms · err: {p.errorRate}%
                  </span>
                </div>
              )}

              {p?.kind === 'log' && (
                <div className="font-mono text-[11px] text-white/70 truncate">
                  [{p.level.toUpperCase()}] {p.messageSummary}
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
