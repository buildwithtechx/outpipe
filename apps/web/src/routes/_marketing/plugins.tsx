import { createFileRoute, Outlet } from '@tanstack/react-router';

export const Route = createFileRoute('/_marketing/plugins')({
  component: PluginsLayout,
});

function PluginsLayout() {
  return <Outlet />;
}
