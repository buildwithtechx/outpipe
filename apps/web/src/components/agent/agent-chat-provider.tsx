import {
  type ReactNode,
  useCallback,
  useEffect,
  useMemo,
  useReducer,
  useRef,
} from 'react';
import { readAgentConfig } from '#/lib/agent/config';
import { runAgentInvestigation } from '#/lib/agent/runtime';
import {
  deleteStoredThread,
  listStoredThreads,
  saveStoredThread,
} from '#/lib/agent/store';
import { getApiBaseURL } from '#/lib/api-client';
import {
  AgentActionsContext,
  type AgentChatActions,
  type AgentChatStateValue,
  AgentStateContext,
} from './agent-chat-context';
import {
  type AgentRequestContext,
  type AgentThread,
  agentChatReducer,
  initialAgentChatState,
} from './agent-chat-data';
import { streamAgentChatServer } from './agent-chat-transport';

export interface AgentChatProviderProps {
  organizationId: string;
  children: ReactNode;
}

export function AgentChatProvider({
  organizationId,
  children,
}: AgentChatProviderProps) {
  const [state, dispatch] = useReducer(agentChatReducer, initialAgentChatState);
  const activeControllers = useRef<Map<string, AbortController>>(new Map());

  // Load configuration and existing saved threads on mount
  useEffect(() => {
    const config = readAgentConfig();
    dispatch({
      type: 'set-provider',
      provider: config.provider,
      model: config.model,
    });

    const stored = listStoredThreads(organizationId);
    if (stored.length > 0) {
      const threads: AgentThread[] = stored.map((t) => ({
        ...t,
        messages: t.messages.map((m) => ({ ...m, localStreaming: false })),
        draft: '',
      }));
      dispatch({ type: 'load', threads });
    }
  }, [organizationId]);

  // Persist updated threads to local storage
  const activeThread = state.threads.find((t) => t.id === state.activeThreadId);
  useEffect(() => {
    if (activeThread && activeThread.messages.length > 0) {
      saveStoredThread(organizationId, activeThread);
    }
  }, [organizationId, activeThread]);

  const openAgent = useCallback(() => {
    dispatch({ type: 'open' });
  }, []);

  const closeAgent = useCallback(() => {
    dispatch({ type: 'close' });
  }, []);

  const startThread = useCallback((context?: AgentRequestContext) => {
    const threadId = crypto.randomUUID();
    const newThread: AgentThread = {
      id: threadId,
      title: context ? `Explain ${context.label}` : 'New Investigation',
      sourceRequestId: context?.requestId || null,
      context,
      createdAt: Date.now(),
      updatedAt: Date.now(),
      messages: [],
      draft: '',
    };
    dispatch({ type: 'create', thread: newThread });
  }, []);

  const stopInvestigation = useCallback((threadId: string) => {
    const controller = activeControllers.current.get(threadId);
    if (controller) {
      controller.abort();
      activeControllers.current.delete(threadId);
    }
  }, []);

  const sendMessage = useCallback(
    async (textToSend?: string) => {
      const currentThread =
        state.threads.find((t) => t.id === state.activeThreadId) ||
        state.threads[0];
      if (!currentThread) {
        startThread();
        return;
      }

      const prompt = (textToSend || currentThread.draft).trim();
      if (!prompt) return;

      const messageId = crypto.randomUUID();
      const assistantId = crypto.randomUUID();
      const controller = new AbortController();
      activeControllers.current.set(currentThread.id, controller);

      dispatch({
        type: 'send',
        threadId: currentThread.id,
        messageId,
        text: prompt,
        assistantId,
      });

      try {
        let serverHandled = false;
        try {
          serverHandled = await streamAgentChatServer({
            organizationId,
            threadId: currentThread.id,
            message: prompt,
            sourceRequestId: currentThread.sourceRequestId,
            signal: controller.signal,
            onEvent: (event) => {
              dispatch({
                type: 'stream',
                threadId: currentThread.id,
                assistantId,
                event,
              });
            },
          });
        } catch (serverErr) {
          if (controller.signal.aborted) throw serverErr;
          serverHandled = false;
        }

        if (!serverHandled && !controller.signal.aborted) {
          await runAgentInvestigation({
            prompt,
            sourceRequestId: currentThread.sourceRequestId,
            history: currentThread.messages,
            signal: controller.signal,
            emit: (event) => {
              dispatch({
                type: 'stream',
                threadId: currentThread.id,
                assistantId,
                event,
              });
            },
          });
        }
      } catch (err) {
        if (!controller.signal.aborted) {
          const msg =
            err instanceof Error ? err.message : 'Investigation failed';
          dispatch({
            type: 'failed',
            threadId: currentThread.id,
            assistantId,
            error: msg,
          });
        }
      } finally {
        activeControllers.current.delete(currentThread.id);
      }
    },
    [state.threads, state.activeThreadId, startThread, organizationId],
  );

  const handleDeleteThread = useCallback(
    (threadId: string) => {
      stopInvestigation(threadId);
      deleteStoredThread(organizationId, threadId);
      const baseURL = getApiBaseURL();
      fetch(
        `${baseURL}/api/v1/organizations/${organizationId}/agent/threads/${threadId}`,
        {
          method: 'DELETE',
          credentials: 'include',
        },
      ).catch(() => undefined);
      dispatch({ type: 'delete', threadId });
    },
    [organizationId, stopInvestigation],
  );

  const actionsValue: AgentChatActions = useMemo(
    () => ({
      openAgent,
      closeAgent,
      startThread,
      isOpen: state.panelOpen,
    }),
    [openAgent, closeAgent, startThread, state.panelOpen],
  );

  const stateValue: AgentChatStateValue = useMemo(
    () => ({
      state,
      dispatch: (action) => {
        if (action.type === 'delete') {
          handleDeleteThread(action.threadId);
        } else {
          dispatch(action);
        }
      },
      sendMessage,
      stopInvestigation,
    }),
    [state, sendMessage, stopInvestigation, handleDeleteThread],
  );

  return (
    <AgentActionsContext.Provider value={actionsValue}>
      <AgentStateContext.Provider value={stateValue}>
        {children}
      </AgentStateContext.Provider>
    </AgentActionsContext.Provider>
  );
}
