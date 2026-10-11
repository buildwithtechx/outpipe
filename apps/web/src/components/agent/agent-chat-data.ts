import type { RequestExplanationPreview } from '#/components/observability/request-explanation-data';
import type {
  AgentSavedMessage,
  AgentSavedThread,
  AgentStreamEvent,
} from '#/lib/agent/protocol';

export interface AgentRequestContext {
  kind: 'request';
  label: string;
  sourceHref: string;
  preview?: RequestExplanationPreview;
  requestId: string;
  method?: string;
  path?: string;
  statusCode?: number;
  durationMs?: number;
}

export type AgentMessage = AgentSavedMessage & { localStreaming?: boolean };

export interface AgentThread extends AgentSavedThread {
  context?: AgentRequestContext;
  messages: AgentMessage[];
  draft: string;
}

export interface AgentChatState {
  threads: AgentThread[];
  activeThreadId: string | null;
  panelOpen: boolean;
  historyOpen: boolean;
  provider: string;
  model: string;
  availableProviders: string[];
}

export type AgentChatAction =
  | { type: 'open' }
  | { type: 'close' }
  | { type: 'toggle-history' }
  | { type: 'create'; thread: AgentThread }
  | { type: 'select'; threadId: string }
  | { type: 'draft'; threadId: string; value: string }
  | {
      type: 'send';
      threadId: string;
      messageId: string;
      text: string;
      assistantId: string;
    }
  | {
      type: 'stream';
      threadId: string;
      assistantId: string;
      event: AgentStreamEvent;
    }
  | {
      type: 'failed';
      threadId: string;
      assistantId: string;
      error: string;
    }
  | { type: 'load'; threads: AgentThread[] }
  | { type: 'delete'; threadId: string }
  | {
      type: 'set-provider';
      provider: string;
      model?: string;
      availableProviders?: string[];
    };

export const initialAgentChatState: AgentChatState = {
  threads: [],
  activeThreadId: null,
  panelOpen: false,
  historyOpen: false,
  provider: 'default',
  model: '',
  availableProviders: [],
};

export function agentChatReducer(
  state: AgentChatState,
  action: AgentChatAction,
): AgentChatState {
  switch (action.type) {
    case 'open':
      return { ...state, panelOpen: true };

    case 'close':
      return { ...state, panelOpen: false, historyOpen: false };

    case 'toggle-history':
      return { ...state, historyOpen: !state.historyOpen };

    case 'create': {
      const existing = state.threads.find((t) => t.id === action.thread.id);
      if (existing) {
        return {
          ...state,
          activeThreadId: action.thread.id,
          panelOpen: true,
        };
      }
      return {
        ...state,
        threads: [action.thread, ...state.threads],
        activeThreadId: action.thread.id,
        panelOpen: true,
      };
    }

    case 'select':
      return {
        ...state,
        activeThreadId: action.threadId,
        historyOpen: false,
      };

    case 'draft':
      return {
        ...state,
        threads: state.threads.map((t) =>
          t.id === action.threadId ? { ...t, draft: action.value } : t,
        ),
      };

    case 'send': {
      const userMsg: AgentMessage = {
        id: action.messageId,
        role: 'user',
        text: action.text,
        status: 'complete',
        steps: [],
        evidence: [],
      };
      const assistantMsg: AgentMessage = {
        id: action.assistantId,
        role: 'assistant',
        text: '',
        status: 'running',
        steps: [],
        evidence: [],
        localStreaming: true,
      };
      return {
        ...state,
        threads: state.threads.map((t) =>
          t.id === action.threadId
            ? {
                ...t,
                draft: '',
                updatedAt: Date.now(),
                messages: [...t.messages, userMsg, assistantMsg],
              }
            : t,
        ),
      };
    }

    case 'stream': {
      const event = action.event;
      return {
        ...state,
        threads: state.threads.map((t) => {
          if (t.id !== action.threadId) return t;
          return {
            ...t,
            updatedAt: Date.now(),
            messages: t.messages.map((m) => {
              if (m.id !== action.assistantId) return m;
              if (event.type === 'step') {
                const step = event.step;
                const existingIdx = m.steps.findIndex((s) => s.id === step.id);
                let newSteps = [...m.steps];
                if (existingIdx >= 0) {
                  newSteps[existingIdx] = step;
                } else {
                  newSteps = [...newSteps, step];
                }
                return { ...m, steps: newSteps };
              }
              if (event.type === 'evidence') {
                return { ...m, evidence: event.evidence };
              }
              if (event.type === 'text') {
                return { ...m, text: m.text + event.delta };
              }
              if (event.type === 'finish') {
                return {
                  ...m,
                  status: event.status,
                  error: event.error,
                  localStreaming: false,
                };
              }
              return m;
            }),
          };
        }),
      };
    }

    case 'failed': {
      return {
        ...state,
        threads: state.threads.map((t) => {
          if (t.id !== action.threadId) return t;
          return {
            ...t,
            messages: t.messages.map((m) =>
              m.id === action.assistantId
                ? {
                    ...m,
                    status: 'failed',
                    error: action.error,
                    localStreaming: false,
                  }
                : m,
            ),
          };
        }),
      };
    }

    case 'load':
      return { ...state, threads: action.threads };

    case 'delete': {
      const remaining = state.threads.filter((t) => t.id !== action.threadId);
      const nextActive =
        state.activeThreadId === action.threadId
          ? remaining[0]?.id || null
          : state.activeThreadId;
      return {
        ...state,
        threads: remaining,
        activeThreadId: nextActive,
      };
    }

    case 'set-provider':
      return {
        ...state,
        provider: action.provider,
        model: action.model !== undefined ? action.model : state.model,
        availableProviders:
          action.availableProviders !== undefined
            ? action.availableProviders
            : state.availableProviders,
      };

    default:
      return state;
  }
}
