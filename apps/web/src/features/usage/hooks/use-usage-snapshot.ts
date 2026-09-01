import { useQuery } from '@tanstack/react-query';
import { getUsageSnapshot } from '../services/usage-service';

export function useUsageSnapshot(organizationId?: string) {
  return useQuery({
    queryKey: ['usage', organizationId],
    queryFn: () => getUsageSnapshot(organizationId as string),
    enabled: Boolean(organizationId),
  });
}
