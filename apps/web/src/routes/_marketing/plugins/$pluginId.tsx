import { createFileRoute, notFound } from '@tanstack/react-router';
import { getPluginDefinition, PluginDetailPage } from '#/features/marketing';
import { createSeo } from '#/lib/seo';

export const Route = createFileRoute('/_marketing/plugins/$pluginId')({
  head: ({ params }) => {
    const plugin = getPluginDefinition(params.pluginId);
    return createSeo({
      title: `${plugin?.name ?? 'Integration'} — Outpipe`,
      description: plugin?.description,
      path: `/plugins/${params.pluginId}`,
    });
  },
  component: PluginRoute,
});

function PluginRoute() {
  const plugin = getPluginDefinition(Route.useParams().pluginId);
  if (!plugin) throw notFound();
  return <PluginDetailPage plugin={plugin} />;
}
