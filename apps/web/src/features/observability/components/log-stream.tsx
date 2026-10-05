import { useState } from 'react';
import type { TelemetryLog } from '../services/observability-service';

interface LogStreamProps {
  logs: TelemetryLog[];
  onSelectTrace?: (traceId: string) => void;
}

export function LogStream({ logs, onSelectTrace }: LogStreamProps) {
  const [filterSeverity, setFilterSeverity] = useState('ALL');
  const [searchTerm, setSearchTerm] = useState('');

  const filteredLogs = logs.filter((log) => {
    const matchesSeverity =
      filterSeverity === 'ALL' ||
      log.severity.toUpperCase() === filterSeverity.toUpperCase();
    const matchesSearch =
      !searchTerm ||
      log.body.toLowerCase().includes(searchTerm.toLowerCase()) ||
      log.traceId?.toLowerCase().includes(searchTerm.toLowerCase());
    return matchesSeverity && matchesSearch;
  });

  function getSeverityColor(sev: string) {
    switch (sev.toUpperCase()) {
      case 'ERROR':
      case 'FATAL':
        return 'bg-rose-500/15 text-rose-400 border-rose-500/30';
      case 'WARN':
      case 'WARNING':
        return 'bg-amber-500/15 text-amber-400 border-amber-500/30';
      case 'DEBUG':
        return 'bg-purple-500/15 text-purple-400 border-purple-500/30';
      default:
        return 'bg-blue-500/15 text-blue-400 border-blue-500/30';
    }
  }

  return (
    <div className="space-y-3">
      <div className="flex flex-wrap items-center gap-3">
        <input
          type="text"
          placeholder="Search logs or trace IDs..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          className="flex-1 min-w-[200px] rounded-lg border border-white/10 bg-white/5 px-3 py-1.5 text-xs text-white placeholder-white/30 focus:border-indigo-500 focus:outline-hidden font-mono"
        />
        <select
          value={filterSeverity}
          onChange={(e) => setFilterSeverity(e.target.value)}
          className="rounded-lg border border-white/10 bg-zinc-800 px-3 py-1.5 text-xs text-white focus:border-indigo-500 focus:outline-hidden"
        >
          <option value="ALL">All Severities</option>
          <option value="INFO">INFO</option>
          <option value="WARN">WARN</option>
          <option value="ERROR">ERROR</option>
          <option value="DEBUG">DEBUG</option>
        </select>
        <span className="text-xs text-white/40">
          Showing {filteredLogs.length} logs
        </span>
      </div>

      <div className="rounded-xl border border-white/10 bg-zinc-950 font-mono text-xs overflow-hidden">
        {filteredLogs.length === 0 ? (
          <div className="p-8 text-center text-white/40">
            No telemetry logs match the current filters.
          </div>
        ) : (
          <div className="divide-y divide-white/5 max-h-[500px] overflow-y-auto">
            {filteredLogs.map((log) => (
              <div
                key={log.id}
                className="p-2.5 hover:bg-white/5 transition-colors flex items-start gap-3"
              >
                <span className="text-white/40 shrink-0 text-[11px]">
                  {new Date(log.timestamp).toLocaleTimeString()}
                </span>
                <span
                  className={`px-1.5 py-0.5 rounded text-[10px] font-bold border shrink-0 ${getSeverityColor(
                    log.severity,
                  )}`}
                >
                  {log.severity.toUpperCase()}
                </span>
                <div className="flex-1 break-all text-white/90">{log.body}</div>
                {log.traceId && onSelectTrace && (
                  <button
                    type="button"
                    onClick={() => onSelectTrace(log.traceId as string)}
                    className="text-[11px] text-indigo-400 hover:text-indigo-300 underline shrink-0 font-mono"
                  >
                    View Trace
                  </button>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
