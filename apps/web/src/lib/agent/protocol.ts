export type AgentRunStatus = 'running' | 'complete' | 'failed' | 'cancelled';

export interface AgentStep {
  id: string;
  label: string;
  status: 'running' | 'complete' | 'failed';
  detail?: string;
}

export type AgentEvidencePresentation =
  | {
      kind: 'request';
      method: string | null;
      path: string | null;
      service: string | null;
      statusCode: number | null;
      durationMs: number | null;
      timestamp: string | null;
      captureState: 'metadata' | 'redacted' | 'full' | 'unknown';
      requestSizeBytes: number | null;
      responseSizeBytes: number | null;
    }
  | {
      kind: 'trace';
      spanCount: number;
      returnedSpanCount: number;
      truncated: boolean;
      spans: Array<{
        spanId: string;
        parentSpanId: string | null;
        operationName: string | null;
        service: string | null;
        startedAt: string | null;
        offsetMs: number | null;
        durationMs: number | null;
        status: 'ok' | 'error' | 'unknown';
      }>;
    }
  | {
      kind: 'comparison';
      hours: 1 | 24;
      service: string | null;
      path: string | null;
      totalRequests: number | null;
      errorRequests: number | null;
      errorRate: number | null;
      p95DurationMs: number | null;
      averageDurationMs: number | null;
    }
  | {
      kind: 'log';
      level: string;
      timestamp: string | null;
      service: string | null;
      messageSummary: string;
      spanId: string | null;
    };

export interface AgentEvidenceReference {
  id: string;
  label: string;
  href: string;
  observedAt: string;
  presentation?: AgentEvidencePresentation;
}

export interface AgentSavedMessage {
  id: string;
  role: 'user' | 'assistant';
  text: string;
  status: AgentRunStatus;
  steps: AgentStep[];
  evidence: AgentEvidenceReference[];
  error?: string;
}

export interface AgentSavedThread {
  id: string;
  title: string;
  sourceRequestId: string | null;
  createdAt: number;
  updatedAt: number;
  messages: AgentSavedMessage[];
}

export interface AgentChatRequest {
  threadId: string;
  clientMessageId: string;
  assistantId: string;
  message: string;
  sourceRequestId?: string;
  provider?: string;
  model?: string;
}

export type AgentStreamEvent =
  | { type: 'run'; runId: string }
  | { type: 'step'; step: AgentStep }
  | { type: 'evidence'; evidence: AgentEvidenceReference[] }
  | { type: 'text'; delta: string }
  | {
      type: 'finish';
      status: 'complete' | 'failed' | 'cancelled';
      error?: string;
    };

export interface AgentThreadListResponse {
  threads: AgentSavedThread[];
  configured: boolean;
  provider: string;
  model: string | null;
  availableProviders?: string[];
}
