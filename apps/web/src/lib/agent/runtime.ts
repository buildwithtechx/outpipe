import { type AgentConfig, readAgentConfig } from './config';
import type { AgentSavedMessage, AgentStreamEvent } from './protocol';
import { getAgentProvider } from './providers/registry';

export interface RunAgentOptions {
  config?: AgentConfig;
  prompt: string;
  sourceRequestId?: string | null;
  history?: AgentSavedMessage[];
  signal?: AbortSignal;
  emit: (event: AgentStreamEvent) => Promise<void> | void;
}

export async function runAgentInvestigation(options: RunAgentOptions) {
  const config = options.config || readAgentConfig();
  const provider = getAgentProvider(config);
  const controller = new AbortController();
  const signal = options.signal || controller.signal;

  const runId = crypto.randomUUID();
  await options.emit({ type: 'run', runId });

  try {
    const result = await provider.investigate({
      prompt: options.prompt,
      sourceRequestId: options.sourceRequestId,
      history: options.history || [],
      signal,
      emit: options.emit,
    });

    await options.emit({ type: 'finish', status: 'complete' });
    return result;
  } catch (error) {
    const isCancelled = signal.aborted;
    const errorMsg =
      error instanceof Error ? error.message : 'Investigation failed';
    await options.emit({
      type: 'finish',
      status: isCancelled ? 'cancelled' : 'failed',
      error: errorMsg,
    });
    throw error;
  }
}
