import { useQuery } from '@tanstack/react-query';
import { getMembers } from '../services/member-service';

export function useMembers(organizationId?: string) {
  return useQuery({
    queryKey: ['members', organizationId],
    queryFn: () => getMembers(organizationId as string),
    enabled: Boolean(organizationId),
  });
}
