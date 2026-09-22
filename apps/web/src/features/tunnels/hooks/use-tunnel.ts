import { useQuery } from '@tanstack/react-query';
import { getTunnel } from '../services/tunnel-service';

export function useTunnel(tunnelID: string) {
  return useQuery({
    queryKey: ['tunnel', tunnelID],
    queryFn: () => getTunnel(tunnelID),
    enabled: Boolean(tunnelID),
  });
}
