export type ContactRequest = {
  name: string;
  email: string;
  topic: string;
  message: string;
};

export type BugReportRequest = {
  name: string;
  email: string;
  category: string;
  summary: string;
  reproduction: string;
  expected: string;
  actual: string;
};
