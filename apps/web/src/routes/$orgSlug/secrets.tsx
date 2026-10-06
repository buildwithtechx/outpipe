import { createFileRoute } from '@tanstack/react-router';
import { SecretsPage } from '#/features/secrets';

export const Route = createFileRoute('/$orgSlug/secrets')({
  component: SecretsRoute,
});

function SecretsRoute() {
  const { orgSlug } = Route.useParams();
  return <SecretsPage key={orgSlug} orgSlug={orgSlug} />;
}
