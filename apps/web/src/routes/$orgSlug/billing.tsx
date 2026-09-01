import { createFileRoute } from '@tanstack/react-router';
import { BillingPage } from '#/features/billing';

export const Route = createFileRoute('/$orgSlug/billing')({
  component: BillingRoute,
});

function BillingRoute() {
  const { orgSlug } = Route.useParams();
  return <BillingPage orgSlug={orgSlug} />;
}
