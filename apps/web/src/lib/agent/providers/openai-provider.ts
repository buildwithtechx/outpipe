import type { AgentStep } from '../protocol';
import type {
  AgentInvestigationContext,
  AgentInvestigationResult,
  AgentProvider,
} from './types';

const SYSTEM_PROMPT = `You are Outpipe Agent, an expert read-only observability and tunneling investigator.
Analyze requests, status codes, traces, and latency patterns.
Give concise, structured Markdown conclusions with:
### Diagnostic Summary
### Key Observations
### Unconfirmed Possibilities
### Recommended Next Checks`;

export class OpenAIAgentProvider implements AgentProvider {
  readonly name = 'openai';
  readonly model: string;
  private readonly apiKey: string;

  constructor(apiKey: string, model = 'gpt-4o') {
    this.apiKey = apiKey;
    this.model = model;
  }

  async investigate(
    context: AgentInvestigationContext,
  ): Promise<AgentInvestigationResult> {
    const { prompt, history, signal, emit } = context;

    const steps: AgentStep[] = [
      {
        id: 'step-1',
        label: 'Connect to OpenAI & analyze telemetry',
        status: 'running',
      },
    ];
    await emit({ type: 'step', step: steps[0] });

    const messages = [
      { role: 'system', content: SYSTEM_PROMPT },
      ...history.map((m) => ({ role: m.role, content: m.text })),
      { role: 'user', content: prompt },
    ];

    const response = await fetch('https://api.openai.com/v1/chat/completions', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${this.apiKey}`,
      },
      body: JSON.stringify({
        model: this.model,
        messages,
        stream: true,
      }),
      signal,
    });

    if (!response.ok) {
      const errText = await response.text();
      steps[0].status = 'failed';
      steps[0].detail = `OpenAI API error: ${response.status}`;
      await emit({ type: 'step', step: steps[0] });
      throw new Error(`OpenAI error (${response.status}): ${errText}`);
    }

    steps[0].status = 'complete';
    steps[0].detail = `Investigating with ${this.model}`;
    await emit({ type: 'step', step: steps[0] });

    let fullText = '';
    const reader = response.body?.getReader();
    if (!reader) throw new Error('No response body stream');

    const decoder = new TextDecoder();
    let buffer = '';

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split('\n');
      buffer = lines.pop() || '';

      for (const line of lines) {
        const trimmed = line.trim();
        if (!trimmed.startsWith('data: ') || trimmed === 'data: [DONE]')
          continue;
        try {
          const parsed = JSON.parse(trimmed.slice(6));
          const delta = parsed.choices?.[0]?.delta?.content || '';
          if (delta) {
            fullText += delta;
            await emit({ type: 'text', delta });
          }
        } catch {
          // Ignore malformed SSE chunk
        }
      }
    }

    return {
      text: fullText,
      steps,
      evidence: [],
      inputTokens: 500,
      outputTokens: 300,
    };
  }
}
