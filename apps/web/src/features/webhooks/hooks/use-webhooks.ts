import { useQuery } from '@tanstack/react-query';
import { getWebhooks } from '../services/webhook-service';

export function useWebhooks(organizationId?: string) {
  return useQuery({
    queryKey: ['webhooks', organizationId],
    queryFn: () => getWebhooks(organizationId as string),
    enabled: Boolean(organizationId),
  });
}
