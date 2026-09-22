import { createFileRoute } from '@tanstack/react-router';
import { UsagePage } from '#/features/usage';

export const Route = createFileRoute('/$orgSlug/usage')({
  component: UsageRoute,
});

function UsageRoute() {
  const { orgSlug } = Route.useParams();
  return <UsagePage orgSlug={orgSlug} />;
}
