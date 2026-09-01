import { useQuery } from '@tanstack/react-query';
import { getRequestEvents } from '../services/requests-service';

export function useRequestEvents(organizationId?: string) {
  return useQuery({
    queryKey: ['requests', organizationId],
    queryFn: () => getRequestEvents(organizationId as string),
    enabled: Boolean(organizationId),
  });
}
