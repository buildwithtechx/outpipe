import { useNavigate } from '@tanstack/react-router';
import { useEffect } from 'react';
import { useAuthSession } from '#/features/auth/hooks/use-auth-session';
import { getLastOrganizationSlug } from '#/features/auth/services/auth-service';
import { getOrganizations } from '#/features/organizations/services/organization-service';

export function useAuthRedirect() {
  const navigate = useNavigate();
  const {
    data: session,
    isAuthenticated,
    isError,
    isLoading,
  } = useAuthSession();

  useEffect(() => {
    if (isLoading || isError || !isAuthenticated || !session) return;
    let cancelled = false;

    void getOrganizations()
      .then((organizations) => {
        if (cancelled) return;
        const lastSlug = getLastOrganizationSlug();
        const lastOrganization = organizations.find(
          (organization) => organization.slug === lastSlug,
        );
        if (lastOrganization) {
          void navigate({
            to: '/$orgSlug',
            params: { orgSlug: lastOrganization.slug },
            replace: true,
          });
          return;
        }
        if (organizations.length === 1) {
          void navigate({
            to: '/$orgSlug',
            params: { orgSlug: organizations[0].slug },
            replace: true,
          });
          return;
        }
        void navigate({ to: '/select', replace: true });
      })
      .catch(() => {
        if (!cancelled) void navigate({ to: '/select', replace: true });
      });

    return () => {
      cancelled = true;
    };
  }, [isAuthenticated, isError, isLoading, navigate, session]);
}
