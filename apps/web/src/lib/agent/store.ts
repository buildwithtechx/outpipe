import type { AgentSavedThread } from './protocol';

const STORAGE_PREFIX = 'outpipe:agent:threads:';

function getStorageKey(organizationId: string) {
  return `${STORAGE_PREFIX}${organizationId}`;
}

export function listStoredThreads(organizationId: string): AgentSavedThread[] {
  if (typeof window === 'undefined' || !window.localStorage) return [];
  try {
    const raw = localStorage.getItem(getStorageKey(organizationId));
    if (!raw) return [];
    const threads = JSON.parse(raw);
    return Array.isArray(threads) ? threads : [];
  } catch {
    return [];
  }
}

export function getStoredThread(
  organizationId: string,
  threadId: string,
): AgentSavedThread | null {
  const threads = listStoredThreads(organizationId);
  return threads.find((t) => t.id === threadId) || null;
}

export function saveStoredThread(
  organizationId: string,
  thread: AgentSavedThread,
): void {
  if (typeof window === 'undefined' || !window.localStorage) return;
  try {
    const threads = listStoredThreads(organizationId);
    const index = threads.findIndex((t) => t.id === thread.id);
    let updated: AgentSavedThread[];
    if (index >= 0) {
      updated = [...threads];
      updated[index] = thread;
    } else {
      updated = [thread, ...threads].slice(0, 50);
    }
    localStorage.setItem(
      getStorageKey(organizationId),
      JSON.stringify(updated),
    );
  } catch {
    // Storage quota or disabled localStorage
  }
}

export function deleteStoredThread(
  organizationId: string,
  threadId: string,
): void {
  if (typeof window === 'undefined' || !window.localStorage) return;
  try {
    const threads = listStoredThreads(organizationId);
    const filtered = threads.filter((t) => t.id !== threadId);
    localStorage.setItem(
      getStorageKey(organizationId),
      JSON.stringify(filtered),
    );
  } catch {
    // Storage error
  }
}
