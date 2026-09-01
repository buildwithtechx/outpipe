import { createFileRoute, Outlet, redirect } from '@tanstack/react-router';
import { getAuthSession } from '#/features/auth/services/auth-service';
import { getOrganizations } from '#/features/organizations/services/organization-service';

export const Route = createFileRoute('/$orgSlug')({
  beforeLoad: async ({ context, params }) => {
    try {
      await context.queryClient.fetchQuery({
        queryKey: ['auth', 'session'],
        queryFn: getAuthSession,
        retry: false,
      });
      const organizations = await context.queryClient.fetchQuery({
        queryKey: ['organizations'],
        queryFn: getOrganizations,
      });
      const organization = organizations.find(
        (item) => item.slug === params.orgSlug,
      );
      if (!organization) throw redirect({ to: '/select' });
    } catch (error) {
      if (error instanceof Response) throw error;
      if (error && typeof error === 'object' && 'isRedirect' in error) {
        throw error;
      }
      throw redirect({ to: '/login' });
    }
  },
  component: OrganizationLayout,
});

function OrganizationLayout() {
  return <Outlet />;
}
