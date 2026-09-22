import { createFileRoute } from '@tanstack/react-router';
import { ApiKeysPage } from '#/features/api-keys';

export const Route = createFileRoute('/$orgSlug/api-keys')({
  component: ApiKeysRoute,
});

function ApiKeysRoute() {
  const { orgSlug } = Route.useParams();
  return <ApiKeysPage orgSlug={orgSlug} />;
}
