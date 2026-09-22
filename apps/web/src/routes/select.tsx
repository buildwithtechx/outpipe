import { createFileRoute } from '@tanstack/react-router';
import { OrganizationSelectionPage } from '#/features/organizations/organization-selection-page';
import { createSeo } from '#/lib/seo';

export const Route = createFileRoute('/select')({
  head: () => createSeo({ title: 'Choose a workspace', path: '/select' }),
  component: OrganizationSelectionPage,
});
