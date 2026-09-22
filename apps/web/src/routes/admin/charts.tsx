import { createFileRoute } from '@tanstack/react-router';
import { AdminChartsPage } from '#/features/admin/admin-charts-page';

export const Route = createFileRoute('/admin/charts')({
  component: AdminChartsRoute,
});

function AdminChartsRoute() {
  return <AdminChartsPage />;
}
