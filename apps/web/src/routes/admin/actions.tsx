import { createFileRoute } from '@tanstack/react-router';
import { AdminActionsPage } from '#/features/admin/admin-actions-page';

export const Route = createFileRoute('/admin/actions')({
  component: AdminActionsRoute,
});

function AdminActionsRoute() {
  return <AdminActionsPage />;
}
