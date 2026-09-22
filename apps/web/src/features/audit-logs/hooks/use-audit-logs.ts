import { useQuery } from '@tanstack/react-query';
import { getOrganizationAuditLogs } from '../services/audit-log-service';

export function useAuditLogs(organizationId?: string) {
  return useQuery({
    queryKey: ['audit-logs', organizationId],
    queryFn: () => getOrganizationAuditLogs(organizationId as string),
    enabled: Boolean(organizationId),
  });
}
