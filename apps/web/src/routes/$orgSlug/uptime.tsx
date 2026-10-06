import { createFileRoute } from '@tanstack/react-router';
import { UptimePage } from '#/features/uptime';

export const Route = createFileRoute('/$orgSlug/uptime')({
  component: UptimeRoute,
});

function UptimeRoute() {
  const { orgSlug } = Route.useParams();
  return <UptimePage key={orgSlug} orgSlug={orgSlug} />;
}
