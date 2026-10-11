import { createContext, type Dispatch, useContext } from 'react';
import type {
  AgentChatAction,
  AgentChatState,
  AgentRequestContext,
} from './agent-chat-data';

export interface AgentChatActions {
  openAgent: () => void;
  closeAgent: () => void;
  startThread: (context?: AgentRequestContext) => void;
  isOpen: boolean;
}

export interface AgentChatStateValue {
  state: AgentChatState;
  dispatch: Dispatch<AgentChatAction>;
  sendMessage: (text?: string) => void;
  stopInvestigation: (threadId: string) => void;
}

export const AgentActionsContext = createContext<AgentChatActions | null>(null);
export const AgentStateContext = createContext<AgentChatStateValue | null>(
  null,
);

export function useAgentChat(): AgentChatActions {
  const value = useContext(AgentActionsContext);
  if (!value) {
    throw new Error('useAgentChat must be used within an AgentChatProvider');
  }
  return value;
}

export function useAgentChatState(): AgentChatStateValue {
  const value = useContext(AgentStateContext);
  if (!value) {
    throw new Error(
      'useAgentChatState must be used within an AgentChatProvider',
    );
  }
  return value;
}
