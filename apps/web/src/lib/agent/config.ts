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
  env: Record<string, string | undefined> = typeof process !== 'undefined'
    ? process.env
    : {},
): AgentConfig {
  const explicitProvider = (
    env.DEFAULT_AGENT_PROVIDER ||
    env.AGENT_PROVIDER ||
    ''
  )
    .toLowerCase()
    .trim() as AgentProviderType;

  const openaiKey = env.OPENAI_API_KEY?.trim();
  const xaiKey = env.XAI_API_KEY?.trim();
  const anthropicKey = env.ANTHROPIC_API_KEY?.trim();

  let provider: AgentProviderType = 'mock';
  let apiKey: string | undefined;
  let model = env.AGENT_MODEL?.trim() || '';

  if (explicitProvider === 'openai' || (!explicitProvider && openaiKey)) {
    provider = 'openai';
    apiKey = openaiKey;
    model = model || env.OPENAI_MODEL?.trim() || 'gpt-4o';
  } else if (explicitProvider === 'xai' || (!explicitProvider && xaiKey)) {
    provider = 'xai';
    apiKey = xaiKey;
    model =
      model ||
      env.AGENT_GROK_MODEL?.trim() ||
      env.XAI_MODEL?.trim() ||
      'grok-4.7';
  } else if (
    explicitProvider === 'anthropic' ||
    (!explicitProvider && anthropicKey)
  ) {
    provider = 'anthropic';
    apiKey = anthropicKey;
    model =
      model || env.ANTHROPIC_MODEL?.trim() || 'claude-3-5-sonnet-20241022';
  } else {
    provider = 'mock';
    model = model || 'outpipe-heuristic-engine-v1';
  }

  const isConfigured =
    provider === 'mock' || (Boolean(apiKey) && Boolean(model));

  return {
    configured: isConfigured,
    provider,
    model,
    apiKey,
    timeoutMs: 60_000,
    maxSteps: 6,
    maxToolCalls: 8,
    maxOutputTokens: 1_536,
    maxTotalTokens: 24_000,
  };
}
