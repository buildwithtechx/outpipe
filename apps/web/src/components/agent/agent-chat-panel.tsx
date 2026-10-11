import {
  ArrowUp,
  History,
  Plus,
  Sparkles,
  StopCircle,
  Trash2,
  X,
} from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { Button } from '#/components/ui/button';
import { useAgentChat, useAgentChatState } from './agent-chat-context';
import { AgentMessageResponse } from './agent-message-response';

export interface AgentChatPanelProps {
  onClose: () => void;
}

export function AgentChatPanel({ onClose }: AgentChatPanelProps) {
  const { startThread } = useAgentChat();
  const { state, dispatch, sendMessage, stopInvestigation } =
    useAgentChatState();
  const [input, setInput] = useState('');
  const scrollRef = useRef<HTMLDivElement>(null);

  const activeThread =
    state.threads.find((t) => t.id === state.activeThreadId) ||
    state.threads[0];
  const isRunning = activeThread?.messages.some((m) => m.status === 'running');

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  });

  const handleSubmit = (e?: React.FormEvent) => {
    e?.preventDefault();
    const trimmed = input.trim();
    if (!trimmed || isRunning) return;
    sendMessage(trimmed);
    setInput('');
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit();
    }
  };

  return (
    <div className="flex h-full flex-col bg-zinc-950 text-white select-none">
      <div className="flex h-14 shrink-0 items-center justify-between border-b border-white/10 px-4">
        <div className="flex items-center gap-2.5">
          <div className="flex size-7 items-center justify-center rounded-lg bg-indigo-500/20 text-indigo-400 border border-indigo-500/30">
            <Sparkles className="size-3.5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <span className="font-semibold text-xs text-white">
                Troubleshooting Agent
              </span>
              {state.availableProviders &&
              state.availableProviders.length > 1 ? (
                <select
                  value={state.provider}
                  onChange={(e) =>
                    dispatch({
                      type: 'set-provider',
                      provider: e.target.value,
                    })
                  }
                  className="rounded bg-white/10 border border-white/10 px-1.5 py-0.5 text-[10px] font-mono text-white/80 focus:outline-none focus:border-indigo-500 cursor-pointer"
                  title="Select AI Provider"
                >
                  {state.availableProviders.map((p) => (
                    <option
                      key={p}
                      value={p}
                      className="bg-zinc-900 text-white"
                    >
                      {p}
                    </option>
                  ))}
                </select>
              ) : (
                <span className="rounded bg-white/10 px-1.5 py-0.5 text-[10px] font-mono text-white/60">
                  {state.provider}
                </span>
              )}
            </div>
          </div>
        </div>

        <div className="flex items-center gap-1.5">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => startThread()}
            className="size-8 p-0 text-white/60 hover:text-white hover:bg-white/5"
            title="Start New Investigation"
          >
            <Plus className="size-4" />
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => dispatch({ type: 'toggle-history' })}
            className={`size-8 p-0 text-white/60 hover:text-white hover:bg-white/5 ${
              state.historyOpen ? 'bg-white/10 text-white' : ''
            }`}
            title="Conversation History"
          >
            <History className="size-4" />
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={onClose}
            className="size-8 p-0 text-white/60 hover:text-white hover:bg-white/5"
            title="Close Panel"
          >
            <X className="size-4" />
          </Button>
        </div>
      </div>

      <div className="flex flex-1 min-h-0 relative">
        {state.historyOpen && (
          <div className="absolute inset-y-0 left-0 z-20 w-64 border-r border-white/10 bg-zinc-900/95 backdrop-blur-md p-3 flex flex-col space-y-2">
            <span className="text-[11px] font-semibold text-white/50 uppercase tracking-wider px-1">
              Investigations ({state.threads.length})
            </span>
            <div className="flex-1 overflow-y-auto space-y-1">
              {state.threads.map((t) => (
                <div
                  key={t.id}
                  className={`group flex items-center justify-between rounded-lg px-2.5 py-2 text-xs transition ${
                    t.id === activeThread?.id
                      ? 'bg-indigo-600 text-white'
                      : 'text-white/70 hover:bg-white/5 hover:text-white'
                  }`}
                >
                  <button
                    type="button"
                    onClick={() => dispatch({ type: 'select', threadId: t.id })}
                    className="truncate flex-1 text-left"
                  >
                    {t.title}
                  </button>
                  <button
                    type="button"
                    onClick={(e) => {
                      e.stopPropagation();
                      dispatch({ type: 'delete', threadId: t.id });
                    }}
                    className="opacity-0 group-hover:opacity-100 text-white/40 hover:text-rose-400 p-1"
                    title="Delete Thread"
                  >
                    <Trash2 className="size-3" />
                  </button>
                </div>
              ))}
              {state.threads.length === 0 && (
                <div className="p-4 text-center text-xs text-white/40">
                  No previous conversations
                </div>
              )}
            </div>
          </div>
        )}

        <div className="flex-1 flex flex-col min-h-0">
          <div
            ref={scrollRef}
            className="flex-1 overflow-y-auto p-4 space-y-4 select-text"
          >
            {(!activeThread || activeThread.messages.length === 0) && (
              <div className="py-8 text-center space-y-4 max-w-sm mx-auto">
                <div className="size-10 rounded-2xl bg-indigo-500/10 border border-indigo-500/20 text-indigo-400 mx-auto flex items-center justify-center">
                  <Sparkles className="size-5" />
                </div>
                <div>
                  <h3 className="text-sm font-semibold text-white">
                    Tunnel & Telemetry Investigator
                  </h3>
                  <p className="text-xs text-white/60 mt-1">
                    Ask questions about your live tunnels, latency spikes, or
                    correlate HTTP failure outcomes.
                  </p>
                </div>
                <div className="space-y-1.5 pt-2">
                  {[
                    'Explain why my tunnel is returning 503 errors',
                    'Where is latency concentrating in recent requests?',
                    'Check for upstream socket timeouts',
                  ].map((preset) => (
                    <button
                      key={preset}
                      type="button"
                      onClick={() => sendMessage(preset)}
                      className="w-full text-left text-xs p-2.5 rounded-lg border border-white/10 bg-white/2 hover:bg-white/5 text-white/70 hover:text-white transition"
                    >
                      • {preset}
                    </button>
                  ))}
                </div>
              </div>
            )}

            {activeThread?.messages.map((m) => (
              <AgentMessageResponse
                key={m.id}
                message={m}
                provider={state.provider}
                model={state.model}
              />
            ))}
          </div>

          <div className="border-t border-white/10 p-3 bg-zinc-900/50">
            {activeThread?.context && (
              <div className="mb-2 flex items-center justify-between rounded-md border border-indigo-500/30 bg-indigo-500/10 px-2.5 py-1 text-[11px] text-indigo-300">
                <span className="truncate">
                  Attached Context: {activeThread.context.label}
                </span>
                <span className="font-mono text-[10px]">
                  {activeThread.context.method} {activeThread.context.path}
                </span>
              </div>
            )}

            <form onSubmit={handleSubmit} className="relative">
              <textarea
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder="Ask about tunnel health, errors, or latency… (Enter to send)"
                rows={2}
                className="w-full resize-none rounded-xl border border-white/10 bg-black/40 px-3 py-2 pr-12 text-xs text-white placeholder-white/40 focus:border-indigo-500 focus:outline-none"
              />
              <div className="absolute bottom-3 right-3 flex items-center gap-1.5">
                {isRunning ? (
                  <Button
                    type="button"
                    size="sm"
                    onClick={() =>
                      activeThread && stopInvestigation(activeThread.id)
                    }
                    className="size-7 p-0 rounded-lg bg-rose-600 hover:bg-rose-500 text-white"
                    title="Stop Investigation"
                  >
                    <StopCircle className="size-3.5" />
                  </Button>
                ) : (
                  <Button
                    type="submit"
                    size="sm"
                    disabled={!input.trim()}
                    className="size-7 p-0 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white disabled:opacity-30"
                    title="Send Message"
                  >
                    <ArrowUp className="size-3.5" />
                  </Button>
                )}
              </div>
            </form>
          </div>
        </div>
      </div>
    </div>
  );
}
