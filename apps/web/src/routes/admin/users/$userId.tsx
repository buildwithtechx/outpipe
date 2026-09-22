import { createFileRoute } from '@tanstack/react-router';
import { AdminUserPage } from '#/features/admin/admin-user-page';

export const Route = createFileRoute('/admin/users/$userId')({
  component: AdminUserRoute,
});

function AdminUserRoute() {
  const { userId } = Route.useParams();
  return <AdminUserPage userId={userId} />;
}
