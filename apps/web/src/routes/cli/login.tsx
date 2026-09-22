import { createFileRoute } from '@tanstack/react-router';
import { PagePlaceholder } from '#/components/layout/page-placeholder';

export const Route = createFileRoute('/cli/login')({ component: CliLoginPage });

function CliLoginPage() {
  return (
    <PagePlaceholder
      title="Authorize CLI"
      description="Approve a device login request from the Outpipe CLI."
    />
  );
}
