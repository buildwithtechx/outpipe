import { Button } from '#/components/ui/button';
import type { Plan } from '#/interfaces/billing';
import { useBillingCheckout } from '../hooks/use-billing-checkout';

export function BillingPlanPicker({
  organizationId,
  plans,
}: {
  organizationId: string;
  plans: Plan[];
}) {
  const checkout = useBillingCheckout(organizationId);
  return (
    <section className="mt-8 rounded-2xl border border-white/10 bg-white/[0.025] p-6">
      <h2 className="text-lg font-medium">Change plan</h2>
      <div className="mt-4 grid gap-3 sm:grid-cols-2">
        {plans
          .filter((plan) => plan.key !== 'free')
          .map((plan) => (
            <div
              key={plan.id}
              className="flex items-center justify-between gap-4 rounded-xl border border-white/10 p-4"
            >
              <div>
                <p className="font-medium">{plan.name}</p>
                <p className="mt-1 text-sm text-white/45">
                  {formatPrice(plan.priceMinor, plan.currency)} /{' '}
                  {plan.billingInterval}
                </p>
              </div>
              <Button
                type="button"
                variant="outline"
                disabled={checkout.isPending}
                onClick={() =>
                  checkout.mutate({
                    planKey: plan.key,
                    billingInterval:
                      plan.billingInterval === 'year' ? 'year' : 'month',
                  })
                }
              >
                Choose
              </Button>
            </div>
          ))}
      </div>
      {checkout.isError && (
        <p className="mt-3 text-sm text-rose-200">
          Checkout could not be started.
        </p>
      )}
    </section>
  );
}

function formatPrice(minor: number, currency: string) {
  return new Intl.NumberFormat(undefined, {
    style: 'currency',
    currency,
  }).format(minor / 100);
}
