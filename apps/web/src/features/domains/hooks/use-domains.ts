import { useQuery } from '@tanstack/react-query';
import { getDomains } from '../services/domain-service';

export function useDomains(organizationId?: string) {
  return useQuery({
    queryKey: ['domains', organizationId],
    queryFn: () => getDomains(organizationId as string),
    enabled: Boolean(organizationId),
  });
}
