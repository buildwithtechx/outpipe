import { createFileRoute } from '@tanstack/react-router';
import { StatusPageEditor } from '#/features/uptime/status-page-editor';

export const Route = createFileRoute('/$orgSlug/uptime_/status-page')({
  component: StatusPageRoute,
});

function StatusPageRoute() {
  const { orgSlug } = Route.useParams();
  return <StatusPageEditor key={orgSlug} orgSlug={orgSlug} />;
}
