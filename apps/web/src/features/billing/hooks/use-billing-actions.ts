import { useMutation, useQueryClient } from '@tanstack/react-query';
import {
  cancelBilling,
  getBillingPortal,
  resumeBilling,
} from '../services/billing-service';

export function useBillingActions(organizationId: string | undefined) {
  const queryClient = useQueryClient();

  const invalidate = () =>
    queryClient.invalidateQueries({ queryKey: ['billing', organizationId] });

  const portal = useMutation({
    mutationFn: () => getBillingPortal(organizationId ?? ''),
    onSuccess: ({ url }) => window.open(url, '_blank', 'noopener,noreferrer'),
  });

  const cancel = useMutation({
    mutationFn: () => cancelBilling(organizationId ?? ''),
    onSuccess: invalidate,
  });

  const resume = useMutation({
    mutationFn: () => resumeBilling(organizationId ?? ''),
    onSuccess: invalidate,
  });

  return { portal, cancel, resume };
}
