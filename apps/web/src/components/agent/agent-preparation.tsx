import { Check, Loader2, X } from 'lucide-react';
import type { AgentStep } from '#/lib/agent/protocol';

export interface AgentPreparationProps {
  steps: AgentStep[];
}

export function AgentPreparation({ steps }: AgentPreparationProps) {
  if (steps.length === 0) return null;

  return (
    <div className="rounded-lg border border-white/10 bg-black/30 p-3 space-y-2">
      <span className="text-[11px] font-semibold text-white/50 uppercase tracking-wider block">
        Investigation Steps
      </span>
      <div className="space-y-1.5">
        {steps.map((step) => (
          <div key={step.id} className="flex items-start gap-2.5 text-xs">
            <span className="mt-0.5 shrink-0">
              {step.status === 'running' && (
                <Loader2 className="size-3.5 text-indigo-400 animate-spin" />
              )}
              {step.status === 'complete' && (
                <Check className="size-3.5 text-emerald-400" />
              )}
              {step.status === 'failed' && (
                <X className="size-3.5 text-rose-400" />
              )}
            </span>
            <div className="min-w-0 flex-1">
              <span
                className={
                  step.status === 'running'
                    ? 'text-indigo-300 font-medium'
                    : step.status === 'complete'
                      ? 'text-white/80'
                      : 'text-rose-300'
                }
              >
                {step.label}
              </span>
              {step.detail && (
                <p className="text-[11px] text-white/50 mt-0.5 leading-tight">
                  {step.detail}
                </p>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
