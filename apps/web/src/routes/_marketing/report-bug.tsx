import { createFileRoute } from '@tanstack/react-router';
import { ReportBugPage } from '#/features/marketing';
import { createSeo } from '#/lib/seo';

export const Route = createFileRoute('/_marketing/report-bug')({
  head: () =>
    createSeo({
      title: 'Report a bug — Outpipe',
      description:
        'Send the Outpipe team a clear, safe reproduction of a product issue.',
      path: '/report-bug',
    }),
  component: ReportBugPage,
});
