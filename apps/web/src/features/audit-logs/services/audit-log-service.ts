import type { AuditEvent } from '#/interfaces/audit';
import { apiClient } from '#/lib/api-client';

export function getOrganizationAuditLogs(organizationId: string) {
  return apiClient.get<{ events: AuditEvent[] }>(
    `/api/v1/organizations/${organizationId}/audit-logs?limit=100`,
  );
}
