import { useQuery } from '@tanstack/react-query';
import { getOrganizations } from '../services/organization-service';

export function useOrganization(orgSlug: string) {
  const query = useQuery({
    queryKey: ['organizations'],
    queryFn: getOrganizations,
  });
  return {
    ...query,
    organization: query.data?.find((item) => item.slug === orgSlug),
  };
}
