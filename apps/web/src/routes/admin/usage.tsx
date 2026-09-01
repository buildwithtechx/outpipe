import { createFileRoute } from '@tanstack/react-router';
import { AdminUsagePage } from '#/features/admin/admin-usage-page';

export const Route = createFileRoute('/admin/usage')({
  component: AdminUsageRoute,
});

function AdminUsageRoute() {
  return <AdminUsagePage />;
}
