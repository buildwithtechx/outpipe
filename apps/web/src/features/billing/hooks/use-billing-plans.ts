import { useQuery } from '@tanstack/react-query';
import { getBillingPlans } from '../services/billing-service';

export function useBillingPlans(organizationId: string | undefined) {
  return useQuery({
    queryKey: ['billing', 'plans', organizationId],
    queryFn: () => getBillingPlans(organizationId ?? ''),
    enabled: Boolean(organizationId),
  });
}
