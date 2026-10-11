import { env } from '#/env';

export type AgentProviderType = 'openai' | 'xai' | 'anthropic' | 'mock';

export interface AgentConfig {
  configured: boolean;
  provider: AgentProviderType;
  model: string;
  apiKey?: string;
  timeoutMs: number;
  maxSteps: number;
  maxToolCalls: number;
  maxOutputTokens: number;
  maxTotalTokens: number;
}

export function readAgentConfig(
  customEnv?: Record<string, string | undefined>,
): AgentConfig {
  let explicitProvider = (
    customEnv?.DEFAULT_AGENT_PROVIDER ||
    customEnv?.AGENT_PROVIDER ||
    env.VITE_DEFAULT_AGENT_PROVIDER ||
    ''
  )
    .toLowerCase()
    .trim() as AgentProviderType;

  const openaiKey = customEnv?.OPENAI_API_KEY?.trim();
  const xaiKey = customEnv?.XAI_API_KEY?.trim();
  const anthropicKey = customEnv?.ANTHROPIC_API_KEY?.trim();

  if (!explicitProvider) {
    if (openaiKey) {
      explicitProvider = 'openai';
    } else if (xaiKey) {
      explicitProvider = 'xai';
    } else if (anthropicKey) {
      explicitProvider = 'anthropic';
    } else {
      explicitProvider = 'mock';
    }
  }

  const model =
    customEnv?.AGENT_MODEL ||
    env.VITE_AGENT_MODEL ||
    (explicitProvider === 'openai'
      ? 'gpt-4o'
      : explicitProvider === 'xai'
        ? 'grok-4.7'
        : explicitProvider === 'anthropic'
          ? 'claude-3-5-sonnet-20241022'
          : 'outpipe-heuristic-engine-v1');

  return {
    configured: true,
    provider: explicitProvider,
    model,
    apiKey: openaiKey || xaiKey || anthropicKey,
    timeoutMs: 60_000,
    maxSteps: 6,
    maxToolCalls: 8,
    maxOutputTokens: 1_536,
    maxTotalTokens: 24_000,
  };
}
