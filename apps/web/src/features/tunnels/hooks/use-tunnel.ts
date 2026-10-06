import { useQuery } from '@tanstack/react-query';
import { ApiError } from '#/lib/api-client';
import { getTunnel } from '../services/tunnel-service';

export function useTunnel(tunnelID: string) {
  return useQuery({
    queryKey: ['tunnel', tunnelID],
    queryFn: () => getTunnel(tunnelID),
    enabled: Boolean(tunnelID),
    retry: (count, error) =>
      !(
        error instanceof ApiError &&
        error.status >= 400 &&
        error.status < 500
      ) && count < 3,
    refetchInterval: (query) =>
      query.state.error instanceof ApiError &&
      query.state.error.status >= 400 &&
      query.state.error.status < 500
        ? false
        : 5000,
    refetchIntervalInBackground: false,
  });
}
