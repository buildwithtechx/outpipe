import { useQuery } from '@tanstack/react-query';
import { getBilling } from '../services/billing-service';

export function useBilling(organizationId?: string) {
  return useQuery({
    queryKey: ['billing', organizationId],
    queryFn: () => getBilling(organizationId as string),
    enabled: Boolean(organizationId),
  });
}
