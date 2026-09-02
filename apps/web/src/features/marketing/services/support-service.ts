import type { BugReportRequest, ContactRequest } from '#/interfaces/support';
import { apiClient } from '#/lib/api-client';

export function submitContact(input: ContactRequest) {
  return apiClient.post<void>('/api/v1/support/contact', input);
}

export function submitBugReport(input: BugReportRequest) {
  return apiClient.post<void>('/api/v1/support/bug-report', input);
}
