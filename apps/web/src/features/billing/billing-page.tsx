import { Button } from '#/components/ui/button';
import { useOrganization } from '#/features/organizations/hooks/use-organization';
import { BillingPlanPicker } from './components/billing-plan-picker';
import { useBilling } from './hooks/use-billing';
import { useBillingActions } from './hooks/use-billing-actions';
import { useBillingInvoices } from './hooks/use-billing-invoices';
import { useBillingPlans } from './hooks/use-billing-plans';

export function BillingPage({ orgSlug }: { orgSlug: string }) {
  const organizationQuery = useOrganization(orgSlug);
  const query = useBilling(organizationQuery.organization?.id);
  const actions = useBillingActions(organizationQuery.organization?.id);
  const invoices = useBillingInvoices(organizationQuery.organization?.id);
  const plans = useBillingPlans(organizationQuery.organization?.id);
  if (organizationQuery.isLoading || query.isLoading)
    return <p className="p-8 text-sm text-white/55">Loading billing…</p>;
  if (
    organizationQuery.isError ||
    query.isError ||
    !organizationQuery.organization
  )
    return (
      <p className="p-8 text-sm text-rose-200">
        We could not load billing details.
      </p>
    );
  const organization = organizationQuery.organization;
  const { plan, subscription } = query.data ?? {};
  if (!plan || !subscription)
    return (
      <p className="p-8 text-sm text-rose-200">
        Billing details are not available for this workspace.
      </p>
    );
  return (
    <main className="mx-auto w-full max-w-6xl px-6 py-12 text-white sm:px-8 lg:py-16">
      <header className="border-b border-white/10 pb-8">
        <p className="mb-3 text-sm text-indigo-200">{organization.name}</p>
        <h1 className="text-3xl font-semibold tracking-tight">Billing</h1>
        <p className="mt-3 text-sm text-white/55">
          Your current plan and subscription status.
        </p>
      </header>
      <section className="mt-8 rounded-2xl border border-indigo-300/20 bg-indigo-300/[0.05] p-6">
        <p className="text-sm text-indigo-200">Current plan</p>
        <div className="mt-3 flex flex-wrap items-center justify-between gap-4">
          <h2 className="text-2xl font-semibold">{plan.name}</h2>
          <span className="rounded-full bg-emerald-400/10 px-3 py-1 text-sm text-emerald-200">
            {subscription.status}
          </span>
        </div>
        <p className="mt-3 text-sm text-white/55">
          {formatPrice(plan.priceMinor, plan.currency)} · {plan.billingInterval}
        </p>
        <div className="mt-5 flex flex-wrap gap-3">
          <Button
            type="button"
            variant="outline"
            onClick={() => actions.portal.mutate()}
            disabled={actions.portal.isPending}
          >
            Manage billing
          </Button>
          {subscription.cancelAtPeriodEnd ? (
            <Button
              type="button"
              variant="outline"
              onClick={() => actions.resume.mutate()}
              disabled={actions.resume.isPending}
            >
              Resume subscription
            </Button>
          ) : (
            <Button
              type="button"
              variant="outline"
              onClick={() => actions.cancel.mutate()}
              disabled={actions.cancel.isPending}
            >
              Cancel at period end
            </Button>
          )}
        </div>
      </section>
      {!plans.isLoading && plans.data?.plans && (
        <BillingPlanPicker
          organizationId={organization.id}
          plans={plans.data.plans}
        />
      )}
      <section className="mt-8 rounded-2xl border border-white/10 bg-white/[0.025] p-6">
        <h2 className="text-lg font-medium">Invoices</h2>
        <div className="mt-4 grid gap-3">
          {invoices.data?.invoices?.length ? (
            invoices.data.invoices.map((invoice) => (
              <div
                key={invoice.id}
                className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-white/5 px-4 py-3 text-sm"
              >
                <span className="text-white/70">
                  {formatPrice(invoice.amountMinor, invoice.currency)}
                </span>
                <span className="text-white/45">{invoice.status}</span>
                {invoice.invoiceUrl && (
                  <a
                    className="text-indigo-200 hover:text-indigo-100"
                    href={invoice.invoiceUrl}
                    target="_blank"
                    rel="noreferrer"
                  >
                    View invoice
                  </a>
                )}
              </div>
            ))
          ) : (
            <p className="text-sm text-white/45">
              No invoices have been issued yet.
            </p>
          )}
        </div>
      </section>
    </main>
  );
}

function formatPrice(minor: number, currency: string) {
  return new Intl.NumberFormat(undefined, {
    style: 'currency',
    currency,
  }).format(minor / 100);
}
