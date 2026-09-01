import { useQuery } from '@tanstack/react-query';
import { getOrganizations } from '#/features/organizations/services/organization-service';
import { getTunnels } from '../services/tunnel-service';

export function useOrganizationTunnels(orgSlug: string) {
  const organizationsQuery = useQuery({
    queryKey: ['organizations'],
    queryFn: getOrganizations,
  });
  const organization = organizationsQuery.data?.find(
    (candidate) => candidate.slug === orgSlug,
  );
  const tunnelsQuery = useQuery({
    queryKey: ['tunnels', organization?.id],
    queryFn: () => getTunnels(organization?.id ?? ''),
    enabled: Boolean(organization?.id),
  });

  return { organization, organizationsQuery, tunnelsQuery };
}
