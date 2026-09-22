import { createFileRoute, redirect } from '@tanstack/react-router';
import { AdminDashboardLayout } from '#/components/navigation';
import { getAuthSession } from '#/features/auth/services/auth-service';
import { requirePlatformAdmin } from '#/lib/route-guards';

export const Route = createFileRoute('/admin')({
  beforeLoad: async ({ context }) => {
    try {
      const session = await context.queryClient.fetchQuery({
        queryKey: ['auth', 'session'],
        queryFn: getAuthSession,
        retry: false,
      });
      requirePlatformAdmin(session);
    } catch (error) {
      if (
        error instanceof Error &&
        error.message === 'platform admin access is required'
      ) {
        throw redirect({ to: '/' });
      }
      throw redirect({ to: '/login' });
    }
  },
  component: AdminLayout,
});

function AdminLayout() {
  return <AdminDashboardLayout />;
}
