import { useMutation } from '@tanstack/react-query';
import { checkoutBilling } from '../services/billing-service';

export function useBillingCheckout(organizationId: string | undefined) {
  return useMutation({
    mutationFn: (input: {
      planKey: string;
      billingInterval: 'month' | 'year';
    }) =>
      checkoutBilling(
        organizationId ?? '',
        input.planKey,
        input.billingInterval,
      ),
    onSuccess: ({ url }) => window.location.assign(url),
  });
}
