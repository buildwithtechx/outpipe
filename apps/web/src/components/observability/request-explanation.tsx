import {
  AlertCircle,
  ArrowRight,
  Bot,
  Check,
  ChevronDown,
  Copy,
  HelpCircle,
  Info,
  RotateCcw,
  Sparkles,
} from 'lucide-react';
import { useEffect, useState } from 'react';
import { Button } from '#/components/ui/button';
import type { RequestCapture } from '#/interfaces';
import {
  buildRequestExplanationPreview,
  captureToSummary,
  type HttpRequestSummary,
  type RequestExplanationPreview,
} from './request-explanation-data';

export interface RequestExplanationProps {
  request?: HttpRequestSummary;
  capture?: RequestCapture;
  onAskAgent?: (preview: RequestExplanationPreview) => void;
  onInspectEvidence?: (target: 'response' | 'context') => void;
  autoStart?: boolean;
}

const PREPARATION_STEPS = [
  'Reading request telemetry',
  'Evaluating latency & status markers',
  'Synthesizing diagnostic findings',
];

export function RequestExplanation({
  request: rawRequest,
  capture,
  onAskAgent,
  onInspectEvidence,
  autoStart = true,
}: RequestExplanationProps) {
  const request = rawRequest ?? (capture ? captureToSummary(capture) : null);
  const [phase, setPhase] = useState<'idle' | 'preparing' | 'ready'>(
    autoStart ? 'preparing' : 'idle',
  );
  const [step, setStep] = useState(0);
  const [activeFollowUp, setActiveFollowUp] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const [showLimitations, setShowLimitations] = useState(false);

  useEffect(() => {
    if (phase !== 'preparing') return;
    const timers = [
      setTimeout(() => setStep(1), 250),
      setTimeout(() => setStep(2), 500),
      setTimeout(() => setPhase('ready'), 800),
    ];
    return () => timers.forEach(clearTimeout);
  }, [phase]);

  if (!request) {
    return (
      <div className="p-4 text-xs text-white/50 text-center">
        No request data available to analyze.
      </div>
    );
  }

  const preview = buildRequestExplanationPreview(request);

  const handleCopySummary = () => {
    const text = [
      `[Outpipe Request Analysis] ${preview.title}`,
      `Summary: ${preview.summary}`,
      ...(preview.hypothesis
        ? [`Hypothesis: ${preview.hypothesis.detail}`]
        : []),
      'Evidence:',
      ...preview.evidence.map((e) => `• ${e.label}: ${e.value} (${e.detail})`),
      'Recommended Next Steps:',
      ...preview.nextChecks.map((c, i) => `${i + 1}. ${c}`),
    ].join('\n');
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="rounded-xl border border-white/10 bg-zinc-900/90 overflow-hidden text-white text-xs">
      <div className="flex items-center justify-between px-4 py-3 border-b border-white/10 bg-white/2">
        <div className="flex items-center gap-2">
          <div className="flex size-7 items-center justify-center rounded-lg bg-indigo-500/20 text-indigo-400">
            <Sparkles className="size-4" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <span className="font-semibold text-white">
                Request Diagnostic Preview
              </span>
              <span className="px-1.5 py-0.5 rounded text-[10px] bg-white/10 text-white/70">
                Deterministic
              </span>
            </div>
            <p className="text-[11px] text-white/50">
              Heuristic analysis based on captured telemetry
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          {phase === 'ready' && onAskAgent && (
            <Button
              size="sm"
              onClick={() => onAskAgent(preview)}
              className="bg-indigo-600 hover:bg-indigo-500 text-white h-7 text-xs flex items-center gap-1.5"
            >
              <Bot className="size-3.5" />
              <span>Ask AI Agent</span>
            </Button>
          )}
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              setStep(0);
              setPhase('preparing');
            }}
            disabled={phase === 'preparing'}
            className="border-white/10 bg-white/5 hover:bg-white/10 text-xs h-7 text-white/80"
          >
            <RotateCcw className="size-3 mr-1" />
            {phase === 'preparing' ? 'Analyzing…' : 'Re-run'}
          </Button>
        </div>
      </div>

      {phase === 'preparing' && (
        <div className="p-6 space-y-4">
          <div className="space-y-2.5">
            {PREPARATION_STEPS.map((label, idx) => (
              <div key={label} className="flex items-center gap-2.5 text-xs">
                <span
                  className={`flex size-4 items-center justify-center rounded-full ${
                    idx < step
                      ? 'text-emerald-400'
                      : idx === step
                        ? 'text-indigo-400 animate-pulse'
                        : 'text-white/20'
                  }`}
                >
                  {idx < step ? (
                    <Check className="size-3" />
                  ) : (
                    <span className="size-1.5 rounded-full bg-current" />
                  )}
                </span>
                <span
                  className={idx <= step ? 'text-white/90' : 'text-white/30'}
                >
                  {label}
                </span>
              </div>
            ))}
          </div>
          <div className="space-y-1.5 pt-2 animate-pulse">
            <div className="h-3 w-4/5 rounded bg-white/5" />
            <div className="h-3 w-3/5 rounded bg-white/5" />
          </div>
        </div>
      )}

      {phase === 'ready' && (
        <div className="p-4 space-y-4">
          <div>
            <h4 className="text-sm font-semibold text-white">
              {preview.title}
            </h4>
            <p className="mt-1 text-white/70 leading-relaxed">
              {preview.summary}
            </p>
          </div>

          {preview.hypothesis && (
            <div className="p-3 rounded-lg border border-amber-500/20 bg-amber-500/10 flex items-start gap-2.5">
              <HelpCircle className="size-4 text-amber-400 shrink-0 mt-0.5" />
              <div>
                <span className="font-semibold text-amber-300 block">
                  {preview.hypothesis.title}
                </span>
                <p className="mt-0.5 text-amber-200/80 leading-normal text-[11px]">
                  {preview.hypothesis.detail}
                </p>
              </div>
            </div>
          )}

          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <span className="font-medium text-white/80">
                Observed Evidence
              </span>
              {onInspectEvidence && (
                <span className="text-[11px] text-white/40">
                  Select item to inspect
                </span>
              )}
            </div>
            <div className="rounded-lg border border-white/10 divide-y divide-white/10 overflow-hidden bg-black/20">
              {preview.evidence.map((item) => (
                <button
                  type="button"
                  key={item.id}
                  disabled={!onInspectEvidence}
                  onClick={() => onInspectEvidence?.(item.target)}
                  className={`w-full flex items-center justify-between px-3 py-2 text-left transition ${
                    onInspectEvidence
                      ? 'hover:bg-white/5 cursor-pointer'
                      : 'cursor-default'
                  }`}
                >
                  <div className="min-w-0 pr-2">
                    <span className="font-medium text-white/90 block">
                      {item.label}
                    </span>
                    <span className="text-[11px] text-white/50 block truncate">
                      {item.detail}
                    </span>
                  </div>
                  <div className="flex items-center gap-2 shrink-0">
                    <span className="font-mono text-white/80 font-medium">
                      {item.value}
                    </span>
                    {onInspectEvidence && (
                      <ArrowRight className="size-3 text-white/30" />
                    )}
                  </div>
                </button>
              ))}
            </div>
          </div>

          {preview.nextChecks.length > 0 && (
            <div className="space-y-2">
              <span className="font-medium text-white/80">
                Suggested Next Checks
              </span>
              <ul className="space-y-1.5">
                {preview.nextChecks.map((check, idx) => (
                  <li
                    key={check}
                    className="flex items-start gap-2 text-white/70"
                  >
                    <span className="flex size-4 shrink-0 items-center justify-center rounded bg-white/10 text-[10px] font-mono text-white/80 mt-0.5">
                      {idx + 1}
                    </span>
                    <span className="leading-tight">{check}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}

          <div className="space-y-2">
            <span className="font-medium text-white/80">
              Follow-up Analysis
            </span>
            <div className="flex flex-wrap gap-1.5">
              {preview.followUps.map((q) => (
                <Button
                  key={q.id}
                  variant="outline"
                  size="sm"
                  onClick={() =>
                    setActiveFollowUp(activeFollowUp === q.id ? null : q.id)
                  }
                  className={`h-6 text-[11px] rounded-md border-white/10 ${
                    activeFollowUp === q.id
                      ? 'bg-indigo-600 text-white border-indigo-500'
                      : 'bg-white/5 text-white/70 hover:bg-white/10 hover:text-white'
                  }`}
                >
                  {q.label}
                </Button>
              ))}
            </div>
            {activeFollowUp && (
              <div className="p-3 rounded-lg border border-white/10 bg-white/3">
                <span className="text-[10px] uppercase font-semibold text-white/40 block mb-1">
                  Analysis Detail
                </span>
                <p className="text-white/80 leading-relaxed text-[11px]">
                  {
                    preview.followUps.find((q) => q.id === activeFollowUp)
                      ?.answer
                  }
                </p>
              </div>
            )}
          </div>

          <div className="border border-white/10 rounded-lg overflow-hidden">
            <button
              type="button"
              onClick={() => setShowLimitations(!showLimitations)}
              className="w-full flex items-center justify-between px-3 py-2 bg-white/2 hover:bg-white/5 text-white/60 hover:text-white/80"
            >
              <div className="flex items-center gap-1.5 text-[11px]">
                <Info className="size-3.5" />
                <span>Scope & Limitations</span>
              </div>
              <ChevronDown
                className={`size-3.5 transition-transform ${
                  showLimitations ? 'rotate-180' : ''
                }`}
              />
            </button>
            {showLimitations && (
              <ul className="p-3 space-y-1 bg-black/20 text-white/50 text-[11px] border-t border-white/10">
                {preview.limitations.map((lim) => (
                  <li key={lim}>• {lim}</li>
                ))}
              </ul>
            )}
          </div>

          <div className="pt-2 border-t border-white/10 flex items-center justify-between">
            <span className="text-[10px] text-white/40 flex items-center gap-1">
              <AlertCircle className="size-3" />
              Local heuristic analysis · No external telemetry transmission
            </span>
            <Button
              variant="ghost"
              size="sm"
              onClick={handleCopySummary}
              className="h-6 text-[11px] text-white/60 hover:text-white px-2"
            >
              <Copy className="size-3 mr-1" />
              {copied ? 'Copied' : 'Copy summary'}
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
