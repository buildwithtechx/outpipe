import { useBlocker } from '@tanstack/react-router';
import { useCallback } from 'react';

export function useUptimeUnsavedChanges(dirty: boolean) {
  const shouldBlockFn = useCallback(() => dirty, [dirty]);
  return useBlocker({
    shouldBlockFn,
    enableBeforeUnload: dirty,
    withResolver: true,
  });
}
