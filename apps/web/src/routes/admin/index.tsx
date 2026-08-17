import { createFileRoute } from '@tanstack/react-router';
import { PagePlaceholder } from '#/components/layout/page-placeholder';

export const Route = createFileRoute('/admin/')({ component: AdminPage });

function AdminPage() {
  return (
    <PagePlaceholder
      title="Platform administration"
      description="Monitor and operate the Outpipe platform."
    />
  );
}
