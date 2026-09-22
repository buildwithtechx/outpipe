import { useQuery } from '@tanstack/react-query';
import { useEffect } from 'react';
import { getAuthSession } from '#/features/auth/services/auth-service';
import { useAuthStore } from '#/stores/auth-store';

export function useAuthSession() {
  const setUser = useAuthStore((state) => state.setUser);
  const clear = useAuthStore((state) => state.clear);

  const query = useQuery({
    queryKey: ['auth', 'session'],
    queryFn: getAuthSession,
    enabled: typeof window !== 'undefined',
    retry: false,
    refetchOnMount: 'always',
  });

  useEffect(() => {
    if (query.isError) {
      clear();
    } else if (query.data) {
      setUser(query.data.user);
    }
  }, [clear, query.data, query.isError, setUser]);

  const isAuthenticated = !query.isError && query.data !== undefined;
  const user = !query.isError ? (query.data?.user ?? null) : null;

  return {
    ...query,
    isAuthenticated,
    user,
  };
}
