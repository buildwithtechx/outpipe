import { describe, expect, it } from 'vitest';
import type { AgentStreamEvent } from '#/lib/agent/protocol';
import {
  consumeAgentStream,
  createAgentSseParser,
} from './agent-chat-transport';

describe('agent-chat-transport', () => {
  it('parses valid SSE events across chunk boundaries', () => {
    const received: unknown[] = [];
    const parser = createAgentSseParser((data) => received.push(data));

    parser.push('data: {"type":"run","runId":"run-1"}\n\n');
    expect(received).toHaveLength(1);
    expect(received[0]).toEqual({ type: 'run', runId: 'run-1' });

    // Multi-chunk splitting
    parser.push('data: {"type":"text"');
    parser.push(',"delta":"Hello ');
    parser.push('world"}\n\n');

    expect(received).toHaveLength(2);
    expect(received[1]).toEqual({ type: 'text', delta: 'Hello world' });

    parser.finish();
  });

  it('handles CRLF line terminators and ignores non-data comments', () => {
    const received: unknown[] = [];
    const parser = createAgentSseParser((data) => received.push(data));

    parser.push(
      ': ping\r\ndata: {"type":"step","step":{"id":"1","label":"Inspect","status":"complete"}}\r\n\r\n',
    );
    expect(received).toHaveLength(1);
    expect((received[0] as { type: string }).type).toBe('step');

    parser.finish();
  });

  it('throws on invalid JSON in data frames', () => {
    const parser = createAgentSseParser(() => {});
    expect(() => {
      parser.push('data: not-json\n\n');
    }).toThrow(/Agent returned an invalid response stream/);
  });

  it('consumes a mock Response stream into typed events', async () => {
    const events: AgentStreamEvent[] = [];
    const payload =
      'data: {"type":"run","runId":"run-42"}\n\n' +
      'data: {"type":"text","delta":"Investigating..."}\n\n' +
      'data: {"type":"finish","status":"complete"}\n\n';

    const stream = new ReadableStream({
      start(controller) {
        controller.enqueue(new TextEncoder().encode(payload));
        controller.close();
      },
    });

    const mockResponse = new Response(stream, {
      headers: { 'Content-Type': 'text/event-stream' },
    });

    await consumeAgentStream(mockResponse, (event) => {
      events.push(event);
    });

    expect(events).toHaveLength(3);
    expect(events[0]).toEqual({ type: 'run', runId: 'run-42' });
    expect(events[1]).toEqual({ type: 'text', delta: 'Investigating...' });
    expect(events[2]).toEqual({ type: 'finish', status: 'complete' });
  });
});
