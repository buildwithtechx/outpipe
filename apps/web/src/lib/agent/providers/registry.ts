import { type AgentConfig, readAgentConfig } from '../config';
import { AnthropicAgentProvider } from './anthropic-provider';
import { MockAgentProvider } from './mock-provider';
import { OpenAIAgentProvider } from './openai-provider';
import type { AgentProvider } from './types';
import { XaiAgentProvider } from './xai-provider';

export function getAgentProvider(customConfig?: AgentConfig): AgentProvider {
  const config = customConfig || readAgentConfig();

  switch (config.provider) {
    case 'openai':
      if (config.apiKey) {
        return new OpenAIAgentProvider(config.apiKey, config.model);
      }
      return new MockAgentProvider(config.model);

    case 'xai':
      if (config.apiKey) {
        return new XaiAgentProvider(config.apiKey, config.model);
      }
      return new MockAgentProvider(config.model);

    case 'anthropic':
      if (config.apiKey) {
        return new AnthropicAgentProvider(config.apiKey, config.model);
      }
      return new MockAgentProvider(config.model);

    default:
      return new MockAgentProvider(config.model);
  }
}
