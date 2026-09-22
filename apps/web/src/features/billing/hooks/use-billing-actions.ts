import { useMutation, useQueryClient } from '@tanstack/react-query';
import {
  cancelBilling,
  getBillingPortal,
  resumeBilling,
} from '../services/billing-service';

export function useBillingActions(organizationId: string | undefined) {
  const queryClient = useQueryClient();

  const invalidate = () => {
    setTimeout(() => {
      void queryClient.invalidateQueries({
        queryKey: ['billing', organizationId],
      });
    }, 2000);
    return queryClient.invalidateQueries({
      queryKey: ['billing', organizationId],
    });
  };

  const portal = useMutation({
    mutationFn: () => getBillingPortal(organizationId ?? ''),
    onSuccess: ({ url }) => {
      if (url) {
        window.location.assign(url);
      }
    },
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
