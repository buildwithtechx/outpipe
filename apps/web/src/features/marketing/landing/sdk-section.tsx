import { Link } from '@tanstack/react-router';
import { ArrowRight, Braces } from 'lucide-react';
import { MarketingContainer } from '#/components/layout';
import { pluginDefinitions } from '../plugins/plugin-data';

const sdkPlugins = pluginDefinitions.filter((plugin) =>
  ['sdk', 'go', 'rust', 'php', 'angular'].includes(plugin.id),
);

export function SdkSection() {
  return (
    <section className="py-24 sm:py-32">
      <MarketingContainer>
        <div className="mx-auto max-w-2xl text-center">
          <div className="mx-auto inline-flex items-center gap-2 text-sm text-indigo-300">
            <Braces className="size-4" />
            Official SDKs
          </div>
          <h2 className="mt-4 text-3xl font-semibold tracking-[-0.05em] sm:text-5xl">
            Keep your stack. Add a public edge.
          </h2>
          <p className="mt-5 text-base leading-7 text-white/50 sm:text-lg">
            Native packages for the runtimes teams use to ship services,
            previews, and integrations.
          </p>
        </div>
        <div className="mx-auto mt-12 grid max-w-5xl gap-3 sm:grid-cols-2 lg:grid-cols-5">
          {sdkPlugins.map((plugin) => {
            const Icon = plugin.icon;
            return (
              <Link
                key={plugin.id}
                to="/plugins/$pluginId"
                params={{ pluginId: plugin.id }}
                className="group flex items-center gap-3 rounded-2xl border border-white/10 bg-white/[0.025] px-4 py-4 transition-colors hover:border-indigo-300/35 hover:bg-indigo-300/[0.05]"
              >
                <Icon className={`size-5 ${plugin.colorClass}`} />
                <span className="text-sm font-medium text-white/70 group-hover:text-white">
                  {plugin.name}
                </span>
                <ArrowRight className="ml-auto size-3.5 text-white/25 transition-transform group-hover:translate-x-0.5 group-hover:text-white" />
              </Link>
            );
          })}
        </div>
      </MarketingContainer>
    </section>
  );
}
