import { useOrganization } from '#/features/organizations/hooks/use-organization';
import { useBilling } from './hooks/use-billing';

export function BillingPage({ orgSlug }: { orgSlug: string }) {
  const organizationQuery = useOrganization(orgSlug);
  const query = useBilling(organizationQuery.organization?.id);
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
