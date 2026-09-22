import { useQuery } from '@tanstack/react-query';
import type { AuditEvent } from '#/interfaces/audit';
import { apiClient } from '#/lib/api-client';

export function useAdminAuditLogs() {
  return useQuery({
    queryKey: ['admin', 'audit-logs'],
    queryFn: () =>
      apiClient.get<{ items: AuditEvent[]; total: number }>(
        '/api/v1/admin/audit-logs?limit=100',
      ),
  });
}
