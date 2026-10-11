import type {
  AgentEvidenceReference,
  AgentSavedMessage,
  AgentStep,
  AgentStreamEvent,
} from '../protocol';

export interface AgentInvestigationContext {
  prompt: string;
  sourceRequestId?: string | null;
  history: AgentSavedMessage[];
  signal: AbortSignal;
  emit: (event: AgentStreamEvent) => Promise<void> | void;
}

export interface AgentInvestigationResult {
  text: string;
  steps: AgentStep[];
  evidence: AgentEvidenceReference[];
  inputTokens?: number;
  outputTokens?: number;
}

export interface AgentProvider {
  readonly name: string;
  readonly model: string;
  investigate(
    context: AgentInvestigationContext,
  ): Promise<AgentInvestigationResult>;
}
