import { describe, expect, it } from 'vitest';
import {
  type AgentThread,
  agentChatReducer,
  initialAgentChatState,
} from './agent-chat-data';
import { createAgentSseParser } from './agent-chat-transport';

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

  it('updates provider, model and availableProviders from backend config', () => {
    let state = agentChatReducer(initialAgentChatState, {
      type: 'set-provider',
      provider: 'openai',
      model: 'gpt-4o-mini',
      availableProviders: ['mock', 'openai', 'anthropic'],
    });

    expect(state.provider).toBe('openai');
    expect(state.model).toBe('gpt-4o-mini');
    expect(state.availableProviders).toEqual(['mock', 'openai', 'anthropic']);

    state = agentChatReducer(state, {
      type: 'set-provider',
      provider: 'anthropic',
    });
    expect(state.provider).toBe('anthropic');
    expect(state.model).toBe('gpt-4o-mini');
    expect(state.availableProviders).toEqual(['mock', 'openai', 'anthropic']);
  });

  it('parses SSE chunks correctly in createAgentSseParser', () => {
    const events: unknown[] = [];
    const parser = createAgentSseParser((data) => events.push(data));

    parser.push('data: {"type":"text","delta":"Hello "}\n\n');
    parser.push('data: {"type":"text","delta":"world"}\n\n');
    parser.push('data: {"type":"finish","status":"complete"}\n\n');
    parser.finish();

    expect(events.length).toBe(3);
    expect(events[0]).toEqual({ type: 'text', delta: 'Hello ' });
    expect(events[1]).toEqual({ type: 'text', delta: 'world' });
    expect(events[2]).toEqual({ type: 'finish', status: 'complete' });
  });
});
