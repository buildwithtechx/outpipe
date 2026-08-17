import { createFileRoute } from '@tanstack/react-router';
import { ContactPage } from '#/features/marketing';
import { createSeo } from '#/lib/seo';

export const Route = createFileRoute('/_marketing/contact')({
  head: () =>
    createSeo({
      title: 'Contact — Outpipe',
      description:
        'Contact the Outpipe team with product and integration questions.',
      path: '/contact',
    }),
  component: ContactPage,
});
