import { createFileRoute } from '@tanstack/react-router';
import { MembersPage } from '#/features/organizations';

export const Route = createFileRoute('/$orgSlug/members')({
  component: MembersRoute,
});

function MembersRoute() {
  const { orgSlug } = Route.useParams();
  return <MembersPage orgSlug={orgSlug} />;
}
