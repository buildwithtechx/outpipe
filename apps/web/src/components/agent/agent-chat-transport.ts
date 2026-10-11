import type { AgentStreamEvent } from '#/lib/agent/protocol';
import { getApiBaseURL } from '#/lib/api-client';

export function createAgentSseParser(onData: (data: unknown) => void) {
  let buffer = '';
  const frame = (raw: string) => {
    const data = raw
      .split(/\r?\n/)
      .filter((line) => line.startsWith('data:'))
      .map((line) => line.slice(5).replace(/^ /, ''))
      .join('\n');
    if (!data) return;
    try {
      const parsed = JSON.parse(data);
      onData(parsed);
    } catch {
      throw new Error('Agent returned an invalid response stream.');
    }
  };

  return {
    push(chunk: string) {
      buffer += chunk;
      if (buffer.length > 1_000_000) {
        throw new Error('Agent response stream exceeded limit.');
      }
      while (true) {
        const separator = /\r?\n\r?\n/.exec(buffer);
        if (!separator) break;
        frame(buffer.slice(0, separator.index));
        buffer = buffer.slice(separator.index + separator[0].length);
      }
    },
    finish() {
      if (buffer.trim()) {
        frame(buffer);
      }
    },
  };
}

export async function consumeAgentStream(
  response: Response,
  onEvent: (event: AgentStreamEvent) => void,
) {
  if (!response.body) {
    throw new Error('Agent returned an empty response body.');
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let finished = false;

  const parser = createAgentSseParser((data) => {
    if (finished) return;
    if (typeof data === 'object' && data !== null) {
      const evt = data as AgentStreamEvent;
      onEvent(evt);
      if (evt.type === 'finish') {
        finished = true;
      }
    }
  });

  try {
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      parser.push(decoder.decode(value, { stream: true }));
    }
    parser.push(decoder.decode());
    parser.finish();
  } finally {
    reader.releaseLock();
  }
}

export async function streamAgentChatServer({
  organizationId,
  threadId,
  message,
  sourceRequestId,
  signal,
  onEvent,
}: {
  organizationId: string;
  threadId: string;
  message: string;
  sourceRequestId?: string | null;
  signal?: AbortSignal;
  onEvent: (event: AgentStreamEvent) => void;
}): Promise<boolean> {
  const baseURL = getApiBaseURL();
  const url = `${baseURL}/api/v1/organizations/${organizationId}/agent/chat`;

  const res = await fetch(url, {
    method: 'POST',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      Accept: 'text/event-stream',
    },
    body: JSON.stringify({
      threadId,
      message,
      sourceRequestId: sourceRequestId || undefined,
    }),
    signal,
  });

  if (!res.ok) {
    if (res.status === 404 || res.status === 501 || res.status === 502) {
      return false;
    }
    let errorMsg = `Server error (${res.status})`;
    try {
      const body = await res.json();
      if (body?.error) errorMsg = body.error;
    } catch {
      // ignore
    }
    throw new Error(errorMsg);
  }

  await consumeAgentStream(res, onEvent);
  return true;
}
