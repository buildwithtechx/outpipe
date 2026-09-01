import { useNavigate } from '@tanstack/react-router';
import { useEffect } from 'react';
import { useAuthSession } from '#/features/auth/hooks/use-auth-session';
import { getLastOrganizationSlug } from '#/features/auth/services/auth-service';
import { getOrganizations } from '#/features/organizations/services/organization-service';

export function useAuthRedirect() {
  const navigate = useNavigate();
  const { data: session } = useAuthSession();

  useEffect(() => {
    if (!session) return;
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
          });
          return;
        }
        if (organizations.length === 1) {
          void navigate({
            to: '/$orgSlug',
            params: { orgSlug: organizations[0].slug },
          });
          return;
        }
        void navigate({ to: '/select' });
      })
      .catch(() => {
        if (!cancelled) void navigate({ to: '/select' });
      });

    return () => {
      cancelled = true;
    };
  }, [navigate, session]);
}
