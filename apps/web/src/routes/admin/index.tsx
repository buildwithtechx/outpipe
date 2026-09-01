import { createFileRoute } from '@tanstack/react-router';
import { AdminOverviewPage } from '#/features/admin/admin-overview-page';

export const Route = createFileRoute('/admin/')({ component: AdminPage });

function AdminPage() {
  return <AdminOverviewPage />;
}
