import { AlertTriangle, Bot, ChevronDown, Sparkles, User } from 'lucide-react';
import { useState } from 'react';
import type { AgentMessage } from './agent-chat-data';
import { AgentEvidenceCards } from './agent-evidence-cards';
import { AgentPreparation } from './agent-preparation';

export interface AgentMessageResponseProps {
  message: AgentMessage;
  provider: string;
  model: string;
}

export function AgentMessageResponse({
  message,
  provider,
  model,
}: AgentMessageResponseProps) {
  const isUser = message.role === 'user';
  const [showSteps, setShowSteps] = useState(
    message.status === 'running' || message.steps.length > 0,
  );

  if (isUser) {
    return (
      <div className="flex items-start gap-3 justify-end pl-8">
        <div className="rounded-2xl rounded-tr-sm bg-indigo-600 px-4 py-2.5 text-xs text-white shadow-sm max-w-xl leading-relaxed whitespace-pre-wrap">
          {message.text}
        </div>
        <div className="flex size-7 items-center justify-center rounded-lg bg-white/10 text-white/70 shrink-0 mt-0.5">
          <User className="size-3.5" />
        </div>
      </div>
    );
  }

  return (
    <div className="flex items-start gap-3 pr-8">
      <div className="flex size-7 items-center justify-center rounded-lg bg-indigo-500/20 text-indigo-400 shrink-0 mt-0.5 border border-indigo-500/30">
        <Bot className="size-4" />
      </div>

      <div className="space-y-3 min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <span className="font-semibold text-xs text-white">
            Outpipe Troubleshooting Agent
          </span>
          <span className="rounded bg-white/10 px-1.5 py-0.5 text-[10px] font-mono text-white/60">
            {provider}:{model}
          </span>
          {message.status === 'running' && (
            <span className="flex items-center gap-1 text-[10px] text-indigo-400 animate-pulse">
              <Sparkles className="size-2.5" />
              <span>Investigating…</span>
            </span>
          )}
        </div>

        {message.steps.length > 0 && (
          <div className="space-y-1">
            <button
              type="button"
              onClick={() => setShowSteps(!showSteps)}
              className="flex items-center gap-1.5 text-[11px] text-white/50 hover:text-white/80 transition"
            >
              <span>{message.steps.length} investigation steps</span>
              <ChevronDown
                className={`size-3 transition-transform ${
                  showSteps ? 'rotate-180' : ''
                }`}
              />
            </button>
            {showSteps && <AgentPreparation steps={message.steps} />}
          </div>
        )}

        {message.evidence.length > 0 && (
          <AgentEvidenceCards evidence={message.evidence} />
        )}

        {message.text && (
          <div className="rounded-xl border border-white/10 bg-white/2 p-3.5 text-xs text-white/90 leading-relaxed space-y-2 whitespace-pre-wrap font-sans">
            {message.text}
          </div>
        )}

        {message.error && (
          <div className="rounded-lg border border-rose-500/30 bg-rose-500/10 p-3 text-xs text-rose-300 flex items-start gap-2">
            <AlertTriangle className="size-4 shrink-0 text-rose-400 mt-0.5" />
            <div>
              <span className="font-semibold block">Investigation Error</span>
              <p className="text-[11px] text-rose-200/80 mt-0.5">
                {message.error}
              </p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
