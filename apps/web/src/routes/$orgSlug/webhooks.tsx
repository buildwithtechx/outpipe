import { createFileRoute } from '@tanstack/react-router';
import { WebhooksPage } from '#/features/webhooks';

export const Route = createFileRoute('/$orgSlug/webhooks')({
  component: WebhooksRoute,
});

function WebhooksRoute() {
  const { orgSlug } = Route.useParams();
  return <WebhooksPage orgSlug={orgSlug} />;
}
