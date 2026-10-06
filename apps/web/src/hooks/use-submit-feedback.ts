import { useEffect, useState } from 'react';

export function useSubmitFeedback(open: boolean) {
  const [error, setError] = useState(false);
  useEffect(() => {
    if (!open) setError(false);
  }, [open]);
  async function submit(action: () => Promise<unknown>) {
    setError(false);
    try {
      await action();
      return true;
    } catch {
      setError(true);
      return false;
    }
  }
  return { error, submit };
}
