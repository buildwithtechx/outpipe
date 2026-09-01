import { createFileRoute } from '@tanstack/react-router';
import { AdminSubscriptionsPage } from '#/features/admin/admin-subscriptions-page';

export const Route = createFileRoute('/admin/subscriptions')({
  component: AdminSubscriptionsRoute,
});

function AdminSubscriptionsRoute() {
  return <AdminSubscriptionsPage />;
}
