import { useQuery } from '@tanstack/react-query';
import { getInvoices } from '../services/billing-service';

export function useBillingInvoices(organizationId?: string) {
  return useQuery({
    queryKey: ['billing-invoices', organizationId],
    queryFn: () => getInvoices(organizationId as string),
    enabled: Boolean(organizationId),
  });
}
