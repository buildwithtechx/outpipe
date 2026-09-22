import { createFileRoute } from '@tanstack/react-router';
import { ProfileSettingsPage } from '#/features/organizations';

export const Route = createFileRoute('/$orgSlug/settings/profile')({
  component: ProfileSettingsRoute,
});

function ProfileSettingsRoute() {
  return <ProfileSettingsPage />;
}
