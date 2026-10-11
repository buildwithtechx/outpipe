import { describe, expect, it } from 'vitest';
import { readAgentConfig } from '#/lib/agent/config';
import type { AgentStreamEvent } from '#/lib/agent/protocol';
import { MockAgentProvider } from '#/lib/agent/providers/mock-provider';
import {
  type AgentThread,
  agentChatReducer,
  initialAgentChatState,
} from './agent-chat-data';

describe('Agent Chat Reducer & Provider', () => {
  it('handles open and close actions', () => {
    let state = agentChatReducer(initialAgentChatState, { type: 'open' });
    expect(state.panelOpen).toBe(true);

    state = agentChatReducer(state, { type: 'close' });
    expect(state.panelOpen).toBe(false);
  });

  it('creates and selects a new conversation thread', () => {
    const thread: AgentThread = {
      id: 'thread-1',
      title: 'Explain Request #123',
      sourceRequestId: 'req-123',
      createdAt: Date.now(),
      updatedAt: Date.now(),
      messages: [],
      draft: '',
    };

    const state = agentChatReducer(initialAgentChatState, {
      type: 'create',
      thread,
    });

    expect(state.threads.length).toBe(1);
    expect(state.activeThreadId).toBe('thread-1');
    expect(state.panelOpen).toBe(true);
  });

  it('appends user and assistant messages on send and streams text', () => {
    const thread: AgentThread = {
      id: 'thread-1',
      title: 'Investigation',
      sourceRequestId: null,
      createdAt: Date.now(),
      updatedAt: Date.now(),
      messages: [],
      draft: 'Why did the request fail?',
    };

    let state = agentChatReducer(initialAgentChatState, {
      type: 'create',
      thread,
    });

    state = agentChatReducer(state, {
      type: 'send',
      threadId: 'thread-1',
      messageId: 'user-msg-1',
      text: 'Why did the request fail?',
      assistantId: 'asst-msg-1',
    });

    const active = state.threads[0];
    expect(active.messages.length).toBe(2);
    expect(active.messages[0].role).toBe('user');
    expect(active.messages[1].role).toBe('assistant');
    expect(active.messages[1].status).toBe('running');

    // Stream text delta
    state = agentChatReducer(state, {
      type: 'stream',
      threadId: 'thread-1',
      assistantId: 'asst-msg-1',
      event: { type: 'text', delta: 'The payment upstream timed out.' },
    });

    expect(state.threads[0].messages[1].text).toBe(
      'The payment upstream timed out.',
    );

    // Finish event
    state = agentChatReducer(state, {
      type: 'stream',
      threadId: 'thread-1',
      assistantId: 'asst-msg-1',
      event: { type: 'finish', status: 'complete' },
    });

    expect(state.threads[0].messages[1].status).toBe('complete');
  });

  it('runs MockAgentProvider and emits steps, evidence, and finish', async () => {
    const provider = new MockAgentProvider();
    const emittedEvents: AgentStreamEvent[] = [];
    const controller = new AbortController();

    const result = await provider.investigate({
      prompt: 'Check tunnel errors',
      sourceRequestId: 'req-test-99',
      history: [],
      signal: controller.signal,
      emit: (e) => {
        emittedEvents.push(e);
      },
    });

    expect(result.text).toContain('HTTP 503');
    expect(result.steps.length).toBeGreaterThanOrEqual(3);
    expect(result.evidence.length).toBeGreaterThanOrEqual(2);
    expect(emittedEvents.some((e) => e.type === 'step')).toBe(true);
    expect(emittedEvents.some((e) => e.type === 'evidence')).toBe(true);
    expect(emittedEvents.some((e) => e.type === 'text')).toBe(true);
  });

  it('reads pluggable configuration defaults correctly', () => {
    const config = readAgentConfig({});
    expect(config.provider).toBe('mock');
    expect(config.configured).toBe(true);

    const openaiConfig = readAgentConfig({
      OPENAI_API_KEY: 'sk-test-123',
    });
    expect(openaiConfig.provider).toBe('openai');
    expect(openaiConfig.configured).toBe(true);

    const xaiConfig = readAgentConfig({
      DEFAULT_AGENT_PROVIDER: 'xai',
      XAI_API_KEY: 'xai-test-key',
    });
    expect(xaiConfig.provider).toBe('xai');
    expect(xaiConfig.model).toBe('grok-4.7');
  });
});
