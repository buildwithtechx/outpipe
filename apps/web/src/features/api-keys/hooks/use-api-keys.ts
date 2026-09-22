import { useQuery } from '@tanstack/react-query';
import { getApiKeys } from '../services/api-key-service';

export function useApiKeys(organizationId?: string) {
  return useQuery({
    queryKey: ['api-keys', organizationId],
    queryFn: () => getApiKeys(organizationId as string),
    enabled: Boolean(organizationId),
  });
}
