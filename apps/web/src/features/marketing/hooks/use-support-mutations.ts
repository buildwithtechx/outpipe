import { useMutation } from '@tanstack/react-query';
import type { BugReportRequest, ContactRequest } from '#/interfaces/support';
import { submitBugReport, submitContact } from '../services/support-service';

export function useSupportMutations() {
  const contact = useMutation({
    mutationFn: (input: ContactRequest) => submitContact(input),
  });
  const bugReport = useMutation({
    mutationFn: (input: BugReportRequest) => submitBugReport(input),
  });
  return { contact, bugReport };
}
