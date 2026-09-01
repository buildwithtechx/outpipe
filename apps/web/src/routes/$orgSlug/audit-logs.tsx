import { createFileRoute } from '@tanstack/react-router';
import { AuditLogsPage } from '#/features/audit-logs';

export const Route = createFileRoute('/$orgSlug/audit-logs')({
  component: AuditLogsRoute,
});

function AuditLogsRoute() {
  const { orgSlug } = Route.useParams();
  return <AuditLogsPage orgSlug={orgSlug} />;
}
