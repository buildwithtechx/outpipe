import { useEffect, useState } from 'react';

export function useAuthNotice() {
  const [notice, setNotice] = useState<string | null>(null);

  useEffect(() => {
    if (typeof window === 'undefined') return;
    const search = new URLSearchParams(window.location.search);
    const error = search.get('error');
    if (error === 'oauth_failed') {
      setNotice('We could not complete that sign-in. Please try again.');
    }
    if (error === 'oauth_start_failed') {
      setNotice(
        'That sign-in provider is unavailable right now. Please try again later.',
      );
    }
    if (error) {
      search.delete('error');
      const qs = search.toString();
      window.history.replaceState(
        null,
        '',
        qs ? `${window.location.pathname}?${qs}` : window.location.pathname,
      );
    }
  }, []);

  return notice;
}
