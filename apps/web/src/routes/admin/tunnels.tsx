import { createFileRoute } from '@tanstack/react-router';
import { AdminTunnelsPage } from '#/features/admin/admin-tunnels-page';

export const Route = createFileRoute('/admin/tunnels')({
  component: AdminTunnelsRoute,
});

function AdminTunnelsRoute() {
  return <AdminTunnelsPage />;
}
