import { createFileRoute } from '@tanstack/react-router';
import { AdminOrganizationsPage } from '#/features/admin/admin-organizations-page';

export const Route = createFileRoute('/admin/organizations/')({
  component: AdminOrganizationsRoute,
});

function AdminOrganizationsRoute() {
  return <AdminOrganizationsPage />;
}
